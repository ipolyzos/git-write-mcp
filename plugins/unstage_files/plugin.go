// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package unstage_files registers the unstage_files MCP tool and handler.
package unstage_files

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(meta{})
}

type meta struct{}

func (meta) Name() string { return "unstage_files" }
