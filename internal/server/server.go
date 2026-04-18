// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package server wires stdio MCP startup (bootstrap: plugin registration only).
package server

import (
	"github.com/ipolyzos/git-write-mcp/internal/plugin"

	_ "github.com/ipolyzos/git-write-mcp/plugins/commit_changes"
	_ "github.com/ipolyzos/git-write-mcp/plugins/create_branch"
	_ "github.com/ipolyzos/git-write-mcp/plugins/stage_files"
	_ "github.com/ipolyzos/git-write-mcp/plugins/tag_revision"
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
