package main

import (
	"plane.watch/lib/export"
	"testing"
	"time"
)

func Test_getDistanceBetween(t *testing.T) {
	type args struct {
		lat1 float64
		lon1 float64
		lat2 float64
		lon2 float64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "statue-liberty-eiffel-tower",
			args: args{
				lat1: 40.6892,
				lon1: -74.0444,
				lat2: 48.8583,
				lon2: 2.2945,
			},
			want: 5_837_413,
		},
		{
			name: "water-park",
			args: args{
				lat1: -32.290026755368224,
				lon1: 115.85115467567556,
				lat2: -32.28937695712494,
				lon2: 115.85751010601774,
			},
			want: 602,
		},
		{
			name: "near-beverley",
			args: args{
				lat1: -32.22732514010243,
				lon1: 116.65858573202698,
				lat2: -32.163891745518725,
				lon2: 117.4699887407923,
			},
			want: 76_675,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDistanceBetween(tt.args.lat1, tt.args.lon1, tt.args.lat2, tt.args.lon2); got != tt.want {
				t.Errorf("getDistanceBetween() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_getDistanceBetween(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getDistanceBetween(float64(i), float64(i), float64(i+b.N), float64(i+b.N))
	}
}

func Test_pwAlertBot_alertUser(t *testing.T) {
	// make sure we only send one update per 5 minutes
	var alertCount int
	a := &pwAlertBot{
		sendAlert: func(pa *proximityAlert) {
			alertCount++
		},
	}

	a.alertUser(nil)
	if 0 != alertCount {
		t.Errorf("Sent an alert when nil")
	}

	pa := proximityAlert{
		time: time.Now(),
		alert: &location{
			LocationName:  "test-1",
			DiscordUserId: "testerer",
		},
		update: &export.PlaneLocationJSON{
			Icao: "01AB23",
		},
		distanceMtr: 23,
	}
	expected := "testerertest-101AB23"
	if expected != pa.Key() {
		t.Errorf("Did not generate the correct key. %s != %s", expected, pa.Key())
	}

	a.alertUser(&pa)
	if 1 != alertCount {
		t.Errorf("Expected to send an alert")
	}
	a.alertUser(&pa)
	if 1 != alertCount {
		t.Errorf("Should not have send the same alert twice")
	}

	for i := 0; i < 300; i++ {
		pa.time = pa.time.Add(time.Second)
		a.alertUser(&pa)
		if 1 != alertCount {
			t.Errorf("Should not have sent the proximity alert again, offset %d seconds", i)
		}
	}
}
