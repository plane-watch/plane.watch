package sbs1

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	sbsMsgTypeField      = 0
	sbsMsgSubCatField    = 1
	sbsIcaoField         = 4
	sbsRecvDate          = 6
	sbsRecvTime          = 7
	sbsCallsignField     = 10
	sbsAltitudeField     = 11
	sbsGroundSpeedField  = 12
	sbsTrackField        = 13
	sbsLatField          = 14
	sbsLonField          = 15
	sbsVerticalRateField = 16
	sbsSquawkField       = 17
	sbsAlertSquawkField  = 18
	sbsEmergencyField    = 19
	sbsSpiIdentField     = 20
	sbsOnGroundField     = 21
)

type Frame struct {
	// original is our unadulterated string
	MsgType      string
	original     string
	icaoStr      string
	IcaoInt      uint32
	Received     time.Time
	CallSign     string
	Altitude     int
	GroundSpeed  int
	Track        float64
	Lat, Lon     float64
	VerticalRate int
	Squawk       string
	Alert        string
	Emergency    string
	OnGround     bool

	HasPosition bool
}

func NewFrame(sbsString string) *Frame {
	return &Frame{
		original: sbsString,
	}
}

func (f *Frame) TimeStamp() time.Time {
	return f.Received
}

func (f *Frame) Parse() error {
	// decode the string
	var err error

	bits := strings.Split(f.original, ",")
	if len(bits) != 22 {
		return fmt.Errorf("Failed to Parse Input - not enough parameters: %s", f.original)
	}

	f.icaoStr = bits[sbsIcaoField]
	f.IcaoInt, err = icaoStringToInt(bits[sbsIcaoField])
	if nil != err {
		return err
	}
	sTime := bits[sbsRecvDate] + " " + bits[sbsRecvTime]
	//2016/06/03 00:00:38.350
	f.Received, err = time.Parse("2006/01/02 15:04:05.999999999", sTime)
	if nil != err {
		f.Received = time.Now()
	}

	f.MsgType = bits[sbsMsgTypeField]

	switch bits[sbsMsgTypeField] { // message type
	case "SEL": // SELECTION_CHANGE
		f.CallSign = bits[sbsCallsignField]
	case "ID": // NEW_ID
		f.CallSign = bits[sbsCallsignField]
	case "AIR": // NEW_AIRCRAFT - just indicates when a new aircraft pops up
	case "STA": // STATUS_AIRCRAFT
	// call sign field (10) contains one of:
	//	PL (Position Lost)
	// 	SL (Signal Lost)
	// 	RM (Remove)
	// 	AD (Delete)
	// 	OK (used to reset time-outs if aircraft returns into cover).
	case "CLK": // CLICK
	case "MSG": // TRANSMISSION
		switch bits[sbsMsgSubCatField] {
		case "1": // ES Identification and Category
			f.CallSign = bits[sbsCallsignField]

		case "2": // ES Surface Position Message
			f.Altitude, _ = strconv.Atoi(bits[sbsAltitudeField])
			f.GroundSpeed, _ = strconv.Atoi(bits[sbsGroundSpeedField])
			f.Track, _ = strconv.ParseFloat(bits[sbsTrackField], 32)
			f.Lat, _ = strconv.ParseFloat(bits[sbsLatField], 32)
			f.Lon, _ = strconv.ParseFloat(bits[sbsLonField], 32)
			f.HasPosition = true
			f.OnGround = "-1" == bits[sbsOnGroundField]

		case "3": // ES Airborne Position Message
			f.Altitude, _ = strconv.Atoi(bits[sbsAltitudeField])
			f.Lat, _ = strconv.ParseFloat(bits[sbsLatField], 32)
			f.Lon, _ = strconv.ParseFloat(bits[sbsLonField], 32)
			f.HasPosition = true
			f.Alert = bits[sbsAlertSquawkField]
			f.Emergency = bits[sbsEmergencyField]
			f.OnGround = "-1" == bits[sbsOnGroundField]
		//SPI Flag Ignored

		case "4": // ES Airborne velocity Message
			f.GroundSpeed, _ = strconv.Atoi(bits[sbsGroundSpeedField])
			f.Track, _ = strconv.ParseFloat(bits[sbsTrackField], 32)
			f.VerticalRate, _ = strconv.Atoi(bits[sbsVerticalRateField])
			f.OnGround = "-1" == bits[sbsOnGroundField]

		case "5": // Surveillance Alt Message
			f.Altitude, _ = strconv.Atoi(bits[sbsAltitudeField])
			f.Alert = bits[sbsAlertSquawkField]
			f.OnGround = "-1" == bits[sbsOnGroundField]
			f.CallSign = bits[sbsCallsignField]
		//SPI Flag Ignored

		case "6": // Surveillance ID Message
			f.CallSign = bits[sbsCallsignField]
			f.Altitude, _ = strconv.Atoi(bits[sbsAltitudeField])
			f.Squawk = bits[sbsSquawkField]
			f.Alert = bits[sbsAlertSquawkField]
			f.Emergency = bits[sbsEmergencyField]
			f.OnGround = "-1" == bits[sbsOnGroundField]
		//SPI Flag Ignored

		case "7": //Air To Air Message
			f.Altitude, _ = strconv.Atoi(bits[sbsAltitudeField])
			f.OnGround = "-1" == bits[sbsOnGroundField]

		case "8": // All Call Reply
			f.OnGround = "-1" == bits[sbsOnGroundField]
		}
	}

	return nil
}

func icaoStringToInt(icao string) (uint32, error) {
	btoi, err := hex.DecodeString(icao)
	if nil != err {
		return 0, fmt.Errorf("Failed to decode ICAO HEX (%s) into UINT32. %s", icao, err)
	}
	return uint32(btoi[0])<<16 | uint32(btoi[1])<<8 | uint32(btoi[2]), nil
}

func (f *Frame) Icao() uint32 {
	if nil == f {
		return 0
	}
	return f.IcaoInt
}
func (f *Frame) IcaoStr() string {
	return f.icaoStr
}

func (f *Frame) Decode() error {
	return f.Parse()
}

func (f *Frame) Raw() []byte {
	return []byte(f.original)
}
