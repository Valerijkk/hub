// Package lifecycle — helpers graceful shutdown для cmd/server.
package lifecycle

import (
	"context"
	"log/slog"
)

// CancelOnListenError логирует ошибку app.Listen и отменяет rootCtx.
// Не вызывает os.Exit — иначе defer runCleanups в main не выполняется (A8-008).
func CancelOnListenError(cancel context.CancelFunc, err error) {
	slog.Error("server failed", "error", err)
	if cancel != nil {
		cancel()
	}
}
