package export

import (
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
