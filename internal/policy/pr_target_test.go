// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func TestAllowedPRTargetBranches_Default(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "")
	got := policy.AllowedPRTargetBranches()
	want := []string{"develop"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAllowedPRTargetBranches_FromEnv(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop, main ")
	got := policy.AllowedPRTargetBranches()
	want := []string{"develop", "main"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestValidatePRTargetBranch_Allowed(t *testing.T) {
	t.Helper()
	if err := policy.ValidatePRTargetBranch("develop", []string{"develop"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePRTargetBranch_NotAllowed(t *testing.T) {
	t.Helper()
	err := policy.ValidatePRTargetBranch("evil", []string{"develop"})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "target_branch" {
		t.Fatalf("got %v", err)
	}
}
