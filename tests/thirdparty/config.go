// Package thirdparty executes integration tests based on third-party projects that use avo.
package thirdparty

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// GithubRepository specifies a repository on github.
type GithubRepository struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (r GithubRepository) String() string {
	return path.Join(r.Owner, r.Name)
}

// URL returns the Github repository URL.
func (r GithubRepository) URL() string {
	return fmt.Sprintf("https://github.com/%s", r)
}

// CloneURL returns the git clone URL.
func (r GithubRepository) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s.git", r)
}

// Metadata about the repository.
type Metadata struct {
	// Repository description.
	Description string `json:"description,omitempty"`

	// Homepage URL. Not the same as the Github page.
	Homepage string `json:"homepage,omitempty"`

	// Number of Github stars.
	Stars int `json:"stars,omitempty"`
}

// Project defines an integration test based on a third-party project using avo.
type Project struct {
	// Repository for the project. At the moment, all projects are available on
	// github.
	Repository GithubRepository `json:"repository"`

	// Repository metadata.
	Metadata Metadata `json:"metadata"`

	// Default git branch. This is used when testing against the latest version.
	DefaultBranch string `json:"default_branch,omitempty"`

	// Version as a git sha, tag or branch.
	Version string `json:"version"`

	// If the project test has a known problem, record it by setting this to a
	// non-zero avo issue number.  If set, the project will be skipped in
	// testing.
	KnownIssue int `json:"known_issue,omitempty"`

	// Packages within the project to test.
	Packages []*Package `json:"packages"`
}

func (p *Project) defaults(set bool) {
	for _, pkg := range p.Packages {
		pkg.defaults(set)
	}
}

// Validate project definition.
func (p *Project) Validate() error {
	if p.DefaultBranch == "" {
		return errors.New("missing default branch")
	}
	if p.Version == "" {
		return errors.New("missing version")
	}
	if len(p.Packages) == 0 {
		return errors.New("missing packages")
	}
	for _, pkg := range p.Packages {
		if err := pkg.Validate(); err != nil {
			return fmt.Errorf("package %s: %w", pkg.Name(), err)
		}
	}
	return nil
}

// ID returns an identifier for the project.
func (p *Project) ID() string {
	return strings.ReplaceAll(p.Repository.String(), "/", "-")
}

// Skip reports whether the project test should be skipped. If skipped, a known
// issue will be set.
func (p *Project) Skip() bool {
	return p.KnownIssue != 0
}

// Reason returns the reason why the test is skipped.
func (p *Project) Reason() string {
	return fmt.Sprintf("https://github.com/mmcloughlin/avo/issues/%d", p.KnownIssue)
}

// Step represents a set of commands to run as part of the testing plan for a
// third-party package.
type Step struct {
	Name             string   `json:"name,omitempty"`
	WorkingDirectory string   `json:"dir,omitempty"`
	Commands         []string `json:"commands"`
}

// Validate step parameters.
func (s *Step) Validate() error {
	if s.Name == "" {
		return errors.New("missing name")
	}
	if len(s.Commands) == 0 {
		return errors.New("missing commands")
	}
	return nil
}

// Package defines an integration test for a package within a project.
type Package struct {
	// Sub-package within the project under test. All file path references will
	// be relative to this directory. If empty the root of the repository is
	// used.
	SubPackage string `json:"pkg,omitempty"`

	// Path to the module file for the avo generator package. This is necessary
	// so the integration test can insert replace directives to point at the avo
	// version under test.
	Module string `json:"module"`

	// Setup steps. These run prior to the insertion of avo replace directives,
	// therefore should be used if it's necessary to initialize new go modules
	// within the repository.
	Setup []*Step `json:"setup,omitempty"`

	// Steps to run the avo code generator.
	Generate []*Step `json:"generate"`

	// Test steps. If empty, defaults to "go test ./...".
	Test []*Step `json:"test,omitempty"`
}

// defaults sets or removes default field values.
func (p *Package) defaults(set bool) {
	for _, stage := range []struct {
		Steps       []*Step
		DefaultName string
	}{
		{p.Setup, "Setup"},
		{p.Generate, "Generate"},
		{p.Test, "Test"},
	} {
		if len(stage.Steps) == 1 {
			stage.Steps[0].Name = applydefault(set, stage.Steps[0].Name, stage.DefaultName)
		}
	}
}

func applydefault(set bool, s, def string) string {
	switch {
	case set && s == "":
		return def
	case !set && s == def:
		return ""
	default:
		return s
	}
}

// Validate package definition.
func (p *Package) Validate() error {
	if p.Module == "" {
		return errors.New("missing module")
	}
	if len(p.Generate) == 0 {
		return errors.New("no generate commands")
	}

	stages := map[string][]*Step{
		"setup":    p.Setup,
		"generate": p.Generate,
		"test":     p.Test,
	}
	for name, steps := range stages {
		for _, s := range steps {
			if err := s.Validate(); err != nil {
				return fmt.Errorf("%s step: %w", name, err)
			}
		}
	}

	return nil
}

// Name of the package.
func (p *Package) Name() string {
	if p.IsRoot() {
		return "root"
	}
	return p.SubPackage
}

// IsRoot reports whether the package is the root of the containing project.
func (p *Package) IsRoot() bool {
	return p.SubPackage == ""
}

// Context specifies execution environment parameters for a third-party test.
type Context struct {
	// Path to the avo version under test.
	AvoDirectory string

	// Path to the checked out third-party repository.
	RepositoryDirectory string
}

// Steps generates the list of steps required to execute the integration test
// for this package. Context specifies execution environment parameters.
func (p *Package) Steps(c *Context) []*Step {
	var steps []*Step

	// Optional setup.
	steps = append(steps, p.Setup...)

	// Replace avo dependency.
	const invalid = "v0.0.0-00010101000000-000000000000"
	moddir := filepath.Dir(p.Module)
	modfile := filepath.Base(p.Module)
	steps = append(steps, &Step{
		Name:             "Avo Module Replacement",
		WorkingDirectory: moddir,
		Commands: []string{
			"go mod edit -modfile=" + modfile + " -require=github.com/mmcloughlin/avo@" + invalid,
			"go mod edit -modfile=" + modfile + " -replace=github.com/mmcloughlin/avo=" + c.AvoDirectory,
			"go mod tidy -modfile=" + modfile,
		},
	})

	// Run generation.
	steps = append(steps, p.Generate...)

	// Display changes.
	steps = append(steps, &Step{
		Name:     "Diff",
		Commands: []string{"git diff"},
	})

	// Tests.
	if len(p.Test) > 0 {
		steps = append(steps, p.Test...)
	} else {
		steps = append(steps, &Step{
			Name: "Test",
			Commands: []string{
				"go test ./...",
			},
		})
	}

	// Prepend sub-directory to every step.
	if p.SubPackage != "" {
		for _, s := range steps {
			s.WorkingDirectory = filepath.Join(p.SubPackage, s.WorkingDirectory)
		}
	}

	return steps
}

// Test case for a given package within a project.
type Test struct {
	Project *Project
	Package *Package
}

// ID returns an identifier for the test case.
func (t *Test) ID() string {
	pkgpath := path.Join(t.Project.Repository.String(), t.Package.SubPackage)
	return strings.ReplaceAll(pkgpath, "/", "-")
}

// Projects is a collection of third-party integration tests.
type Projects []*Project

func (p Projects) defaults(set bool) {
	for _, prj := range p {
		prj.defaults(set)
	}
}

// Validate the project collection.
func (p Projects) Validate() error {
	seen := map[string]bool{}
	for _, prj := range p {
		// Project is valid.
		if err := prj.Validate(); err != nil {
			return fmt.Errorf("project %s: %w", prj.ID(), err)
		}

		// No duplicate entries.
		id := prj.ID()
		if seen[id] {
			return fmt.Errorf("duplicate project %q", id)
		}
		seen[id] = true
	}
	return nil
}

// Tests returns all test cases for the projects collection.
func (p Projects) Tests() []*Test {
	var ts []*Test
	for _, prj := range p {
		for _, pkg := range prj.Packages {
			ts = append(ts, &Test{
				Project: prj,
				Package: pkg,
			})
		}
	}
	return ts
}

// Ranked returns a copy of the projects list ranked in desending order of
// popularity.
func (p Projects) Ranked() Projects {
	ranked := append(Projects(nil), p...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Metadata.Stars > ranked[j].Metadata.Stars
	})
	return ranked
}

// Top returns the top n most popular projects.
func (p Projects) Top(n int) Projects {
	top := p.Ranked()
	if len(top) > n {
		top = top[:n]
	}
	return top
}

// LoadProjects loads a list of project configurations from JSON format.
func LoadProjects(r io.Reader) (Projects, error) {
	var prjs Projects
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&prjs); err != nil {
		return nil, err
	}
	prjs.defaults(true)
	return prjs, nil
}

// LoadProjectsFile loads a list of project configurations from a JSON file.
func LoadProjectsFile(filename string) (Projects, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadProjects(f)
}

// StoreProjects writes a list of project configurations in JSON format.
func StoreProjects(w io.Writer, prjs Projects) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	prjs.defaults(false)
	err := e.Encode(prjs)
	prjs.defaults(true)
	return err
}

// StoreProjectsFile writes a list of project configurations to a JSON file.
func StoreProjectsFile(filename string, prjs Projects) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return StoreProjects(f, prjs)
}
