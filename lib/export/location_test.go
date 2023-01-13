package export

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func getRefLocation() *PlaneLocation {
	return &PlaneLocation{
		PlaneLocationPB: PlaneLocationPB{
			Icao:            1234,
			Lat:             -31.1,
			Lon:             115.9,
			Heading:         223.4,
			Velocity:        666,
			Altitude:        31_000,
			VerticalRate:    -200,
			AltitudeUnits:   AltitudeUnits_FEET,
			FlightStatus:    FlightStatus_NormalAirborne,
			OnGround:        false,
			AirframeType:    AirframeType_CODE_01_LIGHT,
			HasAltitude:     true,
			HasLocation:     true,
			HasHeading:      true,
			HasOnGround:     true,
			HasFlightStatus: true,
			HasVerticalRate: true,
			HasVelocity:     true,
			SourceTag:       "benchmark",
			Squawk:          1234,
			Special:         "",
			TileLocation:    "tile35",
			TrackedSince:    timestamppb.Now(),
			LastMsg:         timestamppb.Now(),
			Updates: &FieldUpdates{
				Location:     timestamppb.Now(),
				Altitude:     timestamppb.Now(),
				Velocity:     timestamppb.Now(),
				Heading:      timestamppb.Now(),
				OnGround:     timestamppb.Now(),
				VerticalRate: timestamppb.Now(),
				FlightStatus: timestamppb.Now(),
				Special:      timestamppb.Now(),
				Squawk:       timestamppb.Now(),
			},
			SignalRssi:      0,
			AircraftWidth:   31,
			AircraftLength:  44,
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
			Segments:        []*PlaneSegment{},
		},
	}
}

func BenchmarkPlaneLocationEncode(b *testing.B) {
	loc := getRefLocation()
	for i := 0; i < b.N; i++ {
		_, err := loc.ToJsonBytes()
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationStdJsonEncode(b *testing.B) {
	loc := getRefLocation()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(loc)
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationProtobufEncode(b *testing.B) {
	loc := getRefLocation()
	var err error
	for i := 0; i < b.N; i++ {
		_, err = proto.Marshal(loc)
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationDecode(b *testing.B) {
	loc := getRefLocation()
	msg, err := loc.ToJsonBytes()
	if err != nil {
		b.Fail()
		return
	}
	b.Logf("Encoded Length: %d", len(msg))
	update := PlaneLocation{}

	var jsonFast = jsoniter.ConfigFastest
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = jsonFast.Unmarshal(msg, &update); nil != err {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationStdJsonDecode(b *testing.B) {
	loc := getRefLocation()
	msg, err := loc.ToJsonBytes()
	if err != nil {
		b.Fail()
		return
	}
	b.Logf("Encoded Length: %d", len(msg))
	update := PlaneLocation{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = json.Unmarshal(msg, &update); nil != err {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationProtobufDecode(b *testing.B) {
	loc := getRefLocation()
	protobufBytes, err := proto.Marshal(loc)
	if err != nil {
		b.Fail()
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := PlaneLocation{}
		err = proto.Unmarshal(protobufBytes, &msg)
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationEncodeJsonArray100(b *testing.B) {
	loc := getRefLocation()
	list := make([]*PlaneLocation, 100)
	for i := 0; i < 100; i++ {
		list[i] = loc
	}
	var jsonFast = jsoniter.ConfigFastest
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsonFast.Marshal(list)
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationDecodeJsonArray100(b *testing.B) {
	loc := getRefLocation()
	list := make([]*PlaneLocation, 100)
	list2 := make([]*PlaneLocation, 0)
	for i := 0; i < 100; i++ {
		list[i] = loc
	}
	var jsonFast = jsoniter.ConfigFastest
	listBytes, err := jsonFast.Marshal(list)
	if err != nil {
		b.Fail()
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = jsonFast.Unmarshal(listBytes, &list2)
		if err != nil {
			b.Errorf("Failed to decode: %s", err)
			return
		}
	}
}

func BenchmarkPlaneLocationEncodeProtobufArray100(b *testing.B) {
	loc := getRefLocation()
	list := PlaneLocations{
		PlaneLocation: make([]*PlaneLocationPB, 100),
	}
	for i := 0; i < 100; i++ {
		list.PlaneLocation[i] = &loc.PlaneLocationPB
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := proto.Marshal(&list)
		if err != nil {
			b.Fail()
			return
		}
	}
}

func BenchmarkPlaneLocationDecodeProtobufArray100(b *testing.B) {
	loc := getRefLocation()
	list := PlaneLocations{
		PlaneLocation: make([]*PlaneLocationPB, 100),
	}
	list2 := PlaneLocations{}
	for i := 0; i < 100; i++ {
		list.PlaneLocation[i] = &loc.PlaneLocationPB
	}
	listBytes, err := proto.Marshal(&list)
	if err != nil {
		b.Fail()
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = proto.Unmarshal(listBytes, &list2)
		if err != nil {
			b.Errorf("Failed to decode: %s", err)
			return
		}
	}
}
