package main

import (
	"context"
	"log/slog"
	"sync"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

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
