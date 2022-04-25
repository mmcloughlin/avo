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
			Name: "project_missing_default_branch",
			Item: &Project{
				Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
			},
			ErrorSubstring: "missing default branch",
		},
		{
			Name: "project_missing_version",
			Item: &Project{
				Repository:    GithubRepository{Owner: "octocat", Name: "hello-world"},
				DefaultBranch: "main",
			},
			ErrorSubstring: "missing version",
		},
		{
			Name:           "package_missing_module",
			Item:           &Package{},
			ErrorSubstring: "missing module",
		},
		{
			Name: "package_no_generate_commands",
			Item: &Package{
				Module: "avo/go.mod",
			},
			ErrorSubstring: "no generate commands",
		},
		{
			Name: "package_invalid_generate_commands",
			Item: &Package{
				Module: "avo/go.mod",
				Generate: []*Step{
					{},
				},
			},
			ErrorSubstring: "generate step: missing name",
		},
		{
			Name: "projects_invalid_package",
			Item: Projects{
				{
					Repository: GithubRepository{Owner: "octocat", Name: "hello-world"},
				},
			},
			ErrorSubstring: "missing default branch",
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

func TestLoadProjectsBad(t *testing.T) {
	r := strings.NewReader(`[{"unknown_field": "value"}]`)
	_, err := LoadProjects(r)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestLoadProjectsFileNotExist(t *testing.T) {
	prjs, err := LoadProjectsFile("does_not_exist")
	if prjs != nil {
		t.Fatal("expected nil return")
	}
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestProjectsFileValid(t *testing.T) {
	prjs, err := LoadProjectsFile("projects.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, prj := range prjs {
		t.Logf("read: %s", prj.ID())
	}
	if len(prjs) == 0 {
		t.Fatal("no packages loaded")
	}
	if err := prjs.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProjectsFileStepsValid(t *testing.T) {
	prjs, err := LoadProjectsFile("projects.json")
	if err != nil {
		t.Fatal(err)
	}
	c := &Context{
		AvoDirectory:        "avo",
		RepositoryDirectory: "repo",
	}
	for _, prj := range prjs {
		for _, pkg := range prj.Packages {
			for _, s := range pkg.Steps(c) {
				if err := s.Validate(); err != nil {
					t.Errorf("project %s: package %s: %s", prj.ID(), pkg.Name(), err)
				}
			}
		}
	}
}

func TestProjectsFileRoundtrip(t *testing.T) {
	prjs, err := LoadProjectsFile("projects.json")
	if err != nil {
		t.Fatal(err)
	}

	// Write and read back.
	buf := bytes.NewBuffer(nil)
	if err := StoreProjects(buf, prjs); err != nil {
		t.Fatal(err)
	}

	roundtrip, err := LoadProjects(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Should be identical.
	if !reflect.DeepEqual(prjs, roundtrip) {
		t.Fatal("roundtrip mismatch")
	}
}
