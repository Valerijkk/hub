package main

import (
	"log/slog"
	"os"

	auditmod "github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/platform/database"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	redisclient "github.com/valerii/hub/backend/internal/platform/redis"
)

func wireInfra(d *serverDeps) {
	// --- Инфраструктурные клиенты ---

	// PostgreSQL + Ent
	// BE-H5: rawDB разделяет connection pool с Ent — переиспользуется
	// в /health probe, чтобы не открывать новое подключение на каждом запросе.
	db, rawDB, err := database.NewPostgresClient(d.ctx, d.cfg.Postgres.URL)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	d.db = db
	d.rawDB = rawDB
	d.addCleanup(func() { db.Close() })

	// Клиент Redis
	rdb, err := redisclient.NewClient(d.ctx, d.cfg.Redis.URL)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	d.rdb = rdb
	d.addCleanup(func() { rdb.Close() })

	// NATS
	nc, err := natsclient.NewClient(d.ctx, d.cfg.NATS.URL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	d.nc = nc
	d.addCleanup(func() { nc.Close() })

	// Создаём единый JetStream stream EVENTS для всех модулей.
	if err := nc.EnsureStreams(d.ctx); err != nil {
		slog.Error("failed to ensure JetStream streams", "error", err)
		os.Exit(1)
	}

	// Outbox worker (ctx == rootCtx — отменится по SIGINT/SIGTERM)
	outboxWorker := outbox.NewWorker(db, nc)
	go outboxWorker.ProcessLoop(d.rootCtx)
	d.outboxWorker = outboxWorker

	// Audit service (ранняя инициализация — нужен для HTTP-audit middleware на /api).
	auditRepo := auditmod.NewRepository(db)
	auditSvc := auditmod.NewAuditService(auditRepo, rdb)
	d.auditRepo = auditRepo
	d.auditSvc = auditSvc
}
