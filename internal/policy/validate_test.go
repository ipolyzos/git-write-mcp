// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"strings"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestValidateBranchName_AcceptsCommonNames(t *testing.T) {
	t.Helper()
	for _, name := range []string{"main", "develop", "feature/foo", "release/1.x"} {
		if err := policy.ValidateBranchName(name); err != nil {
			t.Fatalf("ValidateBranchName(%q): %v", name, err)
		}
	}
}

func TestValidateBranchName_RejectsDotDot(t *testing.T) {
	t.Helper()
	if err := policy.ValidateBranchName(".."); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateBranchName_RejectsAbsolutePath(t *testing.T) {
	t.Helper()
	if err := policy.ValidateBranchName("/etc/passwd"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateTagName_RejectsInvalid(t *testing.T) {
	t.Helper()
	if err := policy.ValidateTagName(""); err == nil {
		t.Fatal("expected error")
	}
	if err := policy.ValidateTagName("bad tag"); err == nil {
		t.Fatal("expected error for space")
	}
}

func TestValidateTagName_AcceptsSemverLike(t *testing.T) {
	t.Helper()
	if err := policy.ValidateTagName("v1.2.3"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRefName_RejectsLongString(t *testing.T) {
	t.Helper()
	long := strings.Repeat("a", 300)
	if err := policy.ValidateRefName(long); err == nil {
		t.Fatal("expected error for excessive length")
	}
}
