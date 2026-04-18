// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

import "strings"

// resetUnstageIndexPathsOnly allows only `git reset HEAD -- <path>...` (index-only;
// working tree unchanged). Any other `reset` argv stays blocked by allowlist/destructive checks.
func resetUnstageIndexPathsOnly(argv []string) bool {
	if len(argv) < 4 {
		return false
	}
	if argv[0] != "reset" || argv[1] != "HEAD" || argv[2] != "--" {
		return false
	}
	for _, p := range argv[3:] {
		if strings.TrimSpace(p) == "" {
			return false
		}
		if strings.HasPrefix(p, "-") {
			return false
		}
	}
	return true
}

// allowedFirst maps the first argv token after "git" to permitted subcommands.
// Narrow set; expand deliberately as tools are implemented.
var allowedFirst = map[string]struct{}{
	"add":          {},
	"branch":       {},
	"checkout":     {},
	"commit":       {},
	"merge":        {},
	"push":         {},
	"rev-parse":    {},
	"stash":        {},
	"switch":       {},
	"symbolic-ref": {},
	"tag":          {},
	"version":      {},
}

func allowlisted(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	switch argv[0] {
	case "diff":
		return len(argv) == 3 && argv[1] == "--cached" && argv[2] == "--quiet"
	case "reset":
		return resetUnstageIndexPathsOnly(argv)
	default:
		_, ok := allowedFirst[argv[0]]
		return ok
	}
}
