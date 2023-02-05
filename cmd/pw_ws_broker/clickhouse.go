package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/clickhouse"
	"plane.watch/lib/ws_protocol"
	"time"
)

type (
	ClickHouseData struct {
		server *clickhouse.Server
	}
)

var (
	GlobalClickHouseData *ClickHouseData
)

func NewClickHouseData(url string) (*ClickHouseData, error) {
	s, err := clickhouse.New(url)
	if nil != err {
		return nil, err
	}
	return &ClickHouseData{
		server: s,
	}, nil
}

func (chd *ClickHouseData) PlaneLocationHistory(icao, callSign string) []ws_protocol.LocationHistory {
	history := make([]ws_protocol.LocationHistory, 0, 2200) // if we increase from 6 hours, increase our initial allocation
	query := `WITH t AS (
SELECT *
FROM location_updates_low
WHERE Icao = '` + icao + `' AND CallSign = '` + callSign + `' AND HasLocation = 1 AND TileLocation != 'tileUnknown'
  AND LastMsg > timestamp_sub(hour, 12, now())
), t_over AS (
    SELECT *, ROW_NUMBER() OVER(PARTITION BY toInt64(toInt64(LastMsg)/10) ORDER BY LastMsg) AS N FROM t
)

SELECT Lat, Lon, Velocity, Altitude, Heading FROM t_over where N=1 ORDER BY LastMsg`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := chd.server.Select(ctx, &history, query, icao, callSign); nil != err {
		log.Error().Err(err).Str("query", query).Msg("Failed to get aircraft location history")
		return history
	}
	log.Debug().Int("num items", len(history)).Str("query", query).Send()

	return history
}
