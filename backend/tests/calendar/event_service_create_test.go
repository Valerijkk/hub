package calendar_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	createActorID        = uuid.MustParse("10101010-1010-4010-8010-101010101010")
	createOrganizationID = uuid.MustParse("20202020-2020-4020-8020-202020202020")
	createReceiptID      = uuid.MustParse("30303030-3030-4030-8030-303030303030")
	createApplicationID  = uuid.MustParse("40404040-4040-4040-8040-404040404040")
)

func TestEventServiceCreateCalendarEvent_PersistsCompleteEnvelopeAndReplays(t *testing.T) {
	backend := newCreateBackendFake()
	repository := &createRepositoryFake{backend: backend}
	clock := &createClockFake{now: time.Date(2031, time.February, 3, 4, 5, 6, 0, time.UTC)}
	ids := &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}
	service := calendar.NewEventService(repository, clock, ids, createSerializerUnused{})

	exdates := []string{"2031-02-10T04:05:06Z"}
	guestName := "Guest"
	attendees := []calendar.CalendarEventCreateAttendeeCandidate{{
		Email:       "guest@example.test",
		DisplayName: &guestName,
		Role:        calendar.CalendarEventAttendeeRoleRequired,
	}}
	command, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{
		IdempotencyKey:  "create-key-0000001",
		Title:           "Design review",
		Description:     "Private details stay outside the event envelope.",
		Location:        "Room 4",
		Temporal:        calendar.CalendarEvent{StartsAt: time.Date(2031, time.February, 10, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, time.February, 10, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Exdates:         exdates,
		ReminderMinutes: intPointer(15),
		Attendees:       attendees,
	})
	if err != nil {
		t.Fatalf("NewCalendarEventCreateCommand() error = %v", err)
	}
	exdates[0] = "mutated"
	attendees[0].Email = "mutated@example.test"

	first, err := service.CreateCalendarEvent(context.Background(), createActorID, command)
	if err != nil {
		t.Fatalf("CreateCalendarEvent() first error = %v", err)
	}
	if first.Replayed() {
		t.Fatal("first CreateCalendarEvent() result is replayed")
	}
	if got := first.Response().Status(); got != 201 {
		t.Fatalf("first response status = %d, want 201", got)
	}

	var response struct {
		ID             uuid.UUID `json:"id"`
		UID            string    `json:"uid"`
		Sequence       uint      `json:"sequence"`
		Version        uint      `json:"version"`
		OrganizerEmail string    `json:"organizer_email"`
		Attendees      []struct {
			Email      string              `json:"email"`
			UserID     *uuid.UUID          `json:"user_id"`
			RSVPStatus calendar.RSVPStatus `json:"rsvp_status"`
		} `json:"attendees"`
	}
	if err := json.Unmarshal(first.Response().JSON(), &response); err != nil {
		t.Fatalf("json.Unmarshal(create response) error = %v", err)
	}
	if response.ID == uuid.Nil || response.UID != "hub-"+response.ID.String() {
		t.Fatalf("created id/uid = %s/%q, want deterministic hub UUID UID", response.ID, response.UID)
	}
	if response.Sequence != 0 || response.Version != 1 || response.OrganizerEmail != "owner@example.test" {
		t.Fatalf("created response metadata = %#v", response)
	}
	if len(response.Attendees) != 1 || response.Attendees[0].Email != "guest@example.test" || response.Attendees[0].UserID != nil || response.Attendees[0].RSVPStatus != calendar.RSVPNeedsAction {
		t.Fatalf("external attendee response = %#v", response.Attendees)
	}
	if strings.Contains(string(first.Response().JSON()), "create-key-0000001") {
		t.Fatal("create response leaked raw Idempotency-Key")
	}
	if len(backend.events) != 1 || len(backend.attendees) != 2 || len(backend.outboxes) != 1 || len(backend.receipts) != 1 {
		t.Fatalf("first transaction mutations = events:%d attendees:%d outboxes:%d receipts:%d, want 1:2:1:1", len(backend.events), len(backend.attendees), len(backend.outboxes), len(backend.receipts))
	}

	envelope := backend.outboxes[0].Envelope()
	if envelope.Subject() != calendar.CalendarInviteCreatedSubjectV2 || envelope.EventID() != createApplicationID {
		t.Fatalf("outbox envelope subject/id = %q/%s", envelope.Subject(), envelope.EventID())
	}
	encodedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}
	var completeEnvelope map[string]json.RawMessage
	if err := json.Unmarshal(encodedEnvelope, &completeEnvelope); err != nil {
		t.Fatalf("json.Unmarshal(envelope) error = %v", err)
	}
	for _, field := range []string{"subject", "schema_version", "event_id", "occurred_at", "organization_id", "actor_id", "correlation_id", "payload"} {
		if _, found := completeEnvelope[field]; !found {
			t.Fatalf("complete envelope does not contain %q", field)
		}
	}
	if strings.Contains(string(encodedEnvelope), "Private details") || strings.Contains(string(encodedEnvelope), "create-key-0000001") || strings.Contains(string(encodedEnvelope), "calendar.invite.created.v1") {
		t.Fatalf("safe create envelope leaked content or raw key: %s", encodedEnvelope)
	}
	var exactEnvelope struct {
		Subject        string    `json:"subject"`
		SchemaVersion  int       `json:"schema_version"`
		EventID        uuid.UUID `json:"event_id"`
		OccurredAt     time.Time `json:"occurred_at"`
		OrganizationID uuid.UUID `json:"organization_id"`
		ActorID        *uuid.UUID `json:"actor_id"`
		CorrelationID  string    `json:"correlation_id"`
	}
	if err := json.Unmarshal(encodedEnvelope, &exactEnvelope); err != nil {
		t.Fatalf("json.Unmarshal(exact envelope) error = %v", err)
	}
	if exactEnvelope.Subject != calendar.CalendarInviteCreatedSubjectV2 || exactEnvelope.SchemaVersion != 1 || exactEnvelope.EventID != createApplicationID ||
		exactEnvelope.OccurredAt != clock.now || exactEnvelope.OrganizationID != createOrganizationID || exactEnvelope.ActorID == nil ||
		*exactEnvelope.ActorID != createActorID || exactEnvelope.CorrelationID != response.ID.String() {
		t.Fatalf("exact EnvelopeV1 values = %#v", exactEnvelope)
	}
	payload, err := calendar.DecodeCalendarInviteDeliveryEnvelopeV2(envelope)
	if err != nil {
		t.Fatalf("DecodeCalendarInviteDeliveryEnvelopeV2() error = %v", err)
	}
	eventSnapshot := backend.events[0].Snapshot()
	if eventSnapshot.OrganizerID == nil || payload.EventID != response.ID || payload.EventSequence != 0 || payload.OrganizerID != *eventSnapshot.OrganizerID || payload.RecurrenceID != nil || len(payload.Deliveries) != 1 {
		t.Fatalf("strict v2 create payload = %#v, event = %#v", payload, eventSnapshot)
	}
	attendeeID := backend.attendees[1].AttendeeID()
	inviteNamespace := uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.calendar.invite-delivery.id.v1"))
	wantInviteID := uuid.NewSHA1(inviteNamespace, []byte(createApplicationID.String()+"/"+attendeeID.String()))
	if payload.Deliveries[0].AttendeeID != attendeeID || payload.Deliveries[0].InviteID != wantInviteID {
		t.Fatalf("deterministic v2 delivery = %#v, want attendee %s invite %s", payload.Deliveries[0], attendeeID, wantInviteID)
	}

	replayed, err := service.CreateCalendarEvent(context.Background(), createActorID, command)
	if err != nil {
		t.Fatalf("CreateCalendarEvent() replay error = %v", err)
	}
	if !replayed.Replayed() || string(replayed.Response().JSON()) != string(first.Response().JSON()) {
		t.Fatalf("replay result = replayed:%t body:%s, want byte-equivalent original", replayed.Replayed(), replayed.Response().JSON())
	}
	if len(backend.events) != 1 || len(backend.attendees) != 2 || len(backend.outboxes) != 1 || len(backend.receipts) != 1 {
		t.Fatalf("replay added mutation = events:%d attendees:%d outboxes:%d receipts:%d", len(backend.events), len(backend.attendees), len(backend.outboxes), len(backend.receipts))
	}
	if clock.calls != 1 || ids.calls != 1 {
		t.Fatalf("replay invoked injected clock/id ports = %d/%d, want 1/1 total", clock.calls, ids.calls)
	}
	if backend.authorizationCalls != 1 {
		t.Fatalf("replay invoked authorization resolver %d times, want 1", backend.authorizationCalls)
	}
}

func TestEventServiceCreateCalendarEvent_ChangedInputConflictsWithoutSecondMutation(t *testing.T) {
	backend := newCreateBackendFake()
	service := calendar.NewEventService(
		&createRepositoryFake{backend: backend},
		&createClockFake{now: time.Date(2031, time.February, 3, 4, 5, 6, 0, time.UTC)},
		&createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}},
		createSerializerUnused{},
	)
	command := mustCreateCommand(t, "Original title")
	if _, err := service.CreateCalendarEvent(context.Background(), createActorID, command); err != nil {
		t.Fatalf("first CreateCalendarEvent() error = %v", err)
	}
	changed := mustCreateCommand(t, "Changed title")
	_, err := service.CreateCalendarEvent(context.Background(), createActorID, changed)
	if !errors.Is(err, calendar.ErrEventConflict) {
		t.Fatalf("changed-input CreateCalendarEvent() error = %v, want %v", err, calendar.ErrEventConflict)
	}
	if len(backend.events) != 1 || len(backend.attendees) != 2 || len(backend.outboxes) != 1 || len(backend.receipts) != 1 {
		t.Fatalf("changed input mutated state = events:%d attendees:%d outboxes:%d receipts:%d", len(backend.events), len(backend.attendees), len(backend.outboxes), len(backend.receipts))
	}
}

func TestEventServiceCreateCalendarEvent_RollsBackEveryWriteAndRedactsBackendError(t *testing.T) {
	backend := newCreateBackendFake()
	backend.appendError = errors.New("SENSITIVE-CREATE-BACKEND-ERROR")
	service := calendar.NewEventService(
		&createRepositoryFake{backend: backend},
		&createClockFake{now: time.Date(2031, time.February, 3, 4, 5, 6, 0, time.UTC)},
		&createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}},
		createSerializerUnused{},
	)

	_, err := service.CreateCalendarEvent(context.Background(), createActorID, mustCreateCommand(t, "Rollback"))
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("rollback CreateCalendarEvent() error = %v, want %v", err, calendar.ErrEventInvalid)
	}
	if strings.Contains(err.Error(), "SENSITIVE-CREATE-BACKEND-ERROR") {
		t.Fatalf("safe error leaked backend detail: %q", err)
	}
	if len(backend.events) != 0 || len(backend.attendees) != 0 || len(backend.outboxes) != 0 || len(backend.receipts) != 0 {
		t.Fatalf("rollback retained partial state = events:%d attendees:%d outboxes:%d receipts:%d", len(backend.events), len(backend.attendees), len(backend.outboxes), len(backend.receipts))
	}
}

func TestEventServiceCreateCalendarEvent_RejectsInvalidCommandBeforeTransaction(t *testing.T) {
	repository := &createRepositoryFake{backend: newCreateBackendFake()}
	service := calendar.NewEventService(repository, &createClockFake{}, &createIDGeneratorFake{}, createSerializerUnused{})

	_, err := service.CreateCalendarEvent(context.Background(), createActorID, calendar.CalendarEventCreateCommand{})
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("invalid command CreateCalendarEvent() error = %v, want %v", err, calendar.ErrEventInvalid)
	}
	if repository.calls != 0 {
		t.Fatalf("invalid command transaction calls = %d, want 0", repository.calls)
	}
}

func TestEventServiceCreateCalendarEvent_RejectsAuthorizationAndScheduledCallBeforeWrites(t *testing.T) {
	wrongActor := uuid.MustParse("90909090-9090-4090-8090-909090909090")
	callID := uuid.MustParse("80808080-8080-4080-8080-808080808080")
	replacementCallID := uuid.MustParse("70707070-7070-4070-8070-707070707070")
	for _, testCase := range []struct {
		name          string
		command       func(*testing.T) calendar.CalendarEventCreateCommand
		authorization calendar.CalendarEventCreateAuthorization
		resolverError error
		want          error
	}{
		{
			name: "delegated organizer",
			command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommand(t, "Delegation") },
			authorization: calendar.CalendarEventCreateAuthorization{Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: wrongActor, Email: "other@example.test"}},
			want: calendar.ErrEventConflict,
		},
		{
			name: "hidden scheduled call",
			command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Hidden call", &callID) },
			resolverError: calendar.ErrEventNotFound,
			want: calendar.ErrEventNotFound,
		},
		{name: "inactive owned scheduled call", command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Inactive call", &callID) }, resolverError: calendar.ErrEventConflict, want: calendar.ErrEventConflict},
		{name: "linked scheduled call", command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Linked call", &callID) }, resolverError: calendar.ErrEventConflict, want: calendar.ErrEventConflict},
		{name: "context mismatched scheduled call", command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Context call", &callID) }, resolverError: calendar.ErrEventConflict, want: calendar.ErrEventConflict},
		{name: "mapping mismatched scheduled call", command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Mapping call", &callID) }, resolverError: calendar.ErrEventConflict, want: calendar.ErrEventConflict},
		{name: "external guest scheduled call", command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "External guest call", &callID) }, resolverError: calendar.ErrEventConflict, want: calendar.ErrEventConflict},
		{
			name: "scheduled call dropped by resolver",
			command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Substituted call", &callID) },
			authorization: calendar.CalendarEventCreateAuthorization{
				Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
				Attendees: []calendar.CalendarEventCreateResolvedAttendee{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
			},
			want: calendar.ErrEventConflict,
		},
		{
			name: "scheduled call substituted by resolver",
			command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommandWithCall(t, "Substituted call", &callID) },
			authorization: calendar.CalendarEventCreateAuthorization{
				Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
				Attendees: []calendar.CalendarEventCreateResolvedAttendee{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
				ScheduledCallID: &replacementCallID,
			},
			want: calendar.ErrEventConflict,
		},
		{
			name: "scheduled call inserted by resolver",
			command: func(t *testing.T) calendar.CalendarEventCreateCommand { return mustCreateCommand(t, "Inserted call") },
			authorization: calendar.CalendarEventCreateAuthorization{
				Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
				Attendees: []calendar.CalendarEventCreateResolvedAttendee{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
				ScheduledCallID: &callID,
			},
			want: calendar.ErrEventConflict,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newCreateBackendFake()
			backend.authorization = testCase.authorization
			backend.authorizationError = testCase.resolverError
			service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
			_, err := service.CreateCalendarEvent(context.Background(), createActorID, testCase.command(t))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("CreateCalendarEvent() error = %v, want %v", err, testCase.want)
			}
			if len(backend.events) != 0 || len(backend.attendees) != 0 || len(backend.outboxes) != 0 || len(backend.receipts) != 0 {
				t.Fatalf("authorization/call failure retained writes = events:%d attendees:%d outboxes:%d receipts:%d", len(backend.events), len(backend.attendees), len(backend.outboxes), len(backend.receipts))
			}
		})
	}
}

func TestEventServiceCreateCalendarEvent_PersistsAndReturnsExactResolvedScheduledCall(t *testing.T) {
	callID := uuid.MustParse("61616161-6161-4161-8161-616161616161")
	internalAttendeeID := uuid.MustParse("62626262-6262-4262-8262-626262626262")
	backend := newCreateBackendFake()
	backend.authorization = calendar.CalendarEventCreateAuthorization{
		Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
		Attendees: []calendar.CalendarEventCreateResolvedAttendee{{UserID: &internalAttendeeID, Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
		ScheduledCallID: &callID,
	}
	service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
	result, err := service.CreateCalendarEvent(context.Background(), createActorID, mustCreateCommandWithCall(t, "Exact call", &callID))
	if err != nil {
		t.Fatalf("CreateCalendarEvent() error = %v", err)
	}
	var response struct { ScheduledCallID *uuid.UUID `json:"scheduled_call_id"` }
	if err := json.Unmarshal(result.Response().JSON(), &response); err != nil {
		t.Fatalf("json.Unmarshal(create response) error = %v", err)
	}
	if len(backend.events) != 1 {
		t.Fatalf("persisted event count = %d, want 1", len(backend.events))
	}
	persisted := backend.events[0].Snapshot()
	if response.ScheduledCallID == nil || *response.ScheduledCallID != callID || persisted.ScheduledCallID == nil || *persisted.ScheduledCallID != callID {
		t.Fatalf("resolved scheduled-call persistence/response = response:%#v persisted:%#v", response, persisted)
	}
}

func TestEventServiceCreateCalendarEvent_RejectsScheduledCallWithExternalGuestBeforeDomainWrites(t *testing.T) {
	callID := uuid.MustParse("63636363-6363-4363-8363-636363636363")
	backend := newCreateBackendFake()
	backend.authorization = calendar.CalendarEventCreateAuthorization{
		Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
		Attendees: []calendar.CalendarEventCreateResolvedAttendee{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
		ScheduledCallID: &callID,
	}
	service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
	_, err := service.CreateCalendarEvent(context.Background(), createActorID, mustCreateCommandWithCall(t, "External guest call", &callID))
	if !errors.Is(err, calendar.ErrEventConflict) {
		t.Fatalf("CreateCalendarEvent() error = %v, want %v", err, calendar.ErrEventConflict)
	}
	if len(backend.events) != 0 || len(backend.attendees) != 0 || len(backend.outboxes) != 0 {
		t.Fatalf("external guest scheduled-call failure retained domain writes = events:%d attendees:%d outboxes:%d", len(backend.events), len(backend.attendees), len(backend.outboxes))
	}
	for _, receipt := range backend.receipts {
		if receipt.completion != nil {
			t.Fatal("external guest scheduled-call failure completed a mutation receipt")
		}
	}
}

func TestEventServiceCreateCalendarEvent_UsesAuthoritativeResolvedAttendeeIdentity(t *testing.T) {
	resolvedUserID := uuid.MustParse("51515151-5151-4151-8151-515151515151")
	candidateName := "Untrusted candidate display"
	resolvedName := "Directory canonical display"
	backend := newCreateBackendFake()
	backend.authorization = calendar.CalendarEventCreateAuthorization{
		Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: createActorID, Email: "owner@example.test"},
		Attendees: []calendar.CalendarEventCreateResolvedAttendee{{
			UserID: &resolvedUserID, Email: "IDENTITY@example.test", DisplayName: &resolvedName, Role: calendar.CalendarEventAttendeeRoleOptional,
		}},
	}
	service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
	command, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{
		IdempotencyKey: "create-key-0000001", Title: "Resolved attendee",
		Temporal: calendar.CalendarEvent{StartsAt: time.Date(2031, 2, 10, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, 2, 10, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Attendees: []calendar.CalendarEventCreateAttendeeCandidate{{Email: "identity@example.test", DisplayName: &candidateName, Role: calendar.CalendarEventAttendeeRoleOptional}},
	})
	if err != nil { t.Fatalf("NewCalendarEventCreateCommand() error = %v", err) }
	result, err := service.CreateCalendarEvent(context.Background(), createActorID, command)
	if err != nil { t.Fatalf("CreateCalendarEvent() error = %v", err) }
	var response struct {
		Attendees []struct { UserID *uuid.UUID `json:"user_id"`; Email string `json:"email"`; DisplayName *string `json:"display_name"` } `json:"attendees"`
	}
	if err := json.Unmarshal(result.Response().JSON(), &response); err != nil { t.Fatalf("json.Unmarshal(create response) error = %v", err) }
	if len(response.Attendees) != 1 || response.Attendees[0].UserID == nil || *response.Attendees[0].UserID != resolvedUserID ||
		response.Attendees[0].Email != "identity@example.test" || response.Attendees[0].DisplayName == nil || *response.Attendees[0].DisplayName != resolvedName {
		t.Fatalf("response did not use resolved attendee = %#v", response.Attendees)
	}
	organizerCount := 0
	attendeeCount := 0
	unexpectedRoleCount := 0
	var persisted calendar.CalendarEventAttendeeCommandInput
	for _, command := range backend.attendees {
		snapshot := command.Snapshot()
		switch snapshot.Role {
		case calendar.CalendarEventAttendeeRoleOrganizer:
			organizerCount++
		case calendar.CalendarEventAttendeeRoleOptional:
			attendeeCount++
			persisted = snapshot
		default:
			unexpectedRoleCount++
		}
	}
	if organizerCount != 1 || attendeeCount != 1 || unexpectedRoleCount != 0 {
		t.Fatalf("persisted attendee roles = organizer:%d attendee:%d unexpected:%d, want 1:1:0", organizerCount, attendeeCount, unexpectedRoleCount)
	}
	if persisted.UserID == nil || *persisted.UserID != resolvedUserID || persisted.Email != "identity@example.test" || persisted.DisplayName != resolvedName {
		t.Fatalf("persisted attendee did not use resolver output = %#v", persisted)
	}
}

func TestNewCalendarEventCreateCommand_CanonicalizesRecurrenceAndExdatesAndRejectsBoundaries(t *testing.T) {
	timed := calendar.CalendarEvent{StartsAt: time.Date(2031, time.February, 10, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, time.February, 10, 10, 0, 0, 0, time.UTC), Timezone: "UTC"}
	allDay := calendar.CalendarEvent{AllDay: true, StartDate: "2031-02-10", EndDateExclusive: "2031-02-11", Timezone: "Europe/Moscow"}
	for _, testCase := range []struct { name string; temporal calendar.CalendarEvent; raw string; want string }{
		{name: "weekly reordered", temporal: timed, raw: "INTERVAL=2;FREQ=WEEKLY;BYDAY=WE,MO;COUNT=2", want: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;COUNT=2"},
		{name: "daily default interval count", temporal: timed, raw: "FREQ=DAILY;COUNT=2", want: "FREQ=DAILY;INTERVAL=1;COUNT=2"},
		{name: "weekly seven weekdays", temporal: timed, raw: "FREQ=WEEKLY;BYDAY=SU,SA,FR,TH,WE,TU,MO;COUNT=2", want: "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU;COUNT=2"},
		{name: "daily upper interval and count", temporal: timed, raw: "FREQ=DAILY;INTERVAL=365;COUNT=1000", want: "FREQ=DAILY;INTERVAL=365;COUNT=1000"},
		{name: "timed until", temporal: timed, raw: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;UNTIL=20311231T090000Z", want: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;UNTIL=20311231T090000Z"},
		{name: "all-day monthly until", temporal: allDay, raw: "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31;UNTIL=20311231", want: "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31;UNTIL=20311231"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := createCalendarCreateRRULE(t, testCase.temporal, testCase.raw); got != testCase.want {
				t.Fatalf("normalized RRULE = %q, want %q", got, testCase.want)
			}
		})
	}
	first, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Canonical recurrence", Temporal: timed, RRULE: "INTERVAL=2;FREQ=WEEKLY;BYDAY=WE,MO;COUNT=2", Exdates: []string{"2031-02-12T09:00:00Z", "2031-02-10T09:00:00Z"}})
	if err != nil { t.Fatalf("NewCalendarEventCreateCommand(first) error = %v", err) }
	second, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Canonical recurrence", Temporal: timed, RRULE: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE;COUNT=2", Exdates: []string{"2031-02-10T09:00:00Z", "2031-02-12T09:00:00Z"}})
	if err != nil { t.Fatalf("NewCalendarEventCreateCommand(second) error = %v", err) }
	backend := newCreateBackendFake()
	service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
	if _, err := service.CreateCalendarEvent(context.Background(), createActorID, first); err != nil { t.Fatalf("first canonical create error = %v", err) }
	if replay, err := service.CreateCalendarEvent(context.Background(), createActorID, second); err != nil || !replay.Replayed() { t.Fatalf("canonical equivalent replay = %#v, %v", replay, err) }
	if got := backend.events[0].Exdates(); len(got) != 2 || got[0] != "2031-02-10T09:00:00Z" || got[1] != "2031-02-12T09:00:00Z" { t.Fatalf("canonical EXDATE persistence = %#v", got) }

	for _, testCase := range []struct { name string; temporal calendar.CalendarEvent; raw string }{
		{name: "unknown frequency", temporal: timed, raw: "FREQ=YEARLY;COUNT=1"},
		{name: "duplicate key", temporal: timed, raw: "FREQ=DAILY;FREQ=WEEKLY;COUNT=1"},
		{name: "unknown selector", temporal: timed, raw: "FREQ=DAILY;BYHOUR=9;COUNT=1"},
		{name: "daily byday", temporal: timed, raw: "FREQ=DAILY;BYDAY=MO;COUNT=1"},
		{name: "weekly missing byday", temporal: timed, raw: "FREQ=WEEKLY;COUNT=1"},
		{name: "weekly duplicate byday", temporal: timed, raw: "FREQ=WEEKLY;BYDAY=MO,MO;COUNT=1"},
		{name: "weekly numbered byday", temporal: timed, raw: "FREQ=WEEKLY;BYDAY=1MO;COUNT=1"},
		{name: "monthly missing bymonthday", temporal: timed, raw: "FREQ=MONTHLY;COUNT=1"},
		{name: "monthly byday", temporal: timed, raw: "FREQ=MONTHLY;BYDAY=MO;BYMONTHDAY=1;COUNT=1"},
		{name: "monthly day lower bound", temporal: timed, raw: "FREQ=MONTHLY;BYMONTHDAY=0;COUNT=1"},
		{name: "monthly day upper bound", temporal: timed, raw: "FREQ=MONTHLY;BYMONTHDAY=32;COUNT=1"},
		{name: "interval lower bound", temporal: timed, raw: "FREQ=DAILY;INTERVAL=0;COUNT=1"},
		{name: "interval upper bound", temporal: timed, raw: "FREQ=DAILY;INTERVAL=366;COUNT=1"},
		{name: "count lower bound", temporal: timed, raw: "FREQ=DAILY;COUNT=0"},
		{name: "count upper bound", temporal: timed, raw: "FREQ=DAILY;COUNT=1001"},
		{name: "count until exclusive", temporal: timed, raw: "FREQ=DAILY;COUNT=2;UNTIL=20261231T090000Z"},
		{name: "timed until date", temporal: timed, raw: "FREQ=DAILY;UNTIL=20261231"},
		{name: "all-day until timestamp", temporal: allDay, raw: "FREQ=DAILY;UNTIL=20261231T090000Z"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Invalid RRULE", Temporal: testCase.temporal, RRULE: testCase.raw})
			if !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("NewCalendarEventCreateCommand() error = %v, want %v", err, calendar.ErrEventInvalid) }
		})
	}
	for _, name := range []string{strings.Repeat("x", 160), strings.Repeat("😀", 160)} {
		_, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Accepted name", Temporal: timed, Attendees: []calendar.CalendarEventCreateAttendeeCandidate{{Email: "guest@example.test", DisplayName: stringPointer(name), Role: calendar.CalendarEventAttendeeRoleRequired}}})
		if err != nil { t.Fatalf("accepted display name error = %v", err) }
	}
	for _, name := range []string{strings.Repeat("x", 161), strings.Repeat("😀", 161)} {
		_, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Rejected name", Temporal: timed, Attendees: []calendar.CalendarEventCreateAttendeeCandidate{{Email: "guest@example.test", DisplayName: stringPointer(name), Role: calendar.CalendarEventAttendeeRoleRequired}}})
		if !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("rejected display name error = %v, want %v", err, calendar.ErrEventInvalid) }
	}
	validLongIANA := timed; validLongIANA.Timezone = "America/Argentina/ComodRivadavia"
	if _, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Accepted IANA", Temporal: validLongIANA}); err != nil { t.Fatalf("accepted IANA timezone error = %v", err) }
	for _, timezone := range []string{"Mars/Olympus", strings.Repeat("U", 65)} {
		invalidTimezone := timed; invalidTimezone.Timezone = timezone
		if _, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Rejected timezone", Temporal: invalidTimezone}); !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("rejected timezone error = %v, want %v", err, calendar.ErrEventInvalid) }
	}
	for _, input := range []calendar.CalendarEventCreateInput{
		{IdempotencyKey: "create-key-0000001", Title: "Bad timed EXDATE", Temporal: timed, Exdates: []string{"2031-02-10T09:00:00+03:00"}},
		{IdempotencyKey: "create-key-0000001", Title: "Bad all-day EXDATE", Temporal: allDay, Exdates: []string{"2031-02-10T09:00:00Z"}},
		{IdempotencyKey: "create-key-0000001", Title: "Duplicate EXDATE", Temporal: allDay, Exdates: []string{"2031-02-10", "2031-02-10"}},
	} {
		if _, err := calendar.NewCalendarEventCreateCommand(input); !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("invalid EXDATE error = %v, want %v", err, calendar.ErrEventInvalid) }
	}
	tooMany := make([]string, 1001)
	for index := range tooMany { tooMany[index] = "2031-02-10" }
	if _, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Too many EXDATE", Temporal: allDay, Exdates: tooMany}); !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("1001 EXDATE error = %v, want %v", err, calendar.ErrEventInvalid) }
}

func TestNewCalendarEventCreateCommand_ProvesBoundedRRuleOccurrences(t *testing.T) {
	dailyStart := time.Date(2031, time.January, 1, 9, 0, 0, 0, time.UTC)
	daily := calendar.CalendarEvent{StartsAt: dailyStart, EndsAt: dailyStart.Add(time.Hour), Timezone: "UTC"}
	frozenDailyStart := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)
	frozenDaily := calendar.CalendarEvent{StartsAt: frozenDailyStart, EndsAt: frozenDailyStart.Add(time.Hour), Timezone: "UTC"}
	frozenAllDay := calendar.CalendarEvent{AllDay: true, StartDate: "2026-01-01", EndDateExclusive: "2026-01-02", Timezone: "UTC"}
	weeklyMondayStart := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	weeklyMonday := calendar.CalendarEvent{StartsAt: weeklyMondayStart, EndsAt: weeklyMondayStart.Add(time.Hour), Timezone: "UTC"}
	monthlyStart := time.Date(2024, time.January, 31, 9, 0, 0, 0, time.UTC)
	monthly := calendar.CalendarEvent{StartsAt: monthlyStart, EndsAt: monthlyStart.Add(time.Hour), Timezone: "UTC"}
	monthlySelectorStart := time.Date(2031, time.January, 1, 9, 0, 0, 0, time.UTC)
	monthlySelector := calendar.CalendarEvent{StartsAt: monthlySelectorStart, EndsAt: monthlySelectorStart.Add(time.Hour), Timezone: "UTC"}
	monthlySelectorAllDay := calendar.CalendarEvent{AllDay: true, StartDate: "2031-01-01", EndDateExclusive: "2031-01-02", Timezone: "UTC"}
	newYorkGapStart := time.Date(2026, time.March, 1, 7, 30, 0, 0, time.UTC)
	newYorkGap := calendar.CalendarEvent{StartsAt: newYorkGapStart, EndsAt: newYorkGapStart.Add(time.Hour), Timezone: "America/New_York"}
	newYorkFoldStart := time.Date(2024, time.February, 5, 6, 30, 0, 0, time.UTC)
	newYorkFold := calendar.CalendarEvent{StartsAt: newYorkFoldStart, EndsAt: newYorkFoldStart.Add(time.Hour), Timezone: "America/New_York"}
	for _, testCase := range []struct {
		name string
		temporal calendar.CalendarEvent
		rrule string
		valid bool
	}{
		{name: "count 1000", temporal: daily, rrule: "FREQ=DAILY;COUNT=1000", valid: true},
		{name: "count 1001", temporal: daily, rrule: "FREQ=DAILY;COUNT=1001"},
		{name: "neither count nor until", temporal: daily, rrule: "FREQ=DAILY"},
		{name: "count and until", temporal: daily, rrule: "FREQ=DAILY;COUNT=2;UNTIL=20310102T090000Z"},
		{name: "until before DTSTART", temporal: daily, rrule: "FREQ=DAILY;UNTIL=20301231T090000Z"},
		{name: "frozen daily literal exactly 1000", temporal: frozenDaily, rrule: "FREQ=DAILY;UNTIL=20280926T090000Z", valid: true},
		{name: "frozen daily literal 1001st", temporal: frozenDaily, rrule: "FREQ=DAILY;UNTIL=20280927T090000Z"},
		{name: "all-day daily literal exactly 1000", temporal: frozenAllDay, rrule: "FREQ=DAILY;UNTIL=20280926", valid: true},
		{name: "all-day daily literal 1001st", temporal: frozenAllDay, rrule: "FREQ=DAILY;UNTIL=20280927"},
		{name: "weekly interval two Monday", temporal: weeklyMonday, rrule: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO;UNTIL=20240212T090000Z", valid: true},
		{name: "monthly BYMONTHDAY 31 skips missing dates", temporal: monthly, rrule: "FREQ=MONTHLY;BYMONTHDAY=31;UNTIL=20240531T090000Z", valid: true},
		{name: "timed DTSTART-month selector makes exact 1000", temporal: monthlySelector, rrule: "FREQ=MONTHLY;BYMONTHDAY=2;UNTIL=21140302T090000Z", valid: true},
		{name: "timed DTSTART-month selector makes 1001", temporal: monthlySelector, rrule: "FREQ=MONTHLY;BYMONTHDAY=2;UNTIL=21140402T090000Z"},
		{name: "all-day DTSTART-month selector makes exact 1000", temporal: monthlySelectorAllDay, rrule: "FREQ=MONTHLY;BYMONTHDAY=2;UNTIL=21140302", valid: true},
		{name: "all-day DTSTART-month selector makes 1001", temporal: monthlySelectorAllDay, rrule: "FREQ=MONTHLY;BYMONTHDAY=2;UNTIL=21140402"},
		{name: "New York spring gap skips March 8 and March 15 is occurrence two", temporal: newYorkGap, rrule: "FREQ=WEEKLY;BYDAY=SU;COUNT=2", valid: true},
		{name: "New York fall fold earlier instant reaches the 1001st cap", temporal: newYorkFold, rrule: "FREQ=DAILY;UNTIL=20261101T053000Z"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "Occurrence cap", Temporal: testCase.temporal, RRULE: testCase.rrule})
			if testCase.valid && err != nil { t.Fatalf("NewCalendarEventCreateCommand() error = %v", err) }
			if !testCase.valid && !errors.Is(err, calendar.ErrEventInvalid) { t.Fatalf("NewCalendarEventCreateCommand() error = %v, want %v", err, calendar.ErrEventInvalid) }
		})
	}
}

func createCalendarCreateRRULE(t *testing.T, temporal calendar.CalendarEvent, raw string) string {
	t.Helper()
	backend := newCreateBackendFake()
	service := calendar.NewEventService(&createRepositoryFake{backend: backend}, &createClockFake{now: time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)}, &createIDGeneratorFake{ids: []uuid.UUID{createApplicationID}}, createSerializerUnused{})
	command, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{IdempotencyKey: "create-key-0000001", Title: "RRULE oracle", Temporal: temporal, RRULE: raw})
	if err != nil { t.Fatalf("NewCalendarEventCreateCommand() error = %v", err) }
	result, err := service.CreateCalendarEvent(context.Background(), createActorID, command)
	if err != nil { t.Fatalf("CreateCalendarEvent() error = %v", err) }
	var response struct { RRULE string `json:"rrule"` }
	if err := json.Unmarshal(result.Response().JSON(), &response); err != nil { t.Fatalf("json.Unmarshal(create response) error = %v", err) }
	return response.RRULE
}

func mustCreateCommand(t *testing.T, title string) calendar.CalendarEventCreateCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{
		IdempotencyKey: "create-key-0000001",
		Title:          title,
		Temporal:       calendar.CalendarEvent{StartsAt: time.Date(2031, time.February, 10, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, time.February, 10, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Attendees: []calendar.CalendarEventCreateAttendeeCandidate{{
			Email: "guest@example.test",
			Role:  calendar.CalendarEventAttendeeRoleRequired,
		}},
	})
	if err != nil {
		t.Fatalf("NewCalendarEventCreateCommand() error = %v", err)
	}
	return command
}

func mustCreateCommandWithCall(t *testing.T, title string, scheduledCallID *uuid.UUID) calendar.CalendarEventCreateCommand {
	t.Helper()
	command, err := calendar.NewCalendarEventCreateCommand(calendar.CalendarEventCreateInput{
		IdempotencyKey: "create-key-0000001", Title: title, ScheduledCallID: scheduledCallID,
		Temporal: calendar.CalendarEvent{StartsAt: time.Date(2031, time.February, 10, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2031, time.February, 10, 10, 0, 0, 0, time.UTC), Timezone: "UTC"},
		Attendees: []calendar.CalendarEventCreateAttendeeCandidate{{Email: "guest@example.test", Role: calendar.CalendarEventAttendeeRoleRequired}},
	})
	if err != nil { t.Fatalf("NewCalendarEventCreateCommand() error = %v", err) }
	return command
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string { return &value }

type createRepositoryFake struct {
	backend *createBackendFake
	calls   int
}

func (repository *createRepositoryFake) WithEventTransaction(ctx context.Context, callback func(*calendar.EventTransaction) error) error {
	repository.calls++
	snapshot := repository.backend.snapshot()
	err := calendar.RunEventTransaction(repository.backend, callback)
	if err != nil {
		repository.backend.restore(snapshot)
	}
	return err
}

type createReceiptKey struct {
	actorID        uuid.UUID
	organizationID uuid.UUID
	operation      calendar.MutationReceiptOperation
	keyHash        calendar.MutationReceiptHash
}

type createReceipt struct {
	requestHash calendar.MutationReceiptHash
	receiptID   uuid.UUID
	completion  *calendar.MutationReceiptResponse
}

type createBackendFake struct {
	organizationByActor map[uuid.UUID]uuid.UUID
	receipts            map[createReceiptKey]createReceipt
	events              []calendar.CalendarEventCommand
	attendees           []calendar.CalendarEventAttendeeCommand
	eventSnapshots      []string
	attendeeSnapshots   []string
	outboxes            []calendar.EventOutboxDescriptor
	appendError         error
	authorization       calendar.CalendarEventCreateAuthorization
	authorizationError  error
	authorizationCalls  int
	lastAuthorizationRequest calendar.CalendarEventCreateAuthorizationRequest
}

type createBackendSnapshot struct {
	receipts  map[createReceiptKey]createReceipt
	events    []calendar.CalendarEventCommand
	attendees []calendar.CalendarEventAttendeeCommand
	eventSnapshots []string
	attendeeSnapshots []string
	outboxes  []calendar.EventOutboxDescriptor
}

func newCreateBackendFake() *createBackendFake {
	return &createBackendFake{
		organizationByActor: map[uuid.UUID]uuid.UUID{createActorID: createOrganizationID},
		receipts:            make(map[createReceiptKey]createReceipt),
	}
}

func (backend *createBackendFake) ResolveActorOrganization(_ context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	organizationID, found := backend.organizationByActor[actorID]
	if !found {
		return uuid.Nil, calendar.ErrEventForbidden
	}
	return organizationID, nil
}

func (backend *createBackendFake) ResolveCalendarEventCreateAuthorization(_ context.Context, scope calendar.EventMutationScope, request calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) {
	backend.authorizationCalls++
	backend.lastAuthorizationRequest = request
	if backend.authorizationError != nil {
		return calendar.CalendarEventCreateAuthorization{}, backend.authorizationError
	}
	if backend.authorization.Organizer.UserID != uuid.Nil || backend.authorization.Organizer.Email != "" || backend.authorization.Attendees != nil || backend.authorization.ScheduledCallID != nil {
		return backend.authorization, nil
	}
	organizerName := "Resolved owner"
	authorization := calendar.CalendarEventCreateAuthorization{
		Organizer: calendar.CalendarEventCreateResolvedOrganizer{UserID: scope.ActorID(), Email: "owner@example.test", DisplayName: &organizerName},
		Attendees: make([]calendar.CalendarEventCreateResolvedAttendee, len(request.Attendees)),
		ScheduledCallID: request.ScheduledCallID,
	}
	for index, candidate := range request.Attendees {
		authorization.Attendees[index] = calendar.CalendarEventCreateResolvedAttendee{UserID: candidate.UserID, Email: candidate.Email, DisplayName: candidate.DisplayName, Role: candidate.Role}
	}
	return authorization, nil
}

func (backend *createBackendFake) ReserveMutationReceipt(_ context.Context, scope calendar.EventMutationScope, reservation calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	key := createReceiptKey{actorID: scope.ActorID(), organizationID: scope.OrganizationID(), operation: scope.Operation(), keyHash: reservation.KeyHash()}
	if existing, found := backend.receipts[key]; found {
		if existing.requestHash != reservation.RequestHash() || existing.completion == nil {
			return calendar.MutationReceiptReservationResult{}, calendar.ErrEventConflict
		}
		return calendar.NewReplayedMutationReceiptResult(existing.receiptID, *existing.completion)
	}
	backend.receipts[key] = createReceipt{requestHash: reservation.RequestHash(), receiptID: createReceiptID}
	return calendar.NewReservedMutationReceiptResult(createReceiptID)
}

func (backend *createBackendFake) CompleteMutationReceipt(_ context.Context, scope calendar.EventMutationScope, completion calendar.MutationReceiptCompletion) error {
	for key, receipt := range backend.receipts {
		if key.actorID != scope.ActorID() || key.organizationID != scope.OrganizationID() || key.operation != scope.Operation() || receipt.receiptID != completion.ReceiptID() {
			continue
		}
		response := completion.Response()
		receipt.completion = &response
		backend.receipts[key] = receipt
		return nil
	}
	return calendar.ErrEventInvalid
}

func (backend *createBackendFake) PersistCalendarEvent(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	backend.events = append(backend.events, command)
	backend.eventSnapshots = append(backend.eventSnapshots, fmt.Sprintf("%+v", command))
	return calendar.NewCalendarEventPersistenceResult(command.EventID(), command.Version())
}

func (backend *createBackendFake) PersistCalendarEventAttendee(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	backend.attendees = append(backend.attendees, command)
	backend.attendeeSnapshots = append(backend.attendeeSnapshots, fmt.Sprintf("%+v", command))
	return calendar.NewCalendarEventAttendeePersistenceResult(command.AttendeeID(), command.Version())
}

func (backend *createBackendFake) PersistCalendarEventException(_ context.Context, _ calendar.EventMutationScope, _ calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid
}

func (backend *createBackendFake) AppendOutbox(_ context.Context, _ calendar.EventMutationScope, descriptor calendar.EventOutboxDescriptor) error {
	if backend.appendError != nil {
		return backend.appendError
	}
	backend.outboxes = append(backend.outboxes, descriptor)
	return nil
}

func (backend *createBackendFake) snapshot() createBackendSnapshot {
	receipts := make(map[createReceiptKey]createReceipt, len(backend.receipts))
	for key, receipt := range backend.receipts {
		if receipt.completion != nil {
			response := *receipt.completion
			receipt.completion = &response
		}
		receipts[key] = receipt
	}
	return createBackendSnapshot{
		receipts:  receipts,
		events:    append([]calendar.CalendarEventCommand(nil), backend.events...),
		attendees: append([]calendar.CalendarEventAttendeeCommand(nil), backend.attendees...),
		eventSnapshots: append([]string(nil), backend.eventSnapshots...),
		attendeeSnapshots: append([]string(nil), backend.attendeeSnapshots...),
		outboxes:  append([]calendar.EventOutboxDescriptor(nil), backend.outboxes...),
	}
}

func (backend *createBackendFake) restore(snapshot createBackendSnapshot) {
	backend.receipts = snapshot.receipts
	backend.events = snapshot.events
	backend.attendees = snapshot.attendees
	backend.eventSnapshots = snapshot.eventSnapshots
	backend.attendeeSnapshots = snapshot.attendeeSnapshots
	backend.outboxes = snapshot.outboxes
}

type createClockFake struct {
	now   time.Time
	calls int
}

func (clock *createClockFake) Now() time.Time {
	clock.calls++
	return clock.now
}

type createIDGeneratorFake struct {
	ids   []uuid.UUID
	calls int
}

func (generator *createIDGeneratorFake) NewEventID() uuid.UUID {
	if generator.calls >= len(generator.ids) {
		return uuid.Nil
	}
	id := generator.ids[generator.calls]
	generator.calls++
	return id
}

type createSerializerUnused struct{}

func (createSerializerUnused) Serialize(calendar.ITIPMessage) ([]byte, error) {
	panic("create calendar event must not serialize or send iTIP outside the transactional outbox")
}
