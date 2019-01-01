package textflag

import "testing"

//go:generate go run make_attrtest.go -output zattrtest.s -seed 42 -num 256

func AttrTest() bool

func TestAttributes(t *testing.T) {
	if !AttrTest() {
		t.Fail()
	}
}
