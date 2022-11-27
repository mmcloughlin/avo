//go:build ignore
// +build ignore

package main

import (
	"fmt"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	variants := []struct {
		Size     int
		Register func() GPVirtual
		XOR      func(Op, Op)
	}{
		{8, GP8L, XORB},
		{16, GP16, XORW},
		{32, GP32, XORL},
		{64, GP64, XORQ},
	}

	for _, v := range variants {
		name := fmt.Sprintf("Not%d", v.Size)
		TEXT(name, NOSPLIT, "func(x bool) bool")
		Doc(fmt.Sprintf("%s returns the boolean negation of x using a %d-bit intermediate.", name, v.Size))
		x := v.Register()
		Load(Param("x"), x)
		v.XOR(U8(1), x)
		Store(x.As8L(), ReturnIndex(0))
		RET()
	}

	Generate()
}
