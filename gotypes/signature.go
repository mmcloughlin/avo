package gotypes

import (
	"errors"
	"go/token"
	"go/types"
	"strconv"
)

type Signature struct {
	sig     *types.Signature
	params  *Tuple
	results *Tuple
}

func NewSignature(sig *types.Signature) *Signature {
	s := &Signature{
		sig: sig,
	}
	s.init()
	return s
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
	return NewSignature(s), nil
}

func (s *Signature) Params() *Tuple { return s.params }

func (s *Signature) Results() *Tuple { return s.results }

func (s *Signature) Bytes() int { return s.Params().Bytes() + s.Results().Bytes() }

func (s *Signature) init() {
	p := s.sig.Params()
	r := s.sig.Results()

	// Compute offsets of parameters and return values.
	vs := tuplevars(p)
	vs = append(vs, tuplevars(r)...)
	vs = append(vs, types.NewParam(token.NoPos, nil, "sentinel", types.Typ[types.Uint64]))
	offsets := Sizes.Offsetsof(vs)

	// Build components for them.
	s.params = newTuple(p, offsets, "arg")
	s.results = newTuple(r, offsets[p.Len():], "ret")
}

type Tuple struct {
	components []Component
	byname     map[string]Component
	size       int
}

func newTuple(t *types.Tuple, offsets []int64, defaultprefix string) *Tuple {
	tuple := &Tuple{
		byname: map[string]Component{},
		size:   int(offsets[t.Len()]),
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

func (t *Tuple) Lookup(name string) Component { return t.byname[name] }

func (t *Tuple) At(i int) Component { return t.components[i] }

func (t *Tuple) Bytes() int { return t.size }

func tuplevars(t *types.Tuple) []*types.Var {
	vs := make([]*types.Var, t.Len())
	for i := 0; i < t.Len(); i++ {
		vs[i] = t.At(i)
	}
	return vs
}
