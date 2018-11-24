package gen

import "testing"

func TestBuilderInterfaces(t *testing.T) {
	var _ Builder = NewAsmTest
	var _ Builder = NewGoData
}
