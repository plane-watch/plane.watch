package main

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/url"
	"plane.watch/lib/export"
	"strconv"
	"time"
)

type (
	ClickHouseServer struct {
		// expecting something like clickhouse://user:pass@127.0.0.1:9000
		connUrl string
		conn    driver.Conn

		connected bool

		log zerolog.Logger
	}

	dataStream struct {
		low, high chan *export.PlaneLocation
		chs       *ClickHouseServer
		log       zerolog.Logger
	}

	chRow struct {
		New             uint8
		Removed         uint8
		Icao            string
		Lat             float64
		Lon             float64
		Heading         float64
		Velocity        float64
		Altitude        int32
		VerticalRate    int32
		AltitudeUnits   string
		CallSign        string
		FlightStatus    string
		OnGround        uint8
		Airframe        string
		AirframeType    string
		HasLocation     uint8
		HasHeading      uint8
		HasVerticalRate uint8
		HasVelocity     uint8
		SourceTag       string
		Squawk          uint32
		Special         string
		TrackedSince    int64
		LastMsg         int64
		FlagCode        string
		Operator        string
		RegisteredOwner string
		Registration    string
		RouteCode       string
		Serial          string
		TileLocation    string
		TypeCode        string
	}
)

func NewClickHouse(url string) (*ClickHouseServer, error) {
	chs := &ClickHouseServer{
		connUrl: url,
		log:     log.With().Str("section", "clickhouse").Logger(),
	}
	var err error
	for i := 0; i < 5; i++ {
		err = chs.Connect()

		if nil == err {
			log.Info().Str("ClickHouse", url).Msg("Connected")
			return chs, nil
		}
	}
	log.Error().Err(err).Str("ClickHouse", url).Msg("Failed to Connect")
	return nil, err
}

func (chs *ClickHouseServer) Connect() error {
	var err error

	urlParts, err := url.Parse(chs.connUrl)
	if nil != err {
		return err
	}
	username := urlParts.User.Username()
	password, _ := urlParts.User.Password()
	database := "plane_watch"
	if "" != urlParts.Path && "/" != urlParts.Path {
		database = urlParts.Path
	}
	urlParts.User = nil
	urlParts.Path = ""

	chs.log.Info().Str("URL", urlParts.String()).Msg("Attempting to connect to")

	chs.conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{urlParts.Host},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	})
	if nil != err {
		return err
	}

	chs.connected = true
	return nil
}

func (chs *ClickHouseServer) Inserts(table string, d []any, max int) error {
	if chs.log.Trace().Enabled() {
		chs.log.Trace().Str("table", table).Interface("data", d).Msg("insert")
	}
	t := time.Now()
	ctx := context.Background()
	batch, err := chs.conn.PrepareBatch(ctx, "INSERT INTO "+table)
	if nil != err {
		return err
	}
	for i := 0; i < max; i++ {
		err = batch.AppendStruct(d[i])
		if nil != err {
			chs.log.Error().Err(err).Msg("Did not insert data")
			return err
		}
	}
	chs.log.Debug().
		TimeDiff("Time Taken", time.Now(), t).
		Str("table", table).
		Int("Num Rows", max).
		Msg("Insert Batch")
	return batch.Send()
}

func NewDataStreams(chs *ClickHouseServer) *dataStream {
	ds := &dataStream{
		low:  make(chan *export.PlaneLocation, 1000),
		high: make(chan *export.PlaneLocation, 2000),
		chs:  chs,
		log:  log.With().Str("section", "ch data stream").Logger(),
	}
	go ds.handleQueue(ds.low, "location_updates_low")
	go ds.handleQueue(ds.high, "location_updates_high")

	return ds
}

func (ds *dataStream) AddLow(frame *export.PlaneLocation) {
	ds.low <- frame
}

func (ds *dataStream) AddHigh(frame *export.PlaneLocation) {
	ds.high <- frame
}

func (ds *dataStream) handleQueue(q chan *export.PlaneLocation, table string) {
	ticker := time.NewTicker(time.Second)
	max := 50000
	updates := make([]any, max)
	updateId := 0
	bool2int := func(x bool) uint8 {
		if x {
			return 1
		}
		return 0
	}
	unPtr := func(s *string) string {
		if nil == s {
			return ""
		}
		return *s
	}
	send := func() {
		ds.log.Debug().Int("num", updateId).Msg("Sending Batch To Clickhouse")
		if err := ds.chs.Inserts(table, updates, updateId); nil != err {
			ds.log.Err(err).Msg("Did not save location update to clickhouse")
		}
		updateId = 0
	}
	for {
		select {
		case <-ticker.C:
			send()
		case loc := <-q:
			squawk, _ := strconv.ParseUint(loc.Squawk, 10, 32)
			updates[updateId] = &chRow{
				New:             bool2int(loc.New),
				Removed:         bool2int(loc.Removed),
				Icao:            loc.Icao,
				Lat:             loc.Lat,
				Lon:             loc.Lon,
				Heading:         loc.Heading,
				Velocity:        loc.Velocity,
				Altitude:        int32(loc.Altitude),
				VerticalRate:    int32(loc.VerticalRate),
				AltitudeUnits:   loc.AltitudeUnits,
				CallSign:        unPtr(loc.CallSign),
				FlightStatus:    loc.FlightStatus,
				OnGround:        bool2int(loc.OnGround),
				Airframe:        loc.Airframe,
				AirframeType:    loc.AirframeType,
				HasLocation:     bool2int(loc.HasLocation),
				HasHeading:      bool2int(loc.HasHeading),
				HasVerticalRate: bool2int(loc.HasVerticalRate),
				HasVelocity:     bool2int(loc.HasVelocity),
				SourceTag:       loc.SourceTag,
				Squawk:          uint32(squawk),
				Special:         loc.Special,
				TrackedSince:    loc.TrackedSince.UTC().UnixNano(),
				LastMsg:         loc.LastMsg.UTC().UnixNano(),
				FlagCode:        unPtr(loc.FlagCode),
				Operator:        unPtr(loc.Operator),
				RegisteredOwner: unPtr(loc.RegisteredOwner),
				Registration:    unPtr(loc.Registration),
				RouteCode:       unPtr(loc.RouteCode),
				Serial:          unPtr(loc.Serial),
				TileLocation:    loc.TileLocation,
				TypeCode:        unPtr(loc.TypeCode),
			}

			updateId++
			if updateId >= max-1 {
				send()
			}
		}
	}
}
