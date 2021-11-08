// The godata test relies on a huge generated file, so we limit it to test-only
// builds with the test build tag.

//go:build test
// +build test

package inst

//go:generate avogen -bootstrap -data ../data -output ztable_test.go godatatest
//go:generate go test -tags=test
