// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

type repoProbe struct {
	ex *Executor
}

// NewRepoProbe returns a [policy.RepoGitProbe] backed by allowlisted read-only git.
func NewRepoProbe(ex *Executor) policy.RepoGitProbe {
	if ex == nil {
		return nil
	}
	return repoProbe{ex: ex}
}

func (p repoProbe) GitRepositoryOK(ctx context.Context, workDir string) error {
	res, err := p.ex.Run(ctx, workDir, []string{"rev-parse", "--is-inside-work-tree"})
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(res.Stdout)) != "true" {
		return fmt.Errorf("git: work tree is not a git repository")
	}
	return nil
}

func (p repoProbe) IsDetachedHEAD(ctx context.Context, workDir string) (bool, error) {
	_, err := p.ex.Run(ctx, workDir, []string{"symbolic-ref", "-q", "HEAD"})
	if err == nil {
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, err
}

func (p repoProbe) RefExists(ctx context.Context, workDir, ref string) (bool, error) {
	verify := ref + "^{commit}"
	_, err := p.ex.Run(ctx, workDir, []string{"rev-parse", "--verify", verify})
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		switch exitErr.ExitCode() {
		case 1, 128:
			return false, nil
		}
	}
	return false, err
}

func (p repoProbe) HasStagedChanges(ctx context.Context, workDir string) (bool, error) {
	_, err := p.ex.Run(ctx, workDir, []string{"diff", "--cached", "--quiet"})
	if err == nil {
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, err
}
