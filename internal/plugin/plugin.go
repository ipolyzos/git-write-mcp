// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package plugin defines the MCP tool plugin surface and registration.
package plugin

// Plugin is a thin tool binding (bootstrap: name only; logic comes later).
type Plugin interface {
	Name() string
}
