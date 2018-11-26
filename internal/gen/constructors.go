package gen

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/mmcloughlin/avo/internal/inst"
)

type constructors struct {
	cfg Config
	printer
}

func NewConstructors(cfg Config) Interface {
	return GoFmt(&constructors{cfg: cfg})
}

func (c *constructors) Generate(is []inst.Instruction) ([]byte, error) {
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

func (c *constructors) instruction(i inst.Instruction) {
	for _, line := range c.doc(i) {
		c.Printf("// %s\n", line)
	}

	s := params(i)

	c.Printf("func %s(%s) (*avo.Instruction, error) {\n", i.Opcode, s.ParameterList())
	c.checkargs(i, s)
	c.Printf("\treturn nil, nil\n")
	c.Printf("}\n\n")
}

// doc generates the lines of the function comment.
func (c *constructors) doc(i inst.Instruction) []string {
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

func (c *constructors) checkargs(i inst.Instruction, s signature) {
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

// signature provides access to details about the signature of an instruction function.
type signature interface {
	ParameterList() string
	ParameterName(int) string
	Length() string
}

// argslist is the signature for a function with the given named parameters.
type argslist []string

func (a argslist) ParameterList() string      { return strings.Join(a, ", ") + " avo.Operand" }
func (a argslist) ParameterName(i int) string { return a[i] }
func (a argslist) Length() string             { return strconv.Itoa(len(a)) }

// variadic is the signature for a variadic function.
type variadic struct {
	name string
}

func (v variadic) ParameterList() string      { return v.name + " ...avo.Operand" }
func (v variadic) ParameterName(i int) string { return fmt.Sprintf("%s[%d]", v.name, i) }
func (v variadic) Length() string             { return fmt.Sprintf("len(%s)", v.name) }

// niladic is the signature for a function with no arguments.
type niladic struct{}

func (n niladic) ParameterList() string      { return "" }
func (n niladic) ParameterName(i int) string { panic("niladic function has no parameters") }
func (n niladic) Length() string             { return "0" }

// params generates the function parameters and a function.
func params(i inst.Instruction) signature {
	// Handle the case of forms with multiple arities.
	switch {
	case i.IsVariadic():
		return variadic{name: "ops"}
	case i.IsNiladic():
		return niladic{}
	}

	// Generate nice-looking variable names.
	n := i.Arity()
	ops := make([]string, n)
	count := map[string]int{}
	for j := 0; j < n; j++ {
		// Collect unique lowercase bytes from first characters of operand types.
		s := map[byte]bool{}
		for _, f := range i.Forms {
			c := f.Operands[j].Type[0]
			if 'a' <= c && c <= 'z' {
				s[c] = true
			}
		}

		// Operand name is the sorted bytes.
		var b []byte
		for c := range s {
			b = append(b, c)
		}
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		name := string(b)

		// Append a counter if we've seen it already.
		m := count[name]
		count[name]++
		if m > 0 {
			name += strconv.Itoa(m)
		}
		ops[j] = name
	}

	return argslist(ops)
}

// checkername returns the name of the function that checks an operand of type t.
func checkername(t string) string {
	return "operand.Is" + strings.Title(t)
}
