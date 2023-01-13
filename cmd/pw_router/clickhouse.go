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
	//unPtr := func(s *string) string {
	//	if nil == s {
	//		return ""
	//	}
	//	return *s
	//}
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
			updates[updateId] = &chRow{
				New:             0,
				Removed:         0,
				Icao:            strconv.FormatUint(uint64(loc.Icao), 10),
				Lat:             loc.Lat,
				Lon:             loc.Lon,
				Heading:         loc.Heading,
				Velocity:        loc.Velocity,
				Altitude:        loc.Altitude,
				VerticalRate:    loc.VerticalRate,
				AltitudeUnits:   loc.AltitudeUnits.Describe(),
				CallSign:        loc.CallSign,
				FlightStatus:    loc.FlightStatus.Describe(),
				OnGround:        bool2int(loc.OnGround),
				AirframeType:    loc.AirframeType.Describe(),
				HasLocation:     bool2int(loc.HasLocation),
				HasHeading:      bool2int(loc.HasHeading),
				HasVerticalRate: bool2int(loc.HasVerticalRate),
				HasVelocity:     bool2int(loc.HasVelocity),
				SourceTag:       loc.SourceTag,
				Squawk:          loc.Squawk,
				Special:         loc.Special,
				TrackedSince:    loc.TrackedSince.AsTime().Format("2006-01-02 15:04:05.999999999"),
				LastMsg:         loc.LastMsg.AsTime().Format("2006-01-02 15:04:05.999999999"),
				FlagCode:        loc.FlagCode,
				Operator:        loc.Operator,
				RegisteredOwner: loc.RegisteredOwner,
				Registration:    loc.Registration,
				RouteCode:       loc.RouteCode,
				Serial:          loc.Serial,
				TileLocation:    loc.TileLocation,
				TypeCode:        loc.TypeCode,
			}

			updateId++
			if updateId >= max-1 {
				send()
			}
		}
	}
}
