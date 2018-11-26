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
	c.Printf("import \"github.com/mmcloughlin/avo\"\n\n")

	for _, i := range is {
		c.instruction(i)
	}

	return c.Result()
}

func (c *constructors) instruction(i inst.Instruction) {
	for _, line := range c.doc(i) {
		c.Printf("// %s\n", line)
	}

	paramlist, _ := params(i)

	c.Printf("func %s(%s) error {\n", i.Opcode, paramlist)
	c.Printf("\treturn nil\n")
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

// params generates the function parameters and a function.
func params(i inst.Instruction) (string, func(int) string) {
	a := i.Arities()

	// Handle the case of forms with multiple arities.
	if len(a) > 1 {
		return "ops ...avo.Operand", func(j int) string {
			return fmt.Sprintf("ops[%d]", j)
		}
	}

	// All forms have the same arity.
	n := a[0]
	if n == 0 {
		return "", func(int) string { panic("unreachable") }
	}

	// Generate nice-looking variable names.
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

	return strings.Join(ops, ", ") + " avo.Operand", func(j int) string { return ops[j] }
}
