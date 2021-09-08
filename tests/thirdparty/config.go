// Package thirdparty executes integration tests based on third-party packages that use avo.
package thirdparty

import (
	"encoding/json"
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

type Repository struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (r Repository) String() string {
	return path.Join(r.Owner, r.Name)
}

func (r Repository) ID() string {
	return r.Owner + "-" + r.Name
}

// CloneURL returns the git clone URL.
func (r Repository) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s.git", r)
}

type Step struct {
	WorkingDirectory string     `json:"dir"`
	Commands         [][]string `json:"commands"`
}

// Package defines an integration test based on a third-party package using avo.
type Package struct {
	Repository Repository `json:"repository"`
	Version    string     `json:"version"`  // git sha, tag or branch
	Module     string     `json:"module"`   // path to module file
	Setup      []*Step    `json:"setup"`    // setup commands to run
	Generate   []*Step    `json:"generate"` // generate commands to run
	Test       []*Step    `json:"test"`     // test commands (defaults to "go test ./...")
}

// ID returns an identifier for the package.
func (p Package) ID() string {
	return p.Repository.ID()
}

func (p Package) Steps(c *Context) []*Step {
	var steps []*Step

	// Optional setup.
	steps = append(steps, p.Setup...)

	// Replace avo dependency.
	const invalid = "v0.0.0-00010101000000-000000000000"
	moddir := filepath.Dir(p.Module)
	modfile := filepath.Base(p.Module)
	steps = append(steps, &Step{
		WorkingDirectory: moddir,
		Commands: [][]string{
			{"go", "mod", "edit", "-modfile=" + modfile, "-require=github.com/mmcloughlin/avo@" + invalid},
			{"go", "mod", "edit", "-modfile=" + modfile, "-replace=github.com/mmcloughlin/avo=" + c.AvoDirectory},
			{"go", "mod", "tidy", "-modfile=" + modfile},
		},
	})

	// Run generation.
	steps = append(steps, p.Generate...)

	// Display changes.
	steps = append(steps, &Step{
		Commands: [][]string{
			{"git", "-C", c.RepositoryDirectory, "diff"},
		},
	})

	// Tests.
	if len(p.Test) > 0 {
		steps = append(steps, p.Test...)
	} else {
		steps = append(steps, &Step{
			Commands: [][]string{
				{"go", "test", "./..."},
			},
		})
	}

	return steps
}

// LoadPackages loads a list of package configurations from JSON format.
func LoadPackages(r io.Reader) ([]Package, error) {
	var pkgs []Package
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&pkgs); err != nil {
		return nil, err
	}
	return pkgs, nil
}

// LoadPackagesFile loads a list of package configurations from a JSON file.
func LoadPackagesFile(filename string) ([]Package, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadPackages(f)
}
