package thirdparty

import (
	"context"
	"flag"
	"testing"

	"github.com/mmcloughlin/avo/internal/github"
	"github.com/mmcloughlin/avo/internal/test"
)

var update = flag.Bool("update", false, "update project metadata")

func TestSuiteFileMetadata(t *testing.T) {
	test.RequiresNetwork(t)
	ctx := context.Background()

	s, err := LoadSuiteFile("suite.json")
	if err != nil {
		t.Fatal(err)
	}

	g := github.NewClient(github.WithTokenFromEnvironment())

	for _, prj := range s.Projects {
		// Fetch metadata.
		r, err := g.Repository(ctx, prj.Repository.Owner, prj.Repository.Name)
		if err != nil {
			t.Fatal(err)
		}

		// Update, if requested.
		if *update {
			prj.DefaultBranch = r.DefaultBranch
			prj.Metadata.Description = r.Description
			prj.Metadata.Homepage = r.Homepage
			prj.Metadata.Stars = r.StargazersCount

			t.Logf("%s: metadata updated", prj.ID())
		}

		// Check up to date. Potentially fast-changing properties not included.
		uptodate := true
		uptodate = prj.DefaultBranch == r.DefaultBranch && uptodate
		uptodate = prj.Metadata.Description == r.Description && uptodate
		uptodate = prj.Metadata.Homepage == r.Homepage && uptodate

		if !uptodate {
			t.Errorf("%s: metadata out of date (use -update flag to fix)", prj.ID())
		}
	}

	if err := StoreSuiteFile("suite.json", s); err != nil {
		t.Fatal(err)
	}
}

func TestSuiteFileKnownIssues(t *testing.T) {
	test.RequiresNetwork(t)
	ctx := context.Background()

	s, err := LoadSuiteFile("suite.json")
	if err != nil {
		t.Fatal(err)
	}

	g := github.NewClient(github.WithTokenFromEnvironment())

	for _, prj := range s.Projects {
		// Skipped packages must refer to an open issue.
		if !prj.Skip() {
			continue
		}

		if prj.KnownIssue == 0 {
			t.Errorf("%s: skipped package must refer to known issue", prj.ID())
		}

		issue, err := g.Issue(ctx, "mmcloughlin", "avo", prj.KnownIssue)
		if err != nil {
			t.Fatal(err)
		}

		if issue.State != "open" {
			t.Errorf("%s: known issue in %s state", prj.ID(), issue.State)
		}

		if prj.Reason() != issue.HTMLURL {
			t.Errorf("%s: expected skip reason to be the issue url %s", prj.ID(), issue.HTMLURL)
		}
	}
}
