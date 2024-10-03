package gen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type optab struct {
	prnt.Generator

	cfg printer.Config

	table *Table
}

// NewOptab builds the operator table. This contains a more compact
// representation of the instruction database, containing the data needed for
// instruction builders to match against provided operands, and build the
// selected instruction.
func NewOptab(cfg printer.Config) Interface {
	return GoFmt(&optab{cfg: cfg})
}

func (t *optab) Generate(is []inst.Instruction) ([]byte, error) {
	t.Printf("// %s\n\n", t.cfg.GeneratedWarning())
	t.Printf("package x86\n\n")
	t.Printf("import (\n")
	t.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	t.Printf("\t%q\n", api.ImportPath(api.RegisterPackage))
	t.Printf(")\n\n")

	// Generate instruction data table.
	t.table = NewTable(is)

	// Size constants.
	t.maxOperands(is)

	// Types.
	t.operandTypeEnum(is)
	t.implicitRegisterEnum(is)
	t.enum(t.table.Suffix())
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

	t.Comment("maxoperands is the maximum number of operands in an instruction form, including implicit operands.")
	t.Printf("const maxoperands = %d\n\n", max)
}

func (t *optab) operandTypeEnum(is []inst.Instruction) {
	// Operand type enum.
	e := t.table.OperandType()
	t.enum(e)

	// Operand match function.
	types := inst.OperandTypes(is)
	t.Printf("func (%s %s) Match(op %s) bool {\n", e.Receiver(), e.Name(), api.OperandType)
	t.Printf("\tswitch %s {\n", e.Receiver())
	t.Printf("\t\tdefault: return false\n")
	for _, typ := range types {
		t.Printf("\t\tcase %s: return %s(op)\n", t.table.OperandTypeConst(typ), api.CheckerName(typ))
	}
	t.Printf("\t}\n")
	t.Printf("}\n\n")
}

func (t *optab) implicitRegisterEnum(is []inst.Instruction) {
	// Implicit register enum.
	e := t.table.ImplicitRegister()
	t.enum(e)

	// Register conversion function.
	registers := inst.ImplicitRegisters(is)
	t.Printf("func (%s %s) Register() %s {\n", e.Receiver(), e.Name(), api.RegisterType)
	t.Printf("\tswitch %s {\n", e.Receiver())
	t.Printf("\t\tdefault: panic(\"unexpected implicit register type\")\n")
	for _, r := range registers {
		t.Printf("\t\tcase %s: return %s\n", t.table.ImplicitRegisterConst(r), api.ImplicitRegister(r))
	}
	t.Printf("\t}\n")
	t.Printf("}\n\n")
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

	t.Comment("maxsuffixes is the maximum number of suffixes an instruction can have.")
	t.Printf("const maxsuffixes = %d\n\n", max)

	name := t.table.SuffixesTypeName()
	t.Printf("type %s [maxsuffixes]%s\n", name, t.table.Suffix().Name())

	// Conversion function to list of strings.
	mapname := name + "stringsmap"

	t.Printf("func (s %s) Strings() []string {\n", name)
	t.Printf("return %s[s]", mapname)
	t.Printf("}\n")

	var entries []string
	for _, class := range inst.SuffixesClasses(is) {
		for _, suffixes := range class {
			entry := fmt.Sprintf("%s: %s", t.table.SuffixesList(suffixes), stringsliteral(suffixes.Strings()))
			entries = append(entries, entry)
		}
	}

	t.Printf("var %s = map[%s][]string{\n", mapname, name)
	sort.Strings(entries)
	for _, entry := range entries {
		t.Printf("%s,\n", entry)
	}
	t.Printf("}\n")
}

func (t *optab) suffixesClassEnum(is []inst.Instruction) {
	// Suffixes class enum.
	e := t.table.SuffixesClass()
	t.enum(e)

	// Mapping method to the set of accepted suffixes.
	sets := map[string]string{}
	for key, class := range inst.SuffixesClasses(is) {
		var entries []string
		for _, suffixes := range class {
			entry := fmt.Sprintf("%s: true", t.table.SuffixesConst(suffixes))
			entries = append(entries, entry)
		}

		sort.Strings(entries)
		sets[api.SuffixesClassIdentifier(key)] = "{" + strings.Join(entries, ", ") + "}"
	}

	settype := fmt.Sprintf("map[%s]bool", t.table.SuffixesTypeName())
	t.mapping(e, "SuffixesSet", settype, "nil", sets)
}

func (t *optab) isasEnum(is []inst.Instruction) {
	// ISAs enum.
	e := t.table.ISAs()
	t.enum(e)

	// Mapping method to produce the list of ISAs.
	lists := map[string]string{}
	for _, isas := range inst.ISACombinations(is) {
		lists[api.ISAsIdentifier(isas)] = stringsliteral(isas)
	}
	t.mapping(e, "List", "[]string", "nil", lists)
}

func (t *optab) opcodeEnum(is []inst.Instruction) {
	e := t.table.Opcode()
	t.enum(e)
	t.stringmethod(e)
}

func (t *optab) forms(is []inst.Instruction) {
	// We require all instructions for a given opcode to be in a contiguous
	// block. This is likely true already but we'll make a sorted copy to ensure
	// the optab is robust to changes elsewhere.
	is = append([]inst.Instruction(nil), is...)
	sort.Slice(is, func(i, j int) bool {
		return is[i].Opcode < is[j].Opcode
	})

	// Output instruction forms table.
	table := "forms"
	t.Printf("var %s = []form{\n", table)
	for _, i := range is {
		for _, f := range i.Forms {
			t.Printf("{")

			// Basic properties.
			t.Printf("%s, ", t.table.OpcodeConst(i.Opcode))
			t.Printf("%s, ", t.table.SuffixesClassConst(f.SuffixesClass()))
			t.Printf("%s, ", Features(i, f))
			t.Printf("%s, ", t.table.ISAsConst(f.ISA))

			// Operands.
			t.Printf("%d, ", len(f.Operands))
			t.Printf("oprnds{")
			for _, op := range f.Operands {
				t.Printf(
					"{uint8(%s),false,%s},",
					t.table.OperandTypeConst(op.Type),
					Action(op.Action),
				)
			}
			for _, op := range f.ImplicitOperands {
				t.Printf(
					"{uint8(%s),true,%s},",
					t.table.ImplicitRegisterConst(op.Register),
					Action(op.Action),
				)
			}
			t.Printf("}")

			t.Printf("},\n")
		}
	}
	t.Printf("}\n\n")

	// Build mapping from opcode to corresponding forms.
	forms := map[string]string{}
	n := 0
	for _, i := range is {
		e := n + len(i.Forms)
		forms[i.Opcode] = fmt.Sprintf("%s[%d:%d]", table, n, e)
		n = e
	}

	t.mapping(t.table.Opcode(), "Forms", "[]form", "nil", forms)
}

func (t *optab) enum(e *Enum) {
	// Type declaration.
	t.Comment(e.Doc()...)
	t.Printf("type %s %s\n\n", e.Name(), e.UnderlyingType())

	// Supported values.
	t.Printf("const (\n")
	t.Printf("\t%s %s = iota\n", e.None(), e.name)
	for _, name := range e.ConstNames() {
		t.Printf("\t%s\n", name)
	}
	t.Printf("\t%s\n", e.MaxName())
	t.Printf(")\n\n")
}

func (t *optab) mapping(e *Enum, name, ret, zero string, to map[string]string) {
	table := strings.ToLower(e.Name() + name + "table")

	r := e.Receiver()
	t.Printf("func (%s %s) %s() %s {\n", r, e.Name(), name, ret)
	t.Printf("if %s < %s && %s < %s {\n", e.None(), r, r, e.MaxName())
	t.Printf("return %s[%s-1]\n", table, r)
	t.Printf("}\n")
	t.Printf("return %s\n", zero)
	t.Printf("}\n\n")

	t.Printf("var %s = []%s{\n", table, ret)
	for _, value := range e.Values() {
		t.Printf("\t%s,\n", to[value])
	}
	t.Printf("}\n\n")
}

func (t *optab) stringmethod(e *Enum) {
	s := map[string]string{}
	for _, value := range e.Values() {
		s[value] = strconv.Quote(value)
	}
	t.mapping(e, "String", "string", `""`, s)
}

func stringsliteral(ss []string) string {
	if ss == nil {
		return "nil"
	}
	var quoted []string
	for _, s := range ss {
		quoted = append(quoted, strconv.Quote(s))
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}
