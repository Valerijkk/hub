package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

// mailboxEngineRuntimeInputs keeps provider construction outside the shared
// runtime while exposing only the narrow public Mail ports it needs.
type mailboxEngineRuntimeInputs struct {
	mailboxService        *mailmod.MailboxService
	mailboxRepository     *mailmod.MailboxEntRepository
	eventRepository       *calendarmod.EventEntRepository
	eventService          *calendarmod.EventService
	syncEngine            mailmod.MailboxSyncEngine
	messageReadEngine     mailmod.MailboxMessageReadEngine
	messageMutationEngine mailmod.MailboxMessageMutationEngine
	calendarPartEngine    mailmod.MailboxCalendarPartReadEngine

	calendarMailDeliveryDraftCreator mailmod.CalendarMailDeliveryDraftCreator
	draftService                     *mailmod.MailboxDraftService
	submissionDispatcher             *mailmod.MailboxSubmissionDispatcher
	submissionReconciler             *mailmod.MailboxSubmissionReconciler
	submissionFinalizer              *mailmod.MailboxSubmissionFinalizer
}

func wireMailboxEngineRuntime(d *serverDeps, inputs mailboxEngineRuntimeInputs) {
	mailboxSubmissionService, err := mailmod.NewMailboxSubmissionService(inputs.mailboxService, inputs.mailboxRepository)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	wireCalendarOutboundRuntime(d, inputs.mailboxRepository, inputs.calendarMailDeliveryDraftCreator, mailboxSubmissionService)

	mailboxSubmissionHandler, err := mailmod.NewMailboxSubmissionHandler(mailboxSubmissionService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxSubmissionHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxSubmissionWorker, err := mailmod.NewMailboxSubmissionWorker(
		inputs.mailboxRepository,
		inputs.submissionDispatcher,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionStaleWorker, err := mailmod.NewMailboxSubmissionStaleWorker(
		inputs.mailboxRepository,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionStaleWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionReconciliationWorker, err := mailmod.NewMailboxSubmissionReconciliationWorker(
		inputs.mailboxRepository,
		inputs.submissionReconciler,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionReconciliationWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionFinalizationWorker, err := mailmod.NewMailboxSubmissionFinalizationWorker(
		inputs.mailboxRepository,
		inputs.submissionFinalizer,
		mailmod.NewMailboxSubmissionWorkerTimer(),
		mailmod.DefaultMailboxSubmissionFinalizationWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxDraftHandler, err := mailmod.NewMailboxDraftCreateHandler(inputs.draftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyDraftService, err := mailmod.NewMailboxReplyDraftService(inputs.mailboxService, inputs.mailboxRepository, inputs.messageReadEngine, inputs.draftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyDraftHandler, err := mailmod.NewMailboxReplyDraftHandler(mailboxReplyDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxReplyDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyAllDraftService, err := mailmod.NewMailboxReplyAllDraftService(inputs.mailboxService, inputs.mailboxRepository, inputs.messageReadEngine, inputs.draftService, inputs.mailboxRepository)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyAllDraftHandler, err := mailmod.NewMailboxReplyAllDraftHandler(mailboxReplyAllDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxReplyAllDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxForwardDraftService, err := mailmod.NewMailboxForwardDraftService(inputs.mailboxService, inputs.mailboxRepository, inputs.messageReadEngine, inputs.draftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxForwardDraftHandler, err := mailmod.NewMailboxForwardDraftHandler(mailboxForwardDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxForwardDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	wireMailboxMessageDetail(d, inputs.mailboxService, inputs.mailboxRepository, inputs.messageReadEngine)
	mailboxMessageMutationService, err := mailmod.NewMailboxMessageMutationService(inputs.mailboxService, inputs.mailboxRepository, inputs.messageMutationEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxMessageMutationHandler, err := mailmod.NewMailboxMessageMutationHandler(mailboxMessageMutationService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxMessageMutationHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxSyncService, err := mailmod.NewMailboxSyncService(inputs.syncEngine, inputs.mailboxRepository, mailmod.MailboxSyncServiceConfig{
		Now:           serverUTCClock{}.Now,
		NewLeaseID:    uuid.New,
		NewEventID:    uuid.New,
		Jitter:        func(time.Duration) time.Duration { return 0 },
		LeaseDuration: 30 * time.Second,
		PollInterval:  5 * time.Second,
		RetryInterval: 15 * time.Second,
		PageSize:      100,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSyncWorker, err := mailmod.NewMailboxSyncWorker(inputs.mailboxRepository, mailboxSyncService, mailmod.NewMailboxSyncWorkerTimer(), mailmod.MailboxSyncWorkerConfig{
		Enabled:      true,
		ScanInterval: 2 * time.Second,
		StepTimeout:  30 * time.Second,
		RetryMin:     5 * time.Second,
		RetryMax:     time.Minute,
		BatchSize:    25,
		Concurrency:  2,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	wireCalendarInboundRuntime(d, inputs.mailboxRepository, inputs.calendarPartEngine, inputs.eventRepository, inputs.eventService)

	go func() {
		if err := mailboxSyncWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
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

}

func wireMailboxMessageDetail(d *serverDeps, mailboxSvc *mailmod.MailboxService, mailboxRepo *mailmod.MailboxEntRepository, engine mailmod.MailboxMessageReadEngine) {
	mailboxMessageService, err := mailmod.NewMailboxMessageService(mailboxSvc, mailboxRepo, engine, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxMessageHandler, err := mailmod.NewMailboxMessageHandler(mailboxMessageService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxMessageHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
}

func wireCalendarOutboundRuntime(
	d *serverDeps,
	mailboxRepo *mailmod.MailboxEntRepository,
	drafts mailmod.CalendarMailDeliveryDraftCreator,
	submissions *mailmod.MailboxSubmissionService,
) {
	calendarMailDeliveryService, err := mailmod.NewCalendarMailDeliveryService(
		mailboxRepo,
		drafts,
		submissions,
		mailmod.NewCalendarMailMaterializer(),
		mailmod.NewCalendarMailDeliveryTimer(),
		mailmod.DefaultCalendarMailDeliveryServiceOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarInviteDeliveryConsumer, err := calendarmod.NewCalendarInviteDeliveryConsumer(d.nc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := calendarInviteDeliveryConsumer.Start(d.rootCtx); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarMailDeliveryWorkerCtx, calendarMailDeliveryWorkerCancel := context.WithCancel(d.rootCtx)
	calendarMailDeliveryWorkerDone := make(chan struct{})
	go func() {
		defer close(calendarMailDeliveryWorkerDone)
		if err := calendarMailDeliveryService.Run(calendarMailDeliveryWorkerCtx); err != nil && calendarMailDeliveryWorkerCtx.Err() == nil {
			slog.Error("background worker stopped")
		}
	}()
	var calendarMailDeliveryCleanup sync.Once
	d.addCleanup(func() {
		calendarMailDeliveryCleanup.Do(func() {
			calendarInviteDeliveryConsumer.Stop()
			calendarMailDeliveryWorkerCancel()
			<-calendarMailDeliveryWorkerDone
		})
	})
}

func wireCalendarInboundRuntime(d *serverDeps, mailboxRepo *mailmod.MailboxEntRepository, engine mailmod.MailboxCalendarPartReadEngine, eventRepo *calendarmod.EventEntRepository, eventSvc *calendarmod.EventService) {
	calendarInboundSource := mailmod.NewMailboxEntCalendarPartSource(mailboxRepo, engine)
	calendarInboundService, err := calendarmod.NewCalendarInboundMessageService(eventRepo, calendarInboundSource, eventSvc)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarInboundDueWorker, err := calendarmod.NewCalendarInboundMessageDueWorker(
		eventRepo,
		calendarInboundService,
		calendarmod.NewCalendarInboundMessageDueWorkerTimer(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarInboundConsumer, err := mailmod.NewCalendarInboundConsumer(d.nc, calendarInboundService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := calendarInboundConsumer.Start(d.rootCtx); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarInboundWorkerCtx, calendarInboundWorkerCancel := context.WithCancel(d.rootCtx)
	calendarInboundWorkerDone := make(chan struct{})
	go func() {
		defer close(calendarInboundWorkerDone)
		if err := calendarInboundDueWorker.Run(calendarInboundWorkerCtx); err != nil && calendarInboundWorkerCtx.Err() == nil {
			slog.Error("background worker stopped")
		}
	}()
	var calendarInboundCleanup sync.Once
	d.addCleanup(func() {
		calendarInboundCleanup.Do(func() {
			calendarInboundConsumer.Stop()
			calendarInboundWorkerCancel()
			<-calendarInboundWorkerDone
		})
	})
}
