package dedupe

import (
	"fmt"
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
	Filter struct {
		events chan tracker.Event
		list   *forgetfulmap.ForgetfulSyncMap
	}
)

func NewFilter() *Filter {
	return &Filter{
		list:   forgetfulmap.NewForgetfulSyncMap(),
		events: make(chan tracker.Event),
	}
}

func (f *Filter) Listen() chan tracker.Event {
	return f.events
}

func (f *Filter) Stop() {
	close(f.events)
	f.list.Stop()
}

func (f *Filter) String() string {
	return "Dedupe"
}

func (f *Filter) addDedupedFrame(frame tracker.Frame, src *tracker.FrameSource) {
	defer func() {
		if nil != recover() {
			// it's ok, we didn't need that message anyway...
		}
	}()

	event := tracker.NewDedupedFrameEvent(frame, src)
	f.events <- event
}

func (f *Filter) Handle(frame tracker.Frame, src *tracker.FrameSource) tracker.Frame {
	if nil == frame {
		return nil
	}
	var key interface{}
	switch (frame).(type) {
	case *beast.Frame:
		key = fmt.Sprintf("%X", frame.(*beast.Frame).AvrRaw())
	case *mode_s.Frame:
		key = fmt.Sprintf("%X", frame.(*mode_s.Frame).Raw())
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
	f.addDedupedFrame(frame, src)
	return frame
}
