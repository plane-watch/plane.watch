package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"plane.watch/lib/nats_io"
	"sort"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"plane.watch/lib/dedupe/forgetfulmap"

	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"plane.watch/lib/export"
	"plane.watch/lib/tile_grid"
	"plane.watch/lib/ws_protocol"
)

type (
	PwWsBrokerWebJSON struct {
		natsRPC   *nats_io.Server
		Addr      string
		ServeTest bool

		serveMux      *http.ServeMux
		httpServer    http.Server
		cert, certKey string

		domainsToServe []string

		clients   *ClientListJSON
		listening bool

		sendTickDuration time.Duration
	}

	loadedResponseJSON struct {
		out ws_protocol.WsResponse

		highLow, tile string
	}

	WsClientJSON struct {
		conn    *websocket.Conn
		outChan chan loadedResponseJSON
		cmdChan chan WsCmdJSON

		parent     *ClientListJSON
		identifier string
		log        zerolog.Logger

		sendTickDuration time.Duration
	}
	WsCmdJSON struct {
		action     string
		what       string
		extra      string
		tick       time.Duration
		locHistory []ws_protocol.LocationHistory
		results    ws_protocol.SearchResult
	}
	ClientListJSON struct {
		clients sync.Map

		// naive approach
		globalList *forgetfulmap.ForgetfulSyncMap

		broker *PwWsBrokerWebJSON
	}
)

// configureWeb Sets up our serve mux to handle our web endpoints
func (bw *PwWsBrokerWebJSON) configureWeb() error {
	bw.clients = newClientListJSON(bw)

	if err := configureServeMuxCommon(bw.serveMux, bw.certKey, bw.cert, bw.ServeTest); err != nil {
		return err
	}
	bw.serveMux.HandleFunc("/planes", bw.servePlanes)

	log.Info().Str("On Address", bw.Addr).Msg("Starting Serving")

	bw.httpServer = http.Server{
		Addr:         bw.Addr,
		Handler:      bw.serveMux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	return nil
}

// listenAndServe runs the top level serving and channel control
func (bw *PwWsBrokerWebJSON) listenAndServe(exitChan chan bool) {
	log.Info().Str("HttpAddr", bw.Addr).Msg("HTTP Listening on")
	bw.listening = true
	var err error
	isTLS := false
	if bw.cert != "" {
		isTLS = true
		err = bw.httpServer.ListenAndServeTLS(bw.cert, bw.certKey)
	} else {
		err = bw.httpServer.ListenAndServe()
	}
	if nil != err {
		bw.listening = false
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error().
				Err(err).
				Bool("tls", isTLS).
				Str("cert", bw.cert).
				Str("cert-key", bw.certKey).
				Msg("web server error")
		}
	}
	exitChan <- true
}

// logRequest logs all the web requests we get
func (bw *PwWsBrokerWebJSON) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("Remote", r.RemoteAddr).Str("Request", r.RequestURI).Msg("Web RQ")
		handler.ServeHTTP(w, r)
	})
}

// ServeHTTP Asks our internal serveMux to Serve HTTP web requests
func (bw *PwWsBrokerWebJSON) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bw.serveMux.ServeHTTP(w, r)
}

// closeWeb calls Close() on our http server
func (bw *PwWsBrokerWebJSON) closeWeb() error {
	return bw.httpServer.Close()
}

// SendLocationUpdate calls SendLocationUpdate() on our internal clientList
func (bw *PwWsBrokerWebJSON) SendLocationUpdate(highLow string, msg []byte) {
	bw.clients.SendLocationUpdate(highLow, msg)
}

// search handles
func (cl *ClientListJSON) performSearch(query string) ws_protocol.SearchResult {
	results := ws_protocol.SearchResult{
		Query:    query,
		Aircraft: []*export.PlaneLocationJSON{},
		Airport:  []ws_protocol.AirportLocation{},
		Route:    []string{},
	}

	if len(query) >= 2 {
		// find any aircraft
		cl.globalList.Range(func(key, value interface{}) bool {
			loc := value.(*export.PlaneLocationJSON)

			if strings.Contains(strings.ToLower(loc.Icao), query) ||
				(nil != loc.CallSign && strings.Contains(strings.ToLower(*loc.CallSign), query)) ||
				(nil != loc.RegisteredOwner && strings.Contains(strings.ToLower(*loc.RegisteredOwner), query)) ||
				(nil != loc.Registration && strings.Contains(strings.ToLower(*loc.Registration), query)) {
				results.Aircraft = append(results.Aircraft, loc)
			}

			return len(results.Aircraft) < 10
		})

		sort.Sort(results.Aircraft)
	}

	// Airport Lookup, if we have a nats connection
	if nil != cl.broker.natsRPC {
		headers := map[string]string{}
		resp, err := cl.broker.natsRPC.Request(export.NatsAPISearchAirportV1, []byte(query), headers, time.Second)
		if nil != err {
			log.Error().Err(err).Msg("Failed to search for airport")
		} else {
			json := jsoniter.ConfigFastest
			var airports []export.Airport
			err = json.Unmarshal(resp, &airports)
			if nil != err {
				log.Error().Err(err).Msg("Failed to unmarshal airport search results")
			} else {
				for _, airport := range airports {
					results.Airport = append(results.Airport, ws_protocol.AirportLocation{
						Name: airport.Name,
						Icao: airport.IcaoCode,
						Iata: airport.IataCode,
						Lat:  airport.Latitude,
						Lon:  airport.Longitude,
					})
				}
			}
		}
	}

	return results
}

// servePlanes Serves our Websocket Endpoint
func (bw *PwWsBrokerWebJSON) servePlanes(w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("New Connection", r.RemoteAddr).Msg("New /planes WS")

	compress := r.URL.Query().Get("compress")
	wsCompression := websocket.CompressionContextTakeover

	if strings.ToLower(compress) == "false" {
		wsCompression = websocket.CompressionDisabled
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       []string{ws_protocol.WsProtocolPlanes},
		InsecureSkipVerify: false,
		OriginPatterns:     bw.domainsToServe,
		CompressionMode:    wsCompression,
	})
	if nil != err {
		log.Error().Err(err).Msg("Failed to setup websocket connection")
		w.WriteHeader(500)
		_, _ = w.Write([]byte("Failed to setup websocket connection"))
		return
	}

	log.Debug().Str("protocol", conn.Subprotocol()).Msg("Speaking...")
	switch conn.Subprotocol() {
	case ws_protocol.WsProtocolPlanes:
		client := NewWsClientJSON(conn, r.RemoteAddr, bw.sendTickDuration)
		bw.clients.addClient(client)
		client.Handle(r.Context())
		bw.clients.removeClient(client)
	default:
		_ = conn.Close(websocket.StatusPolicyViolation, "Unknown Sub Protocol")
		log.Debug().Str("proto", conn.Subprotocol()).Msg("Bad connection, could not speak protocol")
		return
	}
}

// HealthCheck allows us to check the health of this component
func (bw *PwWsBrokerWebJSON) HealthCheck() bool {
	log.Info().Bool("Web Listening", bw.listening).Msg("Health check")
	return bw.listening
}

// HealthCheckName Gives the name of this health check
func (bw *PwWsBrokerWebJSON) HealthCheckName() string {
	return "WS Broker Web"
}

// NewWsClientJSON creates a new Websocket Client. This represents an individual connection and its handling
func NewWsClientJSON(conn *websocket.Conn, identifier string, defaultSendTick time.Duration) *WsClientJSON {
	client := WsClientJSON{
		conn:             conn,
		cmdChan:          make(chan WsCmdJSON),
		outChan:          make(chan loadedResponseJSON, 500),
		identifier:       identifier,
		log:              log.With().Str("client", identifier).Logger(),
		sendTickDuration: defaultSendTick,
	}
	return &client
}

// Handle is a top level method that is called to Handle a websocket client connection
func (c *WsClientJSON) Handle(ctx context.Context) {
	err := c.planeProtocolHandler(ctx, c.conn)
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if nil != err {
		if websocket.CloseStatus(err) == -1 {
			log.Error().Err(err).Msg("Failure in protocol handler")
		}
		return
	}
}

// AddSub adds a "Please Subscribe this client to this tile" command to the clients command queue
func (c *WsClientJSON) AddSub(tileName string) {
	log.Debug().Msg("Add Sub")
	c.cmdChan <- WsCmdJSON{
		action: ws_protocol.RequestTypeSubscribe,
		what:   tileName,
	}
	log.Debug().Msg("Add Sub Done")
}

// UnSub adds a "Please remove this tile from the clients list" command to the clients command queue
func (c *WsClientJSON) UnSub(tileName string) {
	log.Debug().Msg("Unsub")
	c.cmdChan <- WsCmdJSON{
		action: ws_protocol.RequestTypeUnsubscribe,
		what:   tileName,
	}
	log.Debug().Msg("Unsub done")
}

// SendSubscribedTiles adds a "Please send the list of tiles that we are currently subscribed to" command to the queue
func (c *WsClientJSON) SendSubscribedTiles() {
	log.Debug().Msg("Unsub")
	c.cmdChan <- WsCmdJSON{
		action: ws_protocol.RequestTypeSubscribeList,
		what:   "",
	}
	log.Debug().Msg("Unsub done")
}

// SendTilePlanes adds a "send to the client the list of planes on the requested tile" command to the queue
func (c *WsClientJSON) SendTilePlanes(tileName string) {
	c.log.Debug().Str("tile", tileName).Msg("Planes on Tile")
	c.cmdChan <- WsCmdJSON{
		action: ws_protocol.RequestTypeGridPlanes,
		what:   tileName,
	}
}

// SendPlaneLocationHistory adds a "send the location history (from clickhouse) of the requested flight" command to the queue
func (c *WsClientJSON) SendPlaneLocationHistory(icao, callSign string) {
	c.log.Debug().Str("icao", icao).Str("callSign", callSign).Msg("Request Flight Path")
	go func() {
		c.cmdChan <- WsCmdJSON{
			action:     ws_protocol.RequestTypePlaneLocHistory,
			what:       icao,
			extra:      callSign,
			locHistory: GlobalClickHouseData.PlaneLocationHistory(icao, callSign),
		}
	}()
}

func (c *WsClientJSON) AdjustSendTick(tick int) {
	if tick > 0 {
		tickDuration := time.Duration(tick) * time.Millisecond
		if tickDuration < MaxTickDuration {
			c.log.Debug().Dur("Requested Tick (ms)", tickDuration).Msg("Client Adjust Tick Timing")
			go func() {
				c.cmdChan <- WsCmdJSON{
					action: ws_protocol.RequestTypeTickAdjust,
					tick:   tickDuration,
				}
			}()
		}
	}
}
func (c *WsClientJSON) SendSearchResults(query string) {
	go func() {
		c.log.Debug().Str("Searching for", query).Msg("Performing Search")
		c.cmdChan <- WsCmdJSON{
			action:  ws_protocol.RequestTypeSearch,
			what:    query,
			results: c.parent.performSearch(strings.ToLower(query)),
		}
	}()
}

// planeProtocolHandler handles our websocket protocol
// it runs the command queue that sends information to the client
// it runs (in a go routine) the reading of requests from the client.
// the read requests become commands in the command queue
func (c *WsClientJSON) planeProtocolHandler(ctx context.Context, conn *websocket.Conn) error {
	// read from the connection for commands to perform
	// these get added to the queue to process
	json := jsoniter.ConfigFastest
	go func() {
		for {
			mt, frame, err := conn.Read(ctx)
			if nil != err {
				if !(errors.Is(err, io.EOF) || websocket.CloseStatus(err) >= 0) {
					log.Debug().Err(err).Int("Close Status", int(websocket.CloseStatus(err))).Msg("Error from reading")
				}
				c.cmdChan <- WsCmdJSON{action: "exit"}
				return
			}
			switch mt {
			case websocket.MessageText:
				log.Debug().Bytes("Client Msg", frame).Msg("From Client")
				rq := ws_protocol.WsRequest{}
				if err = json.Unmarshal(frame, &rq); nil != err {
					log.Warn().Err(err).Msg("Failed to understand message from client")
				}
				switch rq.Type {
				case ws_protocol.RequestTypeSubscribe:
					c.AddSub(rq.GridTile)
				case ws_protocol.RequestTypeSubscribeList:
					c.SendSubscribedTiles()
				case ws_protocol.RequestTypeUnsubscribe:
					c.UnSub(rq.GridTile)
				case ws_protocol.RequestTypeGridPlanes:
					c.SendTilePlanes(rq.GridTile)
				case ws_protocol.RequestTypePlaneLocHistory:
					c.SendPlaneLocationHistory(rq.Icao, rq.CallSign)
				case ws_protocol.RequestTypeTickAdjust:
					c.AdjustSendTick(rq.Tick)
				case ws_protocol.RequestTypeSearch:
					c.SendSearchResults(rq.Query)
				default:
					_ = c.sendError(ctx, "Unknown request type")
				}

			case websocket.MessageBinary:
				_ = c.sendError(ctx, "Please speak text")
			}
		}
	}()

	// write a stream of location information
	subs := make(map[string]bool)

	grid := make(map[string]bool)
	gridNames := make(map[string]bool)
	gridNames[""] = true
	grid["all_low"] = true
	grid["all_high"] = true
	for k := range tile_grid.GetGrid() {
		gridNames[k] = true
		grid[k+"_low"] = true
		grid[k+"_high"] = true
	}

	locationMessages := make([]*export.PlaneLocationJSON, 0, 1000)
	icaoIdLookup := make(map[string]int, 1000)

	currentSendTickDuration := c.sendTickDuration
	if currentSendTickDuration == 0 {
		// it still runs, but we do not send anything
		currentSendTickDuration = 10 * time.Second // something long enough that it is not much of an overhead
	}
	sendTick := time.NewTicker(currentSendTickDuration)
	defer sendTick.Stop()

	// this is the command processing main loop
	var err error
	for {
		select {
		case cmdMsg := <-c.cmdChan:
			switch cmdMsg.action {
			case "exit":
				return nil
			case ws_protocol.RequestTypeSubscribe:
				if _, ok := grid[cmdMsg.what]; ok {
					subs[cmdMsg.what] = true
					err = c.sendAck(ctx, ws_protocol.ResponseTypeAckSub, cmdMsg.what)
					prometheusSubscriptions.WithLabelValues(cmdMsg.what).Inc()
				} else {
					err = c.sendError(ctx, "Unknown Tile: "+cmdMsg.what)
				}
			case ws_protocol.RequestTypeUnsubscribe:
				if _, ok := subs[cmdMsg.what]; ok {
					prometheusSubscriptions.WithLabelValues(cmdMsg.what).Dec()
					err = c.sendAck(ctx, ws_protocol.ResponseTypeAckUnsub, cmdMsg.what)
				} else {
					err = c.sendError(ctx, "Not Subbed to: "+cmdMsg.what)
				}
				delete(subs, cmdMsg.what)
			case ws_protocol.RequestTypeSubscribeList:
				tiles := make([]string, 0, len(subs))
				for k, v := range subs {
					if v {
						tiles = append(tiles, k)
					}
				}
				err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
					Type:  ws_protocol.ResponseTypeSubTiles,
					Tiles: tiles,
				})
			case ws_protocol.RequestTypeGridPlanes:
				if _, gridOk := gridNames[cmdMsg.what]; gridOk {
					matching := 0
					// find all things currently in requested grid
					c.parent.globalList.Range(func(key, value interface{}) bool {
						loc := value.(*export.PlaneLocationJSON)
						if cmdMsg.what == loc.TileLocation {
							if id, ok := icaoIdLookup[loc.Icao]; ok {
								locationMessages[id] = loc
							} else {
								locationMessages = append(locationMessages, loc)
								icaoIdLookup[loc.Icao] = len(locationMessages) - 1
							}
							matching++
						}
						return true
					})
				} else {
					err = c.sendError(ctx, "Unknown Tile: "+cmdMsg.what)
				}
			case ws_protocol.RequestTypePlaneLocHistory:
				err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
					Type:     ws_protocol.ResponseTypePlaneLocHistory,
					History:  cmdMsg.locHistory,
					Icao:     cmdMsg.what,
					CallSign: cmdMsg.extra,
				})
			case ws_protocol.RequestTypeTickAdjust:
				// this is already less than 10 seconds (MaxTickDuration)
				if cmdMsg.tick > c.sendTickDuration {
					currentSendTickDuration = cmdMsg.tick
				} else {
					// set the smallest allowed
					currentSendTickDuration = c.sendTickDuration
				}
				c.log.Info().Dur("Tick", currentSendTickDuration).Msg("Changing Tick Rate")
				sendTick.Reset(currentSendTickDuration)

				err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
					Type:    ws_protocol.ResponseTypeMsg,
					Message: fmt.Sprintf("Set Tick Rate To %s", currentSendTickDuration),
				})
			case ws_protocol.RequestTypeSearch:
				err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
					Type:    ws_protocol.ResponseTypeSearchResults,
					Results: &cmdMsg.results,
				})
			default:
				err = c.sendError(ctx, "Unknown Command")
			}
		case planeMsg := <-c.outChan:
			// if we have a subscription to this planes tile or all tiles
			// log.Debug().Str("tile", planeMsg.tile).Str("highlow", planeMsg.highLow).Msg("info")
			tileSub, tileOk := subs[planeMsg.tile]
			allSub, allOk := subs["all"+planeMsg.highLow]
			if (tileSub && tileOk) || (allSub && allOk) {
				if c.sendTickDuration > 0 {
					// limit our updates to only 1 per icao, sent periodically
					if id, ok := icaoIdLookup[planeMsg.out.Location.Icao]; ok {
						locationMessages[id] = planeMsg.out.Location
					} else {
						locationMessages = append(locationMessages, planeMsg.out.Location)
						icaoIdLookup[planeMsg.out.Location.Icao] = len(locationMessages) - 1
					}
				} else {
					err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
						Type:     ws_protocol.ResponseTypePlaneLocation,
						Location: planeMsg.out.Location,
					})
				}
			}
		case <-sendTick.C:
			if len(locationMessages) > 0 {
				err = c.sendPlaneMessage(ctx, &ws_protocol.WsResponse{
					Type:      ws_protocol.ResponseTypePlaneLocations,
					Locations: locationMessages,
				})
				// reset the slice and lookup map
				locationMessages = make([]*export.PlaneLocationJSON, 0, 1000)
				icaoIdLookup = make(map[string]int, 1000)
			}
		}

		if nil != err {
			break
		}
	}

	// tell prometheus we are no longer caring about the tiles
	for k := range subs {
		prometheusSubscriptions.WithLabelValues(k).Dec()
	}
	return err
}

// sendAck sends an acknowledgement message to the client
func (c *WsClientJSON) sendAck(ctx context.Context, ackType, tile string) error {
	rs := ws_protocol.WsResponse{
		Type:  ackType,
		Tiles: []string{tile},
	}
	return c.sendPlaneMessage(ctx, &rs)
}

// sendError sends an error message to the client
func (c *WsClientJSON) sendError(ctx context.Context, msg string) error {
	c.log.Error().Str("protocol", "error").Msg(msg)
	rs := ws_protocol.WsResponse{
		Type:    ws_protocol.ResponseTypeError,
		Message: msg,
	}
	return c.sendPlaneMessage(ctx, &rs)
}

// sendPlaneMessage sends a message to the client, with timeout
func (c *WsClientJSON) sendPlaneMessage(ctx context.Context, planeMsg *ws_protocol.WsResponse) error {
	json := jsoniter.ConfigFastest
	buf, err := json.Marshal(planeMsg)
	if nil != err {
		c.log.Debug().Err(err).Str("type", planeMsg.Type).Msg("Failed to marshal plane msg to send to client")
		return err
	}
	go func() {
		if err = c.writeTimeout(ctx, 3*time.Second, buf); nil != err {
			c.log.Debug().
				Err(err).
				Str("type", planeMsg.Type).
				Msgf("Failed to send message to client. %+v", err)
		}
	}()
	return nil
}

// writeTimeout handles the writing of a message to the actual websocket connection
func (c *WsClientJSON) writeTimeout(ctx context.Context, timeout time.Duration, msg []byte) error {
	ctxW, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prometheusMessagesSent.Inc()
	prometheusMessagesSize.Add(float64(len(msg)))
	return c.conn.Write(ctxW, websocket.MessageText, msg)
}

// newClientListJSON represents a list of websocket clients that we are currently servicing
func newClientListJSON(bw *PwWsBrokerWebJSON) *ClientListJSON {
	cl := ClientListJSON{
		broker: bw,
	}
	cl.globalList = forgetfulmap.NewForgetfulSyncMap(
		forgetfulmap.WithPrometheusCounters(prometheusKnownPlanes),
		forgetfulmap.WithPreEvictionAction(func(key, value any) {
			log.Debug().Str("ICAO", key.(string)).Msg("Removing Aircraft due to inactivity")
		}),
		forgetfulmap.WithForgettableAction(func(key, value any, added time.Time) bool {
			result := true
			if loc, ok := value.(*export.PlaneLocationJSON); ok {
				oldest := time.Now().Add(-10 * time.Minute)
				if loc.OnGround {
					// if a plane is on the ground, remove it 2 minutes after we last saw it
					oldest = time.Now().Add(-2 * time.Minute)
				}
				// TODO: determine if plane is ADS-C more robustly
				if loc.SourceTag == "ADS-C" {
					oldest = time.Now().Add(-time.Hour)
				}
				// remove  the plane from the list if it is older than our oldest allowable
				result = loc.LastMsg.Before(oldest)
			}
			// remove anything that is not a *export.PlaneLocationJSON
			return result
		}),
	)

	return &cl
}

// addClient adds a websocket client to the list
func (cl *ClientListJSON) addClient(c *WsClientJSON) {
	c.log.Debug().Msg("Add Client")
	c.parent = cl
	cl.clients.Store(c, true)
	prometheusNumClients.Inc()
	// c.log.Debug().Msg("Add Client Done")
}

// removeClient removes a websocket client from the list
func (cl *ClientListJSON) removeClient(c *WsClientJSON) {
	c.log.Debug().Msg("Remove Client")
	close(c.outChan)
	cl.clients.Delete(c)
	prometheusNumClients.Dec()
	// log.Debug().Msg("Remove Client Done")
}

func (cl *ClientListJSON) globalListUpdate(loc *export.PlaneLocationJSON) {
	if nil == loc {
		return
	}
	cl.globalList.Store(loc.Icao, loc)
}

// SendLocationUpdate sends an update to each listening client
// todo: make this threaded?
func (cl *ClientListJSON) SendLocationUpdate(highLow string, rawMsg []byte) {

	// TODO - we only have a single worker with this rewrite, might need tuning
	var loc *export.PlaneLocationJSON
	var json = jsoniter.ConfigFastest
	errJson := json.Unmarshal(rawMsg, &loc)
	if nil != errJson {
		log.Debug().Err(errJson).Msg("did not understand msg")
		return
	}
	tile := loc.TileLocation + highLow

	// Add our update to our global list
	cl.globalListUpdate(loc)

	// send the update to each of our clients
	cl.clients.Range(func(key, value interface{}) bool {
		defer func() {
			if r := recover(); nil != r {
				log.Error().Msgf("Panic: %v", r)
			}
		}()
		client := key.(*WsClientJSON)
		client.outChan <- loadedResponseJSON{
			out: ws_protocol.WsResponse{
				Type:     ws_protocol.ResponseTypePlaneLocation,
				Location: loc,
			},
			highLow: highLow,
			tile:    tile,
		}
		return true
	})
}
