package build

import (
	"flag"
	"os"

	"github.com/mmcloughlin/avo/gotypes"

	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo"
)

// ctx provides a global build context.
var ctx = NewContext()

func TEXT(name, signature string) {
	ctx.Function(name)
	ctx.SignatureExpr(signature)
}

func LABEL(name string) { ctx.Label(avo.Label(name)) }

var flags = NewFlags(flag.CommandLine)

func Generate() {
	if !flag.Parsed() {
		flag.Parse()
	}
	cfg := flags.Config()
	os.Exit(Main(cfg, ctx))
}

func GP8v() reg.Virtual  { return ctx.GP8v() }
func GP16v() reg.Virtual { return ctx.GP16v() }
func GP32v() reg.Virtual { return ctx.GP32v() }
func GP64v() reg.Virtual { return ctx.GP64v() }
func Xv() reg.Virtual    { return ctx.Xv() }
func Yv() reg.Virtual    { return ctx.Yv() }

func Param(name string) gotypes.Component  { return ctx.Param(name) }
func ParamIndex(i int) gotypes.Component   { return ctx.ParamIndex(i) }
func Return(name string) gotypes.Component { return ctx.Return(name) }
func ReturnIndex(i int) gotypes.Component  { return ctx.ReturnIndex(i) }

func Load(src gotypes.Component, dst reg.Register) reg.Register { return ctx.Load(src, dst) }
func Store(src reg.Register, dst gotypes.Component)             { ctx.Store(src, dst) }
