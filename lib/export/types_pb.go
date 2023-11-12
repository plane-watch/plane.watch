package export

import (
	"fmt"
	"sync"
)

type (
	// PlaneLocationPBMsg is our exported data format. it encodes to JSON
	PlaneLocationPBMsg struct {
		PlaneLocationPB
		icaoStr         string
		sourceTagsMutex *sync.Mutex
	}
)

// Plane here gives us something to look at
func (pl *PlaneLocationPBMsg) Plane() string {
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
		return "No ADS-B Emitter Category Information"
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

func (pl *PlaneLocationPBMsg) CloneSourceTags() map[string]uint32 {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()

	return Clone(pl.SourceTags)
}

func (pl *PlaneLocationPBMsg) IcaoStr() string {
	if "" == pl.icaoStr {
		pl.icaoStr = fmt.Sprintf("%X", pl.Icao)
	}
	return pl.icaoStr
}
