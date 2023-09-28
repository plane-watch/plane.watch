package tracker

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
)

type (
	// Option allows us to configure our new Tracker as we need it
	Option func(*Tracker)

	EventMaker interface {
		Stopper
		Listen() chan FrameEvent
	}
	EventListener interface {
		OnEvent(Event)
	}
	Stopper interface {
		Stop()
	}
	// Frame is our general object for a tracking update, AVR, SBS1, Modes Beast Binary
	Frame interface {
		Icao() uint32
		IcaoStr() string
		Decode() error
		TimeStamp() time.Time
		Raw() []byte
	}

	// A Producer can listen for or generate Frames, it provides the output via a channel that the handler can then
	// process further.
	// A Producer can send *FrameEvent events
	Producer interface {
		EventMaker
		fmt.Stringer
		monitoring.HealthCheck
	}

	// Sink is something that takes the output from our producers and trackers
	Sink interface {
		EventListener
		Stopper
		monitoring.HealthCheck
	}

	// Middleware has a chance to modify a frame before we send it to the plane Tracker
	Middleware interface {
		fmt.Stringer
		monitoring.HealthCheck
		Handle(*FrameEvent) Frame
	}
)

func WithDecodeWorkerCount(numDecodeWorkers int) Option {
	return func(t *Tracker) {
		t.decodeWorkerCount = numDecodeWorkers
	}
}
func WithPruneTiming(pruneTick, pruneAfter time.Duration) Option {
	return func(t *Tracker) {
		t.pruneTick = pruneTick
		t.pruneAfter = pruneAfter
	}
}
func WithPrometheusCounters(currentPlanes prometheus.Gauge, decodedFrames prometheus.Counter) Option {
	return func(t *Tracker) {
		t.stats.currentPlanes = currentPlanes
		t.stats.decodedFrames = decodedFrames
	}
}

// Finish begins the ending of the tracking by closing our decoding queue
func (t *Tracker) Finish() {
	if t.finishDone {
		return
	}
	t.finishDone = true
	log.Debug().Msg("Starting Finish()")
	for _, p := range t.producers {
		log.Debug().Str("producer", p.String()).Msg("Stopping Producer")
		p.Stop()
	}
	log.Debug().Msg("Closing Decoding Queue")
	close(t.decodingQueue)
	t.planeList.Stop()
}

func (t *Tracker) EventListener(eventSource EventMaker, waiter *sync.WaitGroup) {
	for e := range eventSource.Listen() {
		t.decodingQueue <- &e
	}
	waiter.Done()
	t.log.Debug().Msg("Done with Event Source")
}

// AddProducer wires up a Producer to start feeding data into the tracker
func (t *Tracker) AddProducer(p Producer) {
	if nil == p {
		return
	}
	monitoring.AddHealthCheck(p)

	t.log.Debug().Str("producer", p.String()).Msg("Adding producer")
	t.producers = append(t.producers, p)
	t.producerWaiter.Add(1)

	go t.EventListener(p, &t.producerWaiter)
	t.log.Debug().Msg("Just added a producer")
}

// AddMiddleware wires up a Middleware which each message will go through before being added to the tracker
func (t *Tracker) AddMiddleware(m Middleware) {
	if nil == m {
		return
	}
	monitoring.AddHealthCheck(m)
	t.log.Debug().Str("name", m.String()).Msg("Adding middleware")
	t.middlewares = append(t.middlewares, m)

	t.log.Debug().Msg("Just added a middleware")
}

// SetSink wires up a Sink in the tracker. Whenever an event happens it gets sent to each Sink
func (t *Tracker) SetSink(s Sink) {
	t.log.Debug().Str("name", s.HealthCheckName()).Msg("Set Sink")
	if nil == s {
		return
	}
	t.sink = s
	monitoring.AddHealthCheck(s)
}

// Stop attempts to stop all the things, mid-flight. Use this if you have something else waiting for things to finish
// use this if you are listening to remote sources
func (t *Tracker) Stop() {
	t.Finish()
	t.producerWaiter.Wait()
	t.decodingQueueWaiter.Wait()
	t.eventsWaiter.Wait()
	t.middlewareWaiter.Wait()
}

// StopOnCancel listens for SigInt etc. and gracefully stops
func (t *Tracker) StopOnCancel() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	isStopping := false
	exitChan := make(chan bool, 3)
	for {
		select {
		case sig := <-ch:
			log.Info().Str("Signal", sig.String()).Msg("Received Interrupt, stopping")
			if !isStopping {
				isStopping = true
				go func() {
					t.Stop()
					exitChan <- true
					log.Info().Msg("Done Stopping")
				}()
				go func() {
					time.Sleep(time.Second * 5)
					exitChan <- true
					log.Info().Msg("Timeout after 5 seconds, force stopping")
					os.Exit(1)
				}()
			} else {
				log.Info().Str("Signal", sig.String()).Msg("Second Interrupt, forcing exit")
				os.Exit(1)
			}
		case <-exitChan:
			return
		}
	}
}

// Wait waits for all producers to stop producing input and then returns
// use this method if you are processing a file
func (t *Tracker) Wait() {
	t.producerWaiter.Wait()
	log.Debug().Msg("Producers finished")
	time.Sleep(time.Millisecond * 50)
	t.Finish()
	log.Debug().Msg("Finish() done")
	t.decodingQueueWaiter.Wait()
	log.Debug().Msg("Decoding waiter done")
	t.eventsWaiter.Wait()
	log.Debug().Msg("events waiter done")
}

func (t *Tracker) decodeQueue() {
	for f := range t.decodingQueue {
		if nil == f {
			continue
		}
		if nil != t.stats.decodedFrames {
			t.stats.decodedFrames.Inc()
		}
		frame := f.Frame()
		err := frame.Decode()
		if nil != err {
			if !errors.Is(mode_s.ErrNoOp, err) {
				// the decode operation failed to produce valid output, and we tell someone about it
				t.log.Error().Err(err).Str("Tag", f.Source().Tag).Send()
			}
			continue
		}

		for _, m := range t.middlewares {
			frame = m.Handle(f)
			if nil == frame {
				break
			}
		}
		if nil == frame || frame.Icao() == 0 {
			// invalid frame || unable to determine planes ICAO
			continue
		}
		plane := t.GetPlane(frame.Icao())

		switch typeFrame := frame.(type) {
		case *beast.Frame:

			plane.HandleModeSFrame(typeFrame.AvrFrame(), f.Source().RefLat, f.Source().RefLon)
			plane.setSignalLevel(typeFrame.SignalRssi())
		case *mode_s.Frame:
			plane.HandleModeSFrame(typeFrame, f.Source().RefLat, f.Source().RefLon)
		case *sbs1.Frame:
			plane.HandleSbs1Frame(typeFrame)
		default:
			t.log.Error().Str("Tag", f.Source().Tag).Msg("unknown frame type, cannot track")
		}
	}
	t.decodingQueueWaiter.Done()
}
