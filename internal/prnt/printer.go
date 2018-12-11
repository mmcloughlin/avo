package prnt

import (
	"bytes"
	"fmt"
)

type Generator struct {
	buf bytes.Buffer
	err error
}

func (g *Generator) Printf(format string, args ...interface{}) {
	if g.err != nil {
		return
	}
	if _, err := fmt.Fprintf(&g.buf, format, args...); err != nil {
		g.AddError(err)
	}
}

func (g *Generator) NL() {
	g.Printf("\n")
}

func (g *Generator) Comment(lines ...string) {
	for _, line := range lines {
		g.Printf("// %s\n", line)
	}
}

func (g *Generator) AddError(err error) {
	if err != nil && g.err == nil {
		g.err = err
	}
}

func (g *Generator) Result() ([]byte, error) {
	return g.buf.Bytes(), g.err
}
