// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package unstage_files

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
	"github.com/ipolyzos/git-write-mcp/internal/runtime"
)

// Params holds inputs for unstage_files (MCP JSON and tests).
type Params struct {
	Paths  []string `json:"paths"`
	DryRun bool     `json:"dry_run,omitempty"`
}

// Handler runs unstage_files through the shared mutation [runtime.Pipeline].
type Handler struct {
	Pipeline *runtime.Pipeline
	Actor    string
}

// ToolResult is the deterministic JSON-shaped output for MCP and tests.
type ToolResult struct {
	OK              bool                   `json:"ok"`
	Error           string                 `json:"error,omitempty"`
	Paths           []string               `json:"paths"`
	DryRunRequested bool                   `json:"dry_run_requested"`
	Mode            string                 `json:"mode,omitempty"`
	DryRun          bool                   `json:"dry_run,omitempty"`
	DryRunPreview   *runtime.DryRunPreview `json:"dry_run_preview,omitempty"`
	GitStdout       string                 `json:"git_stdout,omitempty"`
	GitStderr       string                 `json:"git_stderr,omitempty"`
}

// Execute validates paths, resolves them inside the repository root, and removes
// them from the index with git reset HEAD -- <paths> (working tree unchanged).
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("unstage_files: nil handler or pipeline")
	}
	if h.Pipeline.Policy == nil {
		return ToolResult{}, errors.New("unstage_files: nil policy engine")
	}
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return ToolResult{}, err
	}
	actor := strings.TrimSpace(h.Actor)
	if actor == "" {
		actor = "mcp"
	}

	var resolved []string
	mr := &runtime.MutationRequest{
		Actor:           actor,
		Action:          "unstage_files",
		Target:          strings.Join(p.Paths, ","),
		WorkDir:         absRoot,
		DryRunRequested: p.DryRun,
	}
	mr.Validate = func() error {
		rels, err := normalizeUnstagePaths(h.Pipeline.Policy, absRoot, p.Paths)
		if err != nil {
			return err
		}
		resolved = rels
		argv := append([]string{"reset", "HEAD", "--"}, rels...)
		mr.GitArgs = argv
		mr.Target = strings.Join(rels, ",")
		return nil
	}

	res, err := h.Pipeline.Execute(ctx, mr)
	out := toolResultFromRuntime(res, resolved, p.DryRun, err)
	return out, err
}

func normalizeUnstagePaths(eng *policy.Engine, repoRoot string, paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, &policy.ValidationError{Field: "paths", Msg: "required"}
	}
	seen := make(map[string]struct{})
	var rels []string
	for _, raw := range paths {
		rel, err := resolveOnePath(eng, repoRoot, raw)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[rel]; dup {
			continue
		}
		seen[rel] = struct{}{}
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	return rels, nil
}

func resolveOnePath(eng *policy.Engine, repoRoot string, raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", &policy.ValidationError{Field: "paths", Msg: "empty path element"}
	}
	var candidate string
	if filepath.IsAbs(s) {
		candidate = filepath.Clean(s)
	} else {
		candidate = filepath.Join(repoRoot, filepath.Clean(s))
	}
	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if err := eng.MustBeContainedUnderRepo(absPath); err != nil {
		return "", fmt.Errorf("policy: paths: outside repository root: %w", err)
	}
	rel, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return "", err
	}
	return rel, nil
}

func toolResultFromRuntime(r runtime.Result, resolved []string, dryRequested bool, execErr error) ToolResult {
	out := ToolResult{
		Paths:           resolved,
		DryRunRequested: dryRequested,
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
		out.DryRunPreview = r.DryRunPreview
	} else if out.OK {
		out.Mode = "execute"
	}
	return out
}
