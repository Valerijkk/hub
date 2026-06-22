package main

import (
	"fmt"
	"log/slog"
	"os"
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
			slog.Error("server failed", "error", err)
			os.Exit(1)
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
