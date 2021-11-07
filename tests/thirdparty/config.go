// Package thirdparty executes integration tests based on third-party packages that use avo.
package thirdparty

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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

// CloneURL returns the git clone URL.
func (r GithubRepository) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s.git", r)
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

// Package defines an integration test based on a third-party package using avo.
type Package struct {
	// Repository the package belongs to. At the moment, all packages are
	// available on github.
	Repository GithubRepository `json:"repository"`

	// Version as a git sha, tag or branch.
	Version string `json:"version"`

	// Sub-package within the repository under test. All file path references
	// will be relative to this directory. If empty the root of the repository
	// is used.
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
	Generate []*Step `json:"generate"` // generate commands to run

	// Test steps. If empty, defaults to "go test ./...".
	Test []*Step `json:"test,omitempty"`
}

// ID returns an identifier for the package.
func (p *Package) ID() string {
	pkgpath := path.Join(p.Repository.String(), p.SubPackage)
	return strings.ReplaceAll(pkgpath, "/", "-")
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
	if p.Version == "" {
		return errors.New("missing version")
	}
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

// Packages is a collection of third-party integration tests.
type Packages []*Package

func (p Packages) defaults(set bool) {
	for _, pkg := range p {
		pkg.defaults(set)
	}
}

// Validate the package collection.
func (p Packages) Validate() error {
	for _, pkg := range p {
		if err := pkg.Validate(); err != nil {
			return fmt.Errorf("package %s: %w", pkg.ID(), err)
		}
	}
	return nil
}

// LoadPackages loads a list of package configurations from JSON format.
func LoadPackages(r io.Reader) (Packages, error) {
	var pkgs Packages
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&pkgs); err != nil {
		return nil, err
	}
	pkgs.defaults(true)
	return pkgs, nil
}

// LoadPackagesFile loads a list of package configurations from a JSON file.
func LoadPackagesFile(filename string) (Packages, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadPackages(f)
}

// StorePackages writes a list of package configurations in JSON format.
func StorePackages(w io.Writer, pkgs Packages) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	pkgs.defaults(false)
	err := e.Encode(pkgs)
	pkgs.defaults(true)
	return err
}

// StorePackagesFile writes a list of package configurations to a JSON file.
func StorePackagesFile(filename string, pkgs Packages) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return StorePackages(f, pkgs)
}
