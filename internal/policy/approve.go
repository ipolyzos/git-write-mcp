// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import "context"

// InvocationRequest describes a planned git subprocess (argv after the git binary name).
// Actor, Action, Target, and DryRunRequested are optional metadata for centralized approval.
type InvocationRequest struct {
	WorkDir string
	Args    []string

	Actor           string
	Action          string
	Target          string
	DryRunRequested bool
}

// Decision is returned by Approver.Approve. DryRun skips execution while still allowing the request.
type Decision struct {
	Approved bool
	DryRun   bool
}

// Approver is an extension point for human-in-the-loop or policy engines before git runs.
// When nil on an Executor, mutating invocations are rejected with [ErrApprovalRequired].
type Approver interface {
	Approve(ctx context.Context, req InvocationRequest) (Decision, error)
}

// AlwaysApprove grants every invocation (use only in controlled environments).
type AlwaysApprove struct{}

// Approve implements [Approver].
func (AlwaysApprove) Approve(context.Context, InvocationRequest) (Decision, error) {
	return Decision{Approved: true, DryRun: false}, nil
}

// DenyAll rejects every invocation until a real approver is wired.
type DenyAll struct{}

// Approve implements [Approver].
func (DenyAll) Approve(context.Context, InvocationRequest) (Decision, error) {
	return Decision{Approved: false}, nil
}
