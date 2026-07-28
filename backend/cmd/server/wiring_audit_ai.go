package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	aimod "github.com/valerii/hub/backend/internal/ai"
	auditmod "github.com/valerii/hub/backend/internal/audit"
	mailmod "github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/platform/cron"
	"github.com/valerii/hub/backend/internal/platform/middleware"
)

func wireAuditMailAI(d *serverDeps) {
	api := d.api
	requireAuth := d.requireAuth
	cfg := d.cfg

	// Audit routes (журнал аудита, лимиты, корзина, управление пользователями)
	// ME-14: audit consumer использует общий rootCtx, плюс defer Stop(),
	// чтобы JetStream consumer корректно завершился и in-flight Ack-и
	// не потерялись после ухода main(). Stop() блокируется до drain'а.
	auditConsumer := auditmod.NewAuditConsumer(d.nc, d.auditRepo)
	if err := auditConsumer.Start(d.rootCtx); err != nil {
		slog.Error("failed to start audit consumer", "error", err)
	}
	d.addCleanup(auditConsumer.Stop)
	d.auditConsumer = auditConsumer
	auditHandler := auditmod.NewHandler(d.auditSvc)
	auditHandler.RegisterRoutesSimple(api, requireAuth)
	auditHandler.RegisterAlertWebhook(api, cfg.Server.AlertWebhookSecret)

	hubMailURL := os.Getenv("HUB_MAIL_URL")
	if hubMailURL == "" {
		hubMailURL = "http://127.0.0.1:3001"
	}
	fromAlerts := os.Getenv("RESEND_FROM_ALERTS")
	if fromAlerts == "" {
		fromAlerts = cfg.Resend.FromEmail
	}
	mailStore := mailmod.NewEntSettingsStore(d.db)
	mailAdminSvc := mailmod.NewAdminService(mailStore, hubMailURL, fromAlerts, cfg.Resend.APIKey != "")
	mailSender := mailmod.NewService(mailmod.LoadConfig())
	mailSender.SetSettingsStore(mailStore) // W7-A1-003: admin daily_limit / alerts_enabled → SendAlert
	mailAdminHandler := mailmod.NewAdminHandler(mailAdminSvc, mailStore, mailSender, func(c fiber.Ctx, userID uuid.UUID) (bool, error) {
		return d.auditSvc.IsSystemAdmin(c.Context(), userID)
	})
	mailAdminHandler.RegisterRoutes(api, requireAuth)

	aiRepo := aimod.NewEntRepository(d.db, d.integrationCipher)
	aiAudit := aimod.NewAuditRepoLogger(d.auditRepo)
	aiAdapter := aimod.NewOpenAIAdapter()
	aiSvc := aimod.NewService(aiRepo, aiAdapter, aiAudit)
	aiSvc.SetProviderRPMLimiter(aimod.NewRedisProviderRPMLimiter(d.rdb))
	aiSvc.SetSettingsStore(aimod.NewEntSettingsStore(d.db))
	aiSvc.SetOfflineOnly(cfg.AI.OfflineOnly)
	aiHandler := aimod.NewHandler(aiSvc)
	aiAdminHandler := aimod.NewAdminHandler(aiSvc, func(c fiber.Ctx, userID uuid.UUID) (bool, error) {
		return d.auditSvc.IsSystemAdmin(c.Context(), userID)
	})
	aiRateLimit := middleware.RateLimitByUser(d.rootCtx, middleware.AIRateLimit)
	aiHandler.RegisterRoutes(api, requireAuth, aiRateLimit)
	// A2W5-A1-002 / TASK-085: AIRateLimit также на admin TestProvider и /ai/my (BYOK).
	aiAdminHandler.RegisterRoutes(api, requireAuth, aiRateLimit)
	aimod.NewUserHandler(aiSvc).RegisterRoutes(api, requireAuth, aiRateLimit)
	d.aiSvc = aiSvc
	d.aiHandler = aiHandler

	botUserID, err := aimod.EnsureBotUser(d.ctx, d.db)
	if err != nil {
		slog.Error("failed to ensure @hub-ai bot user", "error", err)
		os.Exit(1)
	}
	d.botUserID = botUserID
	d.messengerSvc.SetSystemBotID(botUserID)
	aiHandler.SetBotUserID(botUserID)
	aiBot := aimod.NewMessengerBot(aiSvc, d.messengerSvc, botUserID)
	issueLookup := aimod.NewTrackerIssueLookup(d.trackerSvc, d.projectRepo)
	issueLookup.SetAccessChecker(aimod.RBACProjectAccess{Svc: d.rbacSvc})
	aiBot.SetIssueLookup(issueLookup)
	aiBot.SetDocSearch(aimod.NewSearchDocProvider(d.searchSvc))
	d.messengerSvc.SetMessageSentHook(aiBot)
	d.aiBot = aiBot

	// Запуск фоновой очистки корзины (каждые 24 часа).
	// ME-10: rootCtx.
	cron.Schedule(d.rootCtx, 24*time.Hour, "trash-purge-expired", func(ctx context.Context) error {
		return d.auditSvc.PurgeExpired(ctx)
	})

	// Retention policy: ежедневная очистка устаревших данных. ME-10: rootCtx.
	// BE-H7: передаём S3-клиент, чтобы при удалении row из recording таблицы
	// чистились и соответствующие .mp4-объекты в SeaweedFS. Если S3 не
	// сконфигурирован — RetentionService продолжит работать в fallback-режиме
	// (только метаданные).
	var retentionS3Client *s3.Client
	retentionS3Bucket := d.storageSvc.Bucket()
	if cfg.S3.Endpoint != "" {
		retentionS3Client = d.storageSvc.S3Client()
	}
	d.retentionS3Client = retentionS3Client
	d.retentionS3Bucket = retentionS3Bucket
	retentionSvc := auditmod.NewRetentionService(d.db, d.mongoDB, retentionS3Client, retentionS3Bucket)
	// ME-3: гарантируем индекс на created_at в коллекциях, которые подметает
	// retention-loop. Без него DeleteMany по $lt:cutoff делает COLLSCAN.
	if err := retentionSvc.EnsureIndexes(d.ctx); err != nil {
		slog.Error("failed to ensure retention indexes", "error", err)
	}
	cron.Schedule(d.rootCtx, 24*time.Hour, "retention-cleanup", func(ctx context.Context) error {
		return retentionSvc.RunRetentionCleanup(ctx)
	})

	// Крон проверки приближающихся дедлайнов (каждый час). ME-10: rootCtx.
	cron.Schedule(d.rootCtx, 1*time.Hour, "deadline-check", func(ctx context.Context) error {
		return d.notifSvc.CheckDeadlines(ctx)
	})
}
