package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"plane.watch/lib/export"
	"time"
)

type (
	SearchApiHandler struct {
		ApiHandler
	}
)

func newSearchApi(idx int) *SearchApiHandler {
	api := SearchApiHandler{
		ApiHandler: ApiHandler{
			idx:     idx,
			name:    "search",
			subject: "v1.search.*",
		},
	}
	api.handler = api.searchHandler

	return &api
}

func (sa *SearchApiHandler) searchHandler(msg *nats.Msg) {
	// capture how long we spend searching
	tStart := time.Now()
	defer func() {
		d := time.Since(tStart)
		prometheusCounterSearchSummary.Observe(float64(d.Microseconds()))
	}()
	prometheusCounterSearch.Inc()
	sa.log.Info().
		Str("subject", msg.Subject).
		Str("what", string(msg.Data)).
		Msg("Search Request")

	var respondErr error

	switch msg.Subject {
	case export.NatsAPISearchAirportV1:
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
			sa.log.Error().Err(err).Msg("Failed to insult sam enough, database said no")
			return
		}

		var json = jsoniter.ConfigFastest
		response, err := json.Marshal(&airports)
		if nil != err {
			sa.log.Error().Err(err).Msg("Failed to insult sam enough, json conversion failed")
		}

		respondErr = msg.Respond(response)
	case export.NatsAPISearchRouteV1:
		respondErr = msg.Respond([]byte("unimplemented"))
	default:
		respondErr = msg.Respond([]byte(fmt.Sprintf(ErrUnsupportedResponse, msg.Subject)))
	}

	if nil != respondErr {
		sa.log.Error().Err(respondErr).Msg("Failed sending reply")
		_ = msg.Respond([]byte(fmt.Sprintf(ErrRequestFailed, respondErr)))
	}
}
