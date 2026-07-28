// Package observability предоставляет обёртки для измерения времени операций.
package observability

import (
	"context"
	"log/slog"
	"time"
)

// Timed выполняет fn и логирует длительность в миллисекундах.
// Ошибки — Error; успех — Debug (W4-A3-006 / P3-BE-12: hot-path не шумит в Info).
func Timed[T any](ctx context.Context, operation string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := startOperationSpan(ctx, operation)
	start := time.Now()
	result, err := fn(ctx)
	ms := time.Since(start).Milliseconds()
	recordOperation(operation, ms, err != nil)
	finishOperationSpan(span, err)
	attrs := []any{
		"operation", operation,
		"duration_ms", ms,
	}
	if err != nil {
		slog.ErrorContext(ctx, "operation failed", append(attrs, "error", err)...)
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
