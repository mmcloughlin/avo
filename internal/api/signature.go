package api

import (
	"fmt"
	"strconv"
	"strings"
)

// Signature provides access to details about the signature of an instruction function.
type Signature interface {
	ParameterList() string
	Arguments() string
	ParameterName(int) string
	ParameterSlice() string
	Length() string
}

// ArgsList builds a signature for a function with the named parameters.
func ArgsList(args []string) Signature {
	return argslist(args)
}

type argslist []string

func (a argslist) ParameterList() string      { return strings.Join(a, ", ") + " " + OperandType }
func (a argslist) Arguments() string          { return strings.Join(a, ", ") }
func (a argslist) ParameterName(i int) string { return a[i] }
func (a argslist) ParameterSlice() string {
	return fmt.Sprintf("[]%s{%s}", OperandType, strings.Join(a, ", "))
}
func (a argslist) Length() string { return strconv.Itoa(len(a)) }

// Variadic is the signature for a variadic function with the named argument slice.
func Variadic(name string) Signature {
	return variadic{name: name}
}

type variadic struct {
	name string
}

func (v variadic) ParameterList() string      { return v.name + " ..." + OperandType }
func (v variadic) Arguments() string          { return v.name + "..." }
func (v variadic) ParameterName(i int) string { return fmt.Sprintf("%s[%d]", v.name, i) }
func (v variadic) ParameterSlice() string     { return v.name }
func (v variadic) Length() string             { return fmt.Sprintf("len(%s)", v.name) }

// Niladic is the signature for a function with no arguments.
func Niladic() Signature {
	return niladic{}
}

type niladic struct{}

func (n niladic) ParameterList() string      { return "" }
func (n niladic) Arguments() string          { return "" }
func (n niladic) ParameterName(i int) string { panic("niladic function has no parameters") }
func (n niladic) ParameterSlice() string     { return fmt.Sprintf("[]%s{}", OperandType) }
func (n niladic) Length() string             { return "0" }
