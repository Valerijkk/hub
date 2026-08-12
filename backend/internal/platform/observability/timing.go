// Package observability предоставляет обёртки для измерения времени операций.
package observability

import (
	"context"
	"log/slog"
	"time"
)

// Timed выполняет fn и логирует длительность в миллисекундах.
// Реальные ошибки — Error; успех и штатная отмена (Canceled/DeadlineExceeded) — Debug
// (W4-A3-006 / A3W4-3-001: hot-path и cancel не шумят error_count/slog.Error).
func Timed[T any](ctx context.Context, operation string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := startOperationSpan(ctx, operation)
	start := time.Now()
	result, err := fn(ctx)
	ms := time.Since(start).Milliseconds()
	failed := err != nil && !isBenignContextError(err)
	recordOperation(operation, ms, failed)
	finishOperationSpan(span, err)
	attrs := []any{
		"operation", operation,
		"duration_ms", ms,
	}
	if failed {
		slog.ErrorContext(ctx, "operation failed", append(attrs, "error", err)...)
		return result, err
	}
	if err != nil {
		// A3W4-3-001: cancel/deadline — Debug, не slog.Error (caller всё ещё видит err).
		slog.DebugContext(ctx, "operation canceled", append(attrs, "error", err)...)
		return result, err
	}
	slog.DebugContext(ctx, "operation completed", attrs...)
	return result, nil
}

// TimedVoid — вариант Timed для функций без возвращаемого значения.
func TimedVoid(ctx context.Context, operation string, fn func(context.Context) error) error {
	_, err := Timed(ctx, operation, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}
