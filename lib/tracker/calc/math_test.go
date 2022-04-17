package calc

import (
	"testing"
	"time"
)

func TestFlightLocationValid_Valid(t *testing.T) {
	type args struct {
		prevTime     time.Time
		currentTime  time.Time
		prevVelocity float64
		prevLat      float64
		prevLon      float64
		currentLat   float64
		currentLon   float64
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Constant Velocity",
			args: args{
				// 1 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 1, 0, time.Local),
				prevVelocity: 300, // 154.333 m/s
				prevLat:      -32,
				prevLon:      116,
				currentLat:   -32,
				currentLon:   116.00164, // an extra ~ 155m,
			},
			want: true,
		},
		{
			name: "0.5G Accel",
			args: args{
				// 1 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 1, 0, time.Local),
				prevVelocity: 300, // 154.333 m/s
				prevLat:      -32,
				prevLon:      116,

				currentLat: -32,
				currentLon: 116.001685, // an extra ~ 155m,
			},
			want: true,
		},
		{
			name: "0.5G Accel over 10s",
			args: args{
				// 1 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 10, 0, time.Local),
				prevVelocity: 300, // 154.333 m/s
				prevLat:      -32,
				prevLon:      116,

				currentLat: -32,
				currentLon: 116.019149, // an extra ~ 1800m,
			},
			want: true,
		},
		{
			name: "1G Accel over 5s",
			args: args{
				// 1 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 5, 0, time.Local),
				prevVelocity: 300, // 154.333 m/s
				prevLat:      -32,
				prevLon:      116,

				currentLat: -32,
				currentLon: 116.019149, // an extra ~ 1800m,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FlightLocationValid(tt.args.prevTime, tt.args.currentTime, tt.args.prevVelocity, tt.args.prevLat, tt.args.prevLon, tt.args.currentLat, tt.args.currentLon); got != tt.want {
				t.Errorf("FlightLocationValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
