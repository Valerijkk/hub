package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	auditmod "github.com/valerii/hub/backend/internal/audit"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/internal/platform/database"
	"github.com/valerii/hub/backend/internal/platform/health"
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

	// A14-001: /live = liveness (процесс), /ready = deps (readiness).
	// /health — alias /ready для обратной совместимости (не использовать как liveness).
	pingPG := func(ctx context.Context) error { return d.rawDB.PingContext(ctx) }
	pingRedis := func(ctx context.Context) error { return d.rdb.Ping(ctx).Err() }
	pingMongo := func(ctx context.Context) error { return mongoDB.Client().Ping(ctx, nil) }
	pingNATS := func(ctx context.Context) error {
		if d.nc == nil || d.nc.Conn == nil || !d.nc.Conn.IsConnected() {
			return fmt.Errorf("nats disconnected")
		}
		return nil
	}
	readyHandler := health.Ready(health.Deps{
		PingPostgres: pingPG,
		PingRedis:    pingRedis,
		PingMongo:    pingMongo,
		PingNATS:     pingNATS,
	})
	api.Get("/live", health.Live)
	api.Get("/ready", readyHandler)
	api.Get("/health", readyHandler)

	statusSvc := statusmod.NewService(pingPG, pingRedis, pingMongo, pingNATS)
	// W7-A5-002: публичный aggregate + /status/details за auth.
	statusmod.NewHandler(statusSvc).RegisterRoutes(api, requireAuth)

	messengerEntRepo := messengermod.NewEntRepository(d.db)
	messengerMongoRepo := messengermod.NewMongoRepository(mongoDB)
	if err := messengerMongoRepo.EnsureIndexes(d.ctx); err != nil {
		slog.Error("failed to ensure messenger indexes", "error", err)
	}
	// P1-BE-34 / TASK-023: GDPR erase стирает Mongo-сообщения по sender_id.
	// C-W19-002 / TASK-178: post-Commit fail → JetStream compensating retry (идемпотентный DeleteMany).
	if d.auditSvc != nil {
		eraser := auditmod.NewMongoMessageEraser(mongoDB)
		d.auditSvc.SetMessageEraser(eraser)
		d.auditSvc.SetMongoEraseRetry(auditmod.NewJetStreamMongoEraseRetry(d.nc))
		eraseRetryCons := auditmod.NewMongoEraseRetryConsumer(d.nc, eraser)
		if err := eraseRetryCons.Start(d.rootCtx); err != nil {
			slog.Error("failed to start gdpr mongo erase retry consumer", "error", err)
		}
		d.addCleanup(eraseRetryCons.Stop)
	}
	messengerRepo := messengermod.NewCombinedRepository(messengerEntRepo, messengerMongoRepo)
	presenceSvc := messengermod.NewPresenceService(d.rdb)
	unreadSvc := messengermod.NewUnreadService(d.rdb)
	messengerSvc := messengermod.NewMessengerService(messengerRepo, d.nc, presenceSvc, unreadSvc)
	// Feature #58 (Thread → Issue): MessengerService.PromoteToIssue делегирует в trackerSvc.
	messengerSvc.SetIssuePromoter(trackerSvc)
	// TASK-010: обязательный PermTaskCreate в целевом проекте (fail-closed при nil).
	messengerSvc.SetPromotePermissionChecker(rbacSvc)
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
	// Поиск переиспользует тот же raw `*sql.DB`, что и /ready probe (BE-H5):
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
