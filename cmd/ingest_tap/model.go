package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"plane.watch/lib/export"
	"strings"
	"time"
)

type (
	planesSource int
	model        struct {
		tapper *PlaneWatchTapper

		focusIcaoList []string

		statsTable    table.Model
		selectedTable table.Model
		planesTable   table.Model
		source        planesSource

		inputAvr map[string][]string

		fromIngest   map[string][]export.Aircraft
		fromRouter   map[string][]export.Aircraft
		fromWsBroker map[string][]export.Aircraft

		width, height int
		tickCount     uint64
		tickDuration  time.Duration

		heading lipgloss.Style

		logs *strings.Builder
	}

	timerTick time.Time
)

func initialModel(natsUrl, wsUrl string) (*model, error) {
	logs := &strings.Builder{}
	logger := zerolog.New(logs).With().Timestamp().Logger()

	m := &model{
		tapper:        NewPlaneWatchTapper(WithLogger(logger)),
		tickDuration:  time.Millisecond * 16,
		focusIcaoList: make([]string, 0),
		inputAvr:      make(map[string][]string),
		fromIngest:    make(map[string][]export.Aircraft),
		fromRouter:    make(map[string][]export.Aircraft),
		fromWsBroker:  make(map[string][]export.Aircraft),
		source:        planesSourceWS,
		logs:          logs,
	}
	if err := m.tapper.Connect(natsUrl, wsUrl); err != nil {
		return nil, err
	}
	m.buildTables()
	m.configureStyles()

	return m, nil
}

func (m *model) configureStyles() {
	m.heading = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		PaddingTop(0).
		PaddingLeft(3).
		PaddingBottom(1)

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
			{Title: "# Receivers", Width: 11},
			{Title: "# Beast", Width: 10},
			{Title: "# Frames", Width: 10},
			{Title: "# Low", Width: 10},
			{Title: "# High", Width: 10},
			{Title: "# WS", Width: 10},
		}),
		table.WithRows([]table.Row{
			{"0", "0", "0", "0", "0", "0"},
		}),
		table.WithHeight(2),
		table.WithFocused(false),
	)
	m.selectedTable = table.New(
		table.WithColumns([]table.Column{
			{Title: "# Receivers", Width: 11},
			{Title: "# Beast", Width: 10},
			{Title: "# Frames", Width: 10},
			{Title: "# Low", Width: 10},
			{Title: "# High", Width: 10},
			{Title: "# WS", Width: 10},
		}),
		table.WithRows([]table.Row{
			{"0", "0", "0", "0", "0", "0"},
		}),
		table.WithHeight(2),
		table.WithFocused(false),
	)
	m.planesTable = table.New(
		table.WithColumns([]table.Column{
			{Title: "ICAO", Width: 6},
			{Title: "Receivers", Width: 9},
			{Title: "Lat", Width: 10},
			{Title: "Lon", Width: 10},
			{Title: "Altitude", Width: 10},
			{Title: "Heading", Width: 10},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
	)
}
