// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	Package("github.com/mmcloughlin/avo/tests/fixedbugs/issue61")
	Implement("private")
	RET()
	Generate()
}
