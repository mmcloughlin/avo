package thirdparty

import (
	"strings"
	"testing"
)

func TestLoadPackages(t *testing.T) {
	r := strings.NewReader(`[{"unknown_field": "value"}]`)
	_, err := LoadPackages(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestLoadPackagesFile(t *testing.T) {
	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		t.Log(pkg.ImportPath)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
}

func TestLoadPackagesFileNotExist(t *testing.T) {
	pkgs, err := LoadPackagesFile("does_not_exist")
	if pkgs != nil {
		t.Fatal("expected nil return")
	}
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}
