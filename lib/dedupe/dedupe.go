package dedupe

import (
	"bytes"
	"github.com/google/btree"
	"github.com/prometheus/client_golang/prometheus"
	"plane.watch/lib/dedupe/forgetfulmap"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
	"sync"
	"time"
)

/**
This package provides a way to deduplicate mode_s messages.

Consider a message a duplicate if we have seen it in the last minute
*/

type (
	FrameAndTime struct {
		frame []byte
		time  time.Time
	}

	Option func(*Filter)
	Filter struct {
		events chan tracker.Event
		list   *forgetfulmap.ForgetfulSyncMap

		btreeDegree int
		btree       *btree.BTreeG[FrameAndTime]
		mu          sync.Mutex

		dedupeCounter prometheus.Counter

		sweepInterval      time.Duration
		sweeperMaxAge      time.Duration
		sweeperControlChan chan int
		sweeperTimerChan   *time.Ticker
	}
)

func WithBtreeDegree(degree int) Option {
	return func(f *Filter) {
		f.btreeDegree = degree
	}
}

func WithDedupeCounter(dedupeCounter prometheus.Counter) Option {
	return func(f *Filter) {
		f.dedupeCounter = dedupeCounter
	}
}

func WithSweeperDuration(d time.Duration) Option {
	return func(f *Filter) {
		f.sweepInterval = d
	}
}

// WithDedupeMaxAge sets a time after which we do not consider a frame a duplicate
func WithDedupeMaxAge(d time.Duration) Option {
	return func(f *Filter) {
		if d > 0 {
			d = -d
		}
		f.sweeperMaxAge = d
	}
}

func NewFilter(opts ...Option) *Filter {
	f := Filter{
		btreeDegree:   16,
		sweepInterval: 10 * time.Second,
		sweeperMaxAge: -60 * time.Second,
	}

	for _, opt := range opts {
		opt(&f)
	}

	f.list = forgetfulmap.NewForgetfulSyncMap()
	f.btree = btree.NewG[FrameAndTime](f.btreeDegree, func(a, b FrameAndTime) bool {
		return bytes.Compare(a.frame, b.frame) < 0
	})

	// btree sweeping
	if f.sweepInterval > 0 {
		f.sweeperControlChan = make(chan int)
		f.sweeperTimerChan = time.NewTicker(f.sweepInterval)
		go func() {
			for {
				select {
				case <-f.sweeperControlChan:
					return
				case <-f.sweeperTimerChan.C:
					f.sweep()
				}
			}
		}()
	}

	return &f
}

func (f *Filter) sweep() {
	f.mu.Lock()
	defer f.mu.Unlock()
	// sweep btree and prune old entries
	olderThan := time.Now().Add(f.sweeperMaxAge)
	toRemove := make([]FrameAndTime, 0, f.btree.Len()/2)
	f.btree.Descend(func(item FrameAndTime) bool {
		if item.time.Before(olderThan) {
			toRemove = append(toRemove, item)
		}
		return true
	})
	for _, item := range toRemove {
		f.btree.Delete(item)
	}
}

func (f *Filter) Stop() {
	if nil != f.sweeperControlChan {
		f.sweeperControlChan <- 1
	}
	f.list.Stop()
}

func (f *Filter) String() string {
	return "Dedupe"
}

func (f *Filter) Handle(frame tracker.Frame) tracker.Frame {
	if nil == frame {
		return nil
	}
	var key []byte
	switch (frame).(type) {
	case *beast.Frame:
		key = frame.(*beast.Frame).Raw()
	case *mode_s.Frame:
		key = frame.(*mode_s.Frame).Raw()
	case *sbs1.Frame:
		// todo: investigate better dedupe detection for sbs1
		key = frame.(*sbs1.Frame).Raw()
	default:
		return nil
	}
	item := FrameAndTime{
		frame: key,
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.btree.Has(item) {
		return nil
	}
	item.time = time.Now()
	f.btree.ReplaceOrInsert(item)

	// we have a deduped frame, do send it to the dedupe queue
	if nil != f.dedupeCounter {
		f.dedupeCounter.Inc()
	}
	return frame
}

func (f *Filter) HandleOld(frame tracker.Frame) tracker.Frame {
	if nil == frame {
		return nil
	}
	var key interface{}
	switch (frame).(type) {
	case *beast.Frame:
		key = frame.(*beast.Frame).RawString()
	case *mode_s.Frame:
		key = frame.(*mode_s.Frame).RawString()
	case *sbs1.Frame:
		// todo: investigate better dedupe detection for sbs1
		key = string(frame.(*sbs1.Frame).Raw())
	default:
	}
	if f.list.HasKey(key) {
		return nil
	}
	f.list.AddKey(key)

	// we have a deduped frame, do send it to the dedupe queue
	if nil != f.dedupeCounter {
		f.dedupeCounter.Inc()
	}
	return frame
}
