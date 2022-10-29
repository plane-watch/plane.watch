package sink

import (
	"github.com/rs/zerolog"
	"plane.watch/lib/tracker"
	"testing"
)

type drain struct {
	numJsonPublished int
	numTextPublished int
}

func (d *drain) PublishJson(queue string, msg []byte) error {
	d.numJsonPublished++
	return nil
}

func (d *drain) PublishText(queue string, msg []byte) error {
	d.numTextPublished++
	return nil
}

func (d *drain) Stop() {
}

func (d *drain) HealthCheckName() string {
	return "test drain"
}

func (d *drain) HealthCheck() bool {
	return true
}

func BenchmarkSink_TrackerMsgJson(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.NoLevel)
	sink := NewSink(&Config{sourceTag: "test"}, nil).(*Sink)
	plane := tracker.NewTracker().GetPlane(0x223344)

	le := tracker.NewPlaneLocationEvent(plane)

	for n := 0; n < b.N; n++ {
		_, err := sink.trackerMsgJson(le)
		if err != nil {
			return
		}
	}
}

func BenchmarkSink_OnEvent(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.NoLevel)
	d := drain{}
	sink := NewSink(&Config{sourceTag: "test"}, &d).(*Sink)
	plane := tracker.NewTracker().GetPlane(0x223344)

	le := tracker.NewPlaneLocationEvent(plane)

	for n := 0; n < b.N; n++ {
		sink.OnEvent(le)
	}

	if d.numJsonPublished != b.N {
		b.Errorf("Incorrect number of frames handled. Expected %d, got %d", b.N, d.numJsonPublished)
	}
}
