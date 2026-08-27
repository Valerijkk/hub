package mail_test

import (
	"os"
	"strings"
	"testing"
)

func qaCalendarMailWiringSource(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile("../../cmd/server/wiring_calendar_mail_qa.go")
	if err != nil {
		t.Fatalf("read QA calendar/mail wiring: %v", err)
	}
	return string(source)
}

func TestQACalendarMailSubmissionRuntimeWiringSourceContract(t *testing.T) {
	source := qaCalendarMailWiringSource(t)

	for _, required := range []string{
		"//go:build qa && !release",
		"mailmod.NewMailboxSubmissionService(mailboxSvc, mailboxRepo)",
		"mailmod.NewMailboxSubmissionDispatcher(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionWorker(",
		"mailmod.NewMailboxSubmissionStaleWorker(",
		"mailmod.NewMailboxSubmissionReconciler(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionReconciliationWorker(",
		"mailmod.NewMailboxSubmissionFinalizer(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionFinalizationWorker(",
		"mailmod.NewMailboxSubmissionWorkerTimer()",
		"mailmod.DefaultMailboxSubmissionWorkerOptions()",
		"mailmod.DefaultMailboxSubmissionStaleWorkerOptions()",
		"mailmod.DefaultMailboxSubmissionReconciliationWorkerOptions()",
		"mailmod.DefaultMailboxSubmissionFinalizationWorkerOptions()",
		"mailboxSubmissionWorker.Run(d.rootCtx)",
		"mailboxSubmissionStaleWorker.Run(d.rootCtx)",
		"mailboxSubmissionReconciliationWorker.Run(d.rootCtx)",
		"mailboxSubmissionFinalizationWorker.Run(d.rootCtx)",
		`slog.Error("background worker stopped", "error", err)`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("QA submission runtime wiring is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"NewMailboxSubmissionHandler",
		"NewMailboxEntDraftAttachmentMaterializer",
		"mailboxObjectStore",
		"CalendarMailDelivery",
		"Stub",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("QA submission runtime wiring contains forbidden %q", forbidden)
		}
	}

	previous := -1
	for _, fragment := range []string{
		"mailboxEngine := mailmod.NewFakeMailEngine",
		"controller, err := qamailcalendar.NewController",
		"controller.RegisterRoutes(d.api)",
		"mailmod.NewMailboxSubmissionService(mailboxSvc, mailboxRepo)",
		"mailmod.NewMailboxSubmissionDispatcher(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionWorker(",
		"mailmod.NewMailboxSubmissionStaleWorker(",
		"mailmod.NewMailboxSubmissionReconciler(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionReconciliationWorker(",
		"mailmod.NewMailboxSubmissionFinalizer(mailboxRepo, mailboxEngine)",
		"mailmod.NewMailboxSubmissionFinalizationWorker(",
		"mailboxSubmissionWorker.Run(d.rootCtx)",
		"mailboxSubmissionStaleWorker.Run(d.rootCtx)",
		"mailboxSubmissionReconciliationWorker.Run(d.rootCtx)",
		"mailboxSubmissionFinalizationWorker.Run(d.rootCtx)",
		"wireMailboxMessageDetail(d, mailboxSvc, mailboxRepo, mailboxEngine)",
		"wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc)",
	} {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("QA submission runtime wiring is missing ordered %q", fragment)
		}
		if position <= previous {
			t.Fatalf("QA submission runtime wiring puts %q before its predecessor", fragment)
		}
		previous = position
	}
	if got := strings.Count(source, ".Run(d.rootCtx)"); got != 4 {
		t.Fatalf("QA submission worker Run calls = %d; want 4", got)
	}
	if got := strings.Count(source, "mailmod.NewMailboxSubmissionWorkerTimer()"); got != 4 {
		t.Fatalf("QA submission worker timers = %d; want 4", got)
	}
	if got := strings.Count(source, `slog.Error("background worker stopped", "error", err)`); got != 4 {
		t.Fatalf("QA submission worker error logs = %d; want 4", got)
	}
}

func TestQACalendarMailAccountProvisioningRuntimeWiringSourceContract(t *testing.T) {
	source := qaCalendarMailWiringSource(t)

	for _, required := range []string{
		`"context"`,
		`"sync"`,
		`"time"`,
		"mailboxAccountProvisioner := mailmod.NewQAMailboxAccountProvisioner()",
		"mailboxAccountProvisioningWorkerOptions := mailmod.DefaultMailboxAccountProvisioningWorkerOptions()",
		"mailboxAccountProvisioningWorkerOptions.ScanInterval = time.Second",
		`mailboxAccountProvisioningWorker, err := mailmod.NewMailboxAccountProvisioningWorker(
		mailboxRepo,
		mailboxAccountProvisioner,
		mailmod.NewMailboxAccountProvisioningWorkerTimer(),
		mailboxAccountProvisioningWorkerOptions,
	)`,
		"mailboxAccountProvisioningCtx, cancelMailboxAccountProvisioning := context.WithCancel(d.rootCtx)",
		"mailboxAccountProvisioningDone := make(chan struct{})",
		"var mailboxAccountProvisioningCleanup sync.Once",
		`d.addCleanup(func() {
	mailboxAccountProvisioningCleanup.Do(func() {
		cancelMailboxAccountProvisioning()
		<-mailboxAccountProvisioningDone
	})
})`,
		"defer close(mailboxAccountProvisioningDone)",
		"mailboxAccountProvisioningWorker.Run(mailboxAccountProvisioningCtx)",
		"mailboxAccountProvisioningCtx.Err() == nil",
		`slog.Error("background worker stopped")`,
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("QA account provisioning runtime wiring is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"mailmod.NewMailboxRegistryProvisioningEngine(",
		"stalwart.NewRegistryDomainController(",
		"stalwart.NewRegistryAccountClient(",
		"mailmod.NewMailboxDomainProvisioningWorker(",
		`mailmod.NewMailboxAccountProvisioningWorker(
		mailboxRepo,
		mailboxEngine,`,
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("QA account provisioning runtime wiring contains forbidden %q", forbidden)
		}
	}

	previous := -1
	for _, fragment := range []string{
		"mailboxSubmissionFinalizationWorker, err := mailmod.NewMailboxSubmissionFinalizationWorker(",
		"mailboxAccountProvisioner := mailmod.NewQAMailboxAccountProvisioner()",
		"mailboxAccountProvisioningWorker, err := mailmod.NewMailboxAccountProvisioningWorker(",
		"mailboxAccountProvisioningCtx, cancelMailboxAccountProvisioning := context.WithCancel(d.rootCtx)",
		"mailboxAccountProvisioningWorker.Run(mailboxAccountProvisioningCtx)",
		"mailboxSubmissionWorker.Run(d.rootCtx)",
		"mailboxSubmissionStaleWorker.Run(d.rootCtx)",
		"mailboxSubmissionReconciliationWorker.Run(d.rootCtx)",
		"mailboxSubmissionFinalizationWorker.Run(d.rootCtx)",
		"wireMailboxMessageDetail(d, mailboxSvc, mailboxRepo, mailboxEngine)",
		"wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc)",
	} {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("QA account provisioning runtime wiring is missing ordered %q", fragment)
		}
		if position <= previous {
			t.Fatalf("QA account provisioning runtime wiring puts %q before its predecessor", fragment)
		}
		previous = position
	}

	for fragment, want := range map[string]int{
		"mailmod.NewQAMailboxAccountProvisioner()":                               1,
		"mailmod.NewMailboxAccountProvisioningWorker(":                          1,
		"mailmod.NewMailboxAccountProvisioningWorkerTimer()":                    1,
		"mailmod.DefaultMailboxAccountProvisioningWorkerOptions()":              1,
		"mailboxAccountProvisioningWorkerOptions.ScanInterval = time.Second":    1,
		"context.WithCancel(d.rootCtx)":                                         1,
		"make(chan struct{})":                                                   1,
		"sync.Once":                                                             1,
		"d.addCleanup(func() {":                                                 1,
		"defer close(mailboxAccountProvisioningDone)":                           1,
		"mailboxAccountProvisioningWorker.Run(mailboxAccountProvisioningCtx)":  1,
		`slog.Error("background worker stopped")`:                               1,
		"qamailcalendar.ParseConfig(os.LookupEnv)":                              1,
		"qamailcalendar.NewController(d.db, mailboxEngine, config)":            1,
	} {
		if got := strings.Count(source, fragment); got != want {
			t.Fatalf("QA account provisioning runtime wiring occurrences of %q = %d; want %d", fragment, got, want)
		}
	}
}
