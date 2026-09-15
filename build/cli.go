package build

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/mmcloughlin/avo/pass"
	"github.com/mmcloughlin/avo/printer"
)

// Config contains options for an avo main function.
type Config struct {
	ErrOut     io.Writer
	MaxErrors  int // max errors to report; 0 means unlimited
	CPUProfile io.WriteCloser
	Passes     []pass.Interface
}

// Main is the standard main function for an avo program. This extracts the
// result from the build Context (logging and exiting on error), and performs
// configured passes.
func Main(cfg *Config, context *Context) int {
	diag := log.New(cfg.ErrOut, "", 0)

	if cfg.CPUProfile != nil {
		defer cfg.CPUProfile.Close()
		if err := pprof.StartCPUProfile(cfg.CPUProfile); err != nil {
			diag.Println("could not start CPU profile: ", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	f, err := context.Result()
	if err != nil {
		LogError(diag, err, cfg.MaxErrors)
		return 1
	}

	p := pass.Concat(cfg.Passes...)
	if err := p.Execute(f); err != nil {
		diag.Println(err)
		return 1
	}

	return 0
}

// Flags represents CLI flags for an avo program.
type Flags struct {
	errout    *outputValue
	allerrors bool
	cpuprof   *outputValue
	pkg       string
	arch      string
	arm64BMI2 bool
	goasm     *printerValue
	arm64     *printerValue
	stubs     *printerValue
	printers  []*printerValue
}

// archPrinter maps a GOARCH to the printer that produces its assembly. amd64 is
// emitted by the standard goasm printer; arm64 by the EXPERIMENTAL lowering
// printer (which lowers the amd64 instruction stream).
var archPrinter = map[string]printer.Builder{
	"amd64": printer.NewGoAsm,
	"arm64": printer.NewARM64Asm,
}

// NewFlags initializes avo flags for the given FlagSet.
func NewFlags(fs *flag.FlagSet) *Flags {
	f := &Flags{}

	f.errout = newOutputValue(os.Stderr)
	fs.Var(f.errout, "log", "diagnostics output")

	fs.BoolVar(&f.allerrors, "e", false, "no limit on number of errors reported")

	f.cpuprof = newOutputValue(nil)
	fs.Var(f.cpuprof, "cpuprofile", "write cpu profile to `file`")

	fs.StringVar(&f.pkg, "pkg", "", "package name (defaults to current directory name)")

	fs.StringVar(&f.arch, "arch", "", "comma-separated list of GOARCH values to emit; treats -out as a base path and writes <base>_GOARCH.s for each (amd64 via goasm, arm64 via the EXPERIMENTAL lowering printer)")

	fs.BoolVar(&f.arm64BMI2, "arm64-prefer-bmi2", false, "EXPERIMENTAL arm64 lowering: prefer a function's BMI2 twin over its generic one when both exist (default prefers generic; BMI2 x86 code is tuned for x86 and is not reliably faster once lowered -- measure before enabling)")

	f.goasm = newLazyPrinterValue(printer.NewGoAsm, os.Stdout)
	fs.Var(f.goasm, "out", "assembly output (or, with -arch, the base path)")

	f.arm64 = newLazyPrinterValue(printer.NewARM64Asm, nil)
	fs.Var(f.arm64, "arm64", "EXPERIMENTAL arm64 assembly output (lowered from amd64); prefer -arch")

	f.stubs = newLazyPrinterValue(printer.NewStubs, nil)
	fs.Var(f.stubs, "stubs", "go stub file")

	f.printers = []*printerValue{f.goasm, f.arm64, f.stubs}

	return f
}

// Config builds a configuration object based on flag values.
func (f *Flags) Config() *Config {
	pc := printer.NewGoRunConfig()
	if f.pkg != "" {
		pc.Pkg = f.pkg
	}
	pc.ARM64PreferBMI2 = f.arm64BMI2

	passes := []pass.Interface{pass.Compile}
	if f.arch != "" {
		passes = append(passes, f.archPasses(pc)...)
	} else {
		for _, pv := range f.printers {
			if p := pv.Build(pc); p != nil {
				passes = append(passes, p)
			}
		}
	}

	cfg := &Config{
		ErrOut:     f.errout.w,
		MaxErrors:  10,
		CPUProfile: f.cpuprof.w,
		Passes:     passes,
	}

	if f.allerrors {
		cfg.MaxErrors = 0
	}

	return cfg
}

// archPasses builds one output pass per GOARCH listed in -arch, deriving the
// output filename from the -out base (e.g. base "x.s" + "arm64" -> "x_arm64.s").
func (f *Flags) archPasses(pc printer.Config) []pass.Interface {
	base := f.goasm.filename
	var passes []pass.Interface
	for _, arch := range strings.Split(f.arch, ",") {
		arch = strings.TrimSpace(arch)
		if arch == "" {
			continue
		}
		build, ok := archPrinter[arch]
		if !ok {
			panic(fmt.Sprintf("avo: -arch %q not supported (have amd64, arm64)", arch))
		}
		w, err := createOutput(archFilename(base, arch))
		if err != nil {
			panic(err)
		}
		passes = append(passes, &pass.Output{Writer: w, Printer: build(pc)})
	}
	// Stubs are architecture-independent; emit them once if requested.
	if p := f.stubs.Build(pc); p != nil {
		passes = append(passes, p)
	}
	return passes
}

// archFilename inserts _GOARCH before the extension of base.
func archFilename(base, arch string) string {
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)] + "_" + arch + ext
}

type outputValue struct {
	w        io.WriteCloser
	filename string
	lazy     bool // defer file creation to Build (so the path can be reused, e.g. as an -arch base)
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
	if o.lazy && s != "-" {
		return nil // created later in Build / archPasses
	}
	w, err := createOutput(s)
	if err != nil {
		return err
	}
	o.w = w
	return nil
}

// createOutput opens a writer for the given filename ("-" means stdout).
func createOutput(filename string) (io.WriteCloser, error) {
	if filename == "-" {
		return nopwritecloser{os.Stdout}, nil
	}
	return os.Create(filename)
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

// newLazyPrinterValue is like newPrinterValue but defers file creation until
// Build, so the filename can also serve as a base path for -arch.
func newLazyPrinterValue(b printer.Builder, dflt io.WriteCloser) *printerValue {
	pv := newPrinterValue(b, dflt)
	pv.lazy = true
	return pv
}

func (p *printerValue) Build(cfg printer.Config) pass.Interface {
	// Materialize a deferred -out filename. A lazy printer leaves its writer as
	// the default (goasm defaults to stdout) during Set so the path can double as
	// an -arch base; an explicit -out must still win over that default, so the
	// file is opened here rather than falling back to the (non-nil) default.
	if p.lazy && p.filename != "" && p.filename != "-" {
		w, err := createOutput(p.filename)
		if err != nil {
			panic(err)
		}
		p.outputValue.w = w
		p.lazy = false // materialized; a second Build must not reopen it
	} else if p.outputValue.w == nil && p.outputValue.filename != "" {
		w, err := createOutput(p.outputValue.filename)
		if err != nil {
			panic(err)
		}
		p.outputValue.w = w
	}
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
