package sha1

import (
	"log"
	"reflect"
	"testing"
)

//go:generate go run asm.go -out sha1.s -stubs stub.go

func TestEmptyString(t *testing.T) {
	h := [...]uint32{0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476, 0xc3d2e1f0}
	m := make([]byte, 64)
	m[0] = 0x80

	block(&h, m)

	expect := [...]uint32{0xda39a3ee, 0x5e6b4b0d, 0x3255bfef, 0x95601890, 0xafd80709}

	for i := 0; i < 5; i++ {
		log.Printf("h[%d] = %08x", i, h[i])
	}
	if !reflect.DeepEqual(expect, h) {
		t.Fatal("incorrect hash")
	}
}
