package setup

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"net/url"
	"plane.watch/lib/sink"
	"plane.watch/lib/tracker"
	"strings"
	"time"
)

const (
	Sink             = "sink"
	SinkEncoding     = "sink-encoding"
	SinkCollectDelay = "sink-collect-delay"
)

var (
	prometheusOutputFrame = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ingest_output_frame_total",
		Help: "The total number of raw frames output. (no dedupe)",
	})
	prometheusOutputPlaneLocation = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pw_ingest_output_location_update_total",
		Help: "The total number of plane location events output.",
	})
)

func IncludeSinkFlags(app *cli.App) {
	app.Flags = append(app.Flags, []cli.Flag{
		&cli.StringFlag{
			Name:    Sink,
			Usage:   "The place to send decoded JSON in URL Form. nats://user:pass@host:port/vhost?ttl=60",
			EnvVars: []string{"SINK"},
		},
		&cli.StringFlag{
			Name:    SinkEncoding,
			Usage:   "The data serialisation method. 'json' or 'protobuf'",
			Value:   "json",
			EnvVars: []string{"SINK_ENCODING"},
		},
		&cli.DurationFlag{
			Name:    SinkCollectDelay,
			Value:   300 * time.Millisecond,
			Usage:   "Instead of emitting an update for every update we get, collect updates and send a deduplicated list (based on icao) every period",
			EnvVars: []string{"SINK_COLLECT_DELAY"},
		},
	}...)
}

func HandleSinkFlag(c *cli.Context, connName string) (tracker.Sink, error) {
	sinkURL := c.String(Sink)
	defaultTag := c.String(Tag)
	sinkEncoding := strings.ToLower(c.String(SinkEncoding))
	defaultDelay := c.Duration(SinkCollectDelay)

	if sinkEncoding != sink.EncodingJSON && sinkEncoding != sink.EncodingProtobuf {
		return nil, fmt.Errorf("sink Encoding must be one of [%s, %s]", sink.EncodingJSON, sink.EncodingProtobuf)
	}

	log.Debug().Str("sink-url", sinkURL).Msg("With Sink")
	s, err := handleSink(connName, sinkURL, defaultTag, sinkEncoding, defaultDelay)
	if nil != err {
		log.Error().Err(err).Str("url", sinkURL).Str("what", "sink").Msg("Failed setup sink")
		return nil, err
	}

	return s, nil
}

func handleSink(connName, urlSink, defaultTag, sinkEncoding string, sendDelay time.Duration) (tracker.Sink, error) {
	parsedURL, err := url.Parse(urlSink)
	if nil != err {
		return nil, err
	}

	urlPass, _ := parsedURL.User.Password()

	commonOpts := []sink.Option{
		sink.WithConnectionName(connName),
		sink.WithHost(parsedURL.Hostname(), parsedURL.Port()),
		sink.WithUserPass(parsedURL.User.Username(), urlPass),
		sink.WithSourceTag(getTag(parsedURL, defaultTag)),
		sink.WithPrometheusCounters(prometheusOutputFrame, prometheusOutputPlaneLocation),
		sink.WithSendDelay(sendDelay),
		sink.WithEncoding(sinkEncoding),
	}

	switch strings.ToLower(parsedURL.Scheme) {
	case "nats", "nats.io":
		return sink.NewNatsSink(commonOpts...)

	default:
		return nil, fmt.Errorf("unknown scheme: %s, expected nats://", parsedURL.Scheme)
	}

}
