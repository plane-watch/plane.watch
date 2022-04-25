package nats_io

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"net"
	"net/url"
)

type Server struct {
	url      string
	incoming *nats.Conn
	outgoing *nats.Conn

	channels []chan *nats.Msg
}

func NewServer(serverUrl string) (*Server, error) {
	n := &Server{}
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
	log.Debug().Str("url", n.url).Msg("connecting to nats.io server...")
	n.incoming, err = nats.Connect(n.url)
	if nil != err {
		log.Error().Err(err).Str("dir", "incoming").Msg("Unable to connect to NATS server")
		return err
	}
	n.outgoing, err = nats.Connect(n.url)
	if nil != err {
		log.Error().Err(err).Str("dir", "outgoing").Msg("Unable to connect to NATS server")
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
			log.Error().Err(err).Str("dir", "incoming").Msg("failed to drain connection")
		}
	}
	n.outgoing.Close()
}

func (n *Server) Subscribe(subject string) (chan *nats.Msg, error) {
	ch := make(chan *nats.Msg, 512)
	sub, err := n.incoming.ChanSubscribe(subject, ch)
	if nil != err {
		return nil, err
	}
	if err = sub.SetPendingLimits(512, 5*1024*1024); nil != err {
		return nil, err
	}
	return ch, nil
}

func (n *Server) HealthCheckName() string {
	return "Nats"
}

func (n *Server) HealthCheck() bool {
	return n.incoming.IsConnected() && n.outgoing.IsConnected()
}
