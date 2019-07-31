package gen

import "testing"

func TestBuilderInterfaces(t *testing.T) {
	_ = []Builder{
		NewAsmTest,
		NewGoData,
		NewGoDataTest,
		NewCtors,
		NewCtorsTest,
		NewBuild,
		NewMOV,
	}
}
