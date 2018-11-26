package x86

import "errors"

var ErrBadOperandTypes = errors.New("bad operand types")

//go:generate avogen -output zconstructors.go constructors
