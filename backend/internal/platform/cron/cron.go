// Package cron реализует общий runner для долгоиграющих периодических задач.
//
// До LO-1 в cmd/server/main.go было ~7 копий одного и того же паттерна:
//
//	go func() {
//	    ticker := time.NewTicker(interval)
//	    defer ticker.Stop()
//	    for {
//	        select {
//	        case <-ticker.C:
//	            // do work
//	        case <-rootCtx.Done():
//	            return
//	        }
//	    }
//	}()
//
// Schedule инкапсулирует этот паттерн в одну строку:
//
//	cron.Schedule(rootCtx, 7*24*time.Hour, "weekly-backup",
//	    func(ctx context.Context) error { return backupSvc.PerformBackup(ctx) })
//
// Все запущенные горутины слушают тот же rootCtx и одновременно
// завершаются при SIGINT/SIGTERM (см. ME-10 graceful-shutdown).
package cron

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"
)

// Schedule запускает периодическую задачу в отдельной горутине.
//
// - ctx     — контекст отмены (обычно rootCtx из main); при ctx.Done()
//             горутина завершается без вызова fn.
// - interval — период между tick'ами (первый tick также происходит через
//              interval, как и у time.NewTicker).
// - name    — человеко-читаемое имя задачи для структурных логов.
// - fn      — выполняемая работа. Ошибка логируется на уровне Error;
//             паника перехватывается локальным recover (см. ниже) и
//             логируется со stack-trace — одно сбойное tick'ом не
//             уносит весь периодический job и не крашит сервер.
//
// Schedule логирует:
// - Info "cron task started" при запуске;
// - Debug "cron tick completed" после каждого успешного tick'а с длительностью;
// - Error "cron task failed" при ошибке fn;
// - Info "cron task stopped" при завершении (ctx.Done()).
func Schedule(ctx context.Context, interval time.Duration, name string, fn func(ctx context.Context) error) {
	go func() {
		slog.Info("cron task started", "name", name, "interval", interval.String())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("cron task stopped", "name", name)
				return
			case <-ticker.C:
				runTick(ctx, name, fn)
			}
		}
	}()
}

// runTick выполняет одну итерацию fn с локальным recover'ом, чтобы паника
// в задаче не уносила за собой всю горутину Schedule (а с ней — и весь
// процесс сервера, если паника всплыла бы на runtime). Зеркалит паттерн
// platform/outbox/worker.go:72 (processEvent).
func runTick(ctx context.Context, name string, fn func(ctx context.Context) error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("cron task panicked",
				"name", name,
				"panic", r,
				"stack", string(debug.Stack()),
				"duration", time.Since(start).String(),
			)
		}
	}()
	if err := fn(ctx); err != nil {
		slog.Error("cron task failed",
			"name", name,
			"error", err,
			"duration", time.Since(start).String())
		return
	}
	slog.Debug("cron tick completed",
		"name", name,
		"duration", time.Since(start).String())
}
