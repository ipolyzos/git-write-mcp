// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package audit

import "context"

// Sink receives immutable audit records (structured logging, append-only stores, tests).
type Sink interface {
	Emit(ctx context.Context, r Record) error
}

// SinkFunc adapts a function to [Sink].
type SinkFunc func(ctx context.Context, r Record) error

// Emit implements [Sink].
func (f SinkFunc) Emit(ctx context.Context, r Record) error {
	return f(ctx, r)
}
