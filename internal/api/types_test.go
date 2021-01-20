package api

import (
	"go/token"
	"testing"

	"github.com/mmcloughlin/avo/internal/inst"
)

func TestISAsIdentifier(t *testing.T) {
	for _, isas := range inst.ISACombinations(inst.Instructions) {
		ident := ISAsIdentifier(isas)
		if !token.IsIdentifier(ident) {
			t.Errorf("expected %q to be an identifier", ident)
		}
	}
}
