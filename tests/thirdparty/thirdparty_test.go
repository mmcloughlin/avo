package thirdparty

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mmcloughlin/avo/internal/test"
)

// Custom flags.
var (
	pkgsfilename = flag.String("pkgs", "", "packages configuration")
	preserve     = flag.Bool("preserve", false, "preserve working directories")
)

func TestPackages(t *testing.T) {
	// Load packages.
	if *pkgsfilename == "" {
		t.Skip("no packages specified")
	}

	pkgs, err := LoadPackagesFile(*pkgsfilename)
	if err != nil {
		t.Fatal(err)
	}

	for _, pkg := range pkgs {
		pkg := pkg // scopelint
		t.Run(pkg.Name(), func(t *testing.T) {
			dir, clean := test.TempDir(t)
			if !*preserve {
				defer clean()
			} else {
				t.Logf("working directory: %s", dir)
			}
			pt := PackageTest{
				T:       t,
				Package: pkg,
				workdir: dir,
			}
			pt.Run()
		})
	}
}

type PackageTest struct {
	*testing.T
	Package

	workdir  string
	repopath string
}

func (t *PackageTest) Run() {
	t.checkout()
	t.modinit()
	t.replaceavo()
	t.diff()
	t.generate()
	t.diff()
	t.test()
}

func (t *PackageTest) checkout() {
	// Clone repo.
	dst := t.path(t.Name())
	t.git("clone", "--quiet", t.CloneURL(), dst)
	t.repopath = dst

	// Checkout specific version.
	t.git("-C", t.repopath, "checkout", "--quiet", t.Version)
}

func (t *PackageTest) modinit() {
	// Check if module path already exists.
	gomod := filepath.Join(t.repopath, "go.mod")
	if _, err := os.Stat(gomod); err == nil {
		t.Logf("already a module")
		return
	}

	// Initialize the module.
	cmd := exec.Command("go", "mod", "init", t.ImportPath)
	cmd.Dir = t.repopath
	test.ExecCommand(t.T, cmd)
}

func (t *PackageTest) replaceavo() {
	// Determine the path to avo.
	_, self, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("failed to determine path to avo")
	}
	avodir := filepath.Join(filepath.Dir(self), "..", "..")

	// Edit all go.mod files in the repo.
	err := filepath.Walk(t.repopath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) != "go.mod" {
			return nil
		}
		test.Exec(t.T, "go", "mod", "edit", "-replace=github.com/mmcloughlin/avo="+avodir, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (t *PackageTest) generate() {
	if len(t.Generate) == 0 {
		t.Fatal("no commands specified")
	}
	for _, args := range t.Generate {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = filepath.Join(t.repopath, t.Dir)
		test.ExecCommand(t.T, cmd)
	}
}

// diff runs git diff on the repository.
func (t *PackageTest) diff() {
	t.git("-C", t.repopath, "diff")
}

// test runs go test.
func (t *PackageTest) test() {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = t.repopath
	test.ExecCommand(t.T, cmd)
}

func (t *PackageTest) path(rel string) string {
	return filepath.Join(t.workdir, rel)
}

func (t *PackageTest) git(arg ...string) {
	test.Exec(t.T, "git", arg...)
}
