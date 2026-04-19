// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"sort"
	"strings"
)

// EnvAllowedPushRemotes lists comma-separated remote names permitted for
// push_branch (default: origin only). Example: origin,upstream
const EnvAllowedPushRemotes = "GIT_WRITE_MCP_ALLOWED_PUSH_REMOTES"

// AllowedPushRemotes returns the de-duplicated, sorted allowlist from
// [EnvAllowedPushRemotes], or ["origin"] when unset or only whitespace.
func AllowedPushRemotes() []string {
	raw := strings.TrimSpace(os.Getenv(EnvAllowedPushRemotes))
	if raw == "" {
		return []string{"origin"}
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
		return []string{"origin"}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ValidateAllowedPushRemote rejects empty remote or names outside allowed.
func ValidateAllowedPushRemote(remote string, allowed []string) error {
	r := strings.TrimSpace(remote)
	if r == "" {
		return &ValidationError{Field: "remote", Msg: "required"}
	}
	for _, a := range allowed {
		if a == r {
			return nil
		}
	}
	return &ValidationError{Field: "remote", Msg: "not in allowed remotes list"}
}
