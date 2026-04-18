// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package pathcheck_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/pathcheck"
)

func TestMustBeContainedUnderRoot_RejectsTraversal(t *testing.T) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "repo")
	if err := mkdirTempRepo(t, root); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "..", "outside")
	if err := pathcheck.MustBeContainedUnderRoot(root, outside); err == nil {
		t.Fatal("expected error for escaped path")
	}
}

func TestMustBeContainedUnderRoot_AllowsInside(t *testing.T) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "repo")
	if err := mkdirTempRepo(t, root); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "pkg", "a.go")
	if err := pathcheck.MustBeContainedUnderRoot(root, sub); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func mkdirTempRepo(t *testing.T, root string) error {
	t.Helper()
	return os.MkdirAll(filepath.Join(root, "pkg"), 0o755)
}
