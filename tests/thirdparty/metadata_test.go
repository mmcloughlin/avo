package thirdparty

import (
	"context"
	"flag"
	"net/http"
	"testing"

	"github.com/mmcloughlin/avo/internal/github"
	"github.com/mmcloughlin/avo/internal/test"
)

var (
	update = flag.Bool("update", false, "update package metadata")
)

func TestPackagesFileMetadata(t *testing.T) {
	test.RequiresNetwork(t)
	ctx := context.Background()

	pkgs, err := LoadPackagesFile("packages.json")
	if err != nil {
		t.Fatal(err)
	}

	g := github.NewClient(http.DefaultClient)

	for _, pkg := range pkgs {
		// Fetch metadata.
		r, err := g.Repository(ctx, pkg.Repository.Owner, pkg.Repository.Name)
		if err != nil {
			t.Fatal(err)
		}

		// Update, if requested.
		if *update {
			pkg.DefaultBranch = r.DefaultBranch

			t.Logf("%s: metadata updated", pkg.ID())
		}

		// Check up to date.
		uptodate := pkg.DefaultBranch == r.DefaultBranch

		if !uptodate {
			t.Errorf("%s: metadata out of date (use -update flag to fix)", pkg.ID())
		}
	}

	if err := StorePackagesFile("packages.json", pkgs); err != nil {
		t.Fatal(err)
	}
}
