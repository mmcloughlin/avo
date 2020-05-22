// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("Shuffle", NOSPLIT, "func(mask, data []byte) [4]uint32")
	Doc("Shuffle performs in-place shuffling of data according to a shuffle control mask.")
	maskptr := Mem{Base: Load(Param("mask").Base(), GP64())}
	dataptr := Mem{Base: Load(Param("data").Base(), GP64())}
	data := XMM()
	MOVOU(dataptr, data)
	PSHUFB(maskptr, data)
	MOVOU(data, ReturnIndex(0).MustAddr())
	RET()
	Generate()
}
