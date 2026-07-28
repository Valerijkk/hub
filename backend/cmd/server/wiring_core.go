package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	devmod "github.com/valerii/hub/backend/internal/dev"
	"github.com/valerii/hub/backend/internal/platform/cron"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	projectmod "github.com/valerii/hub/backend/internal/project"
	"github.com/valerii/hub/backend/internal/rbac"
	sprintmod "github.com/valerii/hub/backend/internal/sprint"
	storagemod "github.com/valerii/hub/backend/internal/storage"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

func wireCore(d *serverDeps) {
	api := d.api
	requireAuth := d.requireAuth

	// Маршруты project + RBAC (RegisterRoutes project требует rbacSvc).
	rbacCache := rbac.NewPermissionCache(d.rdb)
	outboxPublisher := outbox.NewPublisher()
	projectRepo := projectmod.NewEntRepository(d.db, outboxPublisher)
	projectSvc := projectmod.NewProjectService(projectRepo, rbacCache, d.nc)
	projectHandler := projectmod.NewHandler(projectSvc)
	d.rbacCache = rbacCache
	d.outboxPublisher = outboxPublisher
	d.projectRepo = projectRepo
	d.projectSvc = projectSvc

	// RBAC service (проверка прав с Redis cache)
	rbacRepo := rbac.NewEntRepository(d.db)
	rbacSvc := rbac.NewRBACService(rbacRepo, rbacCache)
	projectHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	// Обработчик кастомных RBAC-ролей — admin CRUD поверх системных ролей.
	customRoleHandler := rbac.NewCustomRoleHandler(rbacSvc)
	customRoleHandler.RegisterRoutes(api, requireAuth)
	d.rbacRepo = rbacRepo
	d.rbacSvc = rbacSvc

	// W7-A5-001: карта эндпоинтов только для system admin (не публичный recon).
	// Fiber v3: middleware после handler в сигнатуре выполняется ДО handler.
	api.Get("/docs", func(c fiber.Ctx) error {
		return c.Type("html").SendString(devmod.BuildAPIDocsHTML(d.app))
	}, requireAuth, rbac.RequireAdmin(rbacRepo))
	api.Get("/docs/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		body, err := devmod.BuildAPIDocsJSON(d.app)
		if err != nil {
			return err
		}
		return c.SendString(body)
	}, requireAuth, rbac.RequireAdmin(rbacRepo))

	// Storage service (SeaweedFS / S3-compatible) — presigned upload/download.
	storageSvc, err := storagemod.NewStorageService(d.cfg.S3, storagemod.NewEntRepository(d.db))
	if err != nil {
		slog.Error("failed to create storage service", "error", err)
		os.Exit(1)
	}
	d.storageSvc = storageSvc

	// Маршруты storage (/api/files/*)
	storageHandler := storagemod.NewHandler(storageSvc)
	storageHandler.RegisterRoutes(api, requireAuth, rbacSvc)

	// Cron: очистка осиротевших файлов (pending > 24h) — ежедневно.
	// ME-10: контекст rootCtx — отменяется одним общим signal.NotifyContext.
	orphanCleanupSvc := storageSvc.NewOrphanCleanupFromService()
	cron.Schedule(d.rootCtx, 24*time.Hour, "orphan-file-cleanup", func(ctx context.Context) error {
		return orphanCleanupSvc.Cleanup(ctx)
	})

	// Маршруты sprint (Scrum: спринты, Planning Poker, burndown, velocity).
	// Создаём до трекера, т.к. SprintService реализует WIPChecker.
	sprintRepo := sprintmod.NewEntRepository(d.db)
	sprintSvc := sprintmod.NewSprintService(sprintRepo, d.rdb, d.nc)
	pokerSvc := sprintmod.NewPokerService(d.rdb, d.nc)
	sprintHandler := sprintmod.NewHandler(sprintSvc, pokerSvc)
	sprintHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.sprintSvc = sprintSvc

	// Маршруты tracker
	trackerRepo := trackermod.NewEntRepository(d.db, outboxPublisher)
	trackerSvc := trackermod.NewTrackerService(trackerRepo, sprintSvc)
	automationSvc := trackermod.NewAutomationService(trackerRepo)
	trackerHandler := trackermod.NewHandler(trackerSvc, automationSvc)
	trackerHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.trackerRepo = trackerRepo
	d.trackerSvc = trackerSvc
	d.automationSvc = automationSvc

	// Автопровизия трекера при создании проекта (workflow + доска "Main Board")
	trackerProv := trackermod.NewTrackerProjectProvisioner(trackerRepo)
	projectSvc.SetTrackerProvisioner(trackerProv)
}
