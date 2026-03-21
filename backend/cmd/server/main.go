package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/platform/database"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	projectmod "github.com/valerii/hub/backend/internal/project"
	"github.com/valerii/hub/backend/internal/rbac"
	storagemod "github.com/valerii/hub/backend/internal/storage"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	tmsmod "github.com/valerii/hub/backend/internal/tms"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
	redisclient "github.com/valerii/hub/backend/internal/platform/redis"
	wsmod "github.com/valerii/hub/backend/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()
	ctx := context.Background()

	// --- Infrastructure clients ---

	// PostgreSQL + Ent
	db, err := database.NewPostgresClient(ctx, cfg.Postgres.URL)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Redis
	rdb, err := redisclient.NewClient(ctx, cfg.Redis.URL)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// NATS
	nc, err := natsclient.NewClient(ctx, cfg.NATS.URL)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()

	// Outbox worker
	outboxWorker := outbox.NewWorker(db, nc)
	outboxCtx, outboxCancel := context.WithCancel(ctx)
	defer outboxCancel()
	go outboxWorker.ProcessLoop(outboxCtx)

	// --- Services ---

	// Magic link (nil ResendClient = dev mode, logs code instead of sending email)
	var resendClient auth.ResendClient
	if cfg.Resend.APIKey != "" {
		// Production Resend client would be created here.
		// For now, leave nil to use dev-mode logging.
		_ = cfg.Resend.APIKey
	}
	magicLinkSvc := auth.NewMagicLinkService(rdb, resendClient, cfg.Resend.FromEmail)

	authSvc := auth.NewAuthService(
		db, rdb, nc,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		magicLinkSvc,
		cfg.Telegram.BotToken,
	)

	// --- Fiber app ---

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		Immutable:    cfg.Server.Immutable,
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	api := app.Group("/api")
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth routes
	requireAuth := auth.RequireAuth(authSvc)
	authHandler := auth.NewHandler(authSvc, db)
	authHandler.RegisterRoutes(api, requireAuth)

	// Project routes
	rbacCache := rbac.NewPermissionCache(rdb)
	outboxPublisher := outbox.NewPublisher()
	projectRepo := projectmod.NewEntRepository(db, outboxPublisher)
	projectSvc := projectmod.NewProjectService(projectRepo, rbacCache)
	projectHandler := projectmod.NewHandler(projectSvc)
	projectHandler.RegisterRoutes(api, requireAuth)

	// RBAC service (permission checks with Redis cache)
	rbacSvc := rbac.NewRBACService(db, rbacCache)

	// Storage service (SeaweedFS / S3-compatible)
	storageSvc, err := storagemod.NewStorageService(cfg.S3, db)
	if err != nil {
		slog.Error("failed to create storage service", "error", err)
		os.Exit(1)
	}

	// Storage routes (/api/files/*)
	storageHandler := storagemod.NewHandler(storageSvc)
	storageHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Tracker routes
	trackerRepo := trackermod.NewEntRepository(db, outboxPublisher)
	trackerSvc := trackermod.NewTrackerService(trackerRepo)
	trackerHandler := trackermod.NewHandler(trackerSvc)
	trackerHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// TMS routes
	tmsRepo := tmsmod.NewEntRepository(db)
	tmsSvc := tmsmod.NewTMSService(tmsRepo)
	tmsHandler := tmsmod.NewHandler(tmsSvc)
	tmsHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Messenger routes
	mongoDB, err := database.NewMongoClient(ctx, cfg.MongoDB.URL, cfg.MongoDB.Database)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	messengerEntRepo := messengermod.NewEntRepository(db)
	messengerMongoRepo := messengermod.NewMongoRepository(mongoDB)
	if err := messengerMongoRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure messenger indexes", "error", err)
	}
	messengerRepo := messengermod.NewCombinedRepository(messengerEntRepo, messengerMongoRepo)
	presenceSvc := messengermod.NewPresenceService(rdb)
	unreadSvc := messengermod.NewUnreadService(rdb)
	messengerSvc := messengermod.NewMessengerService(messengerRepo, nc, presenceSvc, unreadSvc)
	messengerHandler := messengermod.NewHandler(messengerSvc)
	messengerHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Запуск фоновой очистки устаревших записей присутствия
	presenceCtx, presenceCancel := context.WithCancel(ctx)
	defer presenceCancel()
	go presenceSvc.StartCleanupLoop(presenceCtx)

	// Auto-channel service (создаёт каналы при создании проектов)
	autoChannelSvc := messengermod.NewAutoChannelService(messengerSvc, messengerEntRepo)
	if err := autoChannelSvc.Start(nc); err != nil {
		slog.Error("failed to start auto-channel service", "error", err)
		os.Exit(1)
	}
	defer autoChannelSvc.Stop()

	// --- WebSocket gateway ---

	wsHub := wsmod.NewHub()
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()
	go wsHub.Run(wsCtx)

	wsBridge := wsmod.NewBridge(wsHub, nc)
	if err := wsBridge.Start(); err != nil {
		slog.Error("failed to start ws bridge", "error", err)
		os.Exit(1)
	}
	defer wsBridge.Stop()

	app.Get("/ws", wsmod.Handler(wsHub, authSvc))

	// --- Start server ---

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		slog.Info("starting server", "addr", addr)
		if err := app.Listen(addr); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	outboxCancel()
	if err := app.Shutdown(); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}
