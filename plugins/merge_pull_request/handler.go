// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package merge_pull_request

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds MCP JSON inputs for merge_pull_request.
type Params struct {
	PullRequestNumber int    `json:"pull_request_number"`
	MergeMethod       string `json:"merge_method,omitempty"`
	DryRun            bool   `json:"dry_run,omitempty"`
}

// Handler runs merge_pull_request through [runtime.Pipeline] and [github.MCPGitHub].
type Handler struct {
	Pipeline    *runtime.Pipeline
	GitHub      github.MCPGitHub
	GitHubOwner string
	GitHubRepo  string
	Actor       string
}

// ToolResult is deterministic JSON returned to MCP clients.
type ToolResult struct {
	OK                bool                   `json:"ok"`
	Error             string                 `json:"error,omitempty"`
	PullRequestNumber int                    `json:"pull_request_number"`
	BaseRef           string                 `json:"base_ref,omitempty"`
	MergeMethod       string                 `json:"merge_method"`
	MergedSHA         string                 `json:"merged_sha,omitempty"`
	MergeMessage      string                 `json:"merge_message,omitempty"`
	DryRunRequested   bool                   `json:"dry_run_requested"`
	Mode              string                 `json:"mode,omitempty"`
	DryRun            bool                   `json:"dry_run,omitempty"`
	DryRunPreview     *runtime.DryRunPreview `json:"dry_run_preview,omitempty"`
	GitStdout         string                 `json:"git_stdout,omitempty"`
	GitStderr         string                 `json:"git_stderr,omitempty"`
}

// Execute fetches PR status, evaluates merge policy, then merges via GitHub API unless dry-run.
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("merge_pull_request: nil handler or pipeline")
	}
	if h.Pipeline.Policy == nil {
		return ToolResult{}, errors.New("merge_pull_request: nil policy engine")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}
	owner := strings.TrimSpace(h.GitHubOwner)
	repo := strings.TrimSpace(h.GitHubRepo)
	n := p.PullRequestNumber
	resolved := policy.NormalizeMergeMethod(p.MergeMethod)

	out := ToolResult{
		PullRequestNumber: n,
		MergeMethod:       resolved,
		DryRunRequested:   p.DryRun,
	}

	var merged *github.MergePullRequestResult

	mr := &runtime.MutationRequest{
		Actor:           actor,
		Action:          "merge_pull_request",
		Target:          strconv.Itoa(n),
		WorkDir:         abs,
		DryRunRequested: p.DryRun,
		GitHubOwner:     owner,
		GitHubRepo:      repo,
	}
	mr.Validate = func() error {
		if h.GitHub == nil {
			return fmt.Errorf("merge_pull_request: GitHub integration not configured (set GITHUB_TOKEN and repository env)")
		}
		if n < 1 {
			return &policy.ValidationError{Field: "pull_request_number", Msg: "must be a positive integer"}
		}
		st, err := h.GitHub.GetPullRequestStatus(ctx, owner, repo, n)
		if err != nil {
			return err
		}
		out.BaseRef = strings.TrimSpace(st.BaseRef)
		if err := policy.EvaluateMergePolicy(st, resolved); err != nil {
			return err
		}
		mr.PRMergeNumber = n
		mr.PRMergeMethod = resolved
		mr.PRTargetBranch = strings.TrimSpace(st.BaseRef)
		mr.GitArgs = []string{"merge_pull_request", strconv.Itoa(n), resolved}
		mm := resolved
		mr.APIExecute = func(ctx context.Context) error {
			r, err := h.GitHub.MergePullRequest(ctx, owner, repo, n, github.MergePullRequestInput{MergeMethod: mm})
			if err != nil {
				return err
			}
			merged = r
			return nil
		}
		return nil
	}

	res, err := h.Pipeline.Execute(ctx, mr)
	if len(res.GitResult.Stdout) > 0 {
		out.GitStdout = strings.TrimSpace(string(res.GitResult.Stdout))
	}
	if len(res.GitResult.Stderr) > 0 {
		out.GitStderr = strings.TrimSpace(string(res.GitResult.Stderr))
	}
	if err != nil {
		out.OK = false
		out.Error = err.Error()
		out.Mode = "blocked"
		return out, err
	}
	out.OK = res.OK
	if merged != nil {
		out.MergedSHA = merged.SHA
		out.MergeMessage = merged.Message
	}
	if res.DryRunPreview != nil {
		out.Mode = "dry-run"
		out.DryRun = true
		out.DryRunPreview = res.DryRunPreview
	} else if out.OK {
		out.Mode = "execute"
	}
	return out, nil
}
