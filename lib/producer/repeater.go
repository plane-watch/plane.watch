package producer

import (
	"plane.watch/lib/tracker"
	"sync/atomic"
	"time"
)

type (
	keepAliveRepeater struct {
		listFrames map[uint32]*keepAlive
		chanFrame  chan tracker.FrameEvent
		chanExit   chan bool

		frequency time.Duration
		duration  time.Duration
	}
	keepAlive struct {
		msg      tracker.FrameEvent
		repeater *time.Ticker
		until    time.Time

		numInFlight atomic.Int64
	}
)

func newKeepAliveRepeater() *keepAliveRepeater {
	return &keepAliveRepeater{
		listFrames: make(map[uint32]*keepAlive),
		chanFrame:  make(chan tracker.FrameEvent, 100),
		frequency:  time.Second * 30,
		duration:   time.Hour + time.Minute,
	}
}

func (r *keepAliveRepeater) processor(p *Producer) {
	var ka *keepAlive
	var ok bool
	for msg := range r.chanFrame {
		msg.Source().Tag = "repeat"
		icao := msg.Frame().Icao()
		p.log.Debug().Str("ICAO", msg.Frame().IcaoStr()).Msg("New Repeating Frame")
		ka, ok = r.listFrames[icao]
		if !ok {
			ka = &keepAlive{msg: msg}
			r.listFrames[icao] = ka
		} else {
			ka.stop()
		}
		ka.msg = msg
		ka.repeater = time.NewTicker(r.frequency)
		ka.until = time.Now().Add(r.duration)
		go ka.repeat(p)
	}

	for _, repeater := range r.listFrames {
		repeater.stop()
	}
}
func (r *keepAliveRepeater) stop() {
	close(r.chanFrame)
}

func (ka *keepAlive) repeat(p *Producer) {
	icaoStr := ka.msg.Frame().IcaoStr()
	for t := range ka.repeater.C {
		if t.After(ka.until) {
			p.log.Debug().Str("ICAO", icaoStr).Msg("Repeating Event Expired")
			ka.stop()
		} else {
			p.log.Debug().Str("ICAO", icaoStr).Msg("Repeating Event")
			p.AddEvent(ka.msg)
		}
	}
}

func (ka *keepAlive) stop() {
	ka.repeater.Stop()
}
