package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/clickhouse"
	"plane.watch/lib/export"
	"strconv"
	"time"
)

type (
	DataStream struct {
		low, high chan *export.PlaneLocation
		chs       *clickhouse.Server
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
		TrackedSince    string
		LastMsg         string
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

func NewDataStreams(chs *clickhouse.Server) *DataStream {
	ds := &DataStream{
		low:  make(chan *export.PlaneLocation, 1000),
		high: make(chan *export.PlaneLocation, 2000),
		chs:  chs,
		log:  log.With().Str("section", "ch data stream").Logger(),
	}
	go ds.handleQueue(ds.low, "location_updates_low")
	go ds.handleQueue(ds.high, "location_updates_high")

	return ds
}

func (ds *DataStream) AddLow(frame *export.PlaneLocation) {
	ds.low <- frame
}

func (ds *DataStream) AddHigh(frame *export.PlaneLocation) {
	ds.high <- frame
}

// handleQueue single threadedly accumulates and sends data to clickhouse for the given queue/table
func (ds *DataStream) handleQueue(q chan *export.PlaneLocation, table string) {
	ticker := time.NewTicker(time.Second)
	max := 50_000
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
				TrackedSince:    loc.TrackedSince.UTC().Format("2006-01-02 15:04:05.999999999"),
				LastMsg:         loc.LastMsg.UTC().Format("2006-01-02 15:04:05.999999999"),
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
