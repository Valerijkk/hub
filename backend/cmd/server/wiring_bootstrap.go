package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valerii/hub/backend/internal/platform/config"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

func initLogger() {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "hub-backend"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", serviceName)
	slog.SetDefault(logger)
}

func bootstrapServer(d *serverDeps) {
	initLogger()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	d.cfg = cfg

	// CR-06: при DevMode видны небезопасные эндпоинты и отключён auth rate-limit.
	// Логируем громкий warning, чтобы это не прошло мимо в production-логах.
	if cfg.Server.DevMode {
		slog.Warn("DEV_MODE enabled — /api/dev/* endpoints exposed, auth rate-limit DISABLED — DO NOT USE IN PRODUCTION",
			"environment", cfg.Server.Environment)
	}

	// HI-05: cipher для шифрования access-токенов и webhook-секретов интеграций.
	// В DevMode допустим пустой INTEGRATION_ENCRYPTION_KEY — используем
	// детерминированный dev-ключ (см. cryptopkg.DeriveDevKey). В production
	// config.validate уже отклонил пустое значение, так что NewCipherFromString
	// получит непустую строку.
	var integrationCipher *cryptopkg.Cipher
	if cfg.Integration.EncryptionKey != "" {
		integrationCipher, err = cryptopkg.NewCipherFromString(cfg.Integration.EncryptionKey)
		if err != nil {
			slog.Error("invalid INTEGRATION_ENCRYPTION_KEY", "error", err)
			os.Exit(1)
		}
	} else {
		// Доступно только при DevMode=true (config.validate).
		slog.Warn("INTEGRATION_ENCRYPTION_KEY not set — using deterministic dev key (DEV ONLY)")
		integrationCipher, err = cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
		if err != nil {
			slog.Error("failed to build dev integration cipher", "error", err)
			os.Exit(1)
		}
	}
	d.integrationCipher = integrationCipher

	// ME-10 + ME-14: единый корневой контекст, прерываемый SIGINT/SIGTERM.
	// Все долгоиграющие горутины (outbox, cron-тикеры, notification/audit
	// консьюмеры, WS hub, webhook dispatcher и т. п.) получают этот ctx —
	// при первом сигнале они все одновременно идут в graceful-shutdown.
	rootCtx, rootCancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	d.addCleanup(rootCancel)
	d.rootCtx = rootCtx
	// ctx сохраняем как алиас для read-only бутстрапных вызовов
	// (database connect, EnsureStreams, EnsureIndexes), которые
	// концептуально не "running forever" — но и здесь использование
	// rootCtx безопасно: если процесс получает сигнал раньше окончания
	// инициализации, бутстрап прервётся, и мы быстрее уйдём в exit.
	d.ctx = rootCtx

	traceShutdown, err := observability.InitTracing(d.ctx, observability.TracingConfig{
		ServiceName: os.Getenv("OTEL_SERVICE_NAME"),
	})
	if err != nil {
		slog.Error("failed to init OpenTelemetry tracing", "error", err)
		os.Exit(1)
	}
	d.traceShutdown = traceShutdown
	d.addCleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := traceShutdown(shutdownCtx); err != nil {
			slog.Warn("OpenTelemetry shutdown error", "error", err)
		}
	})
}
