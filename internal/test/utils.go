// Package test provides testing utilities.
package test

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var network = flag.Bool("net", false, "allow network access")

// RequiresNetwork declares that a test requires network access. The test is
// skipped if network access isn't enabled with the -net flag.
func RequiresNetwork(t *testing.T) {
	t.Helper()
	if !*network {
		t.Skip("requires network: enable with -net flag")
	}
}

// Assembles asserts that the given assembly code passes the go assembler.
func Assembles(t *testing.T, asm []byte) {
	t.Helper()

	dir, clean := TempDir(t)
	defer clean()

	asmfilename := filepath.Join(dir, "asm.s")
	if err := os.WriteFile(asmfilename, asm, 0o600); err != nil {
		t.Fatal(err)
	}

	objfilename := filepath.Join(dir, "asm.o")

	goexec(t, "tool", "asm", "-e", "-o", objfilename, asmfilename)
}

// TempDir creates a temp directory. Returns the path to the directory and a
// cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "avo")
	if err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
}

// ExecCommand executes the command, logging the command and output and failing
// the test on error.
func ExecCommand(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	t.Logf("exec: %s", cmd.Args)
	if cmd.Dir != "" {
		t.Logf("dir: %s", cmd.Dir)
	}
	b, err := cmd.CombinedOutput()
	t.Logf("output:\n%s\n", string(b))
	if err != nil {
		t.Fatal(err)
	}
}

// Exec executes the named program with the given arguments, logging the command
// and output and failing the test on error.
func Exec(t *testing.T, name string, arg ...string) {
	t.Helper()
	cmd := exec.Command(name, arg...)
	ExecCommand(t, cmd)
}

// GoTool returns a best guess path to the "go" binary.
func GoTool() string {
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
	Exec(t, GoTool(), arg...)
}

// Logger builds a logger that writes to the test object.
func Logger(tb testing.TB) *log.Logger {
	tb.Helper()
	return log.New(Writer(tb), "test", log.LstdFlags)
}

type writer struct {
	tb testing.TB
}

// Writer builds a writer that logs all writes to the test object.
func Writer(tb testing.TB) io.Writer {
	tb.Helper()
	return writer{tb}
}

func (w writer) Write(p []byte) (n int, err error) {
	w.tb.Log(string(p))
	return len(p), nil
}
