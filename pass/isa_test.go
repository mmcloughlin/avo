package pass

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/ir"
)

func TestRequiredISAExtensions(t *testing.T) {
	f := ir.NewFunction("ISAs")
	f.AddInstruction(&ir.Instruction{ISA: nil})
	f.AddInstruction(&ir.Instruction{ISA: []string{"B", "A"}})
	f.AddInstruction(&ir.Instruction{ISA: []string{"A", "C"}})

	err := RequiredISAExtensions(f)
	if err != nil {
		t.Fatal(err)
	}

	expect := []string{"A", "B", "C"}
	if !reflect.DeepEqual(f.ISA, expect) {
		t.FailNow()
	}
}
