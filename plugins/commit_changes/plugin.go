// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package commit_changes registers the commit_changes tool (logic not implemented).
package commit_changes

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(stub{})
}

type stub struct{}

func (stub) Name() string { return "commit_changes" }
