package gotypes

import (
	"go/types"
	"testing"
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
	c := NewComponent("primitive", typ, 0)
	if _, err := c.Resolve(); err != nil {
		t.Errorf("expected type %s to be primitive: got error '%s'", typ, err)
	}
}
