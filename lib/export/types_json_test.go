package export

import (
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"sync"
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

	pos1 := &PlaneAndLocationInfo{Lat: -31.942017, Lon: 115.964594, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: timestamppb.New(msg1t), Updates: &FieldUpdates{}}
	pos2 := &PlaneAndLocationInfo{Lat: -31.940887, Lon: 115.964897, Heading: 14.116942, HasLocation: true, HasHeading: true, LastMsg: timestamppb.New(msg2t), Updates: &FieldUpdates{}}

	pos1Msg := &PlaneAndLocationInfoMsg{PlaneAndLocationInfo: pos1, sourceTagsMutex: &sync.Mutex{}}
	pos2Msg := &PlaneAndLocationInfoMsg{PlaneAndLocationInfo: pos2, sourceTagsMutex: &sync.Mutex{}}

	if !IsLocationPossible(pos1Msg, pos2Msg) {
		t.Error("Pos1 -> Pos2 is possible")
	}

	if IsLocationPossible(pos2Msg, pos1Msg) {
		t.Error("Pos2 -> Pos1 is not possible")
	}
}

func TestMergeCallSign(t *testing.T) {
	type args struct {
		prev *PlaneAndLocationInfo
		next *PlaneAndLocationInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *PlaneAndLocationInfo
		wantErr bool
	}{
		{
			name: "both-filled",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "ONE", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "TWO", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: "TWO"},
			wantErr: false,
		},
		{
			name: "One-Blank",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "TWO", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: "TWO"},
			wantErr: false,
		},
		{
			name: "Two-Blank",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "ONE", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: "ONE"},
			wantErr: false,
		},
		{
			name: "One-nil",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: ""},
			wantErr: false,
		},
		{
			name: "Two-nil",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "ONE", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: "ONE"},
			wantErr: false,
		},
		{
			name: "both-nil",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: ""},
			wantErr: false,
		},
		{
			name: "both-blank",
			args: args{
				prev: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
				next: &PlaneAndLocationInfo{CallSign: "", Updates: &FieldUpdates{}},
			},
			want:    &PlaneAndLocationInfo{CallSign: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergePlaneLocations(
				&PlaneAndLocationInfoMsg{PlaneAndLocationInfo: tt.args.prev},
				&PlaneAndLocationInfoMsg{PlaneAndLocationInfo: tt.args.next},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergePlaneLocations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want.CallSign != "" {
				if got.CallSign != tt.want.CallSign {
					t.Errorf("MergePlaneLocations() got = %s, want %s", got.CallSign, tt.want.CallSign)
				}
			} else {
				if got.CallSign != "" {
					t.Errorf("Expected a nil callsign, got %s", got.CallSign)
				}
			}
		})
	}
}

func BenchmarkSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%X", 9046412)
	}
}
func BenchmarkFormatInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strings.ToUpper(strconv.FormatInt(9046412, 16))
	}
}
