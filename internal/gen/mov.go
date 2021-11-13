package gen

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mmcloughlin/avo/internal/api"
	"github.com/mmcloughlin/avo/internal/inst"
	"github.com/mmcloughlin/avo/internal/prnt"
	"github.com/mmcloughlin/avo/printer"
)

type mov struct {
	cfg printer.Config
	prnt.Generator
}

// NewMOV generates a function that will auto-select the correct MOV instruction
// based on operand types and and sizes.
func NewMOV(cfg printer.Config) Interface {
	return GoFmt(&mov{cfg: cfg})
}

func (m *mov) Generate(is []inst.Instruction) ([]byte, error) {
	m.Printf("// %s\n\n", m.cfg.GeneratedWarning())
	m.Printf("package build\n\n")

	m.Printf("import (\n")
	m.Printf("\t\"go/types\"\n")
	m.NL()
	m.Printf("\t%q\n", api.ImportPath(api.OperandPackage))
	m.Printf(")\n\n")

	m.Printf("func (c *Context) mov(a, b operand.Op, an, bn int, t *types.Basic) {\n")
	m.Printf("switch {\n")
	for _, i := range is {
		if ismov(i) {
			m.instruction(i)
		}
	}
	m.Printf("default:\n")
	m.Printf("c.adderrormessage(\"could not deduce mov instruction\")\n")
	m.Printf("}\n")
	m.Printf("}\n")
	return m.Result()
}

func (m *mov) instruction(i inst.Instruction) {
	f := flags(i)
	mfs, err := movforms(i)
	if err != nil {
		m.AddError(err)
		return
	}
	for _, mf := range mfs {
		conds := []string{
			fmt.Sprintf("an == %d", opsize[mf.A]),
			fmt.Sprintf("%s(a)", api.CheckerName(mf.A)),
			fmt.Sprintf("bn == %d", opsize[mf.B]),
			fmt.Sprintf("%s(b)", api.CheckerName(mf.B)),
		}
		for c, on := range f {
			cmp := map[bool]string{true: "!=", false: "=="}
			cond := fmt.Sprintf("(t.Info() & %s) %s 0", c, cmp[on])
			conds = append(conds, cond)
		}
		sort.Strings(conds)
		m.Printf("case %s:\n", strings.Join(conds, " && "))
		m.Printf("c.%s(a, b)\n", i.Opcode)
	}
}

// ismov decides whether the given instruction is a plain move instruction.
func ismov(i inst.Instruction) bool {
	// Ignore aliases.
	if i.AliasOf != "" {
		return false
	}

	// Accept specific move instruction prefixes.
	prefixes := []string{"MOV", "KMOV", "VMOV"}
	accept := false
	for _, prefix := range prefixes {
		accept = strings.HasPrefix(i.Opcode, prefix) || accept
	}
	if !accept {
		return false
	}

	// Exclude some cases based on instruction descriptions.
	exclude := []string{"Packed", "Duplicate", "Aligned", "Hint", "Swapping"}
	for _, substring := range exclude {
		if strings.Contains(i.Summary, substring) {
			return false
		}
	}

	return true
}

func flags(i inst.Instruction) map[string]bool {
	f := map[string]bool{}
	switch {
	case strings.Contains(i.Summary, "Floating-Point"):
		f["types.IsFloat"] = true
	case strings.Contains(i.Summary, "Zero-Extend"):
		f["types.IsInteger"] = true
		f["types.IsUnsigned"] = true
	case strings.Contains(i.Summary, "Sign-Extension"):
		f["types.IsInteger"] = true
		f["types.IsUnsigned"] = false
	default:
		f["types.IsInteger"] = true
	}
	return f
}

type movform struct{ A, B string }

func movforms(i inst.Instruction) ([]movform, error) {
	var mfs []movform
	for _, f := range i.Forms {
		if f.Arity() != 2 {
			continue
		}
		mf := movform{
			A: f.Operands[0].Type,
			B: f.Operands[1].Type,
		}
		if opsize[mf.A] < 0 || opsize[mf.B] < 0 {
			continue
		}
		if opsize[mf.A] == 0 || opsize[mf.B] == 0 {
			return nil, errors.New("unknown operand type")
		}
		mfs = append(mfs, mf)
	}
	return mfs, nil
}

var opsize = map[string]int8{
	"imm8":  -1,
	"imm16": -1,
	"imm32": -1,
	"imm64": -1,
	"r8":    1,
	"r16":   2,
	"r32":   4,
	"r64":   8,
	"xmm":   16,
	"ymm":   32,
	"zmm":   64,
	"m8":    1,
	"m16":   2,
	"m32":   4,
	"m64":   8,
	"m128":  16,
	"m256":  32,
	"m512":  64,
	"k":     8,
}
