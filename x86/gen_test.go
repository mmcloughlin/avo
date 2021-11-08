// The full constructors test relies on a huge generated file, so we limit it to
// test-only builds with the test build tag.

//go:build test
// +build test

package x86

//go:generate avogen -output zctors_test.go ctorstest
