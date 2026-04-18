// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestNewEngine_RequiresAbsoluteRoot(t *testing.T) {
	t.Helper()
	_, err := policy.NewEngine("relative")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEngine_MustBeContainedUnderRepo(t *testing.T) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	eng, err := policy.NewEngine(absRoot)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(absRoot, "..", "x")
	if err := eng.MustBeContainedUnderRepo(out); err == nil {
		t.Fatal("expected error")
	}
}
