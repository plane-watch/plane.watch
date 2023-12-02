package export

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"sync"
	"time"
)

type (
	// PlaneAndLocationInfoMsg is our exported data format. it is our wrapper around the Protobuf msg so we can perform locking
	PlaneAndLocationInfoMsg struct {
		*PlaneAndLocationInfo
		icaoStr         string
		sourceTagsMutex *sync.Mutex
	}
)

func NewPlaneAndLocationInfoMsg() *PlaneAndLocationInfoMsg {
	return &PlaneAndLocationInfoMsg{
		PlaneAndLocationInfo: &PlaneAndLocationInfo{
			Icao:            0,
			Lat:             0,
			Lon:             0,
			Heading:         0,
			Velocity:        0,
			Altitude:        0,
			VerticalRate:    0,
			AltitudeUnits:   0,
			FlightStatus:    0,
			OnGround:        false,
			AirframeType:    0,
			HasAltitude:     false,
			HasLocation:     false,
			HasHeading:      false,
			HasOnGround:     false,
			HasFlightStatus: false,
			HasVerticalRate: false,
			HasVelocity:     false,
			SourceTag:       "",
			Squawk:          0,
			Special:         "",
			TileLocation:    "",
			SourceTags:      make(map[string]uint32),
			TrackedSince:    timestamppb.New(time.Now()),
			LastMsg:         timestamppb.New(time.Now()),
			Updates: &FieldUpdates{
				Location:     timestamppb.New(time.Now()),
				Altitude:     timestamppb.New(time.Now()),
				Velocity:     timestamppb.New(time.Now()),
				Heading:      timestamppb.New(time.Now()),
				OnGround:     timestamppb.New(time.Now()),
				VerticalRate: timestamppb.New(time.Now()),
				FlightStatus: timestamppb.New(time.Now()),
				Special:      timestamppb.New(time.Now()),
				Squawk:       timestamppb.New(time.Now()),
			},
			SignalRssi:      0,
			AircraftWidth:   0,
			AircraftLength:  0,
			Registration:    "",
			TypeCode:        "",
			TypeCodeLong:    "",
			Serial:          "",
			RegisteredOwner: "",
			COFAOwner:       "",
			EngineType:      "",
			FlagCode:        "",
			CallSign:        "",
			Operator:        "",
			RouteCode:       "",
			Segments:        make([]*RouteSegment, 2),
		},
		icaoStr:         "",
		sourceTagsMutex: &sync.Mutex{},
	}
}

// Plane here gives us something to look at
func (pl *PlaneAndLocationInfoMsg) Plane() string {
	if pl.CallSign != "" {
		return pl.CallSign
	}

	if pl.Registration != "" {
		return pl.Registration
	}

	if pl.icaoStr == "" {
		pl.icaoStr = fmt.Sprintf("ICAO: %d", pl.Icao)
	}

	return pl.icaoStr
}

func (at AirframeType) Code() string {
	if code, ok := airframeTypeReverseLookup[at]; ok {
		return code
	}
	return ""
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

func (x AltitudeUnits) Describe() string {
	switch x {
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

func (pl *PlaneAndLocationInfoMsg) CloneSourceTags() map[string]uint32 {
	pl.sourceTagsMutex.Lock()
	defer pl.sourceTagsMutex.Unlock()

	return Clone(pl.SourceTags)
}

func (pl *PlaneAndLocationInfoMsg) IcaoStr() string {
	if pl.icaoStr == "" {
		pl.icaoStr = fmt.Sprintf("%06X", pl.Icao)
	}
	return pl.icaoStr
}

func (pl *PlaneAndLocationInfoMsg) SquawkStr() string {
	if nil == pl {
		return ""
	}
	if pl.Squawk == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(pl.Squawk), 10)
}

func (pl *PlaneAndLocationInfoMsg) CallSignStr() string {
	if nil == pl {
		return ""
	}
	return pl.CallSign
}

func (pl *PlaneAndLocationInfoMsg) LatStr() string {
	if nil == pl {
		return ""
	}
	if !pl.HasLocation {
		return ""
	}
	return strconv.FormatFloat(pl.Lat, 'f', 4, 64)
}

func (pl *PlaneAndLocationInfoMsg) LonStr() string {
	if nil == pl {
		return ""
	}
	if !pl.HasLocation {
		return ""
	}
	return strconv.FormatFloat(pl.Lon, 'f', 4, 64)
}

func (pl *PlaneAndLocationInfoMsg) AltitudeStr() string {
	if nil == pl {
		return ""
	}
	if !pl.HasAltitude {
		return ""
	}
	return fmt.Sprintf("%d %s", pl.Altitude, pl.AltitudeUnits)
}

func (pl *PlaneAndLocationInfoMsg) VerticalRateStr() string {
	if nil == pl {
		return ""
	}
	if !pl.HasVerticalRate {
		return ""
	}
	return strconv.FormatInt(int64(pl.VerticalRate), 10)
}

func (pl *PlaneAndLocationInfoMsg) HeadingStr() string {
	if nil == pl {
		return ""
	}
	if !pl.HasHeading {
		return ""
	}
	return strconv.FormatFloat(pl.Heading, 'f', 1, 64)
}

func (pli *PlaneAndLocationInfo) AsPlaneAndLocationInfoMsg() *PlaneAndLocationInfoMsg {
	return &PlaneAndLocationInfoMsg{
		PlaneAndLocationInfo: pli,
		icaoStr:              "",
		sourceTagsMutex:      &sync.Mutex{},
	}
}
