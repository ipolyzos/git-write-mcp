// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "fmt"

// ValidationError describes invalid configuration or inputs for the GitHub layer.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("github: invalid %s: %s", e.Field, e.Msg)
}
