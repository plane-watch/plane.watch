package producer

import (
	"bufio"
	"bytes"
	"plane.watch/lib/tracker/beast"
	"time"
)

func (p *producer) beastScanner(scan *bufio.Scanner) error {
	lastTimeStamp := time.Duration(0)
	for scan.Scan() {
		msg := scan.Bytes()
		frame := beast.NewFrame(msg, false)
		if nil == frame {
			continue
		}
		if p.beastDelay {
			currentTs := frame.BeastTicksNs()
			if lastTimeStamp > 0 && lastTimeStamp < currentTs {
				time.Sleep(currentTs - lastTimeStamp)
			}
			lastTimeStamp = currentTs
		}
		p.addFrame(frame, &p.FrameSource)

		if nil != p.stats.beast {
			p.stats.beast.Inc()
		}
	}
	return scan.Err()
}

// ScanBeast is a splitter for BEAST format messages
func ScanBeast(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// skip until we get our first 0x1A (message start)
	i := bytes.IndexByte(data, 0x1A)
	if -1 == i || len(data) < i+11 {
		// we do not even have the smallest message, let's get some more data
		return 0, nil, nil
	}
	// byte 2 is our message type, so it tells us how long this message is
	msgLen := 0
	switch data[i+1] {
	case 0x31:
		// mode-ac 11 bytes (2+8)
		// 1(esc), 1(type), 6(mlat), 1(signal), 2(mode-ac)
		msgLen = 11
	case 0x32:
		// mode-s short 16 bytes
		// 1(esc), 1(type), 6(mlat), 1(signal), 7(mode-s short)
		msgLen = 16
	case 0x33:
		// mode-s long 23 bytes
		// 1(esc), 1(type), 6(mlat), 1(signal), 14(mode-s extended squitter)
		msgLen = 23
	case 0x34:
		// Config Settings and Stats
		// 1(esc), 1(type), 6(mlat), 1(unused), (1)DIP Config, (1)timestamp error ticks
		msgLen = 11
	default:
		// unknown? assume we got an out of sequence and skip
		return i + 1, nil, nil
	}
	bufLen := len(data)
	//println("type", data[i+1], "input len", bufLen, "msg len",msgLen)
	if bufLen >= i+msgLen {
		// we have enough in our buffer
		// account for double escapes
		advance = msgLen
		max := msgLen * 2
		token = make([]byte, msgLen)
		dataIndex := i // start at the <esc>/0x1a
		tokenIndex := 0
		token[tokenIndex] = data[dataIndex]
		for dataIndex < i+advance-1 && dataIndex < i+max {
			dataIndex++ // first inc skips past the first 0x1a
			tokenIndex++
			// can we get to the next byte?
			if dataIndex+1 > bufLen {
				// run out of buffer, want more
				return 0, nil, nil
			}

			token[tokenIndex] = data[dataIndex]
			if data[dataIndex] != 0x1A { // messages start with 0x1A
				continue
			}
			if dataIndex+2 > bufLen {
				// run out of buffer, want more
				return 0, nil, nil
			}
			if data[dataIndex+1] == 0x1A { // skip over the second <esc>
				advance++
				dataIndex++
			}
		}
		return advance, token, nil
	}
	// we want more data!
	return 0, nil, nil
}
