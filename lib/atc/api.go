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

	Feeders struct {
		Feeders []Feeder
	}

	Feeder struct {
		Altitude      float32 `json:",string"`
		ApiKey        uuid.UUID
		FeedDirection string
		FeedProtocol  string
		ID            int
		Label         string
		Latitude      float32 `json:",string"`
		Longitude     float32 `json:",string"`
		MlatEnabled   bool
		Mux           string
		User          string
	}

	atcCredentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	atcUser struct {
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
	}

	return feeders, nil
}
