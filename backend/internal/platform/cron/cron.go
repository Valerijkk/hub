// Package cron реализует общий runner для долгоиграющих периодических задач.
//
// До LO-1 в cmd/server/main.go было ~7 копий одного и того же паттерна:
//
// асинхронный запуск фоновой задачи
//	    ticker := time.NewTicker(interval)
//	    defer ticker.Stop()
// бесконечный цикл cron-задачи
//	        select {
//	        case <-ticker.C:
//	            // выполнение работы
//	        case <-rootCtx.Done():
//	            return
//	        }
//	    }
//	}()
//
// Schedule инкапсулирует этот паттерн в одну строку:
//
//	cron.Schedule(rootCtx, 7*24*time.Hour, "weekly-backup",
// обёртка для cron: PerformBackup требует context.
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

// defaultTickTimeout — бюджет одного tick'а (W4-A4-008). Зависший fn
// не блокирует следующие тики этой задачи навсегда: контекст отменяется,
// ошибка логируется, Schedule продолжает цикл.
const defaultTickTimeout = 10 * time.Minute

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
	ScheduleWithTimeout(ctx, interval, defaultTickTimeout, name, fn)
}

// ScheduleWithTimeout — как Schedule, но с явным бюджетом одного tick'а.
func ScheduleWithTimeout(ctx context.Context, interval, tickTimeout time.Duration, name string, fn func(ctx context.Context) error) {
	if tickTimeout <= 0 {
		tickTimeout = defaultTickTimeout
	}
	go func() {
		slog.Info("cron task started", "name", name, "interval", interval.String(), "tick_timeout", tickTimeout.String())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-ticker.C:
				// Если ticker.C и ctx.Done() готовы одновременно, select выбирает
				// случайно. Preflight не даёт queued tick запустить fn после отмены.
				if ctx.Err() != nil {
					break loop
				}
				runTick(ctx, tickTimeout, name, fn)
			}
		}
		slog.Info("cron task stopped", "name", name)
	}()
}

// runTick выполняет одну итерацию fn с локальным recover'ом и WithTimeout
// (W4-A4-008), чтобы паника/hang в задаче не уносили горутину Schedule.
func runTick(ctx context.Context, tickTimeout time.Duration, name string, fn func(ctx context.Context) error) {
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

	tickCtx, cancel := context.WithTimeout(ctx, tickTimeout)
	defer cancel()

	if err := fn(tickCtx); err != nil {
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
