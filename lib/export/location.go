package export

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"
	"plane.watch/lib/tracker"
)

var airframeTypeLookup = map[string]AirframeType{
	`0/0`: AirframeType_CODE_00_NO_ADSB,
	`0/1`: AirframeType_CODE_01_LIGHT,
	`0/2`: AirframeType_CODE_02_SMALL,
	`0/3`: AirframeType_CODE_03_LARGE,
	`0/4`: AirframeType_CODE_04_HIGH_VORTEX,
	`0/5`: AirframeType_CODE_05_HEAVY,
	`0/6`: AirframeType_CODE_06_HIGH_PERFORMANCE,
	`0/7`: AirframeType_CODE_07_ROTORCRAFT,

	`1/0`: AirframeType_CODE_10_NO_ADSB,
	`1/1`: AirframeType_CODE_11_GLIDER_SAILPLANE,
	`1/2`: AirframeType_CODE_12_LIGHTER_THAN_AIR,
	`1/3`: AirframeType_CODE_13_PARACHUTIST,
	`1/4`: AirframeType_CODE_14_ULTRALIGHT,
	`1/5`: AirframeType_CODE_15_RESERVED,
	`1/6`: AirframeType_CODE_16_UAV,
	`1/7`: AirframeType_CODE_17_TRANS_ATMO,

	`2/0`: AirframeType_CODE_20_NO_ADSB,
	`2/1`: AirframeType_CODE_21_SURFACE_EMERGENCY_VEHICLE,
	`2/2`: AirframeType_CODE_22_SURFACE_VEHICLE,
	`2/3`: AirframeType_CODE_23_POINT_OBSTACLE,
	`2/4`: AirframeType_CODE_24_CLUSTER_OBSTACLE,
	`2/5`: AirframeType_CODE_25_LINE_OBSTACLE,
	`2/6`: AirframeType_CODE_26_RESERVED,
	`2/7`: AirframeType_CODE_27_RESERVED,

	`3/0`: AirframeType_CODE_30_RESERVED,
	`3/1`: AirframeType_CODE_31_RESERVED,
	`3/2`: AirframeType_CODE_32_RESERVED,
	`3/3`: AirframeType_CODE_33_RESERVED,
	`3/4`: AirframeType_CODE_34_RESERVED,
	`3/5`: AirframeType_CODE_35_RESERVED,
	`3/6`: AirframeType_CODE_36_RESERVED,
	`3/7`: AirframeType_CODE_37_RESERVED,
}

func ToAirframeType(code string) AirframeType {
	if v, ok := airframeTypeLookup[code]; ok {
		return v
	}
	return AirframeType_UNKNOWN
}

func NewPlaneLocation(plane *tracker.Plane, source string) PlaneLocation {
	var altitudeUnits AltitudeUnits

	if "feet" == plane.AltitudeUnits() {
		altitudeUnits = AltitudeUnits_FEET
	} else {
		altitudeUnits = AltitudeUnits_METRES
	}

	return PlaneLocation{
		PlaneLocationPB: PlaneLocationPB{
			Icao:            plane.IcaoIdentifier(),
			Lat:             plane.Lat(),
			Lon:             plane.Lon(),
			Heading:         plane.Heading(),
			Altitude:        plane.Altitude(),
			VerticalRate:    plane.VerticalRate(),
			AltitudeUnits:   altitudeUnits,
			Velocity:        plane.Velocity(),
			CallSign:        plane.FlightNumber(),
			FlightStatus:    FlightStatus(plane.FlightStatusId()),
			OnGround:        plane.OnGround(),
			AirframeType:    ToAirframeType(plane.AirFrameType()),
			Squawk:          plane.SquawkIdentity(),
			Special:         plane.Special(),
			AircraftWidth:   unPtr(plane.AirFrameWidth()),
			AircraftLength:  unPtr(plane.AirFrameLength()),
			Registration:    unPtr(plane.Registration()),
			HasAltitude:     plane.HasAltitude(),
			HasLocation:     plane.HasLocation(),
			HasHeading:      plane.HasHeading(),
			HasVerticalRate: plane.HasVerticalRate(),
			HasVelocity:     plane.HasVelocity(),
			HasFlightStatus: plane.HasFlightStatus(),
			HasOnGround:     plane.HasOnGround(),
			SourceTag:       source,
			TileLocation:    plane.GridTileLocation(),
			LastMsg:         timestamppb.New(plane.LastSeen().UTC()),
			TrackedSince:    timestamppb.New(plane.TrackedSince().UTC()),
			SignalRssi:      unPtr(plane.SignalLevel()),
			Updates: &FieldUpdates{
				Location:     timestamppb.New(plane.LocationUpdatedAt().UTC()),
				Altitude:     timestamppb.New(plane.AltitudeUpdatedAt().UTC()),
				Velocity:     timestamppb.New(plane.VelocityUpdatedAt().UTC()),
				Heading:      timestamppb.New(plane.HeadingUpdatedAt().UTC()),
				VerticalRate: timestamppb.New(plane.VerticalRateUpdatedAt().UTC()),
				OnGround:     timestamppb.New(plane.OnGroundUpdatedAt().UTC()),
				FlightStatus: timestamppb.New(plane.FlightStatusUpdatedAt().UTC()),
				Special:      timestamppb.New(plane.SpecialUpdatedAt().UTC()),
				Squawk:       timestamppb.New(plane.SquawkUpdatedAt().UTC()),
			},
		},
	}
}

func (pl *PlaneLocation) ToJsonBytes() ([]byte, error) {
	json := jsoniter.ConfigFastest
	jsonBuf, err := json.Marshal(pl)
	if nil != err {
		log.Error().Err(err).Msg("could not create json bytes for sending")
		return nil, err
	} else {
		return jsonBuf, nil
	}

}
