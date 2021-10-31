// Package signature tests handling of random function signatures.
package signature

//go:generate go run asm.go -out signature.s -stubs stub.go -seed 42 -num 256
