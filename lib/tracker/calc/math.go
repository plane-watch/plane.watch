package calc

import (
	"math"
	"time"
)

const (
	// Mach1 in Metres/Second
	Mach1 = 343
	// Mach2 in Metres/Second
	Mach2 = 686
	// Mach3 in Metres/Second
	Mach3 = 1029
	// Mach4 in Metres/Second
	Mach4 = 1372
)

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// MaxAllowableDistance calculates how far a plane could have travelled
// interval is the interval between
func MaxAllowableDistance(interval time.Duration, velocity float64) float64 {
	if interval.Seconds() <= 0.001 { // this is a really quick report
		interval = 1
	}
	if velocity > Mach4 {
		velocity = Mach4 // clamp to Mach4
	}

	// SR71A's official was Mach 3.3, not much else gets that fast
	upperRangeLimit := interval.Seconds() * Mach4

	maxSpeed := velocity * 1.5
	if velocity < 20 {
		// when you are going slow, it's easy to double how fast you are going
		// you will need to be pulling some pretty high G's here
		maxSpeed = (velocity + 2) * 10
	}

	expectedMaxDistance := maxSpeed * interval.Seconds()

	if expectedMaxDistance > upperRangeLimit {
		expectedMaxDistance = upperRangeLimit
	}

	return expectedMaxDistance
}
