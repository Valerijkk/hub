package main

import (
	"github.com/gofiber/fiber/v3"
	fiberrecover "github.com/gofiber/fiber/v3/middleware/recover"
	auditmod "github.com/valerii/hub/backend/internal/audit"
	backupmod "github.com/valerii/hub/backend/internal/backup"
	devmod "github.com/valerii/hub/backend/internal/dev"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

func wireFiber(d *serverDeps) {
	cfg := d.cfg

	// --- Fiber-приложение ---

	app := fiber.New(fiber.Config{
		ReadTimeout:             cfg.Server.ReadTimeout,
		WriteTimeout:            cfg.Server.WriteTimeout,
		Immutable:               cfg.Server.Immutable,
		ErrorHandler:            middleware.ErrorHandler,
		EnableTrustedProxyCheck: true,
		TrustedProxies: []string{
			"127.0.0.1",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
		},
		// ME-15: глобальный лимит на размер тела запроса.
		// 25 МБ покрывает CSV-импорт задач и разумный аватар; маршруты,
		// которым нужно больше (например, заливка больших файлов), могут
		// переопределить лимит локально.
		BodyLimit: 25 * 1024 * 1024,
	})
	d.app = app

	app.Use(middleware.RequestID())
	// HI-09: глобальный recover, чтобы паника в обработчике не валила процесс.
	// Стоит сразу после RequestID, чтобы паническая ошибка уже несла request_id
	// в лог — но раньше всех бизнес-middleware (логгер, CORS, rate-limit).
	// A14-007: stack traces только в DevMode — иначе шум + утечка путей в Loki.
	app.Use(fiberrecover.New(fiberrecover.Config{EnableStackTrace: cfg.Server.FiberRecoverEnableStackTrace()}))
	app.Use(observability.HTTPTraceMiddleware())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Get("/metrics", observability.MetricsEndpoint(cfg.Server.MetricsAllowedCIDRs))
	app.Use(observability.HTTPMetricsMiddleware())
	// Глобальный rate limit: 100 req/sec на IP.
	// ME-5: rootCtx даёт cleanup-горутине корректный shutdown.
	app.Use(middleware.RateLimit(d.rootCtx, middleware.APIRateLimit))

	api := app.Group("/api")
	// A7-003: maintenance gate до любых бизнес-маршрутов (создаётся здесь,
	// BackupService подключается к тому же gate в wireIntegrations).
	d.maintGate = backupmod.NewMaintenanceGate()
	api.Use(backupmod.MaintenanceMiddleware(d.maintGate))
	// HTTP-аудит: все API-запросы пишутся в журнал (кроме health/docs).
	api.Use(auditmod.HTTPAuditMiddleware(d.auditSvc.RecordHTTPRequest))

	// Rate limit для загрузки файлов: 10/мин на пользователя
	api.Use("/files", middleware.RateLimitByUser(d.rootCtx, middleware.FileUploadRateLimit))
	// Rate limit для отправки сообщений: 60/мин на пользователя
	api.Use("/messenger", middleware.RateLimitByUser(d.rootCtx, middleware.MessageSendRateLimit))

	// API-документация: динамическая страница со списком всех эндпоинтов.
	// gofiber/swagger несовместим с Fiber v3, поэтому генерируем HTML из зарегистрированных маршрутов.
	api.Get("/docs", func(c fiber.Ctx) error {
		return c.Type("html").SendString(devmod.BuildAPIDocsHTML(app))
	})
	api.Get("/docs/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(devmod.BuildAPIDocsJSON(app))
	})

	d.api = api
}
