// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HTTPDoer performs HTTP requests (typically [*http.Client]). Injectable for tests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a minimal GitHub REST client with validated configuration.
type Client struct {
	cfg     Config
	http    HTTPDoer
	baseURL *url.URL
}

// NewClient returns a client configured for https://api.github.com/ unless
// extended later. httpClient must be non-nil for testability (inject fakes).
func NewClient(cfg Config, httpClient HTTPDoer) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if httpClient == nil {
		return nil, fmt.Errorf("github: nil http client")
	}
	base, err := url.Parse("https://api.github.com")
	if err != nil {
		return nil, err
	}
	return &Client{cfg: cfg, http: httpClient, baseURL: base}, nil
}

// DoAuthorizedGET performs an authenticated GET against the REST API path (must
// start with '/', e.g. "/repos/o/r/pulls/1"). Errors do not include the token.
func (c *Client) DoAuthorizedGET(ctx context.Context, path string) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("github: nil client")
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("github: path must start with /")
	}
	endpoint := strings.TrimRight(c.baseURL.String(), "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return c.http.Do(req)
}

// CreatePullRequest implements [Integration] (placeholder).
func (c *Client) CreatePullRequest(ctx context.Context, in CreatePullRequestInput) (*PullRequest, error) {
	_ = ctx
	_ = in
	return nil, ErrNotImplemented
}

// GetPullRequestStatus implements [Integration] (placeholder).
func (c *Client) GetPullRequestStatus(ctx context.Context, owner, repo string, number int) (*PullRequestStatus, error) {
	_ = ctx
	_ = owner
	_ = repo
	_ = number
	return nil, ErrNotImplemented
}

// MergePullRequest implements [Integration] (placeholder).
func (c *Client) MergePullRequest(ctx context.Context, owner, repo string, number int, in MergePullRequestInput) (*MergePullRequestResult, error) {
	_ = ctx
	_ = owner
	_ = repo
	_ = number
	_ = in
	return nil, ErrNotImplemented
}

// CreateRelease implements [Integration] (placeholder).
func (c *Client) CreateRelease(ctx context.Context, in CreateReleaseInput) (*Release, error) {
	_ = ctx
	_ = in
	return nil, ErrNotImplemented
}
