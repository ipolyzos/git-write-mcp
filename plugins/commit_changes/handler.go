// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package commit_changes

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for commit_changes (MCP JSON and tests).
type Params struct {
	Message string `json:"message"`
	DryRun  bool   `json:"dry_run,omitempty"`
}

// Handler runs commit_changes through the shared mutation [runtime.Pipeline].
type Handler struct {
	Pipeline *runtime.Pipeline
	Actor    string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK              bool   `json:"ok"`
	Error           string `json:"error,omitempty"`
	Message         string `json:"message"`
	DryRunRequested bool   `json:"dry_run_requested"`
	Mode            string `json:"mode,omitempty"`
	DryRun          bool   `json:"dry_run,omitempty"`
	GitStdout       string `json:"git_stdout,omitempty"`
	GitStderr       string `json:"git_stderr,omitempty"`
}

// Execute validates the commit message and runs git commit -m <message>.
// argv is strictly `commit -m <msg>` with no `--amend` or history rewrite flags.
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("commit_changes: nil handler or pipeline")
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
		Action:          "commit_changes",
		Target:          auditTargetFromMessage(p.Message),
		WorkDir:         abs,
		GitArgs:         []string{"commit", "-m", p.Message},
		DryRunRequested: p.DryRun,
		Validate: func() error {
			return policy.ValidateCommitMessage(p.Message)
		},
	}

	res, err := h.Pipeline.Execute(ctx, mr)
	out := toolResultFromRuntime(res, p, err)
	return out, err
}

func auditTargetFromMessage(msg string) string {
	n := strings.ReplaceAll(msg, "\r\n", "\n")
	parts := strings.Split(n, "\n")
	line := ""
	if len(parts) > 0 {
		line = strings.TrimSpace(parts[0])
	}
	if line == "" {
		line = "(empty subject)"
	}
	const maxRunes = 200
	runes := []rune(line)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "..."
	}
	return line
}

func toolResultFromRuntime(r runtime.Result, p Params, execErr error) ToolResult {
	out := ToolResult{
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
