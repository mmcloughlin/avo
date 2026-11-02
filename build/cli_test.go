package build

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/mmcloughlin/avo/printer"
)

// TestOutFlagWritesFile checks that -out (without -arch) writes to the named
// file. The lazy-output plumbing added for -arch left goasm's writer at its
// stdout default, so Build must open the -out file rather than fall through to
// stdout.
func TestOutFlagWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.s")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	f := NewFlags(fs)
	if err := fs.Parse([]string{"-out", path}); err != nil {
		t.Fatal(err)
	}

	p := f.goasm.Build(printer.Config{})
	if p == nil {
		t.Fatal("no goasm output pass built for -out")
	}
	if f.goasm.w == os.Stdout {
		t.Fatal("-out was ignored: goasm still routed to stdout")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("-out did not create %s: %v", path, err)
	}
}

// TestArchFilename covers the base-path derivation used by -arch.
func TestArchFilename(t *testing.T) {
	if got, want := archFilename("seqdec.s", "arm64"), "seqdec_arm64.s"; got != want {
		t.Errorf("archFilename = %q, want %q", got, want)
	}
	if got, want := archFilename("dir/x.s", "amd64"), "dir/x_amd64.s"; got != want {
		t.Errorf("archFilename = %q, want %q", got, want)
	}
}
