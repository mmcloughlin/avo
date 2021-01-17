package api

import (
	"fmt"
	"strings"
)

const (
	// Package is the avo package import path.
	Package = "github.com/mmcloughlin/avo"

	// OperandType is the type used for operands.
	OperandType = "operand.Op"
)

// ImplicitRegister returns avo syntax for the given implicit register type (from Opcodes XML).
func ImplicitRegister(t string) string {
	r := strings.Replace(t, "mm", "", 1) // handles the "xmm0" type
	return fmt.Sprintf("reg.%s", strings.ToUpper(r))
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
