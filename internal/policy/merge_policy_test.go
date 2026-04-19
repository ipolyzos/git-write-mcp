// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"errors"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func goodStatus() *github.PullRequestStatus {
	return &github.PullRequestStatus{
		Number: 1, State: "open", Merged: false, Draft: false,
		MergeableState: "clean", BaseRef: "develop", HeadRef: "feature/a", HeadSHA: "x",
		ChecksSummary: github.ChecksSummary{State: "success", TotalCount: 1},
		ReviewsSummary: github.ReviewsSummary{
			Approved: 1, ChangesRequested: 0,
		},
	}
}

func TestEvaluateMergePolicy_AllowsWhenSatisfied(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeAllowedMethods, "squash")
	t.Setenv(policy.EnvMergeRequireGreenChecks, "1")
	t.Setenv(policy.EnvMergeMinApprovals, "1")
	t.Setenv(policy.EnvMergeBlockOnChangesRequested, "1")
	t.Setenv(policy.EnvMergeRequireCleanMergeability, "1")

	if err := policy.EvaluateMergePolicy(goodStatus(), "squash"); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateMergePolicy_RejectsChecksNotGreen(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	st := goodStatus()
	st.ChecksSummary.State = "failure"
	st.ChecksSummary.TotalCount = 1

	if err := policy.EvaluateMergePolicy(st, "squash"); err == nil || !errors.Is(err, policy.ErrMergeChecksNotSuccessful) {
		t.Fatalf("want ErrMergeChecksNotSuccessful, got %v", err)
	}
}

func TestEvaluateMergePolicy_RejectsChecksPendingWhenForbidden(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeAllowPendingChecks, "0")

	st := goodStatus()
	st.ChecksSummary.State = "pending"

	if err := policy.EvaluateMergePolicy(st, "squash"); err == nil || !errors.Is(err, policy.ErrMergeChecksPending) {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateMergePolicy_RejectsInsufficientApprovals(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeMinApprovals, "2")

	st := goodStatus()
	st.ReviewsSummary.Approved = 1

	if err := policy.EvaluateMergePolicy(st, "squash"); err == nil || !errors.Is(err, policy.ErrMergeInsufficientApprovals) {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateMergePolicy_RejectsChangesRequested(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	st := goodStatus()
	st.ReviewsSummary.Approved = 2
	st.ReviewsSummary.ChangesRequested = 1

	if err := policy.EvaluateMergePolicy(st, "squash"); err == nil || !errors.Is(err, policy.ErrMergeChangesRequested) {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateMergePolicy_RejectsBadMergeMethod(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")
	t.Setenv(policy.EnvMergeAllowedMethods, "squash")

	if err := policy.EvaluateMergePolicy(goodStatus(), "merge"); err == nil || !errors.Is(err, policy.ErrMergeMethodNotAllowed) {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateMergePolicy_RejectsUncleanMergeability(t *testing.T) {
	t.Helper()
	t.Setenv(policy.EnvAllowedPRTargetBranches, "develop")

	st := goodStatus()
	st.MergeableState = "blocked"

	if err := policy.EvaluateMergePolicy(st, "squash"); err == nil || !errors.Is(err, policy.ErrMergeMergeabilityBlocked) {
		t.Fatalf("got %v", err)
	}
}
