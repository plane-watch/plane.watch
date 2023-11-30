package sbs1

import "testing"

func TestIcaoStringToInt(t *testing.T) {
	sut := "7C1BE8"
	expected := uint32(8133608)
	icaoAddr, err := icaoStringToInt(sut)
	if err != nil {
		t.Error(err)
	}
	if icaoAddr != expected {
		t.Errorf("Expected %s to decode to %d, but got %d", sut, expected, icaoAddr)
	}
}

func TestKeepAliveNoOp(t *testing.T) {
	trimableStrings := []string{
		"",
		"\n",
		"\r",
		"\r\n",
		" ",
		" \n",
		"  ",
		"  \n",
		"   ",
		"   \n",
		"    ",
		"    \n",
		"\n\n",
	}

	for _, s := range trimableStrings {
		f := NewFrame(s)
		if nil != f.Parse() {
			t.Errorf("Should not have gotten an error for a newline")
		}
	}
}
