// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"regexp"
	"testing"
)

func TestVersionIsNonEmptySemverLike(t *testing.T) {
	t.Helper()
	if Version == "" {
		t.Fatal("Version must be non-empty")
	}
	ok, err := regexp.MatchString(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`, Version)
	if err != nil {
		t.Fatalf("regexp: %v", err)
	}
	if !ok {
		t.Fatalf("Version %q must match semver-like pattern", Version)
	}
}
