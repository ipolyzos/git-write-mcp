// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package policy

import "errors"

// ErrApprovalRequired is returned when a mutating git invocation has no Approver or approval was denied.
var ErrApprovalRequired = errors.New("policy: approval required for this git invocation")
