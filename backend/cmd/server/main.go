package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	fiberrecover "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/platform/config"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
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
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
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

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

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
		// Reached only when DevMode=true (config.validate enforces this).
		slog.Warn("INTEGRATION_ENCRYPTION_KEY not set — using deterministic dev key (DEV ONLY)")
		integrationCipher, err = cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
		if err != nil {
			slog.Error("failed to build dev integration cipher", "error", err)
			os.Exit(1)
		}
	}

	// ME-10 + ME-14: единый корневой контекст, прерываемый SIGINT/SIGTERM.
	// Все долгоиграющие горутины (outbox, cron-тикеры, notification/audit
	// консьюмеры, WS hub, webhook dispatcher и т. п.) получают этот ctx —
	// при первом сигнале они все одновременно идут в graceful-shutdown.
	rootCtx, rootCancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer rootCancel()
	// ctx сохраняем как алиас для read-only бутстрапных вызовов
	// (database connect, EnsureStreams, EnsureIndexes), которые
	// концептуально не "running forever" — но и здесь использование
	// rootCtx безопасно: если процесс получает сигнал раньше окончания
	// инициализации, бутстрап прервётся, и мы быстрее уйдём в exit.
	ctx := rootCtx

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

	// Создаём единый JetStream stream EVENTS для всех модулей.
	if err := nc.EnsureStreams(ctx); err != nil {
		slog.Error("failed to ensure JetStream streams", "error", err)
		os.Exit(1)
	}

	// Outbox worker (ctx == rootCtx — отменится по SIGINT/SIGTERM)
	outboxWorker := outbox.NewWorker(db, nc)
	go outboxWorker.ProcessLoop(rootCtx)

	// --- Services ---

	// Magic link: при наличии API-ключа Resend — отправка email; иначе — dev-режим (код в логах).
	var resendClient auth.ResendClient
	if cfg.Resend.APIKey != "" {
		resendClient = auth.NewResendEmailClient(cfg.Resend.APIKey, cfg.Resend.FromEmail)
		slog.Info("resend email client initialized")
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
		// ME-15: глобальный лимит на размер тела запроса.
		// 25 МБ покрывает CSV-импорт задач и разумный аватар; маршруты,
		// которым нужно больше (например, заливка больших файлов), могут
		// переопределить лимит локально.
		BodyLimit: 25 * 1024 * 1024,
	})

	app.Use(middleware.RequestID())
	// HI-09: глобальный recover, чтобы паника в обработчике не валила процесс.
	// Стоит сразу после RequestID, чтобы паническая ошибка уже несла request_id
	// в лог — но раньше всех бизнес-middleware (логгер, CORS, rate-limit).
	app.Use(fiberrecover.New(fiberrecover.Config{EnableStackTrace: true}))
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	// Глобальный rate limit: 100 req/sec на IP
	app.Use(middleware.RateLimit(middleware.APIRateLimit))

	api := app.Group("/api")

	// Rate limit для загрузки файлов: 10/мин на пользователя
	api.Use("/files", middleware.RateLimitByUser(middleware.FileUploadRateLimit))
	// Rate limit для отправки сообщений: 60/мин на пользователя
	api.Use("/messenger", middleware.RateLimitByUser(middleware.MessageSendRateLimit))
	// Health endpoint registered after all infrastructure clients are initialized (see below)

	// API-документация: динамическая страница со списком всех эндпоинтов.
	// gofiber/swagger несовместим с Fiber v3, поэтому генерируем HTML из зарегистрированных маршрутов.
	api.Get("/docs", func(c fiber.Ctx) error {
		return c.Type("html").SendString(devmod.BuildAPIDocsHTML(app))
	})
	api.Get("/docs/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(devmod.BuildAPIDocsJSON(app))
	})

	// Auth routes (HI-10: Redis-based sliding-window limiter, не зависит
	// от количества инстансов и устойчив к простой IP-ротации).
	// В dev-режиме пропускаем строгий auth rate limit, чтобы не блокировать разработку.
	if !cfg.Server.DevMode {
		// Отдельный, более жёсткий лимит на verify magic-link: ключ — email
		// из body (атакуемая сущность); 10 попыток / 10 минут. Регистрируем
		// до общего /auth-лимитера, чтобы выполнялся первым на этом пути.
		api.Use("/auth/magic-link/verify",
			middleware.RateLimitRedis(rdb, middleware.RedisRateLimitConfig{
				Max:       10,
				Window:    10 * time.Minute,
				KeyPrefix: "rl:auth-verify-email:",
			}, middleware.ExtractEmailFromBody("email")),
		)
		// Общий лимит auth-эндпоинтов (3 / 10 минут) по IP-хешу.
		api.Use("/auth", middleware.RateLimitRedis(rdb, middleware.RedisRateLimitConfig{
			Max:       middleware.AuthRateLimit.Max,
			Window:    middleware.AuthRateLimit.Window,
			KeyPrefix: "rl:auth:",
		}, nil))
	}
	requireAuth := auth.RequireAuth(authSvc)
	tgBotAuth := auth.NewTelegramBotAuth(cfg.Telegram.BotToken, rdb)
	if tgBotAuth != nil {
		tgBotAuth.StartPolling(ctx)
	}

	// SEC-01: APIKeyAuth должен быть зарегистрирован ДО первого RegisterRoutes —
	// в Fiber `Use` влияет только на handlers, замаунченные ПОЗЖЕ.
	// Раньше Use стоял после всех роутов и был мёртвым кодом.
	apiKeySvc := apigatewaymod.NewAPIKeyService(db)
	api.Use(apigatewaymod.APIKeyAuth(apiKeySvc))

	authHandler := auth.NewHandler(authSvc, tgBotAuth)
	authHandler.RegisterRoutes(api, requireAuth)

	// Dev routes (seed, login без Magic Link — только при DEV_MODE=true)
	devHandler := devmod.NewHandler(db, authSvc)
	devHandler.RegisterRoutes(api, cfg.Server.DevMode)

	// Project routes
	rbacCache := rbac.NewPermissionCache(rdb)
	outboxPublisher := outbox.NewPublisher()
	projectRepo := projectmod.NewEntRepository(db, outboxPublisher)
	projectSvc := projectmod.NewProjectService(projectRepo, rbacCache, db, nc)
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

	// Cron: очистка осиротевших файлов (pending > 24h) — ежедневно.
	// ME-10: контекст rootCtx — отменяется одним общим signal.NotifyContext.
	orphanCleanupSvc := storageSvc.NewOrphanCleanupFromService()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := orphanCleanupSvc.Cleanup(rootCtx); err != nil {
					slog.Error("orphan file cleanup failed", "error", err)
				}
			case <-rootCtx.Done():
				return
			}
		}
	}()

	// Sprint routes (Scrum: спринты, Planning Poker, burndown, velocity)
	// Создаём до трекера, т.к. SprintService реализует WIPChecker.
	sprintRepo := sprintmod.NewEntRepository(db)
	sprintSvc := sprintmod.NewSprintService(sprintRepo, rdb, nc)
	pokerSvc := sprintmod.NewPokerService(rdb, nc)
	sprintHandler := sprintmod.NewHandler(sprintSvc, pokerSvc)
	sprintHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Tracker routes
	trackerRepo := trackermod.NewEntRepository(db, outboxPublisher)
	trackerSvc := trackermod.NewTrackerService(trackerRepo, sprintSvc)
	trackerHandler := trackermod.NewHandler(trackerSvc)
	trackerHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Автопровизия трекера при создании проекта (workflow + доска "Main Board")
	trackerProv := trackermod.NewTrackerProjectProvisioner(trackerRepo)
	projectSvc.SetTrackerProvisioner(trackerProv)

	// Messenger routes
	mongoDB, err := database.NewMongoClient(ctx, cfg.MongoDB.URL, cfg.MongoDB.Database)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	// Health endpoint: checks PostgreSQL, Redis, MongoDB connectivity (unauthenticated)
	api.Get("/health", func(c fiber.Ctx) error {
		healthCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		services := fiber.Map{}
		overall := "ok"

		// PostgreSQL
		rawPgDB, pgErr := database.OpenRawDB(cfg.Postgres.URL)
		if pgErr != nil {
			services["postgres"] = "error"
			overall = "degraded"
		} else {
			if pingErr := rawPgDB.PingContext(healthCtx); pingErr != nil {
				services["postgres"] = "error"
				overall = "degraded"
			} else {
				services["postgres"] = "ok"
			}
			rawPgDB.Close()
		}

		// Redis
		if pingErr := rdb.Ping(healthCtx).Err(); pingErr != nil {
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
	searchRepo := searchmod.NewEntRepository(db)
	searchSvc := searchmod.NewSearchService(searchRepo, rawDB, mongoDB)
	searchHandler := searchmod.NewHandler(searchSvc)
	searchHandler.RegisterRoutes(api, requireAuth)

	// Notification routes (уведомления: in-app + Telegram + Web Push)
	notifRepo := notifmod.NewRepository(db)
	telegramSender := notifmod.NewTelegramSender(cfg.Telegram.BotToken)
	pushSender := notifmod.NewPushSender(db, cfg.WebPush)
	notifSvc := notifmod.NewNotificationService(notifRepo, nc, telegramSender, db, pushSender)
	notifConsumer := notifmod.NewConsumer(nc, notifRepo, notifSvc)
	if err := notifConsumer.Start(rootCtx); err != nil {
		slog.Error("failed to start notification consumer", "error", err)
	}
	if telegramSender != nil {
		telegramSender.Start(rootCtx)
	}
	if pushSender != nil {
		pushSender.Start(rootCtx)
	}
	notifHandler := notifmod.NewHandler(notifSvc)
	notifHandler.RegisterRoutes(api, requireAuth)

	// Audit routes (журнал аудита, лимиты, корзина, управление пользователями)
	auditRepo := auditmod.NewRepository(db)
	auditSvc := auditmod.NewAuditService(auditRepo, rdb)
	// ME-14: audit consumer использует общий rootCtx, плюс defer Stop(),
	// чтобы JetStream consumer корректно завершился и in-flight Ack-и
	// не потерялись после ухода main(). Stop() блокируется до drain'а.
	auditConsumer := auditmod.NewAuditConsumer(nc, auditRepo)
	if err := auditConsumer.Start(rootCtx); err != nil {
		slog.Error("failed to start audit consumer", "error", err)
	}
	defer auditConsumer.Stop()
	auditHandler := auditmod.NewHandler(auditSvc)
	auditHandler.RegisterRoutesSimple(api, requireAuth)

	// Запуск фоновой очистки корзины (каждые 24 часа).
	// ME-10: rootCtx.
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := auditSvc.PurgeExpired(rootCtx); err != nil {
					slog.Error("purge expired items failed", "error", err)
				}
			case <-rootCtx.Done():
				return
			}
		}
	}()

	// Retention policy: ежедневная очистка устаревших данных. ME-10: rootCtx.
	retentionSvc := auditmod.NewRetentionService(db, mongoDB)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := retentionSvc.RunRetentionCleanup(rootCtx); err != nil {
					slog.Error("retention cleanup failed", "error", err)
				}
			case <-rootCtx.Done():
				return
			}
		}
	}()

	// Крон проверки приближающихся дедлайнов (каждый час). ME-10: rootCtx.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := notifSvc.CheckDeadlines(rootCtx); err != nil {
					slog.Error("deadline check failed", "error", err)
				}
			case <-rootCtx.Done():
				return
			}
		}
	}()

	// Dashboard routes (дашборды: персональный, проектный, админский)
	dashSvc := dashboardmod.NewDashboardService(db, rdb, mongoDB)
	dashHandler := dashboardmod.NewHandler(dashSvc)
	dashHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Автопровизия TMS: создаём папку "Общие тесты" при создании проекта
	tmsProv := tmsmod.NewTMSProjectProvisioner(tmsRepo)
	projectSvc.SetTMSProvisioner(tmsProv)

	// Автопровизия дашборда: применяем шаблон "dev" при создании проекта
	dashProv := dashboardmod.NewDashboardProjectProvisioner(dashSvc)
	projectSvc.SetDashboardProvisioner(dashProv)

	// GitHub integration routes (webhook, PR/commit linking, branch creation, link preview)
	githubRepo := githubmod.NewEntRepository(db, integrationCipher)
	githubSvc := githubmod.NewGitHubService(githubRepo, nc)
	linkPreviewSvc := githubmod.NewLinkPreviewService(githubRepo, rdb)
	githubHandler := githubmod.NewHandler(githubSvc, linkPreviewSvc)
	githubHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// GitLab integration routes (webhook, MR/commit linking, pipelines, branch creation)
	gitlabRepo := gitlabmod.NewEntRepository(db, integrationCipher)
	gitlabSvc := gitlabmod.NewGitLabService(gitlabRepo, nc)
	gitlabHandler := gitlabmod.NewHandler(gitlabSvc)
	gitlabHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Calls routes (голосовые звонки через LiveKit)
	callsRepo := callsmod.NewEntRepository(db)
	callsChatRepo := callsmod.NewChatRepository(mongoDB)
	if err := callsChatRepo.EnsureIndexes(ctx); err != nil {
		slog.Error("failed to ensure calls chat indexes", "error", err)
	}
	callsSvc := callsmod.NewCallService(cfg.LiveKit, callsRepo, nc, callsChatRepo)
	recordingSvc := callsmod.NewRecordingService(cfg.LiveKit, cfg.S3, callsRepo, nc)
	scheduledCallSvc := callsmod.NewScheduledCallService(callsRepo, nc)
	callsHandler := callsmod.NewHandler(callsSvc, recordingSvc, scheduledCallSvc, cfg.LiveKit, rbacSvc)
	callsHandler.RegisterRoutes(api, requireAuth)

	// Cron: обработка запланированных звонков (каждые 60 секунд). ME-10: rootCtx.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := scheduledCallSvc.ProcessDueScheduledCalls(rootCtx); err != nil {
					slog.Error("scheduled calls processing failed", "error", err)
				}
			case <-rootCtx.Done():
				return
			}
		}
	}()

	// Запуск фоновой очистки устаревших записей присутствия. ME-10: rootCtx.
	go presenceSvc.StartCleanupLoop(rootCtx)

	// Auto-channel service (создаёт каналы при создании проектов)
	autoChannelSvc := messengermod.NewAutoChannelService(messengerSvc, messengerEntRepo)
	if err := autoChannelSvc.Start(nc); err != nil {
		slog.Error("failed to start auto-channel service", "error", err)
		os.Exit(1)
	}
	defer autoChannelSvc.Stop()

	// API Gateway (API keys + Webhooks). apiKeySvc уже создан выше + middleware
	// зарегистрирован на api-группу (см. SEC-01 выше).
	webhookSvc := apigatewaymod.NewWebhookService(db)
	webhookDispatcher := apigatewaymod.NewWebhookDispatcher(db, nc)
	if err := webhookDispatcher.Start(rootCtx); err != nil {
		slog.Error("failed to start webhook dispatcher", "error", err)
	}
	// ME-10: явный Stop() блокирует до drain'а workerWG (см. HI-08).
	defer webhookDispatcher.Stop()
	apiGatewayHandler := apigatewaymod.NewHandler(apiKeySvc, webhookSvc)
	apiGatewayHandler.RegisterRoutes(api, requireAuth)

	// Backup module (автоматическое еженедельное резервное копирование).
	// INF-04: используем реальный S3-клиент из StorageService — без этого
	// PerformBackup пишет ZIP в no-op (s3Client == nil) и weekly ticker
	// крутится впустую. Bucket — основной S3 bucket приложения, объекты
	// складываются по префиксу "backups/" (см. PerformBackup → s3Key).
	// Если S3 не настроен (cfg.S3.Endpoint == ""), оставляем nil явно
	// и громко логируем — fail-soft, но без молчаливой потери данных.
	var backupS3Client *s3.Client
	backupS3Bucket := storageSvc.Bucket()
	if cfg.S3.Endpoint == "" {
		slog.Warn("backup: S3 not configured (S3_ENDPOINT пустой), еженедельный бэкап будет no-op")
	} else {
		backupS3Client = storageSvc.S3Client()
	}
	backupSvc := backupmod.NewBackupService(db, backupS3Client, backupS3Bucket, cfg.Postgres.URL, cfg.MongoDB.URL)
	backupHandler := backupmod.NewHandler(backupSvc)
	backupHandler.RegisterRoutes(api, requireAuth, rbac.RequireAdmin(db))

	// Еженедельный автоматический бэкап (каждые 7 дней). ME-10: rootCtx.
	go func() {
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				slog.Info("запуск автоматического еженедельного бэкапа")
				if _, err := backupSvc.PerformBackup(rootCtx); err != nil {
					slog.Error("автоматический бэкап не удался", "error", err)
				}
			case <-rootCtx.Done():
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
	// ME-10: rootCtx (вместо локального wsCtx) — отменяется одним сигналом.
	go wsHub.Run(rootCtx)

	wsBridge := wsmod.NewBridge(wsHub, nc)
	if err := wsBridge.Start(); err != nil {
		slog.Error("failed to start ws bridge", "error", err)
		os.Exit(1)
	}
	defer wsBridge.Stop()

	// SEC-H1: WS Authorizer гарантирует, что subscribe идёт только в
	// разрешённые комнаты (членство в канале/проекте/звонке).
	wsAuthz := wsmod.NewAuthorizer(wsmod.AuthorizerConfig{
		IsChannelMember: func(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
			return messengerEntRepo.IsMember(ctx, channelID, userID)
		},
		IsCallParticipantOrChannelMember: func(ctx context.Context, callID, userID uuid.UUID) (bool, error) {
			// Звонок может быть канальным (есть ChannelID) или instant
			// (по invite-коду). Для канального — членство в канале;
			// для instant — fail-closed, т.к. подписка идёт не по invite.
			call, err := callsRepo.GetCall(ctx, callID)
			if err != nil {
				return false, err
			}
			if call.ChannelID == nil {
				return false, nil
			}
			return callsRepo.IsChannelMember(ctx, *call.ChannelID, userID)
		},
		IsProjectMember: func(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
			_, err := projectRepo.FindMember(ctx, projectID, userID)
			if err != nil {
				if ent.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}
			return true, nil
		},
	})
	app.Get("/ws", wsmod.Handler(wsHub, authSvc, wsAuthz))

	// --- Start server ---

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
