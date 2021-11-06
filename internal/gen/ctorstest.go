package gen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"

	"github.com/mmcloughlin/avo/internal/inst"
)

type ctorstest struct {
	cfg printer.Config
	prnt.Generator
}

// NewCtorsTest autogenerates tests for the constructors build by NewCtors.
func NewCtorsTest(cfg printer.Config) Interface {
	return GoFmt(&ctorstest{cfg: cfg})
}

func (c *ctorstest) Generate(is []inst.Instruction) ([]byte, error) {
	c.Printf("// %s\n\n", c.cfg.GeneratedWarning())
	c.BuildTag("stress")
	c.NL()
	c.Printf("package x86\n\n")
	c.Printf("import (\n")
	c.Printf("\t\"math\"\n")
	c.Printf("\t\"reflect\"\n")
	c.Printf("\t\"testing\"\n")
	c.Printf("\t\"time\"\n")
	c.NL()
	c.Printf("\tintrep \"%s/ir\"\n", api.Package)
	c.Printf("\t\"%s/reg\"\n", api.Package)
	c.Printf("\t\"%s/operand\"\n", api.Package)
	c.Printf(")\n\n")

	c.args()

	fns := api.InstructionsFunctions(is)
	for _, fn := range fns {
		c.function(fn)
	}

	c.benchmark(fns)

	return c.Result()
}

func (c *ctorstest) args() {
	c.Printf("var (\n")
	for _, arg := range validArgs {
		c.Printf("\t%s operand.Op = %s\n", argname(arg.Type), arg.Code)
	}
	c.Printf(")\n")
}

func (c *ctorstest) function(fn *api.Function) {
	c.Printf("func Test%sValidForms(t *testing.T) {", fn.Name())

	for _, f := range fn.Forms {
		name := strings.Join(f.Signature(), "_")
		c.Printf("t.Run(\"form=%s\", func(t *testing.T) {\n", name)
		s := formsig(f)
		c.Printf("expect := &%s\n", construct(fn, f, s))
		c.Printf("got, err := %s(%s)\n", fn.Name(), s.Arguments())
		c.Printf("if err != nil { t.Fatal(err) }\n")
		c.Printf("if !reflect.DeepEqual(expect, got) { t.Fatal(\"mismatch\") }\n")
		c.Printf("})\n")
	}

	c.Printf("}\n\n")
}

func (c *ctorstest) benchmark(fns []*api.Function) {
	c.Printf("func BenchmarkConstructors(b *testing.B) {\n")
	c.Printf("start := time.Now()\n")
	c.Printf("for i := 0; i < b.N; i++ {\n")
	n := 0
	for _, fn := range fns {
		for _, f := range fn.Forms {
			n++
			c.Printf("%s(%s)\n", fn.Name(), formsig(f).Arguments())
		}
	}
	c.Printf("}\n")
	c.Printf("elapsed := time.Since(start)\n")
	c.Printf("\tb.ReportMetric(%d * float64(b.N) / elapsed.Seconds(), \"inst/s\")\n", n)
	c.Printf("}\n\n")
}

func construct(fn *api.Function, f inst.Form, s api.Signature) string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "intrep.Instruction{\n")
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
	opexprs := []string{}
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
	return fmt.Sprintf("[]%s{%s}", api.OperandType, strings.Join(opexprs, ", "))
}

func formsig(f inst.Form) api.Signature {
	var names []string
	for _, op := range f.Operands {
		names = append(names, argname(op.Type))
	}
	return api.ArgsList(names)
}

func argname(t string) string {
	return "op" + t
}

var validArgs = []struct {
	Type string
	Code string
}{
	// Immediates
	{"1", "operand.Imm(1)"},
	{"3", "operand.Imm(3)"},
	{"imm2u", "operand.Imm(3)"},
	{"imm8", "operand.Imm(math.MaxInt8)"},
	{"imm16", "operand.Imm(math.MaxInt16)"},
	{"imm32", "operand.Imm(math.MaxInt32)"},
	{"imm64", "operand.Imm(math.MaxInt64)"},

	// Registers
	{"al", "reg.AL"},
	{"cl", "reg.CL"},
	{"ax", "reg.AX"},
	{"eax", "reg.EAX"},
	{"rax", "reg.RAX"},
	{"r8", "reg.CH"},
	{"r16", "reg.R9W"},
	{"r32", "reg.R10L"},
	{"r64", "reg.R11"},
	{"xmm0", "reg.X0"},
	{"xmm", "reg.X7"},
	{"ymm", "reg.Y15"},
	{"zmm", "reg.Z31"},
	{"k", "reg.K7"},

	// Memory
	{"m", "operand.Mem{Base: reg.BX, Index: reg.CX, Scale: 2}"},
	{"m8", "operand.Mem{Base: reg.BL, Index: reg.CH, Scale: 1}"},
	{"m16", "operand.Mem{Base: reg.BX, Index: reg.CX, Scale: 2}"},
	{"m32", "operand.Mem{Base: reg.EBX, Index: reg.ECX, Scale: 4}"},
	{"m64", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m128", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m256", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},
	{"m512", "operand.Mem{Base: reg.RBX, Index: reg.RCX, Scale: 8}"},

	// Vector memory
	{"vm32x", "operand.Mem{Base: reg.R13, Index: reg.X4, Scale: 1}"},
	{"vm64x", "operand.Mem{Base: reg.R13, Index: reg.X8, Scale: 1}"},
	{"vm32y", "operand.Mem{Base: reg.R13, Index: reg.Y4, Scale: 1}"},
	{"vm64y", "operand.Mem{Base: reg.R13, Index: reg.Y8, Scale: 1}"},
	{"vm32z", "operand.Mem{Base: reg.R13, Index: reg.Z4, Scale: 1}"},
	{"vm64z", "operand.Mem{Base: reg.R13, Index: reg.Z8, Scale: 1}"},

	// Relative
	{"rel8", "operand.Rel(math.MaxInt8)"},
	{"rel32", "operand.LabelRef(\"lbl\")"},
}
