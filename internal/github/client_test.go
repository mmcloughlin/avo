package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mmcloughlin/avo/internal/test"
)

func TestClientRepository(t *testing.T) {
	test.RequiresNetwork(t)

	ctx := context.Background()
	g := NewClient(WithHTTPClient(http.DefaultClient), WithTokenFromEnvironment())
	r, err := g.Repository(ctx, "golang", "go")
	if err != nil {
		t.Fatal(err)
	}

	j, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("repository = %s", j)
}

func TestClientIssue(t *testing.T) {
	test.RequiresNetwork(t)

	ctx := context.Background()
	g := NewClient(WithHTTPClient(http.DefaultClient), WithTokenFromEnvironment())
	r, err := g.Issue(ctx, "octocat", "hello-world", 42)
	if err != nil {
		t.Fatal(err)
	}

	j, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("issue = %s", j)
}
