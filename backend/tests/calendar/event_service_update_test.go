package calendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	updateActorID        = uuid.MustParse("51515151-5151-4515-8515-515151515151")
	updateOrganizationID = uuid.MustParse("52525252-5252-4525-8525-525252525252")
	updateEventID        = uuid.MustParse("53535353-5353-4535-8535-535353535353")
	updateAttendeeID     = uuid.MustParse("54545454-5454-4545-8545-545454545454")
	updateApplicationID  = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	cancelApplicationID  = uuid.MustParse("56565656-5656-4565-8565-565656565656")
)

var updateOrganizerAttendeeID = uuid.MustParse("57575757-5757-4575-8575-575757575757")

func TestCalendarEventUpdateCommand_RejectsAmbiguousPresenceAndFreezesInput(t *testing.T) {
	guestName := "Guest"
	attendees := []calendar.CalendarEventCreateAttendeeCandidate{{
		Email: "guest@example.test", DisplayName: &guestName, Role: calendar.CalendarEventAttendeeRoleRequired,
	}}
	validInput := calendar.CalendarEventUpdateInput{
		EventID: updateEventID, ExpectedVersion: 3,
		Title: calendar.CalendarEventUpdateString{Present: true, Value: "Updated title"},
		Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: attendees},
	}
	command, err := calendar.NewCalendarEventUpdateCommand(validInput)
	if err != nil {
		t.Fatalf("NewCalendarEventUpdateCommand() error = %v", err)
	}
	attendees[0].Email = "mutated@example.test"
	guestName = "Mutated"

	backend := newUpdateBackendFake()
	backend.authorization.ResolvedAttendees = []calendar.CalendarEventCreateResolvedAttendee{{
		Email: "guest@example.test", DisplayName: stringPointer("Guest"), Role: calendar.CalendarEventAttendeeRoleRequired,
	}}
	service := newUpdateService(backend)
	if _, err := service.UpdateCalendarEvent(context.Background(), updateActorID, command); err != nil {
		t.Fatalf("UpdateCalendarEvent() error = %v", err)
	}
	if len(backend.lastRequest.Attendees) != 1 || backend.lastRequest.Attendees[0].Email != "guest@example.test" ||
		backend.lastRequest.Attendees[0].DisplayName == nil || *backend.lastRequest.Attendees[0].DisplayName != "Guest" {
		t.Fatalf("authorization request was not frozen: %#v", backend.lastRequest.Attendees)
	}

	for _, testCase := range []struct {
		name  string
		input calendar.CalendarEventUpdateInput
	}{
		{
			name: "no patch",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3},
		},
		{
			name: "omitted string carries value",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Value: "hidden"}},
		},
		{
			name: "omitted time carries pointer",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "x"}, StartsAt: calendar.CalendarEventUpdateTime{Value: timePointer(updateStart())}},
		},
		{
			name: "omitted attendee patch carries candidate",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "x"}, Attendees: calendar.CalendarEventUpdateAttendees{Value: attendees}},
		},
		{
			name: "zero expected version",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, Title: calendar.CalendarEventUpdateString{Present: true, Value: "x"}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := calendar.NewCalendarEventUpdateCommand(testCase.input); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("NewCalendarEventUpdateCommand() error = %v, want %v", err, calendar.ErrEventInvalid)
			}
		})
	}
}

func TestEventTransactionCalendarUpdateScope_RequiresCapabilityBeforeEveryBackendCall(t *testing.T) {
	base := &updateBaseOnlyBackend{organizationID: updateOrganizationID}
	err := calendar.RunEventTransaction(base, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventUpdateScope(context.Background(), updateActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("ResolveCalendarEventUpdateScope() error = %v, want %v", err, calendar.ErrEventInvalid)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if base.organizationCalls != 0 || base.otherCalls != 0 {
		t.Fatalf("missing update capability made backend calls organization:%d other:%d", base.organizationCalls, base.otherCalls)
	}

	backend := newUpdateBackendFake()
	var retained *calendar.EventTransaction
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		scope, err := transaction.ResolveCalendarEventUpdateScope(context.Background(), updateActorID)
		if err != nil {
			t.Fatalf("ResolveCalendarEventUpdateScope() error = %v", err)
		}
		if scope.Operation() != "" || scope.EventOperation() != calendar.EventMutationCalendarUpdate || scope.ActorID() != updateActorID || scope.OrganizationID() != updateOrganizationID {
			t.Fatalf("update scope = %#v", scope)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() update scope error = %v", err)
	}
	if _, err := retained.ResolveCalendarEventUpdateScope(context.Background(), updateActorID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("retained transaction update scope error = %v, want %v", err, calendar.ErrEventTransactionClosed)
	}
}

func TestEventServiceUpdateCalendarEvent_PersistsOneAtomicSafeIntentAndTombstones(t *testing.T) {
	backend := newUpdateBackendFake()
	oldResponse := time.Date(2031, time.January, 10, 8, 0, 0, 0, time.UTC)
	backend.authorization.Attendees[0].Status = calendar.RSVPAccepted
	backend.authorization.Attendees[0].RespondedAt = &oldResponse
	backend.authorization.ResolvedAttendees = []calendar.CalendarEventCreateResolvedAttendee{{
		Email: "new@example.test", DisplayName: stringPointer("New attendee"), Role: calendar.CalendarEventAttendeeRoleOptional,
	}}
	clock := &createClockFake{now: time.Date(2031, time.January, 11, 12, 13, 14, 0, time.FixedZone("UTC+3", 3*60*60))}
	ids := &createIDGeneratorFake{ids: []uuid.UUID{uuid.MustParse("58585858-5858-4585-8585-585858585858"), updateApplicationID, cancelApplicationID}}
	service := calendar.NewEventService(&updateRepositoryFake{backend: backend}, clock, ids, createSerializerUnused{})

	start := updateStart().Add(30 * time.Minute)
	command := mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
		EventID: updateEventID, ExpectedVersion: 3, NotifyAttendees: true,
		Title: calendar.CalendarEventUpdateString{Present: true, Value: "Replanned review"},
		StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: &start},
		Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{{
			Email: "new@example.test", DisplayName: stringPointer("New attendee"), Role: calendar.CalendarEventAttendeeRoleOptional,
		}}},
	})

	result, err := service.UpdateCalendarEvent(context.Background(), updateActorID, command)
	if err != nil {
		t.Fatalf("UpdateCalendarEvent() error = %v", err)
	}
	if result.EventID() != updateEventID || result.Version() != 4 || result.Sequence() != 8 || result.Status() != calendar.CalendarEventStatusConfirmed {
		t.Fatalf("update result = event:%s version:%d sequence:%d status:%q", result.EventID(), result.Version(), result.Sequence(), result.Status())
	}
	if got := result.ChangedFields(); len(got) != 3 || got[0] != calendar.CalendarEventChangedFieldTitle || got[1] != calendar.CalendarEventChangedFieldStartsAt || got[2] != calendar.CalendarEventChangedFieldAttendees {
		t.Fatalf("changed fields = %#v", got)
	}
	if len(backend.persisted) != 1 || backend.reserveCalls != 0 || backend.completeCalls != 0 {
		t.Fatalf("atomic update writes=%d receipts=%d/%d, want 1/0/0", len(backend.persisted), backend.reserveCalls, backend.completeCalls)
	}
	persisted := backend.persisted[0]
	mutations := persisted.Attendees()
	if len(mutations) != 2 || mutations[0].Kind != calendar.CalendarEventUpdateAttendeeRemove || mutations[0].AttendeeID != updateAttendeeID || mutations[0].Version != 2 || mutations[0].RemovedAt == nil || !mutations[0].RemovedAt.Equal(clock.now.UTC()) || mutations[0].Status != calendar.RSVPAccepted || mutations[0].RespondedAt == nil || !mutations[0].RespondedAt.Equal(oldResponse) ||
		mutations[1].Kind != calendar.CalendarEventUpdateAttendeeAdd || mutations[1].Version != 1 || mutations[1].AttendeeID == uuid.Nil || mutations[1].RemovedAt != nil {
		t.Fatalf("attendee mutations = %#v", mutations)
	}
	mutations[1].Email = "caller-mutated@example.test"
	if persisted.Attendees()[1].Email == "caller-mutated@example.test" {
		t.Fatal("CalendarEventUpdatePersistenceCommand.Attendees() exposed mutable slice")
	}

	outboxes := persisted.Outbox()
	if len(outboxes) != 2 || outboxes[0].Envelope().Subject() != calendar.CalendarInviteUpdatedSubjectV2 || outboxes[0].Envelope().EventID() != updateApplicationID ||
		outboxes[1].Envelope().Subject() != calendar.CalendarInviteCancelledSubjectV2 || outboxes[1].Envelope().EventID() != cancelApplicationID {
		t.Fatalf("update outbox = %#v", outboxes)
	}
	encoded, err := json.Marshal(outboxes[0].Envelope())
	if err != nil {
		t.Fatalf("json.Marshal(update envelope) error = %v", err)
	}
	if strings.Contains(string(encoded), "Replanned review") || strings.Contains(string(encoded), "SENSITIVE") || strings.Contains(string(encoded), "iTIP") {
		t.Fatalf("safe update envelope leaked content: %s", encoded)
	}
	assertCalendarEventV2Envelope(t, outboxes[0].Envelope(), calendar.CalendarInviteUpdatedSubjectV2, updateApplicationID, updateEventID, updateActorID, 8, []uuid.UUID{mutations[1].AttendeeID})
	assertCalendarEventV2Envelope(t, outboxes[1].Envelope(), calendar.CalendarInviteCancelledSubjectV2, cancelApplicationID, updateEventID, updateActorID, 8, []uuid.UUID{updateAttendeeID})
}

func TestEventServiceUpdateCalendarEvent_RejectsCollidingNotificationIDsBeforePersistence(t *testing.T) {
	backend := newUpdateBackendFake()
	backend.authorization.ResolvedAttendees = []calendar.CalendarEventCreateResolvedAttendee{{
		Email: "new@example.test", DisplayName: stringPointer("New attendee"), Role: calendar.CalendarEventAttendeeRoleOptional,
	}}
	attendeeID := uuid.MustParse("57575757-5757-4575-8575-575757575757")
	collidingID := uuid.MustParse("58585858-5858-4585-8585-585858585858")
	service := calendar.NewEventService(
		&updateRepositoryFake{backend: backend},
		&createClockFake{now: time.Date(2031, time.January, 11, 12, 13, 14, 0, time.UTC)},
		&createIDGeneratorFake{ids: []uuid.UUID{attendeeID, collidingID, collidingID}},
		createSerializerUnused{},
	)
	_, err := service.UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
		EventID: updateEventID, ExpectedVersion: 3, NotifyAttendees: true,
		Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{{
			Email: "new@example.test", DisplayName: stringPointer("New attendee"), Role: calendar.CalendarEventAttendeeRoleOptional,
		}}},
	}))
	if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), collidingID.String()) {
		t.Fatalf("notification collision error = %v, want content-free invalid", err)
	}
	if backend.persistCalls != 0 || len(backend.persisted) != 0 {
		t.Fatalf("notification collision persisted calls/commands = %d/%d, want 0/0", backend.persistCalls, len(backend.persisted))
	}
}

func TestEventServiceUpdateCalendarEvent_NonDisclosureRecurrenceNoopAndConflictOrdering(t *testing.T) {
	sameRRULE := "FREQ=DAILY;INTERVAL=1;COUNT=2"
	sameExdate := "2031-01-21T09:00:00Z"
	sameStartDate := "2031-01-20"
	for _, testCase := range []struct {
		name string
		prepare func(*updateBackendFake)
		command calendar.CalendarEventUpdateInput
		want error
		wantPersist int
	}{
		{
			name: "missing or foreign resolver result is not found before version",
			prepare: func(backend *updateBackendFake) { backend.authorizationError = calendar.ErrEventNotFound },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 999, Title: calendar.CalendarEventUpdateString{Present: true, Value: "hidden"}},
			want: calendar.ErrEventNotFound,
		},
		{
			name: "terminal is not mutable after resolution",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "x"}},
			want: calendar.ErrEventConflict,
		},
		{
			name: "stale version is version conflict",
			prepare: func(*updateBackendFake) {},
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 99, Title: calendar.CalendarEventUpdateString{Present: true, Value: "x"}},
			want: calendar.ErrEventConflict,
		},
		{
			name: "materialized exception blocks changed recurrence identity before write",
			prepare: func(backend *updateBackendFake) { backend.authorization.HasMaterializedException = true },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Timezone: calendar.CalendarEventUpdateString{Present: true, Value: "Europe/Moscow"}},
			want: calendar.ErrEventConflict,
		},
		{
			name: "materialized exception permits same rrule",
			prepare: func(backend *updateBackendFake) {
				backend.authorization.HasMaterializedException = true
				backend.authorization.Current.RRULE = sameRRULE
			},
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, RRULE: calendar.CalendarEventUpdateRRULE{Present: true, Value: &sameRRULE}},
		},
		{
			name: "materialized exception permits same exdates",
			prepare: func(backend *updateBackendFake) {
				backend.authorization.HasMaterializedException = true
				backend.authorization.Current.Exdates = []string{sameExdate}
			},
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Exdates: calendar.CalendarEventUpdateExdates{Present: true, Value: []string{sameExdate}}},
		},
		{
			name: "materialized exception permits same starts at",
			prepare: func(backend *updateBackendFake) { backend.authorization.HasMaterializedException = true },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: timePointer(updateStart())}},
		},
		{
			name: "materialized exception permits same start date",
			prepare: func(backend *updateBackendFake) {
				backend.authorization.HasMaterializedException = true
				backend.authorization.Current.Temporal = calendar.CalendarEvent{AllDay: true, StartDate: sameStartDate, EndDateExclusive: "2031-01-22", Timezone: "UTC"}
			},
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, StartDate: calendar.CalendarEventUpdateDate{Present: true, Value: &sameStartDate}},
		},
		{
			name: "materialized exception permits same timezone",
			prepare: func(backend *updateBackendFake) { backend.authorization.HasMaterializedException = true },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Timezone: calendar.CalendarEventUpdateString{Present: true, Value: "UTC"}},
		},
		{
			name: "materialized exception permits same all day",
			prepare: func(backend *updateBackendFake) { backend.authorization.HasMaterializedException = true },
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, AllDay: calendar.CalendarEventUpdateBool{Present: true, Value: false}},
		},
		{
			name: "equal patch is a no op",
			prepare: func(*updateBackendFake) {},
			command: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "Original title"}},
			wantPersist: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			testCase.prepare(backend)
			service := newUpdateService(backend)
			result, err := service.UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, testCase.command))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("UpdateCalendarEvent() error = %v, want errors.Is(..., %v)", err, testCase.want)
			}
			if testCase.want == nil && (result.EventID() != updateEventID || result.Version() != 3 || result.Sequence() != 7 || len(result.ChangedFields()) != 0) {
				t.Fatalf("no-op result = %#v", result)
			}
			outboxCount := 0
			for _, persistence := range backend.persisted { outboxCount += len(persistence.Outbox()) }
			if len(backend.persisted) != testCase.wantPersist || backend.persistCalls != testCase.wantPersist || outboxCount != 0 {
				t.Fatalf("persisted/persist calls/outbox = %d/%d/%d, want %d/0/0", len(backend.persisted), backend.persistCalls, outboxCount, testCase.wantPersist)
			}
		})
	}
}

func TestEventServiceUpdateCalendarEvent_RejectsCorruptAuthoritativeCurrentBeforeWrites(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		prepare func(*updateBackendFake)
	}{
		{
			name: "title exceeds create bound",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.Title = strings.Repeat("x", 501) },
		},
		{
			name: "description contains control",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.Description = "unsafe\x00description" },
		},
		{
			name: "location is invalid utf8",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.Location = string([]byte{0xff}) },
		},
		{
			name: "malformed hub uid",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.UID = "not-a-hub-uid" },
		},
		{
			name: "invalid organizer email",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.OrganizerEmail = "not-an-email" },
		},
		{
			name: "noncanonical organizer email",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.OrganizerEmail = "Owner@Example.test" },
		},
		{
			name: "noncanonical rrule",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.RRULE = "COUNT=2;FREQ=DAILY" },
		},
		{
			name: "invalid rrule",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.RRULE = "FREQ=DAILY;COUNT=0" },
		},
		{
			name: "rrule exceeds bound",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.RRULE = strings.Repeat("x", 1001) },
		},
		{
			name: "noncanonical exdates",
			prepare: func(backend *updateBackendFake) {
				backend.authorization.Current.Exdates = []string{"2031-01-21T09:00:00Z", "2031-01-20T09:00:00Z"}
			},
		},
		{
			name: "invalid exdate",
			prepare: func(backend *updateBackendFake) { backend.authorization.Current.Exdates = []string{"not-an-exdate"} },
		},
		{
			name: "exdates exceed bound",
			prepare: func(backend *updateBackendFake) {
				exdates := make([]string, 1001)
				for index := range exdates { exdates[index] = "2031-01-20T09:00:00Z" }
				backend.authorization.Current.Exdates = exdates
			},
		},
		{
			name: "attendee has invalid identifier",
			prepare: func(backend *updateBackendFake) { backend.authorization.Attendees[0].AttendeeID = uuid.Nil },
		},
		{
			name: "attendee email is noncanonical",
			prepare: func(backend *updateBackendFake) { backend.authorization.Attendees[0].Email = "Guest@Example.test" },
		},
		{
			name: "201 active attendees",
			prepare: func(backend *updateBackendFake) {
				attendees := make([]calendar.CalendarEventUpdateAttendeeSnapshot, 201)
				for index := range attendees {
					attendees[index] = calendar.CalendarEventUpdateAttendeeSnapshot{
						AttendeeID: uuid.NewMD5(updateEventID, []byte("attendee-"+strconv.Itoa(index))),
						Email: "attendee" + strconv.Itoa(index) + "@example.test", Role: calendar.CalendarEventAttendeeRoleRequired,
						Status: calendar.RSVPNeedsAction, Version: 1,
					}
				}
				backend.authorization.Attendees = append(
					[]calendar.CalendarEventUpdateAttendeeSnapshot{updateOrganizerAttendeeSnapshot()},
					attendees...,
				)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			testCase.prepare(backend)
			_, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
				EventID: updateEventID, ExpectedVersion: 3, Location: calendar.CalendarEventUpdateString{Present: true, Value: "Room 9"},
			}))
			if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "not-a") || backend.persistCalls != 0 || len(backend.persisted) != 0 {
				t.Fatalf("corrupt current error/persist calls/writes = %v/%d/%d", err, backend.persistCalls, len(backend.persisted))
			}
		})
	}
}

func TestEventServiceUpdateCalendarEvent_RejectsSequenceOverflowBeforeWrites(t *testing.T) {
	for _, testCase := range []struct {
		name string
		invoke func(*calendar.EventService) (calendar.CalendarEventUpdateResult, error)
	}{
		{
			name: "update",
			invoke: func(service *calendar.EventService) (calendar.CalendarEventUpdateResult, error) {
				return service.UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
					EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "Overflow"},
				}))
			},
		},
		{
			name: "cancel",
			invoke: func(service *calendar.EventService) (calendar.CalendarEventUpdateResult, error) {
				return service.CancelCalendarEvent(context.Background(), updateActorID, updateEventID, 3)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			backend.authorization.Current.Sequence = ^uint(0)
			_, err := testCase.invoke(newUpdateService(backend))
			if !errors.Is(err, calendar.ErrEventInvalid) || backend.persistCalls != 0 || len(backend.persisted) != 0 {
				t.Fatalf("sequence overflow error/persist calls/writes = %v/%d/%d", err, backend.persistCalls, len(backend.persisted))
			}
		})
	}
}

func TestEventServiceUpdateCalendarEvent_RejectsEveryPersistenceResultMismatchAndRollsBack(t *testing.T) {
	for _, testCase := range []struct {
		name string
		mutate func(calendar.CalendarEventUpdatePersistenceCommand, calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult
	}{
		{
			name: "event id",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				return mustUpdatePersistenceResult(t, uuid.MustParse("58585858-5858-4585-8585-585858585858"), command.Event().Version(), 8, updatePersistenceEcho(command))
			},
		},
		{
			name: "version",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version()+1, 8, updatePersistenceEcho(command))
			},
		},
		{
			name: "sequence",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 99, updatePersistenceEcho(command))
			},
		},
		{
			name: "missing attendee result",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 8, nil)
			},
		},
		{
			name: "extra attendee result",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				results := updatePersistenceEcho(command)
				results = append(results, calendar.CalendarEventUpdateAttendeePersistenceResult{AttendeeID: uuid.MustParse("59595959-5959-4595-8595-595959595959"), Version: 1})
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 8, results)
			},
		},
		{
			name: "attendee id echo",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				results := updatePersistenceEcho(command)
				results[0].AttendeeID = uuid.MustParse("60606060-6060-4060-8060-606060606060")
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 8, results)
			},
		},
		{
			name: "attendee version echo",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				results := updatePersistenceEcho(command)
				results[0].Version++
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 8, results)
			},
		},
		{
			name: "attendee removed echo",
			mutate: func(command calendar.CalendarEventUpdatePersistenceCommand, _ calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
				results := updatePersistenceEcho(command)
				results[0].Removed = true
				return mustUpdatePersistenceResult(t, command.Event().EventID(), command.Event().Version(), 8, results)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			backend.resultMutator = testCase.mutate
			_, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
				EventID: updateEventID, ExpectedVersion: 3, Location: calendar.CalendarEventUpdateString{Present: true, Value: "Room 7"},
			}))
			if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "SENSITIVE") {
				t.Fatalf("persistence mismatch error = %v, want safe %v", err, calendar.ErrEventInvalid)
			}
			if len(backend.persisted) != 0 {
				t.Fatalf("rollback retained %d persisted commands", len(backend.persisted))
			}
		})
	}
}

func TestEventTransactionCalendarUpdate_ClonesBackendOutputAndRejectsEscapedUse(t *testing.T) {
	backend := newUpdateBackendFake()
	var retained *calendar.EventTransaction
	var returned calendar.CalendarEventUpdateAuthorization
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		scope, err := transaction.ResolveCalendarEventUpdateScope(context.Background(), updateActorID)
		if err != nil { return err }
		returned, err = transaction.ResolveCalendarEventUpdate(context.Background(), scope, calendar.CalendarEventUpdateAuthorizationRequest{EventID: updateEventID})
		if err != nil { return err }
		returned.Current.Title = "caller mutation"
		returned.Attendees[0].Email = "caller@example.test"
		if backend.authorization.Current.Title != "Original title" || backend.authorization.Attendees[0].Email != "guest@example.test" {
			t.Fatalf("resolver output aliases backend state: %#v", backend.authorization)
		}
		command := mustUpdatePersistenceCommand(t)
		if _, err := transaction.PersistCalendarEventUpdate(context.Background(), scope, command); err != nil {
			t.Fatalf("PersistCalendarEventUpdate() error = %v", err)
		}
		return nil
	})
	if err != nil { t.Fatalf("RunEventTransaction() error = %v", err) }
	if retained == nil { t.Fatal("transaction was not retained for lifetime oracle") }
	if _, err := retained.ResolveCalendarEventUpdateScope(context.Background(), updateActorID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped resolve scope error = %v, want %v", err, calendar.ErrEventTransactionClosed)
	}
	if _, err := retained.PersistCalendarEventUpdate(context.Background(), calendar.EventMutationScope{}, mustUpdatePersistenceCommand(t)); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped persist error = %v, want %v", err, calendar.ErrEventTransactionClosed)
	}
	duplicate := []calendar.CalendarEventUpdateAttendeePersistenceResult{{AttendeeID: updateAttendeeID, Version: 2}, {AttendeeID: updateAttendeeID, Version: 2}}
	if _, err := calendar.NewCalendarEventUpdatePersistenceResult(updateEventID, 4, 8, duplicate); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("duplicate persistence result error = %v, want %v", err, calendar.ErrEventInvalid)
	}
}

func TestEventServiceCancelCalendarEvent_UsesSafeCancelledEnvelopeWithoutRecurrenceBlock(t *testing.T) {
	backend := newUpdateBackendFake()
	backend.authorization.HasMaterializedException = true
	service := calendar.NewEventService(
		&updateRepositoryFake{backend: backend},
		&createClockFake{now: time.Date(2031, time.January, 12, 1, 2, 3, 0, time.UTC)},
		&createIDGeneratorFake{ids: []uuid.UUID{cancelApplicationID}},
		createSerializerUnused{},
	)

	result, err := service.CancelCalendarEvent(context.Background(), updateActorID, updateEventID, 3)
	if err != nil {
		t.Fatalf("CancelCalendarEvent() error = %v", err)
	}
	if result.Status() != calendar.CalendarEventStatusCancelled || result.Version() != 4 || result.Sequence() != 8 || len(result.ChangedFields()) != 0 {
		t.Fatalf("cancel result = %#v", result)
	}
	if len(backend.persisted) != 1 || len(backend.persisted[0].Outbox()) != 1 {
		t.Fatalf("cancel persist/outbox = %d/%d", len(backend.persisted), len(backend.persisted[0].Outbox()))
	}
	envelope := backend.persisted[0].Outbox()[0].Envelope()
	if envelope.Subject() != calendar.CalendarInviteCancelledSubjectV2 || envelope.EventID() != cancelApplicationID {
		t.Fatalf("cancel envelope subject/id = %q/%s", envelope.Subject(), envelope.EventID())
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal(cancel envelope) error = %v", err)
	}
	if strings.Contains(string(encoded), "Original title") || strings.Contains(string(encoded), "iTIP") {
		t.Fatalf("safe cancel envelope leaked content: %s", encoded)
	}
	assertCalendarEventV2Envelope(t, envelope, calendar.CalendarInviteCancelledSubjectV2, cancelApplicationID, updateEventID, updateActorID, 8, []uuid.UUID{updateAttendeeID})
}

func TestCalendarInviteUpdateEnvelopeBuilders_ValidateCanonicalSafePayloads(t *testing.T) {
	attendees := []uuid.UUID{updateAttendeeID}
	updated, err := calendar.NewCalendarInviteUpdatedEnvelopeV1(
		updateApplicationID, time.Date(2031, time.January, 13, 1, 2, 3, 0, time.UTC), updateOrganizationID, &updateActorID, updateEventID.String(),
		calendar.CalendarInviteUpdatedPayload{EventID: updateEventID, UID: "hub-" + updateEventID.String(), Sequence: 8, AttendeeIDs: attendees, ResetRSVP: true, ChangedFields: []calendar.CalendarEventChangedField{calendar.CalendarEventChangedFieldTitle, calendar.CalendarEventChangedFieldAttendees}},
	)
	if err != nil || updated.Subject() != "calendar.invite.updated.v1" {
		t.Fatalf("NewCalendarInviteUpdatedEnvelopeV1() = %#v, %v", updated, err)
	}
	attendees[0] = uuid.Nil
	encoded, err := json.Marshal(updated)
	if err != nil || strings.Contains(string(encoded), uuid.Nil.String()) {
		t.Fatalf("updated envelope was not frozen/safe: %s, %v", encoded, err)
	}
	for _, testCase := range []struct {
		name string
		payload calendar.CalendarInviteUpdatedPayload
	}{
		{name: "duplicate attendee", payload: calendar.CalendarInviteUpdatedPayload{EventID: updateEventID, UID: "uid", Sequence: 1, AttendeeIDs: []uuid.UUID{updateAttendeeID, updateAttendeeID}, ChangedFields: []calendar.CalendarEventChangedField{calendar.CalendarEventChangedFieldTitle}}},
		{name: "unordered fields", payload: calendar.CalendarInviteUpdatedPayload{EventID: updateEventID, UID: "uid", Sequence: 1, ChangedFields: []calendar.CalendarEventChangedField{calendar.CalendarEventChangedFieldTimezone, calendar.CalendarEventChangedFieldTitle}}},
		{name: "missing fields", payload: calendar.CalendarInviteUpdatedPayload{EventID: updateEventID, UID: "uid", Sequence: 1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := calendar.NewCalendarInviteUpdatedEnvelopeV1(updateApplicationID, time.Now().UTC(), updateOrganizationID, nil, "correlation", testCase.payload); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("NewCalendarInviteUpdatedEnvelopeV1() error = %v, want %v", err, calendar.ErrEventInvalid)
			}
		})
	}
}

func TestCalendarInviteDeliveryV2MasterProducerEnvelope_AcceptsSequenceZero(t *testing.T) {
	envelope, err := calendar.NewCalendarInviteDeliveryEnvelopeV2(
		calendar.CalendarInviteUpdatedSubjectV2, updateApplicationID, time.Date(2031, time.January, 13, 1, 2, 3, 0, time.UTC), updateOrganizationID, &updateActorID, updateEventID.String(),
		calendar.CalendarInviteDeliveryEnvelopeV2Input{EventID: updateEventID, EventSequence: 0, OrganizerID: updateActorID, AttendeeIDs: []uuid.UUID{updateAttendeeID}},
	)
	if err != nil {
		t.Fatalf("NewCalendarInviteDeliveryEnvelopeV2() error = %v", err)
	}
	assertCalendarEventV2Envelope(t, envelope, calendar.CalendarInviteUpdatedSubjectV2, updateApplicationID, updateEventID, updateActorID, 0, []uuid.UUID{updateAttendeeID})
}

func TestEventServiceUpdateCalendarEvent_RemovalOmitsEmptyUpdatedV2Intent(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input calendar.CalendarEventUpdateInput
		wantFields []calendar.CalendarEventChangedField
	}{
		{
			name: "remove only",
			input: calendar.CalendarEventUpdateInput{
				EventID: updateEventID, ExpectedVersion: 3,
				Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{}},
			},
			wantFields: []calendar.CalendarEventChangedField{calendar.CalendarEventChangedFieldAttendees},
		},
		{
			name: "mixed title and remove",
			input: calendar.CalendarEventUpdateInput{
				EventID: updateEventID, ExpectedVersion: 3,
				Title: calendar.CalendarEventUpdateString{Present: true, Value: "Changed before removal"},
				Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{}},
			},
			wantFields: []calendar.CalendarEventChangedField{calendar.CalendarEventChangedFieldTitle, calendar.CalendarEventChangedFieldAttendees},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			service := calendar.NewEventService(
				&updateRepositoryFake{backend: backend},
				&createClockFake{now: time.Date(2031, time.January, 14, 1, 2, 3, 0, time.UTC)},
				&createIDGeneratorFake{ids: []uuid.UUID{updateApplicationID, cancelApplicationID}},
				createSerializerUnused{},
			)
			result, err := service.UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, testCase.input))
			if err != nil {
				t.Fatalf("UpdateCalendarEvent() error = %v", err)
			}
			if got := result.ChangedFields(); !sameUpdateChangedFields(got, testCase.wantFields) {
				t.Fatalf("changed fields = %#v, want %#v", got, testCase.wantFields)
			}
			if len(backend.persisted) != 1 {
				t.Fatalf("persisted updates = %d, want 1", len(backend.persisted))
			}
			outbox := backend.persisted[0].Outbox()
			if len(outbox) != 1 || outbox[0].Envelope().Subject() != calendar.CalendarInviteCancelledSubjectV2 || outbox[0].Envelope().EventID() != cancelApplicationID {
				t.Fatalf("removal outbox = %#v", outbox)
			}
			assertCalendarEventV2Envelope(t, outbox[0].Envelope(), calendar.CalendarInviteCancelledSubjectV2, cancelApplicationID, updateEventID, updateActorID, 8, []uuid.UUID{updateAttendeeID})
		})
	}
}

func TestEventServiceUpdateCalendarEvent_AuthorizedNonHubIsNotMutableWithoutRecoveryLeak(t *testing.T) {
	backend := newUpdateBackendFake()
	backend.authorization.Current.Source = calendar.CalendarEventSourceMailImport
	_, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
		EventID: updateEventID, ExpectedVersion: 3, Title: calendar.CalendarEventUpdateString{Present: true, Value: "No import mutation"},
	}))
	var conflict calendar.CalendarEventUpdateConflictError
	if !errors.As(err, &conflict) || conflict.Kind() != calendar.CalendarEventUpdateNotMutable || !errors.Is(err, calendar.ErrEventConflict) {
		t.Fatalf("authorized imported event error = %v, want typed not-mutable conflict", err)
	}
	if eventID, found := conflict.EventID(); found || eventID != uuid.Nil {
		t.Fatalf("not-mutable conflict exposed event recovery identity: %s/%t", eventID, found)
	}
	if version, found := conflict.CurrentVersion(); found || version != 0 {
		t.Fatalf("not-mutable conflict exposed recovery version: %d/%t", version, found)
	}
	if len(backend.persisted) != 0 || backend.reserveCalls != 0 || backend.completeCalls != 0 {
		t.Fatalf("not-mutable mutation/receipt calls = %d/%d/%d", len(backend.persisted), backend.reserveCalls, backend.completeCalls)
	}
}

func TestEventServiceUpdateCalendarEvent_RecurrenceAffectedResetAndExceptionBoundary(t *testing.T) {
	t.Run("rrule resets exact active nonorganizer set", func(t *testing.T) {
		backend := newUpdateBackendFake()
		secondID := uuid.MustParse("61616161-6161-4161-8161-616161616161")
		responded := time.Date(2031, time.January, 15, 4, 5, 6, 0, time.UTC)
		backend.authorization.Attendees = append(backend.authorization.Attendees, calendar.CalendarEventUpdateAttendeeSnapshot{
			AttendeeID: secondID, Email: "second@example.test", DisplayName: "Second", Role: calendar.CalendarEventAttendeeRoleOptional, Status: calendar.RSVPTentative, RespondedAt: &responded, Version: 4,
		})
	rrule := "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO;COUNT=2"
		result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3, RRULE: calendar.CalendarEventUpdateRRULE{Present: true, Value: &rrule},
		}))
		if err != nil {
			t.Fatalf("RRULE UpdateCalendarEvent() error = %v", err)
		}
		if result.Sequence() != 8 || len(backend.persisted) != 1 {
			t.Fatalf("RRULE result/persist = %#v/%d", result, len(backend.persisted))
		}
		mutations := backend.persisted[0].Attendees()
		if len(mutations) != 2 {
			t.Fatalf("RRULE reset mutations = %#v", mutations)
		}
		for _, mutation := range mutations {
			if mutation.Kind != calendar.CalendarEventUpdateAttendeeReset || mutation.Status != calendar.RSVPNeedsAction || mutation.RespondedAt != nil ||
				(mutation.AttendeeID == updateAttendeeID && mutation.Version != 2) || (mutation.AttendeeID == secondID && mutation.Version != 5) {
				t.Fatalf("RRULE reset mutation = %#v", mutation)
			}
		}
		outbox := backend.persisted[0].Outbox()
		if len(outbox) != 1 { t.Fatalf("RRULE outbox count = %d, want 1", len(outbox)) }
		encoded, err := json.Marshal(outbox[0].Envelope())
		if err != nil { t.Fatalf("json.Marshal(RRULE envelope) error = %v", err) }
		if !strings.Contains(string(encoded), `"reset_rsvp":true`) || !strings.Contains(string(encoded), `"changed_fields":["rrule"]`) {
			t.Fatalf("RRULE envelope = %s", encoded)
		}
	})

	t.Run("exception blocks recurrence but permits end duration", func(t *testing.T) {
		backend := newUpdateBackendFake()
		backend.authorization.HasMaterializedException = true
		end := updateStart().Add(2 * time.Hour)
		result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3, EndsAt: calendar.CalendarEventUpdateTime{Present: true, Value: &end},
		}))
		if err != nil || result.Sequence() != 8 || len(backend.persisted) != 1 {
			t.Fatalf("end-only exception update = %#v, %v, writes=%d", result, err, len(backend.persisted))
		}
	})
}

func TestEventServiceUpdateCalendarEvent_ResetAndSilentPolicyMapping(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input calendar.CalendarEventUpdateInput
		wantSequence uint
		wantOutbox int
		wantReset bool
	}{
		{
			name: "time resets all",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: timePointer(updateStart().Add(-time.Hour))}},
			wantSequence: 8, wantOutbox: 1, wantReset: true,
		},
		{
			name: "location resets all",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Location: calendar.CalendarEventUpdateString{Present: true, Value: "Room 8"}},
			wantSequence: 8, wantOutbox: 1, wantReset: true,
		},
		{
			name: "description silent",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Description: calendar.CalendarEventUpdateString{Present: true, Value: "Silent description"}},
			wantSequence: 7, wantOutbox: 0,
		},
		{
			name: "description notify",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, NotifyAttendees: true, Description: calendar.CalendarEventUpdateString{Present: true, Value: "Notify description"}},
			wantSequence: 8, wantOutbox: 1,
		},
		{
			name: "reminder only",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, ReminderMinutes: calendar.CalendarEventUpdateInt{Present: true, Value: intPointer(30)}},
			wantSequence: 7, wantOutbox: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			backend.resultSequence = testCase.wantSequence
			result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, testCase.input))
			if err != nil || result.Version() != 4 || result.Sequence() != testCase.wantSequence || len(backend.persisted) != 1 {
				t.Fatalf("policy result/error/writes = %#v/%v/%d", result, err, len(backend.persisted))
			}
			command := backend.persisted[0]
			if len(command.Outbox()) != testCase.wantOutbox {
				t.Fatalf("outbox count = %d, want %d", len(command.Outbox()), testCase.wantOutbox)
			}
			mutations := command.Attendees()
			if testCase.wantReset {
				if len(mutations) != 1 || mutations[0].Kind != calendar.CalendarEventUpdateAttendeeReset || mutations[0].Version != 2 || mutations[0].Status != calendar.RSVPNeedsAction || mutations[0].RespondedAt != nil {
					t.Fatalf("reset mutation = %#v", mutations)
				}
			} else if len(mutations) != 0 {
				t.Fatalf("unexpected attendee mutation = %#v", mutations)
			}
		})
	}
}

func TestCalendarEventUpdateCommand_TemporalPresenceAndAttendeeBounds(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input calendar.CalendarEventUpdateInput
		want error
	}{
		{
			name: "present null timed start reaches merge validator",
			input: calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: nil}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command, err := calendar.NewCalendarEventUpdateCommand(testCase.input)
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) { t.Fatalf("NewCalendarEventUpdateCommand() error = %v, want %v", err, testCase.want) }
				return
			}
			if err != nil { t.Fatalf("NewCalendarEventUpdateCommand() error = %v", err) }
			backend := newUpdateBackendFake()
			if _, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, command); !errors.Is(err, calendar.ErrEventInvalid) || len(backend.persisted) != 0 {
				t.Fatalf("present-null timed update error/writes = %v/%d", err, len(backend.persisted))
			}
		})
	}

	t.Run("present null date is accepted as temporal patch", func(t *testing.T) {
		command, err := calendar.NewCalendarEventUpdateCommand(calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3, StartDate: calendar.CalendarEventUpdateDate{Present: true, Value: nil},
		})
		if err != nil { t.Fatalf("NewCalendarEventUpdateCommand() error = %v", err) }
		backend := newUpdateBackendFake()
		result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, command)
		if err != nil || result.Version() != 3 || len(backend.persisted) != 0 {
			t.Fatalf("present-null date no-op result/error/writes = %#v/%v/%d", result, err, len(backend.persisted))
		}
	})

	t.Run("timed to all-day transition accepts explicit null instants", func(t *testing.T) {
		startDate, endDate := "2031-01-20", "2031-01-22"
		backend := newUpdateBackendFake()
		result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3,
			StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: nil},
			EndsAt: calendar.CalendarEventUpdateTime{Present: true, Value: nil},
			StartDate: calendar.CalendarEventUpdateDate{Present: true, Value: &startDate},
			EndDateExclusive: calendar.CalendarEventUpdateDate{Present: true, Value: &endDate},
			AllDay: calendar.CalendarEventUpdateBool{Present: true, Value: true},
		}))
		wantFields := []calendar.CalendarEventChangedField{
			calendar.CalendarEventChangedFieldStartsAt,
			calendar.CalendarEventChangedFieldEndsAt,
			calendar.CalendarEventChangedFieldStartDate,
			calendar.CalendarEventChangedFieldEndDateExclusive,
			calendar.CalendarEventChangedFieldAllDay,
		}
		if err != nil || result.Sequence() != 8 || len(backend.persisted) != 1 || !sameUpdateChangedFields(result.ChangedFields(), wantFields) {
			t.Fatalf("timed/all-day result/error/writes/fields = %#v/%v/%d/%#v", result, err, len(backend.persisted), result.ChangedFields())
		}
	})

	t.Run("all-day to timed transition accepts explicit null dates", func(t *testing.T) {
		backend := newUpdateBackendFake()
		backend.authorization.Current.Temporal = calendar.CalendarEvent{
			AllDay: true, StartDate: "2031-01-20", EndDateExclusive: "2031-01-22", Timezone: "UTC",
		}
		start, end := updateStart(), updateStart().Add(time.Hour)
		result, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3,
			StartsAt: calendar.CalendarEventUpdateTime{Present: true, Value: &start},
			EndsAt: calendar.CalendarEventUpdateTime{Present: true, Value: &end},
			StartDate: calendar.CalendarEventUpdateDate{Present: true, Value: nil},
			EndDateExclusive: calendar.CalendarEventUpdateDate{Present: true, Value: nil},
			AllDay: calendar.CalendarEventUpdateBool{Present: true, Value: false},
		}))
		wantFields := []calendar.CalendarEventChangedField{
			calendar.CalendarEventChangedFieldStartsAt,
			calendar.CalendarEventChangedFieldEndsAt,
			calendar.CalendarEventChangedFieldStartDate,
			calendar.CalendarEventChangedFieldEndDateExclusive,
			calendar.CalendarEventChangedFieldAllDay,
		}
		if err != nil || result.Sequence() != 8 || len(backend.persisted) != 1 || !sameUpdateChangedFields(result.ChangedFields(), wantFields) {
			t.Fatalf("all-day/timed result/error/writes/fields = %#v/%v/%d/%#v", result, err, len(backend.persisted), result.ChangedFields())
		}
	})

	validCandidates := make([]calendar.CalendarEventCreateAttendeeCandidate, 200)
	for index := range validCandidates {
		validCandidates[index] = calendar.CalendarEventCreateAttendeeCandidate{Email: "guest" + strconv.Itoa(index) + "@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}
	}
	if _, err := calendar.NewCalendarEventUpdateCommand(calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: validCandidates}}); err != nil {
		t.Fatalf("200 attendee command error = %v", err)
	}
	tooMany := append(append([]calendar.CalendarEventCreateAttendeeCandidate(nil), validCandidates...), calendar.CalendarEventCreateAttendeeCandidate{Email: "overflow@example.test", Role: calendar.CalendarEventAttendeeRoleRequired})
	if _, err := calendar.NewCalendarEventUpdateCommand(calendar.CalendarEventUpdateInput{EventID: updateEventID, ExpectedVersion: 3, Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: tooMany}}); !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("201 attendee command error = %v, want %v", err, calendar.ErrEventInvalid)
	}
}

func TestEventServiceUpdateCalendarEvent_ReinviteReplacementAndWriteFailureRollback(t *testing.T) {
	t.Run("inactive reader omitted becomes fresh same email invite", func(t *testing.T) {
		backend := newUpdateBackendFake()
		backend.authorization.Attendees = []calendar.CalendarEventUpdateAttendeeSnapshot{
			updateOrganizerAttendeeSnapshot(),
		}
		backend.authorization.ResolvedAttendees = []calendar.CalendarEventCreateResolvedAttendee{{Email: "guest@example.test", DisplayName: stringPointer("Guest"), Role: calendar.CalendarEventAttendeeRoleRequired}}
		freshID := uuid.MustParse("62626262-6262-4262-8262-626262626262")
		service := calendar.NewEventService(&updateRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, time.January, 15, 1, 2, 3, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{freshID, updateApplicationID}}, createSerializerUnused{})
		_, err := service.UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
			EventID: updateEventID, ExpectedVersion: 3,
			Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}}},
		}))
		if err != nil || len(backend.persisted) != 1 {
			t.Fatalf("reinvite error/writes = %v/%d", err, len(backend.persisted))
		}
		mutations := backend.persisted[0].Attendees()
		if len(mutations) != 1 || mutations[0].Kind != calendar.CalendarEventUpdateAttendeeAdd || mutations[0].AttendeeID != freshID || mutations[0].AttendeeID == updateAttendeeID {
			t.Fatalf("fresh reinvite mutation = %#v", mutations)
		}
	})

	for _, testCase := range []struct {
		name string
		candidate calendar.CalendarEventCreateAttendeeCandidate
		resolved calendar.CalendarEventCreateResolvedAttendee
	}{
		{
			name: "role replacement",
			candidate: calendar.CalendarEventCreateAttendeeCandidate{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleOptional},
			resolved: calendar.CalendarEventCreateResolvedAttendee{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleOptional},
		},
		{
			name: "resolved identity replacement",
			candidate: calendar.CalendarEventCreateAttendeeCandidate{UserID: uuidPointer(uuid.MustParse("63636363-6363-4363-8363-636363636363")), Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired},
			resolved: calendar.CalendarEventCreateResolvedAttendee{UserID: uuidPointer(uuid.MustParse("63636363-6363-4363-8363-636363636363")), Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newUpdateBackendFake()
			backend.authorization.ResolvedAttendees = []calendar.CalendarEventCreateResolvedAttendee{testCase.resolved}
			_, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
				EventID: updateEventID, ExpectedVersion: 3, Attendees: calendar.CalendarEventUpdateAttendees{Present: true, Value: []calendar.CalendarEventCreateAttendeeCandidate{testCase.candidate}},
			}))
			if !errors.Is(err, calendar.ErrEventConflict) || len(backend.persisted) != 0 {
				t.Fatalf("replacement error/writes = %v/%d", err, len(backend.persisted))
			}
		})
	}

	backend := newUpdateBackendFake()
	backend.persistError = errors.New("SENSITIVE-UPDATE-WRITE-ERROR")
	_, err := newUpdateService(backend).UpdateCalendarEvent(context.Background(), updateActorID, mustUpdateCommand(t, calendar.CalendarEventUpdateInput{
		EventID: updateEventID, ExpectedVersion: 3, Location: calendar.CalendarEventUpdateString{Present: true, Value: "Write failure"},
	}))
	if !errors.Is(err, calendar.ErrEventInvalid) || strings.Contains(err.Error(), "SENSITIVE-UPDATE-WRITE-ERROR") || len(backend.persisted) != 0 {
		t.Fatalf("write failure error/writes = %v/%d", err, len(backend.persisted))
	}
}

func assertCalendarEventV2Envelope(t *testing.T, envelope events.EnvelopeV1, wantSubject string, wantSourceID, wantEventID, wantOrganizerID uuid.UUID, wantSequence uint, wantAttendeeIDs []uuid.UUID) {
	t.Helper()
	if envelope.Subject() != wantSubject || envelope.EventID() != wantSourceID {
		t.Fatalf("V2 envelope subject/source = %q/%s, want %q/%s", envelope.Subject(), envelope.EventID(), wantSubject, wantSourceID)
	}
	payload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil {
		t.Fatalf("DecodeCalendarInviteDeliveryEnvelopeV2() error = %v", err)
	}
	if payload.EventID != wantEventID || payload.EventSequence != wantSequence || payload.OrganizerID != wantOrganizerID || payload.RecurrenceID != nil || len(payload.Deliveries) != len(wantAttendeeIDs) {
		t.Fatalf("V2 payload = %#v", payload)
	}
	namespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.calendar.invite-delivery.id.v1"))
	wantDeliveries := make(map[uuid.UUID]uuid.UUID, len(wantAttendeeIDs))
	for _, attendeeID := range wantAttendeeIDs {
		wantDeliveries[attendeeID] = uuid.NewSHA1(namespace, []byte(wantSourceID.String()+"/"+attendeeID.String()))
	}
	for _, delivery := range payload.Deliveries {
		if inviteID, found := wantDeliveries[delivery.AttendeeID]; !found || delivery.InviteID != inviteID {
			t.Fatalf("V2 deterministic delivery = %#v", delivery)
		}
	}
}

func mustUpdateCommand(t *testing.T, input calendar.CalendarEventUpdateInput) calendar.CalendarEventUpdateCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventUpdateCommand(input)
	if err != nil {
		t.Fatalf("NewCalendarEventUpdateCommand() error = %v", err)
	}
	return command
}

func mustUpdatePersistenceCommand(t *testing.T) calendar.CalendarEventUpdatePersistenceCommand {
	t.Helper()
	organizer := updateActorID
	event, err := calendar.NewCalendarEventCommand(calendar.CalendarEventCommandInput{
		EventID: updateEventID, OrganizationID: updateOrganizationID, OwnerID: updateActorID, UID: "hub-" + updateEventID.String(),
		Title: "Persist", Temporal: calendar.CalendarEvent{StartsAt: updateStart(), EndsAt: updateStart().Add(time.Hour), Timezone: "UTC"},
		Version: 4, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed, OrganizerID: &organizer, OrganizerEmail: "owner@example.test",
	})
	if err != nil { t.Fatalf("NewCalendarEventCommand() error = %v", err) }
	command, err := calendar.NewCalendarEventUpdatePersistenceCommand(event, 3, nil, nil)
	if err != nil { t.Fatalf("NewCalendarEventUpdatePersistenceCommand() error = %v", err) }
	return command
}

func updateStart() time.Time { return time.Date(2031, time.January, 20, 9, 0, 0, 0, time.UTC) }

func timePointer(value time.Time) *time.Time { return &value }
func uuidPointer(value uuid.UUID) *uuid.UUID { return &value }

func sameUpdateChangedFields(left, right []calendar.CalendarEventChangedField) bool {
	if len(left) != len(right) { return false }
	for index := range left { if left[index] != right[index] { return false } }
	return true
}

func newUpdateService(backend *updateBackendFake) *calendar.EventService {
	return calendar.NewEventService(
		&updateRepositoryFake{backend: backend},
		&createClockFake{now: time.Date(2031, time.January, 11, 1, 2, 3, 0, time.UTC)},
		&createIDGeneratorFake{ids: []uuid.UUID{updateApplicationID}},
		createSerializerUnused{},
	)
}

type updateRepositoryFake struct {
	backend *updateBackendFake
	calls   int
}

func (repository *updateRepositoryFake) WithEventTransaction(_ context.Context, callback func(*calendar.EventTransaction) error) error {
	repository.calls++
	snapshot := append([]calendar.CalendarEventUpdatePersistenceCommand(nil), repository.backend.persisted...)
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil {
		repository.backend.persisted = snapshot
	}
	return err
}

type updateBackendFake struct {
	organizationID uuid.UUID
	authorization calendar.CalendarEventUpdateAuthorization
	authorizationError error
	persistError error
	resultSequence uint
	resultMutator func(calendar.CalendarEventUpdatePersistenceCommand, calendar.CalendarEventUpdatePersistenceResult) calendar.CalendarEventUpdatePersistenceResult
	lastRequest calendar.CalendarEventUpdateAuthorizationRequest
	persisted []calendar.CalendarEventUpdatePersistenceCommand
	organizationCalls int
	authorizationCalls int
	persistCalls int
	reserveCalls int
	completeCalls int
}

func updateOrganizerAttendeeSnapshot() calendar.CalendarEventUpdateAttendeeSnapshot {
	return calendar.CalendarEventUpdateAttendeeSnapshot{
		AttendeeID:  updateOrganizerAttendeeID,
		UserID:      uuidPointer(updateActorID),
		Email:       "owner@example.test",
		DisplayName: "Owner",
		Role:        calendar.CalendarEventAttendeeRoleOrganizer,
		Status:      calendar.RSVPAccepted,
		Version:     1,
	}
}

func newUpdateBackendFake() *updateBackendFake {
	organizer := updateActorID
	return &updateBackendFake{
		organizationID: updateOrganizationID,
		resultSequence: 8,
		authorization: calendar.CalendarEventUpdateAuthorization{
			Current: calendar.CalendarEventCommandInput{
				EventID: updateEventID, OrganizationID: updateOrganizationID, OwnerID: updateActorID, UID: "hub-" + updateEventID.String(),
				Title: "Original title", Description: "Original description", Location: "Room 4",
				Temporal: calendar.CalendarEvent{StartsAt: updateStart(), EndsAt: updateStart().Add(time.Hour), Timezone: "UTC"},
				Sequence: 7, Version: 3, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed,
				OrganizerID: &organizer, OrganizerEmail: "owner@example.test",
			},
			Attendees: []calendar.CalendarEventUpdateAttendeeSnapshot{
				{
					AttendeeID: updateAttendeeID, Email: "guest@example.test", DisplayName: "Guest", Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPNeedsAction, Version: 1,
				},
				updateOrganizerAttendeeSnapshot(),
			},
		},
	}
}

func (backend *updateBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	backend.organizationCalls++
	if actorID != updateActorID { return uuid.Nil, calendar.ErrEventForbidden }
	return backend.organizationID, nil
}

func (backend *updateBackendFake) ResolveCalendarEventUpdate(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventUpdateAuthorizationRequest) (calendar.CalendarEventUpdateAuthorization, error) {
	backend.authorizationCalls++
	backend.lastRequest = request
	if backend.authorizationError != nil { return calendar.CalendarEventUpdateAuthorization{}, backend.authorizationError }
	return backend.authorization, nil
}

func (backend *updateBackendFake) PersistCalendarEventUpdate(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventUpdatePersistenceCommand) (calendar.CalendarEventUpdatePersistenceResult, error) {
	backend.persistCalls++
	if backend.persistError != nil { return calendar.CalendarEventUpdatePersistenceResult{}, backend.persistError }
	result := mustUpdatePersistenceResultForCommand(command, backend.resultSequence)
	if backend.resultMutator != nil { result = backend.resultMutator(command, result) }
	backend.persisted = append(backend.persisted, command)
	return result, nil
}

func (backend *updateBackendFake) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *updateBackendFake) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) { backend.reserveCalls++; return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid }
func (backend *updateBackendFake) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error { backend.completeCalls++; return calendar.ErrEventInvalid }
func (backend *updateBackendFake) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBackendFake) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBackendFake) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBackendFake) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { return calendar.ErrEventInvalid }

type updateBaseOnlyBackend struct { organizationID uuid.UUID; organizationCalls int; otherCalls int }
func (backend *updateBaseOnlyBackend) ResolveActorOrganization(context.Context, uuid.UUID) (uuid.UUID, error) { backend.organizationCalls++; return backend.organizationID, nil }
func (backend *updateBaseOnlyBackend) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) { backend.otherCalls++; return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) { backend.otherCalls++; return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error { backend.otherCalls++; return calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) { backend.otherCalls++; return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid }
func (backend *updateBaseOnlyBackend) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error { backend.otherCalls++; return calendar.ErrEventInvalid }

func updatePersistenceEcho(command calendar.CalendarEventUpdatePersistenceCommand) []calendar.CalendarEventUpdateAttendeePersistenceResult {
	mutations := command.Attendees()
	results := make([]calendar.CalendarEventUpdateAttendeePersistenceResult, len(mutations))
	for index, mutation := range mutations {
		results[index] = calendar.CalendarEventUpdateAttendeePersistenceResult{AttendeeID: mutation.AttendeeID, Version: mutation.Version, Removed: mutation.Kind == calendar.CalendarEventUpdateAttendeeRemove}
	}
	return results
}

func mustUpdatePersistenceResultForCommand(command calendar.CalendarEventUpdatePersistenceCommand, sequence uint) calendar.CalendarEventUpdatePersistenceResult {
	result, err := calendar.NewCalendarEventUpdatePersistenceResult(command.Event().EventID(), command.Event().Version(), sequence, updatePersistenceEcho(command))
	if err != nil { panic(err) }
	return result
}

func mustUpdatePersistenceResult(t *testing.T, eventID uuid.UUID, version, sequence uint, attendees []calendar.CalendarEventUpdateAttendeePersistenceResult) calendar.CalendarEventUpdatePersistenceResult {
	t.Helper()
	result, err := calendar.NewCalendarEventUpdatePersistenceResult(eventID, version, sequence, attendees)
	if err != nil { t.Fatalf("NewCalendarEventUpdatePersistenceResult() error = %v", err) }
	return result
}
