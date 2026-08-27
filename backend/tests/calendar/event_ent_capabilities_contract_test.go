package calendar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func calendarCapabilitySource(t *testing.T, name string) string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fail()
		return ""
	}
	value, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "calendar", name))
	if err != nil {
		t.Fail()
	}
	return string(value)
}

func requireCalendarSource(t *testing.T, text string, terms ...string) {
	for _, term := range terms {
		if !strings.Contains(text, term) {
			t.Fail()
		}
	}
}

func requireCalendarSourceOrder(t *testing.T, text string, terms ...string) {
	for _, term := range terms {
		index := strings.Index(text, term)
		if index < 0 {
			t.Fail()
			return
		}
		text = text[index+len(term):]
	}
}

func TestEventEntCapabilities_FrozenPortAndUpdateCASContract(t *testing.T) {
	repository := calendarCapabilitySource(t, "event_repository.go")
	requireCalendarSource(t, repository, "type EventRepository interface {\n\tWithEventTransaction(ctx context.Context, callback func(*EventTransaction) error) error\n}")
	update := calendarCapabilitySource(t, "event_ent_update.go")
	requireCalendarSourceOrder(t, update, "backend.activeOrganizationMemberForUpdate", "backend.tx.CalendarEvent.Query()", "backend.tx.CalendarAttendee.Query()", "backend.persistCalendarEventUpdateMaster", "backend.AppendOutbox", "NewCalendarEventUpdatePersistenceResult")
	requireCalendarSource(t, update, "calendarattendee.VersionEQ(expectedVersion)", "calendarevent.VersionEQ(expected)", "return calendarEntCASFailure(err)")
	requireCalendarSourceOrder(t, update, "func (backend *eventEntTransactionBackend) PersistCalendarEventOccurrenceCancel", "backend.activeOrganizationMemberForUpdate", "backend.tx.CalendarEvent.Query()", "backend.tx.CalendarEventException.Query()", "backend.persistCalendarEventUpdateMaster", "backend.AppendOutbox", "NewCalendarEventOccurrenceCancelPersistenceResult")
	requireCalendarSource(t, update, "var _ CalendarEventUpdateTransactionBackend = (*eventEntTransactionBackend)(nil)", "var _ CalendarEventOccurrenceCancelTransactionBackend = (*eventEntTransactionBackend)(nil)")
}

func TestEventEntRSVP_BindsDurableIdentityCASAndReplyDelivery(t *testing.T) {
	source := calendarCapabilitySource(t, "event_ent_rsvp.go")
	requireCalendarSource(t, source, "var _ CalendarEventSeriesRSVPTransactionBackend = (*eventEntTransactionBackend)(nil)", "var _ CalendarEventOccurrenceRSVPTransactionBackend = (*eventEntTransactionBackend)(nil)")
	requireCalendarSourceOrder(t, source, "func (backend *eventEntTransactionBackend) PersistCalendarEventSeriesRSVP", "!command.valid()", "backend.tx.CalendarEvent.Query()", "backend.tx.CalendarAttendee.Query()", "eventEntSeriesRSVPImmutableIdentityMatches", "locked.Version != command.ExpectedVersion()", "calendarattendee.VersionEQ(expectedVersion)", "backend.persistCalendarReplyDelivery(ctx, event.OrganizationID, event.ID, attendee.ID, command.EventSequence(), requested.Version, nil, requested.Status)", "backend.AppendOutbox", "NewCalendarEventSeriesRSVPPersistenceResult")
	requireCalendarSourceOrder(t, source, "func (backend *eventEntTransactionBackend) PersistCalendarEventOccurrenceRSVP", "!command.valid()", "durableSeries, durableSeriesErr := eventEntSeriesRSVPAttendeeSnapshot(currentSeries)", "!reflect.DeepEqual(durableSeries, series)", "durablePrior, durablePriorErr := eventEntOccurrenceRSVPAttendeeSnapshot(currentOccurrence)", "!reflect.DeepEqual(durablePrior, prior)", "validCalendarEventOccurrenceRSVPAttendeeForSeries(durableSeries, requested)", "backend.persistCalendarReplyDelivery(ctx, event.OrganizationID, event.ID, appliedAttendeeID, command.EventSequence(), requested.Version, &recurrenceID, requested.Status)")
	requireCalendarSourceOrder(t, source, "func (backend *eventEntTransactionBackend) persistCalendarReplyDelivery", "databaseNow, err := backend.calendarInboundMessageDatabaseClock(ctx)", "backend.tx.CalendarReplyDelivery.Create()", "SetNextAttemptAt(databaseNow)", "create.Save(ctx)")
	requireCalendarSource(t, source, "durable.Email == requested.Email", "durable.DisplayName == requested.DisplayName", "durable.Role == requested.Role", "durable.RemovedAt == nil && requested.RemovedAt == nil")
}

func TestCalendarInboundReply_ReceiptOutboxReplayAndNoExternalIO(t *testing.T) {
	inbound := calendarCapabilitySource(t, "event_ent_inbound_reply.go")
	requireCalendarSource(t, inbound, "var _ CalendarInboundReplyTransactionBackend = (*eventEntTransactionBackend)(nil)")
	requireCalendarSourceOrder(t, inbound, "backend.ResolveCalendarInboundReplyOwner", "backend.tx.CalendarEvent.Query()", "backend.tx.CalendarAttendee.Query()", "backend.tx.CalendarInboundReplyReceipt.Query()", "calendarattendee.VersionEQ(expectedVersion)", "backend.persistCalendarInboundReplyReceipt", "backend.appendCalendarInboundReplyOutbox", "NewCalendarInboundReplyPersistenceResult")
	service := calendarCapabilitySource(t, "event_service_inbound_reply.go")
	requireCalendarSourceOrder(t, service, "if authorization.ExistingReceipt {", "CalendarInboundReplyDuplicate", "calendarInboundReplyAuthorize")
	for _, name := range []string{"event_ent_update.go", "event_ent_rsvp.go", "event_ent_inbound_reply.go"} {
		source := calendarCapabilitySource(t, name)
		if strings.Contains(source, "net/http") || strings.Contains(source, "github.com/nats-io/nats.go") || strings.Contains(source, "MailEngine") || strings.Contains(source, "NATS") {
			t.Fail()
		}
	}
}
