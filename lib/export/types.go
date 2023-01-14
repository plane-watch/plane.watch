package export

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

type (
	// PlaneLocation is our exported data format. it encodes to JSON
	PlaneLocation struct {
		PlaneLocationPB
		sourceTagsMutex *sync.Mutex
	}
)

var (
	ErrImpossible = errors.New("impossible location")
)

// Plane here gives us something to look at
func (pl *PlaneLocation) Plane() string {
	if "" != pl.CallSign {
		return pl.CallSign
	}

	if "" != pl.Registration {
		return pl.Registration
	}

	return fmt.Sprintf("ICAO: %d", pl.Icao)
}

func (at AirframeType) Describe() string {
	switch at {
	case AirframeType_CODE_00_NO_ADSB:
		return "o ADS-B Emitter Category Information"
	case AirframeType_CODE_01_LIGHT:
		return "Light (< 15500 lbs)"
	case AirframeType_CODE_02_SMALL:
		return "Small (15500 to 75000 lbs)"
	case AirframeType_CODE_03_LARGE:
		return "Large (75000 to 300000 lbs)"
	case AirframeType_CODE_04_HIGH_VORTEX:
		return "High Vortex Large (aircraft such as B-757)"
	case AirframeType_CODE_05_HEAVY:
		return "Heavy (> 300000 lbs)"
	case AirframeType_CODE_06_HIGH_PERFORMANCE:
		return "High Performance (> 5g acceleration and 400 kts)"
	case AirframeType_CODE_07_ROTORCRAFT:
		return "Rotorcraft"
	case AirframeType_CODE_10_NO_ADSB:
		return "No ADS-B Emitter Category Information"
	case AirframeType_CODE_11_GLIDER_SAILPLANE:
		return "Glider / sailplane"
	case AirframeType_CODE_12_LIGHTER_THAN_AIR:
		return "Lighter-than-air"
	case AirframeType_CODE_13_PARACHUTIST:
		return "Parachutist / Skydiver"
	case AirframeType_CODE_14_ULTRALIGHT:
		return "Ultralight / hang-glider / paraglider"
	case AirframeType_CODE_15_RESERVED:
		return "Reserved"
	case AirframeType_CODE_16_UAV:
		return "Unmanned Aerial Vehicle"
	case AirframeType_CODE_17_TRANS_ATMO:
		return "Space / Trans-atmospheric vehicle"
	case AirframeType_CODE_20_NO_ADSB:
		return "No ADS-B Emitter Category Information"
	case AirframeType_CODE_21_SURFACE_EMERGENCY_VEHICLE:
		return "Surface Vehicle – Emergency Vehicle"
	case AirframeType_CODE_22_SURFACE_VEHICLE:
		return "Surface Vehicle – Service Vehicle"
	case AirframeType_CODE_23_POINT_OBSTACLE:
		return "Point Obstacle (includes tethered balloons)"
	case AirframeType_CODE_24_CLUSTER_OBSTACLE:
		return "Cluster Obstacle"
	case AirframeType_CODE_25_LINE_OBSTACLE:
		return "Line Obstacle"
	case AirframeType_CODE_26_RESERVED:
		return "Reserved"
	case AirframeType_CODE_27_RESERVED:
		return "Reserved"
	case AirframeType_CODE_30_RESERVED:
		return "Reserved"
	case AirframeType_CODE_31_RESERVED:
		return "Reserved"
	case AirframeType_CODE_32_RESERVED:
		return "Reserved"
	case AirframeType_CODE_33_RESERVED:
		return "Reserved"
	case AirframeType_CODE_34_RESERVED:
		return "Reserved"
	case AirframeType_CODE_35_RESERVED:
		return "Reserved"
	case AirframeType_CODE_36_RESERVED:
		return "Reserved"
	case AirframeType_CODE_37_RESERVED:
		return "Reserved"
	case AirframeType_UNKNOWN:
		return "Unknown"
	default:
		return "Unknown"
	}
}

func (au AltitudeUnits) Describe() string {
	switch au {
	case AltitudeUnits_FEET:
		return "feet"
	case AltitudeUnits_METRES:
		return "metres"
	default:
		return "unknown"
	}
}

func (fs FlightStatus) Describe() string {
	switch fs {
	case FlightStatus_NormalAirborne:
		return "Normal, Airborne"
	case FlightStatus_NormalOnGround:
		return "Normal, On Ground"
	case FlightStatus_AlertAirborne:
		return "Alert, Airborne"
	case FlightStatus_AlertOnGround:
		return "Alert, On Ground"
	case FlightStatus_AlertSpecialPositionId:
		return "Alert"
	case FlightStatus_NormalSpecialPositionId:
		return ""
	case FlightStatus_Value6Unassigned:
		return ""
	case FlightStatus_Value7Unassigned:
		return ""
	default:
		return "unknown"
	}
}

func unPtr[t any](what *t) t {
	var def t
	if nil == what {
		return def
	}
	return *what
}

//func ptr[t any](what t) *t {
//	return &what
//}

func MergePlaneLocations(prev, next *PlaneLocation) (*PlaneLocation, error) {
	if !IsLocationPossible(prev, next) {
		return prev, ErrImpossible
	}
	merged := prev
	merged.LastMsg = next.LastMsg
	merged.SignalRssi = 0 // makes no sense to merge this value as it is for the individual receiver
	if nil == merged.sourceTagsMutex {
		merged.sourceTagsMutex = &sync.Mutex{}
	}
	merged.IncSourceTag(next.SourceTag)

	if next.TrackedSince.AsTime().Before(prev.TrackedSince.AsTime()) {
		merged.TrackedSince = next.TrackedSince
	}

	if next.HasLocation && next.Updates.Location.AsTime().After(prev.Updates.Location.AsTime()) {
		merged.Lat = next.Lat
		merged.Lon = next.Lon
		merged.Updates.Location = next.Updates.Location
		merged.HasLocation = true
	}
	if next.HasHeading && next.Updates.Heading.AsTime().After(prev.Updates.Heading.AsTime()) {
		merged.Heading = next.Heading
		merged.Updates.Heading = prev.Updates.Heading
		merged.HasHeading = true
	}
	if next.HasVelocity && next.Updates.Velocity.AsTime().After(prev.Updates.Velocity.AsTime()) {
		merged.Velocity = next.Velocity
		merged.Updates.Velocity = next.Updates.Velocity
		merged.HasVelocity = true
	}
	if next.HasAltitude && next.Updates.Altitude.AsTime().After(prev.Updates.Altitude.AsTime()) {
		merged.Altitude = next.Altitude
		merged.AltitudeUnits = next.AltitudeUnits
		merged.Updates.Altitude = next.Updates.Altitude
		merged.HasAltitude = true
	}
	if next.HasVerticalRate && next.Updates.VerticalRate.AsTime().After(prev.Updates.VerticalRate.AsTime()) {
		merged.VerticalRate = next.VerticalRate
		merged.Updates.VerticalRate = next.Updates.VerticalRate
		merged.HasVerticalRate = true
	}
	if next.HasFlightStatus && next.Updates.FlightStatus.AsTime().After(prev.Updates.FlightStatus.AsTime()) {
		merged.FlightStatus = next.FlightStatus
		merged.Updates.FlightStatus = next.Updates.FlightStatus
	}
	if next.HasOnGround && next.Updates.OnGround.AsTime().After(prev.Updates.OnGround.AsTime()) {
		merged.OnGround = next.OnGround
		merged.Updates.OnGround = next.Updates.OnGround
	}
	if AirframeType_UNKNOWN != merged.AirframeType {
		merged.AirframeType = next.AirframeType
	}

	if "" != next.Registration {
		merged.Registration = next.Registration
	}
	if "" != next.CallSign {
		merged.CallSign = next.CallSign
	}
	merged.SourceTag = "merged"

	if next.Updates.Squawk.AsTime().After(prev.Updates.Squawk.AsTime()) {
		if 0 == next.Squawk {
			// setting 0 as the squawk is valid, just when we have badly timed data it can jump around
			// only update to 0 *if* it's been a few seconds to account for delayed feeds
			if next.Updates.Squawk.AsTime().After(prev.Updates.Squawk.AsTime().Add(5 * time.Second)) {
				merged.Squawk = next.Squawk
				merged.Updates.Squawk = next.Updates.Squawk
			}
		} else {
			merged.Squawk = next.Squawk
			merged.Updates.Squawk = next.Updates.Squawk
		}
	}

	if next.Updates.Special.AsTime().After(prev.Updates.Special.AsTime()) {
		merged.Special = next.Special
		merged.Updates.Special = next.Updates.Special
	}

	if "" != next.TileLocation {
		merged.TileLocation = next.TileLocation
	}

	if 0 != next.AircraftWidth {
		merged.AircraftWidth = next.AircraftWidth
	}
	if 0 != next.AircraftLength {
		merged.AircraftLength = next.AircraftLength
	}

	return merged, nil
}

func IsLocationPossible(prev, next *PlaneLocation) bool {
	if nil == prev || nil == next {
		return false
	}
	// simple check, if bearing of prev -> next is more than +-90 degrees of reported value, it is invalid
	if !(prev.HasLocation && next.HasLocation && prev.HasHeading && next.HasHeading) {
		// cannot check, fail open
		return true
	}
	if prev.LastMsg.AsTime().After(next.LastMsg.AsTime()) {
		return false
	}
	if prev.LastMsg.AsTime().Add(3 * time.Second).After(next.LastMsg.AsTime()) {
		// outside of this time, we cannot accurately use heading
		return true
	}

	piDegToRad := math.Pi / 180

	radLat0 := prev.Lat * piDegToRad
	radLon0 := prev.Lon * piDegToRad

	radLat1 := next.Lat * piDegToRad
	radLon1 := next.Lon * piDegToRad

	y := math.Sin(radLon1-radLon0) * math.Cos(radLat1)
	x := math.Cos(radLat0)*math.Sin(radLat1) - math.Sin(radLat0)*math.Cos(radLat1)*math.Cos(radLon1-radLon0)

	ret := math.Atan2(y, x)

	bearing := math.Mod(ret*(180.0/math.Pi)+360.0, 360)

	min := prev.Heading - 90
	max := prev.Heading + 90

	if bearing > min && bearing < max {
		return true
	}

	return false
}
