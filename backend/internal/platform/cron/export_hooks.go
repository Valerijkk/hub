//go:build hubtest

package cron

import (
	"context"
	"time"
)

// ExportDefaultTickTimeout returns the default per-tick budget (W4-A4-008).
func ExportDefaultTickTimeout() time.Duration {
	return defaultTickTimeout
}

// ExportRunTick wraps runTick for hubtest (timeout / hang path).
func ExportRunTick(ctx context.Context, tickTimeout time.Duration, name string, fn func(ctx context.Context) error) {
	runTick(ctx, tickTimeout, name, fn)
}
