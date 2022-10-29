package dedupe

import (
	"github.com/prometheus/client_golang/prometheus"
	"plane.watch/lib/dedupe/forgetfulmap"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
)

/**
This package provides a way to deduplicate mode_s messages.

Consider a message a duplicate if we have seen it in the last minute
*/

type (
	Option func(*Filter)
	Filter struct {
		events chan tracker.Event
		list   *forgetfulmap.ForgetfulSyncMap

		dedupeCounter prometheus.Counter
	}
)

func WithDedupeCounter(dedupeCounter prometheus.Counter) Option {
	return func(f *Filter) {
		f.dedupeCounter = dedupeCounter
	}
}

func NewFilter(opts ...Option) *Filter {
	f := Filter{
		list: forgetfulmap.NewForgetfulSyncMap(),
	}

	for _, opt := range opts {
		opt(&f)
	}

	return &f
}

func (f *Filter) Handle(frame tracker.Frame) tracker.Frame {
	if nil == frame {
		return nil
	}
	var key string
	switch (frame).(type) {
	case *beast.Frame:
		key = string(frame.(*beast.Frame).Raw())
	case *mode_s.Frame:
		key = frame.(*mode_s.Frame).RawString()
	case *sbs1.Frame:
		// todo: investigate better dedupe detection for sbs1
		key = string(frame.(*sbs1.Frame).Raw())
	default:
	}
	if f.list.HasKeyStr(key) {
		return nil
	}
	f.list.AddKeyStr(key)

	// we have a deduped frame, do send it to the dedupe queue
	if nil != f.dedupeCounter {
		f.dedupeCounter.Inc()
	}
	return frame
}
func (f *Filter) String() string {
	return "Dedupe"
}
