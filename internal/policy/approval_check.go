// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"errors"
	"slices"
	"strings"
)

// ApprovalCheckInput describes a mutation for centralized approval evaluation.
// All mutating MCP tools map their work to this shape before approval.
type ApprovalCheckInput struct {
	Actor           string
	Action          string
	Target          string
	WorkDir         string
	GitArgs         []string
	DryRunRequested bool
}

// ApprovalDenyReason classifies why approval was not granted (empty when allowed).
type ApprovalDenyReason string

const (
	// ApprovalDenyNone means the request was allowed or no denial classification applies.
	ApprovalDenyNone ApprovalDenyReason = ""
	// ApprovalDenyMissingContext means actor or action was empty after trim.
	ApprovalDenyMissingContext ApprovalDenyReason = "missing_context"
	// ApprovalDenyUnsupportedAction means Action is not a known MCP mutation.
	ApprovalDenyUnsupportedAction ApprovalDenyReason = "unsupported_action"
	// ApprovalDenyNoApprover means Approver was nil.
	ApprovalDenyNoApprover ApprovalDenyReason = "no_approver"
	// ApprovalDenyRejected means the configured Approver denied the request.
	ApprovalDenyRejected ApprovalDenyReason = "rejected"
	// ApprovalDenyApproverError means Approver.Approve returned an error.
	ApprovalDenyApproverError ApprovalDenyReason = "approver_error"
)

// ApprovalDecision is a deterministic outcome of [EvaluateApproval] for logging,
// audit detail, and policy extension (e.g. human-in-the-loop backends).
type ApprovalDecision struct {
	Allowed    bool               `json:"allowed"`
	DryRun     bool               `json:"dry_run"`
	DenyReason ApprovalDenyReason `json:"deny_reason,omitempty"`
	DenyDetail string             `json:"deny_detail,omitempty"`
}

// KnownMutationActions returns sorted action names accepted by [EvaluateApproval].
// Keep in sync with get_capabilities write_tools (see internal/contracts tests).
func KnownMutationActions() []string {
	out := make([]string, 0, len(supportedMutationActions))
	for k := range supportedMutationActions {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// Known mutation actions accepted by this server build (extend deliberately).
var supportedMutationActions = map[string]struct{}{
	"create_branch":            {},
	"create_pull_request":      {},
	"create_release_from_main": {},
	"merge_pull_request":       {},
	"push_branch":              {},
	"stage_files":              {},
	"unstage_files":            {},
	"commit_changes":           {},
	"tag_revision":             {},
}

// EvaluateApproval runs the shared approval path: input validation, known-action
// check, then delegation to [Approver]. It does not mutate git or audit.
func EvaluateApproval(ctx context.Context, appr Approver, in ApprovalCheckInput) ApprovalDecision {
	if strings.TrimSpace(in.Actor) == "" || strings.TrimSpace(in.Action) == "" {
		return ApprovalDecision{
			Allowed:    false,
			DenyReason: ApprovalDenyMissingContext,
			DenyDetail: "policy: actor and action are required for approval evaluation",
		}
	}
	if _, ok := supportedMutationActions[in.Action]; !ok {
		return ApprovalDecision{
			Allowed:    false,
			DenyReason: ApprovalDenyUnsupportedAction,
			DenyDetail: "policy: unsupported mutation action: " + in.Action,
		}
	}
	if appr == nil {
		return ApprovalDecision{
			Allowed:    false,
			DenyReason: ApprovalDenyNoApprover,
			DenyDetail: ErrApprovalRequired.Error(),
		}
	}
	inv := InvocationRequest{
		WorkDir:         in.WorkDir,
		Args:            in.GitArgs,
		Actor:           in.Actor,
		Action:          in.Action,
		Target:          in.Target,
		DryRunRequested: in.DryRunRequested,
	}
	d, err := appr.Approve(ctx, inv)
	if err != nil {
		return ApprovalDecision{
			Allowed:    false,
			DenyReason: ApprovalDenyApproverError,
			DenyDetail: err.Error(),
		}
	}
	if !d.Approved {
		return ApprovalDecision{
			Allowed:    false,
			DenyReason: ApprovalDenyRejected,
			DenyDetail: ErrApprovalRequired.Error(),
		}
	}
	return ApprovalDecision{
		Allowed: true,
		DryRun:  d.DryRun,
	}
}

// ErrApprovalFromDecision maps a denial decision to an error for pipeline plumbing.
// Approver errors are returned as wrapped errors; policy denials use [ErrApprovalRequired].
func ErrApprovalFromDecision(dec ApprovalDecision) error {
	if dec.Allowed {
		return nil
	}
	if dec.DenyReason == ApprovalDenyApproverError && dec.DenyDetail != "" {
		return errors.New(dec.DenyDetail)
	}
	return ErrApprovalRequired
}
