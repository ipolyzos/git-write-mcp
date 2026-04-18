// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

// DryRunPreview is a deterministic description of what would execute without mutating git state.
type DryRunPreview struct {
	GitArgv []string `json:"git_argv"`
	WorkDir string   `json:"work_dir"`
}
