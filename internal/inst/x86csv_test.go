package inst

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

//go:generate curl --output testdata/x86.v0.2.csv https://raw.githubusercontent.com/golang/arch/master/x86/x86.v0.2.csv

const csvpath = "testdata/x86.v0.2.csv"

func LoadX86CSV(t *testing.T) []Instruction {
	t.Helper()
	f, err := os.Open(csvpath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	is, err := ReadFromX86CSV(f)
	if err != nil {
		t.Fatal(err)
	}

	return is
}

func TestExpandArg(t *testing.T) {
	cases := []struct {
		Arg   string
		Types []string
	}{
		{"imm8", []string{"imm8"}},
		{"r/m32", []string{"r32", "m32"}},
		{"r16", []string{"r16"}},
		{"mm1", []string{"mm"}},
		{"xmm1", []string{"xmm"}},
		{"xmmV", []string{"xmm"}},
		{"xmm2/m128", []string{"xmm", "m128"}},
		{"xmm2/m64", []string{"xmm", "m64"}},
		{"ymm1", []string{"ymm"}},
		{"ymmV", []string{"ymm"}},
		{"ymm2/m256", []string{"ymm", "m256"}},
	}
	for _, c := range cases {
		types := expandArg(c.Arg)
		if !reflect.DeepEqual(c.Types, types) {
			t.Errorf("expanded %v to %s expected %s", c.Arg, types, c.Types)
		}
	}
}

func TestX86CSVOperandTypes(t *testing.T) {
	t.Skip("have not handled all cases yet")

	is := LoadX86CSV(t)

	types := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, op := range f.Operands {
				types[op.Type] = true
			}
		}
	}

	for tipe := range types {
		if strings.Contains(tipe, "/") {
			t.Errorf("operand type %#v contains a slash (should be split)", tipe)
		}
	}
}

// TestCPUIDFlags helps catch any oddities in x86csv CPUID flags.
func TestX86CSVCPUIDFlags(t *testing.T) {
	is := LoadX86CSV(t)

	flags := map[string]bool{}
	for _, i := range is {
		for _, f := range i.Forms {
			for _, flag := range f.CPUID {
				flags[flag] = true
			}
		}
	}

	for flag := range flags {
		if strings.Contains(flag, " ") {
			t.Errorf("CPUID flag %#v contains whitespace", flag)
		}
	}
}
