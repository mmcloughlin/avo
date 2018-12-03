package gen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mmcloughlin/avo/internal/inst"
)

type ctors struct {
	cfg Config
	generator
}

func NewCtors(cfg Config) Interface {
	return GoFmt(&ctors{cfg: cfg})
}

func (c *ctors) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"%s\"\n", pkg)
	c.Printf("\t\"%s/operand\"\n", pkg)
	c.Printf(")\n\n")

	for _, i := range is {
		c.instruction(i)
	}

	return c.Result()
}

func (c *ctors) instruction(i inst.Instruction) {
	for _, line := range doc(i) {
		c.Printf("// %s\n", line)
	}

	s := params(i)

	c.Printf("func %s(%s) (*avo.Instruction, error) {\n", i.Opcode, s.ParameterList())
	c.forms(i, s)
	c.Printf("}\n\n")
}

func (c *ctors) forms(i inst.Instruction, s signature) {
	if i.IsNiladic() {
		if len(i.Forms) != 1 {
			c.AddError(fmt.Errorf("%s breaks assumption that niladic instructions have one form", i.Opcode))
		}
		c.Printf("return &%s, nil\n", construct(i, i.Forms[0], s))
		return
	}

	c.Printf("switch {\n")

	for _, f := range i.Forms {
		var conds []string

		if i.IsVariadic() {
			checklen := fmt.Sprintf("%s == %d", s.Length(), len(f.Operands))
			conds = append(conds, checklen)
		}

		for j, op := range f.Operands {
			checktype := fmt.Sprintf("%s(%s)", checkername(op.Type), s.ParameterName(j))
			conds = append(conds, checktype)
		}

		c.Printf("case %s:\n", strings.Join(conds, " && "))
		c.Printf("return &%s, nil\n", construct(i, f, s))
	}

	c.Printf("}\n")
	c.Printf("return nil, ErrBadOperandTypes\n")
}

func construct(i inst.Instruction, f inst.Form, s signature) string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "avo.Instruction{\n")
	fmt.Fprintf(buf, "\tOpcode: %#v,\n", i.Opcode)
	fmt.Fprintf(buf, "\tOperands: %s,\n", s.ParameterSlice())

	// Input output.
	// TODO(mbm): handle implicit operands
	fmt.Fprintf(buf, "\tInputs: %s,\n", actionfilter(f.Operands, inst.R, s))
	fmt.Fprintf(buf, "\tOutputs: %s,\n", actionfilter(f.Operands, inst.W, s))

	// Branch variables.
	if i.IsBranch() {
		fmt.Fprintf(buf, "\tIsBranch: true,\n")
		fmt.Fprintf(buf, "\tIsConditional: %#v,\n", i.IsConditionalBranch())
	}

	fmt.Fprintf(buf, "}")
	return buf.String()
}

func actionfilter(ops []inst.Operand, a inst.Action, s signature) string {
	opexprs := []string{}
	for i, op := range ops {
		if op.Action.Contains(a) {
			opexprs = append(opexprs, s.ParameterName(i))
		}
	}
	return fmt.Sprintf("[]%s{%s}", operandType, strings.Join(opexprs, ", "))
}

// checkername returns the name of the function that checks an operand of type t.
func checkername(t string) string {
	return "operand.Is" + strings.Title(t)
}
