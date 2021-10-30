//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"go/types"
	"math/rand"
	"strings"

	. "github.com/mmcloughlin/avo/build"
)

var (
	seed = flag.Int64("seed", 0, "random seed")
	num  = flag.Int("num", 32, "number of test functions to generate")
)

func main() {
	flag.Parse()

	rand.Seed(*seed)

	for i := 0; i < *num; i++ {
		TEXT(fmt.Sprintf("Signature%d", i), 0, RandomSignature())
		RET()
	}

	Generate()
}

func RandomSignature() string {
	// Parameters.
	params := RandomTuple()
	sig := fmt.Sprintf("func(%s)", strings.Join(params, ", "))

	// Results.
	results := RandomTuple()
	switch len(results) {
	case 0:
		break
	case 1:
		sig += " " + results[0]
	default:
		sig += " (" + strings.Join(results, ", ") + ")"
	}

	return sig
}

func RandomTuple() []string {
	n := rand.Intn(5)
	vs := make([]string, n)
	for i := 0; i < n; i++ {
		vs[i] = RandomType()
	}
	return vs
}

func RandomType() string {
	for {
		t := types.Typ[rand.Intn(len(types.Typ))]
		info := t.Info()
		if info != 0 && (info&types.IsUntyped) == 0 {
			return t.String()
		}
	}
}
