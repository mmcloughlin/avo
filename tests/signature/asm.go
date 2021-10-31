//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"go/token"
	"go/types"
	"math/rand"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/gotypes"
	. "github.com/mmcloughlin/avo/reg"
)

var (
	seed = flag.Int64("seed", 0, "random seed")
	num  = flag.Int("num", 32, "number of test functions to generate")
)

func main() {
	flag.Parse()

	rand.Seed(*seed)

	for i := 0; i < *num; i++ {
		name := fmt.Sprintf("Signature%d", i)
		sig := RandomSignature()
		SignatureFunction(name, sig)
	}

	Generate()
}

func SignatureFunction(name string, sig *types.Signature) {
	// Declare the function.
	TEXT(name, 0, sig.String())

	// Write to the results. Otherwise, asmdecl would warn us.
	regsize := map[int64]Virtual{1: GP8(), 2: GP16(), 4: GP32(), 8: GP64()}
	rs := sig.Results()
	for i := 0; i < rs.Len(); i++ {
		r := rs.At(i)
		size := Sizes.Sizeof(r.Type())
		Store(regsize[size], ReturnIndex(i))
	}

	RET()
}

func RandomSignature() *types.Signature {
	p := RandomTuple()
	r := RandomTuple()
	return types.NewSignature(nil, p, r, false)
}

func RandomTuple() *types.Tuple {
	n := rand.Intn(5)
	vs := make([]*types.Var, n)
	for i := 0; i < n; i++ {
		t := RandomType()
		vs[i] = types.NewVar(token.NoPos, nil, "", t)
	}
	return types.NewTuple(vs...)
}

func RandomType() types.Type {
	accept := types.IsInteger | types.IsUnsigned
	for {
		t := types.Typ[rand.Intn(len(types.Typ))]
		info := t.Info()
		if info != 0 && (info&^accept) == 0 {
			return t
		}
	}
}
