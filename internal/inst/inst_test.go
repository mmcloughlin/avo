package inst_test

import (
	"testing"

	"github.com/mmcloughlin/avo/internal/gen"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/test"
)

func TestHaveInstructions(t *testing.T) {
	n := len(inst.Instructions)
	t.Logf("number of instructions = %d", n)
	if n == 0 {
		t.Fatalf("no instructions")
	}
}

func TestOpcodeDupes(t *testing.T) {
	count := map[string]int{}
	for _, i := range inst.Instructions {
		count[i.Opcode]++
	}

	for opcode, n := range count {
		if n > 1 {
			t.Errorf("opcode %s appears %d times", opcode, n)
		}
	}
}

func TestInstructionProperties(t *testing.T) {
	for _, i := range inst.Instructions {
		if len(i.Opcode) == 0 {
			t.Errorf("empty opcode")
		}
		if len(i.Forms) == 0 {
			t.Errorf("instruction %s has no forms", i.Opcode)
		}
	}

}

func TestAssembles(t *testing.T) {
	g := gen.NewAsmTest(gen.Config{})
	b, err := g.Generate(inst.Instructions)
	if err != nil {
		t.Fatal(err)
	}
	test.Assembles(t, b)
}
