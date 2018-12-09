package pass

import "github.com/mmcloughlin/avo"

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
