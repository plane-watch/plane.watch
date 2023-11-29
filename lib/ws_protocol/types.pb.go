package ws_protocol

import "strconv"

func (l *AircraftListPB) Len() int {
	if l == nil || l.Aircraft == nil {
		return 0
	}
	return len(l.Aircraft)
}

func (l *AircraftListPB) Less(i, j int) bool {
	if nil == l {
		return false
	}

	left := l.Aircraft[i].CallSign + ":" + l.Aircraft[i].Registration + ":" + strconv.FormatUint(uint64(l.Aircraft[i].Icao), 16)
	right := l.Aircraft[j].CallSign + ":" + l.Aircraft[j].Registration + ":" + strconv.FormatUint(uint64(l.Aircraft[j].Icao), 16)

	return left < right
}

func (l *AircraftListPB) Swap(i, j int) {
	l.Aircraft[i], l.Aircraft[j] = l.Aircraft[j], l.Aircraft[i]
}
