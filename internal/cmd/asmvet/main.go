// Command asmvet checks for correctness of Go assembly.
//
// Standalone version of the assembly checks in go vet.
package main

import (
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(
		asmdecl.Analyzer,
		framepointer.Analyzer,
	)
}
