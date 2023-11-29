// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: plane.watch.websocket.proto

package ws_protocol

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	export "plane.watch/lib/export"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WebsocketRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     string `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	GridTile string `protobuf:"bytes,2,opt,name=GridTile,proto3" json:"GridTile,omitempty"`
	Icao     string `protobuf:"bytes,3,opt,name=Icao,proto3" json:"Icao,omitempty"`
	CallSign string `protobuf:"bytes,4,opt,name=CallSign,proto3" json:"CallSign,omitempty"`
	Tick     string `protobuf:"bytes,5,opt,name=Tick,proto3" json:"Tick,omitempty"`
	Query    string `protobuf:"bytes,6,opt,name=Query,proto3" json:"Query,omitempty"`
}

func (x *WebsocketRequest) Reset() {
	*x = WebsocketRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebsocketRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebsocketRequest) ProtoMessage() {}

func (x *WebsocketRequest) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebsocketRequest.ProtoReflect.Descriptor instead.
func (*WebsocketRequest) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{0}
}

func (x *WebsocketRequest) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *WebsocketRequest) GetGridTile() string {
	if x != nil {
		return x.GridTile
	}
	return ""
}

func (x *WebsocketRequest) GetIcao() string {
	if x != nil {
		return x.Icao
	}
	return ""
}

func (x *WebsocketRequest) GetCallSign() string {
	if x != nil {
		return x.CallSign
	}
	return ""
}

func (x *WebsocketRequest) GetTick() string {
	if x != nil {
		return x.Tick
	}
	return ""
}

func (x *WebsocketRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

type AircraftTrail struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Point []*AircraftTrailPoint `protobuf:"bytes,1,rep,name=Point,proto3" json:"Point,omitempty"`
}

func (x *AircraftTrail) Reset() {
	*x = AircraftTrail{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AircraftTrail) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AircraftTrail) ProtoMessage() {}

func (x *AircraftTrail) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AircraftTrail.ProtoReflect.Descriptor instead.
func (*AircraftTrail) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{1}
}

func (x *AircraftTrail) GetPoint() []*AircraftTrailPoint {
	if x != nil {
		return x.Point
	}
	return nil
}

type AircraftTrailPoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Lat      float64 `protobuf:"fixed64,1,opt,name=Lat,proto3" json:"Lat,omitempty"`
	Lon      float64 `protobuf:"fixed64,2,opt,name=Lon,proto3" json:"Lon,omitempty"`
	Heading  float64 `protobuf:"fixed64,3,opt,name=Heading,proto3" json:"Heading,omitempty"`
	Velocity float64 `protobuf:"fixed64,4,opt,name=Velocity,proto3" json:"Velocity,omitempty"`
	Altitude *int32  `protobuf:"varint,5,opt,name=Altitude,proto3,oneof" json:"Altitude,omitempty"`
}

func (x *AircraftTrailPoint) Reset() {
	*x = AircraftTrailPoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AircraftTrailPoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AircraftTrailPoint) ProtoMessage() {}

func (x *AircraftTrailPoint) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AircraftTrailPoint.ProtoReflect.Descriptor instead.
func (*AircraftTrailPoint) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{2}
}

func (x *AircraftTrailPoint) GetLat() float64 {
	if x != nil {
		return x.Lat
	}
	return 0
}

func (x *AircraftTrailPoint) GetLon() float64 {
	if x != nil {
		return x.Lon
	}
	return 0
}

func (x *AircraftTrailPoint) GetHeading() float64 {
	if x != nil {
		return x.Heading
	}
	return 0
}

func (x *AircraftTrailPoint) GetVelocity() float64 {
	if x != nil {
		return x.Velocity
	}
	return 0
}

func (x *AircraftTrailPoint) GetAltitude() int32 {
	if x != nil && x.Altitude != nil {
		return *x.Altitude
	}
	return 0
}

type AircraftListPB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Aircraft []*export.PlaneAndLocationInfo `protobuf:"bytes,1,rep,name=Aircraft,proto3" json:"Aircraft,omitempty"`
}

func (x *AircraftListPB) Reset() {
	*x = AircraftListPB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AircraftListPB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AircraftListPB) ProtoMessage() {}

func (x *AircraftListPB) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AircraftListPB.ProtoReflect.Descriptor instead.
func (*AircraftListPB) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{3}
}

func (x *AircraftListPB) GetAircraft() []*export.PlaneAndLocationInfo {
	if x != nil {
		return x.Aircraft
	}
	return nil
}

type SearchResultPB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query    string               `protobuf:"bytes,1,opt,name=Query,proto3" json:"Query,omitempty"`
	Aircraft *AircraftListPB      `protobuf:"bytes,2,opt,name=Aircraft,proto3" json:"Aircraft,omitempty"`
	Airport  []*AirportLocationPB `protobuf:"bytes,3,rep,name=Airport,proto3" json:"Airport,omitempty"`
	Route    []string             `protobuf:"bytes,4,rep,name=Route,proto3" json:"Route,omitempty"`
}

func (x *SearchResultPB) Reset() {
	*x = SearchResultPB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchResultPB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResultPB) ProtoMessage() {}

func (x *SearchResultPB) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResultPB.ProtoReflect.Descriptor instead.
func (*SearchResultPB) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{4}
}

func (x *SearchResultPB) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *SearchResultPB) GetAircraft() *AircraftListPB {
	if x != nil {
		return x.Aircraft
	}
	return nil
}

func (x *SearchResultPB) GetAirport() []*AirportLocationPB {
	if x != nil {
		return x.Airport
	}
	return nil
}

func (x *SearchResultPB) GetRoute() []string {
	if x != nil {
		return x.Route
	}
	return nil
}

type AirportLocationPB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string  `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Icao string  `protobuf:"bytes,2,opt,name=Icao,proto3" json:"Icao,omitempty"`
	Iata string  `protobuf:"bytes,3,opt,name=Iata,proto3" json:"Iata,omitempty"`
	Lat  float64 `protobuf:"fixed64,4,opt,name=Lat,proto3" json:"Lat,omitempty"`
	Lon  float64 `protobuf:"fixed64,5,opt,name=Lon,proto3" json:"Lon,omitempty"`
}

func (x *AirportLocationPB) Reset() {
	*x = AirportLocationPB{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AirportLocationPB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AirportLocationPB) ProtoMessage() {}

func (x *AirportLocationPB) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AirportLocationPB.ProtoReflect.Descriptor instead.
func (*AirportLocationPB) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{5}
}

func (x *AirportLocationPB) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AirportLocationPB) GetIcao() string {
	if x != nil {
		return x.Icao
	}
	return ""
}

func (x *AirportLocationPB) GetIata() string {
	if x != nil {
		return x.Iata
	}
	return ""
}

func (x *AirportLocationPB) GetLat() float64 {
	if x != nil {
		return x.Lat
	}
	return 0
}

func (x *AirportLocationPB) GetLon() float64 {
	if x != nil {
		return x.Lon
	}
	return 0
}

type PlaneAndLocationInfoList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Location []*export.PlaneAndLocationInfo `protobuf:"bytes,1,rep,name=Location,proto3" json:"Location,omitempty"`
}

func (x *PlaneAndLocationInfoList) Reset() {
	*x = PlaneAndLocationInfoList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PlaneAndLocationInfoList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlaneAndLocationInfoList) ProtoMessage() {}

func (x *PlaneAndLocationInfoList) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlaneAndLocationInfoList.ProtoReflect.Descriptor instead.
func (*PlaneAndLocationInfoList) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{6}
}

func (x *PlaneAndLocationInfoList) GetLocation() []*export.PlaneAndLocationInfo {
	if x != nil {
		return x.Location
	}
	return nil
}

type WebSocketResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      string                       `protobuf:"bytes,1,opt,name=Type,proto3" json:"Type,omitempty"`
	Message   string                       `protobuf:"bytes,2,opt,name=Message,proto3" json:"Message,omitempty"`
	Tiles     []string                     `protobuf:"bytes,3,rep,name=Tiles,proto3" json:"Tiles,omitempty"`
	Location  *export.PlaneAndLocationInfo `protobuf:"bytes,4,opt,name=Location,proto3,oneof" json:"Location,omitempty"`
	Locations *PlaneAndLocationInfoList    `protobuf:"bytes,5,opt,name=Locations,proto3,oneof" json:"Locations,omitempty"`
	Icao      string                       `protobuf:"bytes,6,opt,name=Icao,proto3" json:"Icao,omitempty"`
	CallSign  *string                      `protobuf:"bytes,7,opt,name=CallSign,proto3,oneof" json:"CallSign,omitempty"`
	History   *AircraftTrail               `protobuf:"bytes,8,opt,name=History,proto3,oneof" json:"History,omitempty"`
	Results   *SearchResultPB              `protobuf:"bytes,9,opt,name=Results,proto3,oneof" json:"Results,omitempty"`
}

func (x *WebSocketResponse) Reset() {
	*x = WebSocketResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plane_watch_websocket_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebSocketResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebSocketResponse) ProtoMessage() {}

func (x *WebSocketResponse) ProtoReflect() protoreflect.Message {
	mi := &file_plane_watch_websocket_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebSocketResponse.ProtoReflect.Descriptor instead.
func (*WebSocketResponse) Descriptor() ([]byte, []int) {
	return file_plane_watch_websocket_proto_rawDescGZIP(), []int{7}
}

func (x *WebSocketResponse) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *WebSocketResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *WebSocketResponse) GetTiles() []string {
	if x != nil {
		return x.Tiles
	}
	return nil
}

func (x *WebSocketResponse) GetLocation() *export.PlaneAndLocationInfo {
	if x != nil {
		return x.Location
	}
	return nil
}

func (x *WebSocketResponse) GetLocations() *PlaneAndLocationInfoList {
	if x != nil {
		return x.Locations
	}
	return nil
}

func (x *WebSocketResponse) GetIcao() string {
	if x != nil {
		return x.Icao
	}
	return ""
}

func (x *WebSocketResponse) GetCallSign() string {
	if x != nil && x.CallSign != nil {
		return *x.CallSign
	}
	return ""
}

func (x *WebSocketResponse) GetHistory() *AircraftTrail {
	if x != nil {
		return x.History
	}
	return nil
}

func (x *WebSocketResponse) GetResults() *SearchResultPB {
	if x != nil {
		return x.Results
	}
	return nil
}

var File_plane_watch_websocket_proto protoreflect.FileDescriptor

var file_plane_watch_websocket_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x77, 0x65,
	0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x70,
	0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73, 0x1a, 0x1d, 0x70,
	0x6c, 0x61, 0x6e, 0x65, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9c, 0x01, 0x0a,
	0x10, 0x57, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x47, 0x72, 0x69, 0x64, 0x54, 0x69, 0x6c,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x47, 0x72, 0x69, 0x64, 0x54, 0x69, 0x6c,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x49, 0x63, 0x61, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x49, 0x63, 0x61, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x43, 0x61, 0x6c, 0x6c, 0x53, 0x69, 0x67,
	0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x43, 0x61, 0x6c, 0x6c, 0x53, 0x69, 0x67,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x69, 0x63, 0x6b, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x54, 0x69, 0x63, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x22, 0x49, 0x0a, 0x0d, 0x41,
	0x69, 0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6c, 0x12, 0x38, 0x0a, 0x05,
	0x50, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x70, 0x6c,
	0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73, 0x2e, 0x41, 0x69, 0x72,
	0x63, 0x72, 0x61, 0x66, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6c, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x52,
	0x05, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x22, 0x9c, 0x01, 0x0a, 0x12, 0x41, 0x69, 0x72, 0x63, 0x72,
	0x61, 0x66, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6c, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x10, 0x0a,
	0x03, 0x4c, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x4c, 0x61, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x4c, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x4c, 0x6f,
	0x6e, 0x12, 0x18, 0x0a, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x56,
	0x65, 0x6c, 0x6f, 0x63, 0x69, 0x74, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x56,
	0x65, 0x6c, 0x6f, 0x63, 0x69, 0x74, 0x79, 0x12, 0x1f, 0x0a, 0x08, 0x41, 0x6c, 0x74, 0x69, 0x74,
	0x75, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x08, 0x41, 0x6c, 0x74,
	0x69, 0x74, 0x75, 0x64, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x41, 0x6c, 0x74,
	0x69, 0x74, 0x75, 0x64, 0x65, 0x22, 0x4f, 0x0a, 0x0e, 0x41, 0x69, 0x72, 0x63, 0x72, 0x61, 0x66,
	0x74, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x42, 0x12, 0x3d, 0x0a, 0x08, 0x41, 0x69, 0x72, 0x63, 0x72,
	0x61, 0x66, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x6c, 0x61, 0x6e,
	0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x50, 0x6c, 0x61, 0x6e, 0x65, 0x41, 0x6e, 0x64,
	0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x41, 0x69,
	0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x22, 0xb5, 0x01, 0x0a, 0x0e, 0x53, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x50, 0x42, 0x12, 0x14, 0x0a, 0x05, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x3a, 0x0a, 0x08, 0x41, 0x69, 0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1e, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f,
	0x77, 0x73, 0x2e, 0x41, 0x69, 0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x50,
	0x42, 0x52, 0x08, 0x41, 0x69, 0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x12, 0x3b, 0x0a, 0x07, 0x41,
	0x69, 0x72, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70,
	0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73, 0x2e, 0x41, 0x69,
	0x72, 0x70, 0x6f, 0x72, 0x74, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x42, 0x52,
	0x07, 0x41, 0x69, 0x72, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x52, 0x6f, 0x75, 0x74,
	0x65, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x22, 0x73,
	0x0a, 0x11, 0x41, 0x69, 0x72, 0x70, 0x6f, 0x72, 0x74, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x50, 0x42, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x49, 0x63, 0x61, 0x6f, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x49, 0x63, 0x61, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x49,
	0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x49, 0x61, 0x74, 0x61, 0x12,
	0x10, 0x0a, 0x03, 0x4c, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x4c, 0x61,
	0x74, 0x12, 0x10, 0x0a, 0x03, 0x4c, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03,
	0x4c, 0x6f, 0x6e, 0x22, 0x59, 0x0a, 0x18, 0x50, 0x6c, 0x61, 0x6e, 0x65, 0x41, 0x6e, 0x64, 0x4c,
	0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x4c, 0x69, 0x73, 0x74, 0x12,
	0x3d, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x21, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e,
	0x50, 0x6c, 0x61, 0x6e, 0x65, 0x41, 0x6e, 0x64, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xda,
	0x03, 0x0a, 0x11, 0x57, 0x65, 0x62, 0x53, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x05, 0x54, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x42, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x6c, 0x61,
	0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x50, 0x6c, 0x61, 0x6e, 0x65, 0x41, 0x6e,
	0x64, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x48, 0x00, 0x52,
	0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x4b, 0x0a, 0x09,
	0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x28, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73,
	0x2e, 0x50, 0x6c, 0x61, 0x6e, 0x65, 0x41, 0x6e, 0x64, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x4c, 0x69, 0x73, 0x74, 0x48, 0x01, 0x52, 0x09, 0x4c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x88, 0x01, 0x01, 0x12, 0x12, 0x0a, 0x04, 0x49, 0x63, 0x61,
	0x6f, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x49, 0x63, 0x61, 0x6f, 0x12, 0x1f, 0x0a,
	0x08, 0x43, 0x61, 0x6c, 0x6c, 0x53, 0x69, 0x67, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x02, 0x52, 0x08, 0x43, 0x61, 0x6c, 0x6c, 0x53, 0x69, 0x67, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x3c,
	0x0a, 0x07, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1d, 0x2e, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73,
	0x2e, 0x41, 0x69, 0x72, 0x63, 0x72, 0x61, 0x66, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6c, 0x48, 0x03,
	0x52, 0x07, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x88, 0x01, 0x01, 0x12, 0x3d, 0x0a, 0x07,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x70, 0x6c, 0x61, 0x6e, 0x65, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x77, 0x73, 0x2e, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x50, 0x42, 0x48, 0x04, 0x52,
	0x07, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x88, 0x01, 0x01, 0x42, 0x0b, 0x0a, 0x09, 0x5f,
	0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x4c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x43, 0x61, 0x6c, 0x6c, 0x53,
	0x69, 0x67, 0x6e, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42,
	0x0a, 0x0a, 0x08, 0x5f, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x42, 0x1d, 0x5a, 0x1b, 0x70,
	0x6c, 0x61, 0x6e, 0x65, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2f, 0x6c, 0x69, 0x62, 0x2f, 0x77,
	0x73, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_plane_watch_websocket_proto_rawDescOnce sync.Once
	file_plane_watch_websocket_proto_rawDescData = file_plane_watch_websocket_proto_rawDesc
)

func file_plane_watch_websocket_proto_rawDescGZIP() []byte {
	file_plane_watch_websocket_proto_rawDescOnce.Do(func() {
		file_plane_watch_websocket_proto_rawDescData = protoimpl.X.CompressGZIP(file_plane_watch_websocket_proto_rawDescData)
	})
	return file_plane_watch_websocket_proto_rawDescData
}

var file_plane_watch_websocket_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_plane_watch_websocket_proto_goTypes = []interface{}{
	(*WebsocketRequest)(nil),            // 0: plane_watch_ws.WebsocketRequest
	(*AircraftTrail)(nil),               // 1: plane_watch_ws.AircraftTrail
	(*AircraftTrailPoint)(nil),          // 2: plane_watch_ws.AircraftTrailPoint
	(*AircraftListPB)(nil),              // 3: plane_watch_ws.AircraftListPB
	(*SearchResultPB)(nil),              // 4: plane_watch_ws.SearchResultPB
	(*AirportLocationPB)(nil),           // 5: plane_watch_ws.AirportLocationPB
	(*PlaneAndLocationInfoList)(nil),    // 6: plane_watch_ws.PlaneAndLocationInfoList
	(*WebSocketResponse)(nil),           // 7: plane_watch_ws.WebSocketResponse
	(*export.PlaneAndLocationInfo)(nil), // 8: plane_watch.PlaneAndLocationInfo
}
var file_plane_watch_websocket_proto_depIdxs = []int32{
	2, // 0: plane_watch_ws.AircraftTrail.Point:type_name -> plane_watch_ws.AircraftTrailPoint
	8, // 1: plane_watch_ws.AircraftListPB.Aircraft:type_name -> plane_watch.PlaneAndLocationInfo
	3, // 2: plane_watch_ws.SearchResultPB.Aircraft:type_name -> plane_watch_ws.AircraftListPB
	5, // 3: plane_watch_ws.SearchResultPB.Airport:type_name -> plane_watch_ws.AirportLocationPB
	8, // 4: plane_watch_ws.PlaneAndLocationInfoList.Location:type_name -> plane_watch.PlaneAndLocationInfo
	8, // 5: plane_watch_ws.WebSocketResponse.Location:type_name -> plane_watch.PlaneAndLocationInfo
	6, // 6: plane_watch_ws.WebSocketResponse.Locations:type_name -> plane_watch_ws.PlaneAndLocationInfoList
	1, // 7: plane_watch_ws.WebSocketResponse.History:type_name -> plane_watch_ws.AircraftTrail
	4, // 8: plane_watch_ws.WebSocketResponse.Results:type_name -> plane_watch_ws.SearchResultPB
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_plane_watch_websocket_proto_init() }
func file_plane_watch_websocket_proto_init() {
	if File_plane_watch_websocket_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plane_watch_websocket_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebsocketRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AircraftTrail); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AircraftTrailPoint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AircraftListPB); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchResultPB); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AirportLocationPB); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PlaneAndLocationInfoList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plane_watch_websocket_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebSocketResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_plane_watch_websocket_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_plane_watch_websocket_proto_msgTypes[7].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_plane_watch_websocket_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_plane_watch_websocket_proto_goTypes,
		DependencyIndexes: file_plane_watch_websocket_proto_depIdxs,
		MessageInfos:      file_plane_watch_websocket_proto_msgTypes,
	}.Build()
	File_plane_watch_websocket_proto = out.File
	file_plane_watch_websocket_proto_rawDesc = nil
	file_plane_watch_websocket_proto_goTypes = nil
	file_plane_watch_websocket_proto_depIdxs = nil
}
