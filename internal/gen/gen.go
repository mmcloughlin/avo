package gen

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/mmcloughlin/avo/internal/inst"
)

type Interface interface {
	Generate([]*inst.Instruction) ([]byte, error)
}

type Func func([]*inst.Instruction) ([]byte, error)

func (f Func) Generate(is []*inst.Instruction) ([]byte, error) {
	return f(is)
}

// GoFmt formats Go code produced from the given generator.
func GoFmt(i Interface) Interface {
	return Func(func(is []*inst.Instruction) ([]byte, error) {
		b, err := i.Generate(is)
		if err != nil {
			return nil, err
		}
		return format.Source(b)
	})
}

type printer struct {
	buf bytes.Buffer
	err error
}

func (p *printer) Printf(format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(&p.buf, format, args...)
}

func (p *printer) Result() ([]byte, error) {
	return p.buf.Bytes(), p.err
}
