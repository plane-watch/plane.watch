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
		&cli.StringSliceFlag{
			Name:    "sink",
			Usage:   "The place to send decoded JSON in URL Form. nats://user:pass@host:port/vhost?ttl=60",
			EnvVars: []string{"SINK"},
		},
		&cli.DurationFlag{
			Name:    "sink-collect-delay",
			Value:   300 * time.Millisecond,
			Usage:   "Instead of emitting an update for every update we get, collect updates and send a deduplicated list (based on icao) every period",
			EnvVars: []string{"SINK_COLLECT_DELAY"},
		},
	}...)
}

func HandleSinkFlags(c *cli.Context, connName string) ([]tracker.Sink, error) {
	defaultDelay := c.Duration("sink-collect-delay")
	defaultTag := c.String("tag")
	sinks := make([]tracker.Sink, 0)

	for _, sinkUrl := range c.StringSlice("sink") {
		log.Debug().Str("sink-url", sinkUrl).Msg("With Sink")
		s, err := handleSink(connName, sinkUrl, defaultTag, defaultDelay)
		if nil != err {
			log.Error().Err(err).Str("url", sinkUrl).Str("what", "sink").Msg("Failed setup sink")
			return nil, err
		} else {
			sinks = append(sinks, s)
		}
	}
	return sinks, nil
}

func handleSink(connName, urlSink, defaultTag string, sendDelay time.Duration) (tracker.Sink, error) {
	parsedUrl, err := url.Parse(urlSink)
	if nil != err {
		return nil, err
	}

	urlPass, _ := parsedUrl.User.Password()

	commonOpts := []sink.Option{
		sink.WithConnectionName(connName),
		sink.WithHost(parsedUrl.Hostname(), parsedUrl.Port()),
		sink.WithUserPass(parsedUrl.User.Username(), urlPass),
		sink.WithSourceTag(getTag(parsedUrl, defaultTag)),
		sink.WithPrometheusCounters(prometheusOutputFrame, prometheusOutputPlaneLocation),
		sink.WithSendDelay(sendDelay),
	}

	switch strings.ToLower(parsedUrl.Scheme) {
	case "nats", "nats.io":
		return sink.NewNatsSink(commonOpts...)

	default:
		return nil, fmt.Errorf("unknown scheme: %s, expected nats://", parsedUrl.Scheme)
	}

}
