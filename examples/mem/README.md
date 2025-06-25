# load

Demonstrates how to load data from a combination of pointers and offsets with `avo`.

The [code generator](asm.go) is as follows:

## Read from a []byte

This function takes a slice of bytes and an offset and returns the byte at that
offset from the start of the slice. You can use an instance of the
`operand.Mem` struct to create a pointer to a memory location. Note that the
`Base` is a register containing the address of the slice `b`'s data and the
`Index` is a register containing the offset value. We indicate the `Scale` here
to calculate the full size-in-bytes of the offset, since the data here is
single bytes the `Scale` is 1. Once this is done we use `MOVB` to read from the
memory location into an 8 bit register. We use `Store(...)` and
`ReturnIndex(0)` to write the return value.

[embedmd]:# (asm.go go /.*TEXT.*ByteFromSlice/ /RET.*/)
```go
        TEXT("ByteFromSlice", NOSPLIT, "func(b []byte, offset int) byte")
        bytesBase := Load(Param("b").Base(), GP64())
        bytesOffset := Load(Param("offset"), GP64())

        bytesData := Mem{Base: bytesBase, Index: bytesOffset, Scale: 1}
        bytesByte := GP8()
        MOVB(bytesData, bytesByte)

        Store(bytesByte, ReturnIndex(0))
        RET()
```

[embedmd]:# (mem.s)
```s
// func ByteFromSlice(b []byte, offset int) byte
TEXT ·ByteFromSlice(SB), NOSPLIT, $0-33
        MOVQ b_base+0(FP), AX
        MOVQ offset+24(FP), CX
        MOVB (AX)(CX*1), AL
        MOVB AL, ret+32(FP)
        RET
```

## Read from a string

This function takes a string and an offset and returns the byte at that offset
from the start of the string. We use an instance of `operand.Mem` just as
above. Because the string is interpreted as being made of bytes here the
`Scale` is still 1.

[embedmd]:# (asm.go go /.*TEXT.*ByteFromSlice/ /RET.*/)
```go
        TEXT("ByteFromString", NOSPLIT, "func(s string, offset int) byte")
        stringBase := Load(Param("s").Base(), GP64())
        stringOffset := Load(Param("offset"), GP64())

        stringData := Mem{Base: stringBase, Index: stringOffset, Scale: 1}
        stringByte := GP8()
        MOVB(stringData, stringByte)

        Store(stringByte, ReturnIndex(0))
        RET()
```

[embedmd]:# (mem.s)
```s
// func ByteFromString(s string, offset int) byte
TEXT ·ByteFromString(SB), NOSPLIT, $0-25
        MOVQ s_base+0(FP), AX
        MOVQ offset+16(FP), CX
        MOVB (AX)(CX*1), AL
        MOVB AL, ret+24(FP)
        RET
```

## Read from an int32 slice

This function takes a slice of int32 and an offset and returns the int32 at
that offset from the start of the slice. We use an instance of `operand.Mem`
just as above. Because the slice contains int32 sized data the `Scale` is 4, as
each int32 is 4 bytes wide.

[embedmd]:# (asm.go go /.*TEXT.*ByteFromSlice/ /RET.*/)
```go
        TEXT("Int32FromSlice", NOSPLIT, "func(i []int32, offset int) int32")
        int32Base := Load(Param("i").Base(), GP64())
        int32Offset := Load(Param("offset"), GP64())

        int32Data := Mem{Base: int32Base, Index: int32Offset, Scale: 4}
        int32Byte := GP32()
        MOVL(int32Data, int32Byte)

        Store(int32Byte, ReturnIndex(0))
        RET()
```

[embedmd]:# (mem.s)
```s
// func Int32FromSlice(i []int32, offset int) int32
TEXT ·Int32FromSlice(SB), NOSPLIT, $0-36
        MOVQ i_base+0(FP), AX
        MOVQ offset+24(FP), CX
        MOVL (AX)(CX*4), AX
        MOVL AX, ret+32(FP)
        RET
```

## Read from an int64 slice

This function takes a slice of int64 and an offset and returns the int64 at
that offset from the start of the slice. We use an instance of `operand.Mem`
just as above. Because the slice contains int64 sized data the `Scale` is 8, as
each int64 is 8 bytes wide.

[embedmd]:# (asm.go go /.*TEXT.*ByteFromSlice/ /RET.*/)
```go
        TEXT("Int64FromSlice", NOSPLIT, "func(i []int64, offset int) int64")
        int64Base := Load(Param("i").Base(), GP64())
        int64Offset := Load(Param("offset"), GP64())

        int64Data := Mem{Base: int64Base, Index: int64Offset, Scale: 4}
        int64Byte := GP64()
        MOVQ(int64Data, int64Byte)

        Store(int64Byte, ReturnIndex(0))
        RET()
```

[embedmd]:# (mem.s)
```s
// func Int64FromSlice(i []int64, offset int) int64
TEXT ·Int64FromSlice(SB), NOSPLIT, $0-40
        MOVQ i_base+0(FP), AX
        MOVQ offset+24(FP), CX
        MOVQ (AX)(CX*4), AX
        MOVQ AX, ret+32(FP)
        RET
```
