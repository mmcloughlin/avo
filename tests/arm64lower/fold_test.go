package arm64lower

import (
	"math/bits"
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestCopyFolds runs the copy-fold shapes (see shiftFolds and setccFolds in
// printer/arm64.go) against pure-Go references, including the shapes the
// folds must refuse.
//
//nolint:gocognit // table-driven differential test; splitting the cases would only add indirection, not clarity.
func TestCopyFolds(t *testing.T) {
	xs := []uint64{0, 1, 3, 63, 64, 65, 0xdeadbeefcafef00d, ^uint64(0), 0x8000000000000000, 127}
	for _, x := range xs {
		for _, n := range xs {
			if got, want := ShlCountFold(x, n), x<<(n&63)+x+n; got != want {
				t.Errorf("ShlCountFold(%#x, %#x) = %#x, want %#x", x, n, got, want)
			}
			if got, want := ShrCountFold32(x, n), uint64(uint32(x)>>(n&31))+x+n; got != want {
				t.Errorf("ShrCountFold32(%#x, %#x) = %#x, want %#x", x, n, got, want)
			}
			if got, want := RolCountFold(x, n), bits.RotateLeft64(x, int(n&63))+x+n; got != want {
				t.Errorf("RolCountFold(%#x, %#x) = %#x, want %#x", x, n, got, want)
			}
			if got, want := CountFoldRefusedRead(x, n), x<<(n&63)+n+n; got != want {
				t.Errorf("CountFoldRefusedRead(%#x, %#x) = %#x, want %#x", x, n, got, want)
			}
			cf := uint64(0)
			if x < 4 {
				cf = 1
			}
			if got, want := AdcAccumQ(x, n), n+cf; got != want {
				t.Errorf("AdcAccumQ(%#x, %#x) = %#x, want %#x", x, n, got, want)
			}
			ge := uint64(0)
			if int64(x) >= int64(n) {
				ge = 1
			}
			if got := SetGeZeroMov(x, n); got != ge {
				t.Errorf("SetGeZeroMov(%#x, %#x) = %d, want %d", x, n, got, ge)
			}
			if got := SetGeRefusedRead(x, n); got != ge {
				t.Errorf("SetGeRefusedRead(%#x, %#x) = %d, want %d", x, n, got, ge)
			}
		}
		if got, want := ShlCountSelf(x), x<<(x&63)+x; got != want {
			t.Errorf("ShlCountSelf(%#x) = %#x, want %#x", x, got, want)
		}
		if got, want := SarCopyFold(x), uint64(int64(x)>>3)+x; got != want {
			t.Errorf("SarCopyFold(%#x) = %#x, want %#x", x, got, want)
		}
		if got, want := ShlCopyFoldCX(x), x<<2+x; got != want {
			t.Errorf("ShlCopyFoldCX(%#x) = %#x, want %#x", x, got, want)
		}
	}
}

// TestCopyFoldsFired checks the generated arm64 text, so a fold that quietly
// stopped matching would fail here rather than pass vacuously through the
// differential tests above (which hold whether or not a fold fires).
func TestCopyFoldsFired(t *testing.T) {
	src, err := os.ReadFile("lower_arm64.s")
	if err != nil {
		t.Fatal(err)
	}
	body := func(name string) string {
		start := strings.Index(string(src), "TEXT ·"+name+"(SB)")
		if start < 0 {
			t.Fatalf("%s not found in lower_arm64.s", name)
		}
		rest := string(src)[start:]
		if end := strings.Index(rest, "\n\n// func"); end >= 0 {
			rest = rest[:end]
		}
		return regexp.MustCompile(`[ \t]+`).ReplaceAllString(rest, " ")
	}
	// Three distinct registers: count, source and destination all read from
	// where they were, neither copy emitted.
	threeOperand := regexp.MustCompile(`(?m)^\s*(LSL|LSRW|ROR)\s+R(\d+), R(\d+), R(\d+)$`)
	for _, name := range []string{"ShlCountFold", "ShrCountFold32", "RolCountFold"} {
		m := threeOperand.FindStringSubmatch(body(name))
		if m == nil || m[2] == m[3] || m[2] == m[4] || m[3] == m[4] {
			t.Errorf("%s: expected a shift with three distinct registers, got:\n%s", name, body(name))
		}
	}
	for _, c := range []struct{ name, want string }{
		{"SarCopyFold", "ASR $0x03, R0, R1"},
		{"ShlCopyFoldCX", "LSL $0x02, R0, R1"},
		{"AdcAccumQ", "CSINC HS, R1, R1, R1"},
		{"SetGe", "CSET GE, R2"},
		{"SetGeZeroMov", "CSET GE, R2"},
	} {
		if !strings.Contains(body(c.name), c.want) {
			t.Errorf("%s: expected %q in:\n%s", c.name, c.want, body(c.name))
		}
	}
	for _, c := range []struct{ name, want string }{
		{"CountFoldRefusedRead", "LSL R1, R"},
		{"ShlCountSelf", "LSL R1, R1, R1"},
		{"SetGeRefusedRead", "BFI $0, R16, $8, R"},
	} {
		if !strings.Contains(body(c.name), c.want) {
			t.Errorf("%s: expected the fold refused, with %q in:\n%s", c.name, c.want, body(c.name))
		}
	}
}
