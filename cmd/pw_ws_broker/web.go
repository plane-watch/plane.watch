package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"net/http"
	"plane.watch/lib/tile_grid"
	"strconv"
	"strings"
	"time"
)

const MaxTickDuration = time.Second * 10

//go:embed test-web
var testWebDir embed.FS

var (
	gridJSONPayload     []byte
	gridJSONPayloadGzip []byte
	gridJSONPayloadETag string
)

func init() {
	var err error
	grid := tile_grid.GetGrid()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	gridJSONPayload, err = json.MarshalIndent(grid, "", "  ")
	if nil != err {
		log.Fatal().Err(err).Msg("Failed to json encode the tile grid")
	}

	gridJSONPayloadETag = fmt.Sprintf(`"%X"`, sha256.Sum224(gridJSONPayload))
	gridJSONPayloadGzip = mustGzipBytes(gridJSONPayload)
}

// logRequest logs all the web requests we get
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("Remote", r.RemoteAddr).Str("Request", r.RequestURI).Msg("Web RQ")
		handler.ServeHTTP(w, r)
	})
}

// mustGzipBytes is a helper function that dies if there is an error GZIP'ing a byte stream
func mustGzipBytes(in []byte) []byte {
	// make a gzip version
	buf := bytes.Buffer{}
	gzw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if nil != err {
		log.Fatal().Err(err).Msg("Failed to create gzip writer")
	}
	_, err = gzw.Write(in)
	if nil != err {
		log.Fatal().Err(err).Msg("Failed to compress json")
	}

	if err = gzw.Close(); nil != err {
		log.Fatal().Err(err).Msg("Failed to gzip json")
	}

	return buf.Bytes()
}

func configureServeMuxCommon(
	serveMux *http.ServeMux,
	certKey, cert string,
	serveTest bool,
) error {
	if serveMux == nil {
		panic("Nil ServeMux given")
	}
	var domainsToServe []string
	serveMux.HandleFunc("/", indexPage)
	serveMux.HandleFunc("/grid", jsonGrid)

	if serveTest {
		serveMux.Handle(
			"/test-web/",
			logRequest(http.FileServer(http.FS(testWebDir))),
		)
	}

	if certKey != "" {
		log.Info().Str("cert", cert).Msg("Using Certificate")
		tlsCert, err := tls.LoadX509KeyPair(cert, certKey)
		if nil != err {
			return err
		}
		x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
		if nil != err {
			return err
		}
		for _, d := range x509Cert.DNSNames {
			domainsToServe = append(domainsToServe, d)
			domainsToServe = append(domainsToServe, d+":*")
		}
		if x509Cert.Subject.CommonName != "" {
			domainsToServe = append(domainsToServe, x509Cert.Subject.CommonName)
		}
	} else {
		domainsToServe = []string{
			"localhost",
			"localhost:3000",
			"localhost:30001",
			"*plane.watch",
			"*plane.watch:3000",
			"*plane.watch:3001",
		}
	}
	log.Info().
		Int("# Domains", len(domainsToServe)).
		Msg("Serving Domain Count")
	for _, d := range domainsToServe {
		log.Info().Str("domain", d).Msg("Serving For Domain")
	}
	return nil
}

// indexPage gives people something to look at if they ask us for the index
func indexPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("Plane.Watch Websocket Broker"))
}

func jsonGrid(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("If-None-Match") == gridJSONPayloadETag {
		w.WriteHeader(304)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("ETag", gridJSONPayloadETag)
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(len(gridJSONPayloadGzip)))
		_, _ = w.Write(gridJSONPayloadGzip)
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(gridJSONPayload)))
		_, _ = w.Write(gridJSONPayload)
	}
}
