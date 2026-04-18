// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package stage_files registers the stage_files tool (logic not implemented).
package stage_files

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(stub{})
}

type stub struct{}

func (stub) Name() string { return "stage_files" }
