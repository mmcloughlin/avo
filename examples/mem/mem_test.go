package mem

import (
	"testing"
)

//go:generate go run asm.go -out mem.s -stubs stub.go

func TestByteFromSlice(t *testing.T) {
	slice := []byte{1, 2, 3, 4}
	for i := range slice {
		val := slice[i]
		asmVal := ByteFromSlice(slice, i)
		if val != asmVal {
			t.Fatalf("ByteFromSlice[%d] got %d, should be %d", i, asmVal, val)
		}
	}
}

func TestByteFromString(t *testing.T) {
	str := "abcdefg"
	for i := range str {
		val := str[i]
		asmVal := ByteFromString(str, i)
		if val != asmVal {
			t.Fatalf("ByteFromString[%d] got %d, should be %d", i, asmVal, val)
		}
	}
}

func TestFromInt32(t *testing.T) {
	slice := []int32{1, 2, 3, 4}
	for i := range slice {
		val := slice[i]
		asmVal := Int32FromSlice(slice, i)
		if val != asmVal {
			t.Fatalf("ByteFromString[%d] got %d, should be %d", i, asmVal, val)
		}
	}
}

func TestFromInt64(t *testing.T) {
	slice := []int64{1, 2, 3, 4}
	for i := range slice {
		val := slice[i]
		asmVal := Int64FromSlice(slice, i)
		if val != asmVal {
			t.Fatalf("ByteFromString[%d] got %d, should be %d", i, asmVal, val)
		}
	}
}
