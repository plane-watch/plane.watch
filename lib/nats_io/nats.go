package nats_io

import (
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/url"
	"time"
)

const DefaultQueueDepth = 2048

type (
	healthItem struct {
		name, subject string
		ch            chan *nats.Msg
	}

	Server struct {
		url            string
		connectionName string
		QueueDepth     int

		incoming *nats.Conn
		outgoing *nats.Conn

		channels []healthItem

		log zerolog.Logger

		droppedMessageCounter prometheus.Counter
	}
)

func NewServer(serverUrl, connectionName string) (*Server, error) {
	n := &Server{
		log:            log.With().Str("section", "nats.io").Logger(),
		QueueDepth:     DefaultQueueDepth,
		connectionName: connectionName,
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
func (n *Server) DroppedCounter(counter prometheus.Counter) {
	n.droppedMessageCounter = counter
}
func (n *Server) NatsErrHandler(conn *nats.Conn, sub *nats.Subscription, err error) {
	l := n.log.Error().Err(err)

	for _, c := range n.channels {
		if c.subject == sub.Subject {
			l.Int(c.name+" len", len(c.ch)).
				Int(c.name+" capacity", cap(c.ch))
		}
	}
	if nil != conn {
		l.Str("addr", conn.ConnectedUrl())
	}
	if nil != sub {
		l.Str("subscription", sub.Subject+"["+sub.Queue+"]")
	}
	l.Send()

	if nil != n.droppedMessageCounter && err == nats.ErrSlowConsumer {
		n.droppedMessageCounter.Inc()
	}
}

func (n *Server) Connect() error {
	var err error
	n.log.Debug().Str("url", n.url).Msg("connecting to server...")
	n.incoming, err = nats.Connect(
		n.url,
		nats.ErrorHandler(n.NatsErrHandler),
		nats.Name(n.connectionName+"+incoming"),
	)
	if nil != err {
		n.log.Error().Err(err).Str("dir", "incoming").Msg("Unable to connect to NATS server")
		return err
	}
	n.outgoing, err = nats.Connect(
		n.url,
		nats.ErrorHandler(n.NatsErrHandler),
		nats.Name(n.connectionName+"+outgoing"),
	)
	if nil != err {
		n.log.Error().Err(err).Str("dir", "outgoing").Msg("Unable to connect to NATS server")
		return err
	}
	n.log.Debug().Str("url", n.url).Msg("Connected")
	return nil
}

// Publish is our simple message publisher
func (n *Server) Publish(queue string, msg []byte) error {
	err := n.outgoing.Publish(queue, msg)
	if nil != err {
		if nats.ErrInvalidConnection == err || nats.ErrConnectionClosed == err || nats.ErrConnectionDraining == err {
			n.log.Error().Err(err).Msg("Connection not in a valid state")
		}
	}
	return err
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
	ch := make(chan *nats.Msg, n.QueueDepth)
	n.channels = append(n.channels, healthItem{
		name:    "sub-" + subject,
		subject: subject,
		ch:      ch,
	})
	_, err := n.incoming.ChanSubscribe(subject, ch)
	if nil != err {
		return nil, err
	}
	n.log.Info().Str("subject", subject).Msg("subscribed")
	return ch, nil
}

// SubscribeQueueGroup allows many workers to feed of a single queue
func (n *Server) SubscribeQueueGroup(subject, queueGroup string) (chan *nats.Msg, error) {
	ch := make(chan *nats.Msg, n.QueueDepth)
	n.channels = append(n.channels, healthItem{
		name:    "sub-queue-" + subject,
		subject: subject,
		ch:      ch,
	})
	_, err := n.incoming.ChanQueueSubscribe(subject, queueGroup, ch)
	if nil != err {
		return nil, err
	}
	n.log.Info().Str("subject", subject).Str("queue-group", queueGroup).Msg("subscribed")
	return ch, nil
}

func (n *Server) Request(subject string, data []byte, timeout time.Duration) ([]byte, error) {
	msg, err := n.outgoing.Request(subject, data, timeout)
	if nil != err {
		n.log.Error().Err(err).Str("subject", subject).Msg("Failed to request")
		return nil, err
	}
	// TODO: instrument so we know how long replies take and how many are successfully served
	return msg.Data, nil
}

func (n *Server) SubscribeReply(subject, queue string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return n.outgoing.QueueSubscribe(subject, queue, handler)
}

func (n *Server) HealthCheckName() string {
	return "Nats"
}

func (n *Server) HealthCheck() bool {
	n.log.Info().
		Int("Num Channels", len(n.channels)).
		Bool("Incoming Connected", n.incoming.IsConnected()).
		Bool("Outgoing Connected", n.outgoing.IsConnected()).
		Send()
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
