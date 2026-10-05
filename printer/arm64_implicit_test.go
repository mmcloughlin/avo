package printer

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/ir"
	"github.com/mmcloughlin/avo/reg"
	"github.com/mmcloughlin/avo/x86"
)

// TestARM64RenameReachesImplicitOperands checks that when coalescing renames a
// register family at an instruction naming it both explicitly and implicitly,
// the lowering renders both roles through the rename: the family's own arm64
// register must not appear. Coalescing leaves implicit-only families alone, but
// a family named explicitly is eligible, so a lowering that spelled its
// implicit half with the package-level rename would read or write the wrong
// register.
func TestARM64RenameReachesImplicitOperands(t *testing.T) {
	mk := func(i *ir.Instruction, err error) *ir.Instruction {
		if err != nil {
			t.Fatal(err)
		}
		return i
	}
	cases := []struct {
		ins *ir.Instruction
		fam reg.Physical
	}{
		{mk(x86.MULQ(reg.RDX)), reg.RDX},
		{mk(x86.MULQ(reg.RAX)), reg.RAX},
		{mk(x86.IMULQ(reg.RDX)), reg.RDX},
		{mk(x86.IMULQ(reg.RAX)), reg.RAX},
		{mk(x86.MULXQ(reg.RCX, reg.RDX, reg.RBX)), reg.RDX},
		{mk(x86.MULXQ(reg.RDX, reg.RCX, reg.RBX)), reg.RDX},
		{mk(x86.SHLQ(reg.CL, reg.RBX)), reg.RCX},
		{mk(x86.SHRQ(reg.CL, reg.RCX)), reg.RCX},
		{mk(x86.SARQ(reg.CL, reg.RBX)), reg.RCX},
		{mk(x86.ROLQ(reg.CL, reg.RBX)), reg.RCX},
	}
	for _, c := range cases {
		f := regFamily(c.fam)
		e := familyEffects(c.ins)
		if e.explicit&(1<<f) == 0 {
			t.Fatalf("%s: family %s is not explicit; the case tests nothing", c.ins.Opcode, c.fam.Asm())
		}
		own := rename(c.fam)
		p := &arm64{renames: map[int]string{f: promoRegs[0]}}
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: lowering panicked: %v", c.ins.Opcode, r)
				}
			}()
			p.lower(c.ins, false, false)
		}()
		var lines []string
		for _, in := range p.pending {
			lines = append(lines, in[0]+" "+in[1])
		}
		out := strings.Join(lines, "\n")
		if regexp.MustCompile(`\b` + own + `\b`).MatchString(out) {
			t.Errorf("%s with %s renamed to %s still names %s:\n%s",
				c.ins.Opcode, c.fam.Asm(), promoRegs[0], own, out)
		}
	}
}
