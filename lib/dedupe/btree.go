package dedupe

import (
	"bytes"
	"github.com/google/btree"
	"github.com/prometheus/client_golang/prometheus"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
	"sync"
	"time"
)

type (
	FrameAndTime struct {
		frame []byte
		time  time.Time
	}

	FilterBTree struct {
		events chan tracker.Event

		btreeDegree int
		btree       *btree.BTreeG[FrameAndTime]
		mu          sync.RWMutex

		dedupeCounter prometheus.Counter

		sweepInterval      time.Duration
		sweeperMaxAge      time.Duration
		sweeperControlChan chan int
		sweeperTimerChan   *time.Ticker
	}
	OptionBTree func(tree *FilterBTree)
)

func WithBtreeDegree(degree int) OptionBTree {
	return func(f *FilterBTree) {
		f.btreeDegree = degree
	}
}

func WithDedupeCounterBTree(dedupeCounter prometheus.Counter) OptionBTree {
	return func(f *FilterBTree) {
		f.dedupeCounter = dedupeCounter
	}
}

func WithSweeperInterval(d time.Duration) OptionBTree {
	return func(f *FilterBTree) {
		f.sweepInterval = d
	}
}

// WithDedupeMaxAge sets a time after which we do not consider a frame a duplicate
func WithDedupeMaxAge(d time.Duration) OptionBTree {
	return func(f *FilterBTree) {
		if d > 0 {
			d = -d
		}
		f.sweeperMaxAge = d
	}
}

func NewFilterBTree(opts ...OptionBTree) *FilterBTree {
	f := FilterBTree{
		btreeDegree:   16,
		sweepInterval: 10 * time.Second,
		sweeperMaxAge: -60 * time.Second,
	}

	for _, opt := range opts {
		opt(&f)
	}

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

func (f *FilterBTree) Handle(frame tracker.Frame) tracker.Frame {
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
	f.mu.RLock()
	has := f.btree.Has(item)
	f.mu.RUnlock()
	if has {
		return nil
	}
	item.time = time.Now()
	f.mu.Lock()
	f.btree.ReplaceOrInsert(item)
	f.mu.Unlock()

	// we have a deduped frame, do send it to the dedupe queue
	if nil != f.dedupeCounter {
		f.dedupeCounter.Inc()
	}
	return frame
}

func (f *FilterBTree) sweep() {
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

func (f *FilterBTree) Stop() {
	if nil != f.sweeperControlChan {
		f.sweeperControlChan <- 1
	}
}

func (f *FilterBTree) String() string {
	return "Dedupe (BTree)"
}
