package api

import (
	"fmt"
	"strings"
)

const (
	Package     = "github.com/mmcloughlin/avo"
	OperandType = "operand.Op"
)

// ImplicitRegister returns avo syntax for the given implicit register type (from Opcodes XML).
func ImplicitRegister(t string) string {
	r := strings.Replace(t, "mm", "", 1) // handles the "xmm0" type
	return fmt.Sprintf("reg.%s", strings.ToUpper(r))
}

// CheckerName returns the name of the function that checks an operand of type t.
func CheckerName(t string) string {
	return "operand.Is" + strings.ToUpper(strings.ReplaceAll(t, "/", ""))
}
