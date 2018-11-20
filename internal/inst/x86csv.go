package inst

import (
	"io"
	"strconv"
	"strings"

	"golang.org/x/arch/x86/x86csv"
)

// ReadFromX86CSV reads instruction list from the Go "x86.csv" format.
func ReadFromX86CSV(r io.Reader) ([]Instruction, error) {
	c := x86csv.NewReader(r)

	rows, err := c.ReadAll()
	if err != nil {
		return nil, err
	}

	// Group by Go opcode.
	groups := map[string][]*x86csv.Inst{}
	for _, row := range rows {
		g := row.GoOpcode()
		groups[g] = append(groups[g], row)
	}

	var is []Instruction
	for opcode, group := range groups {
		i := Instruction{
			Opcode: opcode,
		}
		for _, row := range group {
			forms := formsFromRow(row)
			i.Forms = append(i.Forms, forms...)
		}
		is = append(is, i)
	}

	return is, nil
}

func formsFromRow(row *x86csv.Inst) []Form {
	var fs []Form
	signatures := argsToSignatures(row.GoArgs())
	for _, signature := range signatures {
		ops := make([]Operand, len(signature))
		for i, t := range signature {
			ops[i] = Operand{
				Type: t,
			}
		}
		f := Form{
			Operands: ops,
			CPUID:    splitCPUID(row.CPUID),
		}
		fs = append(fs, f)
	}
	return fs
}

func expandArg(arg string) []string {
	mmprefixes := []string{"", "x", "y"}

	for e := 8; e <= 512; e *= 2 {
		s := strconv.Itoa(e)
		switch arg {
		case "r/m" + s:
			return []string{"r" + s, "m" + s}
		case "r" + s + "V":
			return []string{"r" + s}
		}
		for _, p := range mmprefixes {
			if arg == p+"mm2/m"+s {
				return []string{p + "mm", "m" + s}
			}
		}
	}

	for _, p := range []string{"", "x", "y"} {
		switch arg {
		case p + "mm1", p + "mm2", p + "mmV":
			return []string{p + "mm"}
		}
	}

	return []string{arg}
}

func argsToSignatures(args []string) [][]string {
	n := len(args)
	if n == 0 {
		return [][]string{nil}
	}
	types := expandArg(args[n-1])
	var expanded [][]string
	for _, sub := range argsToSignatures(args[:n-1]) {
		for _, t := range types {
			f := make([]string, n-1, n)
			copy(f, sub)
			f = append(f, t)
			expanded = append(expanded, f)
		}
	}
	return expanded
}

// splitCPUID splits CPUID field into flags, plus handling a few oddities.
func splitCPUID(cpuid string) []string {
	switch cpuid {
	case "Both AES and AVX flags":
		return []string{"AES", "AVX"}
	case "HLE or RTM":
		return []string{"HLE/RTM"}
	}
	return strings.Split(cpuid, ",")
}
