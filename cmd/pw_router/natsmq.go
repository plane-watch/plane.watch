package main

import (
	"github.com/rs/zerolog/log"
	"plane.watch/lib/nats_io"
)

type (
	natsIoRouter struct {
		nats_io.Server
		doneChan chan bool
	}
)

func NewNatsIoRouter(url string) *natsIoRouter {
	if "" == url {
		return nil
	}
	nr := &natsIoRouter{
		doneChan: make(chan bool),
	}

	nr.SetUrl(url)
	return nr
}

func (nr *natsIoRouter) connect() error {
	var err error

	err = nr.Connect()
	if nil != err {
		log.Error().
			Err(err).
			Str("MQ", "nats.io").
			Msg("Unable to determine configuration from URL")
	}
	return err
}

func (nr *natsIoRouter) listen(subject string, incomingMessages chan []byte) error {
	ch, err := nr.Subscribe(subject)
	if nil != err {
		return err
	}

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					log.Error().Msg("failed to get message from nats.io nicely")
					return
				}

				incomingMessages <- msg.Data
			case <-nr.doneChan:
				close(ch)
				return
			}
		}
	}()
	return nil
}

func (nr *natsIoRouter) publish(subject string, msg []byte) error {
	err := nr.Publish(subject, msg)
	if nil != err {
		log.Warn().Err(err).Str("mq", "redis").Msg("Failed to send update")
		return err
	}
	return nil
}

func (nr *natsIoRouter) close() {
	nr.doneChan <- true
}
