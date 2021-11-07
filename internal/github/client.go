// Package github provides a client for the Github REST API.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// Client for the Github REST API.
type Client struct {
	client *http.Client
	base   string
}

// NewClient initializes a client using the given HTTP client.
func NewClient(c *http.Client) *Client {
	return &Client{
		client: c,
		base:   "https://api.github.com",
	}
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
