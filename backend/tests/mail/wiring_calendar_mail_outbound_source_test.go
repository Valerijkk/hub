package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestCalendarMailOutboundRuntimeSourceContract(t *testing.T) {
	runtime := readCalendarMailOutboundSource(t, "../../cmd/server/wiring_calendar_mail_runtime.go")
	for _, required := range []string{
		"func wireCalendarOutboundRuntime(",
		"mailboxRepo *mailmod.MailboxEntRepository",
		"drafts mailmod.CalendarMailDeliveryDraftCreator",
		"submissions *mailmod.MailboxSubmissionService",
		"mailmod.NewCalendarMailDeliveryService(",
		"mailmod.NewCalendarMailMaterializer()",
		"mailmod.NewCalendarMailDeliveryTimer()",
		"mailmod.DefaultCalendarMailDeliveryServiceOptions()",
		"calendarmod.NewCalendarInviteDeliveryConsumer(d.nc, mailboxRepo)",
		"calendarInviteDeliveryConsumer.Start(d.rootCtx)",
		"context.WithCancel(d.rootCtx)",
		"calendarMailDeliveryService.Run(calendarMailDeliveryWorkerCtx)",
		"slog.Error(\"background worker stopped\")",
		"var calendarMailDeliveryCleanup sync.Once",
		"calendarInviteDeliveryConsumer.Stop()",
		"calendarMailDeliveryWorkerCancel()",
		"<-calendarMailDeliveryWorkerDone",
	} {
		if !strings.Contains(runtime, required) {
			t.Fatalf("outbound runtime is missing %q", required)
		}
	}
	if strings.Contains(runtime, "slog.Error(\"background worker stopped\", \"error\", err)") {
		t.Fatal("outbound worker log must remain content-free")
	}
	serviceStart := strings.Index(runtime, "mailmod.NewCalendarMailDeliveryService(")
	consumerStart := strings.Index(runtime, "calendarmod.NewCalendarInviteDeliveryConsumer(d.nc, mailboxRepo)")
	consumerRun := strings.Index(runtime, "calendarInviteDeliveryConsumer.Start(d.rootCtx)")
	workerContext := strings.Index(runtime, "context.WithCancel(d.rootCtx)")
	if serviceStart < 0 || consumerStart <= serviceStart || consumerRun <= consumerStart || workerContext <= consumerRun {
		t.Fatal("outbound runtime must construct and start admission before the delivery worker")
	}
	if fatal := strings.Index(runtime[serviceStart:consumerStart], "d.fatal(\"failed to wire calendar and mail routes\", err)"); fatal < 0 {
		t.Fatal("delivery service construction must fail startup")
	}
	if fatal := strings.Index(runtime[consumerStart:consumerRun], "d.fatal(\"failed to wire calendar and mail routes\", err)"); fatal < 0 {
		t.Fatal("invite consumer construction must fail startup")
	}
	if fatal := strings.Index(runtime[consumerRun:workerContext], "d.fatal(\"failed to wire calendar and mail routes\", err)"); fatal < 0 {
		t.Fatal("invite consumer start must fail startup")
	}
	requireCalendarMailOutboundOrdered(t, runtime,
		"calendarInviteDeliveryConsumer.Start(d.rootCtx)",
		"context.WithCancel(d.rootCtx)",
		"calendarInviteDeliveryConsumer.Stop()",
		"calendarMailDeliveryWorkerCancel()",
		"<-calendarMailDeliveryWorkerDone",
	)
}

func TestCalendarMailOutboundNormalAndQAWiringSourceContract(t *testing.T) {
	normal := readCalendarMailOutboundSource(t, "../../cmd/server/wiring_calendar_mail.go")
	qa := readCalendarMailOutboundSource(t, "../../cmd/server/wiring_calendar_mail_qa.go")

	for _, source := range []string{normal, qa} {
		for _, required := range []string{
			"mailmod.NewCalendarMailDeliveryDraftCreator(mailboxSvc, mailboxRepo, mailboxEngine)",
			"wireCalendarOutboundRuntime(d, mailboxRepo, calendarMailDeliveryDraftCreator, mailboxSubmissionService)",
		} {
			if !strings.Contains(source, required) {
				t.Fatalf("outbound branch wiring is missing %q", required)
			}
		}
	}

	for _, source := range []string{normal, qa} {
		factory := strings.Index(source, "mailmod.NewCalendarMailDeliveryDraftCreator(mailboxSvc, mailboxRepo, mailboxEngine)")
		helper := strings.Index(source, "wireCalendarOutboundRuntime(d, mailboxRepo, calendarMailDeliveryDraftCreator, mailboxSubmissionService)")
		if factory < 0 || helper <= factory {
			t.Fatal("outbound runtime must receive the validated narrow Calendar draft creator")
		}
		if fatal := strings.Index(source[factory:helper], "d.fatal(\"failed to wire calendar and mail routes\", err)"); fatal < 0 {
			t.Fatal("Calendar draft creator factory must fail startup before outbound runtime starts")
		}
	}
	if strings.Contains(qa, "_ = mailboxSubmissionService") {
		t.Fatal("QA wiring must retain its real mailbox submission service for Calendar outbound delivery")
	}
}

func readCalendarMailOutboundSource(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

func requireCalendarMailOutboundOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 || position <= previous {
			t.Fatalf("outbound runtime does not preserve required order for %q", fragment)
		}
		previous = position
	}
}
