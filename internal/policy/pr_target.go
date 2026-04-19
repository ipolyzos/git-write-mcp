// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"os"
	"sort"
	"strings"
)

// EnvAllowedPRTargetBranches lists comma-separated base branches allowed for
// create_pull_request (default: develop only).
const EnvAllowedPRTargetBranches = "GIT_WRITE_MCP_ALLOWED_PR_TARGET_BRANCHES"

// AllowedPRTargetBranches returns the de-duplicated, sorted allowlist from
// [EnvAllowedPRTargetBranches], or ["develop"] when unset or only whitespace.
func AllowedPRTargetBranches() []string {
	raw := strings.TrimSpace(os.Getenv(EnvAllowedPRTargetBranches))
	if raw == "" {
		return []string{"develop"}
	}
	seen := make(map[string]struct{})
	for _, p := range strings.Split(raw, ",") {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		seen[s] = struct{}{}
	}
	if len(seen) == 0 {
		return []string{"develop"}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ValidatePRTargetBranch checks the name with [ValidateBranchName] and the allowlist.
func ValidatePRTargetBranch(name string, allowed []string) error {
	if err := ValidateBranchName(name); err != nil {
		var ve *ValidationError
		if errors.As(err, &ve) && ve != nil {
			cp := *ve
			cp.Field = "target_branch"
			return &cp
		}
		return err
	}
	n := strings.TrimSpace(name)
	for _, a := range allowed {
		if a == n {
			return nil
		}
	}
	return &ValidationError{Field: "target_branch", Msg: "not in allowed target branches list"}
}
