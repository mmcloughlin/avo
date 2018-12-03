package build

import (
	"errors"
	"io"
	"log"

	"github.com/mmcloughlin/avo"
)

type Context struct {
	file     *avo.File
	function *avo.Function
	errs     []error
}

func NewContext() *Context {
	return &Context{
		file: avo.NewFile(),
	}
}

func (c *Context) Function(name string) {
	c.function = avo.NewFunction(name)
	c.file.Functions = append(c.file.Functions, c.function)
}

func (c *Context) Instruction(i *avo.Instruction) {
	c.node(i)
}

func (c *Context) Label(l avo.Label) {
	c.node(l)
}

func (c *Context) node(n avo.Node) {
	if c.function == nil {
		c.AddErrorMessage("no active function")
		return
	}
	c.function.AddNode(n)
}

//go:generate avogen -output zinstructions.go build

func (c *Context) AddError(err error) {
	c.errs = append(c.errs, err)
}

func (c *Context) AddErrorMessage(msg string) {
	c.AddError(errors.New(msg))
}

func (c *Context) Result() (*avo.File, []error) {
	return c.file, c.errs
}

func (c *Context) Main(wout, werr io.Writer) int {
	diag := log.New(werr, "", 0)

	f, errs := c.Result()
	if errs != nil {
		for _, err := range errs {
			diag.Println(err)
		}
		return 1
	}

	p := avo.NewGoPrinter(wout)
	if err := p.Print(f); err != nil {
		diag.Println(err)
		return 1
	}

	return 0
}
