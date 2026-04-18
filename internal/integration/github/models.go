// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

// PullRequest is a minimal, stable subset of the GitHub pulls API response.
type PullRequest struct {
	Number  int
	Title   string
	State   string
	HTMLURL string
	HeadSHA string
}

// PullRequestStatus is a placeholder for future CI/merge readiness fields.
type PullRequestStatus struct{}

// MergePullRequestInput is a placeholder for merge API options.
type MergePullRequestInput struct{}

// MergePullRequestResult is a placeholder for merge API responses.
type MergePullRequestResult struct{}

// CreatePullRequestInput is a placeholder for pull request creation.
type CreatePullRequestInput struct{}

// CreateReleaseInput is a placeholder for release creation.
type CreateReleaseInput struct{}

// Release is a placeholder for release API responses.
type Release struct{}
