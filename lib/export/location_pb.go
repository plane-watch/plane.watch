package export

import (
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/timestamppb"
	"plane.watch/lib/tracker"
	"sync"
)

func NewPBLocationFrom(location PlaneLocationJSON) {

}

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

var locationPool sync.Pool

func init() {
	locationPool = sync.Pool{
		New: func() any {
			return &PlaneLocationPBMsg{
				PlaneLocationPB: PlaneLocationPB{
					state:           protoimpl.MessageState{},
					sizeCache:       0,
					unknownFields:   nil,
					Icao:            0,
					Lat:             0,
					Lon:             0,
					Heading:         0,
					Velocity:        0,
					Altitude:        0,
					VerticalRate:    0,
					AltitudeUnits:   0,
					FlightStatus:    0,
					OnGround:        false,
					AirframeType:    0,
					HasAltitude:     false,
					HasLocation:     false,
					HasHeading:      false,
					HasOnGround:     false,
					HasFlightStatus: false,
					HasVerticalRate: false,
					HasVelocity:     false,
					SourceTag:       "",
					Squawk:          0,
					Special:         "",
					TileLocation:    "",
					SourceTags:      make(map[string]uint32),
					TrackedSince:    timestamppb.Now(),
					LastMsg:         timestamppb.Now(),
					Updates: &FieldUpdates{
						state:         protoimpl.MessageState{},
						sizeCache:     0,
						unknownFields: nil,
						Location:      timestamppb.Now(),
						Altitude:      timestamppb.Now(),
						Velocity:      timestamppb.Now(),
						Heading:       timestamppb.Now(),
						OnGround:      timestamppb.Now(),
						VerticalRate:  timestamppb.Now(),
						FlightStatus:  timestamppb.Now(),
						Special:       timestamppb.Now(),
						Squawk:        timestamppb.Now(),
					},
					SignalRssi:      0,
					AircraftWidth:   0,
					AircraftLength:  0,
					Registration:    "",
					TypeCode:        "",
					TypeCodeLong:    "",
					Serial:          "",
					RegisteredOwner: "",
					COFAOwner:       "",
					EngineType:      "",
					FlagCode:        "",
					CallSign:        "",
					Operator:        "",
					RouteCode:       "",
					Segments:        make([]*PlaneSegment, 0),
				},
			}
		},
	}
}

func NewPlaneLocationPB(plane *tracker.Plane, source string) *PlaneLocationPBMsg {
	var altitudeUnits AltitudeUnits

	if plane.AltitudeUnits() == "feet" {
		altitudeUnits = AltitudeUnits_FEET
	} else {
		altitudeUnits = AltitudeUnits_METRES
	}

	msg := locationPool.Get().(*PlaneLocationPBMsg)

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

func Release(msg *PlaneLocationPBMsg) {
	locationPool.Put(msg)
}

// ToJSONBytes is SO SLOW using the protobuf message encoder, but it is correct output for the Timestamps
// DO NOT USE for anything other than debugging
func (pl *PlaneLocationPBMsg) ToJSONBytes() ([]byte, error) {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()

	jsonBuf, err := protojson.Marshal(&pl.PlaneLocationPB)
	if nil != err {
		log.Error().Err(err).Msg("could not create json bytes for sending")
		return nil, err
	}
	return jsonBuf, nil
}

func FromJSONBytes(jsonBytes []byte) (*PlaneLocationPBMsg, error) {
	var msg PlaneLocationPBMsg
	msg.sourceTagsMutex = &sync.Mutex{}
	err := protojson.Unmarshal(jsonBytes, &msg.PlaneLocationPB)
	return &msg, err
}

func (pl *PlaneLocationPBMsg) ToProtobufBytes() ([]byte, error) {
	protobufBytes, err := proto.Marshal(&pl.PlaneLocationPB)
	if nil != err {
		log.Error().Err(err).Msg("could not create json bytes for sending")
		return nil, err
	}
	return protobufBytes, nil
}

func (pl *PlaneLocationPBMsg) IncSourceTag(source string) {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()
	if nil == pl.SourceTags {
		pl.SourceTags = make(map[string]uint32)
	}
	pl.SourceTags[source]++
}

func FromProtobufBytes(protobufBytes []byte) (*PlaneLocationPBMsg, error) {
	var msg PlaneLocationPBMsg
	msg.sourceTagsMutex = &sync.Mutex{}

	err := proto.Unmarshal(protobufBytes, &msg)
	if nil == msg.SourceTags {
		msg.SourceTags = make(map[string]uint32)
	}
	return &msg, err
}
