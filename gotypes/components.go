package gotypes

import (
	"errors"
	"go/token"
	"go/types"
	"strconv"

	"github.com/mmcloughlin/avo/operand"
)

var Sizes = types.SizesFor("gc", "amd64")

type Basic struct {
	Addr operand.Mem
	Type *types.Basic
}

type Component interface {
	Resolve() (*Basic, error)
	Base() Component
	Len() Component
	Cap() Component
	Index(int) Component
}

type componenterr string

func (c componenterr) Error() string            { return string(c) }
func (c componenterr) Resolve() (*Basic, error) { return nil, c }
func (c componenterr) Base() Component          { return c }
func (c componenterr) Len() Component           { return c }
func (c componenterr) Cap() Component           { return c }
func (c componenterr) Index(int) Component      { return c }

type component struct {
	name   string
	typ    types.Type
	offset int
}

func NewComponent(name string, t types.Type, offset int) Component {
	return &component{
		name:   name,
		typ:    t,
		offset: offset,
	}
}

func NewComponentFromVar(v *types.Var, offset int) Component {
	return NewComponent(v.Name(), v.Type(), offset)
}

func (c *component) Resolve() (*Basic, error) {
	if !isprimitive(c.typ) {
		return nil, errors.New("component is not primitive")
	}
	return &Basic{
		Addr: operand.NewParamAddr(c.name, c.offset),
		Type: c.typ.(*types.Basic),
	}, nil
}

// Reference: https://github.com/golang/go/blob/50bd1c4d4eb4fac8ddeb5f063c099daccfb71b26/src/reflect/value.go#L1800-L1804
//
//	type SliceHeader struct {
//		Data uintptr
//		Len  int
//		Cap  int
//	}
//
var slicehdroffsets = Sizes.Offsetsof([]*types.Var{
	types.NewField(token.NoPos, nil, "Data", types.Typ[types.Uintptr], false),
	types.NewField(token.NoPos, nil, "Len", types.Typ[types.Int], false),
	types.NewField(token.NoPos, nil, "Cap", types.Typ[types.Int], false),
})

func (c *component) Base() Component {
	if !isslice(c.typ) && !isstring(c.typ) {
		return componenterr("only slices and strings have base pointers")
	}
	return c.sub("_base", int(slicehdroffsets[0]), types.Typ[types.Uintptr])
}

func (c *component) Len() Component {
	if !isslice(c.typ) && !isstring(c.typ) {
		return componenterr("only slices and strings have length fields")
	}
	return c.sub("_len", int(slicehdroffsets[1]), types.Typ[types.Int])
}

func (c *component) Cap() Component {
	if !isslice(c.typ) {
		return componenterr("only slices have capacity fields")
	}
	return c.sub("_cap", int(slicehdroffsets[2]), types.Typ[types.Int])
}

func (c *component) Index(i int) Component {
	a, ok := c.typ.(*types.Array)
	if !ok {
		return componenterr("not array type")
	}
	if int64(i) >= a.Len() {
		return componenterr("array index out of bounds")
	}
	// Reference: https://github.com/golang/tools/blob/bcd4e47d02889ebbc25c9f4bf3d27e4124b0bf9d/go/analysis/passes/asmdecl/asmdecl.go#L482-L494
	//
	//		case asmArray:
	//			tu := t.Underlying().(*types.Array)
	//			elem := tu.Elem()
	//			// Calculate offset of each element array.
	//			fields := []*types.Var{
	//				types.NewVar(token.NoPos, nil, "fake0", elem),
	//				types.NewVar(token.NoPos, nil, "fake1", elem),
	//			}
	//			offsets := arch.sizes.Offsetsof(fields)
	//			elemoff := int(offsets[1])
	//			for i := 0; i < int(tu.Len()); i++ {
	//				cc = appendComponentsRecursive(arch, elem, cc, suffix+"_"+strconv.Itoa(i), i*elemoff)
	//			}
	//
	elem := a.Elem()
	elemsize := int(Sizes.Sizeof(types.NewArray(elem, 2)) - Sizes.Sizeof(types.NewArray(elem, 1)))
	return c.sub("_"+strconv.Itoa(i), i*elemsize, elem)
}

func (c *component) sub(suffix string, offset int, t types.Type) *component {
	s := *c
	s.name += suffix
	s.offset += offset
	s.typ = t
	return &s
}

// TODO(mbm): gotypes.Component handling for structs
// TODO(mbm): gotypes.Component handling for complex64/128

func isslice(t types.Type) bool {
	_, ok := t.(*types.Slice)
	return ok
}

func isstring(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Kind() == types.String
}

// isprimitive returns true if the type cannot be broken into components.
func isprimitive(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && (b.Info()&(types.IsString|types.IsComplex)) == 0
}
