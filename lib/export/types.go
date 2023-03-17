package export

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	// Updates Contains the last updated timestamps for their related fields
	Updates struct {
		Location     time.Time
		Altitude     time.Time
		Velocity     time.Time
		Heading      time.Time
		OnGround     time.Time
		VerticalRate time.Time
		FlightStatus time.Time
		Special      time.Time
		Squawk       time.Time
	}

	// PlaneLocation is our exported data format. it encodes to JSON
	PlaneLocation struct {
		// This info is populated by the tracker
		New             bool
		Removed         bool
		Icao            string
		Lat             float64
		Lon             float64
		Heading         float64
		Velocity        float64
		Altitude        int
		VerticalRate    int
		AltitudeUnits   string
		FlightStatus    string
		OnGround        bool
		Airframe        string
		AirframeType    string
		HasAltitude     bool
		HasLocation     bool
		HasHeading      bool
		HasOnGround     bool
		HasFlightStatus bool
		HasVerticalRate bool
		HasVelocity     bool
		SourceTag       string
		Squawk          string
		Special         string
		TileLocation    string

		SourceTags      map[string]uint `json:",omitempty"`
		sourceTagsMutex *sync.Mutex

		// TrackedSince is when we first started tracking this aircraft *this time*
		TrackedSince time.Time

		// LastMsg is the last time we heard from this aircraft
		LastMsg time.Time

		// Updates contains the list of individual fields that contain updated time stamps for various fields
		Updates Updates

		SignalRssi *float64

		AircraftWidth  *float32 `json:",omitempty"`
		AircraftLength *float32 `json:",omitempty"`

		// Enrichment Plane data
		IcaoCode        *string `json:",omitempty"`
		Registration    *string `json:",omitempty"`
		TypeCode        *string `json:",omitempty"`
		TypeCodeLong    *string `json:",omitempty"`
		Serial          *string `json:",omitempty"`
		RegisteredOwner *string `json:",omitempty"`
		COFAOwner       *string `json:",omitempty"`
		EngineType      *string `json:",omitempty"`
		FlagCode        *string `json:",omitempty"`

		// Enrichment Route Data
		CallSign  *string   `json:",omitempty"`
		Operator  *string   `json:",omitempty"`
		RouteCode *string   `json:",omitempty"`
		Segments  []Segment `json:",omitempty"`
	}

	Segment struct {
		Name     string
		ICAOCode string
	}
)

var (
	ErrImpossible = errors.New("impossible location")
)

// Plane here gives us something to look at
func (pl *PlaneLocation) Plane() string {
	if nil != pl.CallSign && "" != *pl.CallSign {
		return *pl.CallSign
	}

	if nil != pl.Registration && "" != *pl.Registration {
		return *pl.Registration
	}

	return "ICAO: " + pl.Icao
}

func unPtr[t any](what *t) t {
	var def t
	if nil == what {
		return def
	}
	return *what
}

func ptr[t any](what t) *t {
	return &what
}

func MergePlaneLocations(prev, next PlaneLocation) (PlaneLocation, error) {
	if !IsLocationPossible(prev, next) {
		return prev, ErrImpossible
	}
	merged := prev
	merged.New = false
	merged.Removed = false
	merged.LastMsg = next.LastMsg
	merged.SignalRssi = nil // makes no sense to merge this value as it is for the individual receiver
	if nil == merged.sourceTagsMutex {
		merged.sourceTagsMutex = &sync.Mutex{}
	}
	merged.sourceTagsMutex.Lock()
	if nil == merged.SourceTags {
		merged.SourceTags = make(map[string]uint)
	}
	merged.SourceTags[next.SourceTag]++
	merged.sourceTagsMutex.Unlock()

	if next.TrackedSince.Before(prev.TrackedSince) {
		merged.TrackedSince = next.TrackedSince
	}

	if next.HasLocation && next.Updates.Location.After(prev.Updates.Location) {
		merged.Lat = next.Lat
		merged.Lon = next.Lon
		merged.Updates.Location = next.Updates.Location
		merged.HasLocation = true
	}
	if next.HasHeading && next.Updates.Heading.After(prev.Updates.Heading) {
		merged.Heading = next.Heading
		merged.Updates.Heading = prev.Updates.Heading
		merged.HasHeading = true
	}
	if next.HasVelocity && next.Updates.Velocity.After(prev.Updates.Velocity) {
		merged.Velocity = next.Velocity
		merged.Updates.Velocity = next.Updates.Velocity
		merged.HasVelocity = true
	}
	if next.HasAltitude && next.Updates.Altitude.After(prev.Updates.Altitude) {
		merged.Altitude = next.Altitude
		merged.AltitudeUnits = next.AltitudeUnits
		merged.Updates.Altitude = next.Updates.Altitude
		merged.HasAltitude = true
	}
	if next.HasVerticalRate && next.Updates.VerticalRate.After(prev.Updates.VerticalRate) {
		merged.VerticalRate = next.VerticalRate
		merged.Updates.VerticalRate = next.Updates.VerticalRate
		merged.HasVerticalRate = true
	}
	if next.HasFlightStatus && next.Updates.FlightStatus.After(prev.Updates.FlightStatus) {
		merged.FlightStatus = next.FlightStatus
		merged.Updates.FlightStatus = next.Updates.FlightStatus
	}
	if next.HasOnGround && next.Updates.OnGround.After(prev.Updates.OnGround) {
		merged.OnGround = next.OnGround
		merged.Updates.OnGround = next.Updates.OnGround
	}
	if "" == merged.Airframe {
		merged.Airframe = next.Airframe
	}
	if "" == merged.AirframeType {
		merged.AirframeType = next.AirframeType
	}

	if "" != unPtr(next.Registration) {
		merged.Registration = ptr(unPtr(next.Registration))
	}
	if "" != unPtr(next.CallSign) {
		merged.CallSign = ptr(unPtr(next.CallSign))
	}
	merged.SourceTag = "merged"

	if next.Updates.Squawk.After(prev.Updates.Squawk) {
		if `0` == next.Squawk {
			// setting 0 as the squawk is valid, just when we have badly timed data it can jump around
			// only update to 0 *if* it's been a few seconds to account for delayed feeds
			if next.Updates.Squawk.After(prev.Updates.Squawk.Add(5 * time.Second)) {
				merged.Squawk = next.Squawk
				merged.Updates.Squawk = next.Updates.Squawk
			}
		} else {
			merged.Squawk = next.Squawk
			merged.Updates.Squawk = next.Updates.Squawk
		}
	}

	if next.Updates.Special.After(prev.Updates.Special) {
		merged.Special = next.Special
		merged.Updates.Special = next.Updates.Special
	}

	if "" != next.TileLocation {
		merged.TileLocation = next.TileLocation
	}

	if 0 != unPtr(next.AircraftWidth) {
		merged.AircraftWidth = ptr(unPtr(next.AircraftWidth))
	}
	if 0 != unPtr(next.AircraftLength) {
		merged.AircraftLength = ptr(unPtr(next.AircraftLength))
	}

	return merged, nil
}

func IsLocationPossible(prev, next PlaneLocation) bool {
	// simple check, if bearing of prev -> next is more than +-90 degrees of reported value, it is invalid
	if !(prev.HasLocation && next.HasLocation && prev.HasHeading && next.HasHeading) {
		// cannot check, fail open
		if log.Trace().Enabled() {
			log.Info().Str("CallSign", unPtr(next.CallSign)).
				Bool("prevHasLocation", prev.HasLocation).
				Bool("nextHasLocation", next.HasLocation).
				Bool("prevHasHeading", prev.HasHeading).
				Bool("nextHasHeading", next.HasHeading).
				Str("SourceNext", next.SourceTag).
				Str("SourcePrev", prev.SourceTag).
				Msg("Blindly accepting lack of data.")
		}
		return true
	}

	// position hasn't changed, don't check it.
	if prev.Lat == next.Lat && prev.Lon == next.Lon {
		// fail open
		if log.Trace().Enabled() {
			log.Info().Str("CallSign", unPtr(next.CallSign)).
				Str("SourceNext", next.SourceTag).
				Str("SourcePrev", prev.SourceTag).
				Msg("Previous = Next Lat Lon.")
		}

		return true
	}
	// check the timestamp against the last one we saw.
	if prev.LastMsg.After(next.LastMsg) {
		if log.Trace().Enabled() {
			log.Info().Str("CallSign", unPtr(next.CallSign)).
				Time("prev", prev.LastMsg).
				Time("next", next.LastMsg).
				Str("SourceNext", next.SourceTag).
				Str("SourcePrev", prev.SourceTag).
				Msg("Rejecting due to timestamp.")
		}

		return false
	}
	// if plane is on the ground, don't check
	// if next.HasOnGround && next.OnGround {
	// 	return true
	// }

	// if prev.LastMsg.Add(10 * time.Second).After(next.LastMsg) {
	// 	// outside of this time, we cannot accurately use heading
	// 	return true
	// }

	piDegToRad := math.Pi / 180

	radLat0 := prev.Lat * piDegToRad
	radLon0 := prev.Lon * piDegToRad

	radLat1 := next.Lat * piDegToRad
	radLon1 := next.Lon * piDegToRad

	y := math.Sin(radLon1-radLon0) * math.Cos(radLat1)
	x := math.Cos(radLat0)*math.Sin(radLat1) - math.Sin(radLat0)*math.Cos(radLat1)*math.Cos(radLon1-radLon0)

	ret := math.Atan2(y, x)

	bearing := math.Mod(ret*(180.0/math.Pi)+360.0, 360)
	deltaBearing := prev.Heading - bearing
	absDeltaBearing := math.Abs(math.Mod(deltaBearing+180, 360) - 180)

	if absDeltaBearing < 90 { //don't make this less than ~45 degrees, otherwise it'll be inaccurate due to possible wind.
		if log.Trace().Enabled() {
			log.Trace().Str("CallSign", unPtr(next.CallSign)).
				Str("Next", fmt.Sprintf("(%f,%f)", next.Lat, next.Lon)).
				Str("Previous", fmt.Sprintf("(%f,%f)", prev.Lat, prev.Lon)).
				Float64("Heading", prev.Heading).
				Float64("Bearing", bearing).
				Float64("DeltaTheta", absDeltaBearing).
				Str("SourceNext", next.SourceTag).
				Str("SourcePrev", prev.SourceTag).
				Msg("Checked Heading vs Bearing")
		}

		return true
	}

	if log.Trace().Enabled() {
		log.Info().Str("CallSign", unPtr(next.CallSign)).
			Str("Next", fmt.Sprintf("(%f,%f)", next.Lat, next.Lon)).
			Str("Previous", fmt.Sprintf("(%f,%f)", prev.Lat, prev.Lon)).
			Float64("Bearing", bearing).
			Float64("Heading", prev.Heading).
			Float64("DeltaTheta", absDeltaBearing).
			Str("SourceNext", next.SourceTag).
			Str("SourcePrev", prev.SourceTag).
			Msg("Rejected Position")
	}
	return false
}
