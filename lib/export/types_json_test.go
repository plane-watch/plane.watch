package export

import (
	"github.com/rs/zerolog"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	m.Run()
}

func TestIsLocationPossible(t *testing.T) {
	msg1t := time.Date(2023, time.January, 9, 19, 0, 0, 0, time.Local)
	msg2t := time.Date(2023, time.January, 9, 19, 0, 1, 0, time.Local)

	if msg2t.Before(msg1t) {
		t.Error("I don't understand time")
	}

	pos1 := PlaneLocationJSON{Lat: -31.942017, Lon: 115.964594, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: msg1t}
	pos2 := PlaneLocationJSON{Lat: -31.940887, Lon: 115.964897, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: msg2t}

	if !IsLocationPossible(pos1, pos2) {
		t.Error("Pos1 -> Pos2 is possible")
	}

	if IsLocationPossible(pos2, pos1) {
		t.Error("Pos2 -> Pos1 is not possible")
	}
}

func TestMergeCallSign(t *testing.T) {
	type args struct {
		prev PlaneLocationJSON
		next PlaneLocationJSON
	}
	tests := []struct {
		name    string
		args    args
		want    PlaneLocationJSON
		wantErr bool
	}{
		{
			name: "both-filled",
			args: args{
				prev: PlaneLocationJSON{CallSign: ptr("ONE")},
				next: PlaneLocationJSON{CallSign: ptr("TWO")},
			},
			want:    PlaneLocationJSON{CallSign: ptr("TWO")},
			wantErr: false,
		},
		{
			name: "One-Blank",
			args: args{
				prev: PlaneLocationJSON{CallSign: ptr("")},
				next: PlaneLocationJSON{CallSign: ptr("TWO")},
			},
			want:    PlaneLocationJSON{CallSign: ptr("TWO")},
			wantErr: false,
		},
		{
			name: "Two-Blank",
			args: args{
				prev: PlaneLocationJSON{CallSign: ptr("ONE")},
				next: PlaneLocationJSON{CallSign: ptr("")},
			},
			want:    PlaneLocationJSON{CallSign: ptr("ONE")},
			wantErr: false,
		},
		{
			name: "One-nil",
			args: args{
				prev: PlaneLocationJSON{CallSign: nil},
				next: PlaneLocationJSON{CallSign: ptr("")},
			},
			want:    PlaneLocationJSON{CallSign: ptr("")},
			wantErr: false,
		},
		{
			name: "Two-nil",
			args: args{
				prev: PlaneLocationJSON{CallSign: ptr("ONE")},
				next: PlaneLocationJSON{CallSign: nil},
			},
			want:    PlaneLocationJSON{CallSign: ptr("ONE")},
			wantErr: false,
		},
		{
			name: "both-nil",
			args: args{
				prev: PlaneLocationJSON{CallSign: nil},
				next: PlaneLocationJSON{CallSign: nil},
			},
			want:    PlaneLocationJSON{CallSign: nil},
			wantErr: false,
		},
		{
			name: "both-blank",
			args: args{
				prev: PlaneLocationJSON{CallSign: ptr("")},
				next: PlaneLocationJSON{CallSign: ptr("")},
			},
			want:    PlaneLocationJSON{CallSign: ptr("")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergePlaneLocations(tt.args.prev, tt.args.next)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergePlaneLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if nil != tt.want.CallSign {
				if unPtr(got.CallSign) != unPtr(tt.want.CallSign) {
					t.Errorf("MergePlaneLocations() got = %s, want %s", unPtr(got.CallSign), unPtr(tt.want.CallSign))
				}
			} else {
				if got.CallSign != nil {
					t.Errorf("Expected a nil callsign, got %s", unPtr(got.CallSign))
				}
			}
		})
	}
}
