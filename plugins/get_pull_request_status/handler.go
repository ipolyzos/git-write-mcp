// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package get_pull_request_status

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
)

// Params holds MCP JSON inputs for get_pull_request_status.
type Params struct {
	PullRequestNumber int `json:"pull_request_number"`
}

// Handler performs read-only GitHub PR inspection (no mutation pipeline).
type Handler struct {
	GitHub github.PRStatusReader
	Owner  string
	Repo   string
}

// ToolResult is deterministic JSON returned to MCP clients.
type ToolResult struct {
	OK          bool                      `json:"ok"`
	Error       string                    `json:"error,omitempty"`
	PullRequest *github.PullRequestStatus `json:"pull_request,omitempty"`
}

// Execute loads PR state, mergeability, checks summary, and review counts.
func (h *Handler) Execute(ctx context.Context, p Params) (ToolResult, error) {
	out := ToolResult{}
	if h == nil || h.GitHub == nil {
		out.OK = false
		out.Error = "get_pull_request_status: GitHub integration not configured (set GITHUB_TOKEN and repository env)"
		return out, fmt.Errorf("%s", out.Error)
	}
	n := p.PullRequestNumber
	if n < 1 {
		out.OK = false
		out.Error = "get_pull_request_status: pull_request_number must be a positive integer"
		return out, fmt.Errorf("%s", out.Error)
	}
	owner := strings.TrimSpace(h.Owner)
	repo := strings.TrimSpace(h.Repo)
	if owner == "" || repo == "" {
		out.OK = false
		out.Error = "get_pull_request_status: repository owner/name not configured"
		return out, fmt.Errorf("%s", out.Error)
	}
	st, err := h.GitHub.GetPullRequestStatus(ctx, owner, repo, n)
	if err != nil {
		out.OK = false
		out.Error = err.Error()
		return out, err
	}
	out.OK = true
	out.PullRequest = st
	return out, nil
}
