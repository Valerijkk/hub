package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/valerii/hub/backend/internal/auth"
	apigatewaymod "github.com/valerii/hub/backend/internal/apigateway"
	devmod "github.com/valerii/hub/backend/internal/dev"
	"github.com/valerii/hub/backend/internal/platform/cron"
	"github.com/valerii/hub/backend/internal/platform/middleware"
)

func wireAuth(d *serverDeps) {
	cfg := d.cfg
	api := d.api

	// --- Сервисы ---

	// Resend delivers magic links when configured. Without it delivery is disabled;
	// only cfg.Server.DevMode controls whether an HTTP response may disclose the code.
	var resendClient auth.ResendClient
	if cfg.Resend.APIKey != "" {
		resendClient = auth.NewResendEmailClient(cfg.Resend.APIKey, cfg.Resend.FromEmail)
		slog.Info("resend email client initialized")
	}
	magicLinkSvc := auth.NewMagicLinkService(d.rdb, resendClient, cfg.Resend.FromEmail)

	authSvc := auth.NewAuthService(
		auth.NewEntRepository(d.db), d.rdb, d.nc,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		magicLinkSvc,
		d.rawDB,
	)
	d.authSvc = authSvc

	// C-W19-001 / TASK-177: Deactivate/Ban → Redis invalidate + NATS revoked → WS kill.
	// auditSvc создан в wireInfra до wireAuth — подключаем SessionRevoker здесь.
	if d.auditSvc != nil {
		d.auditSvc.SetSessionRevoker(authSvc.Sessions)
	}

	// ORG-06: purge pending users старше 7 дней.
	cron.Schedule(d.rootCtx, time.Hour, "pending-user-purge", func(ctx context.Context) error {
		n, err := authSvc.Org.PurgeExpiredPending(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			slog.Info("purged expired pending users", "count", n)
		}
		return nil
	})

	// Auth routes (HI-10: Redis-based sliding-window limiter, не зависит
	// от количества инстансов и устойчив к простой IP-ротации).
	// В dev-режиме пропускаем строгий auth rate limit, чтобы не блокировать разработку.
	if !cfg.Server.DevMode {
		// Отдельный, более жёсткий лимит на verify magic-link: ключ — email
		// из body (атакуемая сущность); 10 попыток / 10 минут. Регистрируем
		// до общего /auth-лимитера, чтобы выполнялся первым на этом пути.
		api.Use("/auth/magic-link/verify",
			middleware.RateLimitRedis(d.rootCtx, d.rdb, middleware.RedisRateLimitConfig{
				Max:       10,
				Window:    10 * time.Minute,
				KeyPrefix: "rl:auth-verify-email:",
			}, middleware.ExtractEmailFromBody("email")),
		)
		// Общий лимит auth-эндпоинтов (3 / 10 минут) по IP-хешу.
		api.Use("/auth", middleware.RateLimitRedis(d.rootCtx, d.rdb, middleware.RedisRateLimitConfig{
			Max:       middleware.AuthRateLimit.Max,
			Window:    middleware.AuthRateLimit.Window,
			KeyPrefix: "rl:auth:",
		}, nil))
	}
	requireAuth := auth.RequireAuth(authSvc)
	d.requireAuth = requireAuth

	// SEC-01: APIKeyAuth должен быть зарегистрирован ДО первого RegisterRoutes.
	apigatewayRepo := apigatewaymod.NewEntRepository(d.db, d.integrationCipher)
	d.apigatewayRepo = apigatewayRepo
	apiKeySvc := apigatewaymod.NewAPIKeyService(apigatewayRepo)
	// async-обновлений last_used_at. Без Close() workers + touchJobs
	// канал утекают на shutdown'е — defer закрывает канал и дожидается
	// drain'а worker-pool'а.
	d.addCleanup(func() { apiKeySvc.Close() })
	d.apiKeySvc = apiKeySvc
	api.Use(apigatewaymod.APIKeyAuth(apiKeySvc))

	// AUTH-01: Telegram login убран; notify bot — Phase 15 Wave 3.
	// DEV_MODE is the sole switch for exposing a one-time code in the response.
	// Missing Resend credentials disable delivery but never enable disclosure.
	authHandler := auth.NewHandler(authSvc, cfg.Server.DevMode)
	authHandler.RegisterRoutes(api, requireAuth)

	// Dev routes (seed, login без Magic Link — только при DEV_MODE=true)
	devHandler := devmod.NewHandler(d.db, authSvc, nil)
	devHandler.RegisterRoutes(api, cfg.Server.DevMode)
}
