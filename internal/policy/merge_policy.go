// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ipolyzos/git-write-mcp/internal/integration/github"
)

// Merge policy is configured via environment variables (see README).
const (
	EnvMergeAllowedMethods           = "GIT_WRITE_MCP_MERGE_ALLOWED_METHODS"
	EnvMergeRequireGreenChecks       = "GIT_WRITE_MCP_MERGE_REQUIRE_GREEN_CHECKS"
	EnvMergeAllowPendingChecks       = "GIT_WRITE_MCP_MERGE_ALLOW_PENDING_CHECKS"
	EnvMergeMinApprovals             = "GIT_WRITE_MCP_MERGE_MIN_APPROVALS"
	EnvMergeBlockOnChangesRequested  = "GIT_WRITE_MCP_MERGE_BLOCK_ON_CHANGES_REQUESTED"
	EnvMergeRequireCleanMergeability = "GIT_WRITE_MCP_MERGE_REQUIRE_CLEAN_MERGEABILITY"
)

// Sentinel errors returned by [EvaluateMergePolicy] for stable tests and branching.
var (
	ErrMergeChecksNotSuccessful   = errors.New("policy: combined checks must succeed before merge")
	ErrMergeChecksPending         = errors.New("policy: pending checks block merge")
	ErrMergeInsufficientApprovals = errors.New("policy: insufficient approving reviews for merge")
	ErrMergeChangesRequested      = errors.New("policy: changes-requested reviews block merge")
	ErrMergeMethodNotAllowed      = errors.New("policy: merge_method not allowed")
	ErrMergeMergeabilityBlocked   = errors.New("policy: merge mergeability_state must be clean")
	ErrMergeDraftPR               = errors.New("policy: cannot merge a draft pull request")
	ErrMergeAlreadyMerged         = errors.New("policy: pull request already merged")
)

// AllowedMergeMethods lists merge_method values permitted by [EnvMergeAllowedMethods]
// (comma-separated). Default when unset or empty: squash only.
func AllowedMergeMethods() []string {
	raw := strings.TrimSpace(os.Getenv(EnvMergeAllowedMethods))
	if raw == "" {
		return []string{"squash"}
	}
	seen := make(map[string]struct{})
	for _, p := range strings.Split(raw, ",") {
		s := strings.TrimSpace(strings.ToLower(p))
		switch s {
		case "merge", "squash", "rebase":
			seen[s] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return []string{"squash"}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// NormalizeMergeMethod returns a non-empty GitHub merge_method value (default squash).
func NormalizeMergeMethod(s string) string {
	v := strings.TrimSpace(strings.ToLower(s))
	switch v {
	case "":
		return "squash"
	case "merge", "squash", "rebase":
		return v
	default:
		return v
	}
}

// EvaluateMergePolicy enforces branch allowlist for the PR base, merge method allowlist,
// GitHub checks/reviews summaries, and mergeability_state against process env vars.
func EvaluateMergePolicy(st *github.PullRequestStatus, mergeMethod string) error {
	if st == nil {
		return errors.New("policy: nil pull request status")
	}
	if st.Merged {
		return ErrMergeAlreadyMerged
	}
	if st.Draft {
		return ErrMergeDraftPR
	}
	base := strings.TrimSpace(st.BaseRef)
	if base == "" {
		return &ValidationError{Field: "base_ref", Msg: "missing from pull request"}
	}
	if err := ValidatePRTargetBranch(base, AllowedPRTargetBranches()); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) && ve != nil {
			cp := *ve
			cp.Field = "base_ref"
			return &cp
		}
		return err
	}
	mm := NormalizeMergeMethod(mergeMethod)
	allowed := AllowedMergeMethods()
	okMethod := false
	for _, m := range allowed {
		if m == mm {
			okMethod = true
			break
		}
	}
	if !okMethod {
		return ErrMergeMethodNotAllowed
	}
	if truthyEnvDefault(EnvMergeRequireCleanMergeability, true) {
		if strings.TrimSpace(strings.ToLower(st.MergeableState)) != "clean" {
			return ErrMergeMergeabilityBlocked
		}
	}

	requireGreen := truthyEnvDefault(EnvMergeRequireGreenChecks, true)
	if requireGreen {
		cs := strings.TrimSpace(strings.ToLower(st.ChecksSummary.State))
		allowPending := truthyEnvDefault(EnvMergeAllowPendingChecks, false)
		switch cs {
		case "success":
			break
		case "pending":
			if !allowPending {
				return ErrMergeChecksPending
			}
		default:
			if cs == "" {
				return ErrMergeChecksNotSuccessful
			}
			return ErrMergeChecksNotSuccessful
		}
	}

	minAppr := mergeMinApprovalsDefault()
	if st.ReviewsSummary.Approved < minAppr {
		return ErrMergeInsufficientApprovals
	}
	if truthyEnvDefault(EnvMergeBlockOnChangesRequested, true) && st.ReviewsSummary.ChangesRequested > 0 {
		return ErrMergeChangesRequested
	}
	return nil
}

func mergeMinApprovalsDefault() int {
	raw := strings.TrimSpace(os.Getenv(EnvMergeMinApprovals))
	if raw == "" {
		return 1
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 1
	}
	return n
}

func truthyEnvDefault(key string, def bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "":
		return def
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
