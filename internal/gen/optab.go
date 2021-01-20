package gen

import (
	"fmt"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type optab struct {
	prnt.Generator

	cfg printer.Config

	operandTypes      *enum
	implicitRegisters *enum
	isas              *enum
	opcodes           *enum
}

func NewOptab(cfg printer.Config) Interface {
	return GoFmt(&optab{cfg: cfg})
}

func (t *optab) Generate(is []inst.Instruction) ([]byte, error) {
	t.Printf("// %s\n\n", t.cfg.GeneratedWarning())
	t.Printf("package x86\n\n")
	t.Printf("import (\n")
	t.Printf("\t%q\n", api.ImportPath("operand"))
	t.Printf(")\n\n")

	// Size constants.
	t.maxOperands(is)

	// Enums.
	t.operandTypesEnum(is)
	t.implicitRegistersEnum(is)
	t.suffixesEnum(is)
	t.isasEnum(is)
	t.opcodesEnum(is)

	// Forms table.
	t.forms(is)

	return t.Result()
}

func (t *optab) maxOperands(is []inst.Instruction) {
	max := 0
	for _, i := range inst.Instructions {
		for _, f := range i.Forms {
			a := len(f.Operands) + len(f.ImplicitOperands)
			if a > max {
				max = a
			}
		}
	}

	t.Comment("MaxOperands is the maximum number of operands in an instruction form, including implicit operands.")
	t.Printf("const MaxOperands = %d\n\n", max)
}

func (t *optab) operandTypesEnum(is []inst.Instruction) {
	types := inst.OperandTypes(is)

	// Operand type enum.
	e := &enum{name: "OperandType"}
	for _, t := range types {
		e.values = append(e.values, api.OperandTypeIdentifier(t))
	}
	e.Print(&t.Generator)

	// Operand match function.
	t.Printf("func (%s %s) Match(op %s) bool {\n", e.Receiver(), e.Name(), api.OperandType)
	t.Printf("\tswitch %s {\n", e.Receiver())
	t.Printf("\t\tdefault: return false\n")
	for _, typ := range types {
		t.Printf("\t\tcase %s: return %s(op)\n", e.ConstName(api.OperandTypeIdentifier(typ)), api.CheckerName(typ))
	}
	t.Printf("\t}\n")
	t.Printf("}\n\n")

	t.operandTypes = e
}

func (t *optab) implicitRegistersEnum(is []inst.Instruction) {
	e := &enum{name: "ImplicitRegister"}
	for _, r := range inst.ImplicitRegisters(is) {
		e.values = append(e.values, api.ImplicitRegisterIdentifier(r))
	}
	e.Print(&t.Generator)
	t.implicitRegisters = e
}

func (t *optab) isasEnum(is []inst.Instruction) {
	combinations := inst.ISACombinations(is)

	// Enum.
	e := &enum{name: "ISAs"}
	for _, isas := range combinations {
		e.values = append(e.values, api.ISAsIdentifier(isas))
	}

	e.Print(&t.Generator)

	// Mapping method to produce the list of ISAs.
	lists := make([]string, len(combinations))
	for i, isas := range combinations {
		lists[i] = fmt.Sprintf("%#v", isas)
	}
	e.MapMethod(&t.Generator, "List", "[]string", "nil", lists)

	t.isas = e
}

func (t *optab) suffixesEnum(is []inst.Instruction) {
	e := &enum{name: "Suffix"}
	for _, s := range inst.UniqueSuffixes(is) {
		e.values = append(e.values, s.String())
	}
	e.Print(&t.Generator)
}

func (t *optab) opcodesEnum(is []inst.Instruction) {
	e := &enum{name: "Opcode"}
	for _, i := range is {
		e.values = append(e.values, i.Opcode)
	}
	e.Print(&t.Generator)
	e.StringMethod(&t.Generator)
	t.opcodes = e
}

func (t *optab) forms(is []inst.Instruction) {
	t.Printf("var forms = []Form{\n")
	for _, i := range is {
		for _, f := range i.Forms {
			t.Printf("{")

			// Basic properties.
			t.Printf("%s, ", t.opcodes.ConstName(i.Opcode))
			t.Printf("%s, ", features(i, f))
			t.Printf("%s, ", t.isas.ConstName(api.ISAsIdentifier(f.ISA)))

			// Operands.
			t.Printf("%d, ", len(f.Operands))
			t.Printf("Operands{")
			for _, op := range f.Operands {
				t.Printf(
					"{uint8(%s),false,%s},",
					t.operandTypes.ConstName(api.OperandTypeIdentifier(op.Type)),
					action(op.Action),
				)
			}
			for _, op := range f.ImplicitOperands {
				t.Printf(
					"{uint8(%s),true,%s},",
					t.implicitRegisters.ConstName(api.ImplicitRegisterIdentifier(op.Register)),
					action(op.Action),
				)
			}
			t.Printf("}")

			t.Printf("},\n")
		}
	}
	t.Printf("}\n\n")
}

func features(i inst.Instruction, f inst.Form) string {
	var enabled []string
	for _, feature := range []struct {
		Name    string
		Enabled bool
	}{
		{"Terminal", i.IsTerminal()},
		{"Branch", i.IsBranch()},
		{"ConditionalBranch", i.IsConditionalBranch()},
		{"CancellingInputs", f.CancellingInputs},
	} {
		if feature.Enabled {
			enabled = append(enabled, "Feature"+feature.Name)
		}
	}

	if len(enabled) == 0 {
		return "0"
	}
	return strings.Join(enabled, "|")
}

func action(a inst.Action) string {
	c := strings.ToUpper(a.String())
	if c == "" {
		c = "None"
	}
	return "Action" + c
}
