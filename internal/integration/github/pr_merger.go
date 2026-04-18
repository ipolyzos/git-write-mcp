// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "context"

// PRMerger merges pull requests via the GitHub REST API.
type PRMerger interface {
	MergePullRequest(ctx context.Context, owner, repo string, number int, in MergePullRequestInput) (*MergePullRequestResult, error)
}
