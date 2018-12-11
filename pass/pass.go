package pass

import (
	"io"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/printer"
)

var Compile = Concat(
	FunctionPass(LabelTarget),
	FunctionPass(CFG),
	FunctionPass(Liveness),
	FunctionPass(AllocateRegisters),
	FunctionPass(BindRegisters),
	FunctionPass(VerifyAllocation),
)

type Interface interface {
	Execute(*avo.File) error
}

type Func func(*avo.File) error

func (p Func) Execute(f *avo.File) error {
	return p(f)
}

type FunctionPass func(*avo.Function) error

func (p FunctionPass) Execute(f *avo.File) error {
	for _, fn := range f.Functions {
		if err := p(fn); err != nil {
			return err
		}
	}
	return nil
}

func Concat(passes ...Interface) Interface {
	return Func(func(f *avo.File) error {
		for _, p := range passes {
			if err := p.Execute(f); err != nil {
				return err
			}
		}
		return nil
	})
}

type Output struct {
	Writer  io.WriteCloser
	Printer printer.Printer
}

func (o *Output) Execute(f *avo.File) error {
	b, err := o.Printer.Print(f)
	if err != nil {
		return err
	}
	if _, err = o.Writer.Write(b); err != nil {
		return err
	}
	return o.Writer.Close()
}
