package api

import (
	"testing"

	"github.com/mmcloughlin/avo/internal/inst"
)

func TestFunctionsUniqueArgNames(t *testing.T) {
	fns := InstructionsFunctions(inst.Instructions)
	for _, fn := range fns {
		s := fn.Signature()
		for _, n := range fn.Arities() {
			if n == 0 {
				continue
			}
			names := map[string]bool{}
			for j := 0; j < n; j++ {
				names[s.ParameterName(j)] = true
			}
			if len(names) != n {
				t.Errorf("repeated argument for instruction %s", fn.Name())
			}
			if _, found := names[""]; found {
				t.Errorf("empty argument name for instruction %s", fn.Name())
			}
		}
	}
}
