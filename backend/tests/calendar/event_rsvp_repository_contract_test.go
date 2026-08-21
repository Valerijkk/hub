package calendar_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/platform/events"
)

var (
	rsvpRepositoryActorID        = uuid.MustParse("b1111111-1111-4111-8111-111111111111")
	rsvpRepositoryOrganizationID = uuid.MustParse("b2222222-2222-4222-8222-222222222222")
	rsvpRepositoryEventID        = uuid.MustParse("b3333333-3333-4333-8333-333333333333")
	rsvpRepositoryAttendeeID     = uuid.MustParse("b4444444-4444-4444-8444-444444444444")
	rsvpRepositoryEnvelopeID     = uuid.MustParse("b5555555-5555-4555-8555-555555555555")
	rsvpRepositoryOtherID        = uuid.MustParse("b6666666-6666-4666-8666-666666666666")
)

var _ calendar.CalendarEventSeriesRSVPTransactionBackend = (*rsvpRepositoryBackend)(nil)

func TestCalendarEventSeriesRSVPPersistenceValues_ValidateAndIsolateAliases(t *testing.T) {
	attendee := rsvpRepositoryAttendee(calendar.RSVPAccepted)
	reply := rsvpRepositoryReply(attendee, 7)
	outbox := rsvpRepositoryOutbox(t)

	command, err := calendar.NewCalendarEventSeriesRSVPPersistenceCommand(attendee, 2, 7, outbox, &reply)
	if err != nil {
		t.Fatalf("NewCalendarEventSeriesRSVPPersistenceCommand() error = %v", err)
	}
	result, err := calendar.NewCalendarEventSeriesRSVPPersistenceResult(attendee, 7, outbox, &reply)
	if err != nil {
		t.Fatalf("NewCalendarEventSeriesRSVPPersistenceResult() error = %v", err)
	}

	*attendee.UserID = rsvpRepositoryOtherID
	*attendee.RespondedAt = attendee.RespondedAt.Add(time.Hour)
	reply.Status = calendar.RSVPDeclined
	commandReply, commandHasReply := command.Reply()
	resultReply, resultHasReply := result.Reply()

	for _, value := range []struct {
		name              string
		attendee          calendar.CalendarEventSeriesRSVPAttendeeSnapshot
		sequence          uint
		expectedVersion   uint
		reply             calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
		hasReply          bool
	}{
		{name: "command", attendee: command.Attendee(), sequence: command.EventSequence(), expectedVersion: command.ExpectedVersion(), reply: mustRSVPRepositoryReply(t, commandReply, commandHasReply), hasReply: commandHasReply},
		{name: "result", attendee: result.Attendee(), sequence: result.EventSequence(), reply: mustRSVPRepositoryReply(t, resultReply, resultHasReply), hasReply: resultHasReply},
	} {
		t.Run(value.name, func(t *testing.T) {
			if value.attendee.UserID == nil || *value.attendee.UserID != rsvpRepositoryActorID || value.attendee.RespondedAt == nil || !value.attendee.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
				t.Fatalf("%s attendee aliases input: %#v", value.name, value.attendee)
			}
			if value.sequence != 7 || (value.name == "command" && value.expectedVersion != 2) || !value.hasReply || value.reply.Status != calendar.RSVPAccepted {
				t.Fatalf("%s getters = sequence:%d expected:%d reply:%#v", value.name, value.sequence, value.expectedVersion, value.reply)
			}
			*value.attendee.UserID = rsvpRepositoryOtherID
			*value.attendee.RespondedAt = value.attendee.RespondedAt.Add(time.Hour)
			if got := value.reply; got.Status != calendar.RSVPAccepted {
				t.Fatalf("%s reply getter = %#v", value.name, got)
			}
		})
	}

	if got := command.Attendee(); got.UserID == nil || *got.UserID != rsvpRepositoryActorID || got.RespondedAt == nil || !got.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
		t.Fatalf("command getter leaked alias: %#v", got)
	}
	if got := result.Attendee(); got.UserID == nil || *got.UserID != rsvpRepositoryActorID || got.RespondedAt == nil || !got.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
		t.Fatalf("result getter leaked alias: %#v", got)
	}

	for _, testCase := range []struct {
		name           string
		attendee       calendar.CalendarEventSeriesRSVPAttendeeSnapshot
		expected       uint
		sequence       uint
		reply          *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
		commandInvalid bool
		resultInvalid  bool
	}{
		{name: "zero expected version", attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted), expected: 0, sequence: 7, reply: rsvpRepositoryReplyPointer(rsvpRepositoryAttendee(calendar.RSVPAccepted), 7), commandInvalid: true},
		{name: "zero sequence", attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted), expected: 2, sequence: 0, reply: rsvpRepositoryReplyPointer(rsvpRepositoryAttendee(calendar.RSVPAccepted), 7), commandInvalid: true, resultInvalid: true},
		{name: "overflow expected version", attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted), expected: ^uint(0), sequence: 7, reply: rsvpRepositoryReplyPointer(rsvpRepositoryAttendee(calendar.RSVPAccepted), 7), commandInvalid: true},
		{name: "optimistic version mismatch", attendee: rsvpRepositoryAttendeeWithVersion(calendar.RSVPAccepted, 2), expected: 2, sequence: 7, reply: rsvpRepositoryReplyPointer(rsvpRepositoryAttendeeWithVersion(calendar.RSVPAccepted, 2), 7), commandInvalid: true},
		{name: "needs action responded", attendee: rsvpRepositoryAttendeeWithResponse(calendar.RSVPNeedsAction, time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)), expected: 2, sequence: 7, commandInvalid: true, resultInvalid: true},
		{name: "accepted without response", attendee: rsvpRepositoryAttendeeWithResponse(calendar.RSVPAccepted, time.Time{}), expected: 2, sequence: 7, commandInvalid: true, resultInvalid: true},
		{name: "non UTC response", attendee: rsvpRepositoryAttendeeWithResponse(calendar.RSVPAccepted, time.Date(2032, time.February, 3, 4, 5, 6, 0, time.FixedZone("UTC+3", 3*60*60))), expected: 2, sequence: 7, commandInvalid: true, resultInvalid: true},
		{name: "noncanonical email", attendee: rsvpRepositoryAttendeeWithEmail(" Attendee@Example.Test "), expected: 2, sequence: 7, commandInvalid: true, resultInvalid: true},
		{name: "removed attendee", attendee: rsvpRepositoryAttendeeWithRemovedAt(), expected: 2, sequence: 7, commandInvalid: true, resultInvalid: true},
		{name: "reply mismatch", attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted), expected: 2, sequence: 7, reply: &calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor{AttendeeID: rsvpRepositoryAttendeeID, EventID: rsvpRepositoryEventID, EventSequence: 8, AttendeeVersion: 3, Status: calendar.RSVPAccepted}, commandInvalid: true, resultInvalid: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, commandErr := calendar.NewCalendarEventSeriesRSVPPersistenceCommand(testCase.attendee, testCase.expected, testCase.sequence, outbox, testCase.reply)
			if testCase.commandInvalid != errors.Is(commandErr, calendar.ErrEventInvalid) {
				t.Fatalf("command error = %v, want invalid:%t", commandErr, testCase.commandInvalid)
			}
			_, resultErr := calendar.NewCalendarEventSeriesRSVPPersistenceResult(testCase.attendee, testCase.sequence, outbox, testCase.reply)
			if testCase.resultInvalid != errors.Is(resultErr, calendar.ErrEventInvalid) {
				t.Fatalf("result error = %v, want invalid:%t", resultErr, testCase.resultInvalid)
			}
		})
	}
}

func TestEventTransactionSeriesRSVP_RequiresCapabilityAndClosesScope(t *testing.T) {
	base := &rsvpBaseOnlyBackend{organizationID: rsvpRepositoryOrganizationID}
	err := calendar.RunEventTransaction(base, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("unsupported scope error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if base.organizationCalls != 0 || base.otherCalls != 0 {
		t.Fatalf("unsupported capability made backend calls org:%d other:%d", base.organizationCalls, base.otherCalls)
	}

	var typedNil *rsvpRepositoryBackend
	err = calendar.RunEventTransaction(typedNil, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("typed-nil scope error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction(typed nil) error = %v", err)
	}

	backend := newRSVPRepositoryBackend()
	var retained *calendar.EventTransaction
	var scope calendar.EventMutationScope
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		var err error
		scope, err = transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		if scope.ActorID() != rsvpRepositoryActorID || scope.OrganizationID() != rsvpRepositoryOrganizationID || scope.Operation() != calendar.MutationReceiptRSVPSeries || scope.EventOperation() != calendar.EventMutationRSVPSeries {
			t.Fatalf("RSVP scope = %#v", scope)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() supported capability error = %v", err)
	}
	if backend.organizationCalls != 1 {
		t.Fatalf("organization calls = %d, want 1", backend.organizationCalls)
	}
	if _, err := retained.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped scope error = %v, want transaction closed", err)
	}
	if _, err := retained.ResolveCalendarEventSeriesRSVP(context.Background(), scope, calendar.CalendarEventSeriesRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID}); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped resolve error = %v, want transaction closed", err)
	}
}

func TestEventTransactionSeriesRSVP_ResolvesOnlyBoundAuthorizationAndCopiesIt(t *testing.T) {
	backend := newRSVPRepositoryBackend()
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		if _, err := transaction.ResolveCalendarEventSeriesRSVP(context.Background(), scope, calendar.CalendarEventSeriesRSVPAuthorizationRequest{}); !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("zero request error = %v, want invalid", err)
		}
		if backend.resolveCalls != 0 {
			t.Fatalf("zero request reached resolver: %d", backend.resolveCalls)
		}

		authorization, err := transaction.ResolveCalendarEventSeriesRSVP(context.Background(), scope, calendar.CalendarEventSeriesRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID})
		if err != nil {
			return err
		}
		if backend.resolveCalls != 1 || backend.lastRequest.EventID != rsvpRepositoryEventID {
			t.Fatalf("resolver calls:%d request:%#v", backend.resolveCalls, backend.lastRequest)
		}
		if authorization.Current.EventID != rsvpRepositoryEventID || authorization.Current.OrganizationID != rsvpRepositoryOrganizationID || authorization.Attendee.UserID == nil || *authorization.Attendee.UserID != rsvpRepositoryActorID {
			t.Fatalf("authorization = %#v", authorization)
		}
		*authorization.Current.OrganizerID = rsvpRepositoryOtherID
		authorization.Current.Exdates[0] = "2032-02-05T04:05:06Z"
		*authorization.Attendee.UserID = rsvpRepositoryOtherID
		*authorization.Attendee.RespondedAt = authorization.Attendee.RespondedAt.Add(time.Hour)
		if backend.authorization.Current.OrganizerID == nil || *backend.authorization.Current.OrganizerID != rsvpRepositoryActorID || backend.authorization.Current.Exdates[0] != "2032-02-04T04:05:06Z" || backend.authorization.Attendee.UserID == nil || *backend.authorization.Attendee.UserID != rsvpRepositoryActorID || backend.authorization.Attendee.RespondedAt == nil || !backend.authorization.Attendee.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
			t.Fatalf("authorization output retained backend aliases: %#v", backend.authorization)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*rsvpRepositoryBackend)
	}{
		{name: "foreign event", mutate: func(backend *rsvpRepositoryBackend) { backend.authorization.Current.EventID = rsvpRepositoryOtherID }},
		{name: "foreign organization", mutate: func(backend *rsvpRepositoryBackend) { backend.authorization.Current.OrganizationID = rsvpRepositoryOtherID }},
		{name: "wrong attendee actor", mutate: func(backend *rsvpRepositoryBackend) { backend.authorization.Attendee.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID) }},
		{name: "foreign attendee event", mutate: func(backend *rsvpRepositoryBackend) { backend.authorization.Attendee.EventID = rsvpRepositoryOtherID }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newRSVPRepositoryBackend()
			testCase.mutate(backend)
			backend.authorization.Current.Title = "attacker controlled title"
			backend.authorization.Attendee.Email = "attacker@example.test"
			authorization := calendar.CalendarEventSeriesRSVPAuthorization{}
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				authorization, err = transaction.ResolveCalendarEventSeriesRSVP(context.Background(), scope, calendar.CalendarEventSeriesRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID})
				return err
			})
			if err != nil {
				t.Fatalf("RunEventTransaction() error = %v", err)
			}
			if authorization.Current.Title != "" || authorization.Current.Description != "" || authorization.Attendee.Email != "" || authorization.Attendee.DisplayName != "" || authorization.Attendee.RespondedAt != nil {
				t.Fatalf("identity mask retained variable material: %#v", authorization)
			}
			if authorization.Current.EventID != backend.authorization.Current.EventID || authorization.Current.OrganizationID != backend.authorization.Current.OrganizationID || authorization.Attendee.AttendeeID != backend.authorization.Attendee.AttendeeID || authorization.Attendee.EventID != backend.authorization.Attendee.EventID || authorization.Attendee.UserID == nil || backend.authorization.Attendee.UserID == nil || *authorization.Attendee.UserID != *backend.authorization.Attendee.UserID {
				t.Fatalf("identity mask = %#v", authorization)
			}
		})
	}
}

func TestEventTransactionSeriesRSVP_PersistsOnlyExactActorBoundEcho(t *testing.T) {
	command := rsvpRepositoryCommand(t, rsvpRepositoryAttendee(calendar.RSVPAccepted), 2, 7, true)
	backend := newRSVPRepositoryBackend()
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		result, err := transaction.PersistCalendarEventSeriesRSVP(context.Background(), scope, command)
		if err != nil {
			return err
		}
		if result.Attendee().AttendeeID != rsvpRepositoryAttendeeID || result.EventSequence() != 7 || backend.persistCalls != 1 || len(backend.persisted) != 1 {
			t.Fatalf("result:%#v persist calls:%d commands:%d", result, backend.persistCalls, len(backend.persisted))
		}
		persistedReplyValue, persistedReplyFound := backend.persisted[0].Reply()
		persistedReply := mustRSVPRepositoryReply(t, persistedReplyValue, persistedReplyFound)
		if persistedReply.AttendeeID != rsvpRepositoryAttendeeID || persistedReply.EventSequence != 7 || persistedReply.AttendeeVersion != 3 || persistedReply.Status != calendar.RSVPAccepted {
			t.Fatalf("persisted reply = %#v", persistedReply)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*rsvpRepositoryBackend)
	}{
		{name: "zero backend result", mutate: func(backend *rsvpRepositoryBackend) { backend.returnZeroResult = true }},
		{name: "mismatched backend result", mutate: func(backend *rsvpRepositoryBackend) {
			backend.resultOverride = func(command calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error) {
				return calendar.NewCalendarEventSeriesRSVPPersistenceResult(command.Attendee(), command.EventSequence()+1, command.Outbox(), nil)
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newRSVPRepositoryBackend()
			testCase.mutate(backend)
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				_, err = transaction.PersistCalendarEventSeriesRSVP(context.Background(), scope, command)
				if !errors.Is(err, calendar.ErrEventInvalid) {
					t.Fatalf("PersistCalendarEventSeriesRSVP() error = %v, want invalid", err)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("RunEventTransaction() error = %v", err)
			}
			if backend.persistCalls != 1 {
				t.Fatalf("persist calls = %d, want 1", backend.persistCalls)
			}
		})
	}

	wrongActorAttendee := rsvpRepositoryAttendee(calendar.RSVPAccepted)
	wrongActorAttendee.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID)
	wrongActorCommand := rsvpRepositoryCommand(t, wrongActorAttendee, 2, 7, true)
	backend = newRSVPRepositoryBackend()
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventSeriesRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		_, err = transaction.PersistCalendarEventSeriesRSVP(context.Background(), scope, wrongActorCommand)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("foreign actor persistence error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}
	if backend.persistCalls != 0 {
		t.Fatalf("foreign actor reached persistence: %d", backend.persistCalls)
	}
}

func rsvpRepositoryCommand(t *testing.T, attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot, expectedVersion, sequence uint, withReply bool) calendar.CalendarEventSeriesRSVPPersistenceCommand {
	t.Helper()
	var reply *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
	if withReply {
		value := rsvpRepositoryReply(attendee, sequence)
		reply = &value
	}
	command, err := calendar.NewCalendarEventSeriesRSVPPersistenceCommand(attendee, expectedVersion, sequence, rsvpRepositoryOutbox(t), reply)
	if err != nil {
		t.Fatalf("NewCalendarEventSeriesRSVPPersistenceCommand() error = %v", err)
	}
	return command
}

func rsvpRepositoryAttendee(status calendar.RSVPStatus) calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	if status == calendar.RSVPNeedsAction {
		return rsvpRepositoryAttendeeWithResponse(status, time.Time{})
	}
	return rsvpRepositoryAttendeeWithResponse(status, time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC))
}

func rsvpRepositoryAttendeeWithVersion(status calendar.RSVPStatus, version uint) calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	attendee := rsvpRepositoryAttendee(status)
	attendee.Version = version
	return attendee
}

func rsvpRepositoryAttendeeWithResponse(status calendar.RSVPStatus, respondedAt time.Time) calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	attendee := calendar.CalendarEventSeriesRSVPAttendeeSnapshot{
		AttendeeID: rsvpRepositoryAttendeeID,
		EventID:    rsvpRepositoryEventID,
		UserID:     rsvpRepositoryUUIDRef(rsvpRepositoryActorID),
		Email:      "attendee@example.test",
		DisplayName: "Attendee",
		Role:       calendar.CalendarEventAttendeeRoleRequired,
		Status:     status,
		Version:    3,
	}
	if !respondedAt.IsZero() {
		attendee.RespondedAt = &respondedAt
	}
	return attendee
}

func rsvpRepositoryAttendeeWithEmail(email string) calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	attendee := rsvpRepositoryAttendee(calendar.RSVPAccepted)
	attendee.Email = email
	return attendee
}

func rsvpRepositoryAttendeeWithRemovedAt() calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	attendee := rsvpRepositoryAttendee(calendar.RSVPAccepted)
	removedAt := time.Date(2032, time.February, 3, 5, 5, 6, 0, time.UTC)
	attendee.RemovedAt = &removedAt
	return attendee
}

func rsvpRepositoryReply(attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot, sequence uint) calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	return calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor{AttendeeID: attendee.AttendeeID, EventID: attendee.EventID, EventSequence: sequence, AttendeeVersion: attendee.Version, Status: attendee.Status}
}

func rsvpRepositoryReplyPointer(attendee calendar.CalendarEventSeriesRSVPAttendeeSnapshot, sequence uint) *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	reply := rsvpRepositoryReply(attendee, sequence)
	return &reply
}

func mustRSVPRepositoryReply(t *testing.T, reply calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor, found bool) calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	t.Helper()
	if !found {
		t.Fatal("Reply() found = false, want true")
	}
	return reply
}

func rsvpRepositoryOutbox(t *testing.T) calendar.EventOutboxDescriptor {
	t.Helper()
	actor := rsvpRepositoryActorID
	envelope, err := events.NewEnvelopeV1("calendar.meeting.rsvp.changed.v1", rsvpRepositoryEnvelopeID, time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC), rsvpRepositoryOrganizationID, &actor, rsvpRepositoryEventID.String(), map[string]any{"event_id": rsvpRepositoryEventID.String()})
	if err != nil {
		t.Fatalf("events.NewEnvelopeV1() error = %v", err)
	}
	descriptor, err := calendar.NewEventOutboxDescriptor(envelope)
	if err != nil {
		t.Fatalf("NewEventOutboxDescriptor() error = %v", err)
	}
	return descriptor
}

func rsvpRepositoryCurrent() calendar.CalendarEventCommandInput {
	return calendar.CalendarEventCommandInput{
		EventID: rsvpRepositoryEventID, OrganizationID: rsvpRepositoryOrganizationID, OwnerID: rsvpRepositoryActorID,
		UID: "hub-" + rsvpRepositoryEventID.String(), Title: "Private title", Description: "Private description", Location: "Private location",
		Temporal: calendar.CalendarEvent{StartsAt: time.Date(2032, time.February, 4, 4, 5, 6, 0, time.UTC), EndsAt: time.Date(2032, time.February, 4, 5, 5, 6, 0, time.UTC), Timezone: "UTC"},
		Exdates: []string{"2032-02-04T04:05:06Z"}, Sequence: 6, Version: 5, Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed,
		OrganizerID: rsvpRepositoryUUIDRef(rsvpRepositoryActorID), OrganizerEmail: "owner@example.test",
	}
}

func rsvpRepositoryUUIDRef(value uuid.UUID) *uuid.UUID {
	return &value
}

type rsvpBaseOnlyBackend struct {
	organizationID    uuid.UUID
	organizationCalls int
	otherCalls        int
}

func (backend *rsvpBaseOnlyBackend) ResolveActorOrganization(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	backend.organizationCalls++
	return backend.organizationID, nil
}
func (backend *rsvpBaseOnlyBackend) ResolveCalendarEventCreateAuthorization(context.Context, calendar.EventMutationScope, calendar.CalendarEventCreateAuthorizationRequest) (calendar.CalendarEventCreateAuthorization, error) {
	backend.otherCalls++
	return calendar.CalendarEventCreateAuthorization{}, calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) ReserveMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptReservation) (calendar.MutationReceiptReservationResult, error) {
	backend.otherCalls++
	return calendar.MutationReceiptReservationResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) CompleteMutationReceipt(context.Context, calendar.EventMutationScope, calendar.MutationReceiptCompletion) error {
	backend.otherCalls++
	return calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) PersistCalendarEvent(context.Context, calendar.EventMutationScope, calendar.CalendarEventCommand) (calendar.CalendarEventPersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventPersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) PersistCalendarEventAttendee(context.Context, calendar.EventMutationScope, calendar.CalendarEventAttendeeCommand) (calendar.CalendarEventAttendeePersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventAttendeePersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) PersistCalendarEventException(context.Context, calendar.EventMutationScope, calendar.CalendarEventExceptionCommand) (calendar.CalendarEventExceptionPersistenceResult, error) {
	backend.otherCalls++
	return calendar.CalendarEventExceptionPersistenceResult{}, calendar.ErrEventInvalid
}
func (backend *rsvpBaseOnlyBackend) AppendOutbox(context.Context, calendar.EventMutationScope, calendar.EventOutboxDescriptor) error {
	backend.otherCalls++
	return calendar.ErrEventInvalid
}

type rsvpRepositoryBackend struct {
	rsvpBaseOnlyBackend
	authorization    calendar.CalendarEventSeriesRSVPAuthorization
	lastRequest      calendar.CalendarEventSeriesRSVPAuthorizationRequest
	persisted        []calendar.CalendarEventSeriesRSVPPersistenceCommand
	resolveCalls     int
	persistCalls     int
	returnZeroResult bool
	resultOverride   func(calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error)
}

func newRSVPRepositoryBackend() *rsvpRepositoryBackend {
	return &rsvpRepositoryBackend{
		rsvpBaseOnlyBackend: rsvpBaseOnlyBackend{organizationID: rsvpRepositoryOrganizationID},
		authorization: calendar.CalendarEventSeriesRSVPAuthorization{
			Current:  rsvpRepositoryCurrent(),
			Attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted),
		},
	}
}

func (backend *rsvpRepositoryBackend) ResolveCalendarEventSeriesRSVP(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventSeriesRSVPAuthorizationRequest) (calendar.CalendarEventSeriesRSVPAuthorization, error) {
	backend.resolveCalls++
	backend.lastRequest = request
	return backend.authorization, nil
}

func (backend *rsvpRepositoryBackend) PersistCalendarEventSeriesRSVP(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventSeriesRSVPPersistenceCommand) (calendar.CalendarEventSeriesRSVPPersistenceResult, error) {
	backend.persistCalls++
	if backend.returnZeroResult {
		return calendar.CalendarEventSeriesRSVPPersistenceResult{}, nil
	}
	if backend.resultOverride != nil {
		return backend.resultOverride(command)
	}
	backend.persisted = append(backend.persisted, command)
	return calendar.NewCalendarEventSeriesRSVPPersistenceResult(command.Attendee(), command.EventSequence(), command.Outbox(), mustRSVPRepositoryReplyPointer(command))
}

func mustRSVPRepositoryReplyPointer(command calendar.CalendarEventSeriesRSVPPersistenceCommand) *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor {
	reply, found := command.Reply()
	if !found {
		return nil
	}
	return &reply
}
