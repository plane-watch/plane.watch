package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"plane.watch/lib/export"
	"plane.watch/lib/middleware"
	"plane.watch/lib/nats_io"
	"plane.watch/lib/randstr"
	"plane.watch/lib/setup"
	"plane.watch/lib/sink"
	"plane.watch/lib/ws_protocol"
	"sync"
	"time"
)

const (
	tapActionAdd    = "add"
	tapActionRemove = "remove"
)

type (
	tapHeaders map[string]string

	wireDecoderFunc func([]byte) *export.PlaneAndLocationInfoMsg
	wsDecoderFunc   func(handlers *[]*wsFilterFunc) func(r *ws_protocol.WsResponse)

	PlaneWatchTapper struct {
		natsSvr       *nats_io.Server
		wsLow, wsHigh *ws_protocol.WsClient

		exitChannels map[string]chan bool
		taps         []tapHeaders

		logger zerolog.Logger

		queueLocations             string
		queueLocationsEnriched     string
		queueLocationsEnrichedLow  string
		queueLocationsEnrichedHigh string

		pwIngestDecoder     wireDecoderFunc
		pwEnrichmentDecoder wireDecoderFunc
		pwRouterDecoder     wireDecoderFunc
		wsProto             string

		wsHandlersLow  []*wsFilterFunc
		wsHandlersHigh []*wsFilterFunc
		mu             sync.Mutex
	}

	wsFilterFunc struct {
		icao      uint32
		feederTag string
		speed     string
		filter    wsHandler
	}

	IngestTapHandler func(frameType, tag string, data []byte)
	wsHandler        func(location *export.PlaneAndLocationInfoMsg)

	Option func(*PlaneWatchTapper)
)

func NewPlaneWatchTapper(opts ...Option) *PlaneWatchTapper {
	pw := &PlaneWatchTapper{
		exitChannels:               make(map[string]chan bool),
		taps:                       make([]tapHeaders, 0),
		logger:                     zerolog.Logger{},
		queueLocations:             sink.QueueLocationUpdates,
		queueLocationsEnriched:     "location-updates-enriched",
		queueLocationsEnrichedLow:  "location-updates-enriched-reduced",
		queueLocationsEnrichedHigh: "location-updates-enriched-merged",

		wsHandlersLow:  make([]*wsFilterFunc, 0),
		wsHandlersHigh: make([]*wsFilterFunc, 0),
	}
	// setup to use JSON wire proto by default
	pw.pwIngestDecoder = pw.decoderJSON
	pw.pwEnrichmentDecoder = pw.decoderJSON
	pw.pwRouterDecoder = pw.decoderJSON
	pw.wsProto = setup.WireProtocolJSON

	for _, opt := range opts {
		opt(pw)
	}

	return pw
}

func WithLogger(logger zerolog.Logger) Option {
	return func(tapper *PlaneWatchTapper) {
		tapper.logger = logger
	}
}

func WithProtocol(protocolType string) Option {
	return func(tapper *PlaneWatchTapper) {
		switch protocolType {
		case setup.WireProtocolJSON:
			tapper.pwIngestDecoder = tapper.decoderJSON
			tapper.pwEnrichmentDecoder = tapper.decoderJSON
			tapper.pwRouterDecoder = tapper.decoderJSON
			tapper.wsProto = setup.WireProtocolJSON
		case setup.WireProtocolProtobuf:
			tapper.pwIngestDecoder = tapper.decoderProtobuf
			tapper.pwEnrichmentDecoder = tapper.decoderProtobuf
			tapper.pwRouterDecoder = tapper.decoderProtobuf
			tapper.wsProto = setup.WireProtocolProtobuf
		default:
			panic(fmt.Sprintf("Unknown wire protocol type, expected on of ['%s', '%s']", setup.WireProtocolJSON, setup.WireProtocolProtobuf))
		}
	}
}

func WithProtocolFor(what, protocolType string) Option {
	return func(tapper *PlaneWatchTapper) {
		if protocolType == "" {
			// NoOp out when we get nothing set
			return
		}
		var wireDecoder wireDecoderFunc

		switch protocolType {
		case setup.WireProtocolJSON:
			wireDecoder = tapper.decoderJSON

		case setup.WireProtocolProtobuf:
			wireDecoder = tapper.decoderProtobuf
		default:
			panic(fmt.Sprintf("Unknown wire protocol type, expected on of ['%s', '%s']", setup.WireProtocolJSON, setup.WireProtocolProtobuf))
		}

		switch what {
		case wireProtocolForIngest:
			tapper.pwIngestDecoder = wireDecoder
		case wireProtocolForEnrichment:
			tapper.pwEnrichmentDecoder = wireDecoder
		case wireProtocolForRouter:
			tapper.pwRouterDecoder = wireDecoder
		case wireProtocolForWsBroker:
			tapper.wsProto = protocolType
		}
	}
}

func (pw *PlaneWatchTapper) Connect(natsServer, wsServer string) error {
	var err error
	pw.natsSvr, err = nats_io.NewServer(
		nats_io.WithConnections(true, true),
		nats_io.WithServer(natsServer, "ingest-nats-tap"),
	)
	if nil != err {
		return err
	}

	pw.wsLow = ws_protocol.NewClient(
		ws_protocol.WithSourceURL(wsServer),
		ws_protocol.WithWireProtocol(pw.wsProto),
		ws_protocol.WithJSONResponseHandler(pw.wsHandlerJSON(&pw.wsHandlersLow)),
		ws_protocol.WithProtobufResponseHandler(pw.wsHandlerProtobuf(&pw.wsHandlersLow)),
		ws_protocol.WithLogger(pw.logger.With().Str("ws", "low").Logger()),
		ws_protocol.WithInsecureConnection(true),
	)
	if err = pw.wsLow.Connect(); err != nil {
		return err
	}
	pw.wsHigh = ws_protocol.NewClient(
		ws_protocol.WithSourceURL(wsServer),
		ws_protocol.WithWireProtocol(pw.wsProto),
		ws_protocol.WithJSONResponseHandler(pw.wsHandlerJSON(&pw.wsHandlersHigh)),
		ws_protocol.WithProtobufResponseHandler(pw.wsHandlerProtobuf(&pw.wsHandlersHigh)),
		ws_protocol.WithLogger(pw.logger.With().Str("ws", "high").Logger()),
		ws_protocol.WithInsecureConnection(true),
	)
	return pw.wsHigh.Connect()
}

func (pw *PlaneWatchTapper) Disconnect() {
	// request things stop sending
	for _, tapName := range pw.taps {
		pw.RemoveIncomingTap(tapName)
	}

	// drains the incoming queues before closing all the things
	pw.natsSvr.Close()
	if err := pw.wsLow.Disconnect(); err != nil {
		pw.logger.Error().Err(err).Msg("Did not disconnect from websocket, low")
	}
	if err := pw.wsHigh.Disconnect(); err != nil {
		pw.logger.Error().Err(err).Msg("Did not disconnect from websocket, high")
	}

	// simple sequential disconnect
	for _, exitChan := range pw.exitChannels {
		exitChan <- true
	}
}

func (pw *PlaneWatchTapper) IncomingDataTap(icao uint32, feederKey string, writer IngestTapHandler) error {
	tapSubject := "ingest-natsTap-" + randstr.RandString(20)
	ch, err := pw.natsSvr.Subscribe(tapSubject)
	if err != nil {
		return fmt.Errorf("failed to open Nats tap: %w", err)
	}
	pw.exitChannels[tapSubject] = make(chan bool)

	go func(ch chan *nats.Msg, exitChan chan bool) {
		for {
			select {
			case msg := <-ch:
				writer(msg.Header.Get("type"), msg.Header.Get("tag"), msg.Data)
			case <-exitChan:
				return
			}
		}
	}(ch, pw.exitChannels[tapSubject])

	icaoStr := ""
	if icao > 0 {
		icaoStr = fmt.Sprintf("%X", icao)
	}

	headers := tapHeaders{
		"action":  tapActionAdd,
		"api-key": feederKey, // the ingest feeder we wish to target
		"icao":    icaoStr,   // the aircraft we wish to look at
		"subject": tapSubject,
	}

	response, errRq := pw.natsSvr.Request(middleware.NatsAPIv1PwIngestTap, []byte{}, headers, time.Second)
	if nil != errRq {
		pw.logger.Error().Err(err).Msg("Failed to request natsTap, are there any pw_ingest's?")
		return errRq
	}
	pw.logger.Debug().Str("response", string(response)).Msg("request response")
	pw.taps = append(pw.taps, headers)

	return nil
}

func (pw *PlaneWatchTapper) RemoveIncomingTap(tap tapHeaders) {
	tap["action"] = tapActionRemove
	response, errRq := pw.natsSvr.Request(middleware.NatsAPIv1PwIngestTap, []byte{}, tap, time.Second)
	if nil != errRq {
		pw.logger.Error().Err(errRq).Msg("Failed to stop natsTap")
		return
	}
	pw.logger.Debug().Str("response", string(response)).Msg("request response")
}

func (pw *PlaneWatchTapper) natsTap(
	icao uint32,
	feederKey string,
	subject string,
	decoder wireDecoderFunc,
	callback func(*export.PlaneAndLocationInfoMsg),
) error {
	listenCh, err := pw.natsSvr.Subscribe(subject)
	if nil != err {
		return err
	}
	tapSubject := subject + "-" + randstr.RandString(20)
	pw.logger.Debug().Str("subject", subject).Str("tap-subject", tapSubject).Msg("Starting Tap")

	pw.exitChannels[tapSubject] = make(chan bool)

	go func(ch chan *nats.Msg, exitChan chan bool) {
		for {
			select {
			case msg := <-ch:

				pli := decoder(msg.Data)
				if icao != 0 && icao != pli.Icao {
					continue
				}
				if feederKey != "" && feederKey != pli.SourceTag {
					continue
				}
				callback(pli)
			case <-exitChan:
				return
			}
		}
	}(listenCh, pw.exitChannels[tapSubject])

	return nil
}

func (pw *PlaneWatchTapper) AfterIngestTap(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	return pw.natsTap(icao, feederKey, pw.queueLocations, pw.pwIngestDecoder, callback)
}

func (pw *PlaneWatchTapper) AfterEnrichmentTap(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	return pw.natsTap(icao, feederKey, pw.queueLocationsEnriched, pw.pwEnrichmentDecoder, callback)
}

func (pw *PlaneWatchTapper) AfterRouterLowTap(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	return pw.natsTap(icao, feederKey, pw.queueLocationsEnrichedLow, pw.pwRouterDecoder, callback)
}

func (pw *PlaneWatchTapper) AfterRouterHighTap(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	return pw.natsTap(icao, feederKey, pw.queueLocationsEnrichedHigh, pw.pwRouterDecoder, callback)
}

func (pw *PlaneWatchTapper) WebSocketTapLow(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	pw.addWsFilterFunc(icao, feederKey, "low", callback)
	return pw.wsLow.SubscribeAllLow()
}
func (pw *PlaneWatchTapper) WebSocketTapHigh(icao uint32, feederKey string, callback func(*export.PlaneAndLocationInfoMsg)) error {
	pw.addWsFilterFunc(icao, feederKey, "high", callback)
	return pw.wsHigh.SubscribeAllHigh()
}

func (pw *PlaneWatchTapper) addWsFilterFunc(icao uint32, feederKey, speed string, filter wsHandler) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	if speed == "low" {
		pw.wsHandlersLow = append(pw.wsHandlersLow, &wsFilterFunc{
			icao:      icao,
			feederTag: feederKey,
			speed:     speed,
			filter:    filter,
		})
	} else {
		pw.wsHandlersHigh = append(pw.wsHandlersHigh, &wsFilterFunc{
			icao:      icao,
			feederTag: feederKey,
			filter:    filter,
		})
	}
}

func (pw *PlaneWatchTapper) decoderJSON(buf []byte) *export.PlaneAndLocationInfoMsg {
	planeLocation := export.NewEmptyPlaneLocationJSON()
	err := json.Unmarshal(buf, &planeLocation)
	if nil != err {
		pw.logger.Error().Err(err).Msg("Failed to decode PlaneLocationJSON")
		return nil
	}
	pli := export.NewPlaneAndLocationInfoMsg()
	if err = planeLocation.ToProtobuf(pli); err != nil {
		pw.logger.Error().Err(err).Msg("Failed to convert to protobuf struct")
		return nil
	}
	return pli
}

func (pw *PlaneWatchTapper) decoderProtobuf(buf []byte) *export.PlaneAndLocationInfoMsg {
	// todo: use our own sync.pool for export.PlaneAndLocationInfoMsg items
	pli, err := export.FromProtobufBytes(buf)

	if err != nil {
		pw.logger.Error().Err(err).Msg("Failed to decode protobuf")
		return nil
	}
	return pli
}

// wsHandlerJSON runs through all the handlers we have and calls maybeSend for each plane location.
// this method is for JSON WS API Endpoints (not Protobuf)
func (pw *PlaneWatchTapper) wsHandlerJSON(handlers *[]*wsFilterFunc) func(r *ws_protocol.WsResponse) {
	return func(r *ws_protocol.WsResponse) {
		var err error
		pw.mu.Lock()
		defer pw.mu.Unlock()
		for _, ff := range *handlers {
			pli := export.NewPlaneAndLocationInfoMsg()
			if err = r.Location.ToProtobuf(pli); err == nil {
				pw.maybeSend(ff, pli)
			}
			if r.Locations != nil {
				for _, loc := range r.Locations {
					if err = loc.ToProtobuf(pli); err == nil {
						pw.maybeSend(ff, pli)
					}
				}
			}
		}
	}
}

// wsHandlerProtobuf runs through all the handlers we have and calls maybeSend for each plane location.
// this method is for Protobuf WS API Endpoints (not JSON)
func (pw *PlaneWatchTapper) wsHandlerProtobuf(handlers *[]*wsFilterFunc) func(r *ws_protocol.WebSocketResponse) {
	return func(r *ws_protocol.WebSocketResponse) {
		pw.mu.Lock()
		defer pw.mu.Unlock()
		for _, ff := range *handlers {
			if r.Location != nil {
				pw.maybeSend(ff, r.Location.AsPlaneAndLocationInfoMsg())

			}
			if r.Locations != nil && r.Locations.Location != nil {
				for _, loc := range r.Locations.Location {
					pw.maybeSend(ff, loc.AsPlaneAndLocationInfoMsg())
				}
			}
		}
	}
}

// maybeSend takes into account the filter and calls the user provided handler if it matches
func (pw *PlaneWatchTapper) maybeSend(ff *wsFilterFunc, loc *export.PlaneAndLocationInfoMsg) {
	if nil == loc {
		return
	}
	if ff.icao != 0 && ff.icao != loc.Icao {
		return
	}
	if ff.feederTag != "" {
		if _, ok := loc.SourceTags[ff.feederTag]; !ok {
			return
		}
	}

	ff.filter(loc)
}
