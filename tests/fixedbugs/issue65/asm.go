// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("accum", NOSPLIT, "func(acc, data, key *byte)")
	acc := Mem{Base: Load(Param("acc"), GP64())}
	data := Mem{Base: Load(Param("data"), GP64())}
	key := Mem{Base: Load(Param("key"), GP64())}

	for n := 0; n < 15; n++ {
		x0, x1 := XMM(), XMM()

		VMOVDQU(data, x0)
		VMOVDQU(key, x1)

		y0, y1, y2 := x0.AsY(), x1.AsY(), YMM()
		VINSERTI128(Imm(1), data.Offset(0x10), y0, y0)
		VINSERTI128(Imm(1), key.Offset(0x10), y1, y1)

		VPADDD(y0, y1, y1)
		VPSHUFD(Imm(0x31), y1, y2)
		VPADDQ(acc, y0, y0)
		VPMULUDQ(y2, y1, y1)
		VPADDQ(y1, y0, y0)

		VMOVDQA(y0, acc)
		// VZEROUPPER()

		data = data.Offset(64)
		key = key.Offset(8)
	}

	RET()

	Generate()
}
