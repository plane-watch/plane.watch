package atc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type (
	Server struct {
		Url      url.URL
		Username string
		Password string
		//log      zerolog.Logger
	}

	Feeders struct { // schema for /api/v1/feeders.json atc endpoint
		Feeders []Feeder
	}

	Feeder struct { // part of schema for /api/v1/feeders.json atc endpoint
		Altitude      float64 `json:",string"`
		ApiKey        uuid.UUID
		FeedDirection string
		FeedProtocol  string
		ID            int
		Label         string
		Latitude      float64 `json:",string"`
		Longitude     float64 `json:",string"`
		MlatEnabled   bool
		Mux           string
		User          string
	}

	FeederB struct { // schema for /api/v1/feeders/{uuid}.json atc endpoint
		Feeder struct {
			ApiKey      uuid.UUID `json:"api_key"`
			Label       string    `json:"label"`
			MlatEnabled bool      `json:"mlat_enabled"`
			Latitude    float64   `json:"latitude,string"`
			Longitude   float64   `json:"longitude,string"`
			Protocol    string    `json:"protocol"`
			Elevation   float64   `json:"elevation,string"`
			Mux         string    `json:"mux"`
		} `json:"feeder"`
	}

	atcCredentials struct { // schema for atc credentials
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	atcUser struct { // schema for atc authentication
		User atcCredentials `json:"user"`
	}
)

func authenticate(server *Server) (authToken string, err error) {

	atcUrl := server.Url.JoinPath("api/user/sign_in")

	creds := atcUser{
		User: atcCredentials{
			Email:    server.Username,
			Password: server.Password,
		},
	}

	credsJson, err := json.Marshal(creds)
	if err != nil {
		return authToken, err
	}

	req, err := http.NewRequest("POST", atcUrl.String(), bytes.NewBuffer(credsJson))
	if err != nil {
		return authToken, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return authToken, err
	}
	defer response.Body.Close()

	// fmt.Println("response Status:", response.Status)
	// fmt.Println("response Headers:", response.Header)
	// body, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("response Body:", string(body))

	if response.StatusCode == 200 {

		authToken = response.Header.Get("Authorization")
		if authToken == "" {
			return authToken, errors.New("No authorization token returned")
		}

	} else {
		errStr := fmt.Sprintf("ATC API response status: %s", response.Status)
		return authToken, errors.New(errStr)
	}

	return authToken, nil

}

func GetFeederLatLon(server *Server, feederApiKey uuid.UUID) (refLat float64, refLon float64, err error) {

	authToken, err := authenticate(server)
	if err != nil {
		return refLat, refLon, err
	}

	atcUrl := server.Url.JoinPath(fmt.Sprintf("/api/v1/feeders/%s.json", feederApiKey.String()))

	req, err := http.NewRequest("GET", atcUrl.String(), nil)
	if err != nil {
		return refLat, refLon, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return refLat, refLon, err
	}
	defer response.Body.Close()

	// fmt.Println("response Status:", response.Status)
	// fmt.Println("response Headers:", response.Header)
	// body, _ := ioutil.ReadAll(response.Body)
	// fmt.Println("response Body:", string(body))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return refLat, refLon, err
	}

	var feeder FeederB

	if response.StatusCode == 200 {

		err := json.Unmarshal(body, &feeder)
		if err != nil {
			fmt.Println(err)
			return refLat, refLon, err
		}

		refLat = feeder.Feeder.Latitude
		refLon = feeder.Feeder.Longitude

	} else {
		errStr := fmt.Sprintf("ATC API response status: %s", response.Status)
		return refLat, refLon, errors.New(errStr)
	}

	return refLat, refLon, err

}

func GetFeeders(server *Server) (feeders Feeders, err error) {

	authToken, err := authenticate(server)
	if err != nil {
		return feeders, err
	}

	atcUrl := server.Url.JoinPath("/api/v1/feeders.json")

	req, err := http.NewRequest("GET", atcUrl.String(), nil)
	if err != nil {
		return feeders, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return feeders, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return feeders, err
	}

	// fmt.Println("response Body:", string(body))

	if response.StatusCode == 200 {

		err := json.Unmarshal(body, &feeders)
		if err != nil {
			return feeders, err
		}
	} else {
		errStr := fmt.Sprintf("ATC API response status: %s", response.Status)
		return feeders, errors.New(errStr)
	}

	return feeders, nil
}
