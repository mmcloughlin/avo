package gen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type ctors struct {
	cfg printer.Config
	prnt.Generator
}

// NewCtors will build instruction constructors. Each constructor will check
// that the provided operands match one of the allowed instruction forms. If so
// it will return an Instruction object that can be added to an avo Function.
func NewCtors(cfg printer.Config) Interface {
	return GoFmt(&ctors{cfg: cfg})
}

func (c *ctors) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"errors\"\n")
	c.NL()
	c.Printf("\tintrep \"%s/ir\"\n", api.Package)
	c.Printf("\t\"%s/reg\"\n", api.Package)
	c.Printf("\t\"%s/operand\"\n", api.Package)
	c.Printf(")\n\n")

	fns := api.InstructionsFunctions(is)
	for _, fn := range fns {
		c.function(fn)
	}

	return c.Result()
}

func (c *ctors) function(fn *api.Function) {
	c.Comment(fn.Doc()...)

	s := fn.Signature()

	c.Printf("func %s(%s) (*intrep.Instruction, error) {\n", fn.Name(), s.ParameterList())
	c.forms(fn, s)
	c.Printf("}\n\n")
}

func (c *ctors) forms(fn *api.Function, s api.Signature) {
	if fn.IsNiladic() {
		if len(fn.Forms) != 1 {
			c.AddError(fmt.Errorf("%s breaks assumption that niladic instructions have one form", fn.Name()))
		}
		c.Printf("return &%s, nil\n", construct(fn, fn.Forms[0], s))
		return
	}

	// Generate switch cases.
	type Case struct {
		Conditions  []string
		Instruction string
	}
	groups := map[string]*Case{}

	for _, f := range fn.Forms {
		var conds []string

		if fn.IsVariadic() {
			checklen := fmt.Sprintf("%s == %d", s.Length(), len(f.Operands))
			conds = append(conds, checklen)
		}

		for j, op := range f.Operands {
			checktype := fmt.Sprintf("%s(%s)", api.CheckerName(op.Type), s.ParameterName(j))
			conds = append(conds, checktype)
		}

		cond := strings.Join(conds, " && ")
		instruction := construct(fn, f, s)

		if _, ok := groups[instruction]; !ok {
			groups[instruction] = &Case{Instruction: instruction}
		}
		groups[instruction].Conditions = append(groups[instruction].Conditions, cond)
	}

	// Collect in slice. (Sorted for reproducibility.)
	var cases []*Case
	for _, cse := range groups {
		cases = append(cases, cse)
	}

	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Instruction < cases[j].Instruction
	})

	// Output switch statement.
	c.Printf("switch {\n")
	for _, cse := range cases {
		c.Printf("case %s:\n", strings.Join(cse.Conditions, ",\n"))
		c.Printf("return &%s, nil\n", cse.Instruction)
	}
	c.Printf("}\n")

	c.Printf("return nil, errors.New(\"%s: bad operands\")\n", fn.Name())
}

func construct(fn *api.Function, f inst.Form, s api.Signature) string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "intrep.Instruction{\n")
	fmt.Fprintf(buf, "\tOpcode: %#v,\n", fn.Instruction.Opcode)
	if len(fn.Suffixes) > 0 {
		fmt.Fprintf(buf, "\tSuffixes: %#v,\n", fn.Suffixes.Strings())
	}
	fmt.Fprintf(buf, "\tOperands: %s,\n", s.ParameterSlice())

	// Inputs.
	fmt.Fprintf(buf, "\tInputs: %s,\n", operandsWithAction(f, inst.R, s))

	// Outputs.
	fmt.Fprintf(buf, "\tOutputs: %s,\n", operandsWithAction(f, inst.W, s))

	// ISAs.
	if len(f.ISA) > 0 {
		fmt.Fprintf(buf, "\tISA: %#v,\n", f.ISA)
	}

	// Branch variables.
	if fn.Instruction.IsTerminal() {
		fmt.Fprintf(buf, "\tIsTerminal: true,\n")
	}

	if fn.Instruction.IsBranch() {
		fmt.Fprintf(buf, "\tIsBranch: true,\n")
		fmt.Fprintf(buf, "\tIsConditional: %#v,\n", fn.Instruction.IsConditionalBranch())
	}

	// Cancelling inputs.
	if f.CancellingInputs {
		fmt.Fprintf(buf, "\tCancellingInputs: true,\n")
	}

	fmt.Fprintf(buf, "}")
	return buf.String()
}

func operandsWithAction(f inst.Form, a inst.Action, s api.Signature) string {
	opexprs := []string{}
	for i, op := range f.Operands {
		if op.Action.ContainsAny(a) {
			opexprs = append(opexprs, s.ParameterName(i))
		}
	}
	for _, op := range f.ImplicitOperands {
		if op.Action.ContainsAny(a) {
			opexprs = append(opexprs, api.ImplicitRegister(op.Register))
		}
	}
	return fmt.Sprintf("[]%s{%s}", api.OperandType, strings.Join(opexprs, ", "))
}
