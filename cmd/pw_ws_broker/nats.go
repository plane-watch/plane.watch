package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/export"
	"plane.watch/lib/nats_io"
	"sync"
)

type (
	PwWsBrokerNats struct {
		routeLow, routeHigh string
		server              *nats_io.Server
		processMessage      processMessage
	}
)

func NewPwWsBrokerNats(url, routeLow, routeHigh string) (*PwWsBrokerNats, error) {
	svr, err := nats_io.NewServer(url, "pw_ws_broker")
	if nil != err {
		return nil, err
	}
	svr.DroppedCounter(promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ws_broker_nats_dropped_message_err_count",
		Help: "The total number slow consumer dropped message errors.",
	}))
	return &PwWsBrokerNats{
		routeLow:  routeLow,
		routeHigh: routeHigh,
		server:    svr,
	}, nil
}

func (n *PwWsBrokerNats) configure() error {
	return nil
}

func (n *PwWsBrokerNats) setProcessMessage(f processMessage) {
	n.processMessage = f
}

func (n *PwWsBrokerNats) consume(exitChan chan bool, subject, what string) {
	log.Debug().Str("Nats Consume", subject).Str("what", what).Send()
	ch, err := n.server.Subscribe(subject)
	if nil != err {
		log.Error().
			Err(err).
			Str("subject", subject).
			Str("what", what).
			Msg("Failed to consume")
		return
	}
	var wg sync.WaitGroup
	worker := func() {
		for msg := range ch {
			planeData, errDecode := export.FromProtobufBytes(msg.Data)
			if nil != errDecode {
				log.Debug().Err(err).Msg("did not understand msg")
				continue
			}
			n.processMessage(what, planeData)
		}
		wg.Done()
	}
	numWorkers := 10
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	wg.Wait()
	log.Info().
		Str("subject", subject).
		Str("what", what).
		Msg("Finished Consuming")
	exitChan <- true
}

func (n *PwWsBrokerNats) consumeAll(exitChan chan bool) {
	go n.consume(exitChan, n.routeLow, "_low")
	go n.consume(exitChan, n.routeHigh, "_high")
}

func (n *PwWsBrokerNats) close() {
	n.server.Close()
}

func (n *PwWsBrokerNats) HealthCheckName() string {
	return n.server.HealthCheckName()
}

func (n *PwWsBrokerNats) HealthCheck() bool {
	return n.server.HealthCheck()
}
