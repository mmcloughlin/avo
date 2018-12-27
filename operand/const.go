package operand

import "fmt"

type Constant interface {
	Op
	Bytes() int
	constant()
}

//go:generate go run make_const.go -output zconst.go

// String is a string constant.
type String string

func (s String) Asm() string { return fmt.Sprintf("$%q", s) }
func (s String) Bytes() int  { return len(s) }
func (s String) constant()   {}

// Imm returns an unsigned integer constant with size guessed from x.
func Imm(x uint64) Constant {
	// TODO(mbm): remove this function
	switch {
	case uint64(uint8(x)) == x:
		return U8(x)
	case uint64(uint16(x)) == x:
		return U16(x)
	case uint64(uint32(x)) == x:
		return U32(x)
	}
	return U64(x)
}
