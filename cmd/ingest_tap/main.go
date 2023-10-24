package main

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"plane.watch/lib/logging"
	"plane.watch/lib/tracker/beast"
	"syscall"
)

const (
	natsURL      = "nats"
	websocketURL = "websocket-url"
	feederAPIKey = "api-key"
	icao         = "icao"
	logFile      = "file"
)

func main() {
	app := cli.NewApp()
	app.Name = "pw_ingest natsTap"
	app.Description = "Ask all of the pw_ingest instances on the nats bus to feed matching data to this terminal.\n" +
		"Matching beast/avr/sbs1 messages will be forwarded to you"

	app.Commands = cli.Commands{
		{
			Name:  "log",
			Usage: "Logs matching data to a file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  logFile,
					Value: "captured-frames",
					Usage: "[.beast|.avr|.sbs1] will be appended to file",
				},
			},
			Action: logMatching,
		},
		{
			Name:   "tui",
			Usage:  "Shows an Interactive Text User Interface",
			Action: runTui,
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  natsURL,
			Usage: "nats url",
			Value: "nats://localhost:4222/",
		},
		&cli.StringFlag{
			Name:  websocketURL,
			Usage: "Plane Watch WebSocket Api URL",
			Value: "https://localhost/planes",
		},
		&cli.StringFlag{
			Name:  icao,
			Usage: "if specified, only frames from the plane with the specified ICAO will be sent",
		},
		&cli.StringFlag{
			Name:  feederAPIKey,
			Usage: "the plane.watch api UUID for the given feeder you want to investigate",
		},
	}
	logging.IncludeVerbosityFlags(app)
	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)

		return nil
	}

	if err := app.Run(os.Args); nil != err {
		log.Error().Err(err).Send()
	}
}

func logMatching(c *cli.Context) error {
	logging.ConfigureForCli()

	tapper := NewPlaneWatchTapper(WithLogger(log.Logger))
	if err := tapper.Connect(c.String(natsURL), ""); err != nil {
		return err
	}
	defer tapper.Disconnect()

	fileHandles := make(map[string]*os.File)
	exts := []string{"beast", "avr", "sbs1"}
	for _, ext := range exts {
		fileName := c.String(logFile) + "." + ext
		fh, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
		if nil != err {
			log.Error().Err(err).Str("filename", fileName).Send()
			return err
		}
		fileHandles[ext] = fh
	}

	if err := tapper.IncomingDataTap(c.String(icao), c.String(feederAPIKey), handleStream(fileHandles)); err != nil {
		return err
	}

	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)
	<-chSignal // wait for our cancel signal
	log.Info().Msg("Shutting down")

	for ext, f := range fileHandles {
		err := f.Close()
		if err != nil {
			log.Error().Err(err).Str("ext", ext).Msg("Failed to close file")
		}
	}

	log.Info().Msg("Shut down complete")
	return nil
}

func handleStream(fileHandles map[string]*os.File) IngestTapHandler {
	return func(frameType, tag string, msgData []byte) {
		switch frameType {
		case "beast":
			b, err := beast.NewFrame(msgData, false)
			if nil != err {
				log.Error().Err(err)
				return
			}
			log.Info().
				Str("ICAO", b.IcaoStr()).
				Str("AVR", b.RawString()).
				Str("tag", tag).
				Send()

			// TODO: Replace timestamp with our own
			_, err = fileHandles["beast"].Write(msgData)
			if err != nil {
				log.Error().Err(err).Send()
			}
		case "avr":
			_, err := fileHandles["avr"].Write(append(msgData, 10)) // new line append
			if err != nil {
				log.Error().Err(err).Send()
			}
		case "sbs1":
			_, err := fileHandles["sbs1"].Write(append(msgData, 10))
			if err != nil {
				log.Error().Err(err).Send()
			}
		}
	}
}
