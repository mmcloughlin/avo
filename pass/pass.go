package pass

import "github.com/mmcloughlin/avo"

// TODO(mbm): pass types

// FunctionPass builds a full pass that operates on all functions independently.
func FunctionPass(p func(*avo.Function) error) func(*avo.File) error {
	return func(f *avo.File) error {
		for _, fn := range f.Functions {
			if err := p(fn); err != nil {
				return err
			}
		}
		return nil
	}
}
