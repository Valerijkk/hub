package main

import (
	"log/slog"

	notifmod "github.com/valerii/hub/backend/internal/notification"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

func wireNotification(d *serverDeps) {
	api := d.api
	requireAuth := d.requireAuth
	cfg := d.cfg

	// Notification routes (уведомления: in-app + Telegram + Web Push)
	notifRepo := notifmod.NewRepository(d.db)
	telegramSender := notifmod.NewTelegramSender(cfg.Telegram.BotToken)
	pushSender := notifmod.NewPushSender(notifRepo, cfg.WebPush)
	notifSvc := notifmod.NewNotificationService(notifRepo, d.nc, telegramSender, pushSender)
	notifConsumer := notifmod.NewConsumer(d.nc, notifRepo, notifSvc)
	if err := notifConsumer.Start(d.rootCtx); err != nil {
		slog.Error("failed to start notification consumer", "error", err)
	}
	d.addCleanup(notifConsumer.Stop)
	if telegramSender != nil {
		telegramSender.Start(d.rootCtx)
		d.addCleanup(telegramSender.Stop)
	}
	if pushSender != nil {
		pushSender.Start(d.rootCtx)
		d.addCleanup(pushSender.Stop)
	}
	telegramBind := notifmod.NewTelegramBindService(
		notifRepo,
		d.rdb,
		telegramSender,
		cfg.Telegram.BotToken,
		cfg.Telegram.BotUsername,
	)
	if telegramBind != nil {
		telegramBind.StartPolling(d.rootCtx)
	}
	notifHandler := notifmod.NewHandler(notifSvc, telegramBind)
	notifHandler.RegisterRoutes(api, requireAuth)
	d.notifSvc = notifSvc

	automationEngine := trackermod.NewAutomationEngine(
		d.trackerRepo,
		&messengerAutomationAdapter{svc: d.messengerSvc},
		&notificationAutomationAdapter{svc: notifSvc},
	)
	automationConsumer := trackermod.NewAutomationConsumer(d.nc, automationEngine)
	if err := automationConsumer.Start(d.rootCtx); err != nil {
		slog.Error("failed to start automation consumer", "error", err)
	}
	d.addCleanup(automationConsumer.Stop)
	d.automationConsumer = automationConsumer
}
