package main

import (
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"plane.watch/lib/atc"
	"plane.watch/lib/logging"
)

type atcFeeders struct {
	mu      sync.Mutex
	feeders []uuid.UUID
}

var (
	validFeeders atcFeeders

	prometheusInputBeastFrames = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ingest_input_beast_total",
		Help: "The total number of beast frames processed.",
	})
	prometheusInputAvrFrames = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ingest_input_avr_total",
		Help: "The total number of AVR frames processed.",
	})
	prometheusInputSbs1Frames = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ingest_input_sbs1_total",
		Help: "The total number of SBS1 frames processed.",
	})
)

func isValidApiKey(clientApiKey uuid.UUID) bool {
	// return true of api key clientApiKey is a valid feeder in atc
	validFeeders.mu.Lock()
	defer validFeeders.mu.Unlock()
	for _, v := range validFeeders.feeders {
		if v == clientApiKey {
			return true
		}
	}
	return false
}

func updateFeederDB(ctx *cli.Context) {
	// updates validFeeders with data from atc
	// TODO: could we use NATS K/V store instead?

	firstRun := true

	for {

		if !firstRun {
			time.Sleep(30 * time.Second)
		} else {
			firstRun = false
		}

		log.Debug().Msg("updating api key cache from atc")

		atcUrl, err := url.Parse(ctx.String("atcurl"))
		if err != nil {
			log.Error().Msg("--atcurl is invalid")
			continue
		}
		s := atc.Server{
			Url:      *atcUrl,
			Username: ctx.String("atcuser"),
			Password: ctx.String("atcpass"),
		}
		f, err := atc.GetFeeders(&s)
		var newValidFeeders []uuid.UUID
		for _, v := range f.Feeders {
			newValidFeeders = append(newValidFeeders, v.ApiKey)
		}

		validFeeders.mu.Lock()
		validFeeders.feeders = newValidFeeders
		validFeeders.mu.Unlock()

	}
}

func clientConnection(ctx *cli.Context, conn net.Conn, tlsConfig *tls.Config) {
	// handles incoming connections

	// TODO: need a way to kill a client connection if the UUID is no longer valid (ie: feeder banned)

	cLog := log.With().Logger()

	var (
		sendRecvBufferSize  = 1024
		clientAuthenticated = false
		clientApiKey        uuid.UUID
		err                 error
	)

	defer conn.Close()

	// update log context with client IP
	remoteIP := net.ParseIP(strings.Split(conn.RemoteAddr().String(), ":")[0])
	cLog = cLog.With().IPAddr("client", remoteIP).Logger()

	cLog.Debug().Msgf("connection established")
	defer cLog.Debug().Msgf("connection closed")

	buf := make([]byte, sendRecvBufferSize)
	for {

		// read data
		_, err = conn.Read(buf)
		if err != nil {
			if err.Error() == "tls: first record does not look like a TLS handshake" {
				cLog.Warn().Msg(err.Error())
			} else if err.Error() == "EOF" {
				if clientAuthenticated {
					cLog.Info().Msg("client disconnected")
				}
			} else {
				cLog.Err(err).Msg("conn.Read")
			}
			break
		}

		// When the first data is sent, the TLS handshake should take place.
		// Accordingly, we need to track the state...
		if !clientAuthenticated {

			// check TLS handshake
			tlscon := conn.(*tls.Conn)
			if tlscon.ConnectionState().HandshakeComplete {

				// check valid uuid was returned as ServerName (sni)
				clientApiKey, err = uuid.Parse(tlscon.ConnectionState().ServerName)
				if err != nil {
					cLog.Warn().Msg("client sent invalid uuid")
					break
				}

				// check valid api key
				if isValidApiKey(clientApiKey) {
					// update log context with client uuid
					cLog = cLog.With().Str("apikey", clientApiKey.String()).Logger()
					cLog.Info().Msg("client connected")
					// if API is valid, then set clientAuthenticated to TRUE
					clientAuthenticated = true
				} else {
					// if API is not valid, then kill the connection
					cLog.Warn().Msg("client sent invalid api key")
					break
				}

			} else {
				// if TLS handshake is not complete, then kill the connection
				cLog.Warn().Msg("data received before tls handshake")
				break
			}
		}

		// If the client has been authenticated, then we can do stuff with the data
		if clientAuthenticated {
			// TODO: do stuff with the data - talk to boxie
			// TODO: need a nice way to update atc that the feeder is online since the time it connected...
			// TODO: maybe have a timer so that it only updates every 5 minutes + some random seconds (to prevent overload of ATC)
			// TODO: do we also need to mark offline on disconnect?
			// cLog.Debug().Msgf("data received: %s", fmt.Sprint(buf[:n]))

			// get feeder lat/long
			atcUrl, err := url.Parse(ctx.String("atcurl"))
			if err != nil {
				log.Error().Msg("--atcurl is invalid")
				continue
			}
			s := atc.Server{
				Url:      *atcUrl,
				Username: ctx.String("atcuser"),
				Password: ctx.String("atcpass"),
			}
			// refLat, refLon, err := atc.GetFeederLatLon(&s, clientApiKey)
			_, _, err = atc.GetFeederLatLon(&s, clientApiKey)
			if err != nil {
				log.Err(err).Msg("atc.GetFeederLatLon")
				continue
			}

			// at this point we should have everything we need to set up a producer for pw_ingest...

			// set up producer
			// producerOpts := make([]producer.Option, 3)
			// producerOpts[0] = producer.WithSourceTag(clientApiKey.String())
			// producerOpts[1] = producer.WithType(producer.Beast)
			// producerOpts[2] = producer.WithPrometheusCounters(prometheusInputAvrFrames, prometheusInputBeastFrames, prometheusInputSbs1Frames)
			// producerOpts = append(producerOpts, producer.WithReferenceLatLon(refLat, refLon))

		}
	}
}

func main() {

	app := &cli.App{
		Name:  "Plane Watch BEAST Feeder Endpoint",
		Usage: "Server for multiple stunnel-based endpoints",
		Description: `This program acts as a server for multiple stunnel-based endpoints, ` +
			`authenticates the feeder based on UUID check against atc.plane.watch, ` +
			`receives BEAST data and ingests via pw_ingest.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen",
				Usage: "Address and TCP port server will listen on",
				Value: "0.0.0.0:12345",
			},
			&cli.StringFlag{
				Name:     "cert",
				Usage:    "Server certificate PEM file name (x509)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "key",
				Usage:    "Server certificate private key PEM file name (x509)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "atcurl",
				Usage:    "URL to ATC API",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "atcuser",
				Usage:    "email username for ATC API",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "atcpass",
				Usage:    "password for ATC API",
				Required: true,
			},
		},
		Action: runServer,
	}

	logging.IncludeVerbosityFlags(app)
	logging.ConfigureForCli()

	// Set logging level
	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)
		return nil
	}

	// Final exit
	if err := app.Run(os.Args); nil != err {
		log.Err(err).Msg("Finishing with an error")
		os.Exit(1)
	}

}

func runServer(ctx *cli.Context) error {

	// load server cert & key
	// TODO: reload certificate on sighup - get boxie help!
	cert, err := tls.LoadX509KeyPair(
		ctx.String("cert"),
		ctx.String("key"),
	)
	if err != nil {
		log.Err(err).Msg("tls.LoadX509KeyPair")
	}

	// tls configuration
	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}

	// start TLS server
	log.Info().Msgf("Starting %s on %s", ctx.App.Name, ctx.String("listen"))
	tlsListener, err := tls.Listen("tcp", ctx.String("listen"), &tlsConfig)
	if err != nil {
		log.Err(err).Msg("tls.Listen")
	}
	defer tlsListener.Close()

	// start api key cache updater
	go updateFeederDB(ctx)

	// handle incoming connections
	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			log.Err(err).Msg("tlsListener.Accept")
			continue
		}
		defer conn.Close()
		go clientConnection(ctx, conn, &tlsConfig)
	}
	return nil
}
