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
	rsvpServiceActorID        = uuid.MustParse("c1111111-1111-4111-8111-111111111111")
	rsvpServiceOrganizationID = uuid.MustParse("c2222222-2222-4222-8222-222222222222")
	rsvpServiceEventID        = uuid.MustParse("c3333333-3333-4333-8333-333333333333")
	rsvpServiceAttendeeID     = uuid.MustParse("c4444444-4444-4444-8444-444444444444")
	rsvpServiceApplicationID  = uuid.MustParse("c5555555-5555-4555-8555-555555555555")
	rsvpServiceReceiptID      = uuid.MustParse("c6666666-6666-4666-8666-666666666666")
	rsvpServiceOtherID        = uuid.MustParse("c7777777-7777-4777-8777-777777777777")
)

var _ calendar.EventRepository = (*rsvpServiceRepositoryFake)(nil)
var _ calendar.EventTransactionBackend = (*rsvpServiceBackendFake)(nil)
var _ calendar.CalendarEventSeriesRSVPTransactionBackend = (*rsvpServiceBackendFake)(nil)

func TestCalendarEventSeriesRSVP_ABIAndConstructorBoundaries(t *testing.T) {
	assertRSVPServiceStructABI(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPInput{}), []string{"IdempotencyKey", "EventID", "ExpectedAttendeeVersion", "Status"})
	for _, opaque := range []reflect.Type{
		reflect.TypeOf(calendar.CalendarEventSeriesRSVPCommand{}),
		reflect.TypeOf(calendar.CalendarEventSeriesRSVPResult{}),
		reflect.TypeOf(calendar.CalendarEventSeriesRSVPConflictError{}),
	} {
		for field := range opaque.NumField() {
			if opaque.Field(field).PkgPath == "" {
				t.Fatalf("%s exposes field %q", opaque, opaque.Field(field).Name)
			}
		}
	}
	assertRSVPServiceExactType(t, reflect.TypeOf(calendar.NewCalendarEventSeriesRSVPCommand), reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPInput) (calendar.CalendarEventSeriesRSVPCommand, error) { panic("oracle") }))
	assertRSVPServiceExactType(t, reflect.TypeOf((*calendar.EventService).RespondToCalendarSeries), reflect.TypeOf(func(*calendar.EventService, context.Context, uuid.UUID, calendar.CalendarEventSeriesRSVPCommand) (calendar.CalendarEventSeriesRSVPResult, error) { panic("oracle") }))
	assertRSVPServiceValueMethod(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPResult{}), "Response", reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPResult) calendar.MutationReceiptResponse { panic("oracle") }))
	assertRSVPServiceValueMethod(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPResult{}), "Replayed", reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPResult) bool { panic("oracle") }))
	assertRSVPServiceValueMethod(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPConflictError{}), "Kind", reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPConflictError) calendar.CalendarEventSeriesRSVPConflictKind { panic("oracle") }))
	assertRSVPServiceValueMethod(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPConflictError{}), "EventID", reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPConflictError) (uuid.UUID, bool) { panic("oracle") }))
	assertRSVPServiceValueMethod(t, reflect.TypeOf(calendar.CalendarEventSeriesRSVPConflictError{}), "CurrentVersion", reflect.TypeOf(func(calendar.CalendarEventSeriesRSVPConflictError) (uint, bool) { panic("oracle") }))

	valid := rsvpServiceInput(calendar.RSVPAccepted)
	for _, testCase := range []struct {
		name  string
		input calendar.CalendarEventSeriesRSVPInput
		valid bool
	}{
		{name: "minimum raw key", input: valid, valid: true},
		{name: "maximum raw key", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 128), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPAccepted}, valid: true},
		{name: "short key", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 15), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPAccepted}},
		{name: "oversized key", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 129), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPAccepted}},
		{name: "trimmed key required", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: " " + strings.Repeat("k", 16), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPAccepted}},
		{name: "zero event", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 16), ExpectedAttendeeVersion: 1, Status: calendar.RSVPAccepted}},
		{name: "zero attendee version", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: rsvpServiceEventID, Status: calendar.RSVPAccepted}},
		{name: "max attendee version", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: ^uint(0), Status: calendar.RSVPAccepted}},
		{name: "needs action target", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPNeedsAction}},
		{name: "unknown target", input: calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: strings.Repeat("k", 16), EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 1, Status: calendar.RSVPStatus("other")}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.NewCalendarEventSeriesRSVPCommand(testCase.input)
			if testCase.valid != (err == nil) {
				t.Fatalf("NewCalendarEventSeriesRSVPCommand() error = %v, valid:%t", err, testCase.valid)
			}
			if !testCase.valid && !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("invalid constructor error = %v, want %v", err, calendar.ErrEventInvalid)
			}
		})
	}

	if calendar.CalendarEventSeriesRSVPAttendeeVersionConflict != "attendee_version_conflict" || calendar.CalendarEventSeriesRSVPIdempotencyKeyReused != "idempotency_key_reused" {
		t.Fatalf("conflict values changed: %q %q", calendar.CalendarEventSeriesRSVPAttendeeVersionConflict, calendar.CalendarEventSeriesRSVPIdempotencyKeyReused)
	}
	zero := calendar.CalendarEventSeriesRSVPConflictError{}
	if zero.Error() != "calendar event series rsvp error" || zero.Unwrap() != nil || zero.Kind() != "" {
		t.Fatalf("zero conflict = error:%q unwrap:%v kind:%q", zero.Error(), zero.Unwrap(), zero.Kind())
	}
	if _, found := zero.EventID(); found {
		t.Fatalf("zero conflict exposed event id: %#v", zero)
	}
	if _, found := zero.CurrentVersion(); found {
		t.Fatalf("zero conflict exposed current version: %#v", zero)
	}
}

func TestEventServiceRespondToCalendarSeries_UpdatesAllPreviousAndTargetStates(t *testing.T) {
	sharedUTC := time.Date(2033, time.March, 4, 5, 6, 7, 0, time.FixedZone("UTC+3", 3*60*60))
	previousStates := []calendar.RSVPStatus{calendar.RSVPNeedsAction, calendar.RSVPAccepted, calendar.RSVPDeclined, calendar.RSVPTentative}
	targetStates := []calendar.RSVPStatus{calendar.RSVPAccepted, calendar.RSVPDeclined, calendar.RSVPTentative}
	for _, previous := range previousStates {
		for _, target := range targetStates {
			t.Run(string(previous)+"_to_"+string(target), func(t *testing.T) {
				backend := newRSVPServiceBackend(previous, false)
				clock := &rsvpServiceClockFake{now: sharedUTC}
				ids := &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}
				service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, ids, rsvpServiceSerializerUnused{})
				result, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(target)))
				if err != nil {
					t.Fatalf("RespondToCalendarSeries() error = %v", err)
				}
				if result.Replayed() || result.Response().Status() != 200 {
					t.Fatalf("result replay:%t status:%d", result.Replayed(), result.Response().Status())
				}
				if !reflect.DeepEqual(backend.order, []string{"scope", "reserve", "resolve", "persist", "complete"}) {
					t.Fatalf("fresh transaction order = %v", backend.order)
				}
				if clock.calls != 1 || ids.calls != 1 || backend.persistCalls != 1 || backend.completeCalls != 1 || backend.completionEventID != rsvpServiceEventID {
					t.Fatalf("calls clock:%d ids:%d persist:%d complete:%d completion event:%s", clock.calls, ids.calls, backend.persistCalls, backend.completeCalls, backend.completionEventID)
				}
				if len(backend.persisted) != 1 {
					t.Fatalf("persisted commands = %d", len(backend.persisted))
				}
				persisted := backend.persisted[0]
				attendee := persisted.Attendee()
				if persisted.ExpectedVersion() != 2 || persisted.EventSequence() != 0 || attendee.Status != target || attendee.Version != 3 || attendee.RespondedAt == nil || !attendee.RespondedAt.Equal(sharedUTC.UTC()) {
					t.Fatalf("persisted attendee = %#v, expected:%d sequence:%d", attendee, persisted.ExpectedVersion(), persisted.EventSequence())
				}
				if reply, found := persisted.Reply(); found || reply != (calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor{}) {
					t.Fatalf("internal reply = %#v found:%t", reply, found)
				}
				assertRSVPServiceResponse(t, result.Response(), attendee, "not_required")
				assertRSVPServiceEnvelope(t, persisted.Outbox(), rsvpServiceApplicationID, sharedUTC.UTC(), previous, attendee, "not_required")
			})
		}
	}
}

func TestEventServiceRespondToCalendarSeries_ExternalOrganizerAddsExactReplyDescriptor(t *testing.T) {
	backend := newRSVPServiceBackend(calendar.RSVPNeedsAction, true)
	backend.authorization.Attendee.DisplayName = ""
	now := time.Date(2033, time.April, 5, 6, 7, 8, 0, time.UTC)
	service := calendar.NewEventService(
		&rsvpServiceRepositoryFake{backend: backend},
		&rsvpServiceClockFake{now: now},
		&rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}},
		rsvpServiceSerializerUnused{},
	)
	result, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPDeclined)))
	if err != nil {
		t.Fatalf("RespondToCalendarSeries() error = %v", err)
	}
	persisted := backend.persisted[0]
	attendee := persisted.Attendee()
	reply, found := persisted.Reply()
	if !found || reply != (calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor{AttendeeID: rsvpServiceAttendeeID, EventID: rsvpServiceEventID, EventSequence: 0, AttendeeVersion: 3, Status: calendar.RSVPDeclined}) {
		t.Fatalf("external reply = %#v found:%t", reply, found)
	}
	assertRSVPServiceResponse(t, result.Response(), attendee, "pending")
	assertRSVPServiceEnvelope(t, persisted.Outbox(), rsvpServiceApplicationID, now, calendar.RSVPNeedsAction, attendee, "pending")
}

func TestEventServiceRespondToCalendarSeries_ReplaysExactBytesAndRejectsChangedFingerprint(t *testing.T) {
	backend := newRSVPServiceBackend(calendar.RSVPNeedsAction, true)
	clock := &rsvpServiceClockFake{now: time.Date(2033, time.May, 6, 7, 8, 9, 0, time.UTC)}
	ids := &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}
	service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, ids, rsvpServiceSerializerUnused{})
	first, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
	if err != nil {
		t.Fatalf("first RespondToCalendarSeries() error = %v", err)
	}
	second, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
	if err != nil {
		t.Fatalf("replay RespondToCalendarSeries() error = %v", err)
	}
	if !second.Replayed() || !reflect.DeepEqual(first.Response().JSON(), second.Response().JSON()) || clock.calls != 1 || ids.calls != 1 || backend.resolveCalls != 1 || backend.persistCalls != 1 || backend.completeCalls != 1 {
		t.Fatalf("replay calls clock:%d ids:%d resolve:%d persist:%d complete:%d first:%q second:%q", clock.calls, ids.calls, backend.resolveCalls, backend.persistCalls, backend.completeCalls, first.Response().JSON(), second.Response().JSON())
	}

	_, err = service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPDeclined)))
	conflict := requireRSVPServiceConflict(t, err, calendar.CalendarEventSeriesRSVPIdempotencyKeyReused)
	if _, found := conflict.EventID(); found {
		t.Fatalf("reused-key conflict exposed event id: %#v", conflict)
	}
	if _, found := conflict.CurrentVersion(); found {
		t.Fatalf("reused-key conflict exposed current version: %#v", conflict)
	}
	if clock.calls != 1 || ids.calls != 1 || backend.resolveCalls != 1 || backend.persistCalls != 1 || backend.completeCalls != 1 {
		t.Fatalf("reused-key reached downstream calls clock:%d ids:%d resolve:%d persist:%d complete:%d", clock.calls, ids.calls, backend.resolveCalls, backend.persistCalls, backend.completeCalls)
	}
}

func TestEventServiceRespondToCalendarSeries_UsesExactReceiptHashes(t *testing.T) {
	backend := newRSVPServiceBackend(calendar.RSVPNeedsAction, false)
	input := rsvpServiceInput(calendar.RSVPTentative)
	service := calendar.NewEventService(
		&rsvpServiceRepositoryFake{backend: backend},
		&rsvpServiceClockFake{now: time.Date(2033, time.June, 7, 8, 9, 10, 0, time.UTC)},
		&rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}},
		rsvpServiceSerializerUnused{},
	)
	if _, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, input)); err != nil {
		t.Fatalf("RespondToCalendarSeries() error = %v", err)
	}
	if backend.reservation.KeyHash() != calendar.MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))) {
		t.Fatalf("key hash = %x", backend.reservation.KeyHash())
	}
	canonical, err := json.Marshal(struct {
		Version                 int                 `json:"version"`
		EventID                 uuid.UUID           `json:"event_id"`
		ExpectedAttendeeVersion uint                `json:"expected_attendee_version"`
		Status                  calendar.RSVPStatus `json:"status"`
	}{Version: 1, EventID: input.EventID, ExpectedAttendeeVersion: input.ExpectedAttendeeVersion, Status: input.Status})
	if err != nil {
		t.Fatalf("canonical hash JSON error = %v", err)
	}
	if backend.reservation.RequestHash() != calendar.MutationReceiptHash(sha256.Sum256(canonical)) {
		t.Fatalf("request hash = %x, canonical:%s", backend.reservation.RequestHash(), canonical)
	}
}

func TestEventServiceRespondToCalendarSeries_RecoversOnlyTypedConflictsAndMasksState(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		mutate        func(*rsvpServiceBackendFake)
		want          error
		conflict      calendar.CalendarEventSeriesRSVPConflictKind
		wantEventID   bool
		wantVersion   bool
		wantResolve   int
		wantClock     int
		wantIDs       int
		wantPersist   int
	}{
		{name: "stale attendee version", mutate: func(backend *rsvpServiceBackendFake) { backend.authorization.Attendee.Version = 9 }, want: calendar.ErrEventConflict, conflict: calendar.CalendarEventSeriesRSVPAttendeeVersionConflict, wantEventID: true, wantVersion: true, wantResolve: 1},
		{name: "cancelled hostile authoritative state", mutate: func(backend *rsvpServiceBackendFake) {
			backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled
			backend.authorization.Current.Title = string([]byte{0xff})
			backend.authorization.Current.Exdates = make([]string, 1001)
		}, want: calendar.ErrEventNotFound, wantResolve: 1},
		{name: "foreign event", mutate: func(backend *rsvpServiceBackendFake) { backend.authorization.Current.EventID = rsvpServiceOtherID }, want: calendar.ErrEventNotFound, wantResolve: 1},
		{name: "foreign attendee actor", mutate: func(backend *rsvpServiceBackendFake) { backend.authorization.Attendee.UserID = rsvpServiceUUID(rsvpServiceOtherID) }, want: calendar.ErrEventNotFound, wantResolve: 1},
		{name: "malformed same scope attendee", mutate: func(backend *rsvpServiceBackendFake) { backend.authorization.Attendee.Email = " Attendee@Example.Test " }, want: calendar.ErrEventInvalid, wantResolve: 1},
		{name: "raw provider error", mutate: func(backend *rsvpServiceBackendFake) { backend.persistError = errors.New("provider rejected attendee@example.test idempotency secret") }, want: calendar.ErrEventInvalid, wantResolve: 1, wantClock: 1, wantIDs: 1, wantPersist: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newRSVPServiceBackend(calendar.RSVPAccepted, false)
			testCase.mutate(backend)
			clock := &rsvpServiceClockFake{now: time.Date(2033, time.July, 8, 9, 10, 11, 0, time.UTC)}
			ids := &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}
			service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, ids, rsvpServiceSerializerUnused{})
			_, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("RespondToCalendarSeries() error = %v, want %v", err, testCase.want)
			}
			if strings.Contains(err.Error(), "attendee@example.test") || strings.Contains(err.Error(), "idempotency") || strings.Contains(err.Error(), rsvpServiceEventID.String()) {
				t.Fatalf("unsafe error = %q", err)
			}
			if testCase.conflict != "" {
				conflict := requireRSVPServiceConflict(t, err, testCase.conflict)
				eventID, eventFound := conflict.EventID()
				version, versionFound := conflict.CurrentVersion()
				if eventFound != testCase.wantEventID || versionFound != testCase.wantVersion || eventFound && eventID != rsvpServiceEventID || versionFound && version != 9 {
					t.Fatalf("conflict recovery event:%s/%t version:%d/%t", eventID, eventFound, version, versionFound)
				}
			}
			if backend.resolveCalls != testCase.wantResolve || clock.calls != testCase.wantClock || ids.calls != testCase.wantIDs || backend.completeCalls != 0 {
				t.Fatalf("blocked path calls resolve:%d clock:%d ids:%d complete:%d", backend.resolveCalls, clock.calls, ids.calls, backend.completeCalls)
			}
			if backend.persistCalls != testCase.wantPersist {
				t.Fatalf("blocked path persisted %d commands, want %d", backend.persistCalls, testCase.wantPersist)
			}
		})
	}
}

func TestEventServiceRespondToCalendarSeries_RejectsMismatchedPersistenceEchoAndRollsBack(t *testing.T) {
	backend := newRSVPServiceBackend(calendar.RSVPNeedsAction, true)
	backend.resultOverride = func(command calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error) {
		attendee := command.Attendee()
		attendee.Version++
		return calendar.NewCalendarEventSeriesRSVPPersistenceResult(attendee, command.EventSequence(), command.Outbox(), nil)
	}
	clock := &rsvpServiceClockFake{now: time.Date(2033, time.August, 9, 10, 11, 12, 0, time.UTC)}
	service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}, rsvpServiceSerializerUnused{})
	_, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("mismatched echo error = %v, want invalid", err)
	}
	if len(backend.persisted) != 0 || backend.receipt != nil || backend.completeCalls != 0 || clock.calls != 1 {
		t.Fatalf("mismatched echo rollback persisted:%d receipt:%#v complete:%d clock:%d", len(backend.persisted), backend.receipt, backend.completeCalls, clock.calls)
	}
}

func TestEventServiceRespondToCalendarSeries_RejectsTypedNilDependenciesBeforeCalls(t *testing.T) {
	var typedRepository *rsvpServiceRepositoryFake
	var typedClock *rsvpServiceClockFake
	var typedIDs *rsvpServiceIDGeneratorFake
	now := time.Date(2033, time.November, 12, 13, 14, 15, 0, time.UTC)
	for _, testCase := range []struct {
		name       string
		repository calendar.EventRepository
		clock      calendar.EventClock
		ids        calendar.EventIDGenerator
	}{
		{name: "repository", repository: typedRepository, clock: &rsvpServiceClockFake{now: now}, ids: &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}},
		{name: "clock", repository: &rsvpServiceRepositoryFake{backend: newRSVPServiceBackend(calendar.RSVPNeedsAction, false)}, clock: typedClock, ids: &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}},
		{name: "id generator", repository: &rsvpServiceRepositoryFake{backend: newRSVPServiceBackend(calendar.RSVPNeedsAction, false)}, clock: &rsvpServiceClockFake{now: now}, ids: typedIDs},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service := calendar.NewEventService(testCase.repository, testCase.clock, testCase.ids, rsvpServiceSerializerUnused{})
			_, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
			if !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("typed-nil %s error = %v, want invalid", testCase.name, err)
			}
			if repository, ok := testCase.repository.(*rsvpServiceRepositoryFake); ok && repository != nil && len(repository.backend.order) != 0 {
				t.Fatalf("typed-nil %s reached backend: %v", testCase.name, repository.backend.order)
			}
		})
	}
}

func TestEventServiceRespondToCalendarSeries_SafeErrorTraversalRejectsHostileChains(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
	}{
		{name: "typed nil", err: rsvpServiceTypedNilErrorValue()},
		{name: "self cycle", err: &rsvpServiceSelfCycleError{}},
		{name: "multi unwrap", err: rsvpServiceMultiUnwrapError{children: []error{errors.New("provider attendee@example.test"), errors.New("idempotency raw key")}}},
		{name: "errors join", err: errors.Join(errors.New("provider attendee@example.test"), errors.New("idempotency raw key"))},
		{name: "panicking unwrap", err: rsvpServicePanickingUnwrapError{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newRSVPServiceBackend(calendar.RSVPNeedsAction, false)
			backend.persistError = testCase.err
			clock := &rsvpServiceClockFake{now: time.Date(2033, time.September, 10, 11, 12, 13, 0, time.UTC)}
			service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}, rsvpServiceSerializerUnused{})
			_, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPAccepted)))
			if !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("hostile %s error = %v, want invalid", testCase.name, err)
			}
			if strings.Contains(err.Error(), "provider") || strings.Contains(err.Error(), "attendee@example.test") || strings.Contains(err.Error(), "idempotency") {
				t.Fatalf("hostile %s leaked error content: %q", testCase.name, err)
			}
			if clock.calls != 1 || backend.completeCalls != 0 {
				t.Fatalf("hostile %s calls clock:%d complete:%d", testCase.name, clock.calls, backend.completeCalls)
			}
		})
	}
}

func TestEventServiceRespondToCalendarSeries_IsolatesCapturedAuthorizationPersistenceAndReplayBytes(t *testing.T) {
	backend := newRSVPServiceBackend(calendar.RSVPAccepted, true)
	backend.authorization.Current.Exdates = []string{"2033-01-03T04:05:06Z"}
	clock := &rsvpServiceClockFake{now: time.Date(2033, time.October, 11, 12, 13, 14, 0, time.UTC)}
	service := calendar.NewEventService(&rsvpServiceRepositoryFake{backend: backend}, clock, &rsvpServiceIDGeneratorFake{ids: []uuid.UUID{rsvpServiceApplicationID}}, rsvpServiceSerializerUnused{})
	command := mustRSVPServiceCommand(t, rsvpServiceInput(calendar.RSVPDeclined))
	first, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, command)
	if err != nil {
		t.Fatalf("fresh RespondToCalendarSeries() error = %v", err)
	}
	wantBytes := first.Response().JSON()
	persisted := backend.persisted[0]
	wantAttendee := persisted.Attendee()

	*backend.authorization.Attendee.UserID = rsvpServiceOtherID
	*backend.authorization.Attendee.RespondedAt = backend.authorization.Attendee.RespondedAt.Add(time.Hour)
	backend.authorization.Current.Exdates[0] = "attacker-controlled"
	returnedCommandAttendee := persisted.Attendee()
	*returnedCommandAttendee.UserID = rsvpServiceOtherID
	*returnedCommandAttendee.RespondedAt = returnedCommandAttendee.RespondedAt.Add(time.Hour)
	returnedResultAttendee := backend.persistedResult.Attendee()
	*returnedResultAttendee.UserID = rsvpServiceOtherID
	*returnedResultAttendee.RespondedAt = returnedResultAttendee.RespondedAt.Add(time.Hour)
	returnedResponseBytes := first.Response().JSON()
	returnedResponseBytes[0] = '!'
	returnedEnvelopeBytes, err := persisted.Outbox().Envelope().MarshalJSON()
	if err != nil {
		t.Fatalf("Envelope.MarshalJSON() error = %v", err)
	}
	returnedEnvelopeBytes[0] = '!'

	if got := persisted.Attendee(); got.UserID == nil || *got.UserID != rsvpServiceAttendeeUserID(wantAttendee) || got.RespondedAt == nil || !got.RespondedAt.Equal(*wantAttendee.RespondedAt) {
		t.Fatalf("persistence command leaked aliases: %#v", got)
	}
	if got := backend.persistedResult.Attendee(); got.UserID == nil || *got.UserID != rsvpServiceAttendeeUserID(wantAttendee) || got.RespondedAt == nil || !got.RespondedAt.Equal(*wantAttendee.RespondedAt) {
		t.Fatalf("persistence result leaked aliases: %#v", got)
	}
	second, err := service.RespondToCalendarSeries(context.Background(), rsvpServiceActorID, command)
	if err != nil {
		t.Fatalf("replay RespondToCalendarSeries() error = %v", err)
	}
	if !second.Replayed() || !reflect.DeepEqual(second.Response().JSON(), wantBytes) || !reflect.DeepEqual(first.Response().JSON(), wantBytes) {
		t.Fatalf("response/replay bytes changed fresh:%q replay:%q want:%q", first.Response().JSON(), second.Response().JSON(), wantBytes)
	}
}

func rsvpServiceInput(status calendar.RSVPStatus) calendar.CalendarEventSeriesRSVPInput {
	return calendar.CalendarEventSeriesRSVPInput{IdempotencyKey: "rsvp-service-key", EventID: rsvpServiceEventID, ExpectedAttendeeVersion: 2, Status: status}
}

func mustRSVPServiceCommand(t *testing.T, input calendar.CalendarEventSeriesRSVPInput) calendar.CalendarEventSeriesRSVPCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventSeriesRSVPCommand(input)
	if err != nil {
		t.Fatalf("NewCalendarEventSeriesRSVPCommand() error = %v", err)
	}
	return command
}

func newRSVPServiceBackend(previous calendar.RSVPStatus, externalOrganizer bool) *rsvpServiceBackendFake {
	var organizerID *uuid.UUID
	if !externalOrganizer {
		organizerID = rsvpServiceUUID(rsvpServiceActorID)
	}
	respondedAt := (*time.Time)(nil)
	if previous != calendar.RSVPNeedsAction {
		respondedAt = rsvpServiceTime(time.Date(2033, time.January, 2, 3, 4, 5, 0, time.UTC))
	}
	return &rsvpServiceBackendFake{
		organizationID: rsvpServiceOrganizationID,
		authorization: calendar.CalendarEventSeriesRSVPAuthorization{
			Current: calendar.CalendarEventCommandInput{
				EventID: rsvpServiceEventID, OrganizationID: rsvpServiceOrganizationID, OwnerID: rsvpServiceActorID,
				UID: "hub-" + rsvpServiceEventID.String(), Title: "RSVP event", Description: "", Location: "",
				Temporal: calendar.CalendarEvent{StartsAt: time.Date(2033, time.January, 3, 4, 5, 6, 0, time.UTC), EndsAt: time.Date(2033, time.January, 3, 5, 5, 6, 0, time.UTC), Timezone: "UTC"},
				Sequence: 0, Version: 7, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed,
				OrganizerID: organizerID, OrganizerEmail: "organizer@example.test",
			},
			Attendee: calendar.CalendarEventSeriesRSVPAttendeeSnapshot{
				AttendeeID: rsvpServiceAttendeeID, EventID: rsvpServiceEventID, UserID: rsvpServiceUUID(rsvpServiceActorID), Email: "attendee@example.test", DisplayName: "Attendee", Role: calendar.CalendarEventAttendeeRoleRequired,
				Status: previous, RespondedAt: respondedAt, Version: 2,
			},
		},
	}
}

type rsvpServiceRepositoryFake struct {
	backend *rsvpServiceBackendFake
}

func (repository *rsvpServiceRepositoryFake) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	snapshot := repository.backend.snapshot()
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil {
		repository.backend.restore(snapshot)
	}
	return err
}

type rsvpServiceReceipt struct {
	keyHash     calendar.MutationReceiptHash
	requestHash calendar.MutationReceiptHash
	response    *calendar.MutationReceiptResponse
}

type rsvpServiceBackendSnapshot struct {
	receipt   *rsvpServiceReceipt
	persisted []calendar.CalendarEventSeriesRSVPPersistenceCommand
}

type rsvpServiceBackendFake struct {
	organizationID uuid.UUID
	authorization  calendar.CalendarEventSeriesRSVPAuthorization
	reservation    calendar.MutationReceiptReservation
	receipt        *rsvpServiceReceipt
	persisted      []calendar.CalendarEventSeriesRSVPPersistenceCommand
	persistedResult calendar.CalendarEventSeriesRSVPPersistenceResult
	resultOverride func(calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error)
	persistError   error
	order          []string
	resolveCalls   int
	persistCalls   int
	completeCalls  int
	completionEventID uuid.UUID
}

func (backend *rsvpServiceBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	backend.order = append(backend.order, "scope")
	if actorID != rsvpServiceActorID {
		return uuid.Nil, calendar.ErrEventForbidden
	}
	return backend.organizationID, nil
}

func (backend *rsvpServiceBackendFake) ReserveMutationReceipt(_ context.Context, _ calendar.EventMutationScope, reservation calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	backend.order = append(backend.order, "reserve")
	backend.reservation = reservation
	if backend.receipt != nil {
		if backend.receipt.keyHash != reservation.KeyHash() || backend.receipt.requestHash != reservation.RequestHash() || backend.receipt.response == nil {
			return calendar.MutationReceiptReservationResult{}, calendar.ErrEventConflict
		}
		return calendar.NewReplayedMutationReceiptResult(rsvpServiceReceiptID, *backend.receipt.response)
	}
	backend.receipt = &rsvpServiceReceipt{keyHash: reservation.KeyHash(), requestHash: reservation.RequestHash()}
	return calendar.NewReservedMutationReceiptResult(rsvpServiceReceiptID)
}

func (backend *rsvpServiceBackendFake) CompleteMutationReceipt(_ context.Context, _ calendar.EventMutationScope, completion calendar.MutationReceiptCompletion) error {
	backend.order = append(backend.order, "complete")
	backend.completeCalls++
	eventID, found := completion.EventID()
	if backend.receipt == nil || completion.ReceiptID() != rsvpServiceReceiptID || !found || eventID != rsvpServiceEventID {
		return calendar.ErrEventInvalid
	}
	backend.completionEventID = eventID
	response := completion.Response()
	backend.receipt.response = &response
	return nil
}

func (backend *rsvpServiceBackendFake) ResolveCalendarEventSeriesRSVP(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventSeriesRSVPAuthorizationRequest) (calendar.CalendarEventSeriesRSVPAuthorization, error) {
	backend.order = append(backend.order, "resolve")
	backend.resolveCalls++
	if request.EventID != rsvpServiceEventID {
		return calendar.CalendarEventSeriesRSVPAuthorization{}, calendar.ErrEventNotFound
	}
	return backend.authorization, nil
}

func (backend *rsvpServiceBackendFake) PersistCalendarEventSeriesRSVP(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error) {
	backend.order = append(backend.order, "persist")
	backend.persistCalls++
	if backend.persistError != nil {
		return calendar.CalendarEventSeriesRSVPPersistenceResult{}, backend.persistError
	}
	if backend.resultOverride != nil {
		return backend.resultOverride(command)
	}
	backend.persisted = append(backend.persisted, command)
	reply, found := command.Reply()
	var replyPointer *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
	if found {
		replyPointer = &reply
	}
	result, err := calendar.NewCalendarEventSeriesRSVPPersistenceResult(command.Attendee(), command.EventSequence(), command.Outbox(), replyPointer)
	if err == nil {
		backend.persistedResult = result
	}
	return result, err
}

func (backend *rsvpServiceBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) {
	return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid
}
func (backend *rsvpServiceBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpServiceBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpServiceBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpServiceBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error {
	return calendar.ErrEventInvalid
}

func (backend *rsvpServiceBackendFake) snapshot() rsvpServiceBackendSnapshot {
	snapshot := rsvpServiceBackendSnapshot{persisted: append([]calendar.CalendarEventSeriesRSVPPersistenceCommand(nil), backend.persisted...)}
	if backend.receipt != nil {
		receipt := *backend.receipt
		if receipt.response != nil {
			response := *receipt.response
			receipt.response = &response
		}
		snapshot.receipt = &receipt
	}
	return snapshot
}

func (backend *rsvpServiceBackendFake) restore(snapshot rsvpServiceBackendSnapshot) {
	backend.persisted = snapshot.persisted
	backend.receipt = snapshot.receipt
}

type rsvpServiceClockFake struct {
	now   time.Time
	calls int
}

func (clock *rsvpServiceClockFake) Now() time.Time {
	clock.calls++
	return clock.now
}

type rsvpServiceIDGeneratorFake struct {
	ids   []uuid.UUID
	calls int
}

func (generator *rsvpServiceIDGeneratorFake) NewEventID() uuid.UUID {
	if generator.calls >= len(generator.ids) {
		return uuid.Nil
	}
	id := generator.ids[generator.calls]
	generator.calls++
	return id
}

type rsvpServiceSerializerUnused struct{}

func (rsvpServiceSerializerUnused) Serialize(calendar.ITIPMessage) ([]byte, error) {
	panic("series RSVP must not serialize or send iTIP outside its atomic persistence port")
}

type rsvpServiceResponseBody struct {
	ID                  uuid.UUID                   `json:"id"`
	UserID              *uuid.UUID                  `json:"user_id"`
	Email               string                      `json:"email"`
	DisplayName         *string                     `json:"display_name"`
	Role                calendar.CalendarEventAttendeeRole `json:"role"`
	RSVPStatus          calendar.RSVPStatus         `json:"rsvp_status"`
	RecurrenceID        *string                     `json:"recurrence_id"`
	RespondedAt         *time.Time                  `json:"responded_at"`
	Version             uint                        `json:"version"`
	ReplyDeliveryStatus string                      `json:"reply_delivery_status"`
}

func assertRSVPServiceResponse(t *testing.T, response calendar.MutationReceiptResponse, attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot, deliveryStatus string) {
	t.Helper()
	if response.Status() != 200 {
		t.Fatalf("response status = %d, want 200", response.Status())
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(response.JSON(), &fields); err != nil {
		t.Fatalf("response JSON = %q: %v", response.JSON(), err)
	}
	if len(fields) != 10 {
		t.Fatalf("response fields = %v, want ten canonical attendee fields", mapsRSVPServiceKeys(fields))
	}
	var displayName *string
	if attendee.DisplayName != "" {
		value := attendee.DisplayName
		displayName = &value
	}
	want, err := json.Marshal(rsvpServiceResponseBody{ID: attendee.AttendeeID, UserID: attendee.UserID, Email: attendee.Email, DisplayName: displayName, Role: attendee.Role, RSVPStatus: attendee.Status, RecurrenceID: nil, RespondedAt: attendee.RespondedAt, Version: attendee.Version, ReplyDeliveryStatus: deliveryStatus})
	if err != nil {
		t.Fatalf("expected response JSON error = %v", err)
	}
	if !reflect.DeepEqual(response.JSON(), want) {
		t.Fatalf("response JSON = %s, want %s", response.JSON(), want)
	}
}

func assertRSVPServiceEnvelope(t *testing.T, descriptor calendar.EventOutboxDescriptor, applicationID uuid.UUID, occurredAt time.Time, previous calendar.RSVPStatus, attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot, deliveryStatus string) {
	t.Helper()
	envelope := descriptor.Envelope()
	if envelope.Subject() != "calendar.rsvp.changed.v1" || envelope.EventID() != applicationID {
		t.Fatalf("envelope subject:%q event:%s", envelope.Subject(), envelope.EventID())
	}
	encoded, err := envelope.MarshalJSON()
	if err != nil {
		t.Fatalf("Envelope.MarshalJSON() error = %v", err)
	}
	var outer struct {
		Subject        string          `json:"subject"`
		EventID        uuid.UUID       `json:"event_id"`
		OccurredAt     time.Time       `json:"occurred_at"`
		OrganizationID uuid.UUID       `json:"organization_id"`
		ActorID        *uuid.UUID      `json:"actor_id"`
		CorrelationID  string          `json:"correlation_id"`
		Payload        json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &outer); err != nil {
		t.Fatalf("envelope JSON = %q: %v", encoded, err)
	}
	if outer.EventID != applicationID || !outer.OccurredAt.Equal(occurredAt) || outer.OrganizationID != rsvpServiceOrganizationID || outer.ActorID == nil || *outer.ActorID != rsvpServiceActorID || outer.CorrelationID != rsvpServiceEventID.String() {
		t.Fatalf("envelope binding = %#v", outer)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(outer.Payload, &payload); err != nil {
		t.Fatalf("payload JSON = %q: %v", outer.Payload, err)
	}
	wantKeys := map[string]struct{}{"event_id": {}, "attendee_id": {}, "attendee_email": {}, "previous_status": {}, "new_status": {}, "version": {}, "recurrence_id": {}, "reply_delivery_status": {}}
	if len(payload) != len(wantKeys) {
		t.Fatalf("payload keys = %v", mapsRSVPServiceKeys(payload))
	}
	for key := range wantKeys {
		if _, found := payload[key]; !found {
			t.Fatalf("payload is missing %q: %v", key, mapsRSVPServiceKeys(payload))
		}
	}
	if _, found := payload["uid"]; found {
		t.Fatal("RSVP payload included forbidden uid")
	}
	if _, found := payload["sequence"]; found {
		t.Fatal("RSVP payload included forbidden sequence")
	}
	var decoded struct {
		EventID             uuid.UUID           `json:"event_id"`
		AttendeeID          uuid.UUID           `json:"attendee_id"`
		AttendeeEmail       string              `json:"attendee_email"`
		PreviousStatus      calendar.RSVPStatus `json:"previous_status"`
		NewStatus           calendar.RSVPStatus `json:"new_status"`
		Version             uint                `json:"version"`
		RecurrenceID        *string             `json:"recurrence_id"`
		ReplyDeliveryStatus string              `json:"reply_delivery_status"`
	}
	if err := json.Unmarshal(outer.Payload, &decoded); err != nil {
		t.Fatalf("decoded payload error = %v", err)
	}
	if decoded.EventID != rsvpServiceEventID || decoded.AttendeeID != attendee.AttendeeID || decoded.AttendeeEmail != attendee.Email || decoded.PreviousStatus != previous || decoded.NewStatus != attendee.Status || decoded.Version != attendee.Version || decoded.RecurrenceID != nil || decoded.ReplyDeliveryStatus != deliveryStatus {
		t.Fatalf("payload = %#v", decoded)
	}
}

func requireRSVPServiceConflict(t *testing.T, err error, kind calendar.CalendarEventSeriesRSVPConflictKind) calendar.CalendarEventSeriesRSVPConflictError {
	t.Helper()
	var conflict calendar.CalendarEventSeriesRSVPConflictError
	if !errors.As(err, &conflict) || conflict.Kind() != kind || conflict.Error() != "calendar event series rsvp conflict" || !errors.Is(conflict, calendar.ErrEventConflict) {
		t.Fatalf("typed conflict = %#v error:%v kind:%q", conflict, err, kind)
	}
	if strings.Contains(conflict.Error(), rsvpServiceEventID.String()) || strings.Contains(conflict.Error(), "attendee@example.test") || strings.Contains(conflict.Error(), "rsvp-service-key") {
		t.Fatalf("typed conflict leaked content: %q", conflict.Error())
	}
	return conflict
}

func assertRSVPServiceStructABI(t *testing.T, typ reflect.Type, names []string) {
	t.Helper()
	if typ.NumField() != len(names) {
		t.Fatalf("%s field count = %d, want %d", typ, typ.NumField(), len(names))
	}
	for index, name := range names {
		if typ.Field(index).Name != name || typ.Field(index).PkgPath != "" {
			t.Fatalf("%s field[%d] = %#v, want exported %q", typ, index, typ.Field(index), name)
		}
	}
}

func assertRSVPServiceExactType(t *testing.T, got, want reflect.Type) {
	t.Helper()
	if got != want {
		t.Fatalf("type = %s, want %s", got, want)
	}
}

func assertRSVPServiceValueMethod(t *testing.T, typ reflect.Type, name string, want reflect.Type) {
	t.Helper()
	method, found := typ.MethodByName(name)
	if !found || method.Type != want {
		t.Fatalf("%s.%s = %v found:%t, want %v", typ, name, method.Type, found, want)
	}
}

func mapsRSVPServiceKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func rsvpServiceUUID(value uuid.UUID) *uuid.UUID { return &value }
func rsvpServiceTime(value time.Time) *time.Time { return &value }

func rsvpServiceAttendeeUserID(attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot) uuid.UUID {
	if attendee.UserID == nil {
		return uuid.Nil
	}
	return *attendee.UserID
}

type rsvpServiceTypedNilError struct{}

func (*rsvpServiceTypedNilError) Error() string { return "typed nil provider error" }

func rsvpServiceTypedNilErrorValue() error {
	var value *rsvpServiceTypedNilError
	return value
}

type rsvpServiceSelfCycleError struct{}

func (*rsvpServiceSelfCycleError) Error() string { return "self cycle provider error" }
func (value *rsvpServiceSelfCycleError) Unwrap() error { return value }

type rsvpServiceMultiUnwrapError struct{ children []error }

func (rsvpServiceMultiUnwrapError) Error() string { return "multi unwrap provider error" }
func (value rsvpServiceMultiUnwrapError) Unwrap() []error { return value.children }

type rsvpServicePanickingUnwrapError struct{}

func (rsvpServicePanickingUnwrapError) Error() string { return "panicking provider error" }
func (rsvpServicePanickingUnwrapError) Unwrap() error { panic("provider panic") }
