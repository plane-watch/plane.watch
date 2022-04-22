package main

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"plane.watch/cmd/pw_discord_bot/config"
	"plane.watch/lib/logging"
	"plane.watch/lib/mapping"
	"plane.watch/lib/monitoring"
)

var version = "dev"

func main() {
	app := cli.NewApp()

	app.Name = "PlaneWatch Discord Bot"

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
			Name:  "config-file",
			Usage: "Provides the location of the config file",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "nuke-commands",
			Usage: "Use the command to re-register all Discord bot commands",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "host",
			Usage: "The default host to talk to for our websocket goodness",
			Value: "localhost:80",
		},
		&cli.BoolFlag{
			Name:  "insecure",
			Usage: "Use this if your connection is not TLS protected",
			Value: false,
		},
	}

	logging.IncludeVerbosityFlags(app)
	monitoring.IncludeMonitoringFlags(app, 9604)

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
	monitoring.RunWebServer(c)

	conf := config.Load(c.String("config-file"))
	if nil == conf {
		return errors.New("failed to load config")
	}

	pwBot, err := NewPlaneWatchBot(conf.Token)
	if nil != err {
		return err
	}
	mapping.SetHereMapsApiKey(conf.HereMapsApiKey)

	return pwBot.Run(c)
}
