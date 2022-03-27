package main

// handles the list of alerting locations

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"path"
	"plane.watch/lib/mapping"
	"plane.watch/lib/tile_grid"
	"strings"
	"sync"
)

const alertLocationsFile = "alert-locations.json"

type (
	alertConfig struct {
		HeightLowFt, HeightHighFt int
		AlertRadiusMtr            int
		Enabled                   bool
	}
	alertConfigs map[string]*alertConfig

	location struct {
		DiscordUserId   string
		DiscordUserName string
		LocationName    string
		Address         string
		Lat             float64
		Lon             float64
		AlertConfig     alertConfigs // The radius of the circle to alert in
		TileGrid        string
	}

	locationMatchFunc func(l *location)
)

var (
	alertLocationsRWLock sync.RWMutex
	alertLocations       []location
	isLoaded             bool
	standardAlerts       = map[string]*alertConfig{
		"very-low": { // LOW flying aircraft, probably looking for something
			HeightLowFt:    -2_000,
			HeightHighFt:   2_000,
			AlertRadiusMtr: 500,
			Enabled:        true,
		},
		"low": { // low flying, probably in transit
			HeightLowFt:    2_000,
			HeightHighFt:   5_000,
			AlertRadiusMtr: 1000,
			Enabled:        true,
		},
		"medium": { // small commercial? ~ 5000ft
			HeightLowFt:    5_000,
			HeightHighFt:   10_000,
			AlertRadiusMtr: 1500,
			Enabled:        true,
		},
		"high": { // aircraft up to 15,000ft - probably coming into land
			HeightLowFt:    10_000,
			HeightHighFt:   25_000,
			AlertRadiusMtr: 3000,
			Enabled:        true,
		},
		"very-high": { // aircraft above 15,000ft - probably commercial in transit, not wanting an alert by default
			HeightLowFt:    25_000,
			HeightHighFt:   328084, // 100k
			AlertRadiusMtr: 20000,
			Enabled:        false,
		},
	}
)

func getPath() string {
	binaryPath, _ := os.Executable()
	if "" != binaryPath && !strings.Contains(binaryPath, "/go-build/") {
		return path.Dir(binaryPath)
	}
	dir, _ := os.Getwd()
	return dir
}

// getExisting returns the id in the array of the existing record, -1 if it does not exist
func getExisting(discordUserId, locationName string) int {
	alertLocationsRWLock.RLock()
	defer alertLocationsRWLock.RUnlock()
	for idx, loc := range alertLocations {
		if loc.DiscordUserId != discordUserId {
			continue
		}
		if loc.LocationName == locationName {
			return idx
		}
	}
	return -1
}

func getLocationsForUser(discordUserId string) []location {
	alertLocationsRWLock.RLock()
	defer alertLocationsRWLock.RUnlock()

	var ret []location
	for _, loc := range alertLocations {
		if loc.DiscordUserId == discordUserId {
			ret = append(ret, loc)
		}
	}
	return ret
}

func addAlertLocation(discordUserId, discordUserName, locationName string, lat, lon float64) error {
	// make sure we do not already have this name
	if -1 != getExisting(discordUserId, locationName) {
		return errors.New("you have an existing location with this name")
	}

	alertLocationsRWLock.Lock()
	loc := location{
		DiscordUserId:   discordUserId,
		DiscordUserName: discordUserName,
		LocationName:    locationName,
		Lat:             lat,
		Lon:             lon,
		AlertConfig:     standardAlerts,
		TileGrid:        tile_grid.LookupTile(lat, lon),
	}
	alertLocations = append(alertLocations, loc)
	alertLocationsRWLock.Unlock()
	return saveLocationsList()
}

func removeAlertLocation(discordUserId, locationName string) error {
	idx := getExisting(discordUserId, locationName)
	if -1 == idx {
		return errors.New("unknown location")
	}
	alertLocationsRWLock.Lock()
	if len(alertLocations) == 1 && idx == 0 {
		alertLocations = []location{}
	} else if idx == len(alertLocations)-1 {
		// last element, just shorten
		alertLocations = alertLocations[:idx-1]
	} else {
		alertLocations = append(alertLocations[:idx], alertLocations[idx+1:]...)
	}
	alertLocationsRWLock.Unlock()
	return saveLocationsList()
}

// setLocationAddress allows us to set the address of a geocoded location
func setLocationAddress(discordUserId, locationName, address string) error {
	idx := getExisting(discordUserId, locationName)
	if -1 == idx {
		return errors.New("that location name does not exist")
	}
	alertLocationsRWLock.Lock()
	alertLocations[idx].Address = address
	alertLocationsRWLock.Unlock()
	return saveLocationsList()
}

// allows updating the radius in which we alert for this location
func setLocationAlertConfigEnabled(discordUserId, locationName, which string, enabled bool) error {
	idx := getExisting(discordUserId, locationName)
	if -1 == idx {
		return errors.New("that location name does not exist")
	}
	alertLocationsRWLock.Lock()
	alertLocations[idx].AlertConfig[which].Enabled = enabled
	alertLocationsRWLock.Unlock()
	return saveLocationsList()
}

func loadLocationsList() {
	alertLocationsRWLock.Lock()
	defer alertLocationsRWLock.Unlock()
	if isLoaded {
		return
	}
	saveLoc := getPath() + "/" + alertLocationsFile
	b, err := ioutil.ReadFile(saveLoc)
	if nil != err {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("No save file. %s does not exist. proceeding with empty list", saveLoc)
			return
		}
		log.Error().Err(err).Msgf("Failed to read %s. %s", saveLoc, err)
		return
	}
	err = json.Unmarshal(b, &alertLocations)
	if nil != err {
		log.Error().Err(err).Msgf("Failed to parse %s JSON perfectly. %s", saveLoc, err)
		panic(err)
	}
	isLoaded = true
}

func saveLocationsList() error {
	alertLocationsRWLock.RLock()
	defer alertLocationsRWLock.RUnlock()
	saveLoc := getPath() + "/" + alertLocationsFile

	b, err := json.MarshalIndent(alertLocations, "", "  ")
	if nil != err {
		if len(b) <= 3 {
			return fmt.Errorf("we don't have a good marshalling, not saving. %s", err)
		} else {
			log.Printf("Failed to marshal JSON, attempting to save what we have. %s", err)
		}
	}

	err = ioutil.WriteFile(saveLoc, b, 0644)
	if nil != err {
		return fmt.Errorf("failed to save locations to %s. %s", saveLoc, err)
	}
	return nil
}

func geoCodeAddress(addr string) (float64, float64, error) {
	log.Info().Msgf("Geocoding user address: %s", addr)
	return mapping.FindCoordinates(addr)
}

func forLocation(tileName string, matchFunc locationMatchFunc) {
	alertLocationsRWLock.RLock()
	defer alertLocationsRWLock.RUnlock()
	for i := range alertLocations {
		if alertLocations[i].TileGrid == tileName {
			matchFunc(&alertLocations[i])
		}
	}
}

func (ac *alertConfigs) configForHeight(altitude int) *alertConfig {
	for _, config := range *ac {
		if altitude >= config.HeightLowFt && altitude < config.HeightHighFt {
			return config
		}
	}
	return nil
}
