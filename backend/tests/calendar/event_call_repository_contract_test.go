package calendar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func eventCallRepositorySource(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	value, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "calendar", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(value)
}

func requireEventCallRepositorySource(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func requireEventCallRepositorySourceOrder(t *testing.T, source string, terms ...string) {
	t.Helper()
	for _, term := range terms {
		index := strings.Index(source, term)
		if index < 0 {
			t.Fatalf("source is missing ordered term %q", term)
		}
		source = source[index+len(term):]
	}
}

func TestEventCallRepository_FrozenSeparatePortAndLifetime(t *testing.T) {
	repository := eventCallRepositorySource(t, "event_call_repository.go")
	requireEventCallRepositorySource(t, repository,
		"type EventCallRepository interface {\n\tWithEventCallTransaction(context.Context, func(*EventCallTransaction) error) error\n}",
		"type EventCallTransactionBackend interface {\n\tResolveEventCallProjection(context.Context, EventCallProjectionRequest) (EventCallProjectionInput, error)\n\tPersistEventCallProjection(context.Context, EventCallProjectionRequest, EventCallProjection) (EventCallProjectionSnapshot, error)\n\tResolveEventCallDue(context.Context, EventCallDueRequest) (EventCallDueInput, error)\n\tPersistEventCallDue(context.Context, EventCallDueRequest, EventCallDueDecision) (EventCallDueSnapshot, error)\n}",
		"type EventCallProjectionRequest struct {\n\tscheduledCallID", "action", "boundary", "expectedEventVersion", "expectedCallVersion",
		"type EventCallDueRequest struct {\n\tscheduledCallID", "now", "envelope", "expectedEventVersion", "expectedCallVersion",
		"func NewEventCallProjectionRequest", "func NewEventCallDueRequest", "func RunEventCallTransaction",
		"ErrEventTransactionClosed", "eventCallTypedNil",
	)
	if strings.Contains(repository, "type EventRepository interface") {
		t.Fatal("event-call port must not redefine or expand EventRepository")
	}

	entRepository := eventCallRepositorySource(t, "event_ent_call.go")
	requireEventCallRepositorySource(t, entRepository, "var _ EventCallRepository = (*EventEntRepository)(nil)", "func (repository *EventEntRepository) WithEventCallTransaction")
	requireEventCallRepositorySourceOrder(t, entRepository,
		"transaction, err := repository.client.Tx(ctx)", "RunEventCallTransaction", "transaction.Commit()")
}

func TestEventEntCallBridge_LocksVersionsAndKeepsDueOutboxAtomic(t *testing.T) {
	source := eventCallRepositorySource(t, "event_ent_call.go")
	requireEventCallRepositorySource(t, source,
		"var _ EventCallTransactionBackend = (*eventEntCallTransactionBackend)(nil)",
		"calendarevent.ScheduledCallIDEQ", "scheduledcall.IDEQ", "ForUpdate()",
		"request.ExpectedEventVersion()", "request.ExpectedCallVersion()",
		"scheduledcall.VersionEQ(expectedCallVersion)", "outbox.NewPublisher().PublishInTx",
		"eventCallProjectionExactUTCSecond(*snapshot.DeletedAt)", "SetDeletedAt(snapshot.DeletedAt.UTC())",
		"calls.channel.", ".scheduled_due.v1",
	)
	requireEventCallRepositorySourceOrder(t, source,
		"backend.tx.CalendarEvent.Query()", "ForUpdate()", "backend.tx.ScheduledCall.Query()", "ForUpdate()")
	for _, forbidden := range []string{"github.com/nats-io/nats.go", "net/http", "MailEngine", "compatibility", "cutover", "time.Now"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("bridge source must not contain %q", forbidden)
		}
	}
}
