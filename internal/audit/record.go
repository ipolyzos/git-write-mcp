// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package audit

import "time"

// Record is a structured audit entry for a mutation pipeline step.
type Record struct {
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Mode      string    `json:"mode"`
	Outcome   string    `json:"outcome"`
	Detail    string    `json:"detail,omitempty"`
}

// Mode values for Record.Mode.
const (
	ModeExecute = "execute"
	ModeDryRun  = "dry-run"
	ModeBlocked = "blocked"
)

// Outcome values for Record.Outcome.
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
	OutcomeSkipped = "skipped"
)
