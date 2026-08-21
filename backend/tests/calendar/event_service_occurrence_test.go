package calendar_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	occurrenceRSVPActorID      = uuid.MustParse("d1111111-1111-4111-8111-111111111111")
	occurrenceRSVPOrganization = uuid.MustParse("d2222222-2222-4222-822222222222")
	occurrenceRSVPEventID      = uuid.MustParse("d3333333-3333-4333-8333-333333333333")
	occurrenceRSVPSeriesID     = uuid.MustParse("d4444444-4444-4444-8444-444444444444")
	occurrenceRSVPOverrideID   = uuid.MustParse("d5555555-5555-4555-8555-555555555555")
	occurrenceRSVPApplication  = uuid.MustParse("d6666666-6666-4666-8666-666666666666")
	occurrenceRSVPReceiptID    = uuid.MustParse("d7777777-7777-4777-8777-777777777777")
	occurrenceRSVPOtherID      = uuid.MustParse("d8888888-8888-4888-8888-888888888888")
)

var _ calendar.EventRepository = (*occurrenceRSVPRepositoryFake)(nil)
var _ calendar.EventTransactionBackend = (*occurrenceRSVPBackendFake)(nil)
var _ calendar.CalendarEventOccurrenceRSVPTransactionBackend = (*occurrenceRSVPBackendFake)(nil)

func TestCalendarEventOccurrenceRSVP_ABIConstructorAndHashBoundaries(t *testing.T) {
	assertOccurrenceRSVPStructABI(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPInput{}), []string{"IdempotencyKey", "EventID", "RecurrenceID", "ExpectedAttendeeVersion", "Status"})
	for _, opaque := range []reflect.Type{
		reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPCommand{}),
		reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPResult{}),
		reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPConflictError{}),
	} {
		for field := range opaque.NumField() {
			if opaque.Field(field).PkgPath == "" {
				t.Fatalf("%s exposes field %q", opaque, opaque.Field(field).Name)
			}
		}
	}
	assertOccurrenceRSVPType(t, reflect.TypeOf(calendar.NewCalendarEventOccurrenceRSVPCommand), reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPInput) (calendar.CalendarEventOccurrenceRSVPCommand, error) { panic("oracle") }))
	assertOccurrenceRSVPType(t, reflect.TypeOf((*calendar.EventService).RespondToCalendarOccurrence), reflect.TypeOf(func(*calendar.EventService, context.Context, uuid.UUID, calendar.CalendarEventOccurrenceRSVPCommand) (calendar.CalendarEventOccurrenceRSVPResult, error) { panic("oracle") }))
	assertOccurrenceRSVPValueMethod(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPResult{}), "Response", reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPResult) calendar.MutationReceiptResponse { panic("oracle") }))
	assertOccurrenceRSVPValueMethod(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPResult{}), "Replayed", reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPResult) bool { panic("oracle") }))
	assertOccurrenceRSVPValueMethod(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPConflictError{}), "Kind", reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPConflictError) calendar.CalendarEventOccurrenceRSVPConflictKind { panic("oracle") }))
	assertOccurrenceRSVPValueMethod(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPConflictError{}), "EventID", reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPConflictError) (uuid.UUID, bool) { panic("oracle") }))
	assertOccurrenceRSVPValueMethod(t, reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPConflictError{}), "CurrentVersion", reflect.TypeOf(func(calendar.CalendarEventOccurrenceRSVPConflictError) (uint, bool) { panic("oracle") }))

	for _, testCase := range []struct {
		name  string
		input calendar.CalendarEventOccurrenceRSVPInput
		valid bool
	}{
		{name: "all day key", input: occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted), valid: true},
		{name: "utc second key", input: occurrenceRSVPInput("2033-01-03T04:05:06Z", calendar.RSVPAccepted), valid: true},
		{name: "maximum raw key", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 128), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}, valid: true},
		{name: "short raw key", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 15), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}},
		{name: "overlong raw key", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 129), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}},
		{name: "surrounding whitespace", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: " " + strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}},
		{name: "zero event", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}},
		{name: "noncanonical recurrence", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03T04:05:06+03:00", ExpectedAttendeeVersion: 2, Status: calendar.RSVPAccepted}},
		{name: "zero version", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Status: calendar.RSVPAccepted}},
		{name: "max version", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: ^uint(0), Status: calendar.RSVPAccepted}},
		{name: "needs action target", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPNeedsAction}},
		{name: "unknown target", input: calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", ExpectedAttendeeVersion: 2, Status: calendar.RSVPStatus("unknown")}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.NewCalendarEventOccurrenceRSVPCommand(testCase.input)
			if testCase.valid != (err == nil) {
				t.Fatalf("NewCalendarEventOccurrenceRSVPCommand() error = %v, valid:%t", err, testCase.valid)
			}
			if !testCase.valid && !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("constructor error = %v, want invalid", err)
			}
		})
	}
	if calendar.CalendarEventOccurrenceRSVPAttendeeVersionConflict != "attendee_version_conflict" || calendar.CalendarEventOccurrenceRSVPIdempotencyKeyReused != "idempotency_key_reused" {
		t.Fatalf("conflict values changed: %q %q", calendar.CalendarEventOccurrenceRSVPAttendeeVersionConflict, calendar.CalendarEventOccurrenceRSVPIdempotencyKeyReused)
	}
}

func TestCalendarEventOccurrenceRSVP_ConflictErrorsExposeOnlyPermittedMetadata(t *testing.T) {
	unknown := calendar.CalendarEventOccurrenceRSVPConflictError{}
	if unknown.Kind() != "" || unknown.Error() != "calendar event occurrence rsvp error" || errors.Unwrap(unknown) != nil {
		t.Fatalf("unknown conflict = %#v", unknown)
	}
	if eventID, found := unknown.EventID(); found || eventID != uuid.Nil {
		t.Fatalf("unknown conflict exposed event %s/%t", eventID, found)
	}
	if version, found := unknown.CurrentVersion(); found || version != 0 {
		t.Fatalf("unknown conflict exposed version %d/%t", version, found)
	}

	staleBackend := newOccurrenceRSVPBackend(calendar.RSVPAccepted, false, false)
	staleBackend.authorization.SeriesAttendee.Version = 9
	staleService := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: staleBackend}, &occurrenceRSVPClockFake{now: time.Date(2033, time.January, 4, 5, 6, 7, 0, time.UTC), backend: staleBackend}, &occurrenceRSVPIDGeneratorFake{backend: staleBackend}, occurrenceRSVPSerializerUnused{})
	_, err := staleService.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)))
	versionConflict := requireOccurrenceRSVPConflict(t, err, calendar.CalendarEventOccurrenceRSVPAttendeeVersionConflict)
	if errors.Unwrap(versionConflict) != calendar.ErrEventConflict {
		t.Fatalf("version conflict unwrap = %v", errors.Unwrap(versionConflict))
	}
	if eventID, found := versionConflict.EventID(); !found || eventID != occurrenceRSVPEventID {
		t.Fatalf("version conflict event = %s/%t", eventID, found)
	}
	if version, found := versionConflict.CurrentVersion(); !found || version != 9 {
		t.Fatalf("version conflict current version = %d/%t", version, found)
	}

	reusedBackend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, false)
	reusedService := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: reusedBackend}, &occurrenceRSVPClockFake{now: time.Date(2033, time.January, 4, 5, 6, 7, 0, time.UTC), backend: reusedBackend}, &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: reusedBackend}, occurrenceRSVPSerializerUnused{})
	if _, err := reusedService.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted))); err != nil {
		t.Fatalf("seed receipt error = %v", err)
	}
	_, err = reusedService.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPDeclined)))
	reusedConflict := requireOccurrenceRSVPConflict(t, err, calendar.CalendarEventOccurrenceRSVPIdempotencyKeyReused)
	if errors.Unwrap(reusedConflict) != calendar.ErrEventConflict {
		t.Fatalf("reused conflict unwrap = %v", errors.Unwrap(reusedConflict))
	}
	if eventID, found := reusedConflict.EventID(); found || eventID != uuid.Nil {
		t.Fatalf("reused conflict exposed event %s/%t", eventID, found)
	}
	if version, found := reusedConflict.CurrentVersion(); found || version != 0 {
		t.Fatalf("reused conflict exposed version %d/%t", version, found)
	}
}

func TestEventServiceRespondToCalendarOccurrence_MaterializesOrUpdatesWithoutChangingSeries(t *testing.T) {
	now := time.Date(2033, time.February, 4, 5, 6, 7, 0, time.FixedZone("UTC+3", 3*60*60))
	for _, testCase := range []struct {
		name        string
		existing    bool
		previous    calendar.RSVPStatus
		target      calendar.RSVPStatus
		expected    uint
		ids         []uuid.UUID
		wantIDCalls int
	}{
		{name: "inherit needs action accepted", previous: calendar.RSVPNeedsAction, target: calendar.RSVPAccepted, expected: 2, ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, wantIDCalls: 2},
		{name: "existing accepted tentative", existing: true, previous: calendar.RSVPAccepted, target: calendar.RSVPTentative, expected: 4, ids: []uuid.UUID{occurrenceRSVPApplication}, wantIDCalls: 1},
		{name: "existing same status remains mutation", existing: true, previous: calendar.RSVPDeclined, target: calendar.RSVPDeclined, expected: 4, ids: []uuid.UUID{occurrenceRSVPApplication}, wantIDCalls: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPBackend(testCase.previous, testCase.existing, false)
			seriesBefore := backend.authorization.SeriesAttendee
			clock := &occurrenceRSVPClockFake{now: now, backend: backend}
			ids := &occurrenceRSVPIDGeneratorFake{ids: testCase.ids, backend: backend}
			service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, clock, ids, occurrenceRSVPSerializerUnused{})
			input := occurrenceRSVPInput("2033-01-03", testCase.target)
			input.ExpectedAttendeeVersion = testCase.expected
			result, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, input))
			if err != nil {
				t.Fatalf("RespondToCalendarOccurrence() error = %v", err)
			}
			if result.Replayed() || result.Response().Status() != 200 || !reflect.DeepEqual(backend.order, []string{"scope", "reserve", "resolve", "persist", "complete"}) {
				t.Fatalf("result replay:%t status:%d order:%v", result.Replayed(), result.Response().Status(), backend.order)
			}
			if clock.calls != 1 || ids.calls != testCase.wantIDCalls || backend.persistCalls != 1 || backend.completeCalls != 1 || backend.completionEventID != occurrenceRSVPEventID {
				t.Fatalf("calls clock:%d ids:%d persist:%d complete:%d completion:%s", clock.calls, ids.calls, backend.persistCalls, backend.completeCalls, backend.completionEventID)
			}
			if !reflect.DeepEqual(seriesBefore, backend.authorization.SeriesAttendee) || !reflect.DeepEqual(seriesBefore, backend.persisted[0].SeriesAttendee()) {
				t.Fatalf("series snapshot changed: before:%#v after:%#v persisted:%#v", seriesBefore, backend.authorization.SeriesAttendee, backend.persisted[0].SeriesAttendee())
			}
			persisted := backend.persisted[0]
			prior, hadPrior := persisted.Prior()
			if hadPrior != testCase.existing || hadPrior && prior.AttendeeID != occurrenceRSVPOverrideID || persisted.ExpectedVersion() != testCase.expected {
				t.Fatalf("prior:%#v found:%t expected:%d", prior, hadPrior, persisted.ExpectedVersion())
			}
			attendee := persisted.Attendee()
			if attendee.AttendeeID != occurrenceRSVPOverrideID || attendee.EventID != occurrenceRSVPEventID || attendee.RecurrenceID != "2033-01-03" || attendee.Status != testCase.target || attendee.Version != testCase.expected+1 || attendee.RespondedAt == nil || !attendee.RespondedAt.Equal(now.UTC()) {
				t.Fatalf("updated occurrence attendee = %#v", attendee)
			}
			assertOccurrenceRSVPResponse(t, result.Response(), attendee, "not_required")
			assertOccurrenceRSVPEnvelope(t, persisted.Outbox(), occurrenceRSVPApplication, now.UTC(), testCase.previous, attendee, "not_required")
		})
	}
}

func TestEventServiceRespondToCalendarOccurrence_ExternalReplyUsesLiveExceptionSequence(t *testing.T) {
	backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, true)
	sequence := uint(0)
	temporal := backend.authorization.Current.Temporal
	backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Temporal: &temporal, Sequence: sequence, Status: calendar.CalendarEventStatusTentative, Cancelled: false}
	now := time.Date(2033, time.March, 4, 5, 6, 7, 0, time.UTC)
	service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, &occurrenceRSVPClockFake{now: now, backend: backend}, &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}, occurrenceRSVPSerializerUnused{})
	result, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPDeclined)))
	if err != nil {
		t.Fatalf("RespondToCalendarOccurrence() error = %v", err)
	}
	persisted := backend.persisted[0]
	attendee := persisted.Attendee()
	reply, found := persisted.Reply()
	if !found || reply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{AttendeeID: attendee.AttendeeID, EventID: occurrenceRSVPEventID, EventSequence: 0, AttendeeVersion: attendee.Version, RecurrenceID: "2033-01-03", Status: calendar.RSVPDeclined}) {
		t.Fatalf("reply = %#v found:%t", reply, found)
	}
	if persisted.EventSequence() != 0 {
		t.Fatalf("effective sequence = %d, want zero", persisted.EventSequence())
	}
	assertOccurrenceRSVPResponse(t, result.Response(), attendee, "pending")
	assertOccurrenceRSVPEnvelope(t, persisted.Outbox(), occurrenceRSVPApplication, now, calendar.RSVPNeedsAction, attendee, "pending")
}

func TestEventServiceRespondToCalendarOccurrence_AllowsNilTemporalForLiveException(t *testing.T) {
	for _, status := range []calendar.CalendarEventStatus{calendar.CalendarEventStatusConfirmed, calendar.CalendarEventStatusTentative} {
		t.Run(string(status), func(t *testing.T) {
			backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, true)
			backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Sequence: 0, Status: status, Cancelled: false}
			now := time.Date(2033, time.March, 5, 6, 7, 8, 0, time.UTC)
			service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, &occurrenceRSVPClockFake{now: now, backend: backend}, &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}, occurrenceRSVPSerializerUnused{})
			result, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPDeclined)))
			if err != nil {
				t.Fatalf("RespondToCalendarOccurrence() error = %v", err)
			}
			persisted := backend.persisted[0]
			reply, found := persisted.Reply()
			if !found || persisted.EventSequence() != 0 || reply.EventSequence != 0 {
				t.Fatalf("nil-temporal exception reply = %#v found:%t sequence:%d", reply, found, persisted.EventSequence())
			}
			assertOccurrenceRSVPResponse(t, result.Response(), persisted.Attendee(), "pending")
		})
	}
}

func TestEventServiceRespondToCalendarOccurrence_ReplaysExactBytesAndHashesOnlyFrozenFingerprint(t *testing.T) {
	backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, false)
	now := time.Date(2033, time.April, 4, 5, 6, 7, 0, time.UTC)
	clock := &occurrenceRSVPClockFake{now: now, backend: backend}
	ids := &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}
	service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, clock, ids, occurrenceRSVPSerializerUnused{})
	input := occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)
	first, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, input))
	if err != nil { t.Fatalf("first error = %v", err) }
	second, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, input))
	if err != nil { t.Fatalf("replay error = %v", err) }
	if !second.Replayed() || !reflect.DeepEqual(first.Response().JSON(), second.Response().JSON()) || clock.calls != 1 || ids.calls != 2 || backend.resolveCalls != 1 || backend.persistCalls != 1 || backend.completeCalls != 1 {
		t.Fatalf("replay calls clock:%d ids:%d resolve:%d persist:%d complete:%d", clock.calls, ids.calls, backend.resolveCalls, backend.persistCalls, backend.completeCalls)
	}
	canonical, err := json.Marshal(struct {
		Version int `json:"version"`; EventID uuid.UUID `json:"event_id"`; RecurrenceID string `json:"recurrence_id"`; ExpectedAttendeeVersion uint `json:"expected_attendee_version"`; Status calendar.RSVPStatus `json:"status"`
	}{Version: 1, EventID: input.EventID, RecurrenceID: input.RecurrenceID, ExpectedAttendeeVersion: input.ExpectedAttendeeVersion, Status: input.Status})
	if err != nil { t.Fatalf("canonical JSON error = %v", err) }
	if backend.reservation.KeyHash() != calendar.MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))) || backend.reservation.RequestHash() != calendar.MutationReceiptHash(sha256.Sum256(canonical)) {
		t.Fatalf("receipt hashes key:%x request:%x canonical:%s", backend.reservation.KeyHash(), backend.reservation.RequestHash(), canonical)
	}
	_, err = service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPDeclined)))
	conflict := requireOccurrenceRSVPConflict(t, err, calendar.CalendarEventOccurrenceRSVPIdempotencyKeyReused)
	if _, found := conflict.EventID(); found { t.Fatalf("reused key exposed event: %#v", conflict) }
	if _, found := conflict.CurrentVersion(); found { t.Fatalf("reused key exposed version: %#v", conflict) }
}

func TestEventServiceRespondToCalendarOccurrence_StaleAndPolicyFailuresMaskStateBeforeClockOrIDs(t *testing.T) {
	for _, testCase := range []struct {
		name string
		existing bool
		mutate func(*occurrenceRSVPBackendFake)
		want error
		kind calendar.CalendarEventOccurrenceRSVPConflictKind
		wantVersion uint
	}{
		{name: "inherited stale series version", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.SeriesAttendee.Version = 9 }, want: calendar.ErrEventConflict, kind: calendar.CalendarEventOccurrenceRSVPAttendeeVersionConflict, wantVersion: 9},
		{name: "override stale version", existing: true, mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.OccurrenceAttendee.Version = 11 }, want: calendar.ErrEventConflict, kind: calendar.CalendarEventOccurrenceRSVPAttendeeVersionConflict, wantVersion: 11},
		{name: "cancelled master", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled }, want: calendar.ErrEventNotFound},
		{name: "foreign event", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Current.EventID = occurrenceRSVPOtherID }, want: calendar.ErrEventNotFound},
		{name: "foreign occurrence recurrence", existing: true, mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.OccurrenceAttendee.RecurrenceID = "2033-01-04" }, want: calendar.ErrEventNotFound},
		{name: "foreign exception event", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPOtherID, RecurrenceID: "2033-01-03", Sequence: 0, Status: calendar.CalendarEventStatusConfirmed} }, want: calendar.ErrEventNotFound},
		{name: "foreign exception recurrence", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-04", Sequence: 0, Status: calendar.CalendarEventStatusConfirmed} }, want: calendar.ErrEventNotFound},
		{name: "non recurring", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Current.RRULE = "" }, want: calendar.ErrEventNotFound},
		{name: "excluded recurrence", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Current.Exdates = []string{"2033-01-03"} }, want: calendar.ErrEventNotFound},
		{name: "cancelled exception", mutate: func(backend *occurrenceRSVPBackendFake) { backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Sequence: 0, Status: calendar.CalendarEventStatusCancelled, Cancelled: true} }, want: calendar.ErrEventNotFound},
		{name: "cancelled exception with valid temporal", mutate: func(backend *occurrenceRSVPBackendFake) { temporal := backend.authorization.Current.Temporal; backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Temporal: &temporal, Sequence: 0, Status: calendar.CalendarEventStatusCancelled, Cancelled: true} }, want: calendar.ErrEventNotFound},
		{name: "malformed live exception", mutate: func(backend *occurrenceRSVPBackendFake) { temporal := backend.authorization.Current.Temporal; backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Temporal: &temporal, Sequence: 99, Status: calendar.CalendarEventStatusTentative, Cancelled: false} }, want: calendar.ErrEventInvalid},
		{name: "malformed live exception temporal", mutate: func(backend *occurrenceRSVPBackendFake) { temporal := calendar.CalendarEvent{}; backend.authorization.Exception = &calendar.CalendarEventExceptionCommandInput{ExceptionID: occurrenceRSVPOtherID, EventID: occurrenceRSVPEventID, RecurrenceID: "2033-01-03", Temporal: &temporal, Sequence: 0, Status: calendar.CalendarEventStatusTentative, Cancelled: false} }, want: calendar.ErrEventInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPBackend(calendar.RSVPAccepted, testCase.existing, false)
			testCase.mutate(backend)
			clock := &occurrenceRSVPClockFake{now: time.Date(2033, time.May, 4, 5, 6, 7, 0, time.UTC), backend: backend}
			ids := &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}
			service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, clock, ids, occurrenceRSVPSerializerUnused{})
			_, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)))
			if !errors.Is(err, testCase.want) { t.Fatalf("error = %v, want %v", err, testCase.want) }
			if clock.calls != 0 || ids.calls != 0 || backend.persistCalls != 0 || backend.completeCalls != 0 { t.Fatalf("blocked calls clock:%d ids:%d persist:%d complete:%d", clock.calls, ids.calls, backend.persistCalls, backend.completeCalls) }
			if testCase.kind != "" {
				conflict := requireOccurrenceRSVPConflict(t, err, testCase.kind)
				if id, found := conflict.EventID(); !found || id != occurrenceRSVPEventID { t.Fatalf("stale event = %s/%t", id, found) }
				if version, found := conflict.CurrentVersion(); !found || version != testCase.wantVersion { t.Fatalf("stale version = %d/%t want:%d", version, found, testCase.wantVersion) }
			}
		})
	}
}

func TestEventServiceRespondToCalendarOccurrence_NonMemberRecurrenceMasksStateBeforeClockOrIDs(t *testing.T) {
	backend := newOccurrenceRSVPBackend(calendar.RSVPAccepted, false, false)
	backend.authorization.Current.Temporal.StartDate = "2033-01-04"
	backend.authorization.Current.Temporal.EndDateExclusive = "2033-01-05"
	clock := &occurrenceRSVPClockFake{now: time.Date(2033, time.May, 4, 5, 6, 7, 0, time.UTC), backend: backend}
	ids := &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}
	service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, clock, ids, occurrenceRSVPSerializerUnused{})

	_, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)))
	if !errors.Is(err, calendar.ErrEventNotFound) {
		t.Fatalf("error = %v, want not found", err)
	}
	if backend.resolveCalls != 1 || !reflect.DeepEqual(backend.order, []string{"scope", "reserve", "resolve"}) {
		t.Fatalf("non-member recurrence must be rejected after authorization: resolve:%d order:%v", backend.resolveCalls, backend.order)
	}
	if clock.calls != 0 || ids.calls != 0 || backend.persistCalls != 0 || backend.completeCalls != 0 || backend.receipt != nil {
		t.Fatalf("blocked calls clock:%d ids:%d persist:%d complete:%d receipt:%#v", clock.calls, ids.calls, backend.persistCalls, backend.completeCalls, backend.receipt)
	}
}

func TestEventServiceRespondToCalendarOccurrence_RollsBackEchoCompletionAndUnsafeBackendFailures(t *testing.T) {
	for _, testCase := range []struct {
		name string
		configure func(*occurrenceRSVPBackendFake)
	}{
		{name: "mismatched echo", configure: func(backend *occurrenceRSVPBackendFake) {
			backend.resultOverride = func(command calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) {
				attendee := command.Attendee(); attendee.Version++
				return calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(attendee, command.EventSequence(), command.Outbox(), nil)
			}
		}},
		{name: "raw backend error", configure: func(backend *occurrenceRSVPBackendFake) { backend.persistError = errors.New("provider leaked attendee@example.test idempotency secret") }},
		{name: "completion error", configure: func(backend *occurrenceRSVPBackendFake) { backend.completeError = errors.New("completion leaked attendee@example.test") }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, false)
			testCase.configure(backend)
			service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, &occurrenceRSVPClockFake{now: time.Date(2033, time.June, 4, 5, 6, 7, 0, time.UTC), backend: backend}, &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}, occurrenceRSVPSerializerUnused{})
			_, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)))
			if !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("error = %v, want invalid", err) }
			if strings.Contains(err.Error(), "attendee@example.test") || strings.Contains(err.Error(), "idempotency") || len(backend.persisted) != 0 || backend.receipt != nil { t.Fatalf("unsafe or unrolled state error:%q persisted:%d receipt:%#v", err, len(backend.persisted), backend.receipt) }
		})
	}
}

func TestEventServiceRespondToCalendarOccurrence_AllowsNormalizedImportedMasterAndCopiesPortSnapshots(t *testing.T) {
	backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, true)
	backend.authorization.Current.Source = calendar.CalendarEventSourceMailImport
	clock := &occurrenceRSVPClockFake{now: time.Date(2033, time.July, 4, 5, 6, 7, 0, time.UTC), backend: backend}
	ids := &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}
	service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, clock, ids, occurrenceRSVPSerializerUnused{})
	if _, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted))); err != nil {
		t.Fatalf("imported-master attendee RSVP error = %v", err)
	}
	persisted := backend.persisted[0]
	attendee := persisted.Attendee()
	attendee.Email = "mutated@example.test"
	if persisted.Attendee().Email != "attendee@example.test" {
		t.Fatal("occurrence attendee getter leaked mutable snapshot")
	}
	series := persisted.SeriesAttendee()
	series.Email = "mutated@example.test"
	if persisted.SeriesAttendee().Email != "attendee@example.test" {
		t.Fatal("series attendee getter leaked mutable snapshot")
	}
	reply, found := persisted.Reply()
	if !found { t.Fatal("external organizer must receive a reply descriptor") }
	reply.RecurrenceID = "2033-01-04"
	persistedReply, persistedFound := persisted.Reply()
	if !persistedFound || persistedReply.RecurrenceID != "2033-01-03" {
		t.Fatalf("reply getter leaked mutable snapshot: %#v found:%t", persistedReply, persistedFound)
	}
}

func TestEventServiceRespondToCalendarOccurrence_ContainmentRejectsHostileErrorTraversal(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
	}{
		{name: "raw", err: errors.New("provider leaked attendee@example.test")},
		{name: "typed nil", err: occurrenceRSVPTypedNilErrorValue()},
		{name: "cycle", err: &occurrenceRSVPSelfCycleError{}},
		{name: "joined", err: occurrenceRSVPMultiUnwrapError{children: []error{errors.New("first secret"), errors.New("second secret")}}},
		{name: "panicking", err: occurrenceRSVPPanickingUnwrapError{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPBackend(calendar.RSVPNeedsAction, false, false)
			backend.persistError = testCase.err
			service := calendar.NewEventService(&occurrenceRSVPRepositoryFake{backend: backend}, &occurrenceRSVPClockFake{now: time.Date(2033, time.August, 4, 5, 6, 7, 0, time.UTC), backend: backend}, &occurrenceRSVPIDGeneratorFake{ids: []uuid.UUID{occurrenceRSVPOverrideID, occurrenceRSVPApplication}, backend: backend}, occurrenceRSVPSerializerUnused{})
			_, err := service.RespondToCalendarOccurrence(context.Background(), occurrenceRSVPActorID, mustOccurrenceRSVPCommand(t, occurrenceRSVPInput("2033-01-03", calendar.RSVPAccepted)))
			if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "attendee@example.test") {
				t.Fatalf("unsafe containment error = %v", err)
			}
		})
	}
}

func occurrenceRSVPInput(recurrenceID string, status calendar.RSVPStatus) calendar.CalendarEventOccurrenceRSVPInput {
	return calendar.CalendarEventOccurrenceRSVPInput{IdempotencyKey: "occurrence-rsvp-key", EventID: occurrenceRSVPEventID, RecurrenceID: recurrenceID, ExpectedAttendeeVersion: 2, Status: status}
}

func mustOccurrenceRSVPCommand(t *testing.T, input calendar.CalendarEventOccurrenceRSVPInput) calendar.CalendarEventOccurrenceRSVPCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventOccurrenceRSVPCommand(input)
	if err != nil { t.Fatalf("NewCalendarEventOccurrenceRSVPCommand() error = %v", err) }
	return command
}

type occurrenceRSVPRepositoryFake struct { backend *occurrenceRSVPBackendFake }
func (repository *occurrenceRSVPRepositoryFake) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	snapshot := repository.backend.snapshot()
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil { repository.backend.restore(snapshot) }
	return err
}

type occurrenceRSVPReceipt struct { keyHash calendar.MutationReceiptHash; requestHash calendar.MutationReceiptHash; response *calendar.MutationReceiptResponse }
type occurrenceRSVPBackendSnapshot struct { receipt *occurrenceRSVPReceipt; persisted []calendar.CalendarEventOccurrenceRSVPPersistenceCommand }

type occurrenceRSVPBackendFake struct {
	organizationID uuid.UUID
	authorization calendar.CalendarEventOccurrenceRSVPAuthorization
	reservation calendar.MutationReceiptReservation
	receipt *occurrenceRSVPReceipt
	persisted []calendar.CalendarEventOccurrenceRSVPPersistenceCommand
	resultOverride func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error)
	persistError error
	completeError error
	order []string
	resolveCalls, persistCalls, completeCalls, clockCalls int
	completionEventID uuid.UUID
	ids occurrenceRSVPIDGeneratorFake
}

func newOccurrenceRSVPBackend(previous calendar.RSVPStatus, existing, externalOrganizer bool) *occurrenceRSVPBackendFake {
	var organizerID *uuid.UUID
	if !externalOrganizer { organizerID = occurrenceRSVPUUID(occurrenceRSVPActorID) }
	responded := (*time.Time)(nil)
	if previous != calendar.RSVPNeedsAction { responded = occurrenceRSVPTime(time.Date(2033, time.January, 2, 3, 4, 5, 0, time.UTC)) }
	backend := &occurrenceRSVPBackendFake{organizationID: occurrenceRSVPOrganization}
	backend.authorization = calendar.CalendarEventOccurrenceRSVPAuthorization{
		Current: calendar.CalendarEventCommandInput{EventID: occurrenceRSVPEventID, OrganizationID: occurrenceRSVPOrganization, OwnerID: occurrenceRSVPActorID, UID: "hub-" + occurrenceRSVPEventID.String(), Title: "Occurrence RSVP event", Description: "", Location: "", Temporal: calendar.CalendarEvent{AllDay: true, StartDate: "2033-01-03", EndDateExclusive: "2033-01-04", Timezone: "UTC"}, RRULE: "FREQ=DAILY;INTERVAL=1;COUNT=2", Sequence: 0, Version: 7, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed, OrganizerID: organizerID, OrganizerEmail: "organizer@example.test"},
		SeriesAttendee: calendar.CalendarEventSeriesRSVPAttendeeSnapshot{AttendeeID: occurrenceRSVPSeriesID, EventID: occurrenceRSVPEventID, UserID: occurrenceRSVPUUID(occurrenceRSVPActorID), Email: "attendee@example.test", DisplayName: "Attendee", Role: calendar.CalendarEventAttendeeRoleRequired, Status: previous, RespondedAt: responded, Version: 2},
	}
	if existing {
		backend.authorization.OccurrenceAttendee = &calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot{AttendeeID: occurrenceRSVPOverrideID, EventID: occurrenceRSVPEventID, UserID: occurrenceRSVPUUID(occurrenceRSVPActorID), Email: "attendee@example.test", DisplayName: "Attendee", Role: calendar.CalendarEventAttendeeRoleRequired, Status: previous, RecurrenceID: "2033-01-03", RespondedAt: responded, Version: 4}
	}
	return backend
}

func (backend *occurrenceRSVPBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	backend.order = append(backend.order, "scope")
	if actorID != occurrenceRSVPActorID { return uuid.Nil, calendar.ErrEventForbidden }
	return backend.organizationID, nil
}
func (backend *occurrenceRSVPBackendFake) ReserveMutationReceipt(_ context.Context, _ calendar.EventMutationScope, reservation calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	backend.order = append(backend.order, "reserve"); backend.reservation = reservation
	if backend.receipt != nil {
		if backend.receipt.keyHash != reservation.KeyHash() || backend.receipt.requestHash != reservation.RequestHash() || backend.receipt.response == nil { return calendar.MutationReceiptReservationResult{}, calendar.ErrEventConflict }
		return calendar.NewReplayedMutationReceiptResult(occurrenceRSVPReceiptID, *backend.receipt.response)
	}
	backend.receipt = &occurrenceRSVPReceipt{keyHash: reservation.KeyHash(), requestHash: reservation.RequestHash()}
	return calendar.NewReservedMutationReceiptResult(occurrenceRSVPReceiptID)
}
func (backend *occurrenceRSVPBackendFake) CompleteMutationReceipt(_ context.Context, _ calendar.EventMutationScope, completion calendar.MutationReceiptCompletion) error {
	backend.order = append(backend.order, "complete"); backend.completeCalls++
	if backend.completeError != nil { return backend.completeError }
	eventID, found := completion.EventID()
	if backend.receipt == nil || completion.ReceiptID() != occurrenceRSVPReceiptID || !found || eventID != occurrenceRSVPEventID { return calendar.ErrEventInvalid }
	backend.completionEventID = eventID; response := completion.Response(); backend.receipt.response = &response
	return nil
}
func (backend *occurrenceRSVPBackendFake) ResolveCalendarEventOccurrenceRSVP(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventOccurrenceRSVPAuthorizationRequest) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) {
	backend.order = append(backend.order, "resolve"); backend.resolveCalls++
	if request.EventID != occurrenceRSVPEventID || request.RecurrenceID != "2033-01-03" { return calendar.CalendarEventOccurrenceRSVPAuthorization{}, calendar.ErrEventNotFound }
	return backend.authorization, nil
}
func (backend *occurrenceRSVPBackendFake) PersistCalendarEventOccurrenceRSVP(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) {
	backend.order = append(backend.order, "persist"); backend.persistCalls++
	if backend.persistError != nil { return calendar.CalendarEventOccurrenceRSVPPersistenceResult{}, backend.persistError }
	if backend.resultOverride != nil { return backend.resultOverride(command) }
	backend.persisted = append(backend.persisted, command)
	reply, found := command.Reply(); var replyPointer *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor; if found { replyPointer = &reply }
	return calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(command.Attendee(), command.EventSequence(), command.Outbox(), replyPointer)
}
func (backend *occurrenceRSVPBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *occurrenceRSVPBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceRSVPBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceRSVPBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *occurrenceRSVPBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { return calendar.ErrEventInvalid }
func (backend *occurrenceRSVPBackendFake) snapshot() occurrenceRSVPBackendSnapshot {
	snapshot := occurrenceRSVPBackendSnapshot{persisted: append([]calendar.CalendarEventOccurrenceRSVPPersistenceCommand(nil), backend.persisted...)}
	if backend.receipt != nil { receipt := *backend.receipt; if receipt.response != nil { response := *receipt.response; receipt.response = &response }; snapshot.receipt = &receipt }
	return snapshot
}
func (backend *occurrenceRSVPBackendFake) restore(snapshot occurrenceRSVPBackendSnapshot) { backend.persisted = snapshot.persisted; backend.receipt = snapshot.receipt }

type occurrenceRSVPClockFake struct { now time.Time; calls int; backend *occurrenceRSVPBackendFake }
func (clock *occurrenceRSVPClockFake) Now() time.Time { clock.calls++; if clock.backend != nil { clock.backend.clockCalls++ }; return clock.now }
type occurrenceRSVPIDGeneratorFake struct { ids []uuid.UUID; calls int; backend *occurrenceRSVPBackendFake }
func (generator *occurrenceRSVPIDGeneratorFake) NewEventID() uuid.UUID { if generator.calls >= len(generator.ids) { return uuid.Nil }; value := generator.ids[generator.calls]; generator.calls++; if generator.backend != nil { generator.backend.ids.calls = generator.calls }; return value }
type occurrenceRSVPSerializerUnused struct{}
func (occurrenceRSVPSerializerUnused) Serialize(calendar.ITIPMessage) ([]byte, error) { panic("occurrence RSVP must not serialize iTIP outside its persistence port") }

type occurrenceRSVPResponseBody struct { ID uuid.UUID `json:"id"`; UserID *uuid.UUID `json:"user_id"`; Email string `json:"email"`; DisplayName *string `json:"display_name"`; Role calendar.CalendarEventAttendeeRole `json:"role"`; RSVPStatus calendar.RSVPStatus `json:"rsvp_status"`; RecurrenceID *string `json:"recurrence_id"`; RespondedAt *time.Time `json:"responded_at"`; Version uint `json:"version"`; ReplyDeliveryStatus string `json:"reply_delivery_status"` }
func assertOccurrenceRSVPResponse(t *testing.T, response calendar.MutationReceiptResponse, attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, deliveryStatus string) {
	t.Helper(); if response.Status() != 200 { t.Fatalf("response status = %d", response.Status()) }
	var fields map[string]json.RawMessage; if err := json.Unmarshal(response.JSON(), &fields); err != nil { t.Fatalf("response JSON = %q: %v", response.JSON(), err) }; if len(fields) != 10 { t.Fatalf("response fields = %v", occurrenceRSVPKeys(fields)) }
	recurrenceID := attendee.RecurrenceID; var displayName *string; if attendee.DisplayName != "" { displayName = &attendee.DisplayName }
	want, err := json.Marshal(occurrenceRSVPResponseBody{ID: attendee.AttendeeID, UserID: attendee.UserID, Email: attendee.Email, DisplayName: displayName, Role: attendee.Role, RSVPStatus: attendee.Status, RecurrenceID: &recurrenceID, RespondedAt: attendee.RespondedAt, Version: attendee.Version, ReplyDeliveryStatus: deliveryStatus})
	if err != nil || !reflect.DeepEqual(response.JSON(), want) { t.Fatalf("response JSON = %s, want %s, error:%v", response.JSON(), want, err) }
}
func assertOccurrenceRSVPEnvelope(t *testing.T, descriptor calendar.EventOutboxDescriptor, applicationID uuid.UUID, occurredAt time.Time, previous calendar.RSVPStatus, attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, deliveryStatus string) {
	t.Helper(); envelope := descriptor.Envelope(); if envelope.Subject() != "calendar.rsvp.changed.v1" || envelope.EventID() != applicationID { t.Fatalf("envelope subject:%q event:%s", envelope.Subject(), envelope.EventID()) }
	encoded, err := envelope.MarshalJSON(); if err != nil { t.Fatalf("marshal envelope: %v", err) }
	var outer struct { EventID uuid.UUID `json:"event_id"`; OccurredAt time.Time `json:"occurred_at"`; OrganizationID uuid.UUID `json:"organization_id"`; ActorID *uuid.UUID `json:"actor_id"`; CorrelationID string `json:"correlation_id"`; Payload json.RawMessage `json:"payload"` }
	if err := json.Unmarshal(encoded, &outer); err != nil || outer.EventID != applicationID || !outer.OccurredAt.Equal(occurredAt) || outer.OrganizationID != occurrenceRSVPOrganization || outer.ActorID == nil || *outer.ActorID != occurrenceRSVPActorID || outer.CorrelationID != occurrenceRSVPEventID.String() { t.Fatalf("envelope binding:%#v error:%v", outer, err) }
	var payload map[string]json.RawMessage; if err := json.Unmarshal(outer.Payload, &payload); err != nil { t.Fatalf("payload JSON: %v", err) }
	wantKeys := map[string]struct{}{"event_id": {}, "attendee_id": {}, "attendee_email": {}, "previous_status": {}, "new_status": {}, "version": {}, "recurrence_id": {}, "reply_delivery_status": {}}
	if len(payload) != len(wantKeys) { t.Fatalf("payload keys = %v", occurrenceRSVPKeys(payload)) }; for key := range wantKeys { if _, found := payload[key]; !found { t.Fatalf("payload missing %q", key) } }
	var decoded struct { EventID uuid.UUID `json:"event_id"`; AttendeeID uuid.UUID `json:"attendee_id"`; AttendeeEmail string `json:"attendee_email"`; Previous calendar.RSVPStatus `json:"previous_status"`; Current calendar.RSVPStatus `json:"new_status"`; Version uint `json:"version"`; RecurrenceID *string `json:"recurrence_id"`; Delivery string `json:"reply_delivery_status"` }
	if err := json.Unmarshal(outer.Payload, &decoded); err != nil || decoded.EventID != occurrenceRSVPEventID || decoded.AttendeeID != attendee.AttendeeID || decoded.AttendeeEmail != attendee.Email || decoded.Previous != previous || decoded.Current != attendee.Status || decoded.Version != attendee.Version || decoded.RecurrenceID == nil || *decoded.RecurrenceID != attendee.RecurrenceID || decoded.Delivery != deliveryStatus { t.Fatalf("payload=%#v error:%v", decoded, err) }
}
func requireOccurrenceRSVPConflict(t *testing.T, err error, kind calendar.CalendarEventOccurrenceRSVPConflictKind) calendar.CalendarEventOccurrenceRSVPConflictError { t.Helper(); var conflict calendar.CalendarEventOccurrenceRSVPConflictError; if !errors.As(err, &conflict) || conflict.Kind() != kind || conflict.Error() != "calendar event occurrence rsvp conflict" || !errors.Is(conflict, calendar.ErrEventConflict) { t.Fatalf("conflict=%#v error:%v", conflict, err) }; return conflict }
func assertOccurrenceRSVPStructABI(t *testing.T, typ reflect.Type, names []string) { t.Helper(); if typ.NumField() != len(names) { t.Fatalf("%s field count %d", typ, typ.NumField()) }; for index, name := range names { if typ.Field(index).Name != name || typ.Field(index).PkgPath != "" { t.Fatalf("%s field[%d]=%#v", typ, index, typ.Field(index)) } } }
func assertOccurrenceRSVPType(t *testing.T, got, want reflect.Type) { t.Helper(); if got != want { t.Fatalf("type=%s want=%s", got, want) } }
func assertOccurrenceRSVPValueMethod(t *testing.T, typ reflect.Type, name string, want reflect.Type) { t.Helper(); method, found := typ.MethodByName(name); if !found || method.Type != want { t.Fatalf("%s.%s=%v found:%t want:%v", typ, name, method.Type, found, want) } }
func occurrenceRSVPKeys(values map[string]json.RawMessage) []string { keys := make([]string, 0, len(values)); for key := range values { keys = append(keys, key) }; return keys }
func occurrenceRSVPUUID(value uuid.UUID) *uuid.UUID { return &value }
func occurrenceRSVPTime(value time.Time) *time.Time { return &value }

type occurrenceRSVPTypedNilError struct{}
func (*occurrenceRSVPTypedNilError) Error() string { return "typed nil provider error" }
func occurrenceRSVPTypedNilErrorValue() error { var value *occurrenceRSVPTypedNilError; return value }
type occurrenceRSVPSelfCycleError struct{}
func (*occurrenceRSVPSelfCycleError) Error() string { return "self cycle provider error" }
func (value *occurrenceRSVPSelfCycleError) Unwrap() error { return value }
type occurrenceRSVPMultiUnwrapError struct{ children []error }
func (occurrenceRSVPMultiUnwrapError) Error() string { return "multi unwrap provider error" }
func (value occurrenceRSVPMultiUnwrapError) Unwrap() []error { return value.children }
type occurrenceRSVPPanickingUnwrapError struct{}
func (occurrenceRSVPPanickingUnwrapError) Error() string { return "panicking provider error" }
func (occurrenceRSVPPanickingUnwrapError) Unwrap() error { panic("provider panic") }
