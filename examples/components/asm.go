// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	Package("github.com/mmcloughlin/avo/examples/components")

	TEXT("StringLen", "func(s string) int")
	strlen := Load(Param("s").Len(), GP64v())
	Store(strlen, ReturnIndex(0))
	RET()

	TEXT("SliceLen", "func(s []int) int")
	slicelen := Load(Param("s").Len(), GP64v())
	Store(slicelen, ReturnIndex(0))
	RET()

	TEXT("SliceCap", "func(s []int) int")
	slicecap := Load(Param("s").Cap(), GP64v())
	Store(slicecap, ReturnIndex(0))
	RET()

	TEXT("ArrayThree", "func(a [7]uint64) uint64")
	a3 := Load(Param("a").Index(3), GP64v())
	Store(a3, ReturnIndex(0))
	RET()

	TEXT("FieldByte", "func(s Struct) byte")
	b := Load(Param("s").Field("Byte"), GP8v())
	Store(b, ReturnIndex(0))
	RET()

	TEXT("FieldInt8", "func(s Struct) int8")
	i8 := Load(Param("s").Field("Int8"), GP8v())
	Store(i8, ReturnIndex(0))
	RET()

	TEXT("FieldUint16", "func(s Struct) uint16")
	u16 := Load(Param("s").Field("Uint16"), GP16v())
	Store(u16, ReturnIndex(0))
	RET()

	TEXT("FieldInt32", "func(s Struct) int32")
	i32 := Load(Param("s").Field("Int32"), GP32v())
	Store(i32, ReturnIndex(0))
	RET()

	TEXT("FieldUint64", "func(s Struct) uint64")
	u64 := Load(Param("s").Field("Uint64"), GP64v())
	Store(u64, ReturnIndex(0))
	RET()

	TEXT("FieldFloat32", "func(s Struct) float32")
	f32 := Load(Param("s").Field("Float32"), Xv())
	Store(f32, ReturnIndex(0))
	RET()

	TEXT("FieldFloat64", "func(s Struct) float64")
	f64 := Load(Param("s").Field("Float64"), Xv())
	Store(f64, ReturnIndex(0))
	RET()

	TEXT("FieldStringLen", "func(s Struct) int")
	l := Load(Param("s").Field("String").Len(), GP64v())
	Store(l, ReturnIndex(0))
	RET()

	TEXT("FieldSliceCap", "func(s Struct) int")
	c := Load(Param("s").Field("Slice").Cap(), GP64v())
	Store(c, ReturnIndex(0))
	RET()

	TEXT("FieldArrayTwoBTwo", "func(s Struct) byte")
	b2 := Load(Param("s").Field("Array").Index(2).Field("B").Index(2), GP8v())
	Store(b2, ReturnIndex(0))
	RET()

	TEXT("FieldArrayOneC", "func(s Struct) uint16")
	c1 := Load(Param("s").Field("Array").Index(1).Field("C"), GP16v())
	Store(c1, ReturnIndex(0))
	RET()

	TEXT("FieldComplex64Imag", "func(s Struct) float32")
	c64i := Load(Param("s").Field("Complex64").Imag(), Xv())
	Store(c64i, ReturnIndex(0))
	RET()

	TEXT("FieldComplex128Real", "func(s Struct) float64")
	c128r := Load(Param("s").Field("Complex128").Real(), Xv())
	Store(c128r, ReturnIndex(0))
	RET()

	Generate()
}
