package add

import (
	"testing"
	"testing/quick"
)

//go:generate go run asm.go -out add.s -stubs stub.go

func TestAdd(t *testing.T) {
	quick.CheckEqual(Add, func(x, y uint64) uint64 { return x + y }, nil)
}
