package middleware

import (
	"testing"
)

func TestLinkedListAddTest(t *testing.T) {
	tap := IngestTap{}
	if nil != tap.head {
		t.Errorf("Zero value of IngestTap has a non nil head")
		return
	}
	if nil != tap.tail {
		t.Errorf("Zero value of IngestTap has a non nil tail")
		return
	}

	cond1 := &condition{}
	cond2 := &condition{}
	cond3 := &condition{}

	///////// First item
	tap.append(cond1)
	if cond1 != tap.head {
		t.Errorf("Failed to set tap.head to our newly added item")
		return
	}
	if cond1 != tap.tail {
		t.Errorf("Failed to set tap.tail to our newly added item")
		return
	}

	///////// Second item
	tap.append(cond2)
	if cond1 != tap.head {
		t.Errorf("Failed to set tap.head to our newly added second item")
		return
	}
	if cond2 != tap.tail {
		t.Errorf("Failed to set tap.tail to our newly added second item")
		return
	}
	if nil != cond1.prevItem {
		t.Errorf("Somehow the tap.head has a previous item")
		return
	}

	if cond2 != cond1.nextItem {
		t.Errorf("Did not correctly set cond2 as cond1's next item")
		return
	}
	if cond2.prevItem != cond1 {
		t.Errorf("Did not correctly set cond2's prevItem to be cond1")
		return
	}

	///////// Third item
	tap.append(cond3)
	if cond3 != tap.tail {
		t.Errorf("Failed to set tap.tail to our newly added third item")
		return
	}
	if cond3 != cond2.nextItem {
		t.Errorf("Failed to set cond2's nextItem to our newly added third item")
		return
	}
	if cond2 != cond3.prevItem {
		t.Errorf("Failed to set cond3's prevItem to our second item")
		return
	}

	if tap.head != cond1 || cond1.nextItem != cond2 {
		t.Errorf("cond1 is not setup correctly after adding third item")
	}
}

func TestLinkedListAddRemoveNone(t *testing.T) {
	tap := IngestTap{}
	if nil != tap.head {
		t.Errorf("Zero value of IngestTap has a non nil head")
		return
	}
	if nil != tap.tail {
		t.Errorf("Zero value of IngestTap has a non nil tail")
		return
	}

	cond1 := &condition{apiKey: "1"}
	cond2 := &condition{apiKey: "2"}
	cond3 := &condition{apiKey: "3"}
	cond4 := &condition{apiKey: "4"}

	tap.append(cond1)
	tap.append(cond2)
	tap.append(cond3)

	tap.remove(cond4)

	///////// Third item
	if tap.head != cond1 {
		t.Errorf("tap.head != cond1")
		return
	}
	if nil != cond1.prevItem {
		t.Errorf("nil != cond1.prevItem")
		return
	}
	if cond1.nextItem != cond2 {
		t.Errorf("cond1.nextItem != cond2")
		return
	}
	if cond2.prevItem != cond1 {
		t.Errorf("cond2.prevItem != cond1")
		return
	}
	if cond2.nextItem != cond3 {
		t.Errorf("cond2.nextItem != cond3")
		return
	}
	if cond3.prevItem != cond2 {
		t.Errorf("cond3.prevItem != cond2")
		return
	}
	if cond3.nextItem != nil {
		t.Errorf("cond3.nextItem != nil")
		return
	}

}

func TestLinkedListAddRemoveHead(t *testing.T) {
	tap := IngestTap{}
	if nil != tap.head {
		t.Errorf("Zero value of IngestTap has a non nil head")
		return
	}
	if nil != tap.tail {
		t.Errorf("Zero value of IngestTap has a non nil tail")
		return
	}

	cond1 := &condition{apiKey: "1"}
	cond2 := &condition{apiKey: "2"}
	cond3 := &condition{apiKey: "3"}

	tap.append(cond1)
	tap.append(cond2)
	tap.append(cond3)

	tap.remove(cond1)

	if nil != cond1.nextItem || nil != cond1.prevItem {
		t.Errorf("Failed to unset next/prev on cond1")
		return
	}

	if cond2 != tap.head {
		t.Errorf("After removal of head, cond2 should be tap.head")
		return
	}

	if cond2.prevItem != nil {
		t.Errorf("tap.head has a prev item")
		return
	}

	if cond2.nextItem != cond3 {
		t.Errorf("cond2.next item should be cond3")
		return
	}

}

func TestLinkedListAddRemoveMiddle(t *testing.T) {
	tap := IngestTap{}
	if nil != tap.head {
		t.Errorf("Zero value of IngestTap has a non nil head")
		return
	}
	if nil != tap.tail {
		t.Errorf("Zero value of IngestTap has a non nil tail")
		return
	}

	cond1 := &condition{apiKey: "1"}
	cond2 := &condition{apiKey: "2"}
	cond3 := &condition{apiKey: "3"}

	tap.append(cond1)
	tap.append(cond2)
	tap.append(cond3)

	tap.remove(cond2)

	if nil != cond2.nextItem || nil != cond2.prevItem {
		t.Errorf("Failed to unset next/prev on cond2")
		return
	}

	if tap.head != cond1 {
		t.Errorf("tap.head should be cond1")
		return
	}
	if tap.tail != cond3 {
		t.Errorf("tap.tail should be cond3")
		return
	}
	if cond1.nextItem != cond3 {
		t.Errorf("cond1.nextItem should be cond3")
		return
	}
	if cond3.prevItem != cond1 {
		t.Errorf("cond3.prevItem should be cond1")
		return
	}
	if nil != cond3.nextItem {
		t.Errorf("cond3.nextItem should be nil")
		return
	}
	if nil != cond1.prevItem {
		t.Errorf("cond1.prevItem should be nil")
		return
	}
}

func TestLinkedListAddRemoveTail(t *testing.T) {
	tap := IngestTap{}
	if nil != tap.head {
		t.Errorf("Zero value of IngestTap has a non nil head")
		return
	}
	if nil != tap.tail {
		t.Errorf("Zero value of IngestTap has a non nil tail")
		return
	}

	cond1 := &condition{apiKey: "1"}
	cond2 := &condition{apiKey: "2"}
	cond3 := &condition{apiKey: "3"}

	tap.append(cond1)
	tap.append(cond2)
	tap.append(cond3)

	tap.remove(cond3)

	if nil != cond3.nextItem || nil != cond3.prevItem {
		t.Errorf("Failed to unset next/prev on cond3")
		return
	}

	if cond2 != tap.tail {
		t.Errorf("tap.tail should be cond2")
	}
	if cond1 != tap.head {
		t.Errorf("tap.head should be cond1")
	}
	if cond1.nextItem != cond2 {
		t.Errorf("cond1.nextItem != cond2")
	}
	if cond2.prevItem != cond1 {
		t.Errorf("cond2.prevItem != cond1")
	}
}
