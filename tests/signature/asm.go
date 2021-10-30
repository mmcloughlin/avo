package main

import (
	"flag"
	"fmt"

	. "github.com/mmcloughlin/avo/build"
)

var (
	// seed = flag.Int64("seed", 0, "random seed")
	num = flag.Int("num", 32, "number of test functions to generate")
)

func main() {
	flag.Parse()

	for i := 0; i < *num; i++ {
		TEXT(fmt.Sprintf("Signature%d", i), 0, "func()")
		RET()
	}

	Generate()
}
