package build

import (
	"errors"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/gotypes"
	"github.com/mmcloughlin/avo/reg"
)

type Context struct {
	file     *avo.File
	function *avo.Function
	errs     []error
	reg.Collection
}

func NewContext() *Context {
	return &Context{
		file:       avo.NewFile(),
		Collection: *reg.NewCollection(),
	}
}

func (c *Context) Function(name string) {
	c.function = avo.NewFunction(name)
	c.file.Functions = append(c.file.Functions, c.function)
}

func (c *Context) Signature(s *gotypes.Signature) {
	c.activefunc().SetSignature(s)
}

func (c *Context) SignatureExpr(expr string) {
	s, err := gotypes.ParseSignature(expr)
	if err != nil {
		c.AddError(err)
		return
	}
	c.Signature(s)
}

func (c *Context) Instruction(i *avo.Instruction) {
	c.activefunc().AddNode(i)
}

func (c *Context) Label(l avo.Label) {
	c.activefunc().AddLabel(l)
}

func (c *Context) activefunc() *avo.Function {
	if c.function == nil {
		c.AddErrorMessage("no active function")
		return avo.NewFunction("")
	}
	return c.function
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
