package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"html"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"plane.watch/lib/export"
	"plane.watch/lib/logging"
	"plane.watch/lib/monitoring"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/mode_s"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// this is a website where you put in one or more Mode S frames and they are decoded
// in a way that is informational

type (
	decodeResponse struct {
		Err         string
		Description string
		Payloads    []string
	}
)

func main() {
	app := cli.NewApp()
	app.Description = "Decode your AVR frames, on the web!"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "listen-http",
			Aliases: []string{"port"},
			Value:   ":8080",
			Usage:   "Port to run the website on (http)",
		},
		&cli.StringFlag{
			Name:  "listen-https",
			Value: ":8443",
			Usage: "Port to run the website on (https)",
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
	}
	logging.IncludeVerbosityFlags(app)
	monitoring.IncludeMonitoringFlags(app, 9605)

	app.Before = func(c *cli.Context) error {
		logging.SetLoggingLevel(c)
		return nil
	}

	app.Action = runHttpServer
	logging.ConfigureForCli()

	if err := app.Run(os.Args); nil != err {
		log.Error().Err(err).Send()
	}
}

func (dr *decodeResponse) Bytes() []byte {
	msg, err := json.Marshal(dr)
	if nil != err {
		log.Error().Err(err).Msg("Failed Encoding response")
	}
	return msg
}

func runHttpServer(c *cli.Context) error {
	cert := c.String("tls-cert")
	certKey := c.String("tls-cert-key")
	if ("" != cert || "" != certKey) && ("" == cert || "" == certKey) {
		return errors.New("please provide both certificate and key")
	}
	monitoring.RunWebServer(c)
	var htdocsPath string
	var err error
	var files fs.FS
	if c.NArg() == 0 {
		log.Info().Msg("Using our embedded filesystem")
		files, err = fs.Sub(embeddedHtdocs, "htdocs")
		if nil != err {
			panic(err)
		}
	} else {
		htdocsPath = path.Clean(c.Args().First())
		log.Info().Str("path", htdocsPath).Msg("Serving HTTP Docs from")
		files = os.DirFS(htdocsPath)
	}
	logging.SetLoggingLevel(c)

	http.Handle("/", http.FileServer(http.FS(files)))

	http.HandleFunc("/decode", decode)

	exitChan := make(chan bool)
	go listenHttp(exitChan, c.String("listen-http"))
	if "" != cert {
		go listenHttps(exitChan, cert, certKey, c.String("listen-https"))
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	select {
	case <-exitChan:
		log.Error().Msg("There as an error and we are exiting")
	case <-sc:
		log.Debug().Msg("Kill Signal Received")
	}

	return nil
}

func decode(w http.ResponseWriter, r *http.Request) {
	resp := decodeResponse{
		Payloads: make([]string, 0),
	}
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		_, errResp := w.Write(resp.Bytes())
		if errResp != nil {
			log.Error().Err(errResp).Msg("Failed to write response")
			return
		}
	}()
	defer func() {
		if rec := recover(); rec != nil {
			resp.Err = "Failed massively!"
			_, _ = w.Write(resp.Bytes())
			log.Error().
				Stack().
				Err(errors.New("failed to handle request")).
				Interface("recover", rec)
		}
	}()
	switch r.URL.Path {
	case "/decode":
		pt := tracker.NewTracker()
		var submittedPackets string
		_ = r.ParseForm()

		refLat := getAsFloat(r.FormValue("refLat"))
		refLon := getAsFloat(r.FormValue("refLon"))
		submittedPackets = r.FormValue("packet")
		if "" == submittedPackets {
			resp.Err = "No Packet Provided"
			return
		}
		packets := strings.Split(submittedPackets, ";")
		icaoList := make(map[uint32]uint32)

		var desc string
		buf := bytes.NewBufferString(desc)
		for _, packet := range packets {
			packet = strings.TrimSpace(packet)
			if "" == packet {
				continue
			}
			log.Debug().Str("frame", packet).Msg("Decoding Frame")
			frame, err := mode_s.DecodeString(packet, time.Now())
			if err != nil {
				resp.Err = fmt.Sprintf("Failed to decode: %s", html.EscapeString(err.Error()))
				return
			}
			if nil == frame {
				resp.Err = fmt.Sprintf("Not an AVR Frame: %s", html.EscapeString(err.Error()))
				return
			}
			pt.GetPlane(frame.Icao()).HandleModeSFrame(frame, refLat, refLon)
			icaoList[frame.Icao()] = frame.Icao()
			frame.Describe(buf)
		}
		resp.Description = buf.String()

		for _, icao := range icaoList {
			plane := pt.GetPlane(icao)
			pl := export.NewPlaneLocation(plane, "")
			encoded, _ := pl.ToJsonBytes() //json.MarshalIndent(&pl, "", "  ")

			resp.Payloads = append(resp.Payloads, string(encoded))
		}

		pt = nil
	default:
		http.NotFound(w, r)
		_, _ = fmt.Fprintln(w, "<br/>\n"+r.RequestURI)
	}

}

func listenHttp(exitChan chan bool, port string) {
	log.Info().Msgf("Decode Listening on %s...", port)
	if err := http.ListenAndServe(port, nil); nil != err {
		if http.ErrServerClosed != err {
			log.Error().Err(err).Send()
		}
	}
	exitChan <- true
}

func listenHttps(exitChan chan bool, cert, certKey, port string) {
	log.Info().Msgf("Decode Listening on %s...", port)
	if err := http.ListenAndServeTLS(port, cert, certKey, nil); nil != err {
		if http.ErrServerClosed != err {
			log.Error().Err(err).Send()
		}
	}
	exitChan <- true
}

func getAsFloat(in string) *float64 {
	f, err := strconv.ParseFloat(in, 64)
	if nil == err {
		return &f
	}
	return nil
}
