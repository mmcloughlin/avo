package gen

import (
	"fmt"
	"io"

	"github.com/mmcloughlin/avo/internal/inst"
)

type Interface interface {
	Generate(io.Writer, []*inst.Instruction) error
}

type Func func(io.Writer, []*inst.Instruction) error

func (f Func) Generate(w io.Writer, is []*inst.Instruction) error {
	return f(w, is)
}

type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	_, p.err = fmt.Fprintf(p.w, format, args...)
}

func (p *printer) Err() error {
	return p.err
}
