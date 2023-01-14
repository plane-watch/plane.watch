package main

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"plane.watch/lib/export"
	"plane.watch/lib/rabbitmq"
	"plane.watch/lib/randstr"
	"time"
)

type (
	PwWsBrokerRabbit struct {
		rabbit      *rabbitmq.RabbitMQ
		queuePrefix string
		queueLow    string
		queueHigh   string
		routeLow    string
		routeHigh   string

		processMessage processMessage
	}
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewPwWsBrokerRabbit(rabbitUrl, routeLow, routeHigh string) (*PwWsBrokerRabbit, error) {
	var err error
	rabbitCfg, err := rabbitmq.NewConfigFromUrl(rabbitUrl)
	if nil != err {
		return nil, err
	}
	prefix := "broker_" + randstr.RandString(10) + "_"

	return &PwWsBrokerRabbit{
		rabbit:      rabbitmq.New(rabbitCfg),
		queuePrefix: prefix,
		queueLow:    prefix + "low",
		queueHigh:   prefix + "high",
		routeLow:    routeLow,
		routeHigh:   routeHigh,
	}, nil
}

func (br *PwWsBrokerRabbit) setProcessMessage(f processMessage) {
	br.processMessage = f
}

func (br *PwWsBrokerRabbit) configure() error {
	if nil == br.rabbit {
		return errors.New("you need to configure the rabbit client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := br.rabbit.ConnectAndWait(ctx); nil != err {
		return err
	}

	if _, err := br.rabbit.QueueDeclare(br.queuePrefix+"high", 60000); nil != err {
		return err
	}

	if _, err := br.rabbit.QueueDeclare(br.queuePrefix+"low", 60000); nil != err {
		return err
	}

	// bind routing keys to our queues
	if err := br.rabbit.QueueBind(br.queueLow, br.routeLow, rabbitmq.PlaneWatchExchange); nil != err {
		return err
	}

	if err := br.rabbit.QueueBind(br.queueHigh, br.routeHigh, rabbitmq.PlaneWatchExchange); nil != err {
		return err
	}
	return nil
}

func (br *PwWsBrokerRabbit) consume(exitChan chan bool, queue, what string) {
	ch, err := br.rabbit.Consume(queue, "pw_ws_broker"+what)
	if nil != err {
		log.Error().
			Err(err).
			Str("queue", queue).
			Str("what", what).
			Msg("Failed to consume")
		return
	}

	for msg := range ch {
		planeData, errDecode := export.FromProtobufBytes(msg.Body)
		if nil != errDecode {
			log.Debug().Err(err).Msg("did not understand msg")
			continue
		}
		br.processMessage(what, planeData)

	}
	log.Info().
		Str("queue", queue).
		Str("what", what).
		Msg("Finished Consuming")
	exitChan <- true
}

func (br *PwWsBrokerRabbit) consumeAll(exitChan chan bool) {
	go br.consume(exitChan, br.queueLow, "_low")
	go br.consume(exitChan, br.queueHigh, "_high")
}

func (br *PwWsBrokerRabbit) close() {
	if err := br.rabbit.QueueRemove(br.queueLow); nil != err {
		log.Error().Err(err).Str("Queue", br.queueLow).Msg("Removing Queue")
	}
	if err := br.rabbit.QueueRemove(br.queueHigh); nil != err {
		log.Error().Err(err).Str("Queue", br.queueHigh).Msg("Removing Queue")
	}

	if nil != br.rabbit {
		br.rabbit.Disconnect()
	}
}

func (br *PwWsBrokerRabbit) HealthCheckName() string {
	return br.rabbit.HealthCheckName()
}

func (br *PwWsBrokerRabbit) HealthCheck() bool {
	return br.rabbit.HealthCheck()
}
