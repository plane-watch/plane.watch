package tracker

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"plane.watch/lib/tracker/mode_s"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
	m.Run()
}

func TestNLFunc(t *testing.T) {
	for i, f := range NLTable {
		if r := getNumLongitudeZone(f - 0.01); i != r {
			t.Errorf("NL Table Fail: Expected %0.2f to yield %d, got %d", f, i, r)
		}
	}
}

func TestCprDecode(t *testing.T) {
	type testDataType struct {
		evenLat, evenLon float64
		oddLat, oddLon   float64

		evenRlatCheck1, evenRlonCheck1 string

		evenRlat, evenRlon string
		oddRlat, oddRlon   string
	}
	testData := []testDataType{
		//odd *8d7c4516581f76e48d95e8ab20ca; even *8d7c4516581f6288f83ade534ae1;
		{evenLat: 83068, evenLon: 15070, oddLat: 94790, oddLon: 103912, oddRlat: "-32.197483", oddRlon: "+116.028629", evenRlat: "-32.197449", evenRlon: "+116.027820"},

		// odd *8d7c4516580f06fc6d8f25d8669d; even *8d7c4516580df2a168340b32212a;
		{evenLat: 86196, evenLon: 13323, oddLat: 97846, oddLon: 102181, oddRlat: "-32.055219", oddRlon: "+115.931602", evenRlat: "-32.054260", evenRlon: "+115.931854"},

		// test data from cprtest.c from mutability dump1090
		{evenLat: 80536, evenLon: 9432, oddLat: 61720, oddLon: 9192, evenRlat: "+51.686646", evenRlon: "+0.700156", oddRlat: "+51.686763", oddRlon: "+0.701294"},
	}
	airDlat0 := "+6.000000"
	airDlat1 := "+6.101695"
	trk := NewTracker()

	for i, d := range testData {
		plane := trk.GetPlane(11234)

		plane.setCprOddLocation(d.oddLat, d.oddLon, time.Now())
		time.Sleep(2)
		plane.setCprEvenLocation(d.evenLat, d.evenLon, time.Now())
		loc, err := plane.cprLocation.decodeGlobalAir()
		if err != nil {
			t.Error(err)
		}

		lat := fmt.Sprintf("%+0.6f", loc.latitude)
		lon := fmt.Sprintf("%+0.6f", loc.longitude)

		if lat != d.oddRlat {
			t.Errorf("Plane latitude is wrong for packet %d: should be %s, was %s", i, d.oddRlat, lat)
		}
		if lon != d.oddRlon {
			t.Errorf("Plane latitude is wrong for packet %d: should be %s, was %s", i, d.oddRlon, lon)
		}

		if airDlat0 != fmt.Sprintf("%+0.6f", plane.cprLocation.airDLat0) {
			t.Error("AirDlat0 is wrong")
		}
		if airDlat1 != fmt.Sprintf("%+0.6f", plane.cprLocation.airDLat1) {
			t.Error("AirDlat1 is wrong")
		}

		plane.setCprEvenLocation(d.evenLat, d.evenLon, time.Now())
		time.Sleep(2)
		plane.setCprOddLocation(d.oddLat, d.oddLon, time.Now())
		loc, err = plane.cprLocation.decodeGlobalAir()
		if err != nil {
			t.Error(err)
		}

		lat = fmt.Sprintf("%+0.6f", loc.latitude)
		lon = fmt.Sprintf("%+0.6f", loc.longitude)

		if lat != d.evenRlat {
			t.Errorf("Plane latitude is wrong for packet %d: should be %s, was %s", i, d.evenRlat, lat)
		}
		if lon != d.evenRlon {
			t.Errorf("Plane latitude is wrong for packet %d: should be %s, was %s", i, d.evenRlon, lon)
		}

		if airDlat0 != fmt.Sprintf("%+0.6f", plane.cprLocation.airDLat0) {
			t.Error("AirDlat0 is wrong")
		}
		if airDlat1 != fmt.Sprintf("%+0.6f", plane.cprLocation.airDLat1) {
			t.Error("AirDlat1 is wrong")
		}

	}
}

func TestTracking(t *testing.T) {
	frames := []string{
		"*8D40621D58C382D690C8AC2863A7;",
		"*8D40621D58C386435CC412692AD6;",
	}
	trk := performTrackingTest(frames, t)

	plane := trk.GetPlane(4219421)
	if alt := plane.Altitude(); alt != 38000 {
		t.Errorf("Plane should be at 38000 feet, was %d", alt)
	}

	lat := "+52.2572021484375"
	lon := "+3.9193725585938"
	if lon != fmt.Sprintf("%+03.13f", plane.Lon()) {
		t.Errorf("longitude Calculation was incorrect: expected %s, got %+0.13f", lon, plane.Lon())
	}
	if lat != fmt.Sprintf("%+03.13f", plane.Lat()) {
		t.Errorf("latitude Calculation was incorrect: expected %s, got %+0.13f", lat, plane.Lat())
	}
}

func TestTracking2(t *testing.T) {
	frames := []string{
		"*8D7C7DAA99146D0980080D6131A1;",
		"*5D7C7DAACD3CE9;",
		"*0005050870B303;",
		"*8D7C7DAA99146C0980040D2A616F;",
		"*8D7C7DAAF80020060049B06CA244;",
		"*8D7C7DAA582886FA618B21ADB377;",
		"*5D7C7DAACD3CE9;",
		"*8D7C7DAA5828829F322FE81F6DD1;",
		"*8D7C7DAA99146C0980040D2A616F;",
		"*8D7C7DAA99146C0980040D2A616F;",
		"*8D7C7DAA99146C0960080D47BBB9;",
		"*8D7C7DAA582886FA778B115D2F89;",
		"*000005084A3646;",
		"*000005084A3646;",
		"*28000A00307264;",
		"*8D7C7DAA99146A09280C0D91E947;",
		"*8D7C7DAA9914690920080DC2621D;",
		"*8D7C7DAA9914690928040DE49A15;",
		"*8D7C7DAA210DA1E0820820472D63;",
		"*5D7C7DAACD3CE9;",
		"*8D7C7DAA582886FB218A9AFB0420;",
		"*5D7C7DAACD3CE9;",
		"*8D7C7DAA5828829FF42F5E556B2D;",
		"*8D7C7DAA9914680920080DC168D3;",
		"*000005084A3646;",
		"*5D7C7DAACD3CE9;",
		"*8D7C7DAA582886FB318A8FD96CD7;",
		"*8D7C7DAA9914670900080D9576E0;",
		"*000005084A3646;",
	}
	performTrackingTest(frames, t).Finish()

}

func performTrackingTest(frames []string, t *testing.T) *Tracker {
	trk := NewTracker()
	for _, msg := range frames {
		frame, err := mode_s.DecodeString(msg, time.Now())
		if nil != err {
			t.Errorf("%s", err)
		}
		trk.GetPlane(frame.Icao()).HandleModeSFrame(frame, nil, nil)
	}
	return trk
}

// Makes sure that we get a location update only when we need one
// The logic we want:
//   Only add something to history if was previously valid and it has now changed
//   example:
//    first frame has Alt, no history
//    second frame has half a location, no history
//    third frame has other half location, no history (alt and location are now valid)
//    forth frame has same alt, no history
//    fifth frame has new alt, 1 history with old alt and location)
//    six frame has heading, 1 history
// Things that change that give us a history
//   Lat, Long, Alt, GroundStatus, Heading
func TestTrackingLocationHistory(t *testing.T) {
	tests := []struct {
		name         string
		frame        string
		numLocations int
	}{
		// ground position does not trigger location history, only lat/lon does
		{name: "DF17/MT31/ST00 Airborne Status Frame", frame: "8D7C4A0CF80300030049B8BA7984", numLocations: 0},
		{name: "DF17/MT31/ST00 Airborne Status Frame", frame: "8D7C4A0CF80300030049B8BA7984", numLocations: 0},

		{name: "DF17/MT31/ST01 Ground Status Frame", frame: "8C7C4A0CF9004103834938E42BD4", numLocations: 0},
		{name: "DF17/MT31/ST01 Ground Status Frame", frame: "8C7C4A0CF9004103834938E42BD4", numLocations: 0},

		{name: "DF17/MT11/Odd", frame: "8D7C75285841B71C2FB174E7746B", numLocations: 0},
		{name: "DF17/MT11/Even", frame: "8D7C75285841C2C178571CF5234E", numLocations: 1},
	}
	trk := NewTracker()
	// our second test should have our plane in the air, so we can put it on the ground
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := mode_s.DecodeString(tt.frame, time.Now())
			if nil != err {
				t.Error(err)
				return
			}
			if nil == frame {
				t.Errorf("nil frame from avr frame %s", tt.frame)
				return
			}
			plane := trk.GetPlane(frame.Icao())
			plane.HandleModeSFrame(frame, nil, nil)
			numHistory := len(plane.locationHistory)
			if tt.numLocations != numHistory {
				t.Errorf("Expected plane to have %d history items, actually has %d", tt.numLocations, numHistory)
			}
		})
	}
	p := trk.GetPlane(0x7C7528)
	if nil == p {
		t.Errorf("Failed to get our plane")
	}
	if !p.HasLocation() {
		t.Errorf("Did not set location correctly")
	}
}

func TestPlane_HasLocation(t *testing.T) {
	trk := NewTracker()
	p := trk.GetPlane(0x010101)
	err := p.addLatLong(0.01, 0.02, time.Now())
	if nil != err {
		t.Errorf("Got error when adding lat/lon: %s", err)
	}
	if !p.HasLocation() {
		t.Error("Did not correctly set plane location has updated flag")
	}
	if 1 != len(p.locationHistory) {
		t.Errorf("Expected plane history to have 1 item. have %d", len(p.locationHistory))
	}
}

func TestPlane_HasHeading(t *testing.T) {
	trk := NewTracker()
	p := trk.GetPlane(0x010101)
	if p.HasLocation() {
		t.Error("Did not expect to have a heading")
	}

	changed := p.setHeading(99)
	if !changed {
		t.Error("Expected that setting our heading got a change")
	}

	if !p.HasHeading() {
		t.Error("Did not correctly set has heading")
	}
}

func TestPlane_HasVerticalRate(t *testing.T) {
	trk := NewTracker()
	p := trk.GetPlane(0x010101)
	if p.HasVerticalRate() {
		t.Error("Did not expect to have a vertical rate")
	}

	changed := p.setVerticalRate(99)
	if !changed {
		t.Error("Expected that setting our vertical rate got a change")
	}

	if !p.HasVerticalRate() {
		t.Error("Did not correctly set has vertical rate")
	}
}

func TestPlane_HasVelocity(t *testing.T) {
	trk := NewTracker()
	p := trk.GetPlane(0x010101)
	if p.HasVelocity() {
		t.Error("Did not expect to have a velocity")
	}

	changed := p.setVelocity(99)
	if !changed {
		t.Error("Expected that setting our velocity got a change")
	}

	if !p.HasVelocity() {
		t.Error("Did not correctly set velocity")
	}
}

func TestCorrectCprDecodeSouthAmerica(t *testing.T) {
	type pair struct {
		odd, even string
		lat, lon  string
		icao      uint32
	}

	icao, _ := strconv.ParseUint("E065D3", 16, 32)

	samples := []pair{
		{
			odd:  "8DE065D358C38797B4F57E1A56F2",
			even: "8DE065D358C3833F06A8657B6B41",
			lat:  "-31.12995",
			lon:  "-54.14777",
			icao: uint32(icao),
		},
	}
	trk := NewTracker()
	for _, sample := range samples {
		t.Run(fmt.Sprintf("decode_%s_%s", sample.odd, sample.even), func(tt *testing.T) {
			oddFrame, err := mode_s.DecodeString(sample.odd, time.Now())
			if nil != err || oddFrame == nil {
				tt.Error(err)
				return
			}
			if oddFrame.IsEven() {
				tt.Error("Odd Frame was Even")
			}
			evenFrame, err := mode_s.DecodeString(sample.even, time.Now())
			if nil != err {
				tt.Error(err)
			}
			if !evenFrame.IsEven() {
				tt.Error("Even frame was Odd")
			}
			p := trk.GetPlane(sample.icao)
			if err = p.setCprEvenLocation(float64(evenFrame.Latitude()), float64(evenFrame.Longitude()), evenFrame.TimeStamp()); nil != err {
				tt.Error(err)
			}
			if err = p.setCprOddLocation(float64(oddFrame.Latitude()), float64(oddFrame.Longitude()), oddFrame.TimeStamp()); nil != err {
				tt.Error(err)
			}

			loc, err := p.cprLocation.decodeGlobalAir()
			if nil != err {
				tt.Error(err)
			}

			if loc.onGround {
				t.Error("Decoding air position resulted in aircraft being on ground")
			}

			if fmt.Sprintf("%0.5f", loc.latitude) != sample.lat {
				t.Errorf("Got incorrect Latitude: expecting: %s != got: %0.5f", sample.lat, loc.latitude)
			}
			if fmt.Sprintf("%0.5f", loc.longitude) != sample.lon {
				t.Errorf("Got incorrect Longitude: expecting: %s != got: %0.5f", sample.lon, loc.longitude)
			}
		})
	}
}
