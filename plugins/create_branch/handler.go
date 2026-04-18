// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_branch

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for create_branch (MCP JSON and tests).
type Params struct {
	BranchName string `json:"branch_name"`
	FromRef    string `json:"from_ref,omitempty"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

// Handler runs create_branch through the shared mutation [runtime.Pipeline].
type Handler struct {
	Pipeline *runtime.Pipeline
	Actor    string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK              bool   `json:"ok"`
	Error           string `json:"error,omitempty"`
	BranchName      string `json:"branch_name"`
	FromRef         string `json:"from_ref,omitempty"`
	DryRunRequested bool   `json:"dry_run_requested"`
	Mode            string `json:"mode,omitempty"`
	DryRun          bool   `json:"dry_run,omitempty"`
	GitStdout       string `json:"git_stdout,omitempty"`
	GitStderr       string `json:"git_stderr,omitempty"`
}

// Execute validates inputs and runs the mutation pipeline.
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{OK: false, Error: "nil pipeline"}, errors.New("create_branch: nil handler or pipeline")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{OK: false}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}

	mr := runtime.MutationRequest{
		Actor:           actor,
		Action:          "create_branch",
		Target:          p.BranchName,
		WorkDir:         abs,
		GitArgs:         gitArgv(p),
		DryRunRequested: p.DryRun,
		Validate: func() error {
			if strings.TrimSpace(p.BranchName) == "" {
				return &policy.ValidationError{Field: "branch_name", Msg: "required"}
			}
			if err := policy.ValidateBranchName(p.BranchName); err != nil {
				return err
			}
			if strings.TrimSpace(p.FromRef) != "" {
				if err := policy.ValidateRefName(p.FromRef); err != nil {
					return err
				}
			}
			return nil
		},
	}

	res, err := h.Pipeline.Execute(ctx, mr)
	out := toolResultFromRuntime(res, p, err)
	return out, err
}

func gitArgv(p Params) []string {
	if strings.TrimSpace(p.FromRef) == "" {
		return []string{"branch", p.BranchName}
	}
	return []string{"branch", p.BranchName, p.FromRef}
}

func toolResultFromRuntime(r runtime.Result, p Params, execErr error) ToolResult {
	out := ToolResult{
		BranchName:      p.BranchName,
		FromRef:         p.FromRef,
		DryRunRequested: p.DryRun,
	}
	if len(r.GitResult.Stdout) > 0 {
		out.GitStdout = strings.TrimSpace(string(r.GitResult.Stdout))
	}
	if len(r.GitResult.Stderr) > 0 {
		out.GitStderr = strings.TrimSpace(string(r.GitResult.Stderr))
	}
	if execErr != nil {
		out.OK = false
		out.Error = execErr.Error()
		out.Mode = "blocked"
		return out
	}
	out.OK = r.OK
	if r.DryRunPreview != nil {
		out.Mode = "dry-run"
		out.DryRun = true
	} else if out.OK {
		out.Mode = "execute"
	}
	return out
}
