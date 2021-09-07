// Package thirdparty executes integration tests based on third-party packages that use avo.
package thirdparty

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type Context struct {
	AvoDirectory        string
	RepositoryDirectory string
}

type Step struct {
	WorkingDirectory string     `json:"dir"`
	Commands         [][]string `json:"commands"`
}

// Package defines an integration test based on a third-party package using avo.
type Package struct {
	ImportPath string  `json:"import_path"` // package import path
	Version    string  `json:"version"`     // git sha, tag or branch
	Module     string  `json:"module"`      // path to module file
	Setup      []*Step `json:"setup"`       // setup commands to run
	Generate   []*Step `json:"generate"`    // generate commands to run
	Test       string  `json:"test"`        // test path relative to repo root (if empty defaults to ./...)
}

// Name returns the package name.
func (p Package) Name() string {
	return filepath.Base(p.ImportPath)
}

// CloneURL returns the git clone URL.
func (p Package) CloneURL() string {
	return "https://" + p.ImportPath + ".git"
}

// TestPath returns the paths to run "go test" on, relative to the repository root.
func (p Package) TestPath() string {
	if p.Test == "" {
		return "./..."
	}
	return p.Test
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
	steps = append(steps, &Step{
		Commands: [][]string{
			{"go", "test", p.TestPath()},
		},
	})

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
