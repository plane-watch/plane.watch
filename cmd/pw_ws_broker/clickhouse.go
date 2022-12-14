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
	history := make([]ws_protocol.LocationHistory, 0, 1000)
	query := "SELECT DISTINCT Lat, Lon, Velocity, Altitude, Heading " +
		"FROM location_updates_low " +
		"WHERE Icao = $1 AND CallSign = $2 AND HasLocation = 1 AND TileLocation != 'tileUnknown' " +
		"AND LastMsg > toStartOfInterval(NOW(), INTERVAL 6 HOUR) " +
		"ORDER BY LastMsg"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := chd.server.Select(ctx, &history, query, icao, callSign); nil != err {
		log.Error().Err(err).Str("query", query).Msg("Failed to get aircraft location history")
		return history
	}
	log.Debug().Int("num items", len(history)).Str("query", query).Send()

	return history
}
