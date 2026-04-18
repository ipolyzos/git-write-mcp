// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

// DryRunPreview is a deterministic description of what would execute without mutating git state.
type DryRunPreview struct {
	Action        string   `json:"action"`
	Targets       []string `json:"targets,omitempty"`
	SecondaryRefs []string `json:"secondary_refs,omitempty"`
	GitArgv       []string `json:"git_argv"`
	GitCommands   []string `json:"git_commands"`
	WorkDir       string   `json:"work_dir"`
	EffectSummary string   `json:"effect_summary"`
}
