//go:build !qa && !release

package main

import (
	"crypto/rand"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
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

		fakeNow := func() time.Time {
			return time.Date(2026, time.September, 1, 12, 1, 0, 0, time.UTC)
		}
		mailFakeIDSource := devdemo.NewMailFakeIDSource()
		mailboxEngine := mailmod.NewFakeMailEngine(mailmod.FakeMailEngineConfig{
			LocalDelivery: true,
			NextID:        mailFakeIDSource.NextID,
			Now: fakeNow,
		})
		if mailboxEngine == nil {
			d.fatal("failed to wire calendar and mail routes", errors.New("dev mail engine is unavailable"))
			return true
		}
		runtimeStart := make(chan struct{})
		var runtimeStartGuard sync.Mutex
		runtimeStarted := false
		startRuntime := func() error {
			runtimeStartGuard.Lock()
			defer runtimeStartGuard.Unlock()
			if runtimeStarted {
				return nil
			}
			if err := mailFakeIDSource.ActivateRuntime(rand.Reader); err != nil {
				return err
			}
			close(runtimeStart)
			runtimeStarted = true
			return nil
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
		mailAdapter, err := newDevMailDemoAdapter(d.db, mailboxEngine, mailboxSvc, mailboxRepo, eventReadSvc, serverUTCClock{}.Now, startRuntime)
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

		d.api.Use("/mail", d.requireAuth, func(ctx fiber.Ctx) error {
			if !mailFakeIDSource.RuntimeReady() {
				ctx.Set("X-Hub-Error-Code", "MAIL_RUNTIME_NOT_READY")
				ctx.Set("Retry-After", "1")
				return fiber.NewError(fiber.StatusServiceUnavailable, "mail runtime is not ready")
			}
			return ctx.Next()
		})
		calendarMailDeliveryServiceOptions := mailmod.DefaultCalendarMailDeliveryServiceOptions()
		calendarMailDeliveryServiceOptions.ScanInterval = 2 * time.Second
		submissionWorkerOptions := mailmod.DefaultMailboxSubmissionWorkerOptions()
		submissionWorkerOptions.ScanInterval = 2 * time.Second
		submissionFinalizationWorkerOptions := mailmod.DefaultMailboxSubmissionFinalizationWorkerOptions()
		submissionFinalizationWorkerOptions.ScanInterval = 2 * time.Second
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
			calendarMailDeliveryServiceOptions:  calendarMailDeliveryServiceOptions,
			submissionWorkerOptions:             submissionWorkerOptions,
			submissionFinalizationWorkerOptions: submissionFinalizationWorkerOptions,
			startGate:                       runtimeStart,
		})
		d.mailCalendarDemoSeeder = mailCalendarDemoSeeder
		return true
	}

	return false
}
