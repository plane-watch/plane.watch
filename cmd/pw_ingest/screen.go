package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"plane.watch/lib/tracker"
	"sort"
	"sync"
	"time"
)

type (
	display struct {
		app    *tview.Application
		top    *tview.Table
		bottom *tview.TextView

		appLock sync.Mutex

		planes sync.Map
	}
	pacerEvent struct {
	}
)

func (p pacerEvent) Type() string {
	return "pacer"
}

func (p pacerEvent) String() string {
	return "beep"
}

func newPacer() *pacerEvent {
	return &pacerEvent{}
}

func (d *display) App() *tview.Application {
	d.appLock.Lock()
	defer d.appLock.Unlock()
	return d.app
}

func (d *display) Run() error {
	go func() {
		c := time.NewTicker(time.Second)
		defer c.Stop()
		for {
			select {
			case <-c.C:
				d.OnEvent(newPacer())
			}
		}
	}()
	if err := d.App().Run(); nil != err {
		return err
	}
	return nil
}

func newAppDisplay() (*display, error) {
	d := display{}
	d.app = tview.NewApplication()
	d.top = tview.NewTable()
	d.top.SetTitle("Current Planes")

	d.bottom = tview.NewTextView().SetDynamicColors(true)
	d.bottom.SetBorder(true).SetTitle("Logs")

	_, _, _, height := d.bottom.GetRect()
	d.bottom.SetWordWrap(false).SetMaxLines(height + 50)
	d.bottom.SetChangedFunc(func() {
		d.app.Draw()
	})
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.top, 0, 1, true).
		AddItem(d.bottom, 0, 1, false)
	d.app.SetRoot(flex, true).EnableMouse(false)

	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Key() == tcell.KeyEscape {
			// we need to exit!
			d.app.Stop()
		}
		return event
	})

	d.drawTable()

	return &d, nil
}

func (d *display) sortedPlaneIdSlice() []uint32 {
	icaos := make([]uint32, 0)
	d.planes.Range(func(key, value interface{}) bool {
		icaos = append(icaos, key.(uint32))
		return true
	})
	sort.Slice(icaos, func(i, j int) bool {
		return icaos[i] < icaos[j]
	})
	return icaos
}
func (d *display) getPlaneRow(icao uint32) int {
	list := d.sortedPlaneIdSlice()
	for i, id := range list {
		if id == icao {
			return i + 1
		}
	}
	return -1
}
func (d *display) drawTable() {
	icaoList := d.sortedPlaneIdSlice()
	if d.top.GetRowCount() != len(icaoList)+1 {
		d.top.Clear()
	}

	hdr := func(title string, width int) *tview.TableCell {
		return &tview.TableCell{
			Text:          title,
			MaxWidth:      width,
			Color:         tcell.ColorYellowGreen,
			NotSelectable: true,
			Expansion:     1,
		}
	}

	d.top.
		SetCell(0, 0, hdr("ICAO", 9)).
		SetCell(0, 1, hdr("Ident", 9)).
		SetCell(0, 2, hdr("Squawk", 6)).
		SetCell(0, 3, hdr("Lat/Lon", 19)).
		SetCell(0, 4, hdr("Altitude", 12)).
		SetCell(0, 5, hdr("Speed", 12)).
		SetCell(0, 6, hdr("Heading", 12)).
		SetCell(0, 7, hdr("# Msgs", 6)).
		SetCell(0, 8, hdr("Age (s)", 7)).
		SetCell(0, 9, hdr("Extra", 10)).
		SetBorder(true)

	row := 0
	for _, icao := range icaoList {
		row++
		d.drawRow(icao, row)
	}
	d.top.ScrollToBeginning()
}

func (d *display) drawRow(icao uint32, row int) {
	var latLon string
	item, found := d.planes.Load(icao)
	if !found {
		return
	}
	plane := item.(*tracker.Plane)

	d.top.SetCellSimple(row, 0, plane.IcaoIdentifierStr())
	d.top.SetCellSimple(row, 1, plane.FlightNumber())
	d.top.SetCellSimple(row, 2, plane.SquawkIdentityStr())
	if plane.HasLocation() {
		latLon = fmt.Sprintf("%0.4f/%0.4f", plane.Lat(), plane.Lon())
	} else {
		latLon = "?"
	}
	d.top.SetCellSimple(row, 3, latLon)
	d.top.SetCellSimple(row, 4, fmt.Sprint(plane.Altitude()))
	d.top.SetCellSimple(row, 5, fmt.Sprintf("%0.2f", plane.Velocity()))
	d.top.SetCellSimple(row, 6, plane.HeadingStr())
	d.top.SetCellSimple(row, 7, fmt.Sprint(plane.MsgCount()))

	since := time.Now().Sub(plane.LastSeen()).Seconds()
	d.top.SetCellSimple(row, 8, fmt.Sprintf("%0.0f", since))
	d.top.SetCellSimple(row, 9, plane.Special())
}

func (d *display) updateAgeColumn() {
	icaoList := d.sortedPlaneIdSlice()

	row := 0
	for _, icao := range icaoList {
		row++

		item, found := d.planes.Load(icao)
		if !found {
			return
		}
		plane := item.(*tracker.Plane)
		since := time.Now().Sub(plane.LastSeen()).Seconds()
		d.top.SetCellSimple(row, 8, fmt.Sprintf("%0.0f", since))
	}
}

func (d *display) Finish() {

}
func (d *display) OnEvent(e tracker.Event) {
	switch e.(type) {
	case *tracker.PlaneLocationEvent:
		d.App().QueueUpdate(func() {
			ple := e.(*tracker.PlaneLocationEvent)
			if ple.Removed() {
				//_, _ = fmt.Fprintln(d.bottom, "Remove Plane", ple.Plane())
				//d.top.RemoveRow(d.getPlaneRow(ple.Plane().IcaoIdentifier()))
				d.planes.Delete(ple.Plane().IcaoIdentifier())
			} else {
				d.planes.Store(ple.Plane().IcaoIdentifier(), ple.Plane())
				//d.drawRow(ple.Plane().IcaoIdentifier(), d.getPlaneRow(ple.Plane().IcaoIdentifier()))
			}
		})
		//d.drawTable()

	case *tracker.FrameEvent:
		// show the received frame
	case *pacerEvent:
		// cleanup our planes list
		d.App().QueueUpdate(func() {
			d.drawTable()
		})
		//d.drawTable()
	}
}
func (d *display) Stop() {
	d.app.Stop()
}

func (d *display) HealthCheck() bool {
	return true
}

func (d *display) HealthCheckName() string {
	return "TUI Display"
}
