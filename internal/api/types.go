package api

import (
	"path"
	"sort"
	"strings"
)

const (
	// Package is the avo package import path.
	Package = "github.com/mmcloughlin/avo"

	// OperandType is the type used for operands.
	OperandType = "operand.Op"
)

// ImportPath returns the full import path for an avo subpackage.
func ImportPath(pkg string) string {
	return path.Join(Package, pkg)
}

// ImplicitRegisterIdentifier maps an implicit register name to a string
// suitable for a related Go identifier.
func ImplicitRegisterIdentifier(r string) string {
	r = strings.Replace(r, "mm", "", 1) // handles the "xmm0" type
	return strings.ToUpper(r)
}

// ImplicitRegister returns avo syntax for the given implicit register type (from Opcodes XML).
func ImplicitRegister(r string) string {
	return "reg." + ImplicitRegisterIdentifier(r)
}

// OperandTypeIdentifier maps an operand type to a string suitable for a related
// Go identifier.
func OperandTypeIdentifier(t string) string {
	return strings.ToUpper(strings.ReplaceAll(t, "/", ""))
}

// CheckerName returns the name of the function that checks an operand of type t.
func CheckerName(t string) string {
	return "operand.Is" + OperandTypeIdentifier(t)
}

// ISAsIdentifier returns a string representation of an ISA list that suitable
// for use in a Go identifier.
func ISAsIdentifier(isas []string) string {
	if len(isas) == 0 {
		return "Base"
	}
	sorted := append([]string(nil), isas...)
	sort.Strings(sorted)
	ident := strings.Join(sorted, "_")
	ident = strings.ReplaceAll(ident, ".", "") // SSE4.1
	ident = strings.ReplaceAll(ident, "+", "") // MMX+
	return ident
}
