// The full constructors test relies on a huge generated file, so we limit it to
// test-only builds with the test build tag.

//go:build test
// +build test

package x86

import (
	"reflect"
	"testing"

	"github.com/mmcloughlin/avo/ir"
)

func AssertInstructionEqual(t *testing.T, got, expect *ir.Instruction) {
	t.Helper()
	if !reflect.DeepEqual(got, expect) {
		t.Fatal("mismatch")
	}
}
