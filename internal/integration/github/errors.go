// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "errors"

// ErrPullRequestNotFound means the pulls API returned 404 for that number.
var ErrPullRequestNotFound = errors.New("github: pull request not found")
