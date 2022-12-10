package tracker

import (
	"plane.watch/lib/tracker/mode_s"
	"testing"
	"time"
)

func TestStackPushPopSingle(t *testing.T) {
	fl := newLossyFrameList(10)

	frame := &mode_s.Frame{}

	if 0 != fl.Len() {
		t.Errorf("stack should be empty")
	}

	fl.Push(frame)

	if c := fl.Len(); 1 != c {
		t.Errorf("incorrect length after push. expected 1, got %d", c)
	}

	popped := fl.Pop()
	if popped != frame {
		t.Errorf("Failed to get the correct frame from the stack")
	}

	if c := fl.Len(); 0 != c {
		t.Errorf("incorrect length after push. expected 0, got %d", c)
	}
}

func TestStackPushPopMultiple(t *testing.T) {
	numItems := 10
	doubleNumItems := 20
	originals := make([]*mode_s.Frame, doubleNumItems)
	fl := newLossyFrameList(numItems)

	for i := 0; i < doubleNumItems; i++ {
		originals[i] = mode_s.NewFrame("8E7C7F0D581176D7BB8D48CD7714", time.Now())
		fl.Push(originals[i])

		max := i + 1
		if i >= numItems {
			max = 10
		}
		if max != fl.Len() {
			t.Errorf("incorrect number of items in stack expected %d != len %d", max, fl.Len())
		}
	}

	// now test we can pop
	expectedNumItems := numItems
	for i := doubleNumItems - 1; i >= numItems; i-- {
		if c := fl.Len(); expectedNumItems != c {
			t.Errorf("Incorrect number of items in the queue. want: %d, got %d", expectedNumItems, c)
		}
		item := fl.Pop()
		expectedNumItems--
		if item != originals[i] {
			t.Errorf("Incorrect item fetched from stack pop %d != %d", item.TimeStamp().UnixNano(), originals[i].TimeStamp().UnixNano())
		}
	}
	if c := fl.Len(); 0 != c {
		t.Errorf("incorrect length after push. expected 0, got %d", c)
	}

	//empty pop
	if nil != fl.Pop() {
		t.Errorf("pop'd something off an empty queue")
	}
}

func BenchmarkLossyFrameList_Push(b *testing.B) {
	fl := newLossyFrameList(10)
	frame := mode_s.NewFrame("8E7C7F0D581176D7BB8D48CD7714", time.Now())

	for n := 0; n < b.N; n++ {
		fl.Push(frame)
	}
}
