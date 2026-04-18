// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package git

import "fmt"

// Kind classifies git.Executor errors.
type Kind uint8

const (
	// KindNotAllowed means argv is not on the allowlist.
	KindNotAllowed Kind = iota + 1
	// KindDestructiveBlocked means argv matches blocked destructive patterns.
	KindDestructiveBlocked
	// KindOutsideRepo means WorkDir or path escapes the repository root.
	KindOutsideRepo
	// KindInvalidRepoRoot means RepoRoot could not be used as a secure base.
	KindInvalidRepoRoot
)

// KindError is returned for policy violations before git runs.
type KindError struct {
	Kind Kind
	Msg  string
}

func (e *KindError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("git: %s", e.Msg)
}
