package main

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"plane.watch/lib/dedupe"
	"plane.watch/lib/logging"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/setup"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
	"strconv"
	"sync"
)

const (
	typeBeast = "frames.beast"
	typeAvr   = "frames.avr"
	typeSbs1  = "frames.sbs1"
)

type (
	frameProcessor struct {
		filterIcaos    []uint32
		events         chan tracker.Event
		logFileHandles sync.Map
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "Frame Recorder"
	app.Usage = "Records incoming streams to disk"
	app.Description = "Records Beast/AVR/SBS1 source to a file, optionally filtered"

	app.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "filter-icao",
			Usage: "when specified, limits the ICAOs to the chosen aircraft",
		},
	}

	setup.IncludeSourceFlags(app)
	logging.IncludeVerbosityFlags(app)
	monitoring.IncludeMonitoringFlags(app, 9606)

	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)
		return nil
	}

	app.Action = runCli

	if err := app.Run(os.Args); nil != err {
		log.Error().Err(err).Msg("did not complete successfully")
		os.Exit(1)
	}
}

func runCli(c *cli.Context) error {
	logging.ConfigureForCli()
	monitoring.RunWebServer(c)

	trackerOpts := make([]tracker.Option, 0)
	trk := tracker.NewTracker(trackerOpts...)

	producers, err := setup.HandleSourceFlags(c)
	if 0 == len(producers) {
		cli.ShowAppHelpAndExit(c, 1)
	}
	if nil != err {
		return err
	}
	for _, p := range producers {
		trk.AddProducer(p)
	}

	// have our own middlewares
	trk.AddMiddleware(dedupe.NewFilter())

	icaos := make([]uint32, 0)
	for _, icao := range c.StringSlice("filter-icaos") {
		ident, err := strconv.ParseUint(icao, 16, 32)
		if nil != err {
			log.Warn().Str("given", icao).Msg("Unable to decode icao")
		}
		log.Info().Uint64("icao", ident).Str("str", icao).Msg("Filtering ICAO")
		icaos = append(icaos, uint32(ident))
	}
	fp := newFrameLogger(icaos)
	trk.AddMiddleware(fp)

	// setup a file sync?

	trk.Wait()
	return nil
}

func newFrameLogger(filterIcaos []uint32) *frameProcessor {
	return &frameProcessor{
		filterIcaos: filterIcaos,
		events:      make(chan tracker.Event),
	}
}

func (fp *frameProcessor) Stop() {
	close(fp.events)
	fp.logFileHandles.Range(func(name, handle interface{}) bool {
		fp.logFileHandles.Delete(name)
		_ = handle.(*os.File).Close()
		return true
	})
}

func (fp *frameProcessor) String() string {
	return "Recorder Frame Processor"
}

func (fp *frameProcessor) Listen() chan tracker.Event {
	return fp.events
}
func (fp *frameProcessor) Handle(frame tracker.Frame) tracker.Frame {
	if nil == frame {
		return nil
	}
	if len(fp.filterIcaos) > 0 {
		found := false
		for _, icao := range fp.filterIcaos {
			if frame.Icao() == icao {
				found = true
				break
			}
		}
		if !found {
			return frame
		}
	}

	write := func(fileName string, b []byte) {
		fh, ok := fp.logFileHandles.Load(fileName)
		if !ok {
			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if nil != err {
				log.Error().Err(err).Bytes("frame", frame.Raw()).Msg("Failed to log frame")
				return
			}
			fp.logFileHandles.Store(fileName, f)
			fh = f
		}

		if _, err := fh.(*os.File).Write(b); nil != err {
			log.Error().Err(err).Msg("Failed to write frame to log file")
		}
	}

	switch f := (frame).(type) {
	case *beast.Frame:
		write(typeBeast, frame.Raw())
		write(typeAvr, append(f.AvrFrame().Raw(), 0x0A))
	case *mode_s.Frame:
		write(typeAvr, frame.Raw())
	case *sbs1.Frame:
		write(typeSbs1, frame.Raw())
	default:
	}

	return frame
}
