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
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/platform/database"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	projectmod "github.com/valerii/hub/backend/internal/project"
	"github.com/valerii/hub/backend/internal/rbac"
	storagemod "github.com/valerii/hub/backend/internal/storage"
	callsmod "github.com/valerii/hub/backend/internal/calls"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	tmsmod "github.com/valerii/hub/backend/internal/tms"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
	redisclient "github.com/valerii/hub/backend/internal/platform/redis"
	auditmod "github.com/valerii/hub/backend/internal/audit"
	notifmod "github.com/valerii/hub/backend/internal/notification"
	dashboardmod "github.com/valerii/hub/backend/internal/dashboard"
	searchmod "github.com/valerii/hub/backend/internal/search"
	sprintmod "github.com/valerii/hub/backend/internal/sprint"
	wsmod "github.com/valerii/hub/backend/internal/ws"
	githubmod "github.com/valerii/hub/backend/internal/github"
	apigatewaymod "github.com/valerii/hub/backend/internal/apigateway"
	backupmod "github.com/valerii/hub/backend/internal/backup"
	devmod "github.com/valerii/hub/backend/internal/dev"
	imexmod "github.com/valerii/hub/backend/internal/imex"

	// fiberswagger "github.com/gofiber/swagger" // TODO: несовместим с Fiber v3
	_ "github.com/valerii/hub/backend/docs"
)

// @title Hub Platform API
// @version 1.0
// @description Единая платформа для команд разработки и тестирования: трекер задач, мессенджер, TMS, созвоны.
// @host localhost:3000
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name X-API-Key
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
	// Глобальный rate limit: 100 req/sec на IP
	app.Use(middleware.RateLimit(middleware.APIRateLimit))

	api := app.Group("/api")

	// Rate limit для загрузки файлов: 10/мин на пользователя
	api.Use("/files", middleware.RateLimitByUser(middleware.FileUploadRateLimit))
	// Rate limit для отправки сообщений: 60/мин на пользователя
	api.Use("/messenger", middleware.RateLimitByUser(middleware.MessageSendRateLimit))
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Swagger UI — документация API
	// TODO: gofiber/swagger несовместим с Fiber v3, нужен swagger middleware для v3
	// Пока отдаём docs.json напрямую
	api.Get("/docs/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(`{"openapi":"3.0.0","info":{"title":"Hub API","version":"1.0.0"},"paths":{}}`)
	})

	// Auth routes
	requireAuth := auth.RequireAuth(authSvc)
	authHandler := auth.NewHandler(authSvc, db)
	authHandler.RegisterRoutes(api, requireAuth)

	// Dev routes (seed, login без Magic Link — только при DEV_MODE=true)
	devHandler := devmod.NewHandler(db, authSvc)
	devHandler.RegisterRoutes(api, cfg.Server.DevMode)

	// Project routes
	rbacCache := rbac.NewPermissionCache(rdb)
	outboxPublisher := outbox.NewPublisher()
	projectRepo := projectmod.NewEntRepository(db, outboxPublisher)
	projectSvc := projectmod.NewProjectService(projectRepo, rbacCache)
	projectSvc.SetDB(db)
	projectSvc.SetNATS(nc)
	projectHandler := projectmod.NewHandler(projectSvc)
	projectHandler.RegisterRoutes(api, requireAuth)

	// RBAC service (permission checks with Redis cache)
	rbacSvc := rbac.NewRBACService(db, rbacCache)
	// Custom RBAC roles handler
	customRoleHandler := rbac.NewCustomRoleHandler(rbacSvc)
	customRoleHandler.RegisterRoutes(api, requireAuth)

	// Storage service (SeaweedFS / S3-compatible)
	storageSvc, err := storagemod.NewStorageService(cfg.S3, db)
	if err != nil {
		slog.Error("failed to create storage service", "error", err)
		os.Exit(1)
	}

	// Storage routes (/api/files/*)
	storageHandler := storagemod.NewHandler(storageSvc)
	storageHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Cron: очистка осиротевших файлов (pending > 24h) — ежедневно
	orphanCleanupSvc := storageSvc.NewOrphanCleanupFromService()
	orphanCtx, orphanCancel := context.WithCancel(ctx)
	defer orphanCancel()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := orphanCleanupSvc.Cleanup(orphanCtx); err != nil {
					slog.Error("orphan file cleanup failed", "error", err)
				}
			case <-orphanCtx.Done():
				return
			}
		}
	}()

	// Tracker routes
	trackerRepo := trackermod.NewEntRepository(db, outboxPublisher)
	trackerSvc := trackermod.NewTrackerService(trackerRepo)
	trackerHandler := trackermod.NewHandler(trackerSvc)
	trackerHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Автопровизия трекера при создании проекта (workflow + доска "Main Board")
	trackerProv := trackermod.NewTrackerProjectProvisioner(trackerRepo)
	projectSvc.SetTrackerProvisioner(trackerProv)

	// Sprint routes (Scrum: спринты, Planning Poker, burndown, velocity)
	sprintRepo := sprintmod.NewEntRepository(db)
	sprintSvc := sprintmod.NewSprintService(sprintRepo, rdb, nc)
	pokerSvc := sprintmod.NewPokerService(rdb, nc)
	sprintHandler := sprintmod.NewHandler(sprintSvc, pokerSvc)
	sprintHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	// Внедряем WIP-лимиты в трекер
	trackerSvc.SetWIPChecker(sprintSvc)

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
	messengerSvc.SetDB(db)
	messengerHandler := messengermod.NewHandler(messengerSvc)
	messengerHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// TMS routes (Ent + MongoDB combined repository)
	tmsEntRepo := tmsmod.NewEntRepository(db)
	tmsMongoRepo := tmsmod.NewMongoRepository(mongoDB)
	if err := tmsMongoRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure TMS indexes", "error", err)
	}
	tmsRepo := tmsmod.NewCombinedRepository(tmsEntRepo, tmsMongoRepo)
	tmsSvc := tmsmod.NewTMSService(tmsRepo, nc)
	tmsHandler := tmsmod.NewHandler(tmsSvc)
	tmsHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Search routes (глобальный поиск + сохранённые фильтры)
	rawDB, err := database.OpenRawDB(cfg.Postgres.URL)
	if err != nil {
		slog.Error("failed to open raw DB for search", "error", err)
		os.Exit(1)
	}
	defer rawDB.Close()
	if err := searchmod.EnsureSearchIndexes(ctx, rawDB); err != nil {
		slog.Error("failed to ensure search indexes", "error", err)
	}
	searchSvc := searchmod.NewSearchService(db, rawDB, mongoDB)
	searchHandler := searchmod.NewHandler(searchSvc)
	searchHandler.RegisterRoutes(api, requireAuth)

	// Notification routes (уведомления: in-app + Telegram)
	notifRepo := notifmod.NewRepository(db)
	telegramSender := notifmod.NewTelegramSender(cfg.Telegram.BotToken)
	notifSvc := notifmod.NewNotificationService(notifRepo, nc, telegramSender, db)
	notifConsumer := notifmod.NewConsumer(nc, notifRepo, notifSvc)
	notifCtx, notifCancel := context.WithCancel(ctx)
	defer notifCancel()
	if err := notifConsumer.Start(notifCtx); err != nil {
		slog.Error("failed to start notification consumer", "error", err)
	}
	if telegramSender != nil {
		telegramSender.Start(notifCtx)
	}
	notifHandler := notifmod.NewHandler(notifSvc)
	notifHandler.RegisterRoutes(api, requireAuth)

	// Audit routes (журнал аудита, лимиты, корзина, управление пользователями)
	auditRepo := auditmod.NewRepository(db)
	auditSvc := auditmod.NewAuditService(db, auditRepo)
	auditConsumer := auditmod.NewAuditConsumer(nc, auditRepo)
	auditCtx, auditCancel := context.WithCancel(ctx)
	defer auditCancel()
	if err := auditConsumer.Start(auditCtx); err != nil {
		slog.Error("failed to start audit consumer", "error", err)
	}
	auditHandler := auditmod.NewHandler(auditSvc)
	auditHandler.RegisterRoutesSimple(api, requireAuth)

	// Запуск фоновой очистки корзины (каждые 24 часа)
	purgeCtx, purgeCancel := context.WithCancel(ctx)
	defer purgeCancel()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := auditSvc.PurgeExpired(purgeCtx); err != nil {
					slog.Error("purge expired items failed", "error", err)
				}
			case <-purgeCtx.Done():
				return
			}
		}
	}()

	// Retention policy: ежедневная очистка устаревших данных
	retentionSvc := auditmod.NewRetentionService(db, mongoDB)
	retentionCtx, retentionCancel := context.WithCancel(ctx)
	defer retentionCancel()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := retentionSvc.RunRetentionCleanup(retentionCtx); err != nil {
					slog.Error("retention cleanup failed", "error", err)
				}
			case <-retentionCtx.Done():
				return
			}
		}
	}()

	// Dashboard routes (дашборды: персональный, проектный, админский)
	dashSvc := dashboardmod.NewDashboardService(db, rdb, mongoDB)
	dashHandler := dashboardmod.NewHandler(dashSvc)
	dashHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// GitHub integration routes (webhook, PR/commit linking, branch creation, link preview)
	githubRepo := githubmod.NewEntRepository(db)
	githubSvc := githubmod.NewGitHubService(githubRepo, nc)
	linkPreviewSvc := githubmod.NewLinkPreviewService(githubRepo, rdb)
	githubHandler := githubmod.NewHandler(githubSvc, linkPreviewSvc)
	githubHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Calls routes (голосовые звонки через LiveKit)
	callsRepo := callsmod.NewEntRepository(db)
	callsChatRepo := callsmod.NewChatRepository(mongoDB)
	if err := callsChatRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure calls chat indexes", "error", err)
	}
	callsSvc := callsmod.NewCallService(cfg.LiveKit, callsRepo, nc, callsChatRepo)
	recordingSvc := callsmod.NewRecordingService(cfg.LiveKit, cfg.S3, callsRepo, nc)
	scheduledCallSvc := callsmod.NewScheduledCallService(callsRepo, nc)
	callsHandler := callsmod.NewHandler(callsSvc, recordingSvc, scheduledCallSvc, cfg.LiveKit)
	callsHandler.RegisterRoutes(api, requireAuth)

	// Cron: обработка запланированных звонков (каждые 60 секунд)
	scheduledCtx, scheduledCancel := context.WithCancel(ctx)
	defer scheduledCancel()
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := scheduledCallSvc.ProcessDueScheduledCalls(scheduledCtx); err != nil {
					slog.Error("scheduled calls processing failed", "error", err)
				}
			case <-scheduledCtx.Done():
				return
			}
		}
	}()

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

	// API Gateway (API keys + Webhooks)
	apiKeySvc := apigatewaymod.NewAPIKeyService(db)
	webhookSvc := apigatewaymod.NewWebhookService(db)
	webhookDispatcher := apigatewaymod.NewWebhookDispatcher(db, nc)
	dispatcherCtx, dispatcherCancel := context.WithCancel(ctx)
	defer dispatcherCancel()
	if err := webhookDispatcher.Start(dispatcherCtx); err != nil {
		slog.Error("failed to start webhook dispatcher", "error", err)
	}
	apiGatewayHandler := apigatewaymod.NewHandler(apiKeySvc, webhookSvc)
	apiGatewayHandler.RegisterRoutes(api, requireAuth)

	// API key auth middleware (параллельно с JWT)
	app.Use(apigatewaymod.APIKeyAuth(apiKeySvc))

	// Backup module (автоматическое еженедельное резервное копирование)
	backupSvc := backupmod.NewBackupService(db, nil, "hub-backups", cfg.Postgres.URL, cfg.MongoDB.URL)
	backupHandler := backupmod.NewHandler(backupSvc)
	backupHandler.RegisterRoutes(api, requireAuth)

	// Еженедельный автоматический бэкап (каждые 7 дней)
	backupCtx, backupCancel := context.WithCancel(ctx)
	defer backupCancel()
	go func() {
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Info("запуск автоматического еженедельного бэкапа")
				if _, err := backupSvc.PerformBackup(backupCtx); err != nil {
					slog.Error("автоматический бэкап не удался", "error", err)
				}
			case <-backupCtx.Done():
				return
			}
		}
	}()

	// Import/Export module
	imexSvc := imexmod.NewImExService(db)
	imexHandler := imexmod.NewHandler(imexSvc)
	imexHandler.RegisterRoutes(api, requireAuth)

	// Project Templates
	templateSvc := projectmod.NewTemplateService(db)
	if err := templateSvc.SeedSystemTemplates(ctx); err != nil {
		slog.Error("failed to seed system templates", "error", err)
	}
	templateHandler := projectmod.NewTemplateHandler(templateSvc)
	templateHandler.RegisterRoutes(api, requireAuth)

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
