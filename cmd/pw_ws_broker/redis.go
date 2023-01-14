package main

import (
	"github.com/rs/zerolog/log"
	"plane.watch/lib/export"
	"plane.watch/lib/redismq"
)

type (
	PwWsBrokerRedis struct {
		server              *redismq.Server
		routeLow, routeHigh string
		processMessage      processMessage
	}
)

func NewPwWsBrokerRedis(url, routeLow, routeHigh string) (*PwWsBrokerRedis, error) {
	r := &PwWsBrokerRedis{
		routeLow:  routeLow,
		routeHigh: routeHigh,
	}
	var err error
	r.server, err = redismq.NewServer(url)
	if nil != err {
		return nil, err
	}

	return r, nil
}

func (r *PwWsBrokerRedis) configure() error {
	return nil
}
func (r *PwWsBrokerRedis) setProcessMessage(f processMessage) {
	r.processMessage = f
}
func (r *PwWsBrokerRedis) consume(exitChan chan bool, subject, what string) {
	log.Debug().Str("Redis Consume", subject).Str("what", what).Send()
	ch, err := r.server.Subscribe(subject)
	if nil != err {
		log.Error().
			Err(err).
			Str("subject", subject).
			Str("what", what).
			Msg("Failed to consume")
		return
	}

	for msg := range ch {
		if log.Trace().Enabled() {
			log.Trace().Str("payload", msg.Payload).Send()
		}

		planeData, errJson := export.FromProtobufBytes([]byte(msg.Payload))
		if nil != errJson {
			log.Debug().Err(err).Msg("did not understand msg")
			continue
		}
		r.processMessage(what, planeData)

	}
	log.Info().
		Str("subject", subject).
		Str("what", what).
		Msg("Finished Consuming")
	exitChan <- true
}

func (r *PwWsBrokerRedis) consumeAll(exitChan chan bool) {
	go r.consume(exitChan, r.routeLow, "_low")
	go r.consume(exitChan, r.routeHigh, "_high")
}

func (r *PwWsBrokerRedis) close() {
	r.server.Close()
}

func (r *PwWsBrokerRedis) HealthCheckName() string {
	return r.server.HealthCheckName()
}

func (r *PwWsBrokerRedis) HealthCheck() bool {
	return r.server.HealthCheck()
}
