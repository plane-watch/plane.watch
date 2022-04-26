package nats_io

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/url"
)

type (
	healthItem struct {
		name string
		ch   chan *nats.Msg
	}

	Server struct {
		url      string
		incoming *nats.Conn
		outgoing *nats.Conn

		channels []healthItem

		log zerolog.Logger
	}
)

func NewServer(serverUrl string) (*Server, error) {
	n := &Server{
		log: log.With().Str("section", "nats.io").Logger(),
	}
	n.SetUrl(serverUrl)
	if err := n.Connect(); nil != err {
		return nil, err
	}
	return n, nil
}

func (n *Server) SetUrl(serverUrl string) {
	serverUrlParts, err := url.Parse(serverUrl)
	if nil == err {
		if "" == serverUrlParts.Port() {
			serverUrlParts.Host = net.JoinHostPort(serverUrlParts.Hostname(), "4222")
		}
	} else {
		log.Error().Err(err).Msg("invalid url")
	}
	n.url = serverUrlParts.String()
}

func (n *Server) Connect() error {
	var err error
	n.log.Debug().Str("url", n.url).Msg("connecting to server...")
	n.incoming, err = nats.Connect(n.url)
	if nil != err {
		n.log.Error().Err(err).Str("dir", "incoming").Msg("Unable to connect to NATS server")
		return err
	}
	n.outgoing, err = nats.Connect(n.url)
	if nil != err {
		n.log.Error().Err(err).Str("dir", "outgoing").Msg("Unable to connect to NATS server")
		return err
	}
	return nil
}

// Publish is our simple message publisher
func (n *Server) Publish(queue string, msg []byte) error {
	if n.outgoing.IsConnected() {
		return n.outgoing.Publish(queue, msg)
	}
	return nil
}

func (n *Server) Close() {
	if n.incoming.IsConnected() {
		if err := n.incoming.Drain(); nil != err {
			n.log.Error().Err(err).Str("dir", "incoming").Msg("failed to drain connection")
		}
	}
	n.outgoing.Close()
}

func (n *Server) Subscribe(subject string) (chan *nats.Msg, error) {
	ch := make(chan *nats.Msg, 512)
	n.channels = append(n.channels, healthItem{
		name: "subscription-" + subject,
		ch:   ch,
	})
	_, err := n.incoming.ChanSubscribe(subject, ch)
	if nil != err {
		return nil, err
	}
	return ch, nil
}

func (n *Server) HealthCheckName() string {
	return "Nats"
}

func (n *Server) HealthCheck() bool {
	for _, item := range n.channels {
		l := len(item.ch)
		c := cap(item.ch)
		p := (float32(l) / float32(c)) * 100
		n.log.Info().
			Int("# items", l).
			Int("max items", c).
			Float32("%", p).
			Str("channel", item.name).
			Msg("Channel Check")
	}
	return n.incoming.IsConnected() && n.outgoing.IsConnected()
}
