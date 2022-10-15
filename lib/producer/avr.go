package producer

import (
	"bufio"
	"plane.watch/lib/tracker/mode_s"
	"time"
)

func (p *Producer) avrScanner(scan *bufio.Scanner) error {
	for scan.Scan() {
		line := scan.Text()
		p.addFrame(mode_s.NewFrame(line, time.Now()), &p.FrameSource)
		p.addDebug("AVR Frame: %s", line)
		if nil != p.stats.avr {
			p.stats.avr.Inc()
		}
	}
	return scan.Err()
}
