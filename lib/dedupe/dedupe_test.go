package dedupe

import (
	"github.com/rs/zerolog"
	"plane.watch/lib/tracker"
	"plane.watch/lib/tracker/beast"
	"testing"
)

var (
	beastModeSShort = []byte{0x1a, 0x32, 0x22, 0x1b, 0x54, 0xf0, 0x81, 0x2b, 0x26, 0x5d, 0x7c, 0x49, 0xf8, 0x28, 0xe9, 0x43}
)

func TestFilter_Handle(t *testing.T) {
	filter := NewFilter()

	frame, err := beast.NewFrame(beastModeSShort, false)
	if nil != err {
		t.Error(err)
	}
	fe := tracker.NewFrameEvent(&frame, nil)

	resp := filter.Handle(&fe)

	if resp == nil {
		t.Errorf("Expected the same frame back")
	}

	if nil != filter.Handle(&fe) {
		t.Errorf("Got a duplicated frame back")
	}
}

func BenchmarkFilter_HandleDuplicates(b *testing.B) {
	filter := NewFilter()

	frame, _ := beast.NewFrame(beastModeSShort, false)
	fe := tracker.NewFrameEvent(&frame, nil)
	filter.Handle(&fe)

	for n := 0; n < b.N; n++ {
		if nil != filter.Handle(&fe) {
			b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
		}
	}
}

func BenchmarkFilter_HandleUnique(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	filter := NewFilter()
	for n := 0; n < b.N; n++ {
		beastModeSTest := []byte{0x1a, 0x32, 0x22, 0x1b, 0x54, 0xf0, 0x81, 0x2b, 0x26, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n), 0, 0, 0}
		msg, _ := beast.NewFrame(beastModeSTest, false)
		fe := tracker.NewFrameEvent(&msg, nil)

		if nil == filter.Handle(&fe) {
			b.Fatal("Expected to insert new message")
		}
		if nil != filter.Handle(&fe) {
			b.Fatal("Failed duplicate insert")
		}
	}

	if int(filter.list.Len()) != b.N {
		b.Errorf("Did not get the same number of items as tested. expected %d, got %d", b.N, filter.list.Len())
	}
}
