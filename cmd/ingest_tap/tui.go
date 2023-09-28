package main

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
	"time"
)

const (
	planesSourceIngest planesSource = iota
	planesSourceLow
	planesSourceHigh
	planesSourceWS
)

func runTui(c *cli.Context) error {
	m, err := initialModel(c.String(natsURL), c.String(websocketURL))
	if err != nil {
		return err
	}
	defer m.tapper.Disconnect()

	if _, err = tea.NewProgram(m).Run(); nil != err {
		return err
	}
	return nil
}

func (m *model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch tMsg := msg.(type) {
	case tea.KeyMsg:
		switch tMsg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = tMsg.Width
		m.height = tMsg.Height
		m.statsTable.SetWidth(m.width)
		m.selectedTable.SetWidth(m.width)
		m.planesTable.SetWidth(m.width)
		return m, nil
	case timerTick:
		m.tickCount++
		m.statsTable.SetRows([]table.Row{
			{"0", "0", "0", "0", "0", "0"},
		})
		m.selectedTable.SetRows([]table.Row{
			{"0", "0", "0", "0", "0", "0"},
		})
		return m, m.tickCmd()
	}
	var cmd1, cmd2, cmd3 tea.Cmd
	m.statsTable, cmd1 = m.statsTable.Update(msg)
	m.selectedTable, cmd2 = m.statsTable.Update(msg)
	m.planesTable, cmd2 = m.statsTable.Update(msg)

	return m, tea.Batch(cmd1, cmd2, cmd3)
}

func (m *model) View() string {
	return m.heading.Render("Received Data Stats") + "\n" +
		m.statsTable.View() + "\n" +
		m.heading.Render("Selected Aircraft") + "\n" +
		m.selectedTable.View() + "\n" +
		m.heading.Render("All Aircraft") + "\n" +
		m.planesTable.View() + "\n"
}

func (m *model) tickCmd() tea.Cmd {
	return tea.Tick(m.tickDuration, func(t time.Time) tea.Msg {
		return timerTick(t)
	})
}
