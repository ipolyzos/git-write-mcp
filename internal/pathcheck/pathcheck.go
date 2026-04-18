// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

// Package pathcheck enforces filesystem containment relative to a repository root.
package pathcheck

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ErrOutsideRoot indicates candidate path escapes the repository root.
var ErrOutsideRoot = fmt.Errorf("pathcheck: path outside repository root")

// MustBeContainedUnderRoot reports whether candidate resolves to a path equal to root or inside it.
// Both root and candidate may be absolute or relative; they are resolved with filepath.Abs.
func MustBeContainedUnderRoot(root string, candidate string) error {
	rootAbs, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return fmt.Errorf("pathcheck: root: %w", err)
	}
	candAbs, err := filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return fmt.Errorf("pathcheck: candidate: %w", err)
	}
	rel, err := filepath.Rel(rootAbs, candAbs)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOutsideRoot, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrOutsideRoot
	}
	return nil
}
