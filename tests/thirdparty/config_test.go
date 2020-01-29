package thirdparty

import (
	"strings"
	"testing"
)

func TestPackageName(t *testing.T) {
	p := Package{ImportPath: "github.com/username/repo"}
	if p.Name() != "repo" {
		t.Fail()
	}
}

func TestPackageCloneURL(t *testing.T) {
	p := Package{ImportPath: "github.com/username/repo"}
	if p.CloneURL() != "https://github.com/username/repo.git" {
		t.Fail()
	}
}

func TestPackagesTestPath(t *testing.T) {
	p := Package{}
	if p.TestPath() != "./..." {
		t.Fail()
	}

	p.Test = "./sub"
	if p.TestPath() != "./sub" {
		t.Fail()
	}
}

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
