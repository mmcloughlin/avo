package thirdparty

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// Package defines an integration test based on a third-party package using avo.
type Package struct {
	ImportPath string     `json:"import_path"` // package import path
	Version    string     `json:"version"`     // git sha, tag or branch
	Generate   [][]string `json:"generate"`    // generate commands to run
	Dir        string     `json:"dir"`         // working directory for generate commands
}

// Name returns the package name.
func (p Package) Name() string {
	return filepath.Base(p.ImportPath)
}

// CloneURL returns the git clone URL.
func (p Package) CloneURL() string {
	return "https://" + p.ImportPath + ".git"
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
