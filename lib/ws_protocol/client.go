package ws_protocol

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
	"net/http"
	"nhooyr.io/websocket"
	"plane.watch/lib/setup"
	"time"
)

const (
	ProdSourceURL = "https://plane.watch/planes"
)

type (
	WsClient struct {
		sourceURL string
		conn      *websocket.Conn

		wireProtocol string

		jsonResponseHandlers     []JSONResponseHandler
		protobufResponseHandlers []ProtobufResponseHandler

		logger zerolog.Logger

		insecure bool
	}

	Option                  func(client *WsClient)
	JSONResponseHandler     func(response *WsResponse)
	ProtobufResponseHandler func(response *WebSocketResponse)
)

var (
	ErrUnexpectedProtocol = errors.New("Unexpected WS Protocol")
)

func WithSourceURL(sourceURL string) Option {
	return func(client *WsClient) {
		client.sourceURL = sourceURL
	}
}

func WithJSONResponseHandler(f JSONResponseHandler) Option {
	return func(client *WsClient) {
		client.jsonResponseHandlers = append(client.jsonResponseHandlers, f)
	}
}

func WithProtobufResponseHandler(f ProtobufResponseHandler) Option {
	return func(client *WsClient) {
		client.protobufResponseHandlers = append(client.protobufResponseHandlers, f)
	}
}

func WithWireProtocol(wireProtocol string) Option {
	return func(client *WsClient) {
		client.wireProtocol = wireProtocol
	}
}

func WithLogger(logger zerolog.Logger) Option {
	return func(client *WsClient) {
		client.logger = logger
	}
}

func WithInsecureConnection(insecue bool) Option {
	return func(client *WsClient) {
		client.insecure = insecue
	}
}

func NewClient(opts ...Option) *WsClient {
	c := &WsClient{
		sourceURL:            ProdSourceURL,
		jsonResponseHandlers: make([]JSONResponseHandler, 0),
		insecure:             false,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *WsClient) Connect() error {
	var err error
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.insecure}
	httpClient := &http.Client{Transport: customTransport}
	cfg := &websocket.DialOptions{
		HTTPClient:   httpClient,
		Subprotocols: []string{WsProtocolPlanes},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c.conn, _, err = websocket.Dial(ctx, c.sourceURL, cfg)

	if nil != err {
		return errors.Wrap(err, "failed to connect to Plane.Watch websocket")
	}

	if c.conn.Subprotocol() != WsProtocolPlanes {
		return ErrUnexpectedProtocol
	}

	c.conn.SetReadLimit(1_048_576) // 1MiB

	go c.reader()
	return nil
}

func (c *WsClient) Disconnect() error {
	return c.conn.Close(websocket.StatusNormalClosure, "Going away")
}

func (c *WsClient) reader() {
	for {
		mType, msg, err := c.conn.Read(context.Background())
		if mType != websocket.MessageText {
			// c.logger.Error().Str("body", string(msg)).Msg("Incorrect Protocol, expected text, got binary")
			continue
		}
		if nil != err {
			if errors.Is(err, websocket.CloseError{}) {
				return
			}
			c.logger.Error().Err(err).Send()
			continue
		}
		switch c.wireProtocol {
		case setup.WireProtocolJSON:
			r := &WsResponse{}
			err = json.Unmarshal(msg, r)
			if nil != err {
				c.logger.Error().Err(err).Send()
				continue
			}

			for _, f := range c.jsonResponseHandlers {
				f(r)
			}
		case setup.WireProtocolProtobuf:
			r := &WebSocketResponse{}
			err = proto.Unmarshal(msg, r)
			if err != nil {
				c.logger.Error().Err(err).Send()
				continue
			}
			if r == nil {
				c.logger.Error().Msg("Failed to decode protobuf message")
				continue
			}

			for _, f := range c.protobufResponseHandlers {
				f(r)
			}
		}
	}
}

func (c *WsClient) writeRequest(rq *WsRequest) error {
	rqJSON, err := json.Marshal(rq)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to tile")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.conn.Write(ctx, websocket.MessageText, rqJSON)
}

func (c *WsClient) Subscribe(gridTile string) error {
	rq := WsRequest{
		Type:     RequestTypeSubscribe,
		GridTile: gridTile,
	}

	return c.writeRequest(&rq)
}

func (c *WsClient) SubscribeAllLow() error {
	rq := WsRequest{
		Type:     RequestTypeSubscribe,
		GridTile: GridTileAllLow,
	}

	return c.writeRequest(&rq)
}

func (c *WsClient) SubscribeAllHigh() error {
	rq := WsRequest{
		Type:     RequestTypeSubscribe,
		GridTile: GridTileAllHigh,
	}

	return c.writeRequest(&rq)
}
