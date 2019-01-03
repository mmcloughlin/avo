package build

import (
	"flag"
	"os"

	"github.com/mmcloughlin/avo/buildtags"
	"github.com/mmcloughlin/avo/gotypes"
	"github.com/mmcloughlin/avo/operand"

	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo"
)

// ctx provides a global build context.
var ctx = NewContext()

func Package(path string) { ctx.Package(path) }

func TEXT(name, signature string) {
	ctx.Function(name)
	ctx.SignatureExpr(signature)
}

func LABEL(name string) { ctx.Label(avo.Label(name)) }

func GLOBL(name string, a avo.Attribute) operand.Mem {
	g := ctx.StaticGlobal(name)
	ctx.DataAttributes(a)
	return g
}

func DATA(offset int, v operand.Constant) {
	ctx.AddDatum(offset, v)
}

var flags = NewFlags(flag.CommandLine)

func Generate() {
	if !flag.Parsed() {
		flag.Parse()
	}
	cfg := flags.Config()

	status := Main(cfg, ctx)

	// To record coverage of integration tests we wrap main() functions in a test
	// functions. In this case we need the main function to terminate, therefore we
	// only exit for failure status codes.
	if status != 0 {
		os.Exit(status)
	}
}

func BuildConstraints(t buildtags.ConstraintsConvertable) { ctx.BuildConstraints(t) }
func BuildConstraint(t buildtags.ConstraintConvertable)   { ctx.BuildConstraint(t) }
func BuildConstraintExpr(expr string)                     { ctx.BuildConstraintExpr(expr) }

func GP8v() reg.GPVirtual  { return ctx.GP8v() }
func GP16v() reg.GPVirtual { return ctx.GP16v() }
func GP32v() reg.GPVirtual { return ctx.GP32v() }
func GP64v() reg.GPVirtual { return ctx.GP64v() }
func Xv() reg.VecVirtual   { return ctx.Xv() }
func Yv() reg.VecVirtual   { return ctx.Yv() }
func Zv() reg.VecVirtual   { return ctx.Zv() }

func Param(name string) gotypes.Component  { return ctx.Param(name) }
func ParamIndex(i int) gotypes.Component   { return ctx.ParamIndex(i) }
func Return(name string) gotypes.Component { return ctx.Return(name) }
func ReturnIndex(i int) gotypes.Component  { return ctx.ReturnIndex(i) }

func Load(src gotypes.Component, dst reg.Register) reg.Register { return ctx.Load(src, dst) }
func Store(src reg.Register, dst gotypes.Component)             { ctx.Store(src, dst) }

func Doc(lines ...string)        { ctx.Doc(lines...) }
func Attributes(a avo.Attribute) { ctx.Attributes(a) }

func AllocLocal(size int) operand.Mem { return ctx.AllocLocal(size) }

func ConstData(name string, v operand.Constant) operand.Mem { return ctx.ConstData(name, v) }
