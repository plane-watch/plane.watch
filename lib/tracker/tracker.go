package tracker

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"plane.watch/lib/dedupe/forgetfulmap"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
)

type (
	Tracker struct {
		planeList *forgetfulmap.ForgetfulSyncMap

		// pruneTick is how long between pruning attempts
		// pruneAfter is how long we wait from the last message before we remove it from the tracker
		pruneTick, pruneAfter time.Duration

		// Input Handling
		producers   []Producer
		middlewares []Middleware
		sinks       []Sink

		producerWaiter   sync.WaitGroup
		middlewareWaiter sync.WaitGroup

		decodeWorkerCount   int
		decodingQueue       chan *FrameEvent
		decodingQueueWaiter sync.WaitGroup

		eventSync    sync.RWMutex
		eventsOpen   bool
		finishDone   bool
		events       chan Event
		eventsWaiter sync.WaitGroup

		startTime time.Time

		stats struct {
			currentPlanes prometheus.Gauge
			decodedFrames prometheus.Counter
		}

		log zerolog.Logger
	}
)

// NewTracker creates a new tracker with which we can populate with plane tracking data
func NewTracker(opts ...Option) *Tracker {
	t := &Tracker{
		producers:         []Producer{},
		middlewares:       []Middleware{},
		decodeWorkerCount: 5,
		pruneTick:         10 * time.Second,
		pruneAfter:        5 * time.Minute,
		decodingQueue:     make(chan *FrameEvent, 1000), // a nice deep buffer
		events:            make(chan Event, 10000),
		eventsOpen:        true,

		startTime: time.Now(),

		log: log.With().Str("Section", "Tracker").Logger(),
	}

	for _, opt := range opts {
		opt(t)
	}

	t.planeList = forgetfulmap.NewForgetfulSyncMap(
		forgetfulmap.WithSweepInterval(t.pruneTick),
		forgetfulmap.WithPreEvictionAction(func(key, value interface{}) {
			if nil != t.stats.currentPlanes {
				t.stats.currentPlanes.Dec()
			}

			if plane, ok := value.(*Plane); ok {
				// now send an event
				t.AddEvent(newPlaneActionEvent(plane, false, true))
			}
			if t.log.Trace().Enabled() {
				t.log.Trace().
					Str("ICAO", fmt.Sprintf("%06X", key)).
					Msg("Plane is being evicted")
			}
		}),
		forgetfulmap.WithForgettableAction(func(key, value any, added time.Time) bool {
			result := true
			if plane, ok := value.(*Plane); ok {
				oldest := time.Now().Add(-t.pruneAfter)
				// remove the plane from the list if it is older than our oldest allowable
				result = plane.LastSeen().Before(oldest)
			}
			if t.log.Trace().Enabled() {
				t.log.Trace().
					Bool("Removing?", result).
					Str("ICAO", fmt.Sprintf("%06X", key)).
					Msg("Should plane be forgotten?")
			}
			// remove anything that is not a *export.PlaneLocation
			return result
		}),
	)

	// Process our event queue and send them to all the Sinks that are currently listening to us
	go t.processEvents()

	t.decodingQueueWaiter.Add(t.decodeWorkerCount)
	for i := 0; i < t.decodeWorkerCount; i++ {
		go t.decodeQueue()
	}

	return t
}

func (t *Tracker) numPlanes() int {
	return int(t.planeList.Len())
}

func (t *Tracker) GetPlane(icao uint32) *Plane {
	plane, ok := t.planeList.Load(icao)
	if ok {
		return plane.(*Plane)
	}
	if t.log.Trace().Enabled() {
		t.log.Trace().
			Str("ICAO", fmt.Sprintf("%06X", icao)).
			Msg("Plane has made an appearance")
	}
	if nil != t.stats.currentPlanes {
		t.stats.currentPlanes.Inc()
	}

	p := newPlane(icao)
	p.tracker = t
	t.planeList.Store(icao, p)
	return p
}

func (t *Tracker) EachPlane(pi PlaneIterator) {
	t.planeList.Range(func(key, value interface{}) bool {
		return pi(value.(*Plane))
	})
}

func (p *Plane) HandleModeSFrame(frame *mode_s.Frame, refLat, refLon *float64) {
	if nil == frame {
		return
	}
	icao := frame.Icao()
	if 0 == icao {
		return
	}
	var planeFormat string
	var hasChanged bool

	p.setLastSeen(frame.TimeStamp())
	p.incMsgCount()

	debugMessage := func(sfmt string, a ...interface{}) {
		if zerolog.GlobalLevel() >= zerolog.DebugLevel {
			planeFormat = fmt.Sprintf("DF%02d - \033[0;97mPlane (\033[38;5;118m%s %-8s\033[0;97m)", frame.DownLinkType(), p.IcaoIdentifierStr(), p.FlightNumber())
			p.tracker.log.Debug().Msgf(planeFormat+sfmt, a...)

		}
	}

	hasChanged = p.setRegistration(frame.DecodeAuIcaoRegistration()) || hasChanged

	if log.Trace().Enabled() {
		log.Trace().
			Str("frame", frame.String()).
			Str("icao", frame.IcaoStr()).
			Str("Downlink Type", "DF"+strconv.Itoa(int(frame.DownLinkType()))).
			Int("Downlink Format", int(frame.DownLinkType())).
			Str("DF17 Msg Type", frame.MessageTypeString()).
			Bytes("RAW", frame.Raw()).
			Send()
	}

	// determine what to do with our given frame
	switch frame.DownLinkType() {
	case 0:
		// grab the altitude
		if frame.AltitudeValid() {
			alt, _ := frame.Altitude()
			hasChanged = p.setAltitude(alt, frame.AltitudeUnits()) || hasChanged
		}
		if frame.VerticalStatusValid() {
			hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
		}
		p.setLocationUpdateTime(frame.TimeStamp())
		debugMessage(" is at %d %s \033[0m", p.Altitude(), p.AltitudeUnits())

	case 1, 2, 3:
		if frame.VerticalStatusValid() {
			hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
		}
		p.setLocationUpdateTime(frame.TimeStamp())
		if frame.Alert() {
			hasChanged = p.setSpecial("alert", "Alert") || hasChanged
		}
	case 6, 7, 8, 9, 10, 12, 13, 14, 15, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31:
		debugMessage(" \033[38;5;52mIgnoring Mode S Frame: %d (%s)\033[0m\n", frame.DownLinkType(), frame.DownLinkFormat())
		break
	case 11:
		if frame.VerticalStatusValid() {
			hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
		}
	case 4, 5:
		if frame.VerticalStatusValid() {
			hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
		}
		if frame.Alert() {
			hasChanged = p.setSpecial("alert", "Alert") || hasChanged
		}
		if frame.AltitudeValid() {
			alt, _ := frame.Altitude()
			hasChanged = p.setAltitude(alt, frame.AltitudeUnits()) || hasChanged
		}
		hasChanged = p.setFlightStatus(frame.FlightStatus(), frame.FlightStatusString()) || hasChanged

		if 5 == frame.DownLinkType() { // || 21 == frame.DownLinkType()
			hasChanged = p.setSquawkIdentity(frame.SquawkIdentity()) || hasChanged
		}

		p.setLocationUpdateTime(frame.TimeStamp())

		debugMessage(" is at %d %s and flight status is: %s. \033[2mMode S Frame: %d \033[0m",
			p.Altitude(), p.AltitudeUnits(), p.FlightStatus(), frame.DownLinkType())
		break
	case 16:
		if frame.AltitudeValid() {
			alt, _ := frame.Altitude()
			hasChanged = p.setAltitude(alt, frame.AltitudeUnits()) || hasChanged
		}
		if frame.VerticalStatusValid() {
			hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
		}
		p.setLocationUpdateTime(frame.TimeStamp())

	case 17, 18, 19: // ADS-B
		//if debug {
		//	frame.Describe(os.Stdout)
		//}
		// i am using the text version because it is easier to program with.
		// if performance is an issue, change over to byte comparing
		messageType := frame.MessageTypeString()
		switch messageType {
		case mode_s.DF17FrameIdCat: // "Aircraft Identification and Category"
			{
				hasChanged = p.setFlightNumber(frame.FlightNumber()) || hasChanged
				if frame.ValidCategory() {
					hasChanged = p.setAirFrameCategory(frame.Category()) || hasChanged
					hasChanged = p.setAirFrameCategoryType(frame.CategoryType()) || hasChanged
				}
				break
			}
		case mode_s.DF17FrameSurfacePos: // "Surface Position"
			{
				if frame.HeadingValid() {
					hasChanged = p.setHeading(frame.MustHeading()) || hasChanged
				}
				if frame.VelocityValid() {
					hasChanged = p.setVelocity(frame.MustVelocity()) || hasChanged
				}
				if frame.VerticalStatusValid() {
					hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
				}

				if frame.IsEven() {
					_ = p.setCprEvenLocation(float64(frame.Latitude()), float64(frame.Longitude()), frame.TimeStamp())
				} else {
					_ = p.setCprOddLocation(float64(frame.Latitude()), float64(frame.Longitude()), frame.TimeStamp())
				}
				if err := p.decodeCprFilledRefLatLon(refLat, refLon, frame.TimeStamp()); nil != err {
					debugMessage("%s", err)
				} else {
					hasChanged = true
				}
				p.setLocationUpdateTime(frame.TimeStamp())

				debugMessage(" is on the ground and has heading %s and is travelling at %0.2f knots\033[0m", p.HeadingStr(), p.Velocity())
				break
			}
		case mode_s.DF17FrameAirPositionBarometric, mode_s.DF17FrameAirPositionGnss: // "Airborne Position (with Barometric altitude)"
			{
				if frame.VerticalStatusValid() {
					hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
				}
				p.setLocationUpdateTime(frame.TimeStamp())

				if frame.IsEven() {
					_ = p.setCprEvenLocation(float64(frame.Latitude()), float64(frame.Longitude()), frame.TimeStamp())
				} else {
					_ = p.setCprOddLocation(float64(frame.Latitude()), float64(frame.Longitude()), frame.TimeStamp())
				}

				altitude, _ := frame.Altitude()
				p.setAltitude(altitude, frame.AltitudeUnits())
				if err := p.decodeCpr(0, 0, frame.TimeStamp()); nil != err {
					debugMessage("%s", err)
				} else {
					hasChanged = true
				}

				if dt := p.DistanceTravelled(); dt.Valid() {
					debugMessage(" travelled %0.2f metres %0.2f seconds", dt.metres, dt.duration)
				}

				if frame.HasSurveillanceStatus() {
					hasChanged = p.setSpecial("surveillance", frame.SurveillanceStatus()) || hasChanged
				} else {
					hasChanged = p.setSpecial("surveillance", "") || hasChanged
				}

				break
			}
		case mode_s.DF17FrameAirVelocity: // "Airborne velocity"
			{
				if frame.HeadingValid() {
					hasChanged = p.setHeading(frame.MustHeading()) || hasChanged
				}
				if frame.VelocityValid() {
					hasChanged = p.setVelocity(frame.MustVelocity()) || hasChanged
				}
				if frame.VerticalStatusValid() {
					hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
				}
				if frame.VerticalRateValid() {
					hasChanged = p.setVerticalRate(frame.MustVerticalRate()) || hasChanged
				}
				p.setLocationUpdateTime(frame.TimeStamp())

				headingStr := "unknown heading"
				if p.HasHeading() {
					headingStr = fmt.Sprintf("heading %0.2f", p.Heading())
				}
				debugMessage(" has %s and is travelling at %0.2f knots\033[0m", headingStr, p.Velocity())
				break
			}
		case mode_s.DF17FrameTestMessage: //, "Test Message":
			debugMessage("\033[2m Ignoring: DF%d %s\033[0m", frame.DownLinkType(), messageType)
			break
		case mode_s.DF17FrameTestMessageSquawk: //, "Test Message":
			{
				if frame.SquawkIdentity() > 0 {
					hasChanged = p.setSquawkIdentity(frame.SquawkIdentity()) || hasChanged
				}
				break
			}
		case mode_s.DF17FrameSurfaceSystemStatus: //, "Surface System status":
			{
				debugMessage("\033[2m Ignoring: DF%d %s\033[0m", frame.DownLinkType(), messageType)
				break
			}
		case mode_s.DF17FrameEmergencyPriority: //, "Extended Squitter Aircraft status (Emergency)":
			{
				debugMessage("\033[2m %s\033[0m", messageType)
				if frame.Alert() {
					hasChanged = p.setSpecial("special", frame.Special()) || hasChanged
					hasChanged = p.setSpecial("emergency", frame.Emergency()) || hasChanged
				}
				hasChanged = p.setSquawkIdentity(frame.SquawkIdentity()) || hasChanged
				break
			}
		case mode_s.DF17FrameTcasRA: //, "Extended Squitter Aircraft status (1090ES TCAS RA)":
			{
				debugMessage("\033[2m Ignoring: DF%d %s\033[0m", frame.DownLinkType(), messageType)
				break
			}
		case mode_s.DF17FrameTargetStateStatus: //, "Target State and status Message":
			{
				debugMessage("\033[2m Ignoring: DF%d %s\033[0m", frame.DownLinkType(), messageType)
				break
			}
		case mode_s.DF17FrameAircraftOperational: //, "Aircraft Operational status Message":
			{
				if frame.VerticalStatusValid() {
					hasChanged = p.setGroundStatus(frame.MustOnGround()) || hasChanged
				}
				hasChanged = p.setAirFrameWidthLength(frame.GetAirplaneLengthWidth()) || hasChanged

				break
			}
		}

	case 20, 21:
		switch frame.BdsMessageType() {
		case mode_s.BdsElsDataLinkCap: // 1.0
			hasChanged = p.setSquawkIdentity(frame.SquawkIdentity()) || hasChanged
		case mode_s.BdsElsGicbCap: // 1.7
			if frame.AltitudeValid() {
				hasChanged = p.setAltitude(frame.MustAltitude(), frame.AltitudeUnits()) || hasChanged
			}
		case mode_s.BdsElsAircraftIdent: // 2.0
			hasChanged = p.setFlightNumber(frame.FlightNumber()) || hasChanged
		default:
			// let's see if we can decode more BDS info
			// TODO: Decode Other BDS frames
		}
	}

	if hasChanged {
		p.tracker.AddEvent(newPlaneLocationEvent(p))
	}
}

func (p *Plane) HandleSbs1Frame(frame *sbs1.Frame) {
	var hasChanged bool
	p.setLastSeen(frame.TimeStamp())
	p.incMsgCount()
	if frame.HasPosition {
		if err := p.addLatLong(frame.Lat, frame.Lon, frame.Received); nil != err {
			p.tracker.log.Warn().Err(err).Send()
		}

		hasChanged = true
		p.tracker.log.Debug().Msgf("Plane %s is at %0.4f, %0.4f", frame.IcaoStr(), frame.Lat, frame.Lon)
	}

	if hasChanged {
		p.tracker.AddEvent(newPlaneLocationEvent(p))
	}
}
