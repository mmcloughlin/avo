package gotypes

import (
	"go/types"
	"strings"
	"testing"

	"github.com/mmcloughlin/avo/reg"

	"github.com/mmcloughlin/avo/operand"
)

func TestBasicKindsArePrimitive(t *testing.T) {
	kinds := []types.BasicKind{
		types.Bool,
		types.Int,
		types.Int8,
		types.Int16,
		types.Int32,
		types.Int64,
		types.Uint,
		types.Uint8,
		types.Uint16,
		types.Uint32,
		types.Uint64,
		types.Uintptr,
		types.Float32,
		types.Float64,
	}
	for _, k := range kinds {
		AssertPrimitive(t, types.Typ[k])
	}
}

func TestPointersArePrimitive(t *testing.T) {
	typ := types.NewPointer(types.Typ[types.Uint32])
	AssertPrimitive(t, typ)
}

func AssertPrimitive(t *testing.T, typ types.Type) {
	c := NewComponent(typ, operand.NewParamAddr("primitive", 0))
	if _, err := c.Resolve(); err != nil {
		t.Errorf("expected type %s to be primitive: got error '%s'", typ, err)
	}
}

func TestComponentErrors(t *testing.T) {
	comp := NewComponent(types.Typ[types.Uint32], operand.Mem{})
	cases := []struct {
		Component      Component
		ErrorSubstring string
	}{
		{comp.Base(), "only slices and strings"},
		{comp.Len(), "only slices and strings"},
		{comp.Cap(), "only slices have"},
		{comp.Real(), "only complex"},
		{comp.Imag(), "only complex"},
		{comp.Index(42), "not array type"},
		{comp.Field("a"), "not struct type"},
		{comp.Dereference(reg.R12), "not pointer type"},
	}
	for _, c := range cases {
		_, err := c.Component.Resolve()
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), c.ErrorSubstring) {
			t.Fatalf("error message %q; expected substring %q", err.Error(), c.ErrorSubstring)
		}
	}
}
