package build

import (
	"errors"
	"go/types"

	"github.com/mmcloughlin/avo"
	"github.com/mmcloughlin/avo/gotypes"
	"github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
	"golang.org/x/tools/go/packages"
)

type Context struct {
	pkg      *packages.Package
	file     *avo.File
	function *avo.Function
	global   *avo.Global
	errs     []error
	reg.Collection
}

func NewContext() *Context {
	return &Context{
		file:       avo.NewFile(),
		Collection: *reg.NewCollection(),
	}
}

func (c *Context) Package(path string) {
	cfg := &packages.Config{
		Mode: packages.LoadTypes,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		c.AddError(err)
		return
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		for _, err := range pkg.Errors {
			c.AddError(err)
		}
		return
	}
	c.pkg = pkg
}

func (c *Context) Function(name string) {
	c.function = avo.NewFunction(name)
	c.file.AddSection(c.function)
}

func (c *Context) Signature(s *gotypes.Signature) {
	c.activefunc().SetSignature(s)
}

func (c *Context) SignatureExpr(expr string) {
	s, err := gotypes.ParseSignatureInPackage(c.types(), expr)
	if err != nil {
		c.AddError(err)
		return
	}
	c.Signature(s)
}

func (c *Context) types() *types.Package {
	if c.pkg == nil {
		return nil
	}
	return c.pkg.Types
}

func (c *Context) AllocLocal(size int) operand.Mem {
	return c.activefunc().AllocLocal(size)
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

func (c *Context) StaticGlobal(name string) operand.Mem {
	c.global = avo.NewStaticGlobal(name)
	c.file.AddSection(c.global)
	return c.global.Base()
}

func (c *Context) AddDatum(offset int, v operand.Constant) {
	if err := c.activeglobal().AddDatum(avo.NewDatum(offset, v)); err != nil {
		c.AddError(err)
	}
}

func (c *Context) AppendDatum(v operand.Constant) {
	c.activeglobal().Append(v)
}

func (c *Context) activeglobal() *avo.Global {
	if c.global == nil {
		c.AddErrorMessage("no active global")
		return avo.NewStaticGlobal("")
	}
	return c.global
}

func (c *Context) AddError(err error) {
	c.errs = append(c.errs, err)
}

func (c *Context) AddErrorMessage(msg string) {
	c.AddError(errors.New(msg))
}

func (c *Context) Result() (*avo.File, []error) {
	return c.file, c.errs
}
