// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package stage_files registers the stage_files MCP tool and handler.
package stage_files

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(meta{})
}

type meta struct{}

func (meta) Name() string { return "stage_files" }
