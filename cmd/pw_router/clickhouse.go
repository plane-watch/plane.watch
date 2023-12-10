package main

import (
	"github.com/paulmach/orb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
	"plane.watch/lib/clickhouse"
	"plane.watch/lib/export"
	"strconv"
	"time"
)

type (
	DataStream struct {
		low, high chan *export.PlaneAndLocationInfoMsg
		chs       *clickhouse.Server
		log       zerolog.Logger
	}

	chRow struct {
		Icao            string
		LatLon          orb.Point
		Lat             float64
		Lon             float64
		Heading         float64
		Velocity        float64
		Altitude        int32
		VerticalRate    int32
		AltitudeUnits   string
		CallSign        string
		FlightStatus    string
		OnGround        bool
		Airframe        string
		AirframeType    string
		HasLocation     bool
		HasHeading      bool
		HasVerticalRate bool
		HasVelocity     bool
		SourceTags      map[string]uint32
		Squawk          uint32
		Special         string
		TrackedSince    time.Time
		LastMsg         time.Time
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
		low:  make(chan *export.PlaneAndLocationInfoMsg, 1000),
		high: make(chan *export.PlaneAndLocationInfoMsg, 2000),
		chs:  chs,
		log:  log.With().Str("section", "ch data stream").Logger(),
	}
	go ds.handleQueue(ds.low, "location_updates_low")
	go ds.handleQueue(ds.high, "location_updates_high")

	return ds
}

func (ds *DataStream) AddLow(frame *export.PlaneAndLocationInfoMsg) {
	if frame.SourceTag != "repeat" {
		ds.low <- frame
	}
}

func (ds *DataStream) AddHigh(frame *export.PlaneAndLocationInfoMsg) {
	if frame.SourceTag != "repeat" {
		ds.high <- frame
	}
}

// handleQueue single threadedly accumulates and sends data to clickhouse for the given queue/table
func (ds *DataStream) handleQueue(q chan *export.PlaneAndLocationInfoMsg, table string) {
	ticker := time.NewTicker(time.Second)
	maxNumItems := 50_000
	updates := make([]any, maxNumItems)
	updateID := 0
	send := func() {
		ds.log.Debug().Int("num", updateID).Msg("Sending Batch To Clickhouse")
		if err := ds.chs.Inserts(table, updates, updateID); nil != err {
			ds.log.Err(err).Msg("Did not save location update to clickhouse")
		}
		updateID = 0
	}
	tags := make(map[string]uint32, 5)
	for {
		select {
		case <-ticker.C:
			send()
		case loc := <-q:
			maps.Clear(tags)
			squawk, _ := strconv.ParseUint(loc.SquawkStr(), 10, 32)
			updates[updateID] = &chRow{
				Icao:            loc.IcaoStr(),
				LatLon:          orb.Point{loc.Lat, loc.Lon},
				Heading:         loc.Heading,
				Velocity:        loc.Velocity,
				Altitude:        loc.Altitude,
				VerticalRate:    loc.VerticalRate,
				AltitudeUnits:   loc.AltitudeUnits.Describe(),
				CallSign:        loc.CallSign,
				FlightStatus:    loc.FlightStatus.Describe(),
				OnGround:        loc.OnGround,
				Airframe:        loc.AirframeType.Describe(),
				AirframeType:    loc.AirframeType.Code(),
				HasLocation:     loc.HasLocation,
				HasHeading:      loc.HasHeading,
				HasVerticalRate: loc.HasVerticalRate,
				HasVelocity:     loc.HasVelocity,
				SourceTags:      loc.PrepareSourceTags(tags),
				Squawk:          uint32(squawk),
				Special:         loc.Special,
				TrackedSince:    loc.TrackedSince.AsTime().UTC(),
				LastMsg:         loc.LastMsg.AsTime().UTC(),
				FlagCode:        loc.FlagCode,
				Operator:        loc.Operator,
				RegisteredOwner: loc.RegisteredOwner,
				Registration:    loc.Registration,
				RouteCode:       loc.RouteCode,
				Serial:          loc.Serial,
				TileLocation:    loc.TileLocation,
				TypeCode:        loc.TypeCode,
			}

			updateID++
			if updateID >= maxNumItems-1 {
				send()
			}
		}
	}
}
