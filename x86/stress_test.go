//go:build stress
// +build stress

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
