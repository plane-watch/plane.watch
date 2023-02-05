package calc

import (
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog/log"
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

	Geforce1 = 9.8 // in Metres/Second
	Geforce2 = 9.8 * 2
	Geforce3 = 9.8 * 3
	Geforce4 = 9.8 * 4
	Geforce5 = 9.8 * 5
)

// Distance function returns the distance (in meters) between two points of
//
//	a given longitude and latitude relatively accurately (using a spherical
//	approximation of the Earth) through the Haversin Distance Formula for
//	great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METRES!!!!!!
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

// AccelerationBetween is a better way of getting acceleration between two points
// Respects the direction of flight to determine the acceleration.
// Returns the acceleration in m/s^2
func AccelerationBetween(time1, time2 time.Time, la1, lo1, heading1, velocity1, la2, lo2 float64) float64 {
	r := 6378100.0      // Earth radius in metres
	ktsToMs := 0.514444 // knots to m/s

	// convert the velocity from kts to m/s.
	// split into it's 2d components - this allows us to take the heading into consideration
	// when calculating the accelleration.
	vXref := velocity1 * ktsToMs * math.Sin(heading1*math.Pi/180)
	vYref := velocity1 * ktsToMs * math.Cos(heading1*math.Pi/180)

	// determine the displacement between the two points in it's 2d components.
	// assumes the earth is flat given how smaller distances we're using
	// should be within 1%
	deltaX := (2 * math.Pi * r) / 360 * (lo2 - lo1) * math.Cos(la1*math.Pi/180)
	deltaY := (2 * math.Pi * r) / 360 * (la2 - la1)

	// difference in time between the two frames
	deltaT := time2.Sub(time1)

	// calculated component velocities
	vX := deltaX / deltaT.Seconds()
	vY := deltaY / deltaT.Seconds()

	// calculated component accellerations
	aX := (vX - vXref) / deltaT.Seconds()
	aY := (vY - vYref) / deltaT.Seconds()

	// combined accelleration
	a := math.Sqrt(math.Pow(aX, 2) + math.Pow(aY, 2))

	return a
}

// hsin is the haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// MaxAllowableDistance calculates how far a plane could have travelled
// interval is the interval between
func MaxAllowableDistance(interval time.Duration, velocity float64) float64 {
	if interval.Seconds() <= 0.001 { // this is a really quick report
		interval = 1
	}

	// SR71A's official was Mach 3.3, not much else gets that fast in atmosphere
	// make sure Velocity is between 1m/s and Mach4
	velocity = math.Max(1, math.Min(Mach4, velocity))

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

func FlightLocationValid(prevTime, currentTime time.Time, prevVelocityKnots, prevLat, prevLon, prevHeading, currentLat, currentLon float64) bool {
	deltaT := currentTime.Sub(prevTime)
	// if these frames are heaps far apart we should probably trust them
	if deltaT.Seconds() > 300 {
		return true
	}
	prevVelocityMS := prevVelocityKnots * 0.514444

	// basic distance calc first, let's see if this location is unreasonably far from our previous
	distance := Distance(prevLat, prevLon, currentLat, currentLon)
	//fmt.Printf("calc distance %0.6f\n", distance)
	currentVelocityMS := distance / deltaT.Seconds()
	//fmt.Printf("calc current velocity %0.6f m/s\n", currentVelocityMS)
	maxAllowedDistance := MaxAllowableDistance(deltaT, math.Max(prevVelocityMS, currentVelocityMS))
	if distance > maxAllowedDistance {
		if log.Trace().Enabled() {
			log.Trace().
				Str("FAIL Coordinates", fmt.Sprintf("(%0.6f,%0.6f) => (%0.6f,%0.6f)", prevLat, prevLon, currentLat, currentLon)).
				Float64("distance", distance).
				Dur("duration", deltaT).
				Float64("m/s", deltaT.Seconds()*distance).
				Msg("Plane has travelled too far")
		}
		return false
	} else {
		if log.Trace().Enabled() {
			log.Trace().
				Str("SUCCESS Coordinates", fmt.Sprintf("(%0.6f,%0.6f) => (%0.6f,%0.6f)", prevLat, prevLon, currentLat, currentLon)).
				Float64("distance", distance).
				Dur("duration", deltaT).
				Float64("m/s", deltaT.Seconds()*distance).
				Msg("Valid distance.")
		}
	}

	accel := AccelerationBetween(prevTime, currentTime, prevLat, prevLon, prevHeading, prevVelocityKnots, currentLat, currentLon)

	if accel > Geforce5 {
		if log.Trace().Enabled() {
			log.Trace().
				Str("FAIL Coordinates", fmt.Sprintf("(%0.6f,%0.6f) => (%0.6f,%0.6f)", prevLat, prevLon, currentLat, currentLon)).
				Float64("Acceleration", accel).
				Float64("distance", distance).
				Dur("duration", deltaT).
				Msg("Plane has accelerated too hard")
		}
		return false
	} else {
		if log.Trace().Enabled() {
			log.Trace().
				Str("SUCCESS Coordinates", fmt.Sprintf("(%0.6f,%0.6f) => (%0.6f,%0.6f)", prevLat, prevLon, currentLat, currentLon)).
				Float64("Acceleration", accel).
				Float64("distance", distance).
				Dur("duration", deltaT).
				Msg("Valid acceleration.")
		}
	}

	return true
}
