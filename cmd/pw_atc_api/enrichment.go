package main

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"plane.watch/lib/export"
	"strings"
	"time"
)

type (
	EnrichmentApiHandler struct {
		ApiHandler
	}

	DbOperator struct {
		IcaoCode           *string `db:"icao_code"`
		IataCode           *string `db:"iata_code"`
		Name               *string `db:"name"`
		PositioningPattern *string `db:"positioning_pattern"`
		CharterPattern     *string `db:"charter_pattern"`
	}

	DbRoute struct {
		Id         int64  `db:"id"`
		CallSign   string `db:"callsign"`
		OperatorId int64  `db:"operator_id"`
	}

	DbRouteSegments struct {
		Name     string `db:"name"`
		IcaoCode string `db:"icao_code"`
	}
)

func newEnrichmentApi(idx int) *EnrichmentApiHandler {
	api := EnrichmentApiHandler{
		ApiHandler: ApiHandler{
			idx:     idx,
			name:    "enrichment",
			subject: "v1.enrich.*",
		},
	}
	api.handler = api.enrichHandler

	return &api
}

func (sa *EnrichmentApiHandler) enrichHandler(msg *nats.Msg) {
	// capture how long we spend searching
	tStart := time.Now()
	defer func() {
		d := time.Now().Sub(tStart)
		prometheusCounterEnrichSummary.Observe(float64(d.Microseconds()))
	}()
	prometheusCounterEnrich.Inc()
	what := string(msg.Data)
	sa.log.Info().
		Str("subject", msg.Subject).
		Str("what", what).
		Msg("enrichment request")

	var respondErr error
	var buf []byte

	switch msg.Subject {
	case export.NatsApiEnrichAircraftV1:
		icao := strings.ToUpper(what)
		aircraft := export.Aircraft{}
		respondErr = db.Get(&aircraft, "SELECT * FROM aircraft WHERE icao_code = ?", icao)
		if nil != respondErr {
			json := jsoniter.ConfigFastest
			buf, respondErr = json.Marshal(aircraft)
			if nil == respondErr {
				respondErr = msg.Respond(buf)
			}
		}
	case export.NatsApiEnrichRouteV1:
		response := export.RouteResponse{}
		route := DbRoute{}
		callSign := strings.ToUpper(what)
		respondErr = db.Get(&route, "SELECT id,operator_id,callsign from routes WHERE callsign = $1 LIMIT 1", callSign)
		if nil == respondErr {
			// we have a route!
			response.Route.CallSign = &route.CallSign

			operator := DbOperator{}
			if err := db.Get(&operator, "SELECT name FROM operators WHERE id = $1", route.OperatorId); nil != err {
				sa.log.Error().Err(err).Send()
			}
			response.Route.Operator = operator.Name

			var segments []DbRouteSegments
			var routeStr string
			_ = db.Select(&segments, `SELECT a.name,a.icao_code FROM route_segments rs left join airports a on a.id = rs.airport_id  WHERE route_id=$1 order by rs."order"`, route.Id)
			for _, segment := range segments {
				routeStr += segment.IcaoCode + "-"
				response.Route.Segments = append(response.Route.Segments, export.Segment{
					Name:     segment.Name,
					ICAOCode: segment.IcaoCode,
				})
			}
			routeStr = strings.Trim(routeStr, "-")
			response.Route.RouteCode = &routeStr
		}
		if nil == respondErr {
			json := jsoniter.ConfigFastest
			buf, respondErr = json.Marshal(response)
			if nil == respondErr {
				respondErr = msg.Respond(buf)
			}
		}
	default:
		respondErr = msg.Respond([]byte("unknown enrichment type:" + msg.Subject))
	}

	if nil != respondErr {
		sa.log.Error().Err(respondErr).Msg("Failed sending reply")
		_ = msg.Respond([]byte(""))
	}
}
