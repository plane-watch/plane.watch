package dedupe

import (
	"fmt"
	"github.com/rs/zerolog"
	"plane.watch/lib/tracker/beast"
	"testing"
	"time"
)

var (
	beastModeSShort = []byte{0x1a, 0x32, 0x22, 0x1b, 0x54, 0xf0, 0x81, 0x2b, 0x26, 0x5d, 0x7c, 0x49, 0xf8, 0x28, 0xe9, 0x43}
)

func TestFilter_HandleOld(t *testing.T) {
	filter := NewFilter()

	frame, err := beast.NewFrame(beastModeSShort, false)
	if nil != err {
		t.Error(err)
	}

	resp := filter.HandleOld(&frame)

	if resp == nil {
		t.Errorf("Expected the same frame back")
	}

	if nil != filter.HandleOld(&frame) {
		t.Errorf("Got a duplicated frame back")
	}
}

func TestFilter_HandleNew(t *testing.T) {
	filter := NewFilter()

	frame, _ := beast.NewFrame(beastModeSShort, false)

	resp := filter.Handle(&frame)

	if resp == nil {
		t.Errorf("Expected the same frame back")
	}

	if nil != filter.Handle(&frame) {
		t.Errorf("Got a duplicated frame back")
	}
}

func TestBTreeSweep1(t *testing.T) {

	filter := NewFilter(WithSweeperDuration(0), WithDedupeMaxAge(0))

	frame, _ := beast.NewFrame(beastModeSShort, false)

	if nil == filter.Handle(&frame) {
		t.Errorf("Expected to add a frame")
	}

	if 1 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length. expected 1: Got %d", filter.btree.Len())
	}

	filter.sweep()

	if 0 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length. expected 0: Got %d", filter.btree.Len())
	}
}

func TestBTreeSweep1000(t *testing.T) {
	filter := NewFilter(WithSweeperDuration(0), WithDedupeMaxAge(0))

	messages := makeBeastMessages(9)
	if len(messages) != 1000 {
		t.Errorf("Wrong number of messages, expected 1000: got %d", len(messages))
	}

	for _, msg := range messages {
		if nil == filter.Handle(&msg) {
			t.Errorf("Expected to add a frame")
		}
	}

	if 1000 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length Before Sweep. expected 1000: Got %d", filter.btree.Len())
	}

	filter.sweep()

	if 0 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length After Sweep. expected 0: Got %d", filter.btree.Len())
	}
}

func TestBTreeSweepNewAndOld(t *testing.T) {
	filter := NewFilter(WithSweeperDuration(0), WithDedupeMaxAge(time.Minute))

	messages := makeBeastMessages(9)
	var tNow = time.Now()
	var tOld = time.Now().Add(-time.Hour)
	for i, msg := range messages {
		if i < 500 {
			filter.btree.ReplaceOrInsert(FrameAndTime{
				frame: msg.Raw(),
				time:  tNow,
			})
		} else {
			filter.btree.ReplaceOrInsert(FrameAndTime{
				frame: msg.Raw(),
				time:  tOld,
			})
		}
	}

	if 1000 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length Before Sweep. expected 1000: Got %d", filter.btree.Len())
	}

	filter.sweep()

	if 500 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length After Sweep. expected 500: Got %d", filter.btree.Len())
	}
	filter.sweep()

	if 500 != filter.btree.Len() {
		t.Errorf("Incorrect BTree Length After Sweep. expected 500: Got %d", filter.btree.Len())
	}
}

func BenchmarkFilter_HandleDuplicates(b *testing.B) {
	filter := NewFilter()

	frame, _ := beast.NewFrame(beastModeSShort, false)
	filter.HandleOld(&frame)

	for n := 0; n < b.N; n++ {
		if nil != filter.HandleOld(&frame) {
			b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
		}
	}
}

func BenchmarkFilter_HandleDuplicatesBtree(b *testing.B) {
	filter := NewFilter()

	frame, _ := beast.NewFrame(beastModeSShort, false)
	filter.Handle(&frame)

	for n := 0; n < b.N; n++ {
		if nil != filter.Handle(&frame) {
			b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
		}
	}
}

func makeBeastMessages(iterMax int) []beast.Frame {
	//max := 0x00FFFFFF
	max := iterMax * iterMax * iterMax
	messages := make([]beast.Frame, 0, max)

	// setup our test data
	template := make([]byte, len(beastModeSShort))
	copy(template, beastModeSShort)
	template[13] = 0
	template[14] = 0
	template[15] = 0
	for x := 0; x <= iterMax; x++ {
		for y := 0; y <= iterMax; y++ {
			for z := 0; z <= iterMax; z++ {
				shrt := make([]byte, len(beastModeSShort))
				copy(shrt, template)
				shrt[13] = byte(x)
				shrt[14] = byte(y)
				shrt[15] = byte(z)
				frame, _ := beast.NewFrame(shrt, false)
				messages = append(messages, frame)
			}
		}
	}
	return messages
}

func BenchmarkFilter_HandleUnique(b *testing.B) {
	b.StopTimer()
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	messages := makeBeastMessages(255)

	b.Run("ForgetfulSyncMap", func(bb *testing.B) {
		filter := NewFilter()
		for n := 0; n < bb.N; n++ {
			if n > 16777215 {
				bb.Error("not prepared that many tests")
			}
			filter.HandleOld(&messages[n])
			if nil != filter.HandleOld(&messages[n]) {
				bb.Fatal("Failed duplicate insert")
			}
		}
		filter.Stop()
	})

	for i := 2; i <= 20; i++ {
		filterG := NewFilter(WithBtreeDegree(i), WithSweeperDuration(0))
		b.Run(fmt.Sprintf("GenericBTree_%d", i), func(bb *testing.B) {
			for n := 0; n < bb.N; n++ {
				if n > 16_777_215 {
					bb.Error("not prepared that many tests")
				}
				filterG.Handle(&messages[n])
				if nil != filterG.Handle(&messages[n]) {
					bb.Fatal("Failed duplicate insert")
				}
			}
		})
		filterG.Stop()
	}

}
