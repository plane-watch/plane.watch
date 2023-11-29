package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"plane.watch/lib/export"
	"plane.watch/lib/setup"
	"plane.watch/lib/tracker/beast"
	"plane.watch/lib/tracker/mode_s"
	"plane.watch/lib/tracker/sbs1"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	planesSource int

	sourceInfo struct {
		mu         sync.Mutex
		frameCount uint64
		planes     map[uint32]*export.PlaneAndLocationInfoMsg
		frames     map[uint32]uint64
		icaos      []uint32
	}

	model struct {
		logger zerolog.Logger

		startTime time.Time

		tapper *PlaneWatchTapper

		focusIcaoList []string

		help help.Model

		statsTable       table.Model
		selectedTable    table.Model
		planesTable      table.Model
		source           planesSource
		selectedIcao     uint32
		selectedCallSign string

		logView      viewport.Model
		logViewReady bool

		width, height int
		tickCount     uint64
		tickDuration  time.Duration

		heading lipgloss.Style

		logs *strings.Builder

		incomingMutex sync.Mutex

		feederSources      map[string]int
		incomingIcaoFrames map[uint32]int
		numIncomingBeast   uint64
		numIncomingAvr     uint64
		numIncomingSbs1    uint64

		afterIngest     sourceInfo
		afterEnrichment sourceInfo
		afterRouterLow  sourceInfo
		afterRouterHigh sourceInfo
		finalLow        sourceInfo
		finalHigh       sourceInfo
	}

	timerTick time.Time
)

func initialModel(natsURL, wsURL string, c *cli.Context) (*model, error) {
	logs := &strings.Builder{}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: logs, TimeFormat: time.UnixDate}).With().Timestamp().Logger()

	m := &model{
		logger:    logger.With().Str("app", "model").Logger(),
		startTime: time.Now(),
		tapper: NewPlaneWatchTapper(
			WithLogger(logger),
			WithProtocol(c.String(setup.WireProtocol)),
			WithProtocolFor(wireProtocolForIngest, c.String(wireProtocolForIngest)),
			WithProtocolFor(wireProtocolForEnrichment, c.String(wireProtocolForEnrichment)),
			WithProtocolFor(wireProtocolForRouter, c.String(wireProtocolForRouter)),
			WithProtocolFor(wireProtocolForWsBroker, c.String(wireProtocolForWsBroker)),
		),
		help:          help.New(),
		tickDuration:  time.Millisecond * 16,
		focusIcaoList: make([]string, 0),
		source:        planesSourceWSLow,
		logs:          logs,

		feederSources:      make(map[string]int),
		incomingIcaoFrames: make(map[uint32]int),
	}
	if err := m.tapper.Connect(natsURL, wsURL); err != nil {
		return nil, err
	}
	m.buildTables()
	m.configureStyles()

	m.afterIngest.init()
	m.afterEnrichment.init()
	m.afterRouterLow.init()
	m.afterRouterHigh.init()
	m.finalLow.init()
	m.finalHigh.init()
	m.logger.Debug().Msg("Startup Init Complete")

	return m, nil
}

func (m *model) configureStyles() {
	m.heading = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.statsTable.SetStyles(s)
	m.selectedTable.SetStyles(s)
	m.planesTable.SetStyles(s)
}

func (m *model) buildTables() {
	m.statsTable = table.New(
		table.WithColumns([]table.Column{
			{Title: "Receivers", Width: 11},
			{Title: "Planes", Width: 10},

			{Title: "Beast", Width: 10},
			{Title: "Avr", Width: 5},
			{Title: "Sbs1", Width: 5},

			{Title: "Frames", Width: 10},

			{Title: "Enriched", Width: 10},

			{Title: "Routed Low", Width: 10},
			{Title: "Routed High", Width: 11},

			{Title: "WS Low", Width: 10},
			{Title: "WS High", Width: 10},
		}),
		table.WithRows([]table.Row{
			{"0", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
		}),
		table.WithHeight(2),
		table.WithFocused(false),
	)
	m.selectedTable = table.New(
		table.WithColumns([]table.Column{
			{Title: "Source", Width: 20},
			{Title: "# Updates", Width: 10},
			{Title: "Squawk", Width: 9},
			{Title: "Lat", Width: 10},
			{Title: "Lon", Width: 10},
			{Title: "Altitude", Width: 10},
			{Title: "Vert Rate", Width: 10},
			{Title: "Heading", Width: 10},
		}),
		table.WithRows(m.defaultSelectedTableRows()),
		table.WithHeight(8),
		table.WithFocused(false),
	)
	m.planesTable = table.New(
		table.WithColumns([]table.Column{
			{Title: "ICAO", Width: 6},
			{Title: "Receivers", Width: 9},
			{Title: "CallSign", Width: 9},
			{Title: "Squawk", Width: 9},
			{Title: "Lat", Width: 10},
			{Title: "Lon", Width: 10},
			{Title: "Altitude", Width: 10},
			{Title: "Vert Rate", Width: 10},
			{Title: "Heading", Width: 10},
		}),
		table.WithHeight(10),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
	)
}

func (m *model) defaultSelectedTableRows() []table.Row {
	return []table.Row{
		{planesSourceIncoming.String(), "", "", "", "", "", "", ""},
		{planesSourceIngest.String(), "", "", "", "", "", "", ""},
		{planesSourceEnriched.String(), "", "", "", "", "", "", ""},
		{planesSourceRoutedLow.String(), "", "", "", "", "", "", ""},
		{planesSourceRoutedHigh.String(), "", "", "", "", "", "", ""},
		{planesSourceWSLow.String(), "", "", "", "", "", "", ""},
		{planesSourceWSHigh.String(), "", "", "", "", "", "", ""},
	}
}

func (m *model) handleIncomingData(frameType, tag string, data []byte) {
	m.incomingMutex.Lock()
	defer m.incomingMutex.Unlock()

	m.feederSources[tag]++

	switch frameType {
	case "beast":
		frame, err := beast.NewFrame(data, false)
		if nil != err {
			return
		}
		m.incomingIcaoFrames[frame.Icao()]++
		m.numIncomingBeast++
	case "avr":
		frame, err := mode_s.DecodeString(string(data), m.startTime)
		if nil == err {
			return
		}
		m.incomingIcaoFrames[frame.Icao()]++
		m.numIncomingAvr++
	case "sbs1":
		frame := sbs1.NewFrame(string(data))
		m.incomingIcaoFrames[frame.Icao()]++
		m.numIncomingSbs1++
	}
}

func (si *sourceInfo) init() {
	si.planes = make(map[uint32]*export.PlaneAndLocationInfoMsg)
	si.frames = make(map[uint32]uint64)
}

func (si *sourceInfo) update(loc *export.PlaneAndLocationInfoMsg) {
	si.mu.Lock()
	defer si.mu.Unlock()
	if _, ok := si.frames[loc.Icao]; !ok {
		si.icaos = append(si.icaos, loc.Icao)
		slices.Sort(si.icaos)
	}
	si.frameCount++
	si.planes[loc.Icao] = loc
	si.frames[loc.Icao]++
}

func (si *sourceInfo) numFrames() string {
	si.mu.Lock()
	defer si.mu.Unlock()
	return strconv.FormatUint(si.frameCount, 10)
}

func (si *sourceInfo) numFramesFor(icao uint32) string {
	si.mu.Lock()
	defer si.mu.Unlock()
	return strconv.FormatUint(si.frames[icao], 10)
}

func (si *sourceInfo) getLoc(icao uint32) *export.PlaneAndLocationInfoMsg {
	si.mu.Lock()
	defer si.mu.Unlock()
	p := si.planes[icao]
	if nil == p {
		return nil
	}
	return p
}

func (ps planesSource) String() string {
	switch ps {
	case planesSourceIncoming:
		return "Incoming Data"
	case planesSourceIngest:
		return "After Ingest"
	case planesSourceEnriched:
		return "After Enrichment"
	case planesSourceRoutedLow:
		return "After Routing - Low"
	case planesSourceRoutedHigh:
		return "After Routing - High"
	case planesSourceWSLow:
		return "Websocket - Low"
	case planesSourceWSHigh:
		return "Websocket - High"
	}
	return "Unknown"
}
