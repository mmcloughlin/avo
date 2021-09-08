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
	Package

	WorkDir string // working directory for the test
	Latest  bool   // use latest version of the package

	repopath string // path the repo is cloned to
	cwd      string // working directory to execute commands in
}

// Run the test.
func (t *PackageTest) Run() {
	t.checkout()
	t.modinit()
	t.replaceavo()
	t.diff()
	t.generate()
	t.diff()
	t.test()
}

// checkout the code at the specified version.
func (t *PackageTest) checkout() {
	// Clone repo.
	dst := filepath.Join(t.WorkDir, t.Name())
	t.git("clone", "--quiet", t.CloneURL(), dst)
	t.repopath = dst
	t.cd(t.repopath)

	// Checkout specific version.
	if t.Latest {
		t.Log("using latest version")
		return
	}
	t.git("-C", t.repopath, "checkout", "--quiet", t.Version)
}

// modinit initializes the repo as a go module if it isn't one already.
func (t *PackageTest) modinit() {
	// Check if module path already exists.
	gomod := filepath.Join(t.repopath, "go.mod")
	if _, err := os.Stat(gomod); err == nil {
		t.Logf("already a module")
		return
	}

	// Initialize the module.
	t.gotool("mod", "init", t.ImportPath)
}

// replaceavo points all avo dependencies to the local version.
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
		dir, base := filepath.Split(path)
		if base != "go.mod" {
			return nil
		}
		t.cd(dir)
		t.gotool("mod", "tidy")
		t.gotool("get", "github.com/mmcloughlin/avo")
		t.gotool("mod", "edit", "-replace=github.com/mmcloughlin/avo="+avodir)
		t.gotool("mod", "download")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.cd(t.repopath)
}

// generate runs generate commands.
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
	t.gotool("test", t.TestPath())
}

// git runs a git command.
func (t *PackageTest) git(arg ...string) {
	test.Exec(t.T, "git", arg...)
}

// gotool runs a go command.
func (t *PackageTest) gotool(arg ...string) {
	cmd := exec.Command(test.GoTool(), arg...)
	cmd.Dir = t.cwd
	test.ExecCommand(t.T, cmd)
}

// cd sets the working directory.
func (t *PackageTest) cd(dir string) {
	t.cwd = dir
}
