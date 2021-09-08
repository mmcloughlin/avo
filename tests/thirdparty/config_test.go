package thirdparty

import (
	"strings"
	"testing"
)

func TestLoadPackagesBad(t *testing.T) {
	r := strings.NewReader(`[{"unknown_field": "value"}]`)
	_, err := LoadPackages(r)
	if err == nil {
		t.Fatal("expected non-nil error")
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

func TestPackagesFileValid(t *testing.T) {
	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		t.Logf("read: %s", pkg.ID())
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	if err := pkgs.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPackagesFileStepsValid(t *testing.T) {
	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}
	c := &Context{
		AvoDirectory:        "avo",
		RepositoryDirectory: "repo",
	}
	for _, pkg := range pkgs {
		for _, s := range pkg.Steps(c) {
			if err := s.Validate(); err != nil {
				t.Errorf("package %s: %s", pkg.ID(), err)
			}
		}
	}
}
