package export

import (
	"reflect"
	"sync"
	"testing"
)

func TestPlaneAndLocationInfoMsg_PrepareSourceTags(t *testing.T) {
	tests := []struct {
		name   string
		fields string
		want   map[string]uint32
	}{
		{
			name:   "feeder key with icao and id decodes correctly",
			fields: "YPPH-0001",
			want: map[string]uint32{
				"0001": 1,
			},
		},
		{
			name:   "feeder key with icao and id decodes correctly (lower)",
			fields: "ypph-0001",
			want: map[string]uint32{
				"0001": 1,
			},
		},
		{
			name:   "feeder key with long id works",
			fields: "long-0000001",
			want: map[string]uint32{
				"0000001": 1,
			},
		},
		{
			name:   "feeder key that's an API Key",
			fields: "550e8400-e29b-41d4-a716-446655440000",
			want: map[string]uint32{
				"550e8400-e29b-41d4-a716-446655440000": 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(map[string]uint32, 1)
			pl := &PlaneAndLocationInfoMsg{
				PlaneAndLocationInfo: &PlaneAndLocationInfo{
					SourceTags: map[string]uint32{
						tt.fields: 1,
					},
				},
				sourceTagsMutex: &sync.Mutex{},
			}
			if got := pl.PrepareSourceTags(m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrepareSourceTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
