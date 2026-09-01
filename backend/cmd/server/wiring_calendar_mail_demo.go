//go:build !qa && !release

package main

import (
	"errors"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

func wireDevMailCalendarRuntime(
	d *serverDeps,
	mailboxSvc *mailmod.MailboxService,
	mailboxRepo *mailmod.MailboxEntRepository,
	eventRepo *calendarmod.EventEntRepository,
	eventSvc *calendarmod.EventService,
	eventReadSvc *calendarmod.EventReadService,
) bool {
	config, err := devdemo.ParseMailDemoRuntimeConfig(os.LookupEnv, d.cfg.Server.DevMode)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}

	if d.cfg.Server.DevMode {
		calendarDemoAdapter, err := newCalendarDemoAdapter(eventSvc, eventReadSvc)
		if err != nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", err)
			return true
		}
		if calendarDemoAdapter == nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("calendar demo adapter is unavailable"))
			return true
		}

		if !config.Enabled {
			mailCalendarDemoSeeder, err := devdemo.NewMailCalendarDemoSeeder(d.db, d.rawDB, d.postgresAdvisoryLocks, calendarDemoAdapter)
			if err != nil {
				d.fatal("failed to wire Mail/Calendar demo seeder", err)
				return true
			}
			if mailCalendarDemoSeeder == nil {
				d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("demo seeder is unavailable"))
				return true
			}
			d.mailCalendarDemoSeeder = mailCalendarDemoSeeder
			return false
		}

		var fakeIDSequence uint64
		fakeNow := func() time.Time {
			return time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC)
		}
		mailboxEngine := mailmod.NewFakeMailEngine(mailmod.FakeMailEngineConfig{
			LocalDelivery: true,
			NextID: func() string {
				return "dev-mail-fake-v1-" + strconv.FormatUint(atomic.AddUint64(&fakeIDSequence, 1), 10)
			},
			Now: fakeNow,
		})
		if mailboxEngine == nil {
			d.fatal("failed to wire calendar and mail routes", errors.New("dev mail engine is unavailable"))
			return true
		}

		calendarMailDeliveryDraftCreator, err := mailmod.NewCalendarMailDeliveryDraftCreator(mailboxSvc, mailboxRepo, mailboxEngine)
		if err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		mailboxSubmissionDispatcher, err := mailmod.NewMailboxSubmissionDispatcher(mailboxRepo, mailboxEngine)
		if err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		mailboxSubmissionReconciler, err := mailmod.NewMailboxSubmissionReconciler(mailboxRepo, mailboxEngine)
		if err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		mailboxSubmissionFinalizer, err := mailmod.NewMailboxSubmissionFinalizer(mailboxRepo, mailboxEngine)
		if err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		mailboxDraftService, err := mailmod.NewMailboxDraftService(
			mailboxSvc,
			mailboxRepo,
			mailboxEngine,
			mailboxRepo,
			mailboxRepo,
			mailmod.NewMailboxDraftAttachmentMaterializerDemo(),
			mailboxRepo,
		)
		if err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		mailAdapter, err := newDevMailDemoAdapter(d.db, mailboxEngine, mailboxSvc, mailboxRepo, eventReadSvc, fakeNow)
		if err != nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", err)
			return true
		}
		if mailAdapter == nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("mail demo adapter is unavailable"))
			return true
		}
		mailCalendarDemoSeeder, err := devdemo.NewMailCalendarDemoSeederWithMail(d.db, d.rawDB, d.postgresAdvisoryLocks, calendarDemoAdapter, mailAdapter)
		if err != nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", err)
			return true
		}
		if mailCalendarDemoSeeder == nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("demo seeder is unavailable"))
			return true
		}

		wireMailboxEngineRuntime(d, mailboxEngineRuntimeInputs{
			mailboxService:                  mailboxSvc,
			mailboxRepository:               mailboxRepo,
			eventRepository:                 eventRepo,
			eventService:                    eventSvc,
			syncEngine:                      mailboxEngine,
			messageReadEngine:               mailboxEngine,
			messageMutationEngine:           mailboxEngine,
			calendarPartEngine:              mailboxEngine,
			calendarMailDeliveryDraftCreator: calendarMailDeliveryDraftCreator,
			draftService:                    mailboxDraftService,
			submissionDispatcher:            mailboxSubmissionDispatcher,
			submissionReconciler:            mailboxSubmissionReconciler,
			submissionFinalizer:             mailboxSubmissionFinalizer,
		})
		d.mailCalendarDemoSeeder = mailCalendarDemoSeeder
		return true
	}

	return false
}
