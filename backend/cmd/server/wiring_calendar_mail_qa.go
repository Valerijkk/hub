//go:build qa && !release

package main

import (
	"log/slog"
	"os"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/qamailcalendar"
)

func wireQAMailCalendarRuntime(d *serverDeps, mailboxSvc *mailmod.MailboxService, mailboxRepo *mailmod.MailboxEntRepository, eventRepo *calendarmod.EventEntRepository, eventSvc *calendarmod.EventService) bool {
	config, err := qamailcalendar.ParseConfig(os.LookupEnv)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}

	mailboxEngine := mailmod.NewFakeMailEngine(mailmod.FakeMailEngineConfig{
		Now: serverUTCClock{}.Now,
	})
	controller, err := qamailcalendar.NewController(d.db, mailboxEngine, config)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	if err := controller.RegisterRoutes(d.api); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionService, err := mailmod.NewMailboxSubmissionService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	calendarMailDeliveryDraftCreator, err := mailmod.NewCalendarMailDeliveryDraftCreator(mailboxSvc, mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	wireCalendarOutboundRuntime(d, mailboxRepo, calendarMailDeliveryDraftCreator, mailboxSubmissionService)
	mailboxSubmissionDispatcher, err := mailmod.NewMailboxSubmissionDispatcher(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionWorker, err := mailmod.NewMailboxSubmissionWorker(
		mailboxRepo,
		mailboxSubmissionDispatcher,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionStaleWorker, err := mailmod.NewMailboxSubmissionStaleWorker(
		mailboxRepo,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionStaleWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionReconciler, err := mailmod.NewMailboxSubmissionReconciler(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionReconciliationWorker, err := mailmod.NewMailboxSubmissionReconciliationWorker(
		mailboxRepo,
		mailboxSubmissionReconciler,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionReconciliationWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionFinalizer, err := mailmod.NewMailboxSubmissionFinalizer(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxSubmissionFinalizationWorker, err := mailmod.NewMailboxSubmissionFinalizationWorker(
		mailboxRepo,
		mailboxSubmissionFinalizer,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionFinalizationWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	go func() {
		if err := mailboxSubmissionWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	go func() {
		if err := mailboxSubmissionStaleWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	go func() {
		if err := mailboxSubmissionReconciliationWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	go func() {
		if err := mailboxSubmissionFinalizationWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	wireMailboxMessageDetail(d, mailboxSvc, mailboxRepo, mailboxEngine)
	wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc)
	return true
}
