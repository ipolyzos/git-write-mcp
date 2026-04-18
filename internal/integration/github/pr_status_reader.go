// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "context"

// PRStatusReader fetches read-only pull request readiness data from the GitHub REST API.
type PRStatusReader interface {
	GetPullRequestStatus(ctx context.Context, owner, repo string, number int) (*PullRequestStatus, error)
}
