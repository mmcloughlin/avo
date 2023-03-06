// The godata test relies on a huge generated file, so we limit it to a
// stress-test only build.

//go:build stress

package inst

//go:generate avogen -bootstrap -data ../data -output ztable_test.go godatatest
