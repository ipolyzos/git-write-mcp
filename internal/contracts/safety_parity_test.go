// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package contracts holds cross-module invariants (no non-test code here).
package contracts_test

import (
	"slices"
	"testing"

	"github.com/ipolyzos/git-write-mcp/internal/plugin"
	"github.com/ipolyzos/git-write-mcp/internal/policy"
)

// Enforce a single source of truth: every write tool in get_capabilities must be
// an action that [policy.EvaluateApproval] accepts, and vice versa.
func TestWriteToolsMatchPolicyKnownMutationActions(t *testing.T) {
	t.Helper()
	capWrites := plugin.DocumentCapabilities().WriteTools
	policyActs := policy.KnownMutationActions()
	if !slices.Equal(capWrites, policyActs) {
		t.Fatalf("capabilities write_tools %v policy KnownMutationActions %v — keep supportedMutationActions and DocumentCapabilities.WriteTools aligned",
			capWrites, policyActs)
	}
}
