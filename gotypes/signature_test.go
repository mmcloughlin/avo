package gotypes

import (
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestParseSignature(t *testing.T) {
	cases := []struct {
		Expr         string
		ExpectParams *types.Tuple
		ExpectReturn *types.Tuple
	}{
		{
			Expr: "func()",
		},
		{
			Expr: "func(x, y uint64)",
			ExpectParams: types.NewTuple(
				types.NewParam(token.NoPos, nil, "x", types.Typ[types.Uint64]),
				types.NewParam(token.NoPos, nil, "y", types.Typ[types.Uint64]),
			),
		},
		{
			Expr: "func(n int, s []string) byte",
			ExpectParams: types.NewTuple(
				types.NewParam(token.NoPos, nil, "n", types.Typ[types.Int]),
				types.NewParam(token.NoPos, nil, "s", types.NewSlice(types.Typ[types.String])),
			),
			ExpectReturn: types.NewTuple(
				types.NewParam(token.NoPos, nil, "", types.Typ[types.Byte]),
			),
		},
		{
			Expr: "func(x, y int) (x0, y0 int, s string)",
			ExpectParams: types.NewTuple(
				types.NewParam(token.NoPos, nil, "x", types.Typ[types.Int]),
				types.NewParam(token.NoPos, nil, "y", types.Typ[types.Int]),
			),
			ExpectReturn: types.NewTuple(
				types.NewParam(token.NoPos, nil, "x0", types.Typ[types.Int]),
				types.NewParam(token.NoPos, nil, "y0", types.Typ[types.Int]),
				types.NewParam(token.NoPos, nil, "s", types.Typ[types.String]),
			),
		},
	}
	for _, c := range cases {
		s, err := ParseSignature(c.Expr)
		if err != nil {
			t.Fatal(err)
		}
		if !TypesTuplesEqual(s.sig.Params(), c.ExpectParams) {
			t.Errorf("parameter mismatch\ngot %#v\nexpect %#v\n", s.sig.Params(), c.ExpectParams)
		}
		if !TypesTuplesEqual(s.sig.Results(), c.ExpectReturn) {
			t.Errorf("return value(s) mismatch\ngot %#v\nexpect %#v\n", s.sig.Results(), c.ExpectReturn)
		}
	}
}

func TestParseSignatureErrors(t *testing.T) {
	cases := []struct {
		Expr          string
		ErrorContains string
	}{
		{"idkjklol", "undeclared name"},
		{"struct{}", "not a function signature"},
		{"uint32(0xfeedbeef)", "should have nil value"},
	}
	for _, c := range cases {
		s, err := ParseSignature(c.Expr)
		if s != nil || err == nil || !strings.Contains(err.Error(), c.ErrorContains) {
			t.Errorf("expect error from expression %s\ngot: %s\nexpect substring: %s\n", c.Expr, err, c.ErrorContains)
		}
	}
}

func TypesTuplesEqual(a, b *types.Tuple) bool {
	if a.Len() != b.Len() {
		return false
	}
	n := a.Len()
	for i := 0; i < n; i++ {
		if !TypesVarsEqual(a.At(i), b.At(i)) {
			return false
		}
	}
	return true
}

func TypesVarsEqual(a, b *types.Var) bool {
	return a.Name() == b.Name() && types.Identical(a.Type(), b.Type())
}

func TestSignatureSizes(t *testing.T) {
	cases := []struct {
		Expr    string
		ArgSize int
	}{
		{"func()", 0},
		{"func(uint64) uint64", 16},
		{"func([7]byte) byte", 9},
		{"func(uint64, uint64) (uint64, uint64)", 32},
	}
	for _, c := range cases {
		s, err := ParseSignature(c.Expr)
		if err != nil {
			t.Fatal(err)
		}
		if s.Bytes() != c.ArgSize {
			t.Errorf("%s: size %d expected %d", s, s.Bytes(), c.ArgSize)
		}
	}
}
