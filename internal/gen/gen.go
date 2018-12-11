package gen

import (
	"go/format"

	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/printer"
)

type Interface interface {
	Generate([]inst.Instruction) ([]byte, error)
}

type Func func([]inst.Instruction) ([]byte, error)

func (f Func) Generate(is []inst.Instruction) ([]byte, error) {
	return f(is)
}

type Builder func(printer.Config) Interface

// GoFmt formats Go code produced from the given generator.
func GoFmt(i Interface) Interface {
	return Func(func(is []inst.Instruction) ([]byte, error) {
		b, err := i.Generate(is)
		if err != nil {
			return nil, err
		}
		return format.Source(b)
	})
}
