//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

var offQ Mem

func generateAddSub(useBase bool) {

	funcName := "AddSubPairs"
	if !useBase {
		funcName += "NoBase"
	}

	TEXT(funcName, 0, "func(a *[8]int64)")
	Doc("Destructively Add/Subtract successive pairs")
	offsets, evenVals, oddVals, result := YMM(), YMM(), YMM(), YMM()

	VMOVDQU64(offQ, offsets)
	var aMem Mem

	if useBase {
		aPtr := Load(Param("a"), GP64())
		aMem = Mem{Base: aPtr, Index: offsets, Scale: 8}
	} else {
		aPtrMem, err := Param("a").Resolve()
		if err != nil {
			panic(err)
		}
		VPSLLQ(Imm(3), offsets, offsets)
		VPADDQ_BCST(aPtrMem.Addr, offsets, offsets)
		aMem = Mem{Index: offsets, Scale: 1}
	}

	k := K()
	KXNORB(K0, K0, k)
	VPGATHERQQ(aMem, k, evenVals)
	KXNORB(K0, K0, k)
	VPGATHERQQ(aMem.Offset(8), k, oddVals)
	VPADDQ(oddVals, evenVals, result)
	KXNORB(K0, K0, k)
	VPSCATTERQQ(result, k, aMem)
	VPSUBQ(oddVals, evenVals, result)
	KXNORB(K0, K0, k)
	VPSCATTERQQ(result, k, aMem.Offset(8))
	RET()
}

func main() {

	offQ = GLOBL("offQ", RODATA|NOPTR)
	for i := 0; i < 4; i++ {
		DATA(i*8, U64(2*i))
	}

	generateAddSub(true)
	generateAddSub(false)

	Generate()
}
