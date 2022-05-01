package ws_protocol

import "plane.watch/lib/export"

const (
	WsProtocolPlanes = "planes"

	RequestTypeSubscribe       = "sub"
	RequestTypeSubscribeList   = "sub-list"
	RequestTypeUnsubscribe     = "unsub"
	RequestTypeGridPlanes      = "grid-planes"            // returns the current plane locations in grid
	RequestTypePlaneLocHistory = "plane-location-history" // returns the requested planes path

	ResponseTypeError           = "error"
	ResponseTypeAckSub          = "ack-sub"
	ResponseTypeAckUnsub        = "ack-unsub"
	ResponseTypeSubTiles        = "sub-list"
	ResponseTypePlaneLocation   = "plane-location"
	ResponseTypePlaneLocations  = "plane-location-list"
	ResponseTypePlaneLocHistory = "plane-location-history"
)

type (
	WsRequest struct {
		Type     string `json:"type"`
		GridTile string `json:"gridTile"`
		Icao     string `json:"icao,omitempty"`
		CallSign string `json:"callSign,omitempty"`
	}
	LocationHistory struct {
		Lat, Lon          float64
		Heading, Velocity float64
		Altitude          int
	}
	WsResponse struct {
		Type      string                  `json:"type"`
		Message   string                  `json:"message,omitempty"`
		Tiles     []string                `json:"tiles,omitempty"`
		Location  *export.PlaneLocation   `json:"location,omitempty"`
		Locations *[]export.PlaneLocation `json:"locations,omitempty"`
		History   *[]LocationHistory      `json:"history,omitempty"`
	}
)
