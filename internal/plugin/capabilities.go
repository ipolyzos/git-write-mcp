// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import "encoding/json"

// ToolGetCapabilities is the MCP tool name for the capabilities document.
const ToolGetCapabilities = "get_capabilities"

// Capabilities is the stable JSON document returned by [CapabilitiesJSON] and the
// get_capabilities MCP tool.
type Capabilities struct {
	WriteTools        []string  `json:"write_tools"`
	Behaviors         Behaviors `json:"behaviors"`
	ExcludedReadTools []string  `json:"excluded_read_tools"`
}

// Behaviors lists cross-cutting safety features supported by git-write-mcp write tools.
type Behaviors struct {
	ApprovalRequired  bool `json:"approval_required"`
	AuditTrailEnabled bool `json:"audit_trail_enabled"`
	DryRunSupported   bool `json:"dry_run_supported"`
}

// DocumentCapabilities returns the canonical capability metadata for this server.
// Slice fields are lexicographically sorted for stable JSON.
func DocumentCapabilities() Capabilities {
	return Capabilities{
		WriteTools: []string{"commit_changes", "create_branch", "create_pull_request", "push_branch", "stage_files", "tag_revision", "unstage_files"},
		Behaviors: Behaviors{
			ApprovalRequired:  true,
			AuditTrailEnabled: true,
			DryRunSupported:   true,
		},
		ExcludedReadTools: []string{
			"git_blame",
			"git_branch_list",
			"git_diff",
			"git_log",
			"git_show",
			"git_status",
		},
	}
}

// CapabilitiesJSON marshals [DocumentCapabilities] with [encoding/json.Marshal]
// (deterministic field order from struct layout).
func CapabilitiesJSON() ([]byte, error) {
	return json.Marshal(DocumentCapabilities())
}
