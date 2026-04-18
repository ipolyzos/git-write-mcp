// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "context"

// PRWriter is the minimal GitHub surface needed to open pull requests.
type PRWriter interface {
	CreatePullRequest(ctx context.Context, in CreatePullRequestInput) (*PullRequest, error)
}
