// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package audit

import "context"

// Logger forwards structured records to a [Sink].
type Logger struct {
	sink Sink
}

// NewLogger returns a logger that writes to sink (sink may be nil; emits become no-ops).
func NewLogger(sink Sink) *Logger {
	return &Logger{sink: sink}
}

// Emit forwards r to the underlying sink.
func (l *Logger) Emit(ctx context.Context, r Record) error {
	if l == nil || l.sink == nil {
		return nil
	}
	return l.sink.Emit(ctx, r)
}
