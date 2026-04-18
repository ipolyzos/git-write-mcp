// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package tag_revision registers the tag_revision tool (logic not implemented).
package tag_revision

import "github.com/ipolyzos/git-write-mcp/internal/plugin"

func init() {
	plugin.Default.Register(stub{})
}

type stub struct{}

func (stub) Name() string { return "tag_revision" }
