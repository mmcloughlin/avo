// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("Issue65", NOSPLIT, "func()")
	VINSERTI128(Imm(1), Y0.AsX(), Y1, Y2)
	RET()
	Generate()
}
