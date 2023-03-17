package main

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/nats_io"
)

type (
	ApiHandler struct {
		name     string
		idx      int
		log      zerolog.Logger
		subject  string
		server   *nats_io.Server
		exitChan chan bool
		handler  func(msg *nats.Msg)
	}
)

func (a *ApiHandler) configure(server *nats_io.Server) *ApiHandler {
	a.server = server
	a.exitChan = make(chan bool)
	return a
}

func (a *ApiHandler) exit() {
	a.exitChan <- true
}

func (a *ApiHandler) listen() {
	a.log = log.With().Str("handler", a.name).Int("ID", a.idx).Logger()
	a.log.Info().Str("subject", a.subject).Msg("Starting Handler")

	sub, err := a.server.SubscribeReply(a.subject, "atc-api", a.handler)

	if nil != err {
		a.log.Error().Err(err).Msg("Failed to API Handler")
		return
	}
	//a.log.Info().Msg("Awaiting...")

	<-a.exitChan

	// clean up
	if err = sub.Unsubscribe(); nil != err {
		a.log.Error().Err(err).Msg("Failed to unsubscribe")
	}
}
