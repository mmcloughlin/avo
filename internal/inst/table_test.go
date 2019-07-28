package inst_test

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/internal/gen"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/test"
	"github.com/mmcloughlin/avo/printer"
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

func TestFormDupes(t *testing.T) {
	for _, i := range inst.Instructions {
		if HasFormDupe(i) {
			t.Errorf("%s has duplicate forms", i.Opcode)
		}
	}
}

func HasFormDupe(i inst.Instruction) bool {
	n := len(i.Forms)
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			if reflect.DeepEqual(i.Forms[a], i.Forms[b]) {
				return true
			}
		}
	}
	return false
}

func TestInstructionProperties(t *testing.T) {
	for _, i := range inst.Instructions {
		if len(i.Opcode) == 0 {
			t.Errorf("empty opcode")
		}
		if len(i.Forms) == 0 {
			t.Errorf("instruction %s has no forms", i.Opcode)
		}
		if len(i.Arities()) == 0 {
			t.Errorf("instruction %s has empty arities list", i.Opcode)
		}
		if i.IsNiladic() && len(i.Forms) != 1 {
			t.Errorf("%s breaks our expectation that niladic functions have one form", i.Opcode)
		}
	}
}

func TestAssembles(t *testing.T) {
	g := gen.NewAsmTest(printer.NewArgvConfig())
	b, err := g.Generate(inst.Instructions)
	if err != nil {
		t.Fatal(err)
	}
	test.Assembles(t, b)
}

func TestLookup(t *testing.T) {
	if _, found := inst.Lookup("CPUID"); !found {
		t.Fatalf("missing CPUID")
	}
	if _, found := inst.Lookup(strings.Repeat("XXX", 13)); found {
		t.Fatalf("lookup returns true on an absurd opcode")
	}
}

func TestInstructionArities(t *testing.T) {
	cases := map[string][]int{
		"AESDEC":    {2},
		"EXTRACTPS": {3},
		"SHRQ":      {2, 3},
		"VMOVHPD":   {2, 3},
	}
	for opcode, expect := range cases {
		i, ok := inst.Lookup(opcode)
		if !ok {
			t.Fatalf("could not find %s", opcode)
		}
		got := i.Arities()
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("arity of %s is %v expected %v", opcode, got, expect)
		}
	}
}

func TestStdLibOpcodes(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/stdlibopcodes.txt")
	if err != nil {
		t.Fatal(err)
	}
	opcodes := strings.Fields(string(b))
	for _, opcode := range opcodes {
		if _, found := inst.Lookup(opcode); !found {
			t.Errorf("missing instruction %s (used in stdlib asm)", opcode)
		}
	}
}

func TestCancellingInputs(t *testing.T) {
	// Expect all instruction forms with cancelling inputs to have two input operands of register type.
	//
	// Reference: https://github.com/Maratyszcza/PeachPy/blob/01d15157a973a4ae16b8046313ddab371ea582db/peachpy/x86_64/instructions.py#L136-L138
	//
	//	            assert len(input_operands) == 2, "Instruction forms with cancelling inputs must have two inputs"
	//	            assert all(map(lambda op: isinstance(op, Register), input_operands)), \
	//	                "Both inputs of instruction form with cancelling inputs must be registers"
	//
	for _, i := range inst.Instructions {
		for _, f := range i.Forms {
			if !f.CancellingInputs {
				continue
			}

			if len(f.ImplicitOperands) > 0 {
				t.Errorf("%s: expected no implicit operands", i.Opcode)
			}

			// Expect two register inputs.
			n := 0
			for _, op := range f.Operands {
				if op.Action.Read() {
					n++
					switch op.Type {
					case "r8", "r16", "r32", "r64", "xmm", "ymm":
						// pass
					default:
						t.Errorf("%s: unexpected operand type %q for self-cancelling input", i.Opcode, op.Type)
					}
				}
			}
			if n != 2 {
				t.Errorf("%s: expected two inputs for self-cancelling form", i.Opcode)
			}
		}
	}
}
