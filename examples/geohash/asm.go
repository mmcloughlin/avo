// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("EncodeInt", "func(lat, lng float64) uint64")
	Doc("EncodeInt computes the 64-bit integer geohash of (lat, lng).")
	lat := Load(Param("lat"), Xv())
	lng := Load(Param("lng"), Xv())

	MULSD(ConstData("reciprocal180", F64(1/180.0)), lat)
	onepointfive := ConstData("onepointfive", F64(1.5))
	ADDSD(onepointfive, lat)

	MULSD(ConstData("reciprocal360", F64(1/360.0)), lng)
	ADDSD(onepointfive, lng)

	lngi, lati := GP64v(), GP64v()
	MOVQ(lat, lati)
	SHRQ(U8(20), lati)
	MOVQ(lng, lngi)
	SHRQ(U8(20), lngi)

	mask := ConstData("mask", U64(0x5555555555555555))
	ghsh := GP64v()
	PDEPQ(mask, lati, ghsh)
	temp := GP64v()
	PDEPQ(mask, lngi, temp)
	SHLQ(U8(1), temp)
	XORQ(temp, ghsh)

	Store(ghsh, ReturnIndex(0))
	RET()

	Generate()
}
