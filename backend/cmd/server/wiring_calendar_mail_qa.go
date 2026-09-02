//go:build qa && !release

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

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
	child, err := newQARecoveryChild(os.LookupEnv)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	var recoverySession qamailcalendar.PublishedRecoveryRouteSession
	if child.supervised() {
		evidence, ok := child.routeEvidence()
		if !ok {
			d.fatal("failed to wire calendar and mail routes", errQARecoveryChildInput)
			return true
		}
		recoverySession = newQARecoveryChildSession(evidence)
	}

	mailboxEngineConfig := mailmod.FakeMailEngineConfig{Now: serverUTCClock{}.Now}
	if child.supervised() {
		var fakeBootEntropy [32]byte
		if _, err := rand.Read(fakeBootEntropy[:]); err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		fakeBootPrefix := "qa-fake-" + hex.EncodeToString(fakeBootEntropy[:])
		var fakeIDSequence uint64
		mailboxEngineConfig.NextID = func() string {
			return fakeBootPrefix + "-" + strconv.FormatUint(atomic.AddUint64(&fakeIDSequence, 1), 10)
		}
	}
	mailboxEngine := mailmod.NewFakeMailEngine(mailboxEngineConfig)
	controller, err := qamailcalendar.NewController(d.db, mailboxEngine, config)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	if child.recovery() {
		if err := controller.RegisterPublishedRecoveryRoute(d.api, recoverySession, nil); err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		if err := child.announce(); err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
		return true
	}
	if err := controller.RegisterRoutes(d.api); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	if child.normal() {
		if err := controller.RegisterPublishedRecoveryRoute(d.api, recoverySession, func() error {
			err := child.requestRestart(d.rootCancel)
			if errors.Is(err, errQARecoveryChildRestart) {
				d.rootCancel()
			}
			return err
		}); err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
	}
	if child.supervised() {
		if err := child.announce(); err != nil {
			d.fatal("failed to wire calendar and mail routes", err)
			return true
		}
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
	wireCalendarOutboundRuntime(d, mailboxRepo, calendarMailDeliveryDraftCreator, mailboxSubmissionService, nil, mailmod.DefaultCalendarMailDeliveryServiceOptions())
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
	mailboxAccountProvisioner := mailmod.NewQAMailboxAccountProvisioner()
	mailboxAccountProvisioningWorkerOptions := mailmod.DefaultMailboxAccountProvisioningWorkerOptions()
	mailboxAccountProvisioningWorkerOptions.ScanInterval = time.Second
	mailboxAccountProvisioningWorker, err := mailmod.NewMailboxAccountProvisioningWorker(
		mailboxRepo,
		mailboxAccountProvisioner,
		mailmod.NewMailboxAccountProvisioningWorkerTimer(),
		mailboxAccountProvisioningWorkerOptions,
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}
	mailboxAccountProvisioningCtx, cancelMailboxAccountProvisioning := context.WithCancel(d.rootCtx)
	mailboxAccountProvisioningDone := make(chan struct{})
	var mailboxAccountProvisioningCleanup sync.Once
	d.addCleanup(func() {
		mailboxAccountProvisioningCleanup.Do(func() {
			cancelMailboxAccountProvisioning()
			<-mailboxAccountProvisioningDone
		})
	})
	go func() {
		defer close(mailboxAccountProvisioningDone)
		if err := mailboxAccountProvisioningWorker.Run(mailboxAccountProvisioningCtx); err != nil && mailboxAccountProvisioningCtx.Err() == nil {
			slog.Error("background worker stopped")
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
	wireMailboxMessageDetail(d, mailboxSvc, mailboxRepo, mailboxEngine)
	wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc, nil)
	return true
}

type qaRecoveryChildSession struct {
	evidence qamailcalendar.PublishedRecoveryChildEvidence
}

func newQARecoveryChildSession(evidence qaRecoveryChildEvidence) qaRecoveryChildSession {
	return qaRecoveryChildSession{evidence: qamailcalendar.PublishedRecoveryChildEvidence{
		ChildPID:   evidence.pid,
		Generation: evidence.generation,
		BootSHA256: evidence.bootSHA256,
	}}
}

func (session qaRecoveryChildSession) PublishedRecoveryChildEvidence() qamailcalendar.PublishedRecoveryChildEvidence {
	return session.evidence
}
