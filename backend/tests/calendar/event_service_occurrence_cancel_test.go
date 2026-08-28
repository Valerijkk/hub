package calendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	ical "github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	occurrenceActorID        = uuid.MustParse("a1111111-1111-4111-8111-111111111111")
	occurrenceOrganizationID = uuid.MustParse("a2222222-2222-4222-8222-222222222222")
	occurrenceEventID        = uuid.MustParse("a3333333-3333-4333-8333-333333333333")
	occurrenceExceptionID    = uuid.MustParse("a4444444-4444-4444-8444-444444444444")
	occurrenceEnvelopeID     = uuid.MustParse("a5555555-5555-4555-8555-555555555555")
	occurrenceAttendeeID     = uuid.MustParse("a6666666-6666-4666-8666-666666666666")
	occurrenceExternalID     = uuid.MustParse("a7777777-7777-4777-8777-777777777777")
)

const occurrenceRecurrenceID = "2031-01-21T09:00:00Z"

func TestResolveNominalOccurrence_ResolvesCanonicalTypedTemporalWithoutExdatePolicy(t *testing.T) {
	timedStart := time.Date(2031, time.March, 1, 14, 30, 0, 0, time.UTC)
	for _, testCase := range []struct {
		name         string
		event        ical.CalendarEvent
		rrule        string
		recurrenceID string
		want         ical.CalendarEvent
		ok           bool
	}{
		{
			name:         "timed weekly preserves master duration",
			event:        ical.CalendarEvent{StartsAt: timedStart, EndsAt: timedStart.Add(90 * time.Minute), Timezone: "America/New_York"},
			rrule:        "FREQ=WEEKLY;INTERVAL=1;BYDAY=SA;COUNT=3",
			recurrenceID: "2031-03-08T14:30:00Z",
			want:         ical.CalendarEvent{StartsAt: time.Date(2031, time.March, 8, 14, 30, 0, 0, time.UTC), EndsAt: time.Date(2031, time.March, 8, 16, 0, 0, 0, time.UTC), Timezone: "America/New_York"},
			ok:           true,
		},
		{
			name:         "all day monthly preserves exclusive duration",
			event:        ical.CalendarEvent{AllDay: true, StartDate: "2031-01-31", EndDateExclusive: "2031-02-02", Timezone: "Europe/Moscow"},
			rrule:        "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31;COUNT=3",
			recurrenceID: "2031-03-31",
			want:         ical.CalendarEvent{AllDay: true, StartDate: "2031-03-31", EndDateExclusive: "2031-04-02", Timezone: "Europe/Moscow"},
			ok:           true,
		},
		{
			name:         "dst gap is not a nominal timed member",
			event:        ical.CalendarEvent{StartsAt: time.Date(2027, time.March, 13, 7, 30, 0, 0, time.UTC), EndsAt: time.Date(2027, time.March, 13, 8, 30, 0, 0, time.UTC), Timezone: "America/New_York"},
			rrule:        "FREQ=DAILY;INTERVAL=1;COUNT=3",
			recurrenceID: "2027-03-14T06:30:00Z",
		},
		{
			name:         "dst gap skips to next valid local member",
			event:        ical.CalendarEvent{StartsAt: time.Date(2027, time.March, 13, 7, 30, 0, 0, time.UTC), EndsAt: time.Date(2027, time.March, 13, 8, 30, 0, 0, time.UTC), Timezone: "America/New_York"},
			rrule:        "FREQ=DAILY;INTERVAL=1;COUNT=3",
			recurrenceID: "2027-03-15T06:30:00Z",
			want:         ical.CalendarEvent{StartsAt: time.Date(2027, time.March, 15, 6, 30, 0, 0, time.UTC), EndsAt: time.Date(2027, time.March, 15, 7, 30, 0, 0, time.UTC), Timezone: "America/New_York"},
			ok:           true,
		},
		{
			name:         "noncanonical rule is rejected",
			event:        ical.CalendarEvent{StartsAt: timedStart, EndsAt: timedStart.Add(time.Hour), Timezone: "UTC"},
			rrule:        "INTERVAL=1;FREQ=DAILY;COUNT=2",
			recurrenceID: "2031-03-02T14:30:00Z",
		},
		{
			name:         "nonmember is rejected",
			event:        ical.CalendarEvent{StartsAt: timedStart, EndsAt: timedStart.Add(time.Hour), Timezone: "UTC"},
			rrule:        "FREQ=DAILY;INTERVAL=1;COUNT=2",
			recurrenceID: "2031-03-03T14:30:00Z",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			occurrence, ok := ical.ResolveNominalOccurrence(testCase.event, testCase.rrule, testCase.recurrenceID)
			if ok != testCase.ok {
				t.Fatalf("ResolveNominalOccurrence() ok = %t, want %t", ok, testCase.ok)
			}
			if !ok {
				return
			}
			if occurrence.RecurrenceID != testCase.recurrenceID || occurrence.Temporal != testCase.want {
				t.Fatalf("ResolveNominalOccurrence() = %#v, want recurrence=%q temporal=%#v", occurrence, testCase.recurrenceID, testCase.want)
			}
		})
	}
}

func TestCalendarEventOccurrenceCancelCommand_ValidatesAndFreezesInput(t *testing.T) {
	command, err := calendar.NewCalendarEventOccurrenceCancelCommand(calendar.CalendarEventOccurrenceCancelInput{
		EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, ExpectedVersion: 3,
	})
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceCancelCommand() error = %v", err)
	}
	if command.EventID() != occurrenceEventID || command.RecurrenceID() != occurrenceRecurrenceID || command.ExpectedVersion() != 3 {
		t.Fatalf("command getters = %#v", command)
	}
	for _, input := range []calendar.CalendarEventOccurrenceCancelInput{
		{},
		{EventID: occurrenceEventID, ExpectedVersion: 3},
		{EventID: occurrenceEventID, RecurrenceID: "\xff", ExpectedVersion: 3},
		{EventID: occurrenceEventID, RecurrenceID: strings.Repeat("x", 256), ExpectedVersion: 3},
		{EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, ExpectedVersion: 0},
	} {
		if _, err := calendar.NewCalendarEventOccurrenceCancelCommand(input); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("NewCalendarEventOccurrenceCancelCommand(%#v) error = %v, want invalid", input, err)
		}
	}
}

func TestCalendarEventOccurrenceCancelResult_RequiresCanonicalUTCSecondRecurrenceKey(t *testing.T) {
	if _, err := calendar.NewCalendarEventOccurrenceCancelResult(occurrenceEventID, occurrenceRecurrenceID, calendar.CalendarEventStatusCancelled, 1, 1); err != nil {
		t.Fatalf("NewCalendarEventOccurrenceCancelResult(canonical) error = %v", err)
	}
	for _, recurrenceID := range []string{
		"2031-01-21T09:00:00+00:00",
		"2031-01-21T09:00:00.000Z",
		"2031-01-21T09:00:00z",
	} {
		if _, err := calendar.NewCalendarEventOccurrenceCancelResult(occurrenceEventID, recurrenceID, calendar.CalendarEventStatusCancelled, 1, 1); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("NewCalendarEventOccurrenceCancelResult(%q) error = %v, want invalid", recurrenceID, err)
		}
	}
}

func TestEventTransactionOccurrenceCancel_RequiresOptionalCapabilityBeforeBackend(t *testing.T) {
	base := &occurrenceBaseOnlyBackend{}
	err := calendar.RunEventTransaction(base, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventOccurrenceCancelScope(context.Background(), occurrenceActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("ResolveCalendarEventOccurrenceCancelScope() error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if base.organizationCalls != 0 || base.otherCalls != 0 {
		t.Fatalf("unsupported optional port made backend calls org:%d other:%d", base.organizationCalls, base.otherCalls)
	}

	backend := newOccurrenceBackendFake()
	var retained *calendar.EventTransaction
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		scope, err := transaction.ResolveCalendarEventOccurrenceCancelScope(context.Background(), occurrenceActorID)
		if err != nil {
			return err
		}
		if scope.Operation() != "" || scope.EventOperation() != calendar.EventMutationCalendarOccurrenceCancel || scope.ActorID() != occurrenceActorID || scope.OrganizationID() != occurrenceOrganizationID {
			t.Fatalf("receipt-free occurrence scope = %#v", scope)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() capability backend error = %v", err)
	}
	if _, err := retained.ResolveCalendarEventOccurrenceCancelScope(context.Background(), occurrenceActorID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped scope error = %v, want transaction closed", err)
	}
}

func TestEventTransactionOccurrenceCancel_RejectsTypedNilCapabilityBeforeBackend(t *testing.T) {
	var typedNil *occurrenceBackendFake
	err := calendar.RunEventTransaction(typedNil, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventOccurrenceCancelScope(context.Background(), occurrenceActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("typed-nil occurrence capability error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil { t.Fatalf("RunEventTransaction(typed nil) error = %v", err) }
}

func TestEventTransactionOccurrenceCancel_RejectsOnlyIngressInvalidResolverRequest(t *testing.T) {
	for _, recurrenceID := range []string{"\xff", "2031-01-21\n09:00:00Z", strings.Repeat("x", 256)} {
		backend := newOccurrenceBackendFake()
		err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
			scope, err := transaction.ResolveCalendarEventOccurrenceCancelScope(context.Background(), occurrenceActorID)
			if err != nil { return err }
			_, err = transaction.ResolveCalendarEventOccurrenceCancel(context.Background(), scope, calendar.CalendarEventOccurrenceCancelAuthorizationRequest{EventID: occurrenceEventID, RecurrenceID: recurrenceID})
			if !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("direct ingress %q error = %v", recurrenceID, err) }
			return nil
		})
		if err != nil || backend.resolveCalls != 0 { t.Fatalf("direct invalid ingress transaction error=%v resolver calls=%d", err, backend.resolveCalls) }
	}
}

func TestEventServiceCancelCalendarEventOccurrence_MasksNoncanonicalKeyOnlyAfterAuthoritativeResolve(t *testing.T) {
	const recurrenceID = "not-a-canonical-recurrence"
	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceBackendFake)
		masked bool
	}{
		{name: "foreign event masks", mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.EventID = uuid.MustParse("a2323232-2323-4232-8232-232323232323") }, masked: true},
		{name: "identity bound not mutable"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			if testCase.mutate != nil { testCase.mutate(backend) }
			command := mustOccurrenceCommandForKey(t, recurrenceID, 3)
			_, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, command)
			if testCase.masked {
				if !errors.Is(err, calendar.ErrEventNotFound) || err.Error() != calendar.ErrEventNotFound.Error() { t.Fatalf("foreign noncanonical error = %v", err) }
			} else {
				var conflict calendar.CalendarEventOccurrenceCancelConflictError
				if !errors.As(err, &conflict) || !errors.Is(err, calendar.ErrEventConflict) || conflict.Kind() != calendar.CalendarEventOccurrenceCancelNotMutable { t.Fatalf("identity-bound noncanonical error = %v", err) }
			}
			if backend.resolveCalls != 1 || backend.persistCalls != 0 { t.Fatalf("noncanonical resolve=%d persist=%d", backend.resolveCalls, backend.persistCalls) }
		})
	}
}

func TestEventServiceCancelCalendarEventOccurrence_PersistsOneSafeAtomicCancellation(t *testing.T) {
	backend := newOccurrenceBackendFake()
	clock := &occurrenceClock{now: time.Date(2031, time.January, 20, 10, 11, 12, 0, time.UTC)}
	service := calendar.NewEventService(&occurrenceRepository{backend: backend}, clock, &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}, occurrenceSerializerUnused{})
	command := mustOccurrenceCommand(t, 3)

	result, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, command)
	if err != nil {
		t.Fatalf("CancelCalendarEventOccurrence() error = %v", err)
	}
	if result.EventID() != occurrenceEventID || result.RecurrenceID() != occurrenceRecurrenceID || result.Status() != calendar.CalendarEventStatusCancelled || result.Sequence() != 8 || result.EventVersion() != 4 {
		t.Fatalf("result = event:%s recurrence:%q status:%q seq:%d version:%d", result.EventID(), result.RecurrenceID(), result.Status(), result.Sequence(), result.EventVersion())
	}
	if backend.organizationCalls != 1 || backend.resolveCalls != 1 || backend.persistCalls != 1 || len(backend.persisted) != 1 {
		t.Fatalf("calls org:%d resolve:%d persist:%d persisted:%d", backend.organizationCalls, backend.resolveCalls, backend.persistCalls, len(backend.persisted))
	}
	if !reflect.DeepEqual(backend.callOrder, []string{"organization", "resolve", "persist"}) {
		t.Fatalf("backend call order = %#v", backend.callOrder)
	}
	persisted := backend.persisted[0]
	if persisted.ExpectedVersion() != 3 || persisted.Event().EventID() != occurrenceEventID || persisted.Event().Version() != 4 || persisted.Exception().ExceptionID() != occurrenceExceptionID || persisted.Exception().EventID() != occurrenceEventID || persisted.Exception().Sequence() != 8 {
		t.Fatalf("persistence command = %#v", persisted)
	}
	envelope := persisted.Outbox().Envelope()
	if envelope.Subject() != calendar.CalendarInviteCancelledSubjectV2 || envelope.EventID() != occurrenceEnvelopeID {
		t.Fatalf("outbox envelope = subject:%q event:%s", envelope.Subject(), envelope.EventID())
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}
	var decoded struct {
		Subject       string `json:"subject"`
		CorrelationID string `json:"correlation_id"`
		Payload calendar.CalendarInviteDeliveryV2Payload `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(envelope) error = %v", err)
	}
	if decoded.Subject != calendar.CalendarInviteCancelledSubjectV2 || decoded.CorrelationID != occurrenceEventID.String() || decoded.Payload.EventID != occurrenceEventID || decoded.Payload.EventSequence != 8 || decoded.Payload.OrganizerID != occurrenceActorID || decoded.Payload.RecurrenceID == nil || *decoded.Payload.RecurrenceID != occurrenceRecurrenceID || len(decoded.Payload.Deliveries) != 2 {
		t.Fatalf("safe occurrence payload = %#v", decoded)
	}
	decodedPayload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil || !reflect.DeepEqual(decoded.Payload, decodedPayload) || decodedPayload.Deliveries[0].InviteID == uuid.Nil || decodedPayload.Deliveries[1].InviteID == uuid.Nil || decodedPayload.Deliveries[0].InviteID == decodedPayload.Deliveries[1].InviteID {
		t.Fatalf("canonical deterministic v2 payload = %#v err=%v", decodedPayload, err)
	}
	if strings.Contains(string(encoded), "Original title") || strings.Contains(string(encoded), "owner@example.test") || strings.Contains(string(encoded), "BEGIN:VCALENDAR") {
		t.Fatalf("occurrence cancellation envelope leaked content: %s", encoded)
	}

	returned := persisted.Outbox()
	_ = returned
}

func TestEventServiceCancelCalendarEventOccurrence_MasksBeforeHostileSnapshotTraversal(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceBackendFake)
	}{
		{name: "foreign event with huge attendees", mutate: func(backend *occurrenceBackendFake) {
			backend.authorization.Current.EventID = uuid.MustParse("a8888888-8888-4888-8888-888888888888")
			backend.authorization.Attendees = make([]calendar.CalendarEventOccurrenceCancelAttendeeSnapshot, 201)
		}},
		{name: "foreign organizer malformed recurrence", mutate: func(backend *occurrenceBackendFake) {
			other := uuid.MustParse("a9999999-9999-4999-8999-999999999999")
			backend.authorization.Current.OrganizerID = &other
			backend.authorization.Current.RRULE = strings.Repeat("x", 1<<20)
		}},
		{name: "missing identity corrupt exception", mutate: func(backend *occurrenceBackendFake) {
			backend.authorization.Current.EventID = uuid.Nil
			backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{RecurrenceID: "\xff"}
		}},
		{name: "cross organization oversized exdates", mutate: func(backend *occurrenceBackendFake) {
			backend.authorization.Current.OrganizationID = uuid.MustParse("a1212121-1212-4121-8121-121212121212")
			backend.authorization.Current.Exdates = make([]string, 1001)
		}},
		{name: "wrong owner pointer aliases", mutate: func(backend *occurrenceBackendFake) {
			owner := uuid.MustParse("a1313131-1313-4131-8131-131313131313")
			backend.authorization.Current.OwnerID = owner
			backend.authorization.Current.OrganizerID = &owner
			backend.authorization.Attendees[0].UserID = &owner
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			testCase.mutate(backend)
			_, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3))
			if !errors.Is(err, calendar.ErrEventNotFound) || err.Error() != calendar.ErrEventNotFound.Error() || strings.Contains(err.Error(), "a888") || backend.persistCalls != 0 {
				t.Fatalf("masked hostile result error=%v writes=%d", err, backend.persistCalls)
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.CalendarEventCommandInput)
	}{
		{name: "master event id", mutate: func(input *calendar.CalendarEventCommandInput) { input.EventID = uuid.MustParse("a1515151-1515-4151-8151-151515151515") }},
		{name: "master organization", mutate: func(input *calendar.CalendarEventCommandInput) { input.OrganizationID = uuid.MustParse("a1616161-1616-4161-8161-161616161616") }},
		{name: "master owner", mutate: func(input *calendar.CalendarEventCommandInput) { input.OwnerID = uuid.MustParse("a1717171-1717-4171-8171-171717171717") }},
		{name: "master title", mutate: func(input *calendar.CalendarEventCommandInput) { input.Title = "different title" }},
		{name: "master description", mutate: func(input *calendar.CalendarEventCommandInput) { input.Description = "different description" }},
		{name: "master location", mutate: func(input *calendar.CalendarEventCommandInput) { input.Location = "different location" }},
		{name: "master temporal", mutate: func(input *calendar.CalendarEventCommandInput) { input.Temporal.EndsAt = input.Temporal.EndsAt.Add(time.Minute) }},
		{name: "master rrule", mutate: func(input *calendar.CalendarEventCommandInput) { input.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=2" }},
		{name: "master exdates", mutate: func(input *calendar.CalendarEventCommandInput) { input.Exdates = []string{"2031-01-22T09:00:00Z"} }},
		{name: "master sequence", mutate: func(input *calendar.CalendarEventCommandInput) { input.Sequence++ }},
		{name: "master version", mutate: func(input *calendar.CalendarEventCommandInput) { input.Version++ }},
		{name: "master active status", mutate: func(input *calendar.CalendarEventCommandInput) { input.Status = calendar.CalendarEventStatusTentative }},
		{name: "master organizer email", mutate: func(input *calendar.CalendarEventCommandInput) { input.OrganizerEmail = "replacement@example.test" }},
		{name: "master scheduled call", mutate: func(input *calendar.CalendarEventCommandInput) { value := uuid.MustParse("a1818181-1818-4181-8181-181818181818"); input.ScheduledCallID = &value }},
		{name: "master reminder", mutate: func(input *calendar.CalendarEventCommandInput) { value := 15; input.ReminderMinutes = &value }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			backend.resultMutator = func(_ calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
				return mustOccurrenceEchoResult(t, result, testCase.mutate, nil)
			}
			if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
				t.Fatalf("master echo tamper error=%v writes=%d", err, len(backend.persisted))
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.CalendarEventExceptionCommandInput)
	}{
		{name: "exception id", mutate: func(input *calendar.CalendarEventExceptionCommandInput) { input.ExceptionID = uuid.MustParse("a1919191-1919-4191-8191-191919191919") }},
		{name: "exception recurrence", mutate: func(input *calendar.CalendarEventExceptionCommandInput) { input.RecurrenceID = "2031-01-22T09:00:00Z" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			backend.resultMutator = func(_ calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
				return mustOccurrenceEchoResult(t, result, nil, testCase.mutate)
			}
			if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
				t.Fatalf("exception echo tamper error=%v writes=%d", err, len(backend.persisted))
			}
		})
	}
}

func TestEventServiceCancelCalendarEventOccurrence_ConflictsIdempotenceAndRollback(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		expected  uint
		mutate    func(*occurrenceBackendFake)
		want      error
		wantWrite int
		wantClock int
	}{
		{name: "stale version", expected: 2, want: calendar.ErrEventConflict},
		{name: "non hub source", expected: 3, want: calendar.ErrEventConflict, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Source = calendar.CalendarEventSourceMailImport }},
		{name: "exdate is not mutable", expected: 3, want: calendar.ErrEventConflict, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Exdates = []string{occurrenceRecurrenceID} }},
		{name: "already cancelled matching version", expected: 3, wantWrite: 0, wantClock: 0, mutate: func(backend *occurrenceBackendFake) {
			backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceExceptionID, EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, Sequence: 7, Status: calendar.CalendarEventStatusCancelled, Cancelled: true}
		}},
		{name: "result event echo mismatch", expected: 3, want: calendar.ErrEventInvalid, mutate: func(backend *occurrenceBackendFake) {
			backend.resultMutator = func(command calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
				snapshot := result.Event().Snapshot()
				snapshot.Title = "echo mismatch"
				bad, err := calendar.NewCalendarEventCommand(snapshot)
				if err != nil { panic(err) }
				changed, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(bad, result.Exception(), result.Outbox())
				if err != nil { panic(err) }
				return changed
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			if testCase.mutate != nil { testCase.mutate(backend) }
			clock := &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}
			service := calendar.NewEventService(&occurrenceRepository{backend: backend}, clock, &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}, occurrenceSerializerUnused{})
			result, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, testCase.expected))
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) { t.Fatalf("error = %v, want %v", err, testCase.want) }
				if len(backend.persisted) != 0 { t.Fatalf("rollback retained %d commands", len(backend.persisted)) }
				return
			}
			if err != nil { t.Fatalf("idempotent cancellation error = %v", err) }
			if result.EventVersion() != 3 || result.Sequence() != 7 || len(backend.persisted) != testCase.wantWrite || clock.calls != testCase.wantClock {
				t.Fatalf("idempotent result:%#v writes:%d clock:%d", result, len(backend.persisted), clock.calls)
			}
		})
	}
}

func TestEventServiceCancelCalendarEventOccurrence_MutabilityPrecedesStaleVersion(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		key    string
		mutate func(*occurrenceBackendFake)
	}{
		{name: "terminal status", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled }},
		{name: "non hub source", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Source = calendar.CalendarEventSourceMailImport; backend.authorization.Current.UID = "import-" + occurrenceEventID.String() }},
		{name: "non recurring", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.RRULE = "" }},
		{name: "non member", key: "2031-01-30T09:00:00Z"},
		{name: "exdate", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Exdates = []string{occurrenceRecurrenceID} }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			if testCase.mutate != nil { testCase.mutate(backend) }
			clock := &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}
			ids := &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}
			service := calendar.NewEventService(&occurrenceRepository{backend: backend}, clock, ids, occurrenceSerializerUnused{})
			_, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommandForKey(t, testCase.key, 2))
			var conflict calendar.CalendarEventOccurrenceCancelConflictError
			if !errors.As(err, &conflict) || conflict.Kind() != calendar.CalendarEventOccurrenceCancelNotMutable || !errors.Is(err, calendar.ErrEventConflict) || backend.persistCalls != 0 || clock.calls != 0 || ids.calls != 0 {
				t.Fatalf("mutability precedence error=%v kind=%q persist=%d clock=%d ids=%d", err, conflict.Kind(), backend.persistCalls, clock.calls, ids.calls)
			}
			if eventID, present := conflict.EventID(); present || eventID != uuid.Nil {
				t.Fatalf("not-mutable event recovery = %s present=%t", eventID, present)
			}
			if version, present := conflict.CurrentVersion(); present || version != 0 {
				t.Fatalf("not-mutable version recovery = %d present=%t", version, present)
			}
		})
	}
}

func TestEventServiceCancelCalendarEventOccurrence_RecipientProjectionAndFractionalControls(t *testing.T) {
	backend := newOccurrenceBackendFake()
	removedAt := time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)
	removedID := uuid.MustParse("a2424242-2424-4242-8242-242424242424")
	backend.authorization.Attendees = []calendar.CalendarEventOccurrenceCancelAttendeeSnapshot{
		{AttendeeID: occurrenceExternalID, EventID: occurrenceEventID, Status: calendar.RSVPNeedsAction, Version: 1},
		{AttendeeID: removedID, EventID: occurrenceEventID, Status: calendar.RSVPAccepted, Version: 1, RemovedAt: &removedAt},
		{AttendeeID: occurrenceAttendeeID, EventID: occurrenceEventID, Status: calendar.RSVPAccepted, Version: 1},
		{AttendeeID: uuid.MustParse("a2525252-2525-4252-8252-252525252525"), EventID: occurrenceEventID, UserID: uuidRef(occurrenceActorID), Status: calendar.RSVPAccepted, Version: 1},
	}
	if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); err != nil { t.Fatalf("recipient projection error = %v", err) }
	encoded, err := backend.persisted[0].Outbox().Envelope().MarshalJSON()
	if err != nil { t.Fatalf("MarshalJSON() error = %v", err) }
	var payload struct { Payload calendar.CalendarInviteDeliveryV2Payload `json:"payload"` }
	if err := json.Unmarshal(encoded, &payload); err != nil { t.Fatalf("json.Unmarshal() error = %v", err) }
	if len(payload.Payload.Deliveries) != 2 || strings.Contains(string(encoded), removedID.String()) {
		t.Fatalf("recipient projection = %s", encoded)
	}

	duplicate := newOccurrenceBackendFake()
	duplicate.authorization.Attendees = append(duplicate.authorization.Attendees, calendar.CalendarEventOccurrenceCancelAttendeeSnapshot{AttendeeID: occurrenceAttendeeID, EventID: occurrenceEventID, Status: calendar.RSVPAccepted, Version: 1})
	if _, err := newOccurrenceService(duplicate).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || duplicate.persistCalls != 0 {
		t.Fatalf("active duplicate error=%v persist=%d", err, duplicate.persistCalls)
	}

	fractional := newOccurrenceBackendFake()
	if _, err := newOccurrenceService(fractional).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommandForKey(t, "2031-01-21T09:00:00.100Z", 3)); !errors.Is(err, calendar.ErrEventConflict) || fractional.persistCalls != 0 {
		t.Fatalf("fractional recurrence error=%v persist=%d", err, fractional.persistCalls)
	}
	for codepoint := 0; codepoint <= 0x1f; codepoint++ {
		assertOccurrenceControlIngress(t, rune(codepoint))
	}
	for codepoint := 0x7f; codepoint <= 0x9f; codepoint++ {
		assertOccurrenceControlIngress(t, rune(codepoint))
	}
}

func TestCalendarInviteOccurrenceCancelledEnvelopeV1_RejectsUnsafePayload(t *testing.T) {
	actor := occurrenceActorID
	payload := calendar.CalendarInviteOccurrenceCancelledPayload{EventID: occurrenceEventID, UID: "hub-" + occurrenceEventID.String(), Sequence: 8, RecurrenceID: occurrenceRecurrenceID, AttendeeIDs: []uuid.UUID{occurrenceAttendeeID}}
	envelope, err := calendar.NewCalendarInviteOccurrenceCancelledEnvelopeV1(occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, &actor, occurrenceEventID.String(), payload)
	if err != nil { t.Fatalf("NewCalendarInviteOccurrenceCancelledEnvelopeV1() error = %v", err) }
	if envelope.Subject() != "calendar.invite.cancelled.v1" { t.Fatalf("subject = %q", envelope.Subject()) }
	empty := payload
	empty.AttendeeIDs = []uuid.UUID{}
	emptyEnvelope, err := calendar.NewCalendarInviteOccurrenceCancelledEnvelopeV1(occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, &actor, occurrenceEventID.String(), empty)
	if err != nil { t.Fatalf("empty attendee payload error = %v", err) }
	emptyJSON, err := json.Marshal(emptyEnvelope)
	if err != nil || !strings.Contains(string(emptyJSON), "\"attendee_ids\":[]") { t.Fatalf("empty attendee JSON=%s err=%v", emptyJSON, err) }
	for _, change := range []func(*calendar.CalendarInviteOccurrenceCancelledPayload){
		func(value *calendar.CalendarInviteOccurrenceCancelledPayload) { value.Sequence = 0 },
		func(value *calendar.CalendarInviteOccurrenceCancelledPayload) { value.RecurrenceID = "" },
		func(value *calendar.CalendarInviteOccurrenceCancelledPayload) { value.AttendeeIDs = []uuid.UUID{occurrenceAttendeeID, occurrenceAttendeeID} },
		func(value *calendar.CalendarInviteOccurrenceCancelledPayload) { value.AttendeeIDs = make([]uuid.UUID, 201) },
	} {
		candidate := payload
		change(&candidate)
		if _, err := calendar.NewCalendarInviteOccurrenceCancelledEnvelopeV1(occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, &actor, occurrenceEventID.String(), candidate); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("unsafe occurrence envelope error = %v, want invalid", err)
		}
	}
}

func TestCalendarEventOccurrenceCancel_IngressAndAuthoritativePreflight(t *testing.T) {
	for _, testCase := range []struct {
		name string
		key  string
		want error
	}{
		{name: "one byte accepted", key: "x"},
		{name: "255 bytes accepted", key: strings.Repeat("x", 255)},
		{name: "empty", key: "", want: calendar.ErrEventInvalid},
		{name: "256 bytes", key: strings.Repeat("x", 256), want: calendar.ErrEventInvalid},
		{name: "control", key: "2031-01-21\n09:00:00Z", want: calendar.ErrEventInvalid},
		{name: "invalid utf8", key: "\xff", want: calendar.ErrEventInvalid},
		{name: "one mebibyte byte first cap", key: strings.Repeat("x", 1<<20), want: calendar.ErrEventInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command, err := calendar.NewCalendarEventOccurrenceCancelCommand(calendar.CalendarEventOccurrenceCancelInput{EventID: occurrenceEventID, RecurrenceID: testCase.key, ExpectedVersion: 3})
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) { t.Fatalf("command error = %v, want %v", err, testCase.want) }
				backend := newOccurrenceBackendFake()
				repository := &occurrenceRepository{backend: backend}
				_, serviceErr := calendar.NewEventService(repository, &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}, &occurrenceIDs{}, occurrenceSerializerUnused{}).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, command)
				if !errors.Is(serviceErr, calendar.ErrEventInvalid) || repository.calls != 0 || backend.organizationCalls != 0 || backend.resolveCalls != 0 || backend.persistCalls != 0 {
					t.Fatalf("invalid ingress made calls repo:%d org:%d resolve:%d persist:%d err:%v", repository.calls, backend.organizationCalls, backend.resolveCalls, backend.persistCalls, serviceErr)
				}
				return
			}
			if err != nil || command.RecurrenceID() != testCase.key { t.Fatalf("accepted ingress command:%#v err:%v", command, err) }
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceBackendFake)
	}{
		{name: "identity bound oversized exdates", mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Exdates = make([]string, 1001) }},
		{name: "identity bound oversized attendees", mutate: func(backend *occurrenceBackendFake) { backend.authorization.Attendees = make([]calendar.CalendarEventOccurrenceCancelAttendeeSnapshot, 201) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			testCase.mutate(backend)
			service := newOccurrenceService(backend)
			if _, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || backend.persistCalls != 0 {
				t.Fatalf("identity-bound cap error=%v persist=%d", err, backend.persistCalls)
			}
		})
	}
}

func TestCalendarEventOccurrenceCancel_RecurrenceCounterClockAndIDFailures(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		key          string
		mutate       func(*occurrenceBackendFake)
		ids          []uuid.UUID
		clock        time.Time
		want         error
		wantIDCalls  int
		wantClock    int
	}{
		{name: "weekly member", key: "2031-01-27T09:00:00Z", mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.RRULE = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO;COUNT=3" }, ids: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}, clock: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)},
		{name: "monthly member", key: "2031-02-20T09:00:00Z", mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.RRULE = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20;COUNT=3" }, ids: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}, clock: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)},
		{name: "version overflow", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Version = ^uint(0) }, want: calendar.ErrEventInvalid},
		{name: "sequence overflow", key: occurrenceRecurrenceID, mutate: func(backend *occurrenceBackendFake) { backend.authorization.Current.Sequence = ^uint(0) }, want: calendar.ErrEventInvalid},
		{name: "zero exception id", key: occurrenceRecurrenceID, ids: []uuid.UUID{uuid.Nil}, want: calendar.ErrEventInvalid, wantIDCalls: 1},
		{name: "colliding envelope id", key: occurrenceRecurrenceID, ids: []uuid.UUID{occurrenceExceptionID, occurrenceExceptionID}, want: calendar.ErrEventInvalid, wantIDCalls: 2},
		{name: "non utc clock", key: occurrenceRecurrenceID, ids: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}, clock: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.FixedZone("UTC+3", 3*60*60)), want: calendar.ErrEventInvalid, wantIDCalls: 2, wantClock: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			if testCase.mutate != nil { testCase.mutate(backend) }
			clock := &occurrenceClock{now: testCase.clock}
			if clock.now.IsZero() { clock.now = time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC) }
			ids := &occurrenceIDs{values: testCase.ids}
			result, err := calendar.NewEventService(&occurrenceRepository{backend: backend}, clock, ids, occurrenceSerializerUnused{}).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommandForKey(t, testCase.key, backend.authorization.Current.Version))
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) || len(backend.persisted) != 0 || ids.calls != testCase.wantIDCalls || clock.calls != testCase.wantClock {
					t.Fatalf("failure result:%#v err:%v writes:%d ids:%d clock:%d", result, err, len(backend.persisted), ids.calls, clock.calls)
				}
				return
			}
			if err != nil || result.RecurrenceID() != testCase.key || len(backend.persisted) != 1 || ids.calls != 2 || clock.calls != 1 {
				t.Fatalf("recurrence success result:%#v err:%v writes:%d ids:%d clock:%d", result, err, len(backend.persisted), ids.calls, clock.calls)
			}
		})
	}

	backend := newOccurrenceBackendFake()
	backend.resolveError = errors.New("raw resolver failure")
	if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "raw resolver") || backend.persistCalls != 0 {
		t.Fatalf("raw resolver error = %v writes:%d", err, backend.persistCalls)
	}
	backend = newOccurrenceBackendFake()
	backend.persistError = errors.New("raw persist failure")
	if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "raw persist") || len(backend.persisted) != 0 {
		t.Fatalf("raw persist error = %v writes:%d", err, len(backend.persisted))
	}
	backend = newOccurrenceBackendFake()
	zeroClock := &occurrenceClock{}
	if _, err := calendar.NewEventService(&occurrenceRepository{backend: backend}, zeroClock, &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}, occurrenceSerializerUnused{}).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 || zeroClock.calls != 1 {
		t.Fatalf("zero clock error = %v writes:%d clock:%d", err, len(backend.persisted), zeroClock.calls)
	}
}

func TestCalendarEventOccurrenceCancel_DeepCopyAndFullEchoRollback(t *testing.T) {
	backend := newOccurrenceBackendFake()
	authorization := backend.authorization
	returned := authorization.Current
	returned.Exdates = append(returned.Exdates, "2031-01-22T09:00:00Z")
	if len(backend.authorization.Current.Exdates) != 0 { t.Fatal("fixture alias setup invalid") }

	for _, testCase := range []struct {
		name   string
		mutate func(calendar.CalendarEventOccurrenceCancelPersistenceCommand, calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult
	}{
		{name: "event title", mutate: func(command calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
			input := result.Event().Snapshot()
			input.Title = "different"
			event, err := calendar.NewCalendarEventCommand(input); if err != nil { panic(err) }
			changed, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, result.Exception(), result.Outbox()); if err != nil { panic(err) }
			return changed
		}},
		{name: "outbox event id", mutate: func(command calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
			actor := occurrenceActorID
			recurrenceID := occurrenceRecurrenceID
			envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
				calendar.CalendarInviteCancelledSubjectV2,
				uuid.MustParse("abbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
				time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC),
				occurrenceOrganizationID,
				&actor,
				occurrenceEventID.String(),
				calendar.CalendarInviteDeliveryEnvelopeV2Input{
					EventID:       occurrenceEventID,
					EventSequence: 8,
					OrganizerID:   actor,
					RecurrenceID:  &recurrenceID,
					AttendeeIDs:   []uuid.UUID{occurrenceAttendeeID, occurrenceExternalID},
				},
			)
			if err != nil { panic(err) }
			descriptor, err := calendar.NewEventOutboxDescriptor(envelope); if err != nil { panic(err) }
			changed, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(result.Event(), result.Exception(), descriptor); if err != nil { panic(err) }
			return changed
		}},
		{name: "exception recurrence id", mutate: func(_ calendar.CalendarEventOccurrenceCancelPersistenceCommand, result calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult {
			_, exception, descriptor := mustOccurrencePersistencePartsFor(t, "2031-01-22T09:00:00Z", nil)
			changed, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(result.Event(), exception, descriptor)
			if err != nil { panic(err) }
			return changed
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			backend.resultMutator = testCase.mutate
			_, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3))
			if !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
				t.Fatalf("echo mismatch error=%v writes=%d", err, len(backend.persisted))
			}
		})
	}

	backend = newOccurrenceBackendFake()
	backend.returnZeroResult = true
	if _, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
		t.Fatalf("zero persistence result error=%v writes=%d", err, len(backend.persisted))
	}

	input := newOccurrenceBackendFake().authorization.Current
	input.Exdates = []string{"2031-01-22T09:00:00Z"}
	reminder := 10
	scheduledCall := uuid.MustParse("a1111111-1111-4111-8111-111111111112")
	input.ReminderMinutes, input.ScheduledCallID = &reminder, &scheduledCall
	event, err := calendar.NewCalendarEventCommand(input)
	if err != nil { t.Fatalf("NewCalendarEventCommand() error = %v", err) }
	input.Exdates[0], reminder, scheduledCall = "caller-mutated", 20, uuid.Nil
	snapshot := event.Snapshot()
	snapshot.Exdates[0], *snapshot.ReminderMinutes, *snapshot.ScheduledCallID = "returned-mutated", 30, uuid.Nil
	retained := event.Snapshot()
	if retained.Exdates[0] != "2031-01-22T09:00:00Z" || retained.ReminderMinutes == nil || *retained.ReminderMinutes != 10 || retained.ScheduledCallID == nil || *retained.ScheduledCallID == uuid.Nil { t.Fatal("event snapshot exposes an alias") }

	exceptionTemporal := occurrenceTemporal()
	exception, err := calendar.NewCalendarEventExceptionCommand(calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceExceptionID, EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, Temporal: &exceptionTemporal, Sequence: 8, Status: calendar.CalendarEventStatusConfirmed})
	if err != nil { t.Fatalf("NewCalendarEventExceptionCommand() alias error = %v", err) }
	exceptionTemporal.StartsAt = exceptionTemporal.StartsAt.Add(24 * time.Hour)
	returnedException := exception.Snapshot()
	returnedException.Temporal.StartsAt = returnedException.Temporal.StartsAt.Add(48 * time.Hour)
	if exception.Snapshot().Temporal.StartsAt != occurrenceTemporal().StartsAt { t.Fatal("exception temporal exposes an alias") }
}

func TestCalendarEventOccurrenceCancel_RepositoryAndResultBoundariesDetachAliases(t *testing.T) {
	backend := newOccurrenceBackendFake()
	reminder := 5
	scheduledCall := uuid.MustParse("a1414141-1414-4141-8141-141414141414")
	expectedScheduledCall := scheduledCall
	backend.authorization.Current.ReminderMinutes = &reminder
	backend.authorization.Current.ScheduledCallID = &scheduledCall
	ids := &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}
	ids.afterCall = func(call int) {
		if call != 1 { return }
		backend.authorization.Current.Exdates = []string{occurrenceRecurrenceID}
		*backend.authorization.Current.ReminderMinutes = 99
		*backend.authorization.Current.ScheduledCallID = uuid.Nil
		removed := time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)
		backend.authorization.Attendees[1].RemovedAt = &removed
	}
	service := calendar.NewEventService(&occurrenceRepository{backend: backend}, &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}, ids, occurrenceSerializerUnused{})
	if _, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3)); err != nil { t.Fatalf("alias boundary cancellation error = %v", err) }
	if len(backend.persisted) != 1 { t.Fatalf("persisted count = %d", len(backend.persisted)) }
	persisted := backend.persisted[0].Event().Snapshot()
	if len(persisted.Exdates) != 0 || persisted.ReminderMinutes == nil || *persisted.ReminderMinutes != 5 || persisted.ScheduledCallID == nil || *persisted.ScheduledCallID != expectedScheduledCall {
		t.Fatalf("repository alias crossed into command: %#v", persisted)
	}
	encoded, err := backend.persisted[0].Outbox().Envelope().MarshalJSON()
	if err != nil { t.Fatalf("MarshalJSON() error = %v", err) }
	if !strings.Contains(string(encoded), occurrenceExternalID.String()) { t.Fatalf("attendee alias crossed into payload: %s", encoded) }
}

func TestCalendarEventOccurrenceCancel_RejectsInconsistentExceptionsAndMalformedConflicts(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceBackendFake)
	}{
		{name: "cancelled exception with temporal", mutate: func(backend *occurrenceBackendFake) {
			temporal := occurrenceTemporal()
			backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceExceptionID, EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, Sequence: 7, Status: calendar.CalendarEventStatusCancelled, Cancelled: true, Temporal: &temporal}
		}},
		{name: "live exception without temporal", mutate: func(backend *occurrenceBackendFake) {
			backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceExceptionID, EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, Sequence: 7, Status: calendar.CalendarEventStatusConfirmed, Cancelled: false}
		}},
		{name: "zero-value conflict", mutate: func(backend *occurrenceBackendFake) { backend.resolveError = calendar.CalendarEventOccurrenceCancelConflictError{} }},
		{name: "typed-nil conflict", mutate: func(backend *occurrenceBackendFake) {
			var typedNil *calendar.CalendarEventOccurrenceCancelConflictError
			backend.resolveError = typedNil
		}},
		{name: "wrapped typed-nil conflict", mutate: func(backend *occurrenceBackendFake) {
			var typedNil *calendar.CalendarEventOccurrenceCancelConflictError
			backend.resolveError = occurrenceWrappedError{cause: typedNil}
		}},
		{name: "joined typed-nil conflict", mutate: func(backend *occurrenceBackendFake) {
			var typedNil *calendar.CalendarEventOccurrenceCancelConflictError
			backend.resolveError = errors.Join(errors.New("contained"), typedNil)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			testCase.mutate(backend)
			_, err := newOccurrenceService(backend).CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommand(t, 3))
			if !errors.Is(err, calendar.ErrEventInvalid) || backend.persistCalls != 0 {
				t.Fatalf("inconsistent exception/conflict was not fail-closed; persist=%d", backend.persistCalls)
			}
		})
	}
}

func TestCalendarEventOccurrenceCancelPersistence_RejectsUnsafeDirectCommands(t *testing.T) {
	event, exception, descriptor := mustOccurrencePersistenceParts(t)
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, descriptor); err != nil {
		t.Fatalf("NewCalendarEventOccurrenceCancelPersistenceCommand() error = %v", err)
	}
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, exception, descriptor); err != nil {
		t.Fatalf("NewCalendarEventOccurrenceCancelPersistenceResult() error = %v", err)
	}

	badExceptionInput := exception.Snapshot()
	badExceptionInput.Sequence = 7
	badException, err := calendar.NewCalendarEventExceptionCommand(badExceptionInput)
	if err != nil { t.Fatalf("NewCalendarEventExceptionCommand() error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, badException, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("bad exception sequence constructor error = %v", err)
	}
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, badException, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("bad exception sequence result error = %v", err)
	}
	badExceptionInput = exception.Snapshot()
	temporal := occurrenceTemporal()
	badExceptionInput.Temporal = &temporal
	badException, err = calendar.NewCalendarEventExceptionCommand(badExceptionInput)
	if err != nil { t.Fatalf("NewCalendarEventExceptionCommand() temporal error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, badException, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("temporal cancelled exception result error = %v", err)
	}

	generic, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, uuidRef(occurrenceActorID), occurrenceEventID.String(), map[string]any{"event_id": occurrenceEventID.String()})
	if err != nil { t.Fatalf("events.NewEnvelopeV1() error = %v", err) }
	unsafeDescriptor, err := calendar.NewEventOutboxDescriptor(generic)
	if err != nil { t.Fatalf("NewEventOutboxDescriptor() error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, unsafeDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("unsafe outbox constructor error = %v", err)
	}

	unknownPayload, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, uuidRef(occurrenceActorID), occurrenceEventID.String(), map[string]any{"event_id": occurrenceEventID.String(), "uid": "hub-" + occurrenceEventID.String(), "sequence": 8, "recurrence_id": occurrenceRecurrenceID, "attendee_ids": []string{occurrenceAttendeeID.String(), occurrenceExternalID.String()}, "title": "private", "raw_itip": "BEGIN:VCALENDAR"})
	if err != nil { t.Fatalf("events.NewEnvelopeV1() unknown payload error = %v", err) }
	unknownDescriptor, err := calendar.NewEventOutboxDescriptor(unknownPayload)
	if err != nil { t.Fatalf("NewEventOutboxDescriptor() unknown payload error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, unknownDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("unknown payload constructor error = %v", err)
	}
	duplicatePayload := json.RawMessage(`{"event_id":"a3333333-3333-4333-8333-333333333333","uid":"hub-a3333333-3333-4333-8333-333333333333","sequence":8,"sequence":8,"recurrence_id":"2031-01-21T09:00:00Z","attendee_ids":["a6666666-6666-4666-8666-666666666666","a7777777-7777-4777-8777-777777777777"]}`)
	duplicateEnvelope, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, uuidRef(occurrenceActorID), occurrenceEventID.String(), duplicatePayload)
	if err != nil { t.Fatalf("events.NewEnvelopeV1() duplicate payload error = %v", err) }
	duplicateDescriptor, err := calendar.NewEventOutboxDescriptor(duplicateEnvelope)
	if err != nil { t.Fatalf("NewEventOutboxDescriptor() duplicate payload error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, duplicateDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("duplicate known payload constructor error = %v", err)
	}
	mixedCasePayload := json.RawMessage(`{"event_id":"a3333333-3333-4333-8333-333333333333","uid":"hub-a3333333-3333-4333-8333-333333333333","UID":"private-case-variant","sequence":8,"recurrence_id":"2031-01-21T09:00:00Z","attendee_ids":["a6666666-6666-4666-8666-666666666666","a7777777-7777-4777-8777-777777777777"]}`)
	mixedCaseEnvelope, err := events.NewEnvelopeV1("calendar.invite.cancelled.v1", occurrenceEnvelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, uuidRef(occurrenceActorID), occurrenceEventID.String(), mixedCasePayload)
	if err != nil { t.Fatalf("events.NewEnvelopeV1() mixed-case payload error = %v", err) }
	mixedCaseDescriptor, err := calendar.NewEventOutboxDescriptor(mixedCaseEnvelope)
	if err != nil { t.Fatalf("NewEventOutboxDescriptor() mixed-case payload error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, mixedCaseDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("mixed-case private payload constructor error = %v", err)
	}
	for _, testCase := range []struct {
		name          string
		subject       string
		envelopeID    uuid.UUID
		correlationID string
		payload       map[string]any
	}{
		{name: "subject", subject: "calendar.private.cancelled.v1", envelopeID: occurrenceEnvelopeID, correlationID: occurrenceEventID.String(), payload: occurrenceSafePayloadMap(occurrenceRecurrenceID)},
		{name: "envelope event id", subject: "calendar.invite.cancelled.v1", envelopeID: occurrenceEventID, correlationID: occurrenceEventID.String(), payload: occurrenceSafePayloadMap(occurrenceRecurrenceID)},
		{name: "correlation", subject: "calendar.invite.cancelled.v1", envelopeID: occurrenceEnvelopeID, correlationID: "private-correlation", payload: occurrenceSafePayloadMap(occurrenceRecurrenceID)},
		{name: "payload sequence", subject: "calendar.invite.cancelled.v1", envelopeID: occurrenceEnvelopeID, correlationID: occurrenceEventID.String(), payload: map[string]any{"event_id": occurrenceEventID.String(), "uid": "hub-" + occurrenceEventID.String(), "sequence": 9, "recurrence_id": occurrenceRecurrenceID, "attendee_ids": []string{occurrenceAttendeeID.String(), occurrenceExternalID.String()}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			envelope, err := events.NewEnvelopeV1(testCase.subject, testCase.envelopeID, time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC), occurrenceOrganizationID, uuidRef(occurrenceActorID), testCase.correlationID, testCase.payload)
			if err != nil { t.Fatalf("events.NewEnvelopeV1() error = %v", err) }
			descriptor, err := calendar.NewEventOutboxDescriptor(envelope)
			if err != nil { t.Fatalf("NewEventOutboxDescriptor() error = %v", err) }
			if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(event, 3, exception, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("direct envelope tamper error = %v", err)
			}
		})
	}

	nonMemberEvent, nonMemberException, nonMemberDescriptor := mustOccurrencePersistencePartsFor(t, "2031-01-30T09:00:00Z", nil)
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(nonMemberEvent, 3, nonMemberException, nonMemberDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("non-member constructor error = %v", err)
	}
	exdateEvent, exdateException, exdateDescriptor := mustOccurrencePersistencePartsFor(t, occurrenceRecurrenceID, []string{occurrenceRecurrenceID})
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(exdateEvent, 3, exdateException, exdateDescriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("exdate constructor error = %v", err)
	}
	nonRecurringInput := event.Snapshot()
	nonRecurringInput.RRULE = ""
	nonRecurring, err := calendar.NewCalendarEventCommand(nonRecurringInput)
	if err != nil { t.Fatalf("NewCalendarEventCommand(nonrecurring) error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(nonRecurring, 3, exception, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("non-recurring constructor error = %v", err)
	}
	allDayInput := event.Snapshot()
	allDayInput.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: "2031-01-20", EndDateExclusive: "2031-01-21", Timezone: "UTC"}
	allDayInput.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=3"
	allDay, err := calendar.NewCalendarEventCommand(allDayInput)
	if err != nil { t.Fatalf("NewCalendarEventCommand(all day) error = %v", err) }
	if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceCommand(allDay, 3, exception, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("wrong temporal recurrence constructor error = %v", err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.CalendarEventExceptionCommandInput)
	}{
		{name: "exception event", mutate: func(input *calendar.CalendarEventExceptionCommandInput) { input.EventID = uuid.MustParse("a2121212-2121-4212-8212-212121212121") }},
		{name: "exception cancelled tuple", mutate: func(input *calendar.CalendarEventExceptionCommandInput) { input.Cancelled = false }},
		{name: "exception live status", mutate: func(input *calendar.CalendarEventExceptionCommandInput) { input.Status, input.Cancelled = calendar.CalendarEventStatusTentative, false }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := exception.Snapshot()
			testCase.mutate(&input)
			changed, err := calendar.NewCalendarEventExceptionCommand(input)
			if err != nil { t.Fatalf("NewCalendarEventExceptionCommand() error = %v", err) }
			if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, changed, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("invalid exception persistence result error = %v", err)
			}
		})
	}
	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.CalendarEventCommandInput)
	}{
		{name: "master uid", mutate: func(input *calendar.CalendarEventCommandInput) { input.UID = "hub-a9999999-9999-4999-8999-999999999999" }},
		{name: "master source", mutate: func(input *calendar.CalendarEventCommandInput) { input.Source = calendar.CalendarEventSourceMailImport }},
		{name: "master organizer identity", mutate: func(input *calendar.CalendarEventCommandInput) { value := uuid.MustParse("a2020202-2020-4202-8202-202020202020"); input.OrganizerID = &value }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := event.Snapshot()
			testCase.mutate(&input)
			changed, err := calendar.NewCalendarEventCommand(input)
			if err != nil { t.Fatalf("NewCalendarEventCommand() error = %v", err) }
			if _, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(changed, exception, descriptor); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("invalid master persistence result error = %v", err)
			}
		})
	}
}

func TestCalendarEventOccurrenceCancel_FrozenABIAndRecurrenceBoundaries(t *testing.T) {
	inputType := reflect.TypeOf(calendar.CalendarEventOccurrenceCancelInput{})
	if inputType.NumField() != 3 || inputType.Field(0).Name != "EventID" || inputType.Field(1).Name != "RecurrenceID" || inputType.Field(2).Name != "ExpectedVersion" {
		t.Fatalf("input ABI = %#v", inputType)
	}
	payloadType := reflect.TypeOf(calendar.CalendarInviteOccurrenceCancelledPayload{})
	if payloadType.NumField() != 5 || payloadType.Field(3).Tag.Get("json") != "recurrence_id" || payloadType.Field(4).Tag.Get("json") != "attendee_ids" {
		t.Fatalf("payload ABI = %#v", payloadType)
	}
	portType := reflect.TypeOf((*calendar.CalendarEventOccurrenceCancelTransactionBackend)(nil)).Elem()
	if portType.NumMethod() != 2 || portType.Method(0).Name != "PersistCalendarEventOccurrenceCancel" || portType.Method(1).Name != "ResolveCalendarEventOccurrenceCancel" {
		t.Fatalf("optional port ABI = %#v", portType)
	}
	if reflect.TypeOf(calendar.EventMutationCalendarOccurrenceCancel).Kind() != reflect.String || string(calendar.EventMutationCalendarOccurrenceCancel) != "event_occurrence_cancel" {
		t.Fatalf("receipt-free operation = %q", calendar.EventMutationCalendarOccurrenceCancel)
	}
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelAuthorizationRequest{}), []string{"EventID", "RecurrenceID"}, []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf("")})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelAttendeeSnapshot{}), []string{"AttendeeID", "EventID", "UserID", "Status", "Version", "RemovedAt"}, []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(uuid.UUID{}), reflect.TypeOf((*uuid.UUID)(nil)), reflect.TypeOf(calendar.RSVPStatus("")), reflect.TypeOf(uint(0)), reflect.TypeOf((*time.Time)(nil))})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelAuthorization{}), []string{"Current", "Attendees", "Exception"}, []reflect.Type{reflect.TypeOf(calendar.CalendarEventCommandInput{}), reflect.TypeOf([]calendar.CalendarEventOccurrenceCancelAttendeeSnapshot{}), reflect.TypeOf((*calendar.CalendarEventExceptionCommandInput)(nil))})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelCommand{}), []string{"input"}, []reflect.Type{reflect.TypeOf(calendar.CalendarEventOccurrenceCancelInput{})})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceCommand{}), []string{"event", "expectedVersion", "exception", "outbox"}, []reflect.Type{reflect.TypeOf(calendar.CalendarEventCommand{}), reflect.TypeOf(uint(0)), reflect.TypeOf(calendar.CalendarEventExceptionCommand{}), reflect.TypeOf(calendar.EventOutboxDescriptor{})})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceResult{}), []string{"event", "exception", "outbox"}, []reflect.Type{reflect.TypeOf(calendar.CalendarEventCommand{}), reflect.TypeOf(calendar.CalendarEventExceptionCommand{}), reflect.TypeOf(calendar.EventOutboxDescriptor{})})
	assertOccurrenceStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), []string{"eventID", "recurrenceID", "status", "sequence", "eventVersion"}, []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(""), reflect.TypeOf(calendar.CalendarEventStatus("")), reflect.TypeOf(uint(0)), reflect.TypeOf(uint(0))})
	assertOccurrenceMethodABI(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelCommand), 1, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelPersistenceCommand), 4, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelPersistenceResult), 3, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelResult), 5, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf((*calendar.EventService).CancelCalendarEventOccurrence), 4, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf((*calendar.EventTransaction).ResolveCalendarEventOccurrenceCancelScope), 3, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf((*calendar.EventTransaction).ResolveCalendarEventOccurrenceCancel), 4, 2)
	assertOccurrenceMethodABI(t, reflect.TypeOf((*calendar.EventTransaction).PersistCalendarEventOccurrenceCancel), 4, 2)
	assertOccurrenceExactType(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelCommand), reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelInput) (calendar.CalendarEventOccurrenceCancelCommand, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelPersistenceCommand), reflect.TypeOf(func(calendar.CalendarEventCommand, uint, calendar.CalendarEventExceptionCommand, calendar.EventOutboxDescriptor) (calendar.CalendarEventOccurrenceCancelPersistenceCommand, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelPersistenceResult), reflect.TypeOf(func(calendar.CalendarEventCommand, calendar.CalendarEventExceptionCommand, calendar.EventOutboxDescriptor) (calendar.CalendarEventOccurrenceCancelPersistenceResult, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceCancelResult), reflect.TypeOf(func(uuid.UUID, string, calendar.CalendarEventStatus, uint, uint) (calendar.CalendarEventOccurrenceCancelResult, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf((*calendar.EventService).CancelCalendarEventOccurrence), reflect.TypeOf(func(*calendar.EventService, context.Context, uuid.UUID, calendar.CalendarEventOccurrenceCancelCommand) (calendar.CalendarEventOccurrenceCancelResult, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf((*calendar.EventTransaction).ResolveCalendarEventOccurrenceCancelScope), reflect.TypeOf(func(*calendar.EventTransaction, context.Context, uuid.UUID) (calendar.EventMutationScope, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf((*calendar.EventTransaction).ResolveCalendarEventOccurrenceCancel), reflect.TypeOf(func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceCancelAuthorizationRequest) (calendar.CalendarEventOccurrenceCancelAuthorization, error) { panic("oracle") }))
	assertOccurrenceExactType(t, reflect.TypeOf((*calendar.EventTransaction).PersistCalendarEventOccurrenceCancel), reflect.TypeOf(func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceCancelPersistenceCommand) (calendar.CalendarEventOccurrenceCancelPersistenceResult, error) { panic("oracle") }))
	resolveMethod, resolved := portType.MethodByName("ResolveCalendarEventOccurrenceCancel")
	if !resolved { t.Fatal("occurrence resolver method missing") }
	assertOccurrenceExactType(t, resolveMethod.Type, reflect.TypeOf(func(context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceCancelAuthorizationRequest) (calendar.CalendarEventOccurrenceCancelAuthorization, error) { panic("oracle") }))
	persistMethod, persisted := portType.MethodByName("PersistCalendarEventOccurrenceCancel")
	if !persisted { t.Fatal("occurrence persistence method missing") }
	assertOccurrenceExactType(t, persistMethod.Type, reflect.TypeOf(func(context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceCancelPersistenceCommand) (calendar.CalendarEventOccurrenceCancelPersistenceResult, error) { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelCommand{}), "EventID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelCommand) uuid.UUID { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelCommand{}), "RecurrenceID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelCommand) string { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelCommand{}), "ExpectedVersion", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelCommand) uint { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceCommand{}), "Event", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceCommand) calendar.CalendarEventCommand { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceCommand{}), "Exception", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceCommand) calendar.CalendarEventExceptionCommand { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceCommand{}), "Outbox", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceCommand) calendar.EventOutboxDescriptor { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceResult{}), "Event", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventCommand { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceResult{}), "Exception", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventExceptionCommand { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelPersistenceResult{}), "Outbox", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.EventOutboxDescriptor { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), "EventID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelResult) uuid.UUID { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), "RecurrenceID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelResult) string { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), "Status", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelResult) calendar.CalendarEventStatus { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), "Sequence", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelResult) uint { panic("oracle") }))
	assertOccurrenceValueMethodType(t, reflect.TypeOf(calendar.CalendarEventOccurrenceCancelResult{}), "EventVersion", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelResult) uint { panic("oracle") }))
	conflictType := reflect.TypeOf(calendar.CalendarEventOccurrenceCancelConflictError{})
	assertOccurrenceValueMethodType(t, conflictType, "Error", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelConflictError) string { panic("oracle") }))
	assertOccurrenceValueMethodType(t, conflictType, "Unwrap", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelConflictError) error { panic("oracle") }))
	assertOccurrenceValueMethodType(t, conflictType, "Kind", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelConflictError) calendar.CalendarEventOccurrenceCancelConflictKind { panic("oracle") }))
	assertOccurrenceValueMethodType(t, conflictType, "EventID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelConflictError) (uuid.UUID, bool) { panic("oracle") }))
	assertOccurrenceValueMethodType(t, conflictType, "CurrentVersion", reflect.TypeOf(func(calendar.CalendarEventOccurrenceCancelConflictError) (uint, bool) { panic("oracle") }))

	for _, testCase := range []struct {
		name      string
		configure func(*occurrenceBackendFake) string
		want      error
	}{
		{name: "timed until inclusive", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;UNTIL=20310122T090000Z"
			return "2031-01-22T09:00:00Z"
		}},
		{name: "all day member", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: "2031-01-20", EndDateExclusive: "2031-01-21", Timezone: "UTC"}
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=3"
			return "2031-01-21"
		}},
		{name: "all day weekly dtstart counted once", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: "2031-01-20", EndDateExclusive: "2031-01-21", Timezone: "UTC"}
			backend.authorization.Current.RRULE = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO;COUNT=2"
			return "2031-01-27"
		}},
		{name: "all day monthly dtstart counted once", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: "2031-01-20", EndDateExclusive: "2031-01-21", Timezone: "UTC"}
			backend.authorization.Current.RRULE = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20;COUNT=2"
			return "2031-02-20"
		}},
		{name: "count one thousand inclusive", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=1000"
			return occurrenceTemporal().StartsAt.AddDate(0, 0, 999).Format("2006-01-02T15:04:05Z")
		}},
		{name: "count one thousand one corrupt current", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=1001"
			return occurrenceRecurrenceID
		}, want: calendar.ErrEventInvalid},
		{name: "all day monthly invalid month stops at until", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: "2031-02-28", EndDateExclusive: "2031-03-01", Timezone: "UTC"}
			backend.authorization.Current.RRULE = "FREQ=MONTHLY;INTERVAL=12;BYMONTHDAY=31;UNTIL=20320228"
			return "2032-02-28"
		}, want: calendar.ErrEventConflict},
		{name: "multi_byday_unbounded_rule_rejected", configure: func(backend *occurrenceBackendFake) string {
			backend.authorization.Current.RRULE = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU"
			return "2050-01-01T09:00:00Z"
		}, want: calendar.ErrEventInvalid},
		{name: "dst gap skipped", configure: func(backend *occurrenceBackendFake) string {
			start := time.Date(2031, time.March, 29, 1, 30, 0, 0, time.UTC)
			backend.authorization.Current.Temporal = calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Hour), Timezone: "Europe/Berlin"}
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=3"
			return "2031-03-31T00:30:00Z"
		}},
		{name: "dst fold earliest", configure: func(backend *occurrenceBackendFake) string {
			start := time.Date(2031, time.October, 25, 0, 30, 0, 0, time.UTC)
			backend.authorization.Current.Temporal = calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Hour), Timezone: "Europe/Berlin"}
			backend.authorization.Current.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=3"
			return "2031-10-26T00:30:00Z"
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceBackendFake()
			key := testCase.configure(backend)
			service := calendar.NewEventService(&occurrenceRepository{backend: backend}, &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}, &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}, occurrenceSerializerUnused{})
			_, err := service.CancelCalendarEventOccurrence(context.Background(), occurrenceActorID, mustOccurrenceCommandForKey(t, key, 3))
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) || backend.persistCalls != 0 { t.Fatalf("recurrence bound error=%v persist=%d", err, backend.persistCalls) }
				return
			}
			if err != nil || backend.persistCalls != 1 { t.Fatalf("recurrence member error=%v persist=%d", err, backend.persistCalls) }
		})
	}
}

func mustOccurrenceCommand(t *testing.T, version uint) calendar.CalendarEventOccurrenceCancelCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventOccurrenceCancelCommand(calendar.CalendarEventOccurrenceCancelInput{EventID: occurrenceEventID, RecurrenceID: occurrenceRecurrenceID, ExpectedVersion: version})
	if err != nil { t.Fatalf("NewCalendarEventOccurrenceCancelCommand() error = %v", err) }
	return command
}

func mustOccurrenceCommandForKey(t *testing.T, key string, version uint) calendar.CalendarEventOccurrenceCancelCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventOccurrenceCancelCommand(calendar.CalendarEventOccurrenceCancelInput{EventID: occurrenceEventID, RecurrenceID: key, ExpectedVersion: version})
	if err != nil { t.Fatalf("NewCalendarEventOccurrenceCancelCommand() error = %v", err) }
	return command
}

func mustOccurrencePersistenceParts(t *testing.T) (calendar.CalendarEventCommand, calendar.CalendarEventExceptionCommand, calendar.EventOutboxDescriptor) {
	return mustOccurrencePersistencePartsFor(t, occurrenceRecurrenceID, nil)
}

func mustOccurrencePersistencePartsFor(t *testing.T, recurrenceID string, exdates []string) (calendar.CalendarEventCommand, calendar.CalendarEventExceptionCommand, calendar.EventOutboxDescriptor) {
	t.Helper()
	input := newOccurrenceBackendFake().authorization.Current
	input.Version, input.Sequence = 4, 8
	input.Exdates = append([]string(nil), exdates...)
	event, err := calendar.NewCalendarEventCommand(input)
	if err != nil { t.Fatalf("NewCalendarEventCommand() error = %v", err) }
	exception, err := calendar.NewCalendarEventExceptionCommand(calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceExceptionID, EventID: occurrenceEventID, RecurrenceID: recurrenceID, Sequence: 8, Status: calendar.CalendarEventStatusCancelled, Cancelled: true})
	if err != nil { t.Fatalf("NewCalendarEventExceptionCommand() error = %v", err) }
	actor := occurrenceActorID
	envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteCancelledSubjectV2,
		occurrenceEnvelopeID,
		time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC),
		occurrenceOrganizationID,
		&actor,
		occurrenceEventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{
			EventID:       occurrenceEventID,
			EventSequence: 8,
			OrganizerID:   actor,
			RecurrenceID:  &recurrenceID,
			AttendeeIDs:   []uuid.UUID{occurrenceAttendeeID, occurrenceExternalID},
		},
	)
	if err != nil { t.Fatalf("NewCalendarInviteOccurrenceCancelledEnvelopeV1() error = %v", err) }
	descriptor, err := calendar.NewEventOutboxDescriptor(envelope)
	if err != nil { t.Fatalf("NewEventOutboxDescriptor() error = %v", err) }
	return event, exception, descriptor
}

func mustOccurrenceEchoResult(t *testing.T, original calendar.CalendarEventOccurrenceCancelPersistenceResult, mutateEvent func(*calendar.CalendarEventCommandInput), mutateException func(*calendar.CalendarEventExceptionCommandInput)) calendar.CalendarEventOccurrenceCancelPersistenceResult {
	t.Helper()
	eventInput := original.Event().Snapshot()
	if mutateEvent != nil { mutateEvent(&eventInput) }
	eventInput.UID = "hub-" + eventInput.EventID.String()
	eventInput.OrganizerID = uuidRef(eventInput.OwnerID)
	exceptionInput := original.Exception().Snapshot()
	if mutateException != nil { mutateException(&exceptionInput) }
	exceptionInput.EventID = eventInput.EventID
	exceptionInput.Sequence = eventInput.Sequence
	event, err := calendar.NewCalendarEventCommand(eventInput)
	if err != nil { panic(err) }
	exception, err := calendar.NewCalendarEventExceptionCommand(exceptionInput)
	if err != nil { panic(err) }
	actor := eventInput.OwnerID
	recurrenceID := exceptionInput.RecurrenceID
	envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteCancelledSubjectV2,
		occurrenceEnvelopeID,
		time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC),
		eventInput.OrganizationID,
		&actor,
		eventInput.EventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{
			EventID:       eventInput.EventID,
			EventSequence: eventInput.Sequence,
			OrganizerID:   actor,
			RecurrenceID:  &recurrenceID,
			AttendeeIDs:   []uuid.UUID{occurrenceAttendeeID, occurrenceExternalID},
		},
	)
	if err != nil { panic(err) }
	descriptor, err := calendar.NewEventOutboxDescriptor(envelope)
	if err != nil { panic(err) }
	result, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(event, exception, descriptor)
	if err != nil { panic(err) }
	return result
}

func occurrenceSafePayloadMap(recurrenceID string) map[string]any {
	return map[string]any{"event_id": occurrenceEventID.String(), "uid": "hub-" + occurrenceEventID.String(), "sequence": 8, "recurrence_id": recurrenceID, "attendee_ids": []string{occurrenceAttendeeID.String(), occurrenceExternalID.String()}}
}

func assertOccurrenceControlIngress(t *testing.T, value rune) {
	t.Helper()
	key := "2031-01-21T09:00:00Z" + string(value)
	if _, err := calendar.NewCalendarEventOccurrenceCancelCommand(calendar.CalendarEventOccurrenceCancelInput{EventID: occurrenceEventID, RecurrenceID: key, ExpectedVersion: 3}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("control U+%04X ingress error = %v", value, err)
	}
}

func occurrenceTemporal() calendar.CalendarEvent {
	return calendar.CalendarEvent{StartsAt: time.Date(2031, time.January, 20, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, time.January, 20, 10, 0, 0, 0, time.UTC), Timezone: "UTC"}
}

func uuidRef(value uuid.UUID) *uuid.UUID { return &value }

func newOccurrenceService(backend *occurrenceBackendFake) *calendar.EventService {
	return calendar.NewEventService(&occurrenceRepository{backend: backend}, &occurrenceClock{now: time.Date(2031, time.January, 20, 1, 2, 3, 0, time.UTC)}, &occurrenceIDs{values: []uuid.UUID{occurrenceExceptionID, occurrenceEnvelopeID}}, occurrenceSerializerUnused{})
}

type occurrenceRepository struct { backend *occurrenceBackendFake; calls int }

func (repository *occurrenceRepository) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	repository.calls++
	snapshot := append([]calendar.CalendarEventOccurrenceCancelPersistenceCommand(nil), repository.backend.persisted...)
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil { repository.backend.persisted = snapshot }
	return err
}

type occurrenceBackendFake struct {
	callOrder []string
	organizationID uuid.UUID
	authorization calendar.CalendarEventOccurrenceCancelAuthorization
	resolveError error
	persistError error
	resultMutator func(calendar.CalendarEventOccurrenceCancelPersistenceCommand, calendar.CalendarEventOccurrenceCancelPersistenceResult) calendar.CalendarEventOccurrenceCancelPersistenceResult
	returnZeroResult bool
	lastRequest calendar.CalendarEventOccurrenceCancelAuthorizationRequest
	persisted []calendar.CalendarEventOccurrenceCancelPersistenceCommand
	organizationCalls int
	resolveCalls int
	persistCalls int
}

func newOccurrenceBackendFake() *occurrenceBackendFake {
	return &occurrenceBackendFake{organizationID: occurrenceOrganizationID, authorization: calendar.CalendarEventOccurrenceCancelAuthorization{
		Current: calendar.CalendarEventCommandInput{EventID: occurrenceEventID, OrganizationID: occurrenceOrganizationID, OwnerID: occurrenceActorID, UID: "hub-" + occurrenceEventID.String(), Title: "Original title", Description: "Private description", Location: "Secret room", Temporal: occurrenceTemporal(), RRULE: "FREQ=DAILY;INTERVAL=1;COUNT=3", Sequence: 7, Version: 3, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed, OrganizerID: uuidRef(occurrenceActorID), OrganizerEmail: "owner@example.test"},
		Attendees: []calendar.CalendarEventOccurrenceCancelAttendeeSnapshot{
			{AttendeeID: occurrenceAttendeeID, EventID: occurrenceEventID, Status: calendar.RSVPAccepted, Version: 1},
			{AttendeeID: occurrenceExternalID, EventID: occurrenceEventID, Status: calendar.RSVPNeedsAction, Version: 1},
			{AttendeeID: uuid.MustParse("a0101010-1010-4010-8010-101010101010"), EventID: occurrenceEventID, UserID: uuidRef(occurrenceActorID), Status: calendar.RSVPAccepted, Version: 1},
		},
	}}
}

func (backend *occurrenceBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	backend.callOrder = append(backend.callOrder, "organization")
	backend.organizationCalls++
	if actorID != occurrenceActorID { return uuid.Nil, calendar.ErrEventForbidden }
	return backend.organizationID, nil
}

func (backend *occurrenceBackendFake) ResolveCalendarEventOccurrenceCancel(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventOccurrenceCancelAuthorizationRequest) (calendar.CalendarEventOccurrenceCancelAuthorization, error) {
	backend.callOrder = append(backend.callOrder, "resolve")
	backend.resolveCalls++
	backend.lastRequest = request
	if backend.resolveError != nil { return calendar.CalendarEventOccurrenceCancelAuthorization{}, backend.resolveError }
	return backend.authorization, nil
}

func (backend *occurrenceBackendFake) PersistCalendarEventOccurrenceCancel(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventOccurrenceCancelPersistenceCommand) (calendar.CalendarEventOccurrenceCancelPersistenceResult, error) {
	backend.callOrder = append(backend.callOrder, "persist")
	backend.persistCalls++
	if backend.persistError != nil { return calendar.CalendarEventOccurrenceCancelPersistenceResult{}, backend.persistError }
	if backend.returnZeroResult { return calendar.CalendarEventOccurrenceCancelPersistenceResult{}, nil }
	result, err := calendar.NewCalendarEventOccurrenceCancelPersistenceResult(command.Event(), command.Exception(), command.Outbox())
	if err != nil { return calendar.CalendarEventOccurrenceCancelPersistenceResult{}, err }
	if backend.resultMutator != nil { result = backend.resultMutator(command, result) }
	backend.persisted = append(backend.persisted, command)
	return result, nil
}

func (backend *occurrenceBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) { return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error { return calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { return calendar.ErrEventInvalid }

type occurrenceBaseOnlyBackend struct { organizationCalls int; otherCalls int }
func (backend *occurrenceBaseOnlyBackend) ResolveActorOrganization(context.Context, uuid.UUID) (uuid.UUID, error) { backend.organizationCalls++; return occurrenceOrganizationID, nil }
func (backend *occurrenceBaseOnlyBackend) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { backend.otherCalls++; return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) { backend.otherCalls++; return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error { backend.otherCalls++; return calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceBaseOnlyBackend) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { backend.otherCalls++; return calendar.ErrEventInvalid }

type occurrenceClock struct { now time.Time; calls int }
func (clock *occurrenceClock) Now() time.Time { clock.calls++; return clock.now }

type occurrenceIDs struct { values []uuid.UUID; calls int; afterCall func(int) }
func (ids *occurrenceIDs) NewEventID() uuid.UUID {
	if ids.calls >= len(ids.values) { return uuid.Nil }
	value := ids.values[ids.calls]
	ids.calls++
	if ids.afterCall != nil { ids.afterCall(ids.calls) }
	return value
}

type occurrenceSerializerUnused struct{}
func (occurrenceSerializerUnused) Serialize(calendar.ITIPMessage) ([]byte, error) { return nil, errors.New("serializer must not be called") }

type occurrenceWrappedError struct { cause error }
func (err occurrenceWrappedError) Error() string { return "wrapped occurrence failure" }
func (err occurrenceWrappedError) Unwrap() error { return err.cause }

func assertOccurrenceStructABI(t *testing.T, value reflect.Type, names []string, types []reflect.Type) {
	t.Helper()
	if value.NumField() != len(names) || len(names) != len(types) { t.Fatalf("%s field count=%d", value, value.NumField()) }
	for index := range names {
		field := value.Field(index)
		if field.Name != names[index] || field.Type != types[index] { t.Fatalf("%s field %d = %s %s", value, index, field.Name, field.Type) }
	}
}

func assertOccurrenceMethodABI(t *testing.T, value reflect.Type, inputs, outputs int) {
	t.Helper()
	if value.NumIn() != inputs || value.NumOut() != outputs { t.Fatalf("method type %s inputs=%d outputs=%d", value, value.NumIn(), value.NumOut()) }
}

func assertOccurrenceExactType(t *testing.T, actual, expected reflect.Type) {
	t.Helper()
	if actual != expected { t.Fatalf("ABI type = %s, want %s", actual, expected) }
}

func assertOccurrenceValueMethodType(t *testing.T, receiver reflect.Type, name string, expected reflect.Type) {
	t.Helper()
	method, found := receiver.MethodByName(name)
	if !found { t.Fatalf("%s method %s missing", receiver, name) }
	assertOccurrenceExactType(t, method.Type, expected)
}
