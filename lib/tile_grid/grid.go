package tile_grid

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"math"
)

type (
	GlobeIndexSpecialTile struct {
		debug bool
		North float64 `json:"north"`
		East  float64 `json:"east"`
		South float64 `json:"south"`
		West  float64 `json:"west"`
	}

	GridLocations map[string]GlobeIndexSpecialTile
)

var (
	preCalcGrid [180][360]string
)

func init() {
	for lat := -90; lat < 90; lat++ {
		for lon := -180; lon < 180; lon++ {
			preCalcGrid[lat+90][lon+180] = lookupTileManual(float64(lat), float64(lon))
		}
	}
}

func LookupTile(lat, lon float64) string {
	return lookupTilePreCalc(lat, lon)
}

func lookupTilePreCalc(lat, lon float64) string {
	latInt := int(math.Floor(lat))
	lonInt := int(math.Floor(lon))
	if latInt < -90 || latInt >= 90 || lonInt < -180 || lonInt >= 180 {
		return "tileUnknown"
	}
	return preCalcGrid[latInt+90][lonInt+180]
}

func lookupTileManual(lat, lon float64) string {
	if lat < -90.0 || lat > 90 || lon < -180 || lon > 180 {
		log.Error().Err(fmt.Errorf("cannot lookup invalid coordinates {%0.6f, %0.6f}", lat, lon)).Msg("Using No Tile")
		return ""
	}

	for name, t := range worldGrid {
		if t.contains(lat, lon) {
			return name
		}
	}

	log.Warn().
		Float64("lat", lat).
		Float64("lon", lon).
		Err(fmt.Errorf("could Not Place {%0.6f, %0.6f} in a grid location", lat, lon)).
		Msg("Using No tileUnknown")
	return "tileUnknown"
}

func InGridLocation(lat, lon float64, tileName string) bool {
	if t, ok := worldGrid[tileName]; ok {
		return t.contains(lat, lon)
	}
	return false
}

func GridLocationNames() []string {
	names := make([]string, len(worldGrid))
	i := 0
	for name := range worldGrid {
		names[i] = name
		i++
	}
	return names
}

// contains determines whether the
// * lat is contained between North and South, and
// * lon is contained between East and West
func (t GlobeIndexSpecialTile) contains(lat, lon float64) bool {
	//contains := (lat <= t.North && lat > t.South) && (lon >= t.East && lon < t.West)

	// 90 = top, -90 == bottom
	s := t.South
	if t.South == -90 {
		s -= 0.1 // so the calc below works nicely
	}
	containsLat := lat <= t.North && lat > s
	// -180 == west, 180 == east
	containsLon := lon >= t.West && lon < t.East
	if t.debug {
		log.Debug().
			Floats64(`NW`, []float64{t.West, t.North}).
			Floats64(`SE`, []float64{t.East, t.South}).
			Floats64(`Pnt`, []float64{lat, lon}).
			Floats64(`EW Range`, []float64{t.West, lon, t.East}).
			Bool(`Contains Lat`, containsLat).
			Bool(`Contains Lon`, containsLon).
			Floats64(`NS Range`, []float64{t.North, lat, t.South}).
			Bool(`Contains`, containsLat && containsLon).
			Send()
	}
	return containsLat && containsLon
}

func GetGrid() GridLocations {
	return worldGrid
}
