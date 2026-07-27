package main

import (
	"fmt"
	"log/slog"

	"github.com/valerii/hub/backend/internal/platform/lifecycle"
)

func runServer(d *serverDeps) {
	app := d.app
	cfg := d.cfg
	rootCtx := d.rootCtx

	// --- Запуск сервера ---

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		slog.Info("starting server", "addr", addr)
		if err := app.Listen(addr); err != nil {
			// A8-008: os.Exit обходит defer runCleanups — отменяем rootCtx,
			// чтобы main дошёл до Shutdown + LIFO cleanup (NATS/Redis/PG/Mongo).
			lifecycle.CancelOnListenError(d.rootCancel, err)
		}
	}()

	// ME-10 + ME-14: ждём отмены rootCtx (SIGINT/SIGTERM из NotifyContext
	// или ручной rootCancel при критической ошибке).
	<-rootCtx.Done()

	slog.Info("shutting down server")
	// rootCancel уже выполнен — все долгоиграющие горутины (outbox, cron,
	// notification/audit consumer, dispatcher, wsHub) одновременно начали
	// graceful-shutdown через свои select на rootCtx.Done().
	// app.Shutdown сливает in-flight HTTP-запросы.
	if err := app.Shutdown(); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}
