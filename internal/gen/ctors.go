package gen

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/mmcloughlin/avo/internal/inst"
)

type ctors struct {
	cfg Config
	printer
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
	for _, line := range c.doc(i) {
		c.Printf("// %s\n", line)
	}

	s := params(i)

	c.Printf("func %s(%s) (*avo.Instruction, error) {\n", i.Opcode, s.ParameterList())
	c.checkargs(i, s)
	c.Printf("\treturn &avo.Instruction{Opcode: %#v, Operands: %s}, nil\n", i.Opcode, s.ParameterSlice())
	c.Printf("}\n\n")
}

// doc generates the lines of the function comment.
func (c *ctors) doc(i inst.Instruction) []string {
	lines := []string{
		fmt.Sprintf("%s: %s.", i.Opcode, i.Summary),
		"",
		"Forms:",
		"",
	}

	// Write a table of instruction forms.
	buf := bytes.NewBuffer(nil)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for _, f := range i.Forms {
		row := i.Opcode + "\t" + strings.Join(f.Signature(), "\t") + "\n"
		fmt.Fprint(w, row)
	}
	w.Flush()

	tbl := strings.TrimSpace(buf.String())
	for _, line := range strings.Split(tbl, "\n") {
		lines = append(lines, "\t"+line)
	}

	return lines
}

func (c *ctors) checkargs(i inst.Instruction, s signature) {
	if i.IsNiladic() {
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
	}

	c.Printf("default:\n")
	c.Printf("return nil, ErrBadOperandTypes\n")

	c.Printf("}\n")
}

// checkername returns the name of the function that checks an operand of type t.
func checkername(t string) string {
	return "operand.Is" + strings.Title(t)
}
