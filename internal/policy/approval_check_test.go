// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

func validInput() policy.ApprovalCheckInput {
	return policy.ApprovalCheckInput{
		Actor:           "actor",
		Action:          "create_branch",
		Target:          "feature/x",
		WorkDir:         "/repo",
		GitArgs:         []string{"branch", "feature/x"},
		DryRunRequested: false,
	}
}

func TestEvaluateApproval_ApprovedMutationRequest(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	dec := policy.EvaluateApproval(ctx, policy.AlwaysApprove{}, validInput())
	if !dec.Allowed {
		t.Fatalf("want allowed, got %+v", dec)
	}
	if dec.DryRun {
		t.Fatal("unexpected dry_run from AlwaysApprove")
	}
	if dec.DenyReason != policy.ApprovalDenyNone {
		t.Fatalf("deny_reason = %q", dec.DenyReason)
	}
}

func TestEvaluateApproval_NilApprover(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	dec := policy.EvaluateApproval(ctx, nil, validInput())
	if dec.Allowed {
		t.Fatal("expected denied")
	}
	if dec.DenyReason != policy.ApprovalDenyNoApprover {
		t.Fatalf("deny_reason = %q", dec.DenyReason)
	}
}

func TestEvaluateApproval_DeniedMutationRequest(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	dec := policy.EvaluateApproval(ctx, policy.DenyAll{}, validInput())
	if dec.Allowed {
		t.Fatal("expected denied")
	}
	if dec.DenyReason != policy.ApprovalDenyRejected {
		t.Fatalf("deny_reason = %q, want rejected", dec.DenyReason)
	}
}

func TestEvaluateApproval_MissingApprovalContext(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	for _, in := range []policy.ApprovalCheckInput{
		{},
		{Actor: "a"},
		{Action: "create_branch"},
	} {
		dec := policy.EvaluateApproval(ctx, policy.AlwaysApprove{}, in)
		if dec.Allowed {
			t.Fatalf("expected deny for input %+v", in)
		}
		if dec.DenyReason != policy.ApprovalDenyMissingContext {
			t.Fatalf("deny_reason = %q", dec.DenyReason)
		}
	}
}

func TestEvaluateApproval_UnsupportedActionType(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	in := validInput()
	in.Action = "unknown_mutation"
	dec := policy.EvaluateApproval(ctx, policy.AlwaysApprove{}, in)
	if dec.Allowed {
		t.Fatal("expected unsupported")
	}
	if dec.DenyReason != policy.ApprovalDenyUnsupportedAction {
		t.Fatalf("deny_reason = %q", dec.DenyReason)
	}
}

func TestEvaluateApproval_ApproverError(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	dec := policy.EvaluateApproval(ctx, errApprover{}, validInput())
	if dec.Allowed {
		t.Fatal("expected error path")
	}
	if dec.DenyReason != policy.ApprovalDenyApproverError || dec.DenyDetail != "boom" {
		t.Fatalf("got %+v", dec)
	}
	err := policy.ErrApprovalFromDecision(dec)
	if errors.Is(err, policy.ErrApprovalRequired) {
		t.Fatal("approver errors must not wrap ErrApprovalRequired")
	}
	if err.Error() != "boom" {
		t.Fatalf("got %v", err)
	}
}

func TestEvaluateApproval_DeterministicOutputStructure(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	in := validInput()
	a := policy.EvaluateApproval(ctx, policy.DenyAll{}, in)
	b := policy.EvaluateApproval(ctx, policy.DenyAll{}, in)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("non-deterministic: %+v vs %+v", a, b)
	}
	j1, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(j1) != string(j2) {
		t.Fatalf("json differs: %s vs %s", j1, j2)
	}
}

type errApprover struct{}

func (errApprover) Approve(context.Context, policy.InvocationRequest) (policy.Decision, error) {
	return policy.Decision{}, errors.New("boom")
}
