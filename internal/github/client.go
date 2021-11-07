// Package github provides a client for the Github REST API.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

// Client for the Github REST API.
type Client struct {
	client *http.Client
	base   string
	token  string
}

// Option configures a Github client.
type Option func(*Client)

// WithHTTPClient configures the HTTP client that should be used for Github API
// requests.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.client = h }
}

// WithToken configures a Client with an authentication token for Github API
// requests.
func WithToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithTokenFromEnvironment configures a Client using the GITHUB_TOKEN
// environment variable.
func WithTokenFromEnvironment() Option {
	return WithToken(os.Getenv("GITHUB_TOKEN"))
}

// NewClient initializes a client using the given HTTP client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		client: http.DefaultClient,
		base:   "https://api.github.com",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Repository gets information about the given Github repository.
func (c *Client) Repository(ctx context.Context, owner, name string) (*Repository, error) {
	// Build request.
	u := c.base + "/repos/" + owner + "/" + name
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	// Execute.
	repo := &Repository{}
	if err := c.request(req, repo); err != nil {
		return nil, err
	}

	return repo, nil
}

func (c *Client) request(req *http.Request, payload interface{}) (err error) {
	// Add common headers.
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Execute the request.
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if errc := res.Body.Close(); errc != nil && err == nil {
			err = errc
		}
	}()

	// Check status.
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %d: %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	// Parse response body.
	d := json.NewDecoder(res.Body)

	if err := d.Decode(payload); err != nil {
		return err
	}

	// Should not have trailing data.
	if d.More() {
		return errors.New("unexpected extra data after JSON")
	}

	return nil
}
