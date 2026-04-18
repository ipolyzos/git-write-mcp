// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git_test

import (
	"context"
	"testing"

	gitpkg "github.com/ipolyzos/git-write-mcp/internal/git"
)

func TestRunApproved_SkipDoesNotInvokeGitBinaryForMutations(t *testing.T) {
	t.Helper()
	root := initGitRepo(t)
	ex, err := gitpkg.NewExecutor(gitpkg.ExecutorOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	res, err := ex.RunApproved(ctx, root, []string{"rev-parse", "--is-inside-work-tree"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Fatal("expected skipped without executing git")
	}
	if len(res.Stdout) != 0 || len(res.Stderr) != 0 {
		t.Fatalf("unexpected output: %+v", res)
	}
}
