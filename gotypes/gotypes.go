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
	offset int
	typ    types.Type
	err    error
}

func NewComponentFromVar(v *types.Var) Component {
	return &component{
		name: v.Name(),
	}
}

func (c *component) Resolve() (*Basic, error) {
	b, ok := c.typ.(*types.Basic)
	if !ok {
		return nil, errors.New("not basic type")
	}
	if b.Kind() == types.String {
		return nil, errors.New("string types are handled specially")
	}
	return &Basic{
		Addr: operand.NewParamAddr(c.name, c.offset),
		Type: b,
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
	var s *component
	*s = *c
	s.name += suffix
	s.offset += offset
	s.typ = t
	return s
}

// TODO(mbm): gotypes.Component handling for structs
// TODO(mbm): gotypes.Component handling for complex

type Signature struct {
	sig *types.Signature
}

func newSignature(sig *types.Signature) *Signature {
	return &Signature{
		sig: sig,
	}
}

func ParseSignature(expr string) (*Signature, error) {
	tv, err := types.Eval(token.NewFileSet(), nil, token.NoPos, expr)
	if err != nil {
		return nil, err
	}
	if tv.Value != nil {
		return nil, errors.New("signature expression should have nil value")
	}
	s, ok := tv.Type.(*types.Signature)
	if !ok {
		return nil, errors.New("provided type is not a function signature")
	}
	return newSignature(s), nil
}

func (s *Signature) Parameter(name string) Component {
	p := s.param(name)
	if p == nil {
		return componenterr("unknown parameter")
	}
	return NewComponentFromVar(p)
}

func (s *Signature) param(name string) *types.Var {
	t := s.sig.Params()
	for i := 0; i < t.Len(); i++ {
		p := t.At(i)
		if p.Name() == name {
			return p
		}
	}
	return nil
}

func isslice(t types.Type) bool {
	_, ok := t.(*types.Slice)
	return ok
}

func isstring(t types.Type) bool {
	b, ok := t.(*types.Basic)
	return ok && b.Kind() == types.String
}
