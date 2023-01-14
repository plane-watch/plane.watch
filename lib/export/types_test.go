package export

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestIsLocationPossible(t *testing.T) {
	msg1t := time.Date(2023, time.January, 9, 19, 0, 0, 0, time.Local)
	msg2t := time.Date(2023, time.January, 9, 19, 0, 1, 0, time.Local)

	if msg2t.Before(msg1t) {
		t.Error("I don't understand time")
	}

	pos1 := &PlaneLocation{PlaneLocationPB: PlaneLocationPB{Lat: -31.942017, Lon: 115.964594, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: timestamppb.New(msg1t)}}
	pos2 := &PlaneLocation{PlaneLocationPB: PlaneLocationPB{Lat: -31.940887, Lon: 115.964897, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: timestamppb.New(msg2t)}}

	if !IsLocationPossible(pos1, pos2) {
		t.Error("Pos1 -> Pos2 is possible")
	}

	if IsLocationPossible(pos2, pos1) {
		t.Error("Pos2 -> Pos1 is not possible")
	}
}

//func TestPlaneLocation_ToJsonBytes(t *testing.T) {
//	// old
//	expected := []byte(`{"Icao":1234,"Lat":-31.1,"Lon":0,"Heading":0,"Velocity":0,"Altitude":0,"VerticalRate":0,"AltitudeUnits":"","FlightStatus":"","OnGround":false,"Airframe":"","AirframeType":"","HasAltitude":false,"HasLocation":false,"HasHeading":false,"HasOnGround":false,"HasFlightStatus":false,"HasVerticalRate":false,"HasVelocity":false,"SourceTag":"","Squawk":"","Special":"","TileLocation":"","TrackedSince":"0001-01-01T00:00:00Z","LastMsg":"0001-01-01T00:00:00Z","Updates":{"Location":"0001-01-01T00:00:00Z","Altitude":"0001-01-01T00:00:00Z","Velocity":"0001-01-01T00:00:00Z","Heading":"0001-01-01T00:00:00Z","OnGround":"0001-01-01T00:00:00Z","VerticalRate":"0001-01-01T00:00:00Z","FlightStatus":"0001-01-01T00:00:00Z","Special":"0001-01-01T00:00:00Z","Squawk":"0001-01-01T00:00:00Z"},"SignalRssi":null}`)
//	// new
//
//	p := getRefLocation()
//	buf, err := p.ToJsonBytes()
//	if nil != err {
//		t.Error(err)
//	}
//
//	if 0 != bytes.Compare(expected, buf) {
//		t.Errorf("Incorrect JSON Format")
//		t.Errorf("Expected  %s", string(expected))
//		t.Errorf("Incorrect %s", string(buf))
//	}
//}
