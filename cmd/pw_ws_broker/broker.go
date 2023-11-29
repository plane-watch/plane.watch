package main

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/nats_io"
	"syscall"
	"time"
)

type (
	PwWsBroker struct {
		input    source
		driver   PwWsWebDriver
		exitChan chan bool
	}

	PwWsWebDriver interface {
		configureWeb() error
		listenAndServe(chan bool)
		closeWeb() error
		SendLocationUpdate(highLow string, msg []byte)
		monitoring.HealthCheck
	}

	source interface {
		configure() error
		setProcessMessage(processMessage)
		consumeAll(chan bool)
		close()
		monitoring.HealthCheck
	}
	processMessage func(highLow string, msgData []byte)
)

func NewPlaneWatchWebSocketJSONBroker(
	input source,
	natsRPC *nats_io.Server,
	httpAddr, cert, certKey string,
	serveTestWeb bool,
	sendTickDuration time.Duration,
) (*PwWsBroker, error) {
	return &PwWsBroker{
		input: input,
		driver: &PwWsBrokerWebJSON{
			natsRPC:          natsRPC,
			Addr:             httpAddr,
			ServeTest:        serveTestWeb,
			cert:             cert,
			certKey:          certKey,
			sendTickDuration: sendTickDuration,
			serveMux:         &http.ServeMux{},
		},
		exitChan: make(chan bool),
	}, nil
}

func NewPlaneWatchWebSocketProtobufBroker(
	input source,
	natsRPC *nats_io.Server,
	httpAddr, cert, certKey string,
	serveTestWeb bool,
	sendTickDuration time.Duration,
) (*PwWsBroker, error) {
	return &PwWsBroker{
		input: input,
		driver: &PwWsBrokerWebProtobuf{
			natsRPC:          natsRPC,
			Addr:             httpAddr,
			ServeTest:        serveTestWeb,
			cert:             cert,
			certKey:          certKey,
			sendTickDuration: sendTickDuration,
			serveMux:         &http.ServeMux{},
		},
		exitChan: make(chan bool),
	}, nil
}

func (b *PwWsBroker) Setup() error {
	if err := b.input.configure(); nil != err {
		return err
	}
	if err := b.driver.configureWeb(); nil != err {
		return err
	}

	b.input.setProcessMessage(func(highLow string, msg []byte) {
		prometheusIncomingMessages.WithLabelValues(highLow).Inc()
		b.driver.SendLocationUpdate(highLow, msg)
	})

	monitoring.AddHealthCheck(b.input)
	monitoring.AddHealthCheck(b.driver)
	return nil
}

func (b *PwWsBroker) Run() {
	go b.driver.listenAndServe(b.exitChan)
	go b.input.consumeAll(b.exitChan)
}

func (b *PwWsBroker) Wait() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	select {
	case <-b.exitChan:
		log.Debug().Msg("We are exiting")
	case <-sc:
		log.Debug().Msg("Kill Signal Received")
	}
}

func (b *PwWsBroker) Close() {
	if err := b.driver.closeWeb(); nil != err {
		log.Error().Err(err).Msg("Failed to close web server cleanly")
	}
	b.input.close()
}
