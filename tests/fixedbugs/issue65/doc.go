// Package issue65 is a regression test for a bug involving casting physical registers.
//
// Regression test for a bug where casting a physical register would give the
// error "non physical register found".
//
// See: https://github.com/mmcloughlin/avo/issues/65#issuecomment-576850145
package issue65

//go:generate go run asm.go -out issue65.s -stubs stub.go
