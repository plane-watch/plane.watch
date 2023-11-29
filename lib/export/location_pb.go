package export

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"plane.watch/lib/tracker"
	"strconv"
	"sync"
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
var airframeTypeReverseLookup = map[AirframeType]string{
	AirframeType_CODE_00_NO_ADSB:          `0/0`,
	AirframeType_CODE_01_LIGHT:            `0/1`,
	AirframeType_CODE_02_SMALL:            `0/2`,
	AirframeType_CODE_03_LARGE:            `0/3`,
	AirframeType_CODE_04_HIGH_VORTEX:      `0/4`,
	AirframeType_CODE_05_HEAVY:            `0/5`,
	AirframeType_CODE_06_HIGH_PERFORMANCE: `0/6`,
	AirframeType_CODE_07_ROTORCRAFT:       `0/7`,

	AirframeType_CODE_10_NO_ADSB:          `1/0`,
	AirframeType_CODE_11_GLIDER_SAILPLANE: `1/1`,
	AirframeType_CODE_12_LIGHTER_THAN_AIR: `1/2`,
	AirframeType_CODE_13_PARACHUTIST:      `1/3`,
	AirframeType_CODE_14_ULTRALIGHT:       `1/4`,
	AirframeType_CODE_15_RESERVED:         `1/5`,
	AirframeType_CODE_16_UAV:              `1/6`,
	AirframeType_CODE_17_TRANS_ATMO:       `1/7`,

	AirframeType_CODE_20_NO_ADSB:                   `2/0`,
	AirframeType_CODE_21_SURFACE_EMERGENCY_VEHICLE: `2/1`,
	AirframeType_CODE_22_SURFACE_VEHICLE:           `2/2`,
	AirframeType_CODE_23_POINT_OBSTACLE:            `2/3`,
	AirframeType_CODE_24_CLUSTER_OBSTACLE:          `2/4`,
	AirframeType_CODE_25_LINE_OBSTACLE:             `2/5`,
	AirframeType_CODE_26_RESERVED:                  `2/6`,
	AirframeType_CODE_27_RESERVED:                  `2/7`,

	AirframeType_CODE_30_RESERVED: `3/0`,
	AirframeType_CODE_31_RESERVED: `3/1`,
	AirframeType_CODE_32_RESERVED: `3/2`,
	AirframeType_CODE_33_RESERVED: `3/3`,
	AirframeType_CODE_34_RESERVED: `3/4`,
	AirframeType_CODE_35_RESERVED: `3/5`,
	AirframeType_CODE_36_RESERVED: `3/6`,
	AirframeType_CODE_37_RESERVED: `3/7`,
}

var flightStatusLookup = map[string]int32{
	"Normal, Airborne":      0,
	"Normal, On the ground": 1,
	"ALERT, Airborne":       2,
	"ALERT, On the ground":  3,
	"ALERT, special Position Identification. Airborne or Ground":  4,
	"Normal, special Position Identification. Airborne or Ground": 5,
	"Value 6 is not assigned":                                     6,
	"Value 7 is not assigned":                                     7,
}

func ToAirframeType(code string) AirframeType {
	if v, ok := airframeTypeLookup[code]; ok {
		return v
	}
	return AirframeType_UNKNOWN
}

var locationPool sync.Pool

func init() {
	locationPool = sync.Pool{
		New: func() any {
			return NewPlaneAndLocationInfoMsg()
		},
	}
}

func NewPlaneInfo(plane *tracker.Plane, source string) *PlaneAndLocationInfoMsg {
	var altitudeUnits AltitudeUnits

	if plane.AltitudeUnits() == "feet" {
		altitudeUnits = AltitudeUnits_FEET
	} else {
		altitudeUnits = AltitudeUnits_METRES
	}

	msg := locationPool.Get().(*PlaneAndLocationInfoMsg)

	msg.Icao = plane.IcaoIdentifier()
	msg.Lat = plane.Lat()
	msg.Lon = plane.Lon()
	msg.Heading = plane.Heading()
	msg.Altitude = plane.Altitude()
	msg.VerticalRate = int32(plane.VerticalRate())
	msg.AltitudeUnits = altitudeUnits
	msg.Velocity = plane.Velocity()
	msg.CallSign = plane.FlightNumber()
	msg.FlightStatus = FlightStatus(plane.FlightStatusId())
	msg.OnGround = plane.OnGround()
	msg.AirframeType = ToAirframeType(plane.AirFrameType())
	msg.Squawk = plane.SquawkIdentity()
	msg.Special = plane.Special()
	msg.AircraftWidth = unPtr(plane.AirFrameWidth())
	msg.AircraftLength = unPtr(plane.AirFrameLength())
	msg.Registration = unPtr(plane.Registration())
	msg.HasAltitude = plane.HasAltitude()
	msg.HasLocation = plane.HasLocation()
	msg.HasHeading = plane.HasHeading()
	msg.HasVerticalRate = plane.HasVerticalRate()
	msg.HasVelocity = plane.HasVelocity()
	msg.HasFlightStatus = plane.HasFlightStatus()
	msg.HasOnGround = plane.HasOnGround()
	msg.SourceTag = source
	msg.TileLocation = plane.GridTileLocation()
	msg.LastMsg = timestamppb.New(plane.LastSeen().UTC())
	msg.TrackedSince = timestamppb.New(plane.TrackedSince().UTC())
	msg.SignalRssi = unPtr(plane.SignalLevel())
	msg.Updates.Location = timestamppb.New(plane.LocationUpdatedAt().UTC())
	msg.Updates.Altitude = timestamppb.New(plane.AltitudeUpdatedAt().UTC())
	msg.Updates.Velocity = timestamppb.New(plane.VelocityUpdatedAt().UTC())
	msg.Updates.Heading = timestamppb.New(plane.HeadingUpdatedAt().UTC())
	msg.Updates.VerticalRate = timestamppb.New(plane.VerticalRateUpdatedAt().UTC())
	msg.Updates.OnGround = timestamppb.New(plane.OnGroundUpdatedAt().UTC())
	msg.Updates.FlightStatus = timestamppb.New(plane.FlightStatusUpdatedAt().UTC())
	msg.Updates.Special = timestamppb.New(plane.SpecialUpdatedAt().UTC())
	msg.Updates.Squawk = timestamppb.New(plane.SquawkUpdatedAt().UTC())
	msg.sourceTagsMutex = &sync.Mutex{}
	return msg
}

func Release(msg *PlaneAndLocationInfoMsg) {
	locationPool.Put(msg)
}

func (pl *PlaneAndLocationInfoMsg) ToJSONBytes() ([]byte, error) {
	// convert back to a PlaneLocationJSON object and convert to JSON
	return pl.ToPlaneLocation().ToJSONBytes()
}

func FromJSONBytes(jsonBytes []byte) (*PlaneAndLocationInfoMsg, error) {
	var msg PlaneAndLocationInfoMsg
	msg.sourceTagsMutex = &sync.Mutex{}
	err := protojson.Unmarshal(jsonBytes, msg.PlaneAndLocationInfo)
	return &msg, err
}

func (pl *PlaneAndLocationInfoMsg) ToProtobufBytes() ([]byte, error) {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()
	protobufBytes, err := proto.Marshal(pl.PlaneAndLocationInfo)
	if nil != err {
		log.Error().Err(err).Msg("could not create protobuf bytes for sending")
		return nil, err
	}
	return protobufBytes, nil
}

func (pl *PlaneAndLocationInfoMsg) ToPlaneLocation() *PlaneLocationJSON {
	segments := make([]Segment, len(pl.Segments))
	for i := 0; i < len(pl.Segments); i++ {
		segments[i].Name = pl.Segments[i].Name
		segments[i].ICAOCode = pl.Segments[i].ICAOCode
	}
	return &PlaneLocationJSON{
		Icao:            pl.IcaoStr(),
		Lat:             pl.Lat,
		Lon:             pl.Lon,
		Heading:         pl.Heading,
		Velocity:        pl.Velocity,
		Altitude:        int(pl.Altitude),
		VerticalRate:    int(pl.VerticalRate),
		New:             false,
		Removed:         false,
		OnGround:        pl.OnGround,
		HasAltitude:     pl.HasAltitude,
		HasLocation:     pl.HasLocation,
		HasHeading:      pl.HasHeading,
		HasOnGround:     pl.HasOnGround,
		HasFlightStatus: pl.HasFlightStatus,
		HasVerticalRate: pl.HasVerticalRate,
		HasVelocity:     pl.HasVelocity,
		AltitudeUnits:   AltitudeUnits_name[int32(pl.AltitudeUnits)],
		FlightStatus:    FlightStatus_name[int32(pl.FlightStatus)],
		Airframe:        pl.AirframeType.Describe(),
		AirframeType:    airframeTypeReverseLookup[pl.AirframeType],
		SourceTag:       pl.SourceTag,
		Squawk:          strconv.FormatUint(uint64(pl.Squawk), 10),
		Special:         pl.Special,
		TileLocation:    pl.TileLocation,
		SourceTags:      pl.CloneSourceTags(),
		sourceTagsMutex: &sync.Mutex{},
		TrackedSince:    pl.TrackedSince.AsTime(),
		LastMsg:         pl.LastMsg.AsTime(),
		Updates: Updates{
			Location:     pl.Updates.Location.AsTime(),
			Altitude:     pl.Updates.Altitude.AsTime(),
			Velocity:     pl.Updates.Velocity.AsTime(),
			Heading:      pl.Updates.Heading.AsTime(),
			OnGround:     pl.Updates.OnGround.AsTime(),
			VerticalRate: pl.Updates.VerticalRate.AsTime(),
			FlightStatus: pl.Updates.FlightStatus.AsTime(),
			Special:      pl.Updates.Special.AsTime(),
			Squawk:       pl.Updates.Squawk.AsTime(),
		},
		SignalRssi:      &pl.SignalRssi,
		AircraftWidth:   &pl.AircraftWidth,
		AircraftLength:  &pl.AircraftLength,
		IcaoCode:        ptr(pl.IcaoStr()),
		Registration:    &pl.Registration,
		TypeCode:        &pl.TypeCode,
		TypeCodeLong:    &pl.TypeCodeLong,
		Serial:          &pl.Serial,
		RegisteredOwner: &pl.RegisteredOwner,
		COFAOwner:       &pl.COFAOwner,
		EngineType:      &pl.EngineType,
		FlagCode:        &pl.FlagCode,
		CallSign:        &pl.CallSign,
		Operator:        &pl.Operator,
		RouteCode:       &pl.RouteCode,
		Segments:        segments,
	}
}

func (pl *PlaneAndLocationInfoMsg) IncSourceTag(source string) {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()
	if nil == pl.SourceTags {
		pl.SourceTags = make(map[string]uint32)
	}
	pl.SourceTags[source]++
}

func FromProtobufBytes(protobufBytes []byte) (*PlaneAndLocationInfoMsg, error) {
	msg := PlaneAndLocationInfoMsg{
		sourceTagsMutex:      &sync.Mutex{},
		PlaneAndLocationInfo: &PlaneAndLocationInfo{},
	}

	err := proto.Unmarshal(protobufBytes, &msg)
	if nil == msg.SourceTags {
		msg.SourceTags = make(map[string]uint32)
	}
	return &msg, err
}
