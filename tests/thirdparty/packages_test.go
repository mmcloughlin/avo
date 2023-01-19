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

//go:generate go run make_workflow.go -suite suite.json -output ../../.github/workflows/packages.yml

// Custom flags.
var (
	suitefilename = flag.String("suite", "", "projects configuration")
	preserve      = flag.Bool("preserve", false, "preserve working directories")
	latest        = flag.Bool("latest", false, "use latest versions of each project")
)

// TestPackages runs integration tests on all packages specified by projects
// file given on the command line.
func TestPackages(t *testing.T) {
	// Load suite.
	if *suitefilename == "" {
		t.Skip("no suite specified")
	}

	s, err := LoadSuiteFile(*suitefilename)
	if err != nil {
		t.Fatal(err)
	}

	for _, tst := range s.Projects.Tests() {
		tst := tst // scopelint
		t.Run(tst.ID(), func(t *testing.T) {
			if tst.Project.Skip() {
				t.Skipf("skip: %s", tst.Project.Reason())
			}
			dir, clean := test.TempDir(t)
			if !*preserve {
				defer clean()
			} else {
				t.Logf("working directory: %s", dir)
			}
			pt := PackageTest{
				T:       t,
				Test:    tst,
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
	*Test

	WorkDir string // working directory for the test
	Latest  bool   // use latest version of the project

	repopath string // path the repo is cloned to
}

// Run the test.
func (t *PackageTest) Run() {
	t.checkout()
	t.validate()
	t.steps()
}

// checkout the code at the specified version.
func (t *PackageTest) checkout() {
	// Determine the version we want to checkout.
	version := t.Project.Version
	if t.Latest {
		version = t.Project.DefaultBranch
	}

	// Clone. Use a shallow clone to speed up large repositories.
	t.repopath = filepath.Join(t.WorkDir, t.Name())
	test.Exec(t.T, "git", "init", t.repopath)
	test.Exec(t.T, "git", "-C", t.repopath, "remote", "add", "origin", t.Project.Repository.CloneURL())
	test.Exec(t.T, "git", "-C", t.repopath, "fetch", "--depth=1", "origin", version)
	test.Exec(t.T, "git", "-C", t.repopath, "checkout", "FETCH_HEAD")
}

// validate the test configuration relative to the checked out project.
func (t *PackageTest) validate() {
	// Confirm expected directories exist.
	expect := map[string]string{
		"package": t.Package.SubPackage,
		"root":    t.Package.WorkingDirectory(),
	}
	for name, subdir := range expect {
		if subdir == "" {
			continue
		}
		path := filepath.Join(t.repopath, subdir)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s directory: %s", name, err)
		}
	}
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

	for _, s := range t.Package.Steps(c) {
		for _, command := range s.Commands {
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = filepath.Join(t.repopath, s.WorkingDirectory)
			test.ExecCommand(t.T, cmd)
		}
	}
}
