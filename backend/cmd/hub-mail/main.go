package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberrecover "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/valerii/hub/backend/internal/mail"
	redisclient "github.com/valerii/hub/backend/internal/platform/redis"
)

func main() {
	cfg := mail.LoadConfig()
	svc := mail.NewService(cfg)

	// A3W7-1-001: shared DailyLimit с Hub-репликами через Redis (INCR+TTL до UTC midnight).
	// Без REDIS_URL — process-local Memory (warn); с URL — fail-closed на ошибках Redis.
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		rdb, err := redisclient.NewClient(ctx, redisURL)
		cancel()
		if err != nil {
			slog.Error("hub-mail: redis required for distributed daily_limit", "error", err)
			os.Exit(1)
		}
		svc.SetDailySlotStore(mail.NewRedisDailySlotStore(rdb))
		defer func() { _ = rdb.Close() }()
	} else {
		slog.Warn("hub-mail: REDIS_URL unset — daily_limit process-local only")
	}

	handler := mail.NewHubHandler(svc)

	app := fiber.New(fiber.Config{
		AppName: "hub-mail",
	})
	app.Use(fiberrecover.New())

	handler.RegisterRoutes(app)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("hub-mail starting", "addr", addr, "resend_ready", cfg.ResendAPIKey != "")

	go func() {
		if err := app.Listen(addr); err != nil {
			slog.Error("hub-mail listen failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("hub-mail shutting down")
	_ = app.Shutdown()
}
