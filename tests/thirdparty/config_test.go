package thirdparty

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestValidateErrors(t *testing.T) {
	cases := []struct {
		Name           string
		Item           interface{ Validate() error }
		ErrorSubstring string
	}{
		{
			Name:           "step_missing_name",
			Item:           &Step{},
			ErrorSubstring: "missing name",
		},
		{
			Name: "step_missing_commands",
			Item: &Step{
				Name: "Setup",
			},
			ErrorSubstring: "missing commands",
		},
		{
			Name: "package_missing_version",
			Item: &Package{
				Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
			},
			ErrorSubstring: "missing version",
		},
		{
			Name: "package_missing_module",
			Item: &Package{
				Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
				Version:    "v1.0.1",
			},
			ErrorSubstring: "missing module",
		},
		{
			Name: "package_no_generate_commands",
			Item: &Package{
				Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
				Version:    "v1.0.1",
				Module:     "avo/go.mod",
			},
			ErrorSubstring: "no generate commands",
		},
		{
			Name: "package_invalid_generate_commands",
			Item: &Package{
				Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
				Version:    "v1.0.1",
				Module:     "avo/go.mod",
				Generate: []*Step{
					{},
				},
			},
			ErrorSubstring: "generate step: missing name",
		},
		{
			Name: "packages_invalid_package",
			Item: Packages{
				{
					Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
				},
			},
			ErrorSubstring: "missing version",
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			err := c.Item.Validate()
			if err == nil {
				t.Fatal("expected error; got nil")
			}
			if !strings.Contains(err.Error(), c.ErrorSubstring) {
				t.Fatalf("expected error message to contain %q; got %q", c.ErrorSubstring, err)
			}
		})
	}
}

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

func TestPackagesFileRoundtrip(t *testing.T) {
	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}

	// Write and read back.
	buf := bytes.NewBuffer(nil)
	if err := StorePackages(buf, pkgs); err != nil {
		t.Fatal(err)
	}

	roundtrip, err := LoadPackages(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Should be identical.
	if !reflect.DeepEqual(pkgs, roundtrip) {
		t.Fatal("roundtrip mismatch")
	}
}
