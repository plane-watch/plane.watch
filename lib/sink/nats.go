package sink

import (
	"github.com/rs/zerolog/log"
	"net"
	"net/url"
	"plane.watch/lib/nats_io"
	"plane.watch/lib/tracker"
	"regexp"
	"time"
)

type (
	NatsSink struct {
		Config
		server *nats_io.Server
	}
)

func NewNatsSink(opts ...Option) (tracker.Sink, error) {
	n := &NatsSink{}
	n.setupConfig(opts)
	delay := time.Second
	for {
		err := n.connect()
		if nil == err {
			break
		}
		// failed connection handling
		if delay > time.Second*10 {
			return nil, err
		}

		log.Error().Err(err).Dur("retry delay", delay).Msg("Unable to setup nats sink. Trying again")
		time.Sleep(delay)
		delay += time.Second
	}
	return NewSink(&n.Config, n), nil
}

func (n *NatsSink) connect() error {
	var err error
	serverUrl := url.URL{
		Scheme:  "nats", // tls for secure
		User:    url.UserPassword(n.user, n.pass),
		Host:    net.JoinHostPort(n.host, n.port),
		Path:    "",
		RawPath: "",
	}
	re := regexp.MustCompile(`/\s/`)
	st := re.ReplaceAllString(n.sourceTag, "_")
	n.server, err = nats_io.NewServer(
		nats_io.WithServer(serverUrl.String(), n.connectionName+"+source="+st),
		nats_io.WithConnections(false, true),
	)
	return err
}

func (n *NatsSink) PublishJson(subject string, msg []byte) error {
	return n.server.Publish(subject, msg)
}

func (n *NatsSink) Stop() {
	n.server.Close()
}

func (n *NatsSink) HealthCheck() bool {
	return n.server.HealthCheck()
}

func (n *NatsSink) HealthCheckName() string {
	return n.server.HealthCheckName()
}
