// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_pull_request

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for create_pull_request (MCP JSON and tests).
type Params struct {
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch,omitempty"`
	Title        string `json:"title"`
	Body         string `json:"body,omitempty"`
	Draft        bool   `json:"draft,omitempty"`
	DryRun       bool   `json:"dry_run,omitempty"`
}

// Handler runs create_pull_request through [runtime.Pipeline] and [github.PRWriter].
type Handler struct {
	Pipeline    *runtime.Pipeline
	GitHub      github.PRWriter
	GitHubOwner string
	GitHubRepo  string
	Actor       string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK                bool                   `json:"ok"`
	Error             string                 `json:"error,omitempty"`
	SourceBranch      string                 `json:"source_branch"`
	TargetBranch      string                 `json:"target_branch"`
	Title             string                 `json:"title"`
	PullRequestNumber int                    `json:"pull_request_number,omitempty"`
	HTMLURL           string                 `json:"html_url,omitempty"`
	Draft             bool                   `json:"draft,omitempty"`
	DryRunRequested   bool                   `json:"dry_run_requested"`
	Mode              string                 `json:"mode,omitempty"`
	DryRun            bool                   `json:"dry_run,omitempty"`
	DryRunPreview     *runtime.DryRunPreview `json:"dry_run_preview,omitempty"`
	GitStdout         string                 `json:"git_stdout,omitempty"`
	GitStderr         string                 `json:"git_stderr,omitempty"`
}

// Execute validates branches and opens a GitHub pull request unless dry-run.
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("create_pull_request: nil handler or pipeline")
	}
	if h.Pipeline.Policy == nil {
		return ToolResult{}, errors.New("create_pull_request: nil policy engine")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}

	title := strings.TrimSpace(p.Title)
	tgt := strings.TrimSpace(p.TargetBranch)
	if tgt == "" {
		tgt = "develop"
	}
	src := strings.TrimSpace(p.SourceBranch)

	out := ToolResult{
		SourceBranch:    src,
		TargetBranch:    tgt,
		Title:           title,
		Draft:           p.Draft,
		DryRunRequested: p.DryRun,
	}

	var created *github.PullRequest

	mr := &runtime.MutationRequest{
		Actor:           actor,
		Action:          "create_pull_request",
		Target:          title,
		WorkDir:         abs,
		DryRunRequested: p.DryRun,
		GitHubOwner:     strings.TrimSpace(h.GitHubOwner),
		GitHubRepo:      strings.TrimSpace(h.GitHubRepo),
	}
	mr.Validate = func() error {
		if h.GitHub == nil {
			return fmt.Errorf("create_pull_request: GitHub integration not configured (set GITHUB_TOKEN and repository env)")
		}
		if title == "" {
			return &policy.ValidationError{Field: "title", Msg: "required"}
		}
		if err := validateSourceBranch(src); err != nil {
			return err
		}
		if err := policy.ValidatePRTargetBranch(tgt, policy.AllowedPRTargetBranches()); err != nil {
			return err
		}
		mr.GitArgs = []string{"create_pull_request", src, tgt}
		mr.PRTitle = title
		mr.PRSourceBranch = src
		mr.PRTargetBranch = tgt
		mr.PRDraft = p.Draft

		in := github.CreatePullRequestInput{
			Title: title,
			Head:  src,
			Base:  tgt,
			Body:  p.Body,
			Draft: p.Draft,
		}
		mr.APIExecute = func(ctx context.Context) error {
			pr, err := h.GitHub.CreatePullRequest(ctx, in)
			if err != nil {
				return err
			}
			created = pr
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
	if created != nil {
		out.PullRequestNumber = created.Number
		out.HTMLURL = created.HTMLURL
		out.Title = created.Title
		if out.Title == "" {
			out.Title = title
		}
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

func validateSourceBranch(name string) error {
	if err := policy.ValidateBranchName(name); err != nil {
		var ve *policy.ValidationError
		if errors.As(err, &ve) && ve != nil {
			cp := *ve
			cp.Field = "source_branch"
			return &cp
		}
		return err
	}
	return nil
}
