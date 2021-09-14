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
)

type Context struct {
	AvoDirectory        string
	RepositoryDirectory string
}

type GithubRepository struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (r GithubRepository) String() string {
	return path.Join(r.Owner, r.Name)
}

func (r GithubRepository) ID() string {
	return r.Owner + "-" + r.Name
}

// CloneURL returns the git clone URL.
func (r GithubRepository) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s.git", r)
}

type Step struct {
	Name             string   `json:"name"`
	WorkingDirectory string   `json:"dir"`
	Commands         []string `json:"commands"`
}

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
	Repository GithubRepository `json:"repository"`
	Version    string           `json:"version"`  // git sha, tag or branch
	Module     string           `json:"module"`   // path to module file
	Setup      []*Step          `json:"setup"`    // setup commands to run
	Generate   []*Step          `json:"generate"` // generate commands to run
	Test       []*Step          `json:"test"`     // test commands (defaults to "go test ./...")
}

// ID returns an identifier for the package.
func (p *Package) ID() string {
	return p.Repository.ID()
}

func (p *Package) setdefaults() {
	for _, stage := range []struct {
		Steps       []*Step
		DefaultName string
	}{
		{p.Setup, "Setup"},
		{p.Generate, "Generate"},
		{p.Test, "Test"},
	} {
		if len(stage.Steps) == 1 && stage.Steps[0].Name == "" {
			stage.Steps[0].Name = stage.DefaultName
		}
	}
}

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

	return steps
}

type Packages []*Package

func (p Packages) setdefaults() {
	for _, pkg := range p {
		pkg.setdefaults()
	}
}

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
	pkgs.setdefaults()
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
