// Package allocfail is a regression test for issue 100 based on the original reported allocation failure.
//
// Based on the pull request https://github.com/klauspost/compress/pull/186 at
// c1f3cf132cd8e214b38cc16e418bf2e501ccda93 with the lines after "FIXME"
// comments re-activated and other minimal edits to make it work in this
// environment.
package allocfail

//go:generate go run asm.go -out allocfail.s -stubs stubs.go
