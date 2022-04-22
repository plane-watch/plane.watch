package main

import (
	"errors"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"plane.watch/lib/logging"
	"plane.watch/lib/monitoring"
)

var (
	version              = "dev"
	prometheusNumClients = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pw_ws_broker_num_clients",
		Help: "The current number of websocket clients we are currently serving",
	})
)

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Name = "Plane.Watch WebSocket Broker (pw_ws_broker)"
	app.Usage = "Websocket Broker"
	app.Description = "Acts as a go between external display elements our the data pipeline"
	app.Authors = []*cli.Author{
		{
			Name:  "Jason Playne",
			Email: "jason@jasonplayne.com",
		},
	}

	app.Commands = cli.Commands{
		{
			Name:        "daemon",
			Description: "For prod, Logging is JSON formatted",
			Action:      runDaemon,
		},
		{
			Name:        "cli",
			Description: "Runs in your terminal with human readable output",
			Action:      runCli,
		},
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "rabbitmq",
			Aliases: []string{"source"},
			Usage:   "A place to fetch data from. amqp://user:pass@host:port/vhost?ttl=60",
			EnvVars: []string{"RABBITMQ", "SOURCE"},
		},
		&cli.StringFlag{
			Name:    "nats",
			Usage:   "Nats.io URL for fetching and publishing updates. nats://guest:guest@host:4222/",
			EnvVars: []string{"NATS"},
		},
		&cli.StringFlag{
			Name:    "redis",
			Usage:   "Redis URL for fetching updates. redis://guest:guest@redis:6379/",
			EnvVars: []string{"REDIS"},
		},
		&cli.StringFlag{
			Name:    "route-key-low",
			Usage:   "The routing key that has only the significant flight update events",
			Value:   "location-updates-enriched-reduced",
			EnvVars: []string{"ROUTE_KEY_LOW"},
		},
		&cli.StringFlag{
			Name:    "route-key-high",
			Usage:   "The routing key that has all of the flight update events",
			Value:   "location-updates-enriched",
			EnvVars: []string{"ROUTE_KEY_HIGH"},
		},
		&cli.StringFlag{
			Name:    "http-addr",
			Usage:   "What the HTTP server listens on",
			Value:   ":80",
			EnvVars: []string{"HTTP_ADDR"},
		},
		&cli.StringFlag{
			Name:    "tls-cert",
			Usage:   "The path to a PEM encoded TLS Full Chain Certificate (cert+intermediate+ca)",
			Value:   "",
			EnvVars: []string{"TLS_CERT"},
		},
		&cli.StringFlag{
			Name:    "tls-cert-key",
			Usage:   "The path to a PEM encoded TLS Certificate Key",
			Value:   "",
			EnvVars: []string{"TLS_CERT_KEY"},
		},

		&cli.BoolFlag{
			Name:    "serve-test-web",
			Usage:   "Serve up a test website for websocket testing",
			EnvVars: []string{"TEST_WEB"},
		},
		&cli.DurationFlag{
			Name:    "send-tick",
			Usage:   "When > 0, how long to collect messages before sending them in one batch",
			EnvVars: []string{"SEND_TICK"},
			Value:   0,
		},
	}

	logging.IncludeVerbosityFlags(app)
	monitoring.IncludeMonitoringFlags(app, 9603)

	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)

		return nil
	}

	if err := app.Run(os.Args); nil != err {
		log.Error().Err(err).Send()
	}
}

func runDaemon(c *cli.Context) error {
	return run(c)
}

func runCli(c *cli.Context) error {
	logging.ConfigureForCli()
	return run(c)
}

func run(c *cli.Context) error {
	cert := c.String("tls-cert")
	certKey := c.String("tls-cert-key")
	if ("" != cert || "" != certKey) && ("" == cert || "" == certKey) {
		return errors.New("please provide both certificate and key")
	}
	if ":80" == c.String("http-addr") && "" != cert {
		return c.Set("http-addr", ":443")
	}

	monitoring.RunWebServer(c)

	rabbitmq := c.String("rabbitmq")
	nats := c.String("nats")
	redis := c.String("redis")
	lowRoute := c.String("route-key-low")
	highRoute := c.String("route-key-high")

	hasRedis := "" != redis
	hasNats := "" != nats
	hasRabbit := "" != rabbitmq

	isValid := true
	if !hasRabbit && !hasNats && !hasRedis {
		log.Info().Msg("Please provide rabbitmq (or nats, redis) connection details. (--rabbitmq, --nats, --redis)")
		isValid = false
	}
	if "" == lowRoute {
		log.Info().Msg("Please provide the routing key for significant updates. (--route-key-low)")
		isValid = false
	}
	if "" == highRoute {
		log.Info().Msg("Please provide the routing key for all updates. (--route-key-high)")
		isValid = false
	}
	if !isValid {
		return errors.New("invalid configuration. You need one of [rabbitmq|nats|redis], route-key-low and, route-key-high configured")
	}

	var input source
	var err error
	if hasRabbit && "" != rabbitmq {
		input, err = NewPwWsBrokerRabbit(rabbitmq, lowRoute, highRoute)
	} else if hasNats && "" != nats {
		input, err = NewPwWsBrokerNats(nats, lowRoute, highRoute)
	} else if hasRedis && "" != redis {
		input, err = NewPwWsBrokerRedis(redis, lowRoute, highRoute)
	}
	if nil != err {
		return err
	}

	broker, err := NewPlaneWatchWebSocketBroker(
		input,
		c.String("http-addr"),
		c.String("tls-cert"),
		c.String("tls-cert-key"),
		c.Bool("serve-test-web"),
		c.Duration("send-tick"),
	)
	if nil != err {
		return err
	}
	defer broker.Close()

	if err = broker.Setup(); nil != err {
		return err
	}
	go broker.Run()

	broker.Wait()

	return nil
}
