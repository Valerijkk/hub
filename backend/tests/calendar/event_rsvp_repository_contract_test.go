package calendar_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
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
	rsvpRepositoryOccurrenceID   = uuid.MustParse("b7777777-7777-4777-8777-777777777777")
)

var _ calendar.CalendarEventSeriesRSVPTransactionBackend = (*rsvpRepositoryBackend)(nil)
var _ calendar.CalendarEventOccurrenceRSVPTransactionBackend = (*occurrenceRSVPRepositoryBackend)(nil)

var (
	_ func(*calendar.EventTransaction, context.Context, uuid.UUID) (calendar.EventMutationScope, error) = (*calendar.EventTransaction).ResolveCalendarEventOccurrenceRSVPScope
	_ func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) = (*calendar.EventTransaction).ResolveCalendarEventOccurrenceRSVP
	_ func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) = (*calendar.EventTransaction).PersistCalendarEventOccurrenceRSVP

	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) calendar.CalendarEventSeriesRSVPAttendeeSnapshot = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.SeriesAttendee
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, bool) = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.Prior
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.Attendee
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) uint = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.ExpectedVersion
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) uint = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.EventSequence
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) calendar.EventOutboxDescriptor = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.Outbox
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, bool) = calendar.CalendarEventOccurrenceRSVPPersistenceCommand.Reply

	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceResult) calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot = calendar.CalendarEventOccurrenceRSVPPersistenceResult.Attendee
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceResult) uint = calendar.CalendarEventOccurrenceRSVPPersistenceResult.EventSequence
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceResult) calendar.EventOutboxDescriptor = calendar.CalendarEventOccurrenceRSVPPersistenceResult.Outbox
	_ func(calendar.CalendarEventOccurrenceRSVPPersistenceResult) (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, bool) = calendar.CalendarEventOccurrenceRSVPPersistenceResult.Reply

	_ func(calendar.CalendarEventSeriesRSVPAttendeeSnapshot, *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, uint, uint, calendar.EventOutboxDescriptor, *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) (calendar.CalendarEventOccurrenceRSVPPersistenceCommand, error) = calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand
	_ func(calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, uint, calendar.EventOutboxDescriptor, *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) = calendar.NewCalendarEventOccurrenceRSVPPersistenceResult
)

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
		{name: "zero sequence reply mismatch", attendee: rsvpRepositoryAttendee(calendar.RSVPAccepted), expected: 2, sequence: 0, reply: rsvpRepositoryReplyPointer(rsvpRepositoryAttendee(calendar.RSVPAccepted), 7), commandInvalid: true, resultInvalid: true},
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

func TestCalendarEventSeriesRSVPPersistenceValues_AcceptsInitialEventSequence(t *testing.T) {
	attendee := rsvpRepositoryAttendee(calendar.RSVPAccepted)
	outbox := rsvpRepositoryOutbox(t)
	reply := rsvpRepositoryReply(attendee, 0)

	for _, testCase := range []struct {
		name      string
		reply     *calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
		hasReply  bool
	} {
		{name: "internal no reply"},
		{name: "external reply", reply: &reply, hasReply: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			command, err := calendar.NewCalendarEventSeriesRSVPPersistenceCommand(attendee, 2, 0, outbox, testCase.reply)
			if err != nil {
				t.Fatalf("NewCalendarEventSeriesRSVPPersistenceCommand() error = %v", err)
			}
			result, err := calendar.NewCalendarEventSeriesRSVPPersistenceResult(attendee, 0, outbox, testCase.reply)
			if err != nil {
				t.Fatalf("NewCalendarEventSeriesRSVPPersistenceResult() error = %v", err)
			}
			commandReply, commandHasReply := command.Reply()
			resultReply, resultHasReply := result.Reply()

			for _, value := range []struct {
				name     string
				sequence uint
				reply    calendar.CalendarEventSeriesRSVPReplyDeliveryDescriptor
				hasReply bool
			}{
				{name: "command", sequence: command.EventSequence(), reply: commandReply, hasReply: commandHasReply},
				{name: "result", sequence: result.EventSequence(), reply: resultReply, hasReply: resultHasReply},
			} {
				if value.sequence != 0 || value.hasReply != testCase.hasReply {
					t.Fatalf("%s sequence:%d has reply:%t", value.name, value.sequence, value.hasReply)
				}
				if value.hasReply && (value.reply.AttendeeID != attendee.AttendeeID || value.reply.EventID != attendee.EventID || value.reply.EventSequence != 0 || value.reply.AttendeeVersion != attendee.Version || value.reply.Status != attendee.Status) {
					t.Fatalf("%s reply = %#v", value.name, value.reply)
				}
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

func TestCalendarEventOccurrenceRSVPRepositoryABI(t *testing.T) {
	transactionType := reflect.TypeOf((*calendar.EventTransaction)(nil))
	for name, expected := range map[string]reflect.Type{
		"ResolveCalendarEventOccurrenceRSVPScope": reflect.TypeOf(func(*calendar.EventTransaction, context.Context, uuid.UUID) (calendar.EventMutationScope, error) { return calendar.EventMutationScope{}, nil }),
		"ResolveCalendarEventOccurrenceRSVP":      reflect.TypeOf(func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) { return calendar.CalendarEventOccurrenceRSVPAuthorization{}, nil }),
		"PersistCalendarEventOccurrenceRSVP":      reflect.TypeOf(func(*calendar.EventTransaction, context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) { return calendar.CalendarEventOccurrenceRSVPPersistenceResult{}, nil }),
	} {
		method, found := transactionType.MethodByName(name)
		if !found || method.Type != expected {
			t.Fatalf("%s ABI = %#v", name, method.Type)
		}
	}

	backendType := reflect.TypeOf((*calendar.CalendarEventOccurrenceRSVPTransactionBackend)(nil)).Elem()
	for name, expected := range map[string]reflect.Type{
		"ResolveCalendarEventOccurrenceRSVP": reflect.TypeOf(func(context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) { return calendar.CalendarEventOccurrenceRSVPAuthorization{}, nil }),
		"PersistCalendarEventOccurrenceRSVP": reflect.TypeOf(func(context.Context, calendar.EventMutationScope, calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) { return calendar.CalendarEventOccurrenceRSVPPersistenceResult{}, nil }),
	} {
		method, found := backendType.MethodByName(name)
		if !found || method.Type != expected {
			t.Fatalf("backend %s ABI = %#v", name, method.Type)
		}
	}

	for _, record := range []struct {
		name   string
		typeOf reflect.Type
		fields []struct {
			name   string
			typeOf reflect.Type
		}
	}{
		{
			name: "authorization request", typeOf: reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{}),
			fields: []struct {
				name   string
				typeOf reflect.Type
			}{
				{name: "EventID", typeOf: reflect.TypeOf(uuid.UUID{})},
				{name: "RecurrenceID", typeOf: reflect.TypeOf("")},
			},
		},
		{
			name: "occurrence attendee", typeOf: reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot{}),
			fields: []struct {
				name   string
				typeOf reflect.Type
			}{
				{name: "AttendeeID", typeOf: reflect.TypeOf(uuid.UUID{})},
				{name: "EventID", typeOf: reflect.TypeOf(uuid.UUID{})},
				{name: "UserID", typeOf: reflect.TypeOf((*uuid.UUID)(nil))},
				{name: "Email", typeOf: reflect.TypeOf("")},
				{name: "DisplayName", typeOf: reflect.TypeOf("")},
				{name: "Role", typeOf: reflect.TypeOf(calendar.CalendarEventAttendeeRole(""))},
				{name: "Status", typeOf: reflect.TypeOf(calendar.RSVPStatus(""))},
				{name: "RecurrenceID", typeOf: reflect.TypeOf("")},
				{name: "RespondedAt", typeOf: reflect.TypeOf((*time.Time)(nil))},
				{name: "Version", typeOf: reflect.TypeOf(uint(0))},
				{name: "RemovedAt", typeOf: reflect.TypeOf((*time.Time)(nil))},
			},
		},
		{
			name: "authorization", typeOf: reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPAuthorization{}),
			fields: []struct {
				name   string
				typeOf reflect.Type
			}{
				{name: "Current", typeOf: reflect.TypeOf(calendar.CalendarEventCommandInput{})},
				{name: "SeriesAttendee", typeOf: reflect.TypeOf(calendar.CalendarEventSeriesRSVPAttendeeSnapshot{})},
				{name: "OccurrenceAttendee", typeOf: reflect.TypeOf((*calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot)(nil))},
				{name: "Exception", typeOf: reflect.TypeOf((*calendar.CalendarEventExceptionCommandInput)(nil))},
			},
		},
		{
			name: "reply descriptor", typeOf: reflect.TypeOf(calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}),
			fields: []struct {
				name   string
				typeOf reflect.Type
			}{
				{name: "AttendeeID", typeOf: reflect.TypeOf(uuid.UUID{})},
				{name: "EventID", typeOf: reflect.TypeOf(uuid.UUID{})},
				{name: "EventSequence", typeOf: reflect.TypeOf(uint(0))},
				{name: "AttendeeVersion", typeOf: reflect.TypeOf(uint(0))},
				{name: "RecurrenceID", typeOf: reflect.TypeOf("")},
				{name: "Status", typeOf: reflect.TypeOf(calendar.RSVPStatus(""))},
			},
		},
	} {
		if record.typeOf.NumField() != len(record.fields) {
			t.Fatalf("%s field count = %d, want %d", record.name, record.typeOf.NumField(), len(record.fields))
		}
		for index, expected := range record.fields {
			field := record.typeOf.Field(index)
			if field.Name != expected.name || field.Type != expected.typeOf || field.Anonymous {
				t.Fatalf("%s field[%d] ABI = %#v, want %s %v", record.name, index, field, expected.name, expected.typeOf)
			}
		}
	}
}

func TestCalendarEventOccurrenceRSVPPersistenceValues_ValidateAndIsolateAliases(t *testing.T) {
	series := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	attendee := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	reply := occurrenceRSVPRepositoryReply(attendee, 0)
	outbox := rsvpRepositoryOutbox(t)

	command, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand(series, nil, attendee, 2, 0, outbox, &reply)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceCommand() error = %v", err)
	}
	result, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(attendee, 0, outbox, &reply)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceResult() error = %v", err)
	}
	noReplyCommand, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand(series, nil, attendee, 2, 0, outbox, nil)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceCommand(no reply) error = %v", err)
	}
	noReplyResult, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(attendee, 0, outbox, nil)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceResult(no reply) error = %v", err)
	}
	if noReply, found := noReplyCommand.Reply(); found || noReply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}) {
		t.Fatalf("command no-reply = %#v found:%t", noReply, found)
	}
	if noReply, found := noReplyResult.Reply(); found || noReply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}) {
		t.Fatalf("result no-reply = %#v found:%t", noReply, found)
	}

	*series.UserID = rsvpRepositoryOtherID
	*series.RespondedAt = series.RespondedAt.Add(time.Hour)
	*attendee.UserID = rsvpRepositoryOtherID
	*attendee.RespondedAt = attendee.RespondedAt.Add(time.Hour)
	reply.Status = calendar.RSVPDeclined

	for _, value := range []struct {
		name           string
		series         calendar.CalendarEventSeriesRSVPAttendeeSnapshot
		attendee       calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot
		expected       uint
		sequence       uint
		hasPrior       bool
		reply          calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
		hasReply       bool
	}{
		{name: "command", series: command.SeriesAttendee(), attendee: command.Attendee(), expected: command.ExpectedVersion(), sequence: command.EventSequence(), hasPrior: func() bool { _, found := command.Prior(); return found }(), reply: occurrenceRSVPRepositoryMustReply(t, command.Reply()), hasReply: func() bool { _, found := command.Reply(); return found }()},
		{name: "result", attendee: result.Attendee(), sequence: result.EventSequence(), reply: occurrenceRSVPRepositoryMustReply(t, result.Reply()), hasReply: func() bool { _, found := result.Reply(); return found }()},
	} {
		t.Run(value.name, func(t *testing.T) {
			if value.attendee.AttendeeID != rsvpRepositoryOccurrenceID || value.attendee.UserID == nil || *value.attendee.UserID != rsvpRepositoryActorID || value.attendee.RespondedAt == nil || !value.attendee.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) || value.attendee.RecurrenceID != "2032-02-04" || value.sequence != 0 || !value.hasReply || value.reply.RecurrenceID != "2032-02-04" || value.reply.Status != calendar.RSVPAccepted {
				t.Fatalf("%s immutable values = %#v", value.name, value)
			}
			if value.name == "command" && (value.series.Version != 2 || value.series.AttendeeID == value.attendee.AttendeeID || value.expected != 2 || value.hasPrior) {
				t.Fatalf("command materialization = %#v", value)
			}
			if value.attendee.UserID != nil {
				*value.attendee.UserID = rsvpRepositoryOtherID
			}
			if value.attendee.RespondedAt != nil {
				*value.attendee.RespondedAt = value.attendee.RespondedAt.Add(time.Hour)
			}
		})
	}
	if got := command.SeriesAttendee(); got.UserID == nil || *got.UserID != rsvpRepositoryActorID || got.RespondedAt == nil || !got.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
		t.Fatalf("series getter leaked alias: %#v", got)
	}
	if got := command.Attendee(); got.UserID == nil || *got.UserID != rsvpRepositoryActorID || got.RespondedAt == nil || !got.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
		t.Fatalf("command attendee getter leaked alias: %#v", got)
	}
	if got := result.Attendee(); got.UserID == nil || *got.UserID != rsvpRepositoryActorID || got.RespondedAt == nil || !got.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) {
		t.Fatalf("result attendee getter leaked alias: %#v", got)
	}

	series = occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	prior := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	updated := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPDeclined, 4)
	override, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand(series, &prior, updated, 3, 0, outbox, nil)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceCommand(existing override) error = %v", err)
	}
	returnedPrior, foundPrior := override.Prior()
	if !foundPrior || returnedPrior.AttendeeID != rsvpRepositoryOccurrenceID || returnedPrior.Version != 3 || override.Attendee().Version != 4 || override.ExpectedVersion() != 3 {
		t.Fatalf("existing override CAS = prior:%#v command:%#v", returnedPrior, override)
	}
	*prior.UserID = rsvpRepositoryOtherID
	if got, found := override.Prior(); !found || got.UserID == nil || *got.UserID != rsvpRepositoryActorID {
		t.Fatalf("prior getter leaked alias: %#v found:%t", got, found)
	}

	validSeries := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	validAttendee := occurrenceRSVPRepositoryAttendee(validSeries, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	validReply := occurrenceRSVPRepositoryReply(validAttendee, 0)
	invalidSeries := validSeries
	removedAt := time.Date(2032, time.February, 3, 5, 5, 6, 0, time.UTC)
	invalidSeries.RemovedAt = &removedAt
	sameID := validAttendee
	sameID.AttendeeID = validSeries.AttendeeID
	badVersion := validAttendee
	badVersion.Version = 4
	badKey := validAttendee
	badKey.RecurrenceID = "2032-02-04T04:05:06+00:00"
	tombstone := validAttendee
	tombstone.RemovedAt = &removedAt
	wrongEmail := validAttendee
	wrongEmail.Email = "other@example.test"
	wrongName := validAttendee
	wrongName.DisplayName = "Other attendee"
	wrongRole := validAttendee
	wrongRole.Role = calendar.CalendarEventAttendeeRoleOptional
	wrongEvent := validAttendee
	wrongEvent.EventID = rsvpRepositoryOtherID
	wrongUser := validAttendee
	wrongUser.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID)
	badResponse := validAttendee
	badResponse.Status = calendar.RSVPNeedsAction
	priorWrongID := validAttendee
	priorWrongID.AttendeeID = rsvpRepositoryOtherID
	priorWrongKey := validAttendee
	priorWrongKey.RecurrenceID = "2032-02-05"
	badReply := validReply
	badReply.RecurrenceID = "2032-02-05"
	zeroOutbox := calendar.EventOutboxDescriptor{}

	for _, testCase := range []struct {
		name     string
		series   calendar.CalendarEventSeriesRSVPAttendeeSnapshot
		prior    *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot
		attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot
		expected uint
		outbox   calendar.EventOutboxDescriptor
		reply    *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
	}{
		{name: "zero expected version", series: validSeries, attendee: validAttendee, expected: 0, outbox: outbox, reply: &validReply},
		{name: "overflow expected version", series: validSeries, attendee: validAttendee, expected: ^uint(0), outbox: outbox, reply: &validReply},
		{name: "invalid series", series: invalidSeries, attendee: validAttendee, expected: 2, outbox: outbox, reply: &validReply},
		{name: "inherited expected version is series version", series: validSeries, attendee: validAttendee, expected: 3, outbox: outbox, reply: &validReply},
		{name: "inherited attendee must be distinct", series: validSeries, attendee: sameID, expected: 2, outbox: outbox, reply: &validReply},
		{name: "inherited version increments", series: validSeries, attendee: badVersion, expected: 2, outbox: outbox, reply: &validReply},
		{name: "occurrence key canonical", series: validSeries, attendee: badKey, expected: 2, outbox: outbox, reply: &validReply},
		{name: "occurrence tombstone", series: validSeries, attendee: tombstone, expected: 2, outbox: outbox, reply: &validReply},
		{name: "occurrence response semantics", series: validSeries, attendee: badResponse, expected: 2, outbox: outbox, reply: &validReply},
		{name: "outbox invalid", series: validSeries, attendee: validAttendee, expected: 2, outbox: zeroOutbox, reply: &validReply},
		{name: "reply key mismatch", series: validSeries, attendee: validAttendee, expected: 2, outbox: outbox, reply: &badReply},
		{name: "existing override ID immutable", series: validSeries, prior: &priorWrongID, attendee: occurrenceRSVPRepositoryAttendee(validSeries, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPDeclined, 4), expected: 3, outbox: outbox},
		{name: "existing override key immutable", series: validSeries, prior: &priorWrongKey, attendee: occurrenceRSVPRepositoryAttendee(validSeries, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPDeclined, 4), expected: 3, outbox: outbox},
		{name: "existing override email immutable", series: validSeries, prior: &validAttendee, attendee: occurrenceRSVPRepositoryAttendeeFrom(validAttendee, calendar.RSVPDeclined, 4, func(value *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot) { value.Email = wrongEmail.Email }), expected: 3, outbox: outbox},
		{name: "existing override display name immutable", series: validSeries, prior: &validAttendee, attendee: occurrenceRSVPRepositoryAttendeeFrom(validAttendee, calendar.RSVPDeclined, 4, func(value *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot) { value.DisplayName = wrongName.DisplayName }), expected: 3, outbox: outbox},
		{name: "existing override role immutable", series: validSeries, prior: &validAttendee, attendee: occurrenceRSVPRepositoryAttendeeFrom(validAttendee, calendar.RSVPDeclined, 4, func(value *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot) { value.Role = wrongRole.Role }), expected: 3, outbox: outbox},
		{name: "occurrence event identity", series: validSeries, attendee: wrongEvent, expected: 2, outbox: outbox},
		{name: "occurrence user identity", series: validSeries, attendee: wrongUser, expected: 2, outbox: outbox},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand(testCase.series, testCase.prior, testCase.attendee, testCase.expected, 0, testCase.outbox, testCase.reply); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceCommand() error = %v, want invalid", err)
			}
		})
	}

	for _, testCase := range []struct {
		name     string
		attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot
		outbox   calendar.EventOutboxDescriptor
		reply    *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
	}{
		{name: "invalid attendee", attendee: tombstone, outbox: outbox},
		{name: "invalid outbox", attendee: validAttendee, outbox: zeroOutbox},
		{name: "invalid reply", attendee: validAttendee, outbox: outbox, reply: &badReply},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(testCase.attendee, 0, testCase.outbox, testCase.reply); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceResult() error = %v, want invalid", err)
			}
		})
	}
}

func TestEventTransactionOccurrenceRSVP_RequiresCapabilityAndClosesScope(t *testing.T) {
	base := &rsvpBaseOnlyBackend{organizationID: rsvpRepositoryOrganizationID}
	err := calendar.RunEventTransaction(base, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
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

	var typedNil *occurrenceRSVPRepositoryBackend
	err = calendar.RunEventTransaction(typedNil, func(transaction *calendar.EventTransaction) error {
		_, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("typed-nil scope error = %v, want invalid", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction(typed nil) error = %v", err)
	}

	backend := newOccurrenceRSVPRepositoryBackend()
	var retained *calendar.EventTransaction
	var scope calendar.EventMutationScope
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		retained = transaction
		var err error
		scope, err = transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		if scope.ActorID() != rsvpRepositoryActorID || scope.OrganizationID() != rsvpRepositoryOrganizationID || scope.Operation() != calendar.MutationReceiptRSVPOccurrence || scope.EventOperation() != calendar.EventMutationRSVPOccurrence {
			t.Fatalf("occurrence RSVP scope = %#v", scope)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() supported capability error = %v", err)
	}
	if backend.organizationCalls != 1 {
		t.Fatalf("organization calls = %d, want 1", backend.organizationCalls)
	}
	if _, err := retained.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped scope error = %v, want transaction closed", err)
	}
	if _, err := retained.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"}); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped resolve error = %v, want transaction closed", err)
	}
	series := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	attendee := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	command := occurrenceRSVPRepositoryCommand(t, series, nil, attendee, 2, 0, false)
	if _, err := retained.PersistCalendarEventOccurrenceRSVP(context.Background(), scope, command); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("escaped persist error = %v, want transaction closed", err)
	}
}

func TestEventTransactionOccurrenceRSVP_ResolvesCanonicalBoundSnapshots(t *testing.T) {
	backend := newOccurrenceRSVPRepositoryBackend()
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		for _, request := range []calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{
			{},
			{EventID: rsvpRepositoryEventID},
			{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04T04:05:06+00:00"},
			{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04\x00"},
			{EventID: rsvpRepositoryEventID, RecurrenceID: strings.Repeat("a", 256)},
		} {
			if _, err := transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, request); !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("ResolveCalendarEventOccurrenceRSVP(%#v) error = %v, want invalid", request, err)
			}
		}
		if backend.resolveCalls != 0 {
			t.Fatalf("invalid request reached resolver: %d", backend.resolveCalls)
		}

		authorization, err := transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"})
		if err != nil {
			return err
		}
		if backend.resolveCalls != 1 || backend.lastRequest.EventID != rsvpRepositoryEventID || backend.lastRequest.RecurrenceID != "2032-02-04" || authorization.OccurrenceAttendee == nil || authorization.Exception == nil {
			t.Fatalf("authorization = %#v resolver:%d request:%#v", authorization, backend.resolveCalls, backend.lastRequest)
		}
		if authorization.SeriesAttendee.AttendeeID == authorization.OccurrenceAttendee.AttendeeID || authorization.SeriesAttendee.Version != 2 || authorization.OccurrenceAttendee.Version != 3 || authorization.OccurrenceAttendee.RecurrenceID != "2032-02-04" {
			t.Fatalf("inherited materialization = %#v", authorization)
		}
		*authorization.Current.OrganizerID = rsvpRepositoryOtherID
		authorization.Current.Exdates[0] = "2032-02-05T04:05:06Z"
		*authorization.SeriesAttendee.UserID = rsvpRepositoryOtherID
		*authorization.OccurrenceAttendee.UserID = rsvpRepositoryOtherID
		*authorization.OccurrenceAttendee.RespondedAt = authorization.OccurrenceAttendee.RespondedAt.Add(time.Hour)
		authorization.Exception.RecurrenceID = "2032-02-05"
		authorization.Exception.Temporal.StartsAt = authorization.Exception.Temporal.StartsAt.Add(time.Hour)
		if backend.authorization.Current.OrganizerID == nil || *backend.authorization.Current.OrganizerID != rsvpRepositoryActorID || backend.authorization.Current.Exdates[0] != "2032-02-04T04:05:06Z" || backend.authorization.SeriesAttendee.UserID == nil || *backend.authorization.SeriesAttendee.UserID != rsvpRepositoryActorID || backend.authorization.OccurrenceAttendee == nil || backend.authorization.OccurrenceAttendee.UserID == nil || *backend.authorization.OccurrenceAttendee.UserID != rsvpRepositoryActorID || backend.authorization.OccurrenceAttendee.RespondedAt == nil || !backend.authorization.OccurrenceAttendee.RespondedAt.Equal(time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)) || backend.authorization.Exception == nil || backend.authorization.Exception.RecurrenceID != "2032-02-04" || backend.authorization.Exception.Temporal == nil || !backend.authorization.Exception.Temporal.StartsAt.Equal(time.Date(2032, time.February, 4, 4, 5, 6, 0, time.UTC)) {
			t.Fatalf("authorization output retained backend aliases: %#v", backend.authorization)
		}
		backend.authorization.OccurrenceAttendee.RecurrenceID = "2032-02-04T04:05:06Z"
		backend.authorization.Exception.RecurrenceID = "2032-02-04T04:05:06Z"
		if _, err := transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04T04:05:06Z"}); err != nil {
			t.Fatalf("timed canonical recurrence error = %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}

	for _, testCase := range []struct {
		name      string
		mutate    func(*occurrenceRSVPRepositoryBackend)
		wantError error
	}{
		{name: "different valid occurrence recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.OccurrenceAttendee.RecurrenceID = "2032-02-05" }},
		{name: "different valid exception recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Exception.RecurrenceID = "2032-02-05" }},
		{name: "malformed occurrence recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.OccurrenceAttendee.RecurrenceID = "2032-02-04T04:05:06+00:00" }, wantError: calendar.ErrEventInvalid},
		{name: "oversized occurrence recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.OccurrenceAttendee.RecurrenceID = strings.Repeat("x", 256) }, wantError: calendar.ErrEventInvalid},
		{name: "malformed exception recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Exception.RecurrenceID = "2032-02-04T04:05:06+00:00" }, wantError: calendar.ErrEventInvalid},
		{name: "oversized exception recurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Exception.RecurrenceID = strings.Repeat("x", 256) }, wantError: calendar.ErrEventInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			testCase.mutate(backend)
			authorization := calendar.CalendarEventOccurrenceRSVPAuthorization{}
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				authorization, err = transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"})
				return err
			})
			if testCase.wantError != nil {
				if !errors.Is(err, testCase.wantError) {
					t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v, want %v", err, testCase.wantError)
				}
				if !reflect.DeepEqual(authorization, calendar.CalendarEventOccurrenceRSVPAuthorization{}) {
					t.Fatalf("invalid recurrence authorization = %#v, want zero", authorization)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v", err)
			}
			if !reflect.DeepEqual(authorization, calendar.CalendarEventOccurrenceRSVPAuthorization{}) {
				t.Fatalf("different recurrence authorization = %#v, want exact zero mask", authorization)
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceRSVPRepositoryBackend)
	}{
		{name: "foreign event", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Current.EventID = rsvpRepositoryOtherID }},
		{name: "foreign organization", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Current.OrganizationID = rsvpRepositoryOtherID }},
		{name: "foreign series actor", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.SeriesAttendee.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID) }},
		{name: "foreign occurrence actor", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.OccurrenceAttendee.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID) }},
		{name: "foreign exception event", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Exception.EventID = rsvpRepositoryOtherID }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			testCase.mutate(backend)
			hostile := strings.Repeat("x", 255)
			backend.authorization.Current.Title = hostile
			backend.authorization.Current.Description = hostile
			backend.authorization.Current.Exdates = []string{hostile}
			backend.authorization.SeriesAttendee.Email = hostile
			backend.authorization.SeriesAttendee.DisplayName = hostile
			backend.authorization.OccurrenceAttendee.Email = hostile
			backend.authorization.OccurrenceAttendee.DisplayName = hostile
			backend.authorization.Exception.Temporal = &calendar.CalendarEvent{
				StartsAt: time.Date(2040, time.February, 4, 4, 5, 6, 0, time.UTC),
				EndsAt:   time.Date(2040, time.February, 4, 5, 5, 6, 0, time.UTC),
				Timezone: hostile,
			}
			authorization := calendar.CalendarEventOccurrenceRSVPAuthorization{}
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				authorization, err = transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"})
				return err
			})
			if err != nil {
				t.Fatalf("RunEventTransaction() error = %v", err)
			}
			if !reflect.DeepEqual(authorization, calendar.CalendarEventOccurrenceRSVPAuthorization{}) {
				t.Fatalf("foreign authorization mask = %#v, want exact zero mask", authorization)
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceRSVPRepositoryBackend)
	}{
		{name: "malformed series", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.SeriesAttendee.Email = " malformed@example.test " }},
		{name: "malformed occurrence", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.OccurrenceAttendee.Email = " malformed@example.test " }},
		{name: "malformed exception", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.authorization.Exception.Temporal = &calendar.CalendarEvent{} }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			testCase.mutate(backend)
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				_, err = transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"})
				if !errors.Is(err, calendar.ErrEventInvalid) {
					t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v, want invalid", err)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("RunEventTransaction() error = %v", err)
			}
		})
	}
}

func TestEventTransactionOccurrenceRSVP_ResolvesCancellationAndBoundedPrecedence(t *testing.T) {
	request := calendar.CalendarEventOccurrenceRSVPAuthorizationRequest{EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04"}
	resolve := func(t *testing.T, backend *occurrenceRSVPRepositoryBackend) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) {
		t.Helper()
		var authorization calendar.CalendarEventOccurrenceRSVPAuthorization
		err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
			scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
			if err != nil {
				return err
			}
			authorization, err = transaction.ResolveCalendarEventOccurrenceRSVP(context.Background(), scope, request)
			return err
		})
		return authorization, err
	}

	t.Run("cancelled master returns fixed sanitized authorization before variable validation", func(t *testing.T) {
		backend := newOccurrenceRSVPRepositoryBackend()
		hostile := strings.Repeat("x", 255)
		hostileRecurrence := strings.Repeat("x", 256)
		backend.authorization.Current.Status = calendar.CalendarEventStatusCancelled
		backend.authorization.Current.Title = hostile
		backend.authorization.Current.Description = hostile
		backend.authorization.Current.Temporal = calendar.CalendarEvent{}
		backend.authorization.Current.Exdates = make([]string, 1001)
		for index := range backend.authorization.Current.Exdates {
			backend.authorization.Current.Exdates[index] = hostile
		}
		backend.authorization.SeriesAttendee.Email = hostile
		backend.authorization.SeriesAttendee.DisplayName = hostile
		backend.authorization.OccurrenceAttendee.Email = hostile
		backend.authorization.OccurrenceAttendee.DisplayName = hostile
		backend.authorization.OccurrenceAttendee.RecurrenceID = hostileRecurrence
		backend.authorization.Exception.RecurrenceID = hostileRecurrence
		backend.authorization.Exception.Temporal = &calendar.CalendarEvent{}

		authorization, err := resolve(t, backend)
		if err != nil {
			t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v", err)
		}
		expected := calendar.CalendarEventOccurrenceRSVPAuthorization{
			Current: calendar.CalendarEventCommandInput{
				EventID: rsvpRepositoryEventID, OrganizationID: rsvpRepositoryOrganizationID, Status: calendar.CalendarEventStatusCancelled,
			},
			SeriesAttendee: calendar.CalendarEventSeriesRSVPAttendeeSnapshot{
				AttendeeID: rsvpRepositoryAttendeeID, EventID: rsvpRepositoryEventID, UserID: rsvpRepositoryUUIDRef(rsvpRepositoryActorID),
			},
		}
		if !reflect.DeepEqual(authorization, expected) {
			t.Fatalf("cancelled master authorization = %#v, want %#v", authorization, expected)
		}
		*authorization.SeriesAttendee.UserID = rsvpRepositoryOtherID
		if backend.authorization.SeriesAttendee.UserID == nil || *backend.authorization.SeriesAttendee.UserID != rsvpRepositoryActorID {
			t.Fatalf("sanitized authorization retained backend attendee alias: %#v", backend.authorization.SeriesAttendee)
		}
	})

	t.Run("live master rejects oversized exdates before clone", func(t *testing.T) {
		backend := newOccurrenceRSVPRepositoryBackend()
		backend.authorization.Current.Exdates = make([]string, 1001)
		for index := range backend.authorization.Current.Exdates {
			backend.authorization.Current.Exdates[index] = "2032-02-04T04:05:06Z"
		}
		authorization, err := resolve(t, backend)
		if !errors.Is(err, calendar.ErrEventInvalid) {
			t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v, want invalid", err)
		}
		if !reflect.DeepEqual(authorization, calendar.CalendarEventOccurrenceRSVPAuthorization{}) {
			t.Fatalf("oversized exdates authorization = %#v, want zero", authorization)
		}
	})

	t.Run("live master accepts maximum exdate bound", func(t *testing.T) {
		backend := newOccurrenceRSVPRepositoryBackend()
		backend.authorization.Current.Exdates = make([]string, 1000)
		for index := range backend.authorization.Current.Exdates {
			backend.authorization.Current.Exdates[index] = "2032-02-04T04:05:06Z"
		}
		authorization, err := resolve(t, backend)
		if err != nil {
			t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v", err)
		}
		if len(authorization.Current.Exdates) != 1000 {
			t.Fatalf("maximum exdates = %d, want 1000", len(authorization.Current.Exdates))
		}
	})

	t.Run("cancelled exception sanitizes hostile temporal", func(t *testing.T) {
		backend := newOccurrenceRSVPRepositoryBackend()
		backend.authorization.Exception.Cancelled = true
		backend.authorization.Exception.Status = calendar.CalendarEventStatusCancelled
		backend.authorization.Exception.Temporal = &calendar.CalendarEvent{}

		authorization, err := resolve(t, backend)
		if err != nil {
			t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v", err)
		}
		if authorization.Exception == nil {
			t.Fatal("cancelled exception was not returned")
		}
		expected := calendar.CalendarEventExceptionCommandInput{
			ExceptionID: rsvpRepositoryEnvelopeID, EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04", Sequence: 6,
			Status: calendar.CalendarEventStatusCancelled, Cancelled: true,
		}
		if !reflect.DeepEqual(*authorization.Exception, expected) {
			t.Fatalf("cancelled exception = %#v, want %#v", *authorization.Exception, expected)
		}
		if backend.authorization.Exception.Temporal == nil {
			t.Fatal("cancelled exception unexpectedly changed backend temporal")
		}
	})

	for _, testCase := range []struct {
		name      string
		status    calendar.CalendarEventStatus
		temporal  *calendar.CalendarEvent
		wantNil   bool
	}{
		{name: "confirmed nil temporal", status: calendar.CalendarEventStatusConfirmed, wantNil: true},
		{name: "tentative nil temporal", status: calendar.CalendarEventStatusTentative, wantNil: true},
		{name: "tentative valid temporal", status: calendar.CalendarEventStatusTentative, temporal: &calendar.CalendarEvent{StartsAt: time.Date(2032, time.February, 4, 4, 5, 6, 0, time.UTC), EndsAt: time.Date(2032, time.February, 4, 5, 5, 6, 0, time.UTC), Timezone: "UTC"}},
	} {
		t.Run("live exception "+testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			backend.authorization.Exception.Cancelled = false
			backend.authorization.Exception.Status = testCase.status
			backend.authorization.Exception.Temporal = testCase.temporal
			authorization, err := resolve(t, backend)
			if err != nil {
				t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v", err)
			}
			if authorization.Exception == nil || authorization.Exception.Status != testCase.status || authorization.Exception.Cancelled || (testCase.wantNil && authorization.Exception.Temporal != nil) || (!testCase.wantNil && (authorization.Exception.Temporal == nil || authorization.Exception.Temporal.Validate() != nil)) {
				t.Fatalf("live exception = %#v", authorization.Exception)
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.CalendarEventExceptionCommandInput)
	}{
		{name: "zero exception id", mutate: func(exception *calendar.CalendarEventExceptionCommandInput) { exception.ExceptionID = uuid.Nil }},
		{name: "future exception sequence", mutate: func(exception *calendar.CalendarEventExceptionCommandInput) { exception.Sequence = 7 }},
		{name: "live cancelled status", mutate: func(exception *calendar.CalendarEventExceptionCommandInput) { exception.Status = calendar.CalendarEventStatusCancelled }},
		{name: "cancelled confirmed status", mutate: func(exception *calendar.CalendarEventExceptionCommandInput) { exception.Cancelled = true }},
		{name: "malformed live temporal", mutate: func(exception *calendar.CalendarEventExceptionCommandInput) { exception.Temporal = &calendar.CalendarEvent{} }},
	} {
		t.Run("malformed exception "+testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			testCase.mutate(backend.authorization.Exception)
			authorization, err := resolve(t, backend)
			if !errors.Is(err, calendar.ErrEventInvalid) {
				t.Fatalf("ResolveCalendarEventOccurrenceRSVP() error = %v, want invalid", err)
			}
			if !reflect.DeepEqual(authorization, calendar.CalendarEventOccurrenceRSVPAuthorization{}) {
				t.Fatalf("malformed exception authorization = %#v, want zero", authorization)
			}
		})
	}
}

func TestEventTransactionOccurrenceRSVP_PersistsOnlyExactActorBoundEcho(t *testing.T) {
	series := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	attendee := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	command := occurrenceRSVPRepositoryCommand(t, series, nil, attendee, 2, 0, true)
	backend := newOccurrenceRSVPRepositoryBackend()
	err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		result, err := transaction.PersistCalendarEventOccurrenceRSVP(context.Background(), scope, command)
		if err != nil {
			return err
		}
		if result.Attendee().AttendeeID != rsvpRepositoryOccurrenceID || result.EventSequence() != 0 || backend.persistCalls != 1 || len(backend.persisted) != 1 {
			t.Fatalf("result:%#v persist calls:%d commands:%d", result, backend.persistCalls, len(backend.persisted))
		}
		persisted := backend.persisted[0]
		if persisted.SeriesAttendee().AttendeeID != series.AttendeeID || persisted.SeriesAttendee().Version != 2 || prior, found := persisted.Prior(); found || prior.AttendeeID != uuid.Nil {
			t.Fatalf("persistence changed series or invented prior: %#v", persisted)
		}
		reply, found := persisted.Reply()
		if !found || reply.AttendeeID != rsvpRepositoryOccurrenceID || reply.EventSequence != 0 || reply.AttendeeVersion != 3 || reply.RecurrenceID != "2032-02-04" || reply.Status != calendar.RSVPAccepted {
			t.Fatalf("persisted reply = %#v found:%t", reply, found)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction() error = %v", err)
	}

	noReplyCommand := occurrenceRSVPRepositoryCommand(t, series, nil, attendee, 2, 0, false)
	if noReply, found := noReplyCommand.Reply(); found || noReply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}) {
		t.Fatalf("command no-reply = %#v found:%t", noReply, found)
	}
	backend = newOccurrenceRSVPRepositoryBackend()
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		result, err := transaction.PersistCalendarEventOccurrenceRSVP(context.Background(), scope, noReplyCommand)
		if err != nil {
			return err
		}
		if noReply, found := result.Reply(); found || noReply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}) {
			t.Fatalf("result no-reply = %#v found:%t", noReply, found)
		}
		if len(backend.persisted) != 1 {
			t.Fatalf("no-reply persisted commands = %d, want 1", len(backend.persisted))
		}
		if noReply, found := backend.persisted[0].Reply(); found || noReply != (calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{}) {
			t.Fatalf("persistence no-reply echo = %#v found:%t", noReply, found)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("RunEventTransaction(no reply) error = %v", err)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*occurrenceRSVPRepositoryBackend)
	}{
		{name: "zero backend result", mutate: func(backend *occurrenceRSVPRepositoryBackend) { backend.returnZeroResult = true }},
		{name: "spoofed result echo", mutate: func(backend *occurrenceRSVPRepositoryBackend) {
			backend.resultOverride = func(command calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) {
				spoofed := command.Attendee()
				spoofed.Status = calendar.RSVPDeclined
				return calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(spoofed, command.EventSequence(), command.Outbox(), nil)
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			backend := newOccurrenceRSVPRepositoryBackend()
			testCase.mutate(backend)
			err := calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
				scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
				if err != nil {
					return err
				}
				_, err = transaction.PersistCalendarEventOccurrenceRSVP(context.Background(), scope, command)
				if !errors.Is(err, calendar.ErrEventInvalid) {
					t.Fatalf("PersistCalendarEventOccurrenceRSVP() error = %v, want invalid", err)
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

	foreignSeries := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	foreignSeries.UserID = rsvpRepositoryUUIDRef(rsvpRepositoryOtherID)
	foreignAttendee := occurrenceRSVPRepositoryAttendee(foreignSeries, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	foreignCommand := occurrenceRSVPRepositoryCommand(t, foreignSeries, nil, foreignAttendee, 2, 0, false)
	backend = newOccurrenceRSVPRepositoryBackend()
	err = calendar.RunEventTransaction(backend, func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveCalendarEventOccurrenceRSVPScope(context.Background(), rsvpRepositoryActorID)
		if err != nil {
			return err
		}
		_, err = transaction.PersistCalendarEventOccurrenceRSVP(context.Background(), scope, foreignCommand)
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

func occurrenceRSVPRepositorySeries(status calendar.RSVPStatus) calendar.CalendarEventSeriesRSVPAttendeeSnapshot {
	series := rsvpRepositoryAttendee(status)
	series.Version = 2
	return series
}

func occurrenceRSVPRepositoryAttendee(series calendar.CalendarEventSeriesRSVPAttendeeSnapshot, attendeeID uuid.UUID, recurrenceID string, status calendar.RSVPStatus, version uint) calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot {
	attendee := calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot{
		AttendeeID: attendeeID, EventID: series.EventID, UserID: rsvpRepositoryUUIDRef(*series.UserID), Email: series.Email, DisplayName: series.DisplayName,
		Role: series.Role, Status: status, RecurrenceID: recurrenceID, Version: version,
	}
	if status != calendar.RSVPNeedsAction {
		respondedAt := time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)
		attendee.RespondedAt = &respondedAt
	}
	return attendee
}

func occurrenceRSVPRepositoryAttendeeFrom(prior calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, status calendar.RSVPStatus, version uint, mutate func(*calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot)) calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot {
	updated := prior
	updated.UserID = rsvpRepositoryUUIDRef(*prior.UserID)
	updated.Status = status
	updated.Version = version
	if status == calendar.RSVPNeedsAction {
		updated.RespondedAt = nil
	} else {
		respondedAt := time.Date(2032, time.February, 3, 4, 5, 6, 0, time.UTC)
		updated.RespondedAt = &respondedAt
	}
	mutate(&updated)
	return updated
}

func occurrenceRSVPRepositoryReply(attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, sequence uint) calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor {
	return calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor{AttendeeID: attendee.AttendeeID, EventID: attendee.EventID, EventSequence: sequence, AttendeeVersion: attendee.Version, RecurrenceID: attendee.RecurrenceID, Status: attendee.Status}
}

func occurrenceRSVPRepositoryMustReply(t *testing.T, reply calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor, found bool) calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor {
	t.Helper()
	if !found {
		t.Fatal("Reply() found = false, want true")
	}
	return reply
}

func occurrenceRSVPRepositoryCommand(t *testing.T, series calendar.CalendarEventSeriesRSVPAttendeeSnapshot, prior *calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, attendee calendar.CalendarEventOccurrenceRSVPAttendeeSnapshot, expectedVersion, sequence uint, withReply bool) calendar.CalendarEventOccurrenceRSVPPersistenceCommand {
	t.Helper()
	var reply *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor
	if withReply {
		value := occurrenceRSVPRepositoryReply(attendee, sequence)
		reply = &value
	}
	command, err := calendar.NewCalendarEventOccurrenceRSVPPersistenceCommand(series, prior, attendee, expectedVersion, sequence, rsvpRepositoryOutbox(t), reply)
	if err != nil {
		t.Fatalf("NewCalendarEventOccurrenceRSVPPersistenceCommand() error = %v", err)
	}
	return command
}

type occurrenceRSVPRepositoryBackend struct {
	rsvpBaseOnlyBackend
	authorization    calendar.CalendarEventOccurrenceRSVPAuthorization
	lastRequest      calendar.CalendarEventOccurrenceRSVPAuthorizationRequest
	persisted        []calendar.CalendarEventOccurrenceRSVPPersistenceCommand
	resolveCalls     int
	persistCalls     int
	returnZeroResult bool
	resultOverride   func(calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error)
}

func newOccurrenceRSVPRepositoryBackend() *occurrenceRSVPRepositoryBackend {
	series := occurrenceRSVPRepositorySeries(calendar.RSVPAccepted)
	occurrence := occurrenceRSVPRepositoryAttendee(series, rsvpRepositoryOccurrenceID, "2032-02-04", calendar.RSVPAccepted, 3)
	temporal := rsvpRepositoryCurrent().Temporal
	return &occurrenceRSVPRepositoryBackend{
		rsvpBaseOnlyBackend: rsvpBaseOnlyBackend{organizationID: rsvpRepositoryOrganizationID},
		authorization: calendar.CalendarEventOccurrenceRSVPAuthorization{
			Current: rsvpRepositoryCurrent(), SeriesAttendee: series, OccurrenceAttendee: &occurrence,
			Exception: &calendar.CalendarEventExceptionCommandInput{ExceptionID: rsvpRepositoryEnvelopeID, EventID: rsvpRepositoryEventID, RecurrenceID: "2032-02-04", Temporal: &temporal, Sequence: 6, Status: calendar.CalendarEventStatusConfirmed},
		},
	}
}

func (backend *occurrenceRSVPRepositoryBackend) ResolveCalendarEventOccurrenceRSVP(_ context.Context, _ calendar.EventMutationScope, request calendar.CalendarEventOccurrenceRSVPAuthorizationRequest) (calendar.CalendarEventOccurrenceRSVPAuthorization, error) {
	backend.resolveCalls++
	backend.lastRequest = request
	return backend.authorization, nil
}

func (backend *occurrenceRSVPRepositoryBackend) PersistCalendarEventOccurrenceRSVP(_ context.Context, _ calendar.EventMutationScope, command calendar.CalendarEventOccurrenceRSVPPersistenceCommand) (calendar.CalendarEventOccurrenceRSVPPersistenceResult, error) {
	backend.persistCalls++
	if backend.returnZeroResult {
		return calendar.CalendarEventOccurrenceRSVPPersistenceResult{}, nil
	}
	if backend.resultOverride != nil {
		return backend.resultOverride(command)
	}
	backend.persisted = append(backend.persisted, command)
	return calendar.NewCalendarEventOccurrenceRSVPPersistenceResult(command.Attendee(), command.EventSequence(), command.Outbox(), occurrenceRSVPRepositoryReplyPointer(command))
}

func occurrenceRSVPRepositoryReplyPointer(command calendar.CalendarEventOccurrenceRSVPPersistenceCommand) *calendar.CalendarEventOccurrenceRSVPReplyDeliveryDescriptor {
	reply, found := command.Reply()
	if !found {
		return nil
	}
	return &reply
}
