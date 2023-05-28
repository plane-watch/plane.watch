package tile_grid

import (
	"fmt"
	"testing"
)

func TestGlobeIndexSpecialTile_contains(t1 *testing.T) {
	type pt struct {
		lat  float64
		long float64
	}
	type fields struct {
		// North+West, South+East
		nw, se pt
	}
	perth := pt{lat: -31.952162, long: 115.943482}
	london := pt{lat: 51.5, long: 10}
	tests := []struct {
		name   string
		fields fields
		args   pt
		want   bool
	}{
		{
			"top left of map",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: 75, long: -165}},
			pt{lat: 80, long: -170},
			true,
		},
		{
			"top left of map no perth",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: 75, long: -165}},
			perth,
			false,
		},
		{
			"contains centre",
			fields{nw: pt{lat: 20, long: -20} /* => */, se: pt{lat: -20, long: 20}},
			pt{lat: 0, long: 0},
			true,
		},
		{
			"world contains Perth",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: -90, long: 180}},
			perth,
			true,
		},
		{
			"tile contains Perth",
			fields{nw: pt{lat: -31, long: -115} /* => */, se: pt{lat: -32, long: 116}},
			perth,
			true,
		},
		{
			"northern hemisphere does not contain Perth",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: 0, long: 180}},
			perth,
			false,
		},
		{
			"northern hemisphere does contain London",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: 0, long: 180}},
			london,
			true,
		},
		{
			"southern hemisphere does not contain london",
			fields{nw: pt{lat: 0, long: -180} /* => */, se: pt{lat: -90, long: 180}},
			london,
			false,
		},
		{
			"southern hemisphere does contain perth",
			fields{nw: pt{lat: 0, long: -180} /* => */, se: pt{lat: -90, long: 180}},
			perth,
			true,
		},

		{
			"western hemisphere does not contain perth",
			fields{nw: pt{lat: 90, long: -180} /* => */, se: pt{lat: -90, long: 0}},
			perth,
			false,
		},
		{
			"eastern hemisphere does contain perth",
			fields{nw: pt{lat: 90, long: 0} /* => */, se: pt{lat: -90, long: 180}},
			perth,
			true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			// sanity check values
			if tt.fields.nw.lat < -90 || tt.fields.nw.lat > 90 {
				t1.Errorf("nw lat out of bounds -90...90. %f", tt.fields.nw.lat)
			}
			if tt.fields.nw.long < -180 || tt.fields.nw.long > 180 {
				t1.Errorf("nw long out of bounds -180...180. %f", tt.fields.nw.long)
			}
			if tt.fields.se.lat < -90 || tt.fields.se.lat > 90 {
				t1.Errorf("se lat out of bounds -90...90. %f", tt.fields.se.lat)
			}
			if tt.fields.se.long < -180 || tt.fields.se.long > 180 {
				t1.Errorf("se long out of bounds -180...180. %f", tt.fields.se.long)
			}
			t := GlobeIndexSpecialTile{
				debug: true,
				North: tt.fields.nw.lat,
				West:  tt.fields.nw.long,

				South: tt.fields.se.lat,
				East:  tt.fields.se.long,
			}
			if got := t.contains(tt.args.lat, tt.args.long); got != tt.want {
				t1.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_LookupTile(t *testing.T) {
	type args struct {
		lat float64
		lon float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Perth is found",
			args{lat: -31.952162, lon: 115.943482},
			"tile35",
		},
		{
			"53.253113, 179.723145",
			args{53.253113, 179.723145},
			"tile74",
		},
		{
			"-16, 1",
			args{-16, 1},
			"tile38",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LookupTile(tt.args.lat, tt.args.lon); got != tt.want {
				t.Errorf("LookupTile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGridLocationNames(t *testing.T) {
	if 0 == len(GridLocationNames()) {
		t.Errorf("Do not have a world")
	}
	if len(worldGrid) != len(GridLocationNames()) {
		t.Errorf("Failed to get the correct number of grid location names")
	}
	for _, name := range GridLocationNames() {
		if "" == name {
			t.Errorf("Got an empty name for tile")
		}
	}
}

func TestTileLookupsSame(t *testing.T) {
	var count, failed int
	for lat := -90.0; lat < 90.0; lat += 1 {
		for lon := -180.0; lon < 180.0; lon += 1 {
			count++
			name := fmt.Sprintf("lookup_%0.2f_%0.2f", lat, lon)
			t.Run(name, func(tt *testing.T) {
				manual := lookupTileManual(lat, lon)
				preCalc := lookupTilePreCalc(lat, lon)
				if manual != preCalc {
					failed++
					tt.Errorf("Lookup Difference. Precalc: %s, manual: %s", preCalc, manual)
				}
			})
		}
	}
	if failed > 0 {
		t.Errorf("%d/%d failed", failed, count)
	}
}

func BenchmarkLookupTileManual(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for lat := -90.0; lat < 90.0; lat += 1 {
			for lon := -180.0; lon < 180.0; lon += 1 {
				lookupTileManual(lat, lon)
			}
		}
	}
}

func BenchmarkLookupTilePreCalc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for lat := -90.0; lat < 90.0; lat += 1 {
			for lon := -180.0; lon < 180.0; lon += 1 {
				lookupTilePreCalc(lat, lon)
			}
		}
	}
}

func TestNoTileUnknown(t *testing.T) {
	numFails := 0
	for lat := -90.0; lat < 90.0; lat += 1 {
		for lon := -180.0; lon < 180.0; lon += 1 {
			if "tileUnknown" == lookupTilePreCalc(lat, lon) {
				t.Errorf("tileUnknown for %0.2f, %0.2f", lat, lon)
				numFails++
			}
		}
	}
	if numFails > 0 {
		t.Errorf("Failed to lookup %d items", numFails)
	}
}
