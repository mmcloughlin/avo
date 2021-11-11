package thirdparty

import (
	"context"
	"flag"
	"testing"

	"github.com/mmcloughlin/avo/internal/github"
	"github.com/mmcloughlin/avo/internal/test"
)

var update = flag.Bool("update", false, "update package metadata")

func TestPackagesFileMetadata(t *testing.T) {
	test.RequiresNetwork(t)
	ctx := context.Background()

	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}

	g := github.NewClient(github.WithTokenFromEnvironment())

	for _, pkg := range pkgs {
		// Fetch metadata.
		r, err := g.Repository(ctx, pkg.Repository.Owner, pkg.Repository.Name)
		if err != nil {
			t.Fatal(err)
		}

		// Update, if requested.
		if *update {
			pkg.DefaultBranch = r.DefaultBranch
			pkg.Metadata.Description = r.Description
			pkg.Metadata.Homepage = r.Homepage
			pkg.Metadata.Stars = r.StargazersCount

			t.Logf("%s: metadata updated", pkg.ID())
		}

		// Check up to date. Potentially fast-changing properties not included.
		uptodate := true
		uptodate = pkg.DefaultBranch == r.DefaultBranch && uptodate
		uptodate = pkg.Metadata.Description == r.Description && uptodate
		uptodate = pkg.Metadata.Homepage == r.Homepage && uptodate

		if !uptodate {
			t.Errorf("%s: metadata out of date (use -update flag to fix)", pkg.ID())
		}

	}

	if err := StorePackagesFile("packages.json", pkgs); err != nil {
		t.Fatal(err)
	}
}

func TestPackagesFileKnownIssues(t *testing.T) {
	test.RequiresNetwork(t)
	ctx := context.Background()

	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}

	g := github.NewClient(github.WithTokenFromEnvironment())

	for _, pkg := range pkgs {
		// Skipped packages must refer to an open issue.
		if !pkg.Skip() {
			continue
		}

		if pkg.KnownIssue == 0 {
			t.Errorf("%s: skipped package must refer to known issue", pkg.ID())
		}

		issue, err := g.Issue(ctx, "mmcloughlin", "avo", pkg.KnownIssue)
		if err != nil {
			t.Fatal(err)
		}

		if issue.State != "open" {
			t.Errorf("%s: known issue in %s state", pkg.ID(), issue.State)
		}

		if pkg.Reason() != issue.HTMLURL {
			t.Errorf("%s: expected skip reason to be the issue url %s", pkg.ID(), issue.HTMLURL)
		}
	}
}
