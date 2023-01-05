package export

import "time"

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
		New, Removed      bool
		Icao              string
		Lat, Lon, Heading float64
		Velocity          float64
		Altitude          int
		VerticalRate      int
		AltitudeUnits     string
		FlightStatus      string
		OnGround          bool
		Airframe          string
		AirframeType      string
		HasAltitude       bool
		HasLocation       bool
		HasHeading        bool
		HasVerticalRate   bool
		HasVelocity       bool
		HasOnGround       bool
		HasFlightStatus   bool
		SourceTag         string
		Squawk            string
		Special           string
		TileLocation      string

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
		Serial          *string `json:",omitempty"`
		RegisteredOwner *string `json:",omitempty"`
		COFAOwner       *string `json:",omitempty"`
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

func MergePlaneLocations(current, next PlaneLocation) PlaneLocation {
	merged := current
	merged.New = false
	merged.Removed = false
	merged.LastMsg = next.LastMsg
	merged.SignalRssi = nil // makes no sense to merge this value as it is for the individual receiver
	if next.TrackedSince.Before(current.TrackedSince) {
		merged.TrackedSince = next.TrackedSince
	}

	if next.HasLocation && next.Updates.Location.After(current.Updates.Location) {
		merged.Lat = next.Lat
		merged.Lon = next.Lon
		merged.Updates.Location = next.Updates.Location
		merged.HasLocation = true
	}
	if next.HasHeading && next.Updates.Heading.After(current.Updates.Heading) {
		merged.Heading = next.Heading
		merged.Updates.Heading = current.Updates.Heading
		merged.HasHeading = true
	}
	if next.HasVelocity && next.Updates.Velocity.After(current.Updates.Velocity) {
		merged.Velocity = next.Velocity
		merged.Updates.Velocity = next.Updates.Velocity
		merged.HasVelocity = true
	}
	if next.HasAltitude && next.Updates.Altitude.After(current.Updates.Altitude) {
		merged.Altitude = next.Altitude
		merged.AltitudeUnits = next.AltitudeUnits
		merged.Updates.Altitude = next.Updates.Altitude
		merged.HasAltitude = true
	}
	if next.HasVerticalRate && next.Updates.VerticalRate.After(current.Updates.VerticalRate) {
		merged.VerticalRate = next.VerticalRate
		merged.Updates.VerticalRate = next.Updates.VerticalRate
		merged.HasVerticalRate = true
	}
	if next.HasFlightStatus && next.Updates.FlightStatus.After(current.Updates.FlightStatus) {
		merged.FlightStatus = next.FlightStatus
		merged.Updates.FlightStatus = next.Updates.FlightStatus
	}
	if next.HasOnGround && next.Updates.OnGround.After(current.Updates.OnGround) {
		merged.OnGround = next.OnGround
		merged.Updates.OnGround = next.Updates.OnGround
	}
	if "" == merged.Airframe {
		merged.Airframe = next.Airframe
	}
	if "" == merged.AirframeType {
		merged.Airframe = next.AirframeType
	}

	if "" != unPtr(next.Registration) {
		merged.Registration = ptr(unPtr(next.Registration))
	}
	if "" != unPtr(next.CallSign) {
		merged.CallSign = ptr(unPtr(next.CallSign))
	}
	// TODO: in the future we probably want a list of sources that contributed to this data
	merged.SourceTag = "merged"

	if next.Updates.Squawk.After(current.Updates.Squawk) {
		merged.Squawk = next.Squawk
		merged.Updates.Squawk = next.Updates.Squawk
	}

	if next.Updates.Special.After(current.Updates.Special) {
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

	return merged
}
