package sink

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"plane.watch/lib/dedupe/forgetfulmap"
	"plane.watch/lib/export"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/tracker"
)

type (
	Destination interface {
		PublishJson(queue string, msg []byte) error
		Stop()
		monitoring.HealthCheck
	}

	Sink struct {
		fsm    *forgetfulmap.ForgetfulSyncMap
		config *Config
		dest   Destination
		events chan tracker.Event

		sendList      map[string]*tracker.PlaneLocationEvent
		sendListMutex sync.Mutex
		sendTicker    *time.Ticker
	}
)

func NewSink(conf *Config, dest Destination) tracker.Sink {
	s := Sink{
		fsm:      forgetfulmap.NewForgetfulSyncMap(),
		config:   conf,
		dest:     dest,
		events:   make(chan tracker.Event),
		sendList: make(map[string]*tracker.PlaneLocationEvent),
	}

	if s.config.sendDelay > 0 {
		s.sendTicker = time.NewTicker(s.config.sendDelay)
		go s.doSend()
	}

	return &s
}

func (s *Sink) Listen() chan tracker.Event {
	return s.events
}

// Server pokes a hole in the abstraction to get to the nats server
func (s *Sink) Server() any {
	if ns, ok := s.dest.(*NatsSink); ok {
		return ns.server
	}
	return nil
}

func (s *Sink) Stop() {
	close(s.events)
	s.config.Finish()
	s.dest.Stop()
	s.fsm.Stop()
	s.sendTicker.Stop()
}

func trackerMsgJSON(le *tracker.PlaneLocationEvent, sourceTag string) ([]byte, error) {
	plane := le.Plane()
	if nil == plane {
		return nil, errors.New("no plane")
	}

	eventStruct := export.NewPlaneLocation(plane, le.New(), le.Removed(), sourceTag)
	return eventStruct.ToJSONBytes()
}

func trackerMsgProtobuf(le *tracker.PlaneLocationEvent, sourceTag string) ([]byte, error) {
	plane := le.Plane()
	if nil == plane {
		return nil, errors.New("no plane")
	}

	eventStruct := export.NewPlaneInfo(plane, sourceTag)
	defer export.Release(eventStruct)
	return eventStruct.ToProtobufBytes()
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
		jsonBuf, err = s.config.byteMaker(le, s.config.sourceTag)
		if nil != jsonBuf && nil == err {
			_ = s.dest.PublishJson(s.config.queueName, jsonBuf)
			if nil != s.config.stats.planeLoc {
				s.config.stats.planeLoc.Inc()
			}
		}
	}
}

// OnEvent gets called once for each message we want to send on the bus
func (s *Sink) OnEvent(e tracker.Event) {
	var err error
	if le, ok := e.(*tracker.PlaneLocationEvent); ok {
		if s.config.sendDelay == 0 {
			// warning, this code is a duplicate of the sendLocationList handling
			var jsonBuf []byte
			jsonBuf, err = s.config.byteMaker(le, s.config.sourceTag)
			if jsonBuf != nil && err == nil {
				err = s.dest.PublishJson(s.config.queueName, jsonBuf)
				if err != nil {
					fmt.Println(err)
				}
				if nil != s.config.stats.planeLoc {
					s.config.stats.planeLoc.Inc()
				}
			}
		} else {
			s.sendListMutex.Lock()
			s.sendList[le.Plane().IcaoIdentifierStr()] = le
			s.sendListMutex.Unlock()
		}
	}
}

func (s *Sink) HealthCheckName() string {
	return s.dest.HealthCheckName()
}

func (s *Sink) HealthCheck() bool {
	return s.dest.HealthCheck()
}
