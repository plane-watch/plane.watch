package main

import (
	"github.com/rs/zerolog"
	"math/rand"
	"plane.watch/lib/export"
	"plane.watch/lib/ws_client"
	"time"
)

type (
	pwWsClient struct {
		wsClient *ws_client.Client
		wsLog    zerolog.Logger

		handleUpdate func(planeLocation *export.PlaneLocationJSON)

		exiting bool
	}
)

// contains interactions with plane.watch to get the location info
func (wsc *pwWsClient) handleWebsocketClient(host string, insecure bool) {
	wsc.wsClient = ws_client.NewClient(host)
	if insecure {
		wsc.wsClient.Secure(insecure)
	}
	backoff := 1

	if nil == wsc.handleUpdate {
		panic("You need to specify the handleUpdate method")
	}

	for {
		if wsc.exiting {
			return
		}
		wsc.wsLog.Info().Str("host", host).Bool("secure", !insecure).Msg("Connecting...")
		if err := wsc.wsClient.Connect(); nil != err {
			wsc.wsLog.Error().Err(err).Int("backoff seconds", backoff).Msg("Cannot connect to websocket")
			time.Sleep(time.Duration(backoff) * time.Second)

			backoff = backoff + backoff + (rand.Intn(10) - 2)
			if backoff > 30 {
				backoff = 30
			}
			continue
		} else {
			backoff = 1
		}

		if err := wsc.wsClient.Subscribe("all_high"); nil != err {
			wsc.wsLog.Error().Err(err).Msg("Unable to subscribe to all_high feed")
			if err := wsc.wsClient.Disconnect(); nil != err {
				wsc.wsLog.Error().Err(err).Msg("Did not disconnect gracefully after subscribe fail")
			}
			continue
		}

		for update := range wsc.wsClient.LocationUpdates() {
			//wsc.wsLog.Debug().Msgf("update: %+v", update)
			wsc.handleUpdate(update)
		}
	}
}

func (wsc *pwWsClient) stop() error {
	// todo: better handle closing
	wsc.exiting = true
	return wsc.wsClient.Disconnect()
}
