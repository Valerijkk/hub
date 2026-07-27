package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	apigatewaymod "github.com/valerii/hub/backend/internal/apigateway"
	backupmod "github.com/valerii/hub/backend/internal/backup"
	callsmod "github.com/valerii/hub/backend/internal/calls"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
	imexmod "github.com/valerii/hub/backend/internal/imex"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	projectmod "github.com/valerii/hub/backend/internal/project"
	"github.com/valerii/hub/backend/internal/platform/cron"
	"github.com/valerii/hub/backend/internal/rbac"
	aimod "github.com/valerii/hub/backend/internal/ai"
)

func wireIntegrations(d *serverDeps) {
	api := d.api
	requireAuth := d.requireAuth
	rbacSvc := d.rbacSvc
	cfg := d.cfg

	// Маршруты GitHub (webhook, PR/commit, branch, link preview)
	githubRepo := githubmod.NewEntRepository(d.db, d.integrationCipher)
	githubSvc := githubmod.NewGitHubService(githubRepo, d.nc)
	linkPreviewSvc := githubmod.NewLinkPreviewService(githubRepo, d.rdb)
	githubHandler := githubmod.NewHandler(githubSvc, linkPreviewSvc)
	githubHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.githubSvc = githubSvc

	// Маршруты GitLab (webhook, MR/commit, pipelines, branch)
	gitlabRepo := gitlabmod.NewEntRepository(d.db, d.integrationCipher)
	gitlabSvc := gitlabmod.NewGitLabService(gitlabRepo, d.nc)
	gitlabHandler := gitlabmod.NewHandler(gitlabSvc)
	gitlabHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.gitlabSvc = gitlabSvc
	d.aiBot.SetCodeReview(aimod.NewUnifiedCodeReviewProvider(githubSvc, gitlabSvc))

	// Calls routes (голосовые звонки через LiveKit)
	callsRepo := callsmod.NewEntRepository(d.db)
	callsChatRepo := callsmod.NewChatRepository(d.mongoDB)
	if err := callsChatRepo.EnsureIndexes(d.ctx); err != nil {
		slog.Error("failed to ensure calls chat indexes", "error", err)
	}
	callsSvc := callsmod.NewCallService(cfg.LiveKit, callsRepo, d.nc, callsChatRepo)
	recordingSvc := callsmod.NewRecordingService(cfg.LiveKit, cfg.S3, callsRepo, d.nc)
	scheduledCallSvc := callsmod.NewScheduledCallService(callsRepo, d.nc)
	callsHandler := callsmod.NewHandler(callsSvc, recordingSvc, scheduledCallSvc, cfg.LiveKit, rbacSvc)
	callsHandler.RegisterRoutes(api, requireAuth)
	d.callsRepo = callsRepo
	d.callsSvc = callsSvc
	d.scheduledCallSvc = scheduledCallSvc

	// Cron: обработка запланированных звонков (каждые 60 секунд). ME-10: rootCtx.
	cron.Schedule(d.rootCtx, 60*time.Second, "scheduled-calls", func(ctx context.Context) error {
		return scheduledCallSvc.ProcessDueScheduledCalls(ctx)
	})

	// Запуск фоновой очистки устаревших записей присутствия. ME-10: rootCtx.
	go d.presenceSvc.StartCleanupLoop(d.rootCtx)

	// Auto-channel service (создаёт каналы при создании проектов)
	autoChannelSvc := messengermod.NewAutoChannelService(d.messengerSvc, d.messengerEntRepo)
	if err := autoChannelSvc.Start(d.rootCtx, d.nc); err != nil {
		slog.Error("failed to start auto-channel service", "error", err)
		os.Exit(1)
	}
	d.addCleanup(autoChannelSvc.Stop)
	d.autoChannelSvc = autoChannelSvc

	// API Gateway (API keys + Webhooks). apiKeySvc уже создан выше + middleware
	// зарегистрирован на api-группу (см. SEC-01 в wireAuth).
	webhookSvc := apigatewaymod.NewWebhookService(d.apigatewayRepo)
	webhookDispatcher := apigatewaymod.NewWebhookDispatcher(d.apigatewayRepo, d.nc)
	if err := webhookDispatcher.Start(d.rootCtx); err != nil {
		slog.Error("failed to start webhook dispatcher", "error", err)
	}
	// ME-10: явный Stop() блокирует до drain'а workerWG (см. HI-08).
	d.addCleanup(webhookDispatcher.Stop)
	d.webhookDispatcher = webhookDispatcher
	apiGatewayHandler := apigatewaymod.NewHandler(d.apiKeySvc, webhookSvc)
	apiGatewayHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Backup module (автоматическое еженедельное резервное копирование).
	// INF-04: используем реальный S3-клиент из StorageService — без этого
	// PerformBackup пишет ZIP в no-op (s3Client == nil) и weekly ticker
	// крутится впустую. Bucket — основной S3 bucket приложения, объекты
	// складываются по префиксу "backups/" (см. PerformBackup → s3Key).
	// Если S3 не настроен (cfg.S3.Endpoint == ""), оставляем nil явно
	// и громко логируем — fail-soft, но без молчаливой потери данных.
	var backupS3Client *s3.Client
	backupS3Bucket := d.storageSvc.Bucket()
	if cfg.S3.Endpoint == "" {
		slog.Warn("backup: S3 not configured (S3_ENDPOINT пустой), еженедельный бэкап будет no-op")
	} else {
		backupS3Client = d.storageSvc.S3Client()
	}
	d.backupS3Client = backupS3Client
	d.backupS3Bucket = backupS3Bucket
	backupSvc := backupmod.NewBackupServiceWithGate(d.rootCtx, backupmod.NewEntRepository(d.db), backupS3Client, backupS3Bucket, cfg.Postgres.URL, cfg.MongoDB.URL, d.maintGate)
	backupHandler := backupmod.NewHandler(backupSvc)
	backupHandler.RegisterRoutes(api, requireAuth, rbac.RequireAdmin(d.rbacRepo))
	d.backupSvc = backupSvc

	// Еженедельный автоматический бэкап (каждые 7 дней). ME-10: rootCtx.
	cron.Schedule(d.rootCtx, 7*24*time.Hour, "weekly-backup", func(ctx context.Context) error {
		slog.Info("запуск автоматического еженедельного бэкапа")
		_, err := backupSvc.PerformBackup(ctx)
		return err
	})

	// Модуль import/export
	imexSvc := imexmod.NewImExService(imexmod.NewEntRepository(d.db))
	imexHandler := imexmod.NewHandler(imexSvc)
	imexHandler.RegisterRoutes(api, requireAuth)

	// Шаблоны проектов — быстрый старт tracker/TMS из preset.
	templateSvc := projectmod.NewTemplateService(d.projectRepo)
	if err := templateSvc.SeedSystemTemplates(d.ctx); err != nil {
		slog.Error("failed to seed system templates", "error", err)
	}
	templateHandler := projectmod.NewTemplateHandler(templateSvc)
	templateHandler.RegisterRoutes(api, requireAuth)
}
