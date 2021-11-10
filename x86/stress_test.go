// Constructors test that rely on huge generated files that bloat compile time
// are limited to stress-test mode.

//go:build stress
// +build stress

package x86

//go:generate avogen -output zstress_test.go ctorsstress
//go:generate avogen -output zbench_test.go ctorsbench
