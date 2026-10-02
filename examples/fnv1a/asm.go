//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const (
	OffsetBasis = 0xcbf29ce484222325
	Prime       = 0x100000001b3
)

func main() {
	TEXT("Hash64", NOSPLIT, "func(data []byte) uint64")
	Doc("Hash64 computes the FNV-1a hash of data.")
	ptr := Load(Param("data").Base(), GP64())
	endPtr := Load(Param("data").Len(), GP64())
	ADDQ(ptr, endPtr)

	h := RAX
	MOVQ(Imm(OffsetBasis), h)
	p := GP64()
	MOVQ(Imm(Prime), p)

	Label("loop")
	CMPQ(ptr, endPtr)
	JE(LabelRef("done"))
	b := GP64()
	MOVBQZX(Mem{Base: ptr}, b)
	XORQ(b, h)
	MULQ(p)
	INCQ(ptr)

	JMP(LabelRef("loop"))
	Label("done")
	Store(h, ReturnIndex(0))
	RET()
	Generate()
}
