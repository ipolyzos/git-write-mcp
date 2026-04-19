// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package server wires MCP startup over stdio and mirrors registered [plugin.Plugin] tools on the MCP server (stub handlers until tools are implemented).
package server

import (
	"github.com/ipolyzos/git-write-mcp/internal/plugin"

	_ "github.com/ipolyzos/git-write-mcp/plugins/commit_changes"
	_ "github.com/ipolyzos/git-write-mcp/plugins/create_branch"
	_ "github.com/ipolyzos/git-write-mcp/plugins/create_pull_request"
	_ "github.com/ipolyzos/git-write-mcp/plugins/create_release_from_main"
	_ "github.com/ipolyzos/git-write-mcp/plugins/get_pull_request_status"
	_ "github.com/ipolyzos/git-write-mcp/plugins/merge_pull_request"
	_ "github.com/ipolyzos/git-write-mcp/plugins/push_branch"
	_ "github.com/ipolyzos/git-write-mcp/plugins/stage_files"
	_ "github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
	_ "github.com/ipolyzos/git-write-mcp/plugins/unstage_files"
)

// Bootstrap exposes server wiring for tests and future stdio integration.
type Bootstrap struct {
	reg *plugin.Registry
}

// NewBootstrap returns the default bootstrap using the process-wide plugin registry.
func NewBootstrap() *Bootstrap {
	return &Bootstrap{reg: plugin.Default}
}

// PluginNames returns registered MCP tool names (sorted).
func (b *Bootstrap) PluginNames() []string {
	if b == nil || b.reg == nil {
		return nil
	}
	return b.reg.Names()
}
