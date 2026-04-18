// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package tag_revision

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for tag_revision (MCP JSON and tests).
type Params struct {
	TagName   string `json:"tag_name"`
	TargetRef string `json:"target_ref,omitempty"`
	Message   string `json:"message"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

// Handler runs tag_revision through the shared mutation [runtime.Pipeline].
type Handler struct {
	Pipeline *runtime.Pipeline
	Actor    string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK              bool   `json:"ok"`
	Error           string `json:"error,omitempty"`
	TagName         string `json:"tag_name"`
	TargetRef       string `json:"target_ref,omitempty"`
	Message         string `json:"message"`
	DryRunRequested bool   `json:"dry_run_requested"`
	Mode            string `json:"mode,omitempty"`
	DryRun          bool   `json:"dry_run,omitempty"`
	GitStdout       string `json:"git_stdout,omitempty"`
	GitStderr       string `json:"git_stderr,omitempty"`
}

// Execute validates inputs and runs `git tag -a <tag_name> -m <message> [<target_ref>]`.
// argv never includes `-f`, `--force`, or `-d` (no forced overwrite or deletion).
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("tag_revision: nil handler or pipeline")
	}
	abs, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}

	mr := runtime.MutationRequest{
		Actor:           actor,
		Action:          "tag_revision",
		Target:          auditTarget(p),
		WorkDir:         abs,
		GitArgs:         gitArgv(p),
		DryRunRequested: p.DryRun,
		Validate: func() error {
			if strings.TrimSpace(p.TagName) == "" {
				return &policy.ValidationError{Field: "tag_name", Msg: "required"}
			}
			if err := policy.ValidateTagName(p.TagName); err != nil {
				return err
			}
			if err := policy.ValidateCommitMessage(p.Message); err != nil {
				return err
			}
			if strings.TrimSpace(p.TargetRef) != "" {
				if err := policy.ValidateRefName(p.TargetRef); err != nil {
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

func auditTarget(p Params) string {
	b := strings.Builder{}
	b.WriteString(p.TagName)
	if tr := strings.TrimSpace(p.TargetRef); tr != "" {
		b.WriteString("@")
		b.WriteString(tr)
	}
	return b.String()
}

func gitArgv(p Params) []string {
	argv := []string{"tag", "-a", p.TagName, "-m", p.Message}
	if tr := strings.TrimSpace(p.TargetRef); tr != "" {
		argv = append(argv, tr)
	}
	return argv
}

func toolResultFromRuntime(r runtime.Result, p Params, execErr error) ToolResult {
	out := ToolResult{
		TagName:         p.TagName,
		TargetRef:       p.TargetRef,
		Message:         p.Message,
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
