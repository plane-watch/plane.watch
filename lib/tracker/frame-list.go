package tracker

import "plane.watch/lib/tracker/mode_s"

type (
	lossyFrame struct {
		item       *mode_s.Frame
		next, prev *lossyFrame
	}
	lossyFrameList struct {
		head, tail *lossyFrame
		capacity   int
		numItems   int
	}
)

func newLossyFrameList(numItems int) *lossyFrameList {
	return &lossyFrameList{
		head:     nil,
		tail:     nil,
		capacity: numItems,
		numItems: 0,
	}
}

func (fl *lossyFrameList) Push(f *mode_s.Frame) {
	item := lossyFrame{
		item: f,
		next: fl.head,
		prev: nil,
	}
	if nil != fl.head {
		fl.head.prev = &item
	}
	fl.head = &item
	if nil == fl.tail {
		fl.tail = fl.head
	}
	fl.numItems++
	if fl.numItems > fl.capacity {
		_ = fl.Unshift()
	}
}

// Unshift shortens the queue by 1, taking from the tail.
func (fl *lossyFrameList) Unshift() *mode_s.Frame {
	if nil == fl.tail {
		return nil
	}
	t := fl.tail
	fl.numItems--
	fl.tail = t.prev
	fl.tail.next = nil
	t.prev = nil
	t.next = nil
	if 0 == fl.numItems {
		fl.tail = nil
		fl.head = nil
	}

	defer func() {
		t = nil
	}()

	return t.item
}

func (fl *lossyFrameList) Pop() *mode_s.Frame {
	if nil == fl.head {
		return nil
	}
	t := fl.head
	fl.numItems--
	fl.head = t.next
	t.prev = nil
	t.next = nil
	if 0 == fl.numItems {
		fl.tail = nil
		fl.head = nil
	}

	defer func() {
		t = nil
	}()

	return t.item
}

func (fl *lossyFrameList) Len() int {
	return fl.numItems
}

func (fl *lossyFrameList) Range(f func(f *mode_s.Frame) bool) {
	t := fl.head
	for t.next != nil {
		if !f(t.item) {
			return
		}
		t = t.next
	}
}
