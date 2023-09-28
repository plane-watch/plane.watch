package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
	"plane.watch/lib/export"
	"plane.watch/lib/middleware"
	"plane.watch/lib/nats_io"
	"plane.watch/lib/randstr"
	"plane.watch/lib/sink"
	"time"
)

const (
	tapActionAdd    = "add"
	tapActionRemove = "remove"
)

type (
	tapHeaders map[string]string

	PlaneWatchTapper struct {
		natsSvr *nats_io.Server
		ws      *websocket.Conn

		exitChannels map[string]chan bool
		taps         []tapHeaders

		logger zerolog.Logger

		queueLocations             string
		queueLocationsEnriched     string
		queueLocationsEnrichedLow  string
		queueLocationsEnrichedHigh string
	}

	IngestTapHandler func(frameType, tag string, data []byte)

	Option func(*PlaneWatchTapper)
)

func NewPlaneWatchTapper(opts ...Option) *PlaneWatchTapper {
	pw := &PlaneWatchTapper{
		natsSvr:                    nil,
		ws:                         nil,
		exitChannels:               make(map[string]chan bool),
		taps:                       make([]tapHeaders, 0),
		logger:                     zerolog.Logger{},
		queueLocations:             sink.QueueLocationUpdates,
		queueLocationsEnriched:     "location-updates-enriched",
		queueLocationsEnrichedLow:  "location-updates-enriched-reduced",
		queueLocationsEnrichedHigh: "location-updates-enriched-merged",
	}

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

func (pw *PlaneWatchTapper) Connect(natsServer, wsServer string) error {
	var err error
	pw.natsSvr, err = nats_io.NewServer(
		nats_io.WithConnections(true, true),
		nats_io.WithServer(natsServer, "ingest-nats-tap"),
	)
	if nil != err {
		return err
	}

	if wsServer != "" {
		pw.ws, _, err = websocket.Dial(context.Background(), wsServer, nil)

		if nil != err {
			return err
		}
	}

	return nil
}

func (pw *PlaneWatchTapper) Disconnect() {
	// request things stop sending
	for _, tapName := range pw.taps {
		pw.RemoveIncomingTap(tapName)
	}

	// drains the incoming queues before closing all the things
	pw.natsSvr.Close()

	// simple sequential disconnect
	for _, exitChan := range pw.exitChannels {
		exitChan <- true
	}
}

func (pw *PlaneWatchTapper) IncomingDataTap(icao, feederKey string, writer IngestTapHandler) error {
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

	headers := tapHeaders{
		"action":  tapActionAdd,
		"api-key": feederKey, // the ingest feeder we wish to target
		"icao":    icao,      // the aircraft we wish to look at
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

func (pw *PlaneWatchTapper) natsTap(subject string, callback func(*export.PlaneLocation)) error {
	listenCh, err := pw.natsSvr.Subscribe(subject)
	if nil != err {
		return err
	}
	tapSubject := subject + "-" + randstr.RandString(20)

	pw.exitChannels[tapSubject] = make(chan bool)

	go func(ch chan *nats.Msg, exitChan chan bool) {
		for {
			select {
			case msg := <-ch:
				var planeLocation export.PlaneLocation
				err = json.Unmarshal(msg.Data, &planeLocation)
				if nil != err {
					pw.logger.Error().Err(err)
					continue
				}
				callback(&planeLocation)
			case <-exitChan:
				return
			}
		}
	}(listenCh, pw.exitChannels[tapSubject])

	return nil
}

func (pw *PlaneWatchTapper) AfterIngestTap(callback func(*export.PlaneLocation)) error {
	return pw.natsTap(pw.queueLocations, callback)
}

func (pw *PlaneWatchTapper) AfterEnrichmentTap(callback func(*export.PlaneLocation)) error {
	return pw.natsTap(pw.queueLocationsEnriched, callback)
}

func (pw *PlaneWatchTapper) AfterRouterLowTap(callback func(*export.PlaneLocation)) error {
	return pw.natsTap(pw.queueLocationsEnrichedLow, callback)
}

func (pw *PlaneWatchTapper) AfterRouterHighTap(callback func(*export.PlaneLocation)) error {
	return pw.natsTap(pw.queueLocationsEnrichedHigh, callback)
}

func (pw *PlaneWatchTapper) WebSocketTap(callback func(*export.PlaneLocation)) error {
	return nil
}
