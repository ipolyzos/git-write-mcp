// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

import "strings"

// isDestructive reports argv (git arguments only) that must be blocked by default.
func isDestructive(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	cmd := argv[0]
	switch cmd {
	case "reset", "rebase", "filter-branch", "filter-repo", "gc", "maintenance", "update-ref":
		return true
	case "clean":
		return hasAnyFlag(argv[1:], "-x", "-f", "-fd", "-ffd", "-dff", "-fx", "-df", "-xd")
	case "push":
		return hasAnyFlag(argv[1:], "--force", "-f", "--mirror", "--delete") || hasForceEquals(argv[1:])
	case "pull":
		return hasAnyFlag(argv[1:], "--force", "-f")
	case "commit":
		return hasAnyFlag(argv[1:], "--amend")
	case "tag":
		return hasAnyFlag(argv[1:], "-f", "--force")
	case "branch":
		return hasAnyFlag(argv[1:], "-D")
	case "checkout", "switch":
		return hasAnyFlag(argv[1:], "-f", "--force", "--orphan")
	case "stash":
		return subCmd(argv[1:], "clear") || subCmd(argv[1:], "drop")
	default:
		return false
	}
}

func hasAnyFlag(args []string, needles ...string) bool {
	for _, a := range args {
		for _, n := range needles {
			if a == n || strings.HasPrefix(a, n+"=") {
				return true
			}
		}
	}
	return false
}

func hasForceEquals(args []string) bool {
	for _, a := range args {
		if strings.HasPrefix(a, "--force=") {
			return true
		}
	}
	return false
}

func subCmd(args []string, name string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == name
}
