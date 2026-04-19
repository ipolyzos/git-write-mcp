// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import "context"

// ReleaseCreator creates GitHub releases via the REST API.
type ReleaseCreator interface {
	CreateRelease(ctx context.Context, in CreateReleaseInput) (*Release, error)
}

var _ ReleaseCreator = (*Client)(nil)
