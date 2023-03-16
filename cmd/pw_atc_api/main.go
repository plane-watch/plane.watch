package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"plane.watch/lib/logging"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/nats_io"
	"runtime"
	"time"
)

var (
	version = "dev"
	db      *sqlx.DB
)

func main() {
	app := cli.NewApp()

	app.Version = version
	app.Name = "Plane Watch ATC API Server"
	app.Usage = "Listens on NATS bus for requests and responds to them"

	logging.IncludeVerbosityFlags(app)
	monitoring.IncludeMonitoringFlags(app, 9602)

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
			Name:    "nats",
			Usage:   "Nats.io URL for fetching and publishing updates. nats://guest:guest@host:4222/",
			EnvVars: []string{"NATS"},
		},
		&cli.StringFlag{
			Name:    "db-host",
			Usage:   "Database Host",
			EnvVars: []string{"DATABASE_HOST"},
		},
		&cli.StringFlag{
			Name:    "db-port",
			Usage:   "Database Port",
			EnvVars: []string{"DATABASE_PORT"},
		},
		&cli.StringFlag{
			Name:    "db-user",
			Usage:   "Database User",
			EnvVars: []string{"DATABASE_USER"},
		},
		&cli.StringFlag{
			Name:    "db-pass",
			Usage:   "Database Pass",
			EnvVars: []string{"DATABASE_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    "db-name",
			Usage:   "Database Name",
			EnvVars: []string{"DATABASE_NAME"},
		},
		&cli.IntFlag{
			Name:    "num-workers",
			Usage:   "How many workers to spin up",
			Value:   runtime.NumCPU(),
			EnvVars: []string{"NUM_WORKERS"},
		},
	}
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

func testSearch(server *nats_io.Server) {
	for {
		time.Sleep(time.Second)
		r, err := server.Request("search.airport", []byte("YPP"), time.Second)
		if nil != err {
			log.Error().Err(err).Send()
		}

		fmt.Println("Reply", string(r))

	}
}

func connectDatabase(c *cli.Context) error {
	host := c.String("db-host")
	port := c.String("db-port")
	user := c.String("db-user")
	pass := c.String("db-pass")
	dbName := c.String("db-name")
	var err error
	s := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbName)
	log.Info().Str("conn", s).Send()
	db, err = sqlx.Connect("postgres", s)

	if nil != err {
		return err
	}
	db.DB.SetMaxOpenConns(50) // because samfty said so
	db.DB.SetMaxIdleConns(10)
	return nil
}

func run(c *cli.Context) error {
	log.Info().Msg("Starting up")
	exitChan := make(chan bool)
	if err := connectDatabase(c); nil != err {
		return err
	}

	server, err := nats_io.NewServer(c.String("nats"), "pw_atc_api")
	if nil != err {
		return err
	}

	//go testSearch(server)

	numWorkers := c.Int("num-workers")
	for i := 0; i < numWorkers; i++ {
		go searchHandler(server, i, exitChan)
	}

	select {}
	return nil
}
