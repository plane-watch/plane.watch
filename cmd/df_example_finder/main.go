package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"plane.watch/lib/logging"
	"plane.watch/lib/setup"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"strings"
	"sync"
	"time"
)

func incoming(c *cli.Context) (chan tracker.Frame, error) {
	producers, err := setup.HandleSourceFlags(c)
	log.Info().Int("Num Sources", len(producers)).Send()
	if nil != err {
		return nil, err
	}
	out := make(chan tracker.Frame)
	wg := sync.WaitGroup{}
	wg.Add(1)

	for _, producer := range producers {

		go func(p tracker.Producer) {
			wg.Add(1)
			log.Debug().
				Bool("Healthy?", p.HealthCheck()).
				Str("Source", p.String()).
				Msg("Starting Read from Producer")
			for e := range p.Listen() {
				log.Debug().Str("type", e.Type()).Str("event", e.String()).Send()
				switch e.(type) {
				case *tracker.FrameEvent:
					out <- e.(*tracker.FrameEvent).Frame()
				}
			}
			wg.Done()
		}(producer)
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		wg.Wait()
		close(out)
	}()

	wg.Done()
	return out, nil
}

func modeSFrame(iframe tracker.Frame) *mode_s.Frame {
	if err := iframe.Decode(); nil != err {
		log.Error().Err(err).Str("frame", fmt.Sprintf("%X", iframe.Raw())).Send()
	}
	switch iframe.(type) {
	case *mode_s.Frame:
		return iframe.(*mode_s.Frame)
	case *beast.Frame:
		return iframe.(*beast.Frame).AvrFrame()
	}
	return nil
}

func gatherSamples(c *cli.Context) error {
	incomingChan, err := incoming(c)
	if nil != err {
		return err
	}
	log.Info().Msg("Processing...")

	countMap := make(map[byte]uint32)
	df17Map := make(map[byte]uint32)
	bdsMap := make(map[string]uint32)
	samples := make(map[byte][]string)
	existingSamples := make(map[string]bool)

	for iframe := range incomingChan {
		frame := modeSFrame(iframe)
		if nil == frame {
			continue
		}

		countMap[frame.DownLinkType()]++

		switch frame.DownLinkType() {
		case 17:
			df17Map[frame.MessageType()]++
			key := fmt.Sprintf("DF17/%d", frame.MessageType())
			if _, ok := existingSamples[key]; ok {
				continue
			}
			existingSamples[key] = true
		case 20, 21:
			bdsMap[frame.BdsMessageType()]++
			if "0.0" == frame.BdsMessageType() {
				continue
			}
		}

		if len(samples[frame.DownLinkType()]) < 100 {
			if _, exist := existingSamples[frame.RawString()]; !exist {
				samples[frame.DownLinkType()] = append(samples[frame.DownLinkType()], frame.RawString())
				existingSamples[frame.RawString()] = true
			}
		}
	}

	println("Frame Type Counts")
	for k, c := range countMap {
		println("DF", k, "=\t", c)
	}
	println("DF17 Frame Breakdown")
	for k, c := range df17Map {
		println("DF17 Type", k, "=\t", c)
	}
	println("DF 20/21 BDS Frame Breakdown")
	for k, c := range bdsMap {
		println("BDS Type", k, "=\t", c)
	}

	println("Sample Frames")
	for k, s := range samples {
		println(k, ":", "['"+strings.Join(s, "', '")+"'],")
	}
	return nil
}

func getFlagByte(c *cli.Context, flag string) *byte {
	for _, f := range c.FlagNames() {
		if f == flag {
			v := c.Int(flag)
			b := byte(v)
			return &b
		}
	}
	return nil
}

func showTypes(c *cli.Context) error {
	incomingChan, err := incoming(c)
	if nil != err {
		return err
	}
	log.Info().Msg("Processing...")

	requestedDf := getFlagByte(c, "df")
	requestedMt := getFlagByte(c, "mt")
	requestedSt := getFlagByte(c, "st")
	var requestedIcao *string
	if v := c.String("icao"); "" != v {
		requestedIcao = &v
	}
	export := c.Bool("export")

	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.SetHeader([]string{"DF", "MT", "ST", "ICAO", "AVR", "DF Desc", "MT Desc"})
	tbl.SetBorder(false)
	tbl.SetAutoWrapText(false)
	exportedFrames := make([]string, 0, 1000)

	for iframe := range incomingChan {
		frame := modeSFrame(iframe)
		if nil == frame {
			continue
		}

		if nil != requestedDf && *requestedDf != frame.DownLinkType() {
			continue
		}
		if nil != requestedMt && *requestedMt != frame.MessageType() {
			continue
		}
		if nil != requestedSt && *requestedSt != frame.MessageSubType() {
			continue
		}
		if nil != requestedIcao && *requestedIcao != frame.IcaoStr() {
			continue
		}
		var fields []string
		exportedFrames = append(exportedFrames, frame.RawString())

		switch frame.DownLinkType() {
		case 0, 4, 5, 11:
			fields = []string{
				fmt.Sprintf("%02d", frame.DownLinkType()),
				"",
				"",
				frame.IcaoStr(),
				frame.RawString(),
				frame.DownLinkFormat(),
				frame.MessageTypeString(),
			}
		case 17, 18, 19:
			fields = []string{
				fmt.Sprintf("%02d", frame.DownLinkType()),
				fmt.Sprintf("%02d", frame.MessageType()),
				fmt.Sprintf("%02d", frame.MessageSubType()),
				frame.IcaoStr(),
				frame.RawString(),
				frame.DownLinkFormat(),
				frame.MessageTypeString(),
			}
		case 20, 21:
			fields = []string{
				fmt.Sprintf("%02d", frame.DownLinkType()),
				frame.BdsMessageType(),
				fmt.Sprintf("%02d", frame.MessageSubType()),
				frame.IcaoStr(),
				frame.RawString(),
				frame.DownLinkFormat(),
				frame.MessageTypeString(),
			}
		default:
			fields = []string{
				fmt.Sprintf("%02d", frame.DownLinkType()),
				fmt.Sprintf("%02d", frame.MessageType()),
				fmt.Sprintf("%02d", frame.MessageSubType()),
				frame.IcaoStr(),
				frame.RawString(),
				frame.DownLinkFormat(),
				frame.MessageTypeString(),
			}
		}
		if !export {
			tbl.Append(fields)
		}
	}
	if export {
		for _, f := range exportedFrames {
			fmt.Println(f)
		}
	} else {
		tbl.Render()
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Name = "DF Example Finder"
	app.Usage = "Find examples of payloads in a file"

	setup.IncludeSourceFlags(app)
	logging.IncludeVerbosityFlags(app)

	app.Commands = []*cli.Command{
		{
			Name:   "types",
			Usage:  "Shows message info for everything in the file",
			Action: showTypes,
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "df",
					Usage: "Only print frames of the specified Downlink format",
				},
				&cli.IntFlag{
					Name:  "mt",
					Usage: "Message Type (for df 17,18,19)",
				},
				&cli.IntFlag{
					Name:  "st",
					Usage: "Message Sub Type (for df 17,18,19)",
				},
				&cli.StringFlag{
					Name:  "icao",
					Usage: "only show messages from Airframe with this 24bit identifier (hex, e.g. 7C6CE8)",
				},
				&cli.BoolFlag{
					Name:  "export",
					Usage: "Output only the AVR frames, one per line",
				},
			},
		},
		{
			Name:      "gather-samples",
			Usage:     "Gather Samples and put them in a JSON array ready for use in website_decode",
			Action:    gatherSamples,
			ArgsUsage: "[app.log - A file name to output to or stdout if not specified]",
		},
	}

	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)
		logging.ConfigureForCli()

		return nil
	}

	if err := app.Run(os.Args); nil != err {
		log.Error().Err(err).Msg("Finishing with an error")
		os.Exit(1)
	}
}
