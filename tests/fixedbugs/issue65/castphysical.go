// +build generate

//go:generate go run $GOFILE

// Regression test for a bug where casting a physical register would give the
// error "non physical register found".
//
// See: https://github.com/mmcloughlin/avo/issues/65#issuecomment-576850145

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
