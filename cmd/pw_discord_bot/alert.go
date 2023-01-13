package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"plane.watch/lib/export"
	"sync"
	"time"
)

// handles the determining if an alert needs to be sent to a user
const (
	earthRadiusMtr = 6371000
)

type (
	pwAlertBot struct {
		locationUpdates  chan *export.PlaneLocation
		numUpdateWorkers int

		// keeps track of when we alerted a user
		alertUpdates sync.Map

		wg  sync.WaitGroup
		log zerolog.Logger

		sendAlert func(pa *proximityAlert)
	}

	proximityAlert struct {
		time        time.Time
		alert       *location
		update      *export.PlaneLocation
		distanceMtr int
	}
)

func (a *pwAlertBot) runAlerts() {
	if nil == a.sendAlert {
		panic("You need to specify the pwAlertBot.sendAlert method")
	}
	a.wg.Add(a.numUpdateWorkers)
	a.log.Info().Int("Num Workers", a.numUpdateWorkers).Msg("Starting Alert Workers")
	for i := 0; i < a.numUpdateWorkers; i++ {
		go func(id int) {
			for update := range a.locationUpdates {
				a.handleUpdate(update)
			}
			a.wg.Done()
			a.log.Debug().Int("Worker #", id).Msg("stopped")
		}(i)
	}
}

func (a *pwAlertBot) stop() error {
	a.log.Info().Int("Num Workers", a.numUpdateWorkers).Msg("Stopping Alert Workers")
	close(a.locationUpdates)
	a.wg.Wait()
	return nil
}

func (a *pwAlertBot) handleUpdate(update *export.PlaneLocation) {
	if nil == update {
		return
	}
	// ignore updates of we do not have enough data on them
	if 0 == update.Altitude && !update.OnGround {
		// probably an update that is incomplete, we can catch the next one
		return
	}

	forLocation(update.TileLocation, func(alert *location) {
		distance := getDistanceBetween(update.Lat, update.Lon, alert.Lat, alert.Lon)
		ac := alert.AlertConfig.configForHeight(update.Altitude)
		if nil == ac {
			log.Error().Int32("altitude", update.Altitude).Msg("Failed to get alert config")
			return
		}
		if !ac.Enabled {
			return
		}
		if log.Trace().Enabled() {
			log.Trace().
				Floats64("alert-location", []float64{alert.Lat, alert.Lon}).
				Floats64("plane-location", []float64{update.Lat, update.Lon}).
				Int("Distance (m)", distance).
				Int("Alert Radius", ac.AlertRadiusMtr).
				Bool("In Air Space", distance <= ac.AlertRadiusMtr).
				Msg("Distance Calc")
		}

		if distance <= ac.AlertRadiusMtr {
			// do alert
			a.alertUser(&proximityAlert{
				time:        time.Now(),
				alert:       alert,
				update:      update,
				distanceMtr: distance,
			})
		}
	})
}

func (a *pwAlertBot) alertUser(pa *proximityAlert) {
	// if we have seen this plane in the last few minutes, ignore it (no need to alert more than once for a circling chopper)
	if nil == pa {
		return
	}
	key := pa.Key()

	if existingAlert, ok := a.alertUpdates.Load(key); ok {
		lastSeen := existingAlert.(*proximityAlert)

		if !lastSeen.time.Before(pa.time.Add(-5 * time.Minute)) {
			// we have seen this plane within the last 5 minutes, no need to update
			return
		}
	}

	a.alertUpdates.Store(key, pa)
	a.log.Debug().Msgf("Sending Alert")

	// we need to send a discord message
	a.sendAlert(pa)
}

func (pa *proximityAlert) Key() string {
	return fmt.Sprintf("%s%s%06X", pa.alert.DiscordUserId, pa.alert.LocationName, pa.update.Icao)
}

// getDistanceBetween takes 2 Lat/Lon pairs and calculates the distance between them, in metres
// we use the Great Circle calculation method
func getDistanceBetween(lat1, lon1, lat2, lon2 float64) int {
	// http://janmatuschek.de/LatitudeLongitudeBoundingCoordinates
	//  1 Distance Between Two Given Points
	//  dist = arccos(sin(lat1) · sin(lat2) + cos(lat1) · cos(lat2) · cos(lon1 - lon2)) · R

	radLat1 := lat1 * math.Pi / 180
	radLon1 := lon1 * math.Pi / 180
	radLat2 := lat2 * math.Pi / 180
	radLon2 := lon2 * math.Pi / 180

	distanceBetween := math.Acos(
		(math.Sin(radLat1)*math.Sin(radLat2))+
			(math.Cos(radLat1)*math.Cos(radLat2)*math.Cos(radLon1-radLon2)),
	) * earthRadiusMtr

	return int(math.Round(distanceBetween))
}
