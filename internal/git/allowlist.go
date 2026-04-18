// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

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
	_, ok := allowedFirst[argv[0]]
	return ok
}
