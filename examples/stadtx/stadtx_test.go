package stadtx

import (
	"testing"
	"testing/quick"

	ref "github.com/dgryski/go-stadtx"
)

//go:generate go run asm.go -out stadtx.s -stubs stub.go

func IUT(s State, key []byte) uint64 {
	return Hash(&s, key)
}

func Expect(s State, key []byte) uint64 {
	t := ref.State(s)
	return ref.Hash(&t, key)
}

func TestCmp(t *testing.T) {
	if err := quick.CheckEqual(IUT, Expect, nil); err != nil {
		t.Fatal(err)
	}
}
