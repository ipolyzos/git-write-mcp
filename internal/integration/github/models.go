// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

// PullRequest is a minimal, stable subset of the GitHub pulls API response.
type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	HeadSHA string `json:"-"`
}

// ChecksSummary aggregates GitHub combined status data for the PR head commit.
type ChecksSummary struct {
	State      string          `json:"state"`
	TotalCount int             `json:"total_count"`
	Contexts   []StatusContext `json:"contexts,omitempty"`
}

// StatusContext is one combined status row (Contexts sorted by Context for stable JSON).
type StatusContext struct {
	Context string `json:"context"`
	State   string `json:"state"`
}

// ReviewsSummary counts pull request review records by state.
type ReviewsSummary struct {
	Approved         int `json:"approved"`
	ChangesRequested int `json:"changes_requested"`
	Commented        int `json:"commented"`
	Pending          int `json:"pending"`
	Dismissed        int `json:"dismissed"`
}

// PullRequestStatus is read-only PR readiness data from the GitHub REST API.
type PullRequestStatus struct {
	Number         int            `json:"number"`
	Title          string         `json:"title"`
	State          string         `json:"state"`
	Merged         bool           `json:"merged"`
	Draft          bool           `json:"draft"`
	Mergeable      *bool          `json:"mergeable,omitempty"`
	MergeableState string         `json:"mergeable_state,omitempty"`
	HTMLURL        string         `json:"html_url,omitempty"`
	HeadRef        string         `json:"head_ref"`
	BaseRef        string         `json:"base_ref"`
	HeadSHA        string         `json:"head_sha"`
	ChecksSummary  ChecksSummary  `json:"checks_summary"`
	ReviewsSummary ReviewsSummary `json:"reviews_summary"`
}

// MergePullRequestInput is the JSON body for PUT /repos/{owner}/{repo}/pulls/{pull_number}/merge.
type MergePullRequestInput struct {
	MergeMethod string // merge, squash, rebase (GitHub REST)
}

// MergePullRequestResult is the minimal successful merge API response shape.
type MergePullRequestResult struct {
	Merged  bool   `json:"merged"`
	Message string `json:"message"`
	SHA     string `json:"sha"`
}

// CreatePullRequestInput is the body for POST /repos/{owner}/{repo}/pulls.
type CreatePullRequestInput struct {
	Title string
	Head  string
	Base  string
	Body  string
	Draft bool
}

// CreateReleaseInput is a placeholder for release creation.
type CreateReleaseInput struct{}

// Release is a placeholder for release API responses.
type Release struct{}
