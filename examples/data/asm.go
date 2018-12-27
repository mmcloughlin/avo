// +build ignore

package main

import (
	"math"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	bytes := GLOBL("bytes")
	DATA(0, U64(0x0011223344556677))
	DATA(8, String("strconst"))
	DATA(16, F32(math.Pi))
	DATA(24, F64(math.Pi))
	DATA(32, U32(0x00112233))
	DATA(36, U16(0x4455))
	DATA(38, U8(0x66))
	DATA(39, U8(0x77))

	// TEXT ·DataAt(SB),0,$0-9
	TEXT("DataAt", "func(i int) byte")
	i := Load(Param("i"), GP64v()) // MOVQ	i+0(FP), AX
	ptr := Mem{Base: GP64v()}
	LEAQ(bytes, ptr.Base) // LEAQ	b<>+0x00(SB), BX
	b := GP8v()
	MOVB(ptr.Idx(i, 1), b)   // MOVB	0(BX)(AX*1), CL
	Store(b, ReturnIndex(0)) // MOVB	CL, ret+8(FP)
	RET()                    // RET

	Generate()
}
