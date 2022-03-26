package ws_protocol

import "plane.watch/lib/export"

const (
	WsProtocolPlanes = "planes"

	RequestTypeSubscribe     = "sub"
	RequestTypeSubscribeList = "sub-list"
	RequestTypeUnsubscribe   = "unsub"
	RequestTypeGridPlanes    = "grid-planes" // returns the current plane locations in grid

	ResponseTypeError          = "error"
	ResponseTypeAckSub         = "ack-sub"
	ResponseTypeAckUnsub       = "ack-unsub"
	ResponseTypeSubTiles       = "sub-list"
	ResponseTypePlaneLocation  = "plane-location"
	ResponseTypePlaneLocations = "plane-location-list"
)

type (
	WsRequest struct {
		Type     string `json:"type"`
		GridTile string `json:"gridTile"`
	}
	WsResponse struct {
		Type      string                  `json:"type"`
		Message   string                  `json:"message,omitempty"`
		Tiles     []string                `json:"tiles,omitempty"`
		Location  *export.PlaneLocation   `json:"location,omitempty"`
		Locations []*export.PlaneLocation `json:"locations,omitempty"`
	}
)
