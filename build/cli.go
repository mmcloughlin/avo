package build

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/mmcloughlin/avo/pass"
	"github.com/mmcloughlin/avo/printer"
)

type Config struct {
	ErrOut io.Writer
	Passes []pass.Interface
}

func Main(cfg *Config, context *Context) int {
	diag := log.New(cfg.ErrOut, "", 0)

	f, errs := context.Result()
	if errs != nil {
		for _, err := range errs {
			diag.Println(err)
		}
		return 1
	}

	p := pass.Concat(cfg.Passes...)
	if err := p.Execute(f); err != nil {
		diag.Println(err)
		return 1
	}

	return 0
}

type Flags struct {
	errout   *outputValue
	printers []*printerValue
}

func NewFlags(fs *flag.FlagSet) *Flags {
	f := &Flags{}

	f.errout = newOutputValue(os.Stderr)
	fs.Var(f.errout, "log", "diagnostics output")

	goasm := newPrinterValue(printer.NewGoAsm, os.Stdout)
	fs.Var(goasm, "out", "assembly output")
	f.printers = append(f.printers, goasm)

	stubs := newPrinterValue(printer.NewStubs, nil)
	fs.Var(stubs, "stubs", "go stub file")
	f.printers = append(f.printers, stubs)

	return f
}

func (f *Flags) Config() *Config {
	pc := printer.NewGoRunConfig()
	passes := []pass.Interface{pass.Compile}
	for _, pv := range f.printers {
		p := pv.Build(pc)
		if p != nil {
			passes = append(passes, p)
		}
	}
	return &Config{
		ErrOut: f.errout.w,
		Passes: passes,
	}
}

type outputValue struct {
	w        io.WriteCloser
	filename string
}

func newOutputValue(dflt io.WriteCloser) *outputValue {
	return &outputValue{w: dflt}
}

func (o *outputValue) String() string {
	if o == nil {
		return ""
	}
	return o.filename
}

func (o *outputValue) Set(s string) error {
	o.filename = s
	if s == "-" {
		o.w = nopwritecloser{os.Stdout}
		return nil
	}
	f, err := os.Create(s)
	if err != nil {
		return err
	}
	o.w = f
	return nil
}

type printerValue struct {
	*outputValue
	Builder printer.Builder
}

func newPrinterValue(b printer.Builder, dflt io.WriteCloser) *printerValue {
	return &printerValue{
		outputValue: newOutputValue(dflt),
		Builder:     b,
	}
}

func (p *printerValue) Build(cfg printer.Config) pass.Interface {
	if p.outputValue.w == nil {
		return nil
	}
	return &pass.Output{
		Writer:  p.outputValue.w,
		Printer: p.Builder(cfg),
	}
}

// nopwritecloser wraps a Writer and provides a null implementation of Close().
type nopwritecloser struct {
	io.Writer
}

func (nopwritecloser) Close() error { return nil }
