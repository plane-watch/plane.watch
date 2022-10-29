package dedupe

import (
	"github.com/rs/zerolog"
	"plane.watch/lib/tracker/beast"
	"testing"
	"time"
)

func TestBTreeSweep1(t *testing.T) {

	filter := NewFilterBTree(WithSweeperInterval(0), WithDedupeMaxAge(0))

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

func TestFilterBTree_Handle(t *testing.T) {
	filter := NewFilterBTree()

	frame, _ := beast.NewFrame(beastModeSShort, false)

	resp := filter.Handle(&frame)

	if resp == nil {
		t.Errorf("Expected the same frame back")
	}

	if nil != filter.Handle(&frame) {
		t.Errorf("Got a duplicated frame back")
	}
}

func TestBTreeSweep1000(t *testing.T) {
	filter := NewFilterBTree(WithSweeperInterval(0), WithDedupeMaxAge(0))

	messages := makeBeastMessages(9)
	if len(messages) != 1000 {
		t.Errorf("Wrong number of messages, expected 1000: got %d", len(messages))
	}

	for _, msg := range messages {
		if nil == filter.Handle(msg) {
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
	filter := NewFilterBTree(WithSweeperInterval(0), WithDedupeMaxAge(time.Minute))

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

func BenchmarkFilterBTree_HandleDuplicates(b *testing.B) {
	filter := NewFilterBTree(WithSweeperInterval(0), WithDedupeMaxAge(time.Minute))

	frame, _ := beast.NewFrame(beastModeSShort, false)
	filter.Handle(&frame)

	for n := 0; n < b.N; n++ {
		if nil != filter.Handle(&frame) {
			b.Error("Should not have gotten a non empty response - duplicate handled incorrectly!?")
		}
	}
}

func BenchmarkFilterBTree_HandleUnique(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	degree := 16
	filter := NewFilterBTree(WithBtreeDegree(degree), WithSweeperInterval(0))
	for n := 0; n < b.N; n++ {
		beastModeSTest := []byte{0x1a, 0x32, 0x22, 0x1b, 0x54, 0xf0, 0x81, 0x2b, 0x26, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n), 0, 0, byte(degree)}
		msg, _ := beast.NewFrame(beastModeSTest, false)
		if nil == filter.Handle(&msg) {
			b.Fatalf("Expected to insert new message %0X", beastModeSTest)
		}
		if nil != filter.Handle(&msg) {
			b.Fatalf("Failed duplicate insert of %0X", beastModeSTest)
		}
	}
	if filter.btree.Len() != b.N {
		b.Errorf("Did not get the same number of items as tested. expected %d, got %d", b.N, filter.btree.Len())
	}
	filter.Stop()
}
