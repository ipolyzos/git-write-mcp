// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxRefLength = 240

// MaxCommitMessageRunes is the maximum allowed length of a commit message body.
const MaxCommitMessageRunes = 65536

// ValidationError describes a rejected string input (branch, tag, or ref).
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("policy: invalid %s: %s", e.Field, e.Msg)
}

// ValidateBranchName applies conservative rules aligned with safe branch naming (no traversal, no ref punctuation).
func ValidateBranchName(name string) error {
	if err := validateRefLike("branch name", name, true); err != nil {
		return err
	}
	return nil
}

// ValidateTagName applies conservative rules for tag identifiers.
func ValidateTagName(name string) error {
	if err := validateRefLike("tag name", name, false); err != nil {
		return err
	}
	return nil
}

// ValidateRefName validates a generic ref string length and dangerous characters.
func ValidateRefName(ref string) error {
	if err := validateRefLike("ref", ref, false); err != nil {
		return err
	}
	return nil
}

// ValidateCommitMessage enforces non-empty trimmed content, length, NUL absence,
// and rejects most control characters (multiline bodies may use newline/tab/CR).
func ValidateCommitMessage(msg string) error {
	s := strings.TrimSpace(msg)
	if s == "" {
		return &ValidationError{Field: "message", Msg: "required"}
	}
	if utf8.RuneCountInString(msg) > MaxCommitMessageRunes {
		return &ValidationError{Field: "message", Msg: "exceeds maximum length"}
	}
	if strings.IndexByte(msg, 0) >= 0 {
		return &ValidationError{Field: "message", Msg: "must not contain NUL"}
	}
	for _, r := range msg {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return &ValidationError{Field: "message", Msg: "contains forbidden control characters"}
		}
	}
	return nil
}

func validateRefLike(field string, name string, branchRules bool) error {
	if name == "" {
		return &ValidationError{Field: field, Msg: "must be non-empty"}
	}
	if utf8.RuneCountInString(name) > maxRefLength {
		return &ValidationError{Field: field, Msg: "exceeds maximum length"}
	}
	if strings.Contains(name, "..") {
		return &ValidationError{Field: field, Msg: `must not contain ".."`}
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return &ValidationError{Field: field, Msg: "must not be an absolute path"}
	}
	for _, r := range name {
		switch {
		case r == '~' || r == '^' || r == ':' || r == '?' || r == '*' || r == '[':
			return &ValidationError{Field: field, Msg: "contains forbidden ref punctuation"}
		case unicode.IsControl(r):
			return &ValidationError{Field: field, Msg: "contains control characters"}
		case r == ' ':
			return &ValidationError{Field: field, Msg: "must not contain spaces"}
		}
	}
	if branchRules && strings.HasSuffix(name, ".lock") {
		return &ValidationError{Field: field, Msg: `must not end with ".lock"`}
	}
	return nil
}
