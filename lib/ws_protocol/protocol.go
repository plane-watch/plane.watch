package ws_protocol

import "plane.watch/lib/export"

const (
	WsProtocolPlanes = "planes"

	RequestTypeSubscribe       = "sub"
	RequestTypeSubscribeList   = "sub-list"
	RequestTypeUnsubscribe     = "unsub"
	RequestTypeGridPlanes      = "grid-planes"            // returns the current plane locations in grid
	RequestTypePlaneLocHistory = "plane-location-history" // returns the requested planes path
	RequestTypeTickAdjust      = "adjust-tick"            // adjusts how often we send updates
	RequestTypeSearch          = "search"                 // adjusts how often we send updates

	ResponseTypeError           = "error"
	ResponseTypeMsg             = "info"
	ResponseTypeAckSub          = "ack-sub"
	ResponseTypeAckUnsub        = "ack-unsub"
	ResponseTypeSubTiles        = "sub-list"
	ResponseTypePlaneLocation   = "plane-location"
	ResponseTypePlaneLocations  = "plane-location-list"
	ResponseTypePlaneLocHistory = "plane-location-history"
	ResponseTypeSearchResults   = "search-results"
)

type (
	WsRequest struct {
		Type     string `json:"type"`
		GridTile string `json:"gridTile"`
		Icao     string `json:"icao,omitempty"`
		CallSign string `json:"callSign,omitempty"`
		Tick     int    `json:"tick,omitempty"`  // in Milliseconds
		Query    string `json:"query,omitempty"` // in Milliseconds
	}
	LocationHistory struct {
		Lat, Lon          float64
		Heading, Velocity float64
		Altitude          *int32
	}
	AircraftList []*export.PlaneLocation
	SearchResult struct {
		Query    string
		Aircraft AircraftList
		Airport  []AirportLocation
		Route    []string
	}
	AirportLocation struct {
		Name     string
		Icao     string
		Iata     string
		Lat, Lon float64
	}
	WsResponse struct {
		Type      string                  `json:"type"`
		Message   string                  `json:"message,omitempty"`
		Tiles     []string                `json:"tiles,omitempty"`
		Location  *export.PlaneLocation   `json:"location,omitempty"`
		Locations []*export.PlaneLocation `json:"locations,omitempty"`

		Icao     string            `json:"icao,omitempty"`
		CallSign string            `json:"callSign,omitempty"`
		History  []LocationHistory `json:"history,omitempty"`
		Results  *SearchResult     `json:"results,omitempty"`
	}
)

func (l AircraftList) Len() int {
	if nil == l {
		return 0
	}
	return len(l)
}
func unPtr[t any](what *t) t {
	var def t
	if nil == what {
		return def
	}
	return *what
}

func (l AircraftList) Less(i, j int) bool {
	if nil == l {
		return false
	}

	left := unPtr(l[i].CallSign) + ":" + unPtr(l[i].Registration) + ":" + l[i].Icao
	right := unPtr(l[j].CallSign) + ":" + unPtr(l[j].Registration) + ":" + l[j].Icao

	return left < right
}

func (l AircraftList) Swap(i, j int) {
	x := l[i]
	l[i] = l[j]
	l[j] = x
}
