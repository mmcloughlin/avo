package gen

import (
	"fmt"
	"strconv"
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
}

func (e *enum) MapMethod(p *prnt.Generator, name, ret, zero string, mapping []string) {
	table := strings.ToLower(e.name + name + "table")

	r := e.Receiver()
	p.Printf("func (%s %s) %s() %s {\n", r, e.name, name, ret)
	p.Printf("if %s < %s && %s < %s {\n", e.None(), r, r, e.MaxName())
	p.Printf("return %s[%s-1]\n", table, r)
	p.Printf("}\n")
	p.Printf("return %s\n", zero)
	p.Printf("}\n\n")

	p.Printf("var %s = []%s{\n", table, ret)
	for _, value := range mapping {
		p.Printf("\t%s,\n", value)
	}
	p.Printf("}\n\n")
}

func (e *enum) StringMethod(p *prnt.Generator) {
	mapping := make([]string, len(e.values))
	for i, s := range e.values {
		mapping[i] = strconv.Quote(s)
	}
	e.MapMethod(p, "String", "string", `""`, mapping)
}

func (e *enum) Name() string {
	return e.name
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
