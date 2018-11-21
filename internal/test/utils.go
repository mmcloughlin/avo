package test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Assembles asserts that the given assembly code passes the go assembler.
func Assembles(t *testing.T, asm []byte) {
	t.Helper()

	dir, clean := TempDir(t)
	defer clean()

	asmfilename := filepath.Join(dir, "asm.s")
	if err := ioutil.WriteFile(asmfilename, asm, 0600); err != nil {
		t.Fatal(err)
	}

	objfilename := filepath.Join(dir, "asm.o")

	goexec(t, "tool", "asm", "-e", "-o", objfilename, asmfilename)
}

// TempDir creates a temp directory. Returns the path to the directory and a
// cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "avo")
	if err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
}

// gobin returns a best guess path to the "go" binary.
func gobin() string {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	path := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return "go"
}

// goexec runs a "go" command and checks the output.
func goexec(t *testing.T, arg ...string) {
	t.Helper()
	cmd := exec.Command(gobin(), arg...)
	t.Logf("exec: %s", cmd.Args)
	b, err := cmd.CombinedOutput()
	t.Logf("output:\n%s\n", string(b))
	if err != nil {
		t.Fatal(err)
	}
}
