package gen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/printer"

	"github.com/mmcloughlin/avo/internal/inst"
)

type ctorstest struct {
	cfg printer.Config
	printer.Generator
}

// NewCtorsTest autogenerates tests for the constructors build by NewCtors.
func NewCtorsTest(cfg printer.Config) Interface {
	return GoFmt(&ctorstest{cfg: cfg})
}

func (c *ctorstest) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.BuildTag("!integration")
	c.NL()
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"math\"\n")
	c.Printf("\t\"testing\"\n")
	c.NL()
	c.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	c.Printf("\t%q\n", api.ImportPath(api.RegisterPackage))
	c.Printf(")\n\n")

	DeclareTestArguments(&c.Generator)

	fns := api.InstructionsFunctions(is)
	for _, fn := range fns {
		c.function(fn)
	}

	return c.Result()
}

func (c *ctorstest) function(fn *api.Function) {
	c.Printf("func Test%sValidFormsNoError(t *testing.T) {", fn.Name())
	for _, f := range fn.Forms {
		s := TestSignature(f)
		c.Printf("if _, err := %s(%s); err != nil { t.Fatal(err) }\n", fn.Name(), s.Arguments())
	}
	c.Printf("}\n\n")
}

type ctorsstress struct {
	cfg printer.Config
	printer.Generator
}

// NewCtorsStress autogenerates stress tests for instruction constructors.
func NewCtorsStress(cfg printer.Config) Interface {
	return GoFmt(&ctorsstress{cfg: cfg})
}

func (c *ctorsstress) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.BuildTag("stress")
	c.NL()
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"reflect\"\n")
	c.Printf("\t\"testing\"\n")
	c.NL()
	c.Printf("\t%q\n", api.ImportPath(api.IRPackage))
	c.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	c.Printf("\t%q\n", api.ImportPath(api.RegisterPackage))
	c.Printf(")\n\n")

	fns := api.InstructionsFunctions(is)
	for _, fn := range fns {
		c.function(fn)
	}

	return c.Result()
}

func (c *ctorsstress) function(fn *api.Function) {
	c.Printf("func Test%sValidFormsCorrectInstruction(t *testing.T) {", fn.Name())
	for _, f := range fn.Forms {
		name := strings.Join(f.Signature(), "_")
		c.Printf("t.Run(\"form=%s\", func(t *testing.T) {\n", name)
		s := TestSignature(f)
		c.Printf("expect := &%s\n", construct(fn, f, s))
		c.Printf("got, err := %s(%s);\n", fn.Name(), s.Arguments())
		c.Printf("if err != nil { t.Fatal(err) }\n")
		c.Printf("if !reflect.DeepEqual(got, expect) { t.Fatal(\"mismatch\") }\n")
		c.Printf("})\n")
	}
	c.Printf("}\n\n")
}

type ctorsbench struct {
	cfg printer.Config
	printer.Generator
}

// NewCtorsBench autogenerates a benchmark for the instruction constructors.
func NewCtorsBench(cfg printer.Config) Interface {
	return GoFmt(&ctorsbench{cfg: cfg})
}

func (c *ctorsbench) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.BuildTag("stress")
	c.NL()
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"time\"\n")
	c.Printf("\t\"testing\"\n")
	c.Printf(")\n\n")

	c.Printf("func BenchmarkConstructors(b *testing.B) {\n")
	c.Printf("start := time.Now()\n")
	c.Printf("for i := 0; i < b.N; i++ {\n")
	n := 0
	for _, fn := range api.InstructionsFunctions(is) {
		for _, f := range fn.Forms {
			n++
			c.Printf("%s(%s)\n", fn.Name(), TestSignature(f).Arguments())
		}
	}
	c.Printf("}\n")
	c.Printf("elapsed := time.Since(start)\n")
	c.Printf("\tb.ReportMetric(%d * float64(b.N) / elapsed.Seconds(), \"inst/s\")\n", n)
	c.Printf("}\n\n")
	return c.Result()
}

func construct(fn *api.Function, f inst.Form, s api.Signature) string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "ir.Instruction{\n")
	fmt.Fprintf(buf, "\tOpcode: %#v,\n", fn.Instruction.Opcode)
	if len(fn.Suffixes) > 0 {
		fmt.Fprintf(buf, "\tSuffixes: %#v,\n", fn.Suffixes.Strings())
	}
	fmt.Fprintf(buf, "\tOperands: %s,\n", s.ParameterSlice())

	// Inputs.
	fmt.Fprintf(buf, "\tInputs: %s,\n", operandsWithAction(f, inst.R, s))

	// Outputs.
	fmt.Fprintf(buf, "\tOutputs: %s,\n", operandsWithAction(f, inst.W, s))

	// ISAs.
	if len(f.ISA) > 0 {
		fmt.Fprintf(buf, "\tISA: %#v,\n", f.ISA)
	}

	// Branch variables.
	if fn.Instruction.IsTerminal() {
		fmt.Fprintf(buf, "\tIsTerminal: true,\n")
	}

	if fn.Instruction.IsBranch() {
		fmt.Fprintf(buf, "\tIsBranch: true,\n")
		fmt.Fprintf(buf, "\tIsConditional: %#v,\n", fn.Instruction.IsConditionalBranch())
	}

	// Cancelling inputs.
	if f.CancellingInputs {
		fmt.Fprintf(buf, "\tCancellingInputs: true,\n")
	}

	fmt.Fprintf(buf, "}")
	return buf.String()
}

func operandsWithAction(f inst.Form, a inst.Action, s api.Signature) string {
	var opexprs []string
	for i, op := range f.Operands {
		if op.Action.ContainsAny(a) {
			opexprs = append(opexprs, s.ParameterName(i))
		}
	}
	for _, op := range f.ImplicitOperands {
		if op.Action.ContainsAny(a) {
			opexprs = append(opexprs, api.ImplicitRegister(op.Register))
		}
	}
	if len(opexprs) == 0 {
		return "nil"
	}
	return fmt.Sprintf("[]%s{%s}", api.OperandType, strings.Join(opexprs, ", "))
}
