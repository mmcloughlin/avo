package gen

import (
	"fmt"
	"strings"

	"github.com/mmcloughlin/avo/internal/prnt"
)

type enum struct {
	name   string
	doc    []string
	values []string
}

func (e *enum) Print(p *prnt.Generator) {
	// Type declaration.
	p.Comment(e.doc...)
	p.Printf("type %s %s\n\n", e.name, e.UnderlyingType())

	// Supported values.
	p.Printf("const (\n")
	p.Printf("\t%s %s = iota\n", e.None(), e.name)
	for _, value := range e.values {
		p.Printf("\t%s\n", e.ConstName(value))
	}
	p.Printf("\t%s\n", e.MaxName())
	p.Printf(")\n\n")

	// String method.
	stringtab := strings.ToLower(e.name) + "strings"

	r := e.Receiver()
	p.Printf("func (%s %s) String() string {\n", r, e.name)
	p.Printf("if %s < %s && %s < %s {\n", e.None(), r, r, e.MaxName())
	p.Printf("return %s[%s-1]\n", stringtab, r)
	p.Printf("}\n")
	p.Printf("return \"\"\n")
	p.Printf("}\n\n")

	p.Printf("var %s = []string{\n", stringtab)
	for _, value := range e.values {
		p.Printf("\t%q,\n", value)
	}
	p.Printf("}\n\n")
}

func (e *enum) Receiver() string {
	return strings.ToLower(e.name[:1])
}

func (e *enum) None() string {
	return e.ConstName("None")
}

func (e *enum) ConstName(value string) string {
	return e.name + value
}

func (e *enum) MaxName() string {
	return strings.ToLower(e.ConstName("max"))
}

func (e *enum) Max() int {
	return len(e.values)
}

func (e *enum) UnderlyingType() string {
	b := uint(8)
	for ; b < 64 && e.Max() > ((1<<b)-1); b <<= 1 {
	}
	return fmt.Sprintf("uint%d", b)
}
