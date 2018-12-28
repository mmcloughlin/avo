// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const (
	OffsetBasis = 0xcbf29ce484222325
	Prime       = 0x100000001b3
)

func main() {
	TEXT("Hash64", "func(data []byte) uint64")
	Doc("Hash64 computes the FNV-1a hash of data.")
	ptr := Load(Param("data").Base(), GP64v())
	n := Load(Param("data").Len(), GP64v())

	h := reg.RAX
	MOVQ(operand.Imm(OffsetBasis), h)
	p := GP64v()
	MOVQ(operand.Imm(Prime), p)

	LABEL("loop")
	CMPQ(n, operand.Imm(0))
	JE(operand.LabelRef("done"))
	b := GP64v()
	MOVBQZX(operand.Mem{Base: ptr}, b)
	XORQ(b, h)
	MULQ(p)
	INCQ(ptr)
	DECQ(n)

	JMP(operand.LabelRef("loop"))
	LABEL("done")
	Store(h, ReturnIndex(0))
	RET()
	Generate()
}
