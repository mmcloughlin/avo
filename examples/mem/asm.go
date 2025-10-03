//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	TEXT("ByteFromSlice", NOSPLIT, "func(b []byte, offset int) byte")
	Doc("Takes a slice of bytes and an offset and returns the byte at that offset, performs no bounds checking - very unsafe")
	bytesBase := Load(Param("b").Base(), GP64())
	bytesOffset := Load(Param("offset"), GP64())

	bytesData := Mem{Base: bytesBase, Index: bytesOffset, Scale: 1}
	bytesByte := GP8()
	MOVB(bytesData, bytesByte)

	Store(bytesByte, ReturnIndex(0))
	RET()

	TEXT("ByteFromString", NOSPLIT, "func(s string, offset int) byte")
	Doc("Takes a string and an offset and returns the byte at that offset, performs no bounds checking - very unsafe")
	stringBase := Load(Param("s").Base(), GP64())
	stringOffset := Load(Param("offset"), GP64())

	stringData := Mem{Base: stringBase, Index: stringOffset, Scale: 1}
	stringByte := GP8()
	MOVB(stringData, stringByte)

	Store(stringByte, ReturnIndex(0))
	RET()

	TEXT("Int32FromSlice", NOSPLIT, "func(i []int32, offset int) int32")
	Doc("Takes a slice of int32 and an offset and returns the int32 at that offset, performs no bounds checking - very unsafe")
	int32Base := Load(Param("i").Base(), GP64())
	int32Offset := Load(Param("offset"), GP64())

	int32Data := Mem{Base: int32Base, Index: int32Offset, Scale: 4}
	int32Byte := GP32()
	MOVL(int32Data, int32Byte)

	Store(int32Byte, ReturnIndex(0))
	RET()

	TEXT("Int64FromSlice", NOSPLIT, "func(i []int64, offset int) int64")
	Doc("Takes a slice of int64 and an offset and returns the int64 at that offset, performs no bounds checking - very unsafe")
	int64Base := Load(Param("i").Base(), GP64())
	int64Offset := Load(Param("offset"), GP64())

	int64Data := Mem{Base: int64Base, Index: int64Offset, Scale: 8}
	int64Byte := GP64()
	MOVQ(int64Data, int64Byte)

	Store(int64Byte, ReturnIndex(0))
	RET()

	Generate()
}
