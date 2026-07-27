package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/internal/platform/database"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	searchmod "github.com/valerii/hub/backend/internal/search"
	statusmod "github.com/valerii/hub/backend/internal/status"
	tmsmod "github.com/valerii/hub/backend/internal/tms"
)

func wireMessenger(d *serverDeps) {
	cfg := d.cfg
	api := d.api
	requireAuth := d.requireAuth
	rbacSvc := d.rbacSvc
	trackerSvc := d.trackerSvc

	// Маршруты messenger
	mongoDB, err := database.NewMongoClient(d.ctx, cfg.MongoDB.URL, cfg.MongoDB.Database)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	d.mongoDB = mongoDB
	// A8-007: PG/Redis/NATS уже в LIFO cleanup — Mongo тоже должен закрываться.
	d.addCleanup(database.MongoDisconnectCleanup(mongoDB))

	// Health: PostgreSQL, Redis, MongoDB (без auth) — K8s liveness.
	api.Get("/health", func(c fiber.Ctx) error {
		healthCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		services := fiber.Map{}
		overall := "ok"

		// PostgreSQL
		// BE-H5: переиспользуем pool вместо нового
		// *sql.DB на каждый probe — утечка connections под нагрузкой.
		if pingErr := d.rawDB.PingContext(healthCtx); pingErr != nil {
			services["postgres"] = "error"
			overall = "degraded"
		} else {
			services["postgres"] = "ok"
		}

		// Клиент Redis
		if pingErr := d.rdb.Ping(healthCtx).Err(); pingErr != nil {
			services["redis"] = "error"
			overall = "degraded"
		} else {
			services["redis"] = "ok"
		}

		// MongoDB
		if pingErr := mongoDB.Client().Ping(healthCtx, nil); pingErr != nil {
			services["mongo"] = "error"
			overall = "degraded"
		} else {
			services["mongo"] = "ok"
		}

		status := fiber.StatusOK
		if overall != "ok" {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(fiber.Map{
			"status":   overall,
			"services": services,
		})
	})

	statusSvc := statusmod.NewService(
		func(ctx context.Context) error { return d.rawDB.PingContext(ctx) },
		func(ctx context.Context) error { return d.rdb.Ping(ctx).Err() },
		func(ctx context.Context) error { return mongoDB.Client().Ping(ctx, nil) },
		func(ctx context.Context) error {
			if d.nc == nil || d.nc.Conn == nil || !d.nc.Conn.IsConnected() {
				return fmt.Errorf("nats disconnected")
			}
			return nil
		},
	)
	statusmod.NewHandler(statusSvc).RegisterRoutes(api)

	messengerEntRepo := messengermod.NewEntRepository(d.db)
	messengerMongoRepo := messengermod.NewMongoRepository(mongoDB)
	if err := messengerMongoRepo.EnsureIndexes(d.ctx); err != nil {
		slog.Error("failed to ensure messenger indexes", "error", err)
	}
	messengerRepo := messengermod.NewCombinedRepository(messengerEntRepo, messengerMongoRepo)
	presenceSvc := messengermod.NewPresenceService(d.rdb)
	unreadSvc := messengermod.NewUnreadService(d.rdb)
	messengerSvc := messengermod.NewMessengerService(messengerRepo, d.nc, presenceSvc, unreadSvc)
	// Feature #58 (Thread → Issue): MessengerService.PromoteToIssue делегирует в trackerSvc.
	messengerSvc.SetIssuePromoter(trackerSvc)
	trackerSvc.SetThreadMirror(&threadMirrorAdapter{svc: messengerSvc})
	messengerSvc.SetIssueThreadLinker(trackerSvc)
	messengerHandler := messengermod.NewHandler(messengerSvc)
	messengerHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.messengerEntRepo = messengerEntRepo
	d.messengerSvc = messengerSvc
	d.presenceSvc = presenceSvc

	// Маршруты TMS (Ent + MongoDB combined repository)
	tmsEntRepo := tmsmod.NewEntRepository(d.db)
	tmsMongoRepo := tmsmod.NewMongoRepository(mongoDB)
	if err := tmsMongoRepo.EnsureIndexes(d.ctx); err != nil {
		slog.Error("failed to ensure TMS indexes", "error", err)
	}
	tmsRepo := tmsmod.NewCombinedRepository(tmsEntRepo, tmsMongoRepo)
	tmsSvc := tmsmod.NewTMSService(tmsRepo, d.nc)
	tmsHandler := tmsmod.NewHandler(tmsSvc)
	tmsHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.tmsRepo = tmsRepo
	d.tmsSvc = tmsSvc

	// Search routes (глобальный поиск + сохранённые фильтры)
	// Поиск переиспользует тот же raw `*sql.DB`, что и /health probe (BE-H5):
	// pool разделяется с Ent, отдельного pgx-подключения не открываем.
	if err := searchmod.EnsureSearchIndexes(d.ctx, d.rawDB); err != nil {
		slog.Error("failed to ensure search indexes", "error", err)
	}
	searchRepo := searchmod.NewEntRepository(d.db)
	searchSvc := searchmod.NewSearchService(searchRepo, d.rawDB, mongoDB)
	searchHandler := searchmod.NewHandler(searchSvc)
	searchHandler.RegisterRoutes(api, requireAuth)
	d.searchSvc = searchSvc
}
