package main

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"plane.watch/lib/logging"
	"plane.watch/lib/nats_io"
	"plane.watch/lib/randstr"
	"plane.watch/lib/tracker/beast"
	"sync"
	"syscall"
	"time"
)

const (
	natsUrl = "nats"
	apiKey  = "api-key"
	icao    = "icao"
	logFile = "file"
)

func main() {
	app := cli.NewApp()
	app.Name = "pw_ingest tap"
	app.Description = "Ask all of the pw_ingest instances on the nats bus to feed matching data to this terminal.\n" +
		"Matching beast/avr/sbs1 messages will be forwarded to you"

	app.Commands = cli.Commands{
		{
			Name:        "log",
			Description: "Logs matching data to a file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  icao,
					Usage: "if specified, only frames from the plane with the specified ICAO will be sent",
				},
				&cli.StringFlag{
					Name:  apiKey,
					Usage: "the plane.watch api UUID for the given feeder you want to investigate",
				},
				&cli.StringFlag{
					Name:  logFile,
					Value: "captured-frames",
					Usage: "[.beast|.avr|.sbs1] will be appended to file",
				},
			},
			Action: logMatching,
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  natsUrl,
			Usage: "nats url",
			Value: "nats://localhost:4222/",
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

	natsSvr, err := nats_io.NewServer(
		nats_io.WithConnections(true, true),
		nats_io.WithServer(c.String(natsUrl), "ingest-tap"),
	)
	if nil != err {
		return err
	}

	tapSubject := "ingest-tap-" + randstr.RandString(20)
	ch, err := natsSvr.Subscribe(tapSubject)

	var wg sync.WaitGroup
	wg.Add(1)

	go handleStream(ch, fileHandles, &wg)

	headers := map[string]string{
		"action":  "add",
		"api-key": c.String(apiKey),
		"icao":    c.String(icao),
		"subject": tapSubject,
	}

	response, errRq := natsSvr.Request("v1.pw-ingest.tap", []byte{}, headers, time.Second)
	if nil != errRq {
		log.Error().Err(err).Msg("Failed to request tap")
		return err
	}
	log.Debug().Str("response", string(response)).Msg("request response")

	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)
	<-chSignal // wait for our cancel signal
	log.Info().Msg("Shutting down")

	// ask to step sending
	headers["action"] = "remove"
	response, errRq = natsSvr.Request("v1.pw-ingest.tap", []byte{}, headers, time.Second)
	if nil != errRq {
		log.Error().Err(err).Msg("Failed to stop tap")
		return err
	}
	log.Debug().Str("response", string(response)).Msg("request response")

	close(ch)
	wg.Wait()

	for ext, f := range fileHandles {
		err = f.Close()
		if err != nil {
			log.Error().Err(err).Str("ext", ext).Msg("Failed to close file")
		}
	}

	log.Info().Msg("Shut down complete")
	return nil
}

func handleStream(ch chan *nats.Msg, fileHandles map[string]*os.File, wg *sync.WaitGroup) {
	for msg := range ch {
		switch msg.Header.Get("type") {
		case "beast":
			b, err := beast.NewFrame(msg.Data, false)
			log.Info().
				Str("ICAO", b.IcaoStr()).
				Str("AVR", b.RawString()).
				Str("tag", msg.Header.Get("tag")).
				Send()

			// TODO: Replace timestamp with our own
			_, err = fileHandles["beast"].Write(msg.Data)
			if err != nil {
				log.Error().Err(err).Send()
			}
		case "avr":
			_, err := fileHandles["avr"].Write(append(msg.Data, 10))
			if err != nil {
				log.Error().Err(err).Send()
			}
		case "sbs1":
			_, err := fileHandles["sbs1"].Write(append(msg.Data, 10))
			if err != nil {
				log.Error().Err(err).Send()
			}

		}
	}
	wg.Done()
}
