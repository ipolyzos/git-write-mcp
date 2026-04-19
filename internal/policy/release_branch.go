// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvReleaseSourceBranch sets the git branch name that must be checked out locally
// when creating a GitHub release (default "main").
const EnvReleaseSourceBranch = "GIT_WRITE_MCP_RELEASE_SOURCE_BRANCH"

// ReleaseSourceBranch returns the configured release provenance branch name.
func ReleaseSourceBranch() string {
	s := strings.TrimSpace(os.Getenv(EnvReleaseSourceBranch))
	if s == "" {
		return "main"
	}
	return s
}

// RequireHEADMatchesReleaseSource ensures the repository HEAD is on the configured
// release branch (see [ReleaseSourceBranch]) before GitHub release creation.
func RequireHEADMatchesReleaseSource(ctx context.Context, probe RepoGitProbe, workDir string) error {
	if probe == nil {
		return fmt.Errorf("policy: nil repo probe for release branch check")
	}
	want := ReleaseSourceBranch()
	got, err := probe.CurrentBranchName(ctx, workDir)
	if err != nil {
		return err
	}
	if got != want {
		return &ValidationError{
			Field: "git_branch",
			Msg:   fmt.Sprintf("must be %q for release from main workflow (current %q)", want, got),
		}
	}
	return nil
}

// ComposeReleaseNotes prefers non-empty release_notes; otherwise uses changelog_section.
func ComposeReleaseNotes(releaseNotes, changelogSection string) string {
	rn := strings.TrimSpace(releaseNotes)
	if rn != "" {
		return rn
	}
	return strings.TrimSpace(changelogSection)
}
