// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

type headNameProbe struct {
	cur string
	err error
}

func (n headNameProbe) GitRepositoryOK(context.Context, string) error { return nil }
func (n headNameProbe) IsDetachedHEAD(context.Context, string) (bool, error) {
	return false, nil
}
func (n headNameProbe) RefExists(context.Context, string, string) (bool, error) { return false, nil }
func (n headNameProbe) HasStagedChanges(context.Context, string) (bool, error)  { return false, nil }
func (n headNameProbe) CurrentBranchName(context.Context, string) (string, error) {
	return n.cur, n.err
}

func TestReleaseSourceBranch_Default(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvReleaseSourceBranch, "")
	if policy.ReleaseSourceBranch() != "main" {
		t.Fatal(policy.ReleaseSourceBranch())
	}
}

func TestComposeReleaseNotes_PrefersReleaseNotes(t *testing.T) {
	t.Helper()
	got := policy.ComposeReleaseNotes("alpha", "beta")
	if got != "alpha" {
		t.Fatal(got)
	}
	got = policy.ComposeReleaseNotes("", "beta")
	if got != "beta" {
		t.Fatal(got)
	}
}

func TestRequireHEADMatchesReleaseSource_OK(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvReleaseSourceBranch, "")
	ctx := context.Background()
	p := headNameProbe{cur: "main"}
	if err := policy.RequireHEADMatchesReleaseSource(ctx, p, "/tmp/x"); err != nil {
		t.Fatal(err)
	}
}

func TestRequireHEADMatchesReleaseSource_WrongBranch(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvReleaseSourceBranch, "")
	ctx := context.Background()
	p := headNameProbe{cur: "develop"}
	err := policy.RequireHEADMatchesReleaseSource(ctx, p, "/tmp/x")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *policy.ValidationError
	if !errors.As(err, &ve) || ve.Field != "git_branch" {
		t.Fatalf("%v", err)
	}
}
