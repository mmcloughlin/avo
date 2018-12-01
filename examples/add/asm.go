package main

import (
	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/reg"
)

func main() {
	TEXT("add")
	ADDQ(reg.R8, reg.R11)
	RET()
	EOF()
}
