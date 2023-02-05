package calc

import (
	"fmt"
	"testing"
	"time"
)

func TestAccelerationBetween(t *testing.T) {
	type args struct {
		prevTime     time.Time
		currentTime  time.Time
		prevVelocity float64
		prevLat      float64
		prevLon      float64
		currentLat   float64
		currentLon   float64
		prevHeading  float64
	}

	tests := []struct {
		name      string
		args      args
		want      float64
		varAmount float64 // +/- amount
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
				prevHeading:  180,
				currentLat:   -32.001394,
				currentLon:   116.0, // an extra ~ 155m,
			},
			want:      0.0,
			varAmount: 0.05,
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
				prevHeading:  180,

				currentLat: -32.001439, // an extra ~ 155m + 4.9m (0.5*9.8) across a second,
				currentLon: 116,
			},
			want:      4.5,
			varAmount: 0.05,
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
				prevHeading:  180,

				currentLat: -32,
				currentLon: 116.019149, // an extra ~ 1800m,
			},
			want:      4.5,
			varAmount: 0.05,
		},
		{
			name: "1G Accel over 3s",
			args: args{
				// 3 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 3, 0, time.Local),
				prevVelocity: 300, // 154.333 m/s
				prevLat:      -32,
				prevLon:      116,
				prevHeading:  180,

				currentLat: -32.004712, // an extra ~ 524m, (155+9.8)+(155+19.6)+(155+29.4)
				currentLon: 116,
			},
			want:      9.8,
			varAmount: 0.05,
		},
		{
			name: "B747 Takeoff - 0-280km/hr in 50 seconds",
			args: args{
				// 50 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 50, 0, time.Local),
				prevVelocity: 0, // starting at zero
				prevLat:      -32,
				prevLon:      116,
				prevHeading:  180,

				currentLat: -32.028778, // an extra 3.2km - found on the internet for 10deg flap T/O of B747
				currentLon: 116,
			},
			want:      5.6,
			varAmount: 0.05,
		},
		{
			name: "Old point, constant velocity.",
			args: args{
				// 50 second change
				prevTime:     time.Date(2022, 04, 17, 13, 30, 0, 0, time.Local),
				currentTime:  time.Date(2022, 04, 17, 13, 30, 1, 0, time.Local),
				prevVelocity: 300, // ~155m/s
				prevLat:      -32,
				prevLon:      116,
				prevHeading:  180,

				currentLat: -31.998606, // Gone backwards 155m
				currentLon: 116,
			},
			want:      200.0,
			varAmount: 150.0,
		},
		{
			name: "Constant Velocity - IRL example",
			args: args{
				// 1 second change
				prevTime:     time.Date(2023, 01, 14, 03, 45, 12, 50, time.Local),
				currentTime:  time.Date(2023, 01, 14, 03, 45, 13, 50, time.Local),
				prevVelocity: 178.263849, // ~91.7m/s
				prevLat:      -31.951361,
				prevLon:      115.961814,
				prevHeading:  193.722297,

				currentLat: -31.952246, // moved ~100m
				currentLon: 115.961590,
			},
			want:      0.0,
			varAmount: 0.05,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fmt.Printf("Running test: %s\n", tt.name)

			got := AccelerationBetween(
				tt.args.prevTime, tt.args.currentTime,
				tt.args.prevLat, tt.args.prevLon,
				tt.args.prevHeading, tt.args.prevVelocity,
				tt.args.currentLat, tt.args.currentLon)

			if got > tt.want-tt.varAmount && got < tt.want+tt.varAmount {
				// valid test.
				fmt.Printf("AccelerationBetween() = %v\n", got)
			} else {
				t.Errorf("AccelerationBetween() = %v, want %v +/- %v\n", got, tt.want, tt.varAmount)
			}
		})
	}
}
