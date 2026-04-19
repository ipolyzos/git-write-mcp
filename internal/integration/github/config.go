// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"strings"
)

// Repository identifies a GitHub repository for REST API paths.
type Repository struct {
	Owner string
	Name  string
}

// Config holds credentials and repository metadata for GitHub REST calls.
// Never log [Config.Token]; use [Config.RedactedSummary] for diagnostics.
type Config struct {
	Repository Repository
	Token      string
}

// Validate checks required fields for an authenticated API client.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Repository.Owner) == "" {
		return &ValidationError{Field: "owner", Msg: "required"}
	}
	if strings.TrimSpace(c.Repository.Name) == "" {
		return &ValidationError{Field: "name", Msg: "required"}
	}
	if c.Token == "" {
		return &ValidationError{Field: "token", Msg: "required"}
	}
	return nil
}

// RedactedSummary returns a deterministic, log-safe description (no secrets).
func (c Config) RedactedSummary() string {
	tok := "<empty>"
	if c.Token != "" {
		tok = "<set>"
	}
	return fmt.Sprintf(
		`owner=%q name=%q token=%s`,
		strings.TrimSpace(c.Repository.Owner),
		strings.TrimSpace(c.Repository.Name),
		tok,
	)
}
