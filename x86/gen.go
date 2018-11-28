package x86

import "errors"

var ErrBadOperandTypes = errors.New("bad operand types")

//go:generate avogen -output zctors.go ctors
//go:generate avogen -output zctors_test.go ctorstest
