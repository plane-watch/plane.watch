package export

import (
	"testing"
	"time"
)

func TestIsLocationPossible(t *testing.T) {
	msg1t := time.Date(2023, time.January, 9, 19, 0, 0, 0, time.Local)
	msg2t := time.Date(2023, time.January, 9, 19, 0, 1, 0, time.Local)

	if msg2t.Before(msg1t) {
		t.Error("I don't understand time")
	}

	pos1 := PlaneLocation{Lat: -31.942017, Lon: 115.964594, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: msg1t}
	pos2 := PlaneLocation{Lat: -31.940887, Lon: 115.964897, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: msg2t}

	if !IsLocationPossible(pos1, pos2) {
		t.Error("Pos1 -> Pos2 is possible")
	}

	if IsLocationPossible(pos2, pos1) {
		t.Error("Pos2 -> Pos1 is not possible")
	}
}
