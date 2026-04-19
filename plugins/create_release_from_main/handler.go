// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package create_release_from_main

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

// Params holds MCP JSON for create_release_from_main.
type Params struct {
	TagName          string `json:"tag_name"`
	ReleaseName      string `json:"release_name"`
	ChangelogSection string `json:"changelog_section,omitempty"`
	ReleaseNotes     string `json:"release_notes,omitempty"`
	DryRun           bool   `json:"dry_run,omitempty"`
}

// Handler runs the tool through [runtime.Pipeline] and [github.MCPGitHub].
type Handler struct {
	Pipeline    *runtime.Pipeline
	GitHub      github.MCPGitHub
	GitHubOwner string
	GitHubRepo  string
	Actor       string
}

// ToolResult is deterministic JSON for MCP and tests.
type ToolResult struct {
	OK               bool                   `json:"ok"`
	Error            string                 `json:"error,omitempty"`
	TagName          string                 `json:"tag_name"`
	ReleaseName      string                 `json:"release_name"`
	ReleaseID        int                    `json:"release_id,omitempty"`
	HTMLURL          string                 `json:"html_url,omitempty"`
	ReleaseSourceRef string                 `json:"release_source_ref,omitempty"`
	DryRunRequested  bool                   `json:"dry_run_requested"`
	Mode             string                 `json:"mode,omitempty"`
	DryRun           bool                   `json:"dry_run,omitempty"`
	DryRunPreview    *runtime.DryRunPreview `json:"dry_run_preview,omitempty"`
	GitStdout        string                 `json:"git_stdout,omitempty"`
	GitStderr        string                 `json:"git_stderr,omitempty"`
}

// Execute validates local branch (must match [policy.ReleaseSourceBranch], default main),
// release inputs, then creates a GitHub release targeting that ref.
func (h *Handler) Execute(ctx context.Context, repoRoot string, p Params) (ToolResult, error) {
	if h == nil || h.Pipeline == nil {
		return ToolResult{}, errors.New("create_release_from_main: nil handler or pipeline")
	}
	if h.Pipeline.Policy == nil {
		return ToolResult{}, errors.New("create_release_from_main: nil policy engine")
	}
	if h.Pipeline.RepoProbe == nil {
		return ToolResult{}, errors.New("create_release_from_main: nil repo probe")
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

	tag := strings.TrimSpace(p.TagName)
	title := strings.TrimSpace(p.ReleaseName)
	notes := policy.ComposeReleaseNotes(p.ReleaseNotes, p.ChangelogSection)
	srcRef := policy.ReleaseSourceBranch()

	out := ToolResult{
		TagName:          tag,
		ReleaseName:      title,
		ReleaseSourceRef: srcRef,
		DryRunRequested:  p.DryRun,
	}

	var created *github.Release

	mr := &runtime.MutationRequest{
		Actor:           actor,
		Action:          "create_release_from_main",
		Target:          tag,
		WorkDir:         abs,
		DryRunRequested: p.DryRun,
		GitHubOwner:     owner,
		GitHubRepo:      repo,
	}
	mr.Validate = func() error {
		if h.GitHub == nil {
			return fmt.Errorf("create_release_from_main: GitHub integration not configured (set GITHUB_TOKEN and repository env)")
		}
		if tag == "" {
			return &policy.ValidationError{Field: "tag_name", Msg: "required"}
		}
		if err := policy.ValidateTagName(tag); err != nil {
			var ve *policy.ValidationError
			if errors.As(err, &ve) && ve != nil {
				cp := *ve
				cp.Field = "tag_name"
				return &cp
			}
			return err
		}
		if title == "" {
			return &policy.ValidationError{Field: "release_name", Msg: "required"}
		}
		if strings.TrimSpace(p.ReleaseNotes) == "" && strings.TrimSpace(p.ChangelogSection) == "" {
			return &policy.ValidationError{Field: "release_notes", Msg: "release_notes or changelog_section is required"}
		}
		if err := policy.RequireHEADMatchesReleaseSource(ctx, h.Pipeline.RepoProbe, abs); err != nil {
			return err
		}
		mr.ReleaseTagName = tag
		mr.ReleaseGitHubTitle = title
		mr.ReleaseTargetRef = srcRef
		mr.GitArgs = []string{"create_release_from_main", tag, title}
		body := notes
		tn := tag
		tit := title
		ref := srcRef
		mr.APIExecute = func(ctx context.Context) error {
			r, err := h.GitHub.CreateRelease(ctx, github.CreateReleaseInput{
				TagName:         tn,
				Name:            tit,
				Body:            body,
				TargetCommitish: ref,
			})
			if err != nil {
				return err
			}
			created = r
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
		out.ReleaseID = created.ID
		out.HTMLURL = created.HTMLURL
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
