// Package arm64lower contains differential execution tests for the EXPERIMENTAL
// arm64 lowering printer. asm.go emits the same functions to both amd64 and
// arm64; the tests run each generated function against a pure-Go reference, so
// running the suite on each architecture verifies the lowering is
// semantics-preserving.
package arm64lower
