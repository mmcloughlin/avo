package api

import (
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/internal/inst"
)

func TestFunctionsDuplicateFormSignatures(t *testing.T) {
	// Test for duplicate form signatures within a given function. This could
	// manifest as duplicate case statements in generated code.
	fns := InstructionsFunctions(inst.Instructions)
	for _, fn := range fns {
		fn := fn // scopelint
		t.Run(fn.Name(), func(t *testing.T) {
			seen := map[string]bool{}
			for _, f := range fn.Forms {
				sig := strings.Join(f.Signature(), "_")
				t.Log(sig)
				if seen[sig] {
					t.Fatalf("duplicate: %s", sig)
				}
				seen[sig] = true
			}
		})
	}
}

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
