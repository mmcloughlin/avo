package thirdparty

import (
	"flag"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mmcloughlin/avo/internal/test"
)

//go:generate go run make_workflow.go -pkgs packages.json -output ../../.github/workflows/thirdparty.yml

// Custom flags.
var (
	pkgsfilename = flag.String("pkgs", "", "packages configuration")
	preserve     = flag.Bool("preserve", false, "preserve working directories")
	latest       = flag.Bool("latest", false, "use latest versions of each package")
)

// TestPackages runs integration tests on all packages specified by packages
// file given on the command line.
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
		t.Run(pkg.ID(), func(t *testing.T) {
			dir, clean := test.TempDir(t)
			if !*preserve {
				defer clean()
			} else {
				t.Logf("working directory: %s", dir)
			}
			pt := PackageTest{
				T:       t,
				Package: pkg,
				WorkDir: dir,
				Latest:  *latest,
			}
			pt.Run()
		})
	}
}

// PackageTest executes an integration test based on a given third-party package.
type PackageTest struct {
	*testing.T
	*Package

	WorkDir string // working directory for the test
	Latest  bool   // use latest version of the package

	repopath string // path the repo is cloned to
}

// Run the test.
func (t *PackageTest) Run() {
	t.checkout()
	t.steps()
}

// checkout the code at the specified version.
func (t *PackageTest) checkout() {
	// Clone repo.
	dst := filepath.Join(t.WorkDir, t.Name())
	test.Exec(t.T, "git", "clone", "--quiet", t.Repository.CloneURL(), dst)
	t.repopath = dst

	// Checkout specific version.
	if t.Latest {
		t.Log("using latest version")
		return
	}
	test.Exec(t.T, "git", "-C", t.repopath, "checkout", "--quiet", t.Version)
}

func (t *PackageTest) steps() {
	// Determine the path to avo.
	_, self, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("failed to determine path to avo")
	}
	avodir := filepath.Join(filepath.Dir(self), "..", "..")

	// Run steps.
	c := &Context{
		AvoDirectory:        avodir,
		RepositoryDirectory: t.repopath,
	}

	for _, s := range t.Steps(c) {
		for _, command := range s.Commands {
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = filepath.Join(t.repopath, s.WorkingDirectory)
			test.ExecCommand(t.T, cmd)
		}
	}
}
