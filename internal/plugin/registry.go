// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import "sort"

// Default is populated before plugin init functions run.
var Default *Registry

func init() {
	Default = NewRegistry()
}

// Registry holds registered plugins by tool name.
type Registry struct {
	byName map[string]Plugin
}

// NewRegistry returns an empty registry (tests may use dedicated instances).
func NewRegistry() *Registry {
	return &Registry{byName: make(map[string]Plugin)}
}

// Register adds a plugin. It panics on duplicate or invalid registration.
func (r *Registry) Register(p Plugin) {
	if r == nil {
		panic("plugin: Register on nil Registry")
	}
	if p == nil {
		panic("plugin: Register(nil)")
	}
	name := p.Name()
	if name == "" {
		panic("plugin: empty name")
	}
	if _, exists := r.byName[name]; exists {
		panic("plugin: duplicate registration: " + name)
	}
	r.byName[name] = p
}

// Names returns registered tool names in lexicographic order.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.byName))
	for k := range r.byName {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
