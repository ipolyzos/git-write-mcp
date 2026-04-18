// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"errors"
)

// ErrNotImplemented indicates a workflow operation is not wired in this build.
var ErrNotImplemented = errors.New("github: not implemented")

// Integration is the abstraction for GitHub REST automation used by MCP or workers.
// Methods will gain concrete implementations incrementally; until then callers
// receive [ErrNotImplemented].
type Integration interface {
	CreatePullRequest(ctx context.Context, in CreatePullRequestInput) (*PullRequest, error)
	GetPullRequestStatus(ctx context.Context, owner, repo string, number int) (*PullRequestStatus, error)
	MergePullRequest(ctx context.Context, owner, repo string, number int, in MergePullRequestInput) (*MergePullRequestResult, error)
	CreateRelease(ctx context.Context, in CreateReleaseInput) (*Release, error)
}
