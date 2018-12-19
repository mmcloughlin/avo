package gotypes

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"strconv"
)

type Signature struct {
	pkg     *types.Package
	sig     *types.Signature
	params  *Tuple
	results *Tuple
}

func NewSignature(pkg *types.Package, sig *types.Signature) *Signature {
	s := &Signature{
		pkg: pkg,
		sig: sig,
	}
	s.init()
	return s
}

func NewSignatureVoid() *Signature {
	return NewSignature(nil, types.NewSignature(nil, nil, nil, false))
}

func ParseSignature(expr string) (*Signature, error) {
	return ParseSignatureInPackage(nil, expr)
}

func ParseSignatureInPackage(pkg *types.Package, expr string) (*Signature, error) {
	tv, err := types.Eval(token.NewFileSet(), pkg, token.NoPos, expr)
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
	return NewSignature(pkg, s), nil
}

func (s *Signature) Params() *Tuple { return s.params }

func (s *Signature) Results() *Tuple { return s.results }

func (s *Signature) Bytes() int { return s.Params().Bytes() + s.Results().Bytes() }

func (s *Signature) String() string {
	var buf bytes.Buffer
	types.WriteSignature(&buf, s.sig, types.RelativeTo(s.pkg))
	return buf.String()
}

func (s *Signature) init() {
	p := s.sig.Params()
	r := s.sig.Results()

	// Compute parameter offsets.
	vs := tuplevars(p)
	vs = append(vs, types.NewParam(token.NoPos, nil, "sentinel", types.Typ[types.Uint64]))
	paramsoffsets := Sizes.Offsetsof(vs)
	paramssize := paramsoffsets[p.Len()]
	s.params = newTuple(p, paramsoffsets, paramssize, "arg")

	// Result offsets.
	vs = tuplevars(r)
	resultsoffsets := Sizes.Offsetsof(vs)
	for i := range resultsoffsets {
		resultsoffsets[i] += paramssize
	}
	resultssize := Sizes.Sizeof(types.NewStruct(vs, nil))
	s.results = newTuple(r, resultsoffsets, resultssize, "ret")
}

type Tuple struct {
	components []Component
	byname     map[string]Component
	size       int
}

func newTuple(t *types.Tuple, offsets []int64, size int64, defaultprefix string) *Tuple {
	tuple := &Tuple{
		byname: map[string]Component{},
		size:   int(size),
	}
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)
		name := v.Name()
		if name == "" {
			name = defaultprefix
			if i > 0 {
				name += strconv.Itoa(i)
			}
		}
		c := NewComponent(name, v.Type(), int(offsets[i]))
		tuple.components = append(tuple.components, c)
		if v.Name() != "" {
			tuple.byname[v.Name()] = c
		}
	}
	return tuple
}

func (t *Tuple) Lookup(name string) Component {
	e := t.byname[name]
	if e == nil {
		return componenterr(fmt.Sprintf("unknown variable \"%s\"", name))
	}
	return e
}

func (t *Tuple) At(i int) Component { return t.components[i] }

func (t *Tuple) Bytes() int { return t.size }

func tuplevars(t *types.Tuple) []*types.Var {
	vs := make([]*types.Var, t.Len())
	for i := 0; i < t.Len(); i++ {
		vs[i] = t.At(i)
	}
	return vs
}
