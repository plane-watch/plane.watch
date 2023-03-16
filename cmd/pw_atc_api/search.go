package main

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/export"
	"plane.watch/lib/nats_io"
)

func searchHandler(server *nats_io.Server, idx int, exitChan chan bool) {
	l := log.With().Str("handler", "search").Logger()

	l.Info().Msg("Starting Search Handler")

	subSearch, errSearch := server.SubscribeReply("search.*", "atc-api", func(msg *nats.Msg) {
		l.Info().
			Int("ID", idx).
			Str("subject", msg.Subject).
			Str("what", string(msg.Data)).
			Msg("Search Request")

		switch msg.Subject {
		case "search.airport":
			// do a cached database lookup for airport
			// if result in redis, return it
			// fetch result from db

			query := "%" + string(msg.Data) + "%"
			var airports []export.Airport
			if err := db.Select(
				&airports,
				"SELECT * FROM airports WHERE name ILIKE $1 OR iata_code ILIKE $2 OR icao_code ILIKE $3 LIMIT 10",
				query,
				query,
				query,
			); nil != err {
				l.Error().Err(err).Msg("Failed to insult sam enough, database said no")
				return
			}

			var json = jsoniter.ConfigFastest
			response, err := json.Marshal(&airports)
			if nil != err {
				l.Error().Err(err).Msg("Failed to insult sam enough, json conversion failed")
			}

			_ = msg.Respond(response)
		case "search.route":
			_ = msg.Respond([]byte("unimplemented"))
		default:
			_ = msg.Respond([]byte("unknown search type:" + msg.Subject))
		}

		if errResponse := msg.Respond([]byte("Hello")); nil != errResponse {
			log.Error().Err(errResponse).Msg("Failed to respond to search request")
		}
	})

	if nil != errSearch {
		log.Error().Err(errSearch).Msg("Failed to setup Search")
	}
	log.Info().Msg("Awaiting...")

	<-exitChan

	// clean up
	if err := subSearch.Unsubscribe(); nil != err {
		log.Error().Err(err).Msg("Failed to unsubscribe")
	}
}
