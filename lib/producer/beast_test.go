package producer

import (
	"bufio"
	"bytes"
	"os"
	"plane.watch/lib/tracker"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	beastModeAc     = []byte{0x1A, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	beastModeSShort = []byte{0x1a, 0x32, 0x22, 0x1b, 0x54, 0xf0, 0x81, 0x2b, 0x26, 0x5d, 0x7c, 0x49, 0xf8, 0x28, 0xe9, 0x43}
	beastModeSLong  = []byte{0x1a, 0x33, 0x22, 0x1b, 0x54, 0xac, 0xc2, 0xe9, 0x28, 0x8d, 0x7c, 0x49, 0xf8, 0x58, 0x41, 0xd2, 0x6c, 0xca, 0x39, 0x33, 0xe4, 0x1e, 0xcf}

	beastModeSLongDoubleEsc     = []byte{0x1a, 0x33, 0x22, 0x1b, 0x55, 0xe4, 0x1a, 0x1a, 0xa2, 0x2d, 0x8d, 0x7c, 0x49, 0xf8, 0xe1, 0x1e, 0x2f, 0x00, 0x00, 0x00, 0x00, 0xee, 0xcc, 0x47}
	beastModeSLongDoubleRemoved = []byte{0x1a, 0x33, 0x22, 0x1b, 0x55, 0xe4, 0x1a, 0xa2, 0x2d, 0x8d, 0x7c, 0x49, 0xf8, 0xe1, 0x1e, 0x2f, 0x00, 0x00, 0x00, 0x00, 0xee, 0xcc, 0x47}

	// let's see if we can get a buffer overrun
	beastModeSShortBad = []byte{0xBB, 0x1A, 0x33, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x8D, 0x4D, 0x22, 0x72, 0x99, 0x08, 0x41, 0xB7, 0x90, 0x6C, 0x28, 0x91, 0xA8, 0x1A}
	//                              | ESC | TYPE| MLAT                              | SIG | MODE S LONG
	noBeast = []byte{0x31, 0x33, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x8D, 0x4D, 0x22, 0x72, 0x99, 0x08, 0x41, 0xB7, 0x90, 0x6C, 0x28, 0x91, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8, 0xA8}

	emptyBuf = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
)

func TestScanBeast(t *testing.T) {
	type args struct {
		data  []byte
		atEOF bool
	}
	tests := []struct {
		name        string
		args        args
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}{
		{
			name:        "Test Not Enough",
			args:        args{data: []byte{0x1a, 0x33, 0x22, 0x1b, 0x54}, atEOF: true},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode AC",
			args:        args{data: append(beastModeAc, emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeAc),
			wantToken:   beastModeAc,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode AC, padded",
			args:        args{data: append(emptyBuf, append(beastModeAc, emptyBuf...)...), atEOF: true},
			wantAdvance: len(beastModeAc) + len(emptyBuf),
			wantToken:   beastModeAc,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode S Short",
			args:        args{data: append(beastModeSShort, emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSShort),
			wantToken:   beastModeSShort,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode S Long",
			args:        args{data: append(beastModeSLong, emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSLong),
			wantToken:   beastModeSLong,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode S Long Double Esc",
			args:        args{data: append(beastModeSLongDoubleEsc, emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSLongDoubleEsc),
			wantToken:   beastModeSLongDoubleRemoved,
			wantErr:     false,
		},
		{
			name:        "Test One Valid Mode S Long Double Esc Buffer Overrun",
			args:        args{data: beastModeSLongDoubleEsc[0:22], atEOF: true},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name:        "Test Two Valid Mode S Short",
			args:        args{data: append(append(beastModeSShort, beastModeSShort...), emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSShort),
			wantToken:   beastModeSShort,
			wantErr:     false,
		},
		{
			name:        "Test Two Valid Mode S Long",
			args:        args{data: append(append(beastModeSLong, beastModeSLong...), emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSLong),
			wantToken:   beastModeSLong,
			wantErr:     false,
		},
		{
			name:        "Test Two Valid Mode S Long Escaped",
			args:        args{data: append(append(beastModeSLongDoubleEsc, beastModeSLongDoubleEsc...), emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSLongDoubleEsc),
			wantToken:   beastModeSLongDoubleRemoved,
			wantErr:     false,
		},
		{
			name:        "Test Most of One Valid Mode S Short and One Valid Mode Short S",
			args:        args{data: append(append(beastModeSShort[3:], beastModeSShort...), emptyBuf...), atEOF: true},
			wantAdvance: len(beastModeSShort) + len(beastModeSShort[3:]),
			wantToken:   beastModeSShort,
			wantErr:     false,
		},
		{
			name:        "Test Overrun",
			args:        args{data: beastModeSShortBad, atEOF: true},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name:        "No Beast Esc",
			args:        args{data: noBeast, atEOF: true},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name:        "No Beast Esc",
			args:        args{data: noBeast, atEOF: false},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := ScanBeast()
			gotAdvance, gotToken, err := scanner(tt.args.data, tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanBeast() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("ScanBeast() gotAdvance = %v, want %v", gotAdvance, tt.wantAdvance)
				return
			}

			if gotToken == nil && tt.wantToken == nil {
				return
			}

			l := len(tt.wantToken)
			if !reflect.DeepEqual(gotToken[0:l], tt.wantToken) {
				t.Errorf("ScanBeast() gotToken (len %d) = %X, want (len %d) %X", len(gotToken), gotToken, len(tt.wantToken), tt.wantToken)
			}
		})
	}
}

func Test_producer_beastScanner(t *testing.T) {
	p := New(WithType(Beast))

	buf := beastModeSLong
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	expectedCounter := 0
	unexpectedCounter := 0
	var lock sync.Mutex

	go func() {
		for m := range p.out {
			if m.Type() == tracker.PlaneLocationEventType {
				lock.Lock()
				expectedCounter++
				lock.Unlock()
			} else {
				lock.Lock()
				unexpectedCounter++
				lock.Unlock()
			}
		}
	}()
	err := p.beastScanner(scanner)
	if nil != err {
		t.Errorf("Failed to scan single message")
	}
	close(p.out)
	time.Sleep(time.Millisecond * 50)
	lock.Lock()
	defer lock.Unlock()
	if expectedCounter != 1 {
		t.Errorf("Got the wrong message count. got %d, expected 1", expectedCounter)
	}
	if unexpectedCounter != 0 {
		t.Errorf("Got Unexpected Messaged. got %d, expected 1", expectedCounter)
	}
}

func TestScanBeastUnique(t *testing.T) {
	scanner := ScanBeast()

	buf := append(append(beastModeSShort, beastModeSLong...), emptyBuf...)

	// get the first message

	gotAdvance1, gotToken1, err1 := scanner(buf, true)
	if err1 != nil {
		t.Errorf("ScanBeast() error = %v", err1)
		return
	}
	if gotAdvance1 != len(beastModeSShort) {
		t.Errorf("ScanBeast() gotAdvance1 = %v, want %v", gotAdvance1, len(beastModeSShort))
	}
	if !reflect.DeepEqual(gotToken1[0:gotAdvance1], beastModeSShort) {
		t.Errorf("ScanBeast() gotToken1 (len %d) = %X, want (len %d) %X", len(gotToken1), gotToken1, len(beastModeSShort), beastModeSShort)
	}
	// get the second message

	gotAdvance2, gotToken2, err2 := scanner(buf[gotAdvance1:], true)
	if err1 != nil {
		t.Errorf("ScanBeast() error = %v", err2)
		return
	}
	if gotAdvance2 != len(beastModeSLong) {
		t.Errorf("ScanBeast() gotAdvance2 = %v, want %v", gotAdvance2, len(beastModeSLong))
	}
	if !reflect.DeepEqual(gotToken2[0:gotAdvance2], beastModeSLong) {
		t.Errorf("ScanBeast() gotToken1 (len %d) = %X, want (len %d) %X", len(gotToken2), gotToken2, len(beastModeSLong), beastModeSLong)
	}

	// and now make sure they are both different

	if reflect.DeepEqual(gotToken1[0:12], gotToken2[0:12]) {
		t.Errorf("Token1 and Token2 are the same, did you use the same base slice?")
	}
}

func BenchmarkScanBeast(b *testing.B) {
	f, err := os.Open("testdata/beast.sample")
	if nil != err {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		f.Seek(0, 0)
		scanner := bufio.NewScanner(f)
		scanner.Split(ScanBeast())
		for scanner.Scan() {
		}
	}
}
