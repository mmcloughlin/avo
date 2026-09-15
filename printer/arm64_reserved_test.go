package printer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/reg"
)

// reservedARM64 lists registers the lowering must never allocate or clobber,
// with the reason each one is off limits. R27 deserves particular care: Go's
// arm64 assembler silently uses it to synthesize out-of-range immediates and
// addresses, so a live value parked there would be destroyed with no
// diagnostic from either the assembler or the lowering.
var reservedARM64 = map[string]string{
	"R18": "platform register, reserved by the OS ABI",
	"R27": "REGTMP: Go's assembler clobbers it to synthesize large immediates",
	"R28": "g, the goroutine pointer",
	"R29": "frame pointer",
	"R30": "link register",
	"RSP": "stack pointer",
}

// TestARM64RegisterMapAvoidsReserved checks the x86->arm64 register mapping and
// the reserved scratch registers against the list above. The mapping is the
// single place the lowering decides which physical registers it may use, so
// asserting it directly is stronger than scanning any particular output.
func TestARM64RegisterMapAvoidsReserved(t *testing.T) {
	for idx, name := range armReg {
		if why, bad := reservedARM64[name]; bad {
			t.Errorf("armReg maps x86 index %d to %s, which is reserved (%s)", idx, name, why)
		}
	}
	for _, s := range []string{scratchAddr, scratchVal} {
		if why, bad := reservedARM64[s]; bad {
			t.Errorf("scratch register %s is reserved (%s)", s, why)
		}
	}

	// The scratch registers must also be distinct from every mapped register,
	// or a lowering would silently corrupt an allocated value.
	for idx, name := range armReg {
		if name == scratchAddr || name == scratchVal {
			t.Errorf("scratch register %s collides with armReg[%d]", name, idx)
		}
	}
	if scratchAddr == scratchVal {
		t.Errorf("scratch registers must be distinct, both are %s", scratchAddr)
	}

	// The scratch vector register must likewise sit outside the range x86's
	// XMM0-15 map onto (V0-V15).
	var vecIdx int
	if _, err := fmt.Sscanf(scratchVec, "V%d", &vecIdx); err != nil {
		t.Fatalf("scratchVec %q is not a V register", scratchVec)
	}
	if vecIdx < 16 {
		t.Errorf("scratchVec %s overlaps the XMM-mapped range V0-V15", scratchVec)
	}
}

// TestARM64RenameRejectsUnmappedRegisters checks that an x86 register with no
// arm64 mapping fails loudly rather than silently rendering as something else.
func TestARM64RenameRejectsUnmappedRegisters(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected a panic for an unmapped register index")
		} else if msg, ok := r.(string); !ok || !strings.Contains(msg, "unmapped") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	// RSP is x86 index 4, which avo never allocates and armReg deliberately omits.
	_ = rename(reg.RSP)
}
