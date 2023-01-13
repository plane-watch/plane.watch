package sink

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"plane.watch/lib/dedupe/forgetfulmap"
	"plane.watch/lib/export"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/rabbitmq"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
)

type (
	Destination interface {
		PublishJson(queue string, msg []byte) error
		PublishText(queue string, msg []byte) error
		Stop()
		monitoring.HealthCheck
	}

	Sink struct {
		fsm    *forgetfulmap.ForgetfulSyncMap
		config *Config
		dest   Destination
		events chan tracker.Event

		sendFrameAll    func(tracker.Frame, *tracker.FrameSource) error
		sendFrameDedupe func(tracker.Frame, *tracker.FrameSource) error

		sendList      map[string]*tracker.PlaneLocationEvent
		sendListMutex sync.Mutex
		sendTicker    *time.Ticker
	}
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func stripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}

func NewSink(conf *Config, dest Destination) tracker.Sink {
	s := Sink{
		fsm:      forgetfulmap.NewForgetfulSyncMap(),
		config:   conf,
		dest:     dest,
		events:   make(chan tracker.Event),
		sendList: make(map[string]*tracker.PlaneLocationEvent),
	}

	s.sendTicker = time.NewTicker(s.config.sendDelay)
	go s.doSend()

	s.sendFrameAll = s.sendFrameEvent(QueueTypeAvrAll, QueueTypeBeastAll, QueueTypeSbs1All)
	s.sendFrameDedupe = s.sendFrameEvent(QueueTypeAvrReduce, QueueTypeBeastReduce, QueueTypeSbs1Reduce)
	return &s
}

func (s *Sink) Listen() chan tracker.Event {
	return s.events
}

func (s *Sink) Stop() {
	close(s.events)
	s.config.Finish()
	s.dest.Stop()
	s.fsm.Stop()
	s.sendTicker.Stop()
}

func (s *Sink) sendLocationEvent(routingKey string, le *tracker.PlaneLocationEvent) error {
	jsonBuf, err := s.trackerMsgJson(le)
	if nil != jsonBuf && nil == err {
		if s.fsm.HasKey(string(jsonBuf)) {
			// sending a message we have already sent!
			return nil
		}
		s.fsm.AddKey(string(jsonBuf))

		err = s.dest.PublishJson(routingKey, jsonBuf)
	}
	return err
}

func (s *Sink) trackerMsgJson(le *tracker.PlaneLocationEvent) ([]byte, error) {
	plane := le.Plane()
	if nil == plane {
		return nil, errors.New("no plane")
	}

	eventStruct := export.NewPlaneLocation(plane, s.config.sourceTag)
	return eventStruct.ToJsonBytes()
}

func (s *Sink) sendFrameEvent(queueAvr, queueBeast, queueSbs1 string) func(tracker.Frame, *tracker.FrameSource) error {
	return func(ourFrame tracker.Frame, source *tracker.FrameSource) error {
		var err error
		var body []byte
		if nil == ourFrame {
			return nil
		}

		sendMessage := func(info rabbitFrameMsg) error {
			if _, ok := s.config.queue[info.RouteKey]; !ok {
				return nil
			}
			json := jsoniter.ConfigFastest
			body, err = json.Marshal(info)
			if nil != err {
				return err
			}
			return s.dest.PublishJson(info.RouteKey, body)
		}

		switch ourFrame.(type) {
		case *mode_s.Frame:
			err = sendMessage(rabbitFrameMsg{Type: "avr", Body: ourFrame.Raw(), RouteKey: queueAvr, Source: source})
		case *beast.Frame:
			err = sendMessage(rabbitFrameMsg{Type: "beast", Body: ourFrame.Raw(), RouteKey: queueBeast, Source: source})
			err = sendMessage(rabbitFrameMsg{Type: "avr", Body: ourFrame.(*beast.Frame).AvrFrame().Raw(), RouteKey: queueAvr, Source: source})
		case *sbs1.Frame:
			err = sendMessage(rabbitFrameMsg{Type: "sbs1", Body: ourFrame.Raw(), RouteKey: queueSbs1, Source: source})
		}
		return err
	}
}

func (s *Sink) doSend() {
	for range s.sendTicker.C {
		s.sendLocationList()
	}
}
func (s *Sink) sendLocationList() {
	var err error
	s.sendListMutex.Lock()
	list := s.sendList
	s.sendList = make(map[string]*tracker.PlaneLocationEvent)
	s.sendListMutex.Unlock()
	for _, le := range list {
		// warning, this code is a duplicate of the OnEvent handling
		var jsonBuf []byte
		jsonBuf, err = s.trackerMsgJson(le)
		if nil != jsonBuf && nil == err {
			err = s.dest.PublishJson(QueueLocationUpdates, jsonBuf)
			if nil != s.config.stats.planeLoc {
				s.config.stats.planeLoc.Inc()
			}
		}
	}
}

// OnEvent gets called once for each message we want to send on the bus
func (s *Sink) OnEvent(e tracker.Event) {
	var err error
	switch e.(type) {
	case *tracker.PlaneLocationEvent:
		le := e.(*tracker.PlaneLocationEvent)

		if 0 == s.config.sendDelay {
			// warning, this code is a duplicate of the sendLocationList handling
			var jsonBuf []byte
			jsonBuf, err = s.trackerMsgJson(le)
			if nil != jsonBuf && nil == err {
				err = s.dest.PublishJson(QueueLocationUpdates, jsonBuf)
				if nil != s.config.stats.planeLoc {
					s.config.stats.planeLoc.Inc()
				}
			}
		} else {
			s.sendListMutex.Lock()
			s.sendList[le.Plane().IcaoIdentifierStr()] = le
			s.sendListMutex.Unlock()
		}

	case *tracker.FrameEvent:
		ourFrame := e.(*tracker.FrameEvent).Frame()
		source := e.(*tracker.FrameEvent).Source()
		err = s.sendFrameAll(ourFrame, source)
		if nil != s.config.stats.frame {
			s.config.stats.frame.Inc()
		}
	}

	if nil != err {
		log.Error().
			Err(err).
			Str("event-type", e.Type()).
			Str("event", e.String()).
			Msg("Unable to handle event")
		if err == rabbitmq.ErrNilChannel {
			panic(err)
		}
	}
}

func (s *Sink) HealthCheckName() string {
	return s.dest.HealthCheckName()
}

func (s *Sink) HealthCheck() bool {
	return s.dest.HealthCheck()
}
