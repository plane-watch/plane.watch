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

func (f *Filter) drain() {
	for range f.Listen() {
	}
}

func TestFilter_Handle(t *testing.T) {
	source := &tracker.FrameSource{}

	filter := NewFilter()
	go filter.drain()

	frame := beast.NewFrame(beastModeSShort, false)

	resp := filter.Handle(frame, source)

	if resp != frame {
		t.Errorf("Expected the same frame back")
	}

	if nil != filter.Handle(frame, source) {
		t.Errorf("Got a duplicated frame back")
	}
}

func BenchmarkFilter_HandleDuplicates(b *testing.B) {
	source := &tracker.FrameSource{}

	filter := NewFilter()
	go filter.drain()

	frame := beast.NewFrame(beastModeSShort, false)
	filter.Handle(frame, source)

	for n := 0; n < b.N; n++ {
		if nil != filter.Handle(frame, source) {
			b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
		}
	}
}

func BenchmarkFilter_HandleUnique(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	source := &tracker.FrameSource{}

	max := 0x00FFFFFF

	messages := make([]*beast.Frame, 0, max)

	// setup our test data
	template := make([]byte, len(beastModeSShort))
	copy(template, beastModeSShort)
	template[13] = 0
	template[14] = 0
	template[15] = 0
	for x := 0; x <= 255; x++ {
		for y := 0; y <= 255; y++ {
			for z := 0; z <= 255; z++ {
				shrt := make([]byte, len(beastModeSShort))
				copy(shrt, template)
				shrt[13] = byte(x)
				shrt[14] = byte(y)
				shrt[15] = byte(z)
				frame := beast.NewFrame(shrt, false)
				messages = append(messages, frame)
			}
		}
	}

	b.ResetTimer()
	filter := NewFilter()
	go filter.drain()

	b.Run("Actual Test", func(bb *testing.B) {
		for n := 0; n < bb.N; n++ {
			if n > 16777215 {
				bb.Error("not prepared that many tests")
			}
			if nil != filter.Handle(messages[n], source) {
				//b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
			}
		}
	})

	filter.Stop()
}
