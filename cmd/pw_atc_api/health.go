package main

import "github.com/rs/zerolog/log"

type health struct {
}

func (h *health) HealthCheckName() string {
	return "pw_atc_api"
}

func (h *health) HealthCheck() bool {
	if err := db.Ping(); nil != err {
		log.Error().Err(err).Send()
	}

	return true
}
