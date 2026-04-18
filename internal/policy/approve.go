// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import "context"

// InvocationRequest describes a planned git subprocess (argv after the git binary name).
type InvocationRequest struct {
	WorkDir string
	Args    []string
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
