package export

import (
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/tracker"
)

func NewPlaneLocation(plane *tracker.Plane, isNew, isRemoved bool, source string) PlaneLocationJSON {
	callSign := strings.TrimSpace(plane.FlightNumber())
	return PlaneLocationJSON{
		New:             isNew,
		Removed:         isRemoved,
		Icao:            plane.IcaoIdentifierStr(),
		Lat:             plane.Lat(),
		Lon:             plane.Lon(),
		Heading:         plane.Heading(),
		Altitude:        int(plane.Altitude()),
		VerticalRate:    plane.VerticalRate(),
		AltitudeUnits:   plane.AltitudeUnits(),
		Velocity:        plane.Velocity(),
		CallSign:        &callSign,
		FlightStatus:    plane.FlightStatus(),
		OnGround:        plane.OnGround(),
		Airframe:        plane.AirFrame(),
		AirframeType:    plane.AirFrameType(),
		Squawk:          plane.SquawkIdentityStr(),
		Special:         plane.Special(),
		AircraftWidth:   plane.AirFrameWidth(),
		AircraftLength:  plane.AirFrameLength(),
		Registration:    plane.Registration(),
		HasAltitude:     plane.HasAltitude(),
		HasLocation:     plane.HasLocation(),
		HasHeading:      plane.HasHeading(),
		HasVerticalRate: plane.HasVerticalRate(),
		HasVelocity:     plane.HasVelocity(),
		HasFlightStatus: plane.HasFlightStatus(),
		HasOnGround:     plane.HasOnGround(),
		SourceTag:       source,
		TileLocation:    plane.GridTileLocation(),
		LastMsg:         plane.LastSeen().UTC(),
		TrackedSince:    plane.TrackedSince().UTC(),
		SignalRssi:      plane.SignalLevel(),
		Updates: Updates{
			Location:     plane.LocationUpdatedAt().UTC(),
			Altitude:     plane.AltitudeUpdatedAt().UTC(),
			Velocity:     plane.VelocityUpdatedAt().UTC(),
			Heading:      plane.HeadingUpdatedAt().UTC(),
			VerticalRate: plane.VerticalRateUpdatedAt().UTC(),
			OnGround:     plane.OnGroundUpdatedAt().UTC(),
			FlightStatus: plane.FlightStatusUpdatedAt().UTC(),
			Special:      plane.SpecialUpdatedAt().UTC(),
			Squawk:       plane.SquawkUpdatedAt().UTC(),
		},
		sourceTagsMutex: &sync.Mutex{},
	}
}
func NewEmptyPlaneLocationJSON() *PlaneLocationJSON {
	return &PlaneLocationJSON{sourceTagsMutex: &sync.Mutex{}}
}

func (pl *PlaneLocationJSON) ToJSONBytes() ([]byte, error) {
	json := jsoniter.ConfigFastest

	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()

	jsonBuf, err := json.Marshal(pl)
	if nil != err {
		log.Error().Err(err).Msg("could not create json bytes for sending")
		return nil, err
	}
	return jsonBuf, nil
}

func (pl *PlaneLocationJSON) ToProtobuf(buf *PlaneAndLocationInfoMsg) error {
	if pl == nil {
		return errors.New("nil plane location")
	}

	icao, _ := strconv.ParseInt(pl.Icao, 16, 32)
	squawk, _ := strconv.ParseUint(pl.Squawk, 10, 32)
	altUnits := AltitudeUnits_FEET
	if pl.AltitudeUnits != "feet" {
		altUnits = AltitudeUnits_METRES
	}

	var segments []*RouteSegment
	if nil != pl.Segments {
		for _, s := range pl.Segments {
			segments = append(segments, &RouteSegment{
				Name:     s.Name,
				ICAOCode: s.ICAOCode,
			})
		}
	}
	buf.sourceTagsMutex = &sync.Mutex{}
	buf.Icao = uint32(icao)
	buf.Lat = pl.Lat
	buf.Lon = pl.Lon
	buf.Heading = pl.Heading
	buf.Velocity = pl.Velocity
	buf.Altitude = int32(pl.Altitude)
	buf.VerticalRate = int32(pl.VerticalRate)
	buf.AltitudeUnits = altUnits
	buf.FlightStatus = FlightStatus(flightStatusLookup[pl.FlightStatus])
	buf.OnGround = pl.OnGround
	buf.AirframeType = 0
	buf.HasAltitude = pl.HasAltitude
	buf.HasLocation = pl.HasLocation
	buf.HasHeading = pl.HasHeading
	buf.HasOnGround = pl.HasOnGround
	buf.HasFlightStatus = pl.HasFlightStatus
	buf.HasVerticalRate = pl.HasVerticalRate
	buf.HasVelocity = pl.HasVelocity
	buf.SourceTag = pl.SourceTag
	buf.Squawk = uint32(squawk)
	buf.Special = pl.Special
	buf.TileLocation = pl.TileLocation
	buf.SourceTags = pl.CloneSourceTags()
	buf.TrackedSince = timestamppb.New(pl.TrackedSince)
	buf.LastMsg = timestamppb.New(pl.LastMsg)
	buf.Updates = &FieldUpdates{
		Location:     timestamppb.New(pl.Updates.Location),
		Altitude:     timestamppb.New(pl.Updates.Altitude),
		Velocity:     timestamppb.New(pl.Updates.Velocity),
		Heading:      timestamppb.New(pl.Updates.Heading),
		OnGround:     timestamppb.New(pl.Updates.OnGround),
		VerticalRate: timestamppb.New(pl.Updates.VerticalRate),
		FlightStatus: timestamppb.New(pl.Updates.FlightStatus),
		Special:      timestamppb.New(pl.Updates.Special),
		Squawk:       timestamppb.New(pl.Updates.Squawk),
	}
	buf.SignalRssi = unPtr(pl.SignalRssi)
	buf.AircraftWidth = unPtr(pl.AircraftWidth)
	buf.AircraftLength = unPtr(pl.AircraftLength)
	buf.Registration = unPtr(pl.Registration)
	buf.TypeCode = unPtr(pl.TypeCode)
	buf.TypeCodeLong = unPtr(pl.TypeCodeLong)
	buf.Serial = unPtr(pl.Serial)
	buf.RegisteredOwner = unPtr(pl.RegisteredOwner)
	buf.COFAOwner = unPtr(pl.COFAOwner)
	buf.EngineType = unPtr(pl.EngineType)
	buf.FlagCode = unPtr(pl.FlagCode)
	buf.CallSign = unPtr(pl.CallSign)
	buf.Operator = unPtr(pl.Operator)
	buf.RouteCode = unPtr(pl.RouteCode)
	buf.Segments = segments
	return nil
}
