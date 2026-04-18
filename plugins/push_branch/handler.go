// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package push_branch

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for push_branch (MCP JSON and tests).
type Params struct {
	BranchName string `json:"branch_name"`
	Remote     string `json:"remote,omitempty"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

// Handler runs push_branch through the shared mutation [runtime.Pipeline].
type Handler struct {
	Pipeline *runtime.Pipeline
	Actor    string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK              bool                   `json:"ok"`
	Error           string                 `json:"error,omitempty"`
	BranchName      string                 `json:"branch_name"`
	Remote          string                 `json:"remote"`
	DryRunRequested bool                   `json:"dry_run_requested"`
	Mode            string                 `json:"mode,omitempty"`
	DryRun          bool                   `json:"dry_run,omitempty"`
	DryRunPreview   *runtime.DryRunPreview `json:"dry_run_preview,omitempty"`
	GitStdout       string                 `json:"git_stdout,omitempty"`
	GitStderr       string                 `json:"git_stderr,omitempty"`
}

// Execute validates branch and remote policy, ensures HEAD matches branch_name,
// and runs git push <remote> <branch> (no flags, no refspec, no force).
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("push_branch: nil handler or pipeline")
	}
	if h.Pipeline.Policy == nil {
		return ToolResult{}, errors.New("push_branch: nil policy engine")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}

	branch := strings.TrimSpace(p.BranchName)
	remote := strings.TrimSpace(p.Remote)
	if remote == "" {
		remote = "origin"
	}

	out := ToolResult{
		BranchName:      branch,
		Remote:          remote,
		DryRunRequested: p.DryRun,
	}

	mr := &runtime.MutationRequest{
		Actor:              actor,
		Action:             "push_branch",
		Target:             branch,
		PushRemote:         remote,
		WorkDir:            abs,
		DryRunRequested:    p.DryRun,
		RejectDetachedHEAD: true,
		RequireHEADBranch:  branch,
	}
	mr.Validate = func() error {
		if branch == "" {
			return &policy.ValidationError{Field: "branch_name", Msg: "required"}
		}
		if err := policy.ValidateBranchName(branch); err != nil {
			return err
		}
		if err := policy.ValidateAllowedPushRemote(remote, policy.AllowedPushRemotes()); err != nil {
			return err
		}
		mr.GitArgs = []string{"push", remote, branch}
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
	if res.DryRunPreview != nil {
		out.Mode = "dry-run"
		out.DryRun = true
		out.DryRunPreview = res.DryRunPreview
	} else if out.OK {
		out.Mode = "execute"
	}
	return out, err
}
