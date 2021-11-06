package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type optab struct {
	prnt.Generator

	cfg printer.Config

	operandType      *enum
	implicitRegister *enum
	suffix           *enum
	suffixesClass    *enum
	isas             *enum
	opcode           *enum
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

	// Types.
	t.operandTypeEnum(is)
	t.implicitRegisterEnum(is)
	t.suffixEnum(is)
	t.suffixesType(is)
	t.suffixesClassEnum(is)
	t.isasEnum(is)
	t.opcodeEnum(is)

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

func (t *optab) operandTypeEnum(is []inst.Instruction) {
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

	t.operandType = e
}

func (t *optab) implicitRegisterEnum(is []inst.Instruction) {
	e := &enum{name: "ImplicitRegister"}
	for _, r := range inst.ImplicitRegisters(is) {
		e.values = append(e.values, api.ImplicitRegisterIdentifier(r))
	}
	e.Print(&t.Generator)
	t.implicitRegister = e
}

func (t *optab) suffixEnum(is []inst.Instruction) {
	e := &enum{name: "Suffix"}
	for _, s := range inst.UniqueSuffixes(is) {
		e.values = append(e.values, s.String())
	}
	e.Print(&t.Generator)

	t.suffix = e
}

func (t *optab) suffixesType(is []inst.Instruction) {
	// Declare the type as an array. This requires us to know the maximum number
	// of suffixes an instruction can have.
	max := 0
	for _, class := range inst.SuffixesClasses(is) {
		for _, suffixes := range class {
			if len(suffixes) > max {
				max = len(suffixes)
			}
		}
	}

	t.Comment("MaxSuffixes is the maximum number of suffixes an instruction can have.")
	t.Printf("const MaxSuffixes = %d\n\n", max)

	t.Printf("type Suffixes [MaxSuffixes]Suffix\n")

	// Conversion function to list of strings.
	mapname := "suffixesstringsmap"

	t.Printf("func (s Suffixes) Strings() []string {\n")
	t.Printf("return %s[s]", mapname)
	t.Printf("}\n")

	var entries []string
	for _, class := range inst.SuffixesClasses(is) {
		for _, suffixes := range class {
			entry := fmt.Sprintf("%s: %#v", t.suffixesConst(suffixes), suffixes.Strings())
			entries = append(entries, entry)
		}

	}

	t.Printf("var %s = map[Suffixes][]string{\n", mapname)
	sort.Strings(entries)
	for _, entry := range entries {
		t.Printf("%s,\n", entry)
	}
	t.Printf("}\n")
}

func (t *optab) suffixesClassEnum(is []inst.Instruction) {
	// Gather suffixes classes.
	classes := inst.SuffixesClasses(is)
	keys := make([]string, 0, len(classes))
	for key := range classes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build enum.
	e := &enum{name: "SuffixesClass"}
	for _, key := range keys {
		e.values = append(e.values, api.SuffixesClassIdentifier(key))
	}
	e.Print(&t.Generator)

	// Mapping method to the set of accepted suffixes.
	sets := make([]string, 0, len(classes))
	for _, key := range keys {
		var entries []string
		for _, suffixes := range classes[key] {
			entry := fmt.Sprintf("%s: true", t.suffixesConst(suffixes))
			entries = append(entries, entry)
		}

		sort.Strings(entries)
		set := "{" + strings.Join(entries, ", ") + "}"
		sets = append(sets, set)
	}

	e.MapMethod(&t.Generator, "SuffixesSet", "map[Suffixes]bool", "nil", sets)

	t.suffixesClass = e
}

func (t *optab) suffixesConst(suffixes inst.Suffixes) string {
	var parts []string
	for _, suffix := range suffixes {
		parts = append(parts, t.suffix.ConstName(suffix.String()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
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

func (t *optab) opcodeEnum(is []inst.Instruction) {
	e := &enum{name: "Opcode"}
	for _, i := range is {
		e.values = append(e.values, i.Opcode)
	}
	e.Print(&t.Generator)
	e.StringMethod(&t.Generator)
	t.opcode = e
}

func (t *optab) forms(is []inst.Instruction) {
	t.Printf("var forms = []Form{\n")
	for _, i := range is {
		for _, f := range i.Forms {
			t.Printf("{")

			// Basic properties.
			t.Printf("%s, ", t.opcode.ConstName(i.Opcode))
			t.Printf("%s, ", t.suffixesClassConst(f))
			t.Printf("%s, ", features(i, f))
			t.Printf("%s, ", t.isas.ConstName(api.ISAsIdentifier(f.ISA)))

			// Operands.
			t.Printf("%d, ", len(f.Operands))
			t.Printf("Operands{")
			for _, op := range f.Operands {
				t.Printf(
					"{uint8(%s),false,%s},",
					t.operandType.ConstName(api.OperandTypeIdentifier(op.Type)),
					action(op.Action),
				)
			}
			for _, op := range f.ImplicitOperands {
				t.Printf(
					"{uint8(%s),true,%s},",
					t.implicitRegister.ConstName(api.ImplicitRegisterIdentifier(op.Register)),
					action(op.Action),
				)
			}
			t.Printf("}")

			t.Printf("},\n")
		}
	}
	t.Printf("}\n\n")
}

func (t *optab) suffixesClassConst(f inst.Form) string {
	ident := api.SuffixesClassIdentifier(f.SuffixesClass())
	if ident == "" {
		return t.suffixesClass.None()
	}
	return t.suffixesClass.ConstName(ident)
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
