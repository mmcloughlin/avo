package thirdparty

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

type Package struct {
	ImportPath string     `json:"import_path"`
	Version    string     `json:"version"`
	Dir        string     `json:"dir"`
	Generate   [][]string `json:"generate"`
}

func (p Package) Name() string {
	return filepath.Base(p.ImportPath)
}

func (p Package) CloneURL() string {
	return "https://" + p.ImportPath + ".git"
}

func LoadPackages(r io.Reader) ([]Package, error) {
	var pkgs []Package
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	if err := d.Decode(&pkgs); err != nil {
		return nil, err
	}
	return pkgs, nil
}

func LoadPackagesFile(filename string) ([]Package, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadPackages(f)
}
