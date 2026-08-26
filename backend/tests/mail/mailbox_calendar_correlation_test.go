package mail_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	mail "github.com/valerii/hub/backend/internal/mail"
)

type calendarCorrelationFixture struct {
	actorID         uuid.UUID
	otherActorID    uuid.UUID
	organizationID  uuid.UUID
	accountID       uuid.UUID
	messageID       uuid.UUID
	calendarEventID uuid.UUID
}

func calendarCorrelationFixtureValues() calendarCorrelationFixture {
	return calendarCorrelationFixture{
		actorID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		otherActorID:    uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		organizationID:  uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		accountID:       uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		messageID:       uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		calendarEventID: uuid.MustParse("66666666-6666-4666-8666-666666666666"),
	}
}

func mailboxCalendarCorrelationRequest(t *testing.T, fixture calendarCorrelationFixture) mail.MailboxCalendarCorrelationRequest {
	t.Helper()
	request, err := mail.NewMailboxCalendarCorrelationRequest(
		fixture.actorID,
		fixture.organizationID,
		fixture.accountID,
		fixture.messageID,
	)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationRequest() error = %v", err)
	}
	return request
}

func calendarMailboxCorrelationRequest(t *testing.T, fixture calendarCorrelationFixture) mail.CalendarMailboxCorrelationRequest {
	t.Helper()
	request, err := mail.NewCalendarMailboxCorrelationRequest(
		fixture.actorID,
		fixture.organizationID,
		fixture.calendarEventID,
	)
	if err != nil {
		t.Fatalf("NewCalendarMailboxCorrelationRequest() error = %v", err)
	}
	return request
}

func mailboxCalendarCorrelationResult(t *testing.T, request mail.MailboxCalendarCorrelationRequest, calendarEventID uuid.UUID, path mail.MailboxCalendarCorrelationPath) mail.MailboxCalendarCorrelationResult {
	t.Helper()
	result, err := mail.NewMailboxCalendarCorrelationResult(request, calendarEventID, path)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationResult() error = %v", err)
	}
	return result
}

func calendarMailboxCorrelationResult(t *testing.T, request mail.CalendarMailboxCorrelationRequest, accountID, messageID uuid.UUID, path mail.MailboxCalendarCorrelationPath) mail.CalendarMailboxCorrelationResult {
	t.Helper()
	result, err := mail.NewCalendarMailboxCorrelationResult(request, accountID, messageID, path)
	if err != nil {
		t.Fatalf("NewCalendarMailboxCorrelationResult() error = %v", err)
	}
	return result
}

func requireMailboxCalendarCorrelationErrorKind(t *testing.T, err error, want mail.MailboxCalendarCorrelationErrorKind) {
	t.Helper()
	got, ok := mail.MailboxCalendarCorrelationErrorKindOf(err)
	if !ok || got != want {
		t.Fatalf("MailboxCalendarCorrelationErrorKindOf(%v) = (%q, %t), want (%q, true)", err, got, ok, want)
	}
}

func TestMailboxCalendarCorrelationRequestConstructors(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()

	request := mailboxCalendarCorrelationRequest(t, fixture)
	if request.ActorID() != fixture.actorID || request.OrganizationID() != fixture.organizationID ||
		request.AccountID() != fixture.accountID || request.MessageID() != fixture.messageID {
		t.Fatalf("NewMailboxCalendarCorrelationRequest() = %#v, want exact scoped IDs", request)
	}

	for _, test := range []struct {
		name    string
		actorID uuid.UUID
		orgID   uuid.UUID
		account uuid.UUID
		message uuid.UUID
	}{
		{name: "nil actor", actorID: uuid.Nil, orgID: fixture.organizationID, account: fixture.accountID, message: fixture.messageID},
		{name: "nil organization", actorID: fixture.actorID, orgID: uuid.Nil, account: fixture.accountID, message: fixture.messageID},
		{name: "nil account", actorID: fixture.actorID, orgID: fixture.organizationID, account: uuid.Nil, message: fixture.messageID},
		{name: "nil message", actorID: fixture.actorID, orgID: fixture.organizationID, account: fixture.accountID, message: uuid.Nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := mail.NewMailboxCalendarCorrelationRequest(test.actorID, test.orgID, test.account, test.message)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}

	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)
	if reverseRequest.ActorID() != fixture.actorID || reverseRequest.OrganizationID() != fixture.organizationID ||
		reverseRequest.CalendarEventID() != fixture.calendarEventID {
		t.Fatalf("NewCalendarMailboxCorrelationRequest() = %#v, want exact scoped IDs", reverseRequest)
	}

	for _, test := range []struct {
		name    string
		actorID uuid.UUID
		orgID   uuid.UUID
		eventID uuid.UUID
	}{
		{name: "nil actor", actorID: uuid.Nil, orgID: fixture.organizationID, eventID: fixture.calendarEventID},
		{name: "nil organization", actorID: fixture.actorID, orgID: uuid.Nil, eventID: fixture.calendarEventID},
		{name: "nil event", actorID: fixture.actorID, orgID: fixture.organizationID, eventID: uuid.Nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := mail.NewCalendarMailboxCorrelationRequest(test.actorID, test.orgID, test.eventID)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}
}

func TestMailboxCalendarCorrelationResultConstructors(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	request := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)
	otherFixture := fixture
	otherFixture.actorID = fixture.otherActorID
	otherRequest := mailboxCalendarCorrelationRequest(t, otherFixture)
	otherReverseRequest := calendarMailboxCorrelationRequest(t, otherFixture)

	for _, path := range []mail.MailboxCalendarCorrelationPath{
		mail.MailboxCalendarCorrelationPathInboundReceipt,
		mail.MailboxCalendarCorrelationPathOutboundInvite,
	} {
		t.Run(string(path), func(t *testing.T) {
			result := mailboxCalendarCorrelationResult(t, request, fixture.calendarEventID, path)
			if result.ActorID() != fixture.actorID || result.OrganizationID() != fixture.organizationID ||
				result.AccountID() != fixture.accountID || result.MessageID() != fixture.messageID ||
				result.CalendarEventID() != fixture.calendarEventID || result.Path() != path {
				t.Fatalf("NewMailboxCalendarCorrelationResult() = %#v, want exact scoped values", result)
			}
			if !result.Matches(request) || result.Matches(otherRequest) {
				t.Fatal("MailboxCalendarCorrelationResult.Matches() did not require the originating exact request")
			}

			reverseResult := calendarMailboxCorrelationResult(t, reverseRequest, fixture.accountID, fixture.messageID, path)
			if reverseResult.ActorID() != fixture.actorID || reverseResult.OrganizationID() != fixture.organizationID ||
				reverseResult.CalendarEventID() != fixture.calendarEventID || reverseResult.AccountID() != fixture.accountID ||
				reverseResult.MessageID() != fixture.messageID || reverseResult.Path() != path {
				t.Fatalf("NewCalendarMailboxCorrelationResult() = %#v, want exact scoped values", reverseResult)
			}
			if !reverseResult.Matches(reverseRequest) || reverseResult.Matches(otherReverseRequest) {
				t.Fatal("CalendarMailboxCorrelationResult.Matches() did not require the originating exact request")
			}
		})
	}

	for _, test := range []struct {
		name    string
		eventID uuid.UUID
		path    mail.MailboxCalendarCorrelationPath
	}{
		{name: "nil linked event", eventID: uuid.Nil, path: mail.MailboxCalendarCorrelationPathInboundReceipt},
		{name: "invalid path", eventID: fixture.calendarEventID, path: mail.MailboxCalendarCorrelationPath("invalid")},
	} {
		t.Run("forward "+test.name, func(t *testing.T) {
			_, err := mail.NewMailboxCalendarCorrelationResult(request, test.eventID, test.path)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}

	for _, test := range []struct {
		name      string
		accountID uuid.UUID
		messageID uuid.UUID
		path      mail.MailboxCalendarCorrelationPath
	}{
		{name: "nil linked account", accountID: uuid.Nil, messageID: fixture.messageID, path: mail.MailboxCalendarCorrelationPathInboundReceipt},
		{name: "nil linked message", accountID: fixture.accountID, messageID: uuid.Nil, path: mail.MailboxCalendarCorrelationPathInboundReceipt},
		{name: "invalid path", accountID: fixture.accountID, messageID: fixture.messageID, path: mail.MailboxCalendarCorrelationPath("invalid")},
	} {
		t.Run("reverse "+test.name, func(t *testing.T) {
			_, err := mail.NewCalendarMailboxCorrelationResult(reverseRequest, test.accountID, test.messageID, test.path)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}
}

type mailboxCalendarCorrelationResolverFake struct {
	result mail.MailboxCalendarCorrelationResult
	err    error
}

func (resolver *mailboxCalendarCorrelationResolverFake) ResolveCalendarEventForMessage(context.Context, mail.MailboxCalendarCorrelationRequest) (mail.MailboxCalendarCorrelationResult, error) {
	return resolver.result, resolver.err
}

type calendarMailboxCorrelationResolverFake struct {
	result mail.CalendarMailboxCorrelationResult
	err    error
}

func (resolver *calendarMailboxCorrelationResolverFake) ResolveMailMessageForCalendarEvent(context.Context, mail.CalendarMailboxCorrelationRequest) (mail.CalendarMailboxCorrelationResult, error) {
	return resolver.result, resolver.err
}

func TestMailboxCalendarCorrelationEntrypointsRejectNilResolvers(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	forwardRequest := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)
	var typedNilForward *mailboxCalendarCorrelationResolverFake
	var typedNilReverse *calendarMailboxCorrelationResolverFake

	for _, test := range []struct {
		name     string
		resolver mail.MailboxCalendarCorrelationResolver
	}{
		{name: "nil", resolver: nil},
		{name: "typed nil", resolver: typedNilForward},
	} {
		t.Run("forward "+test.name, func(t *testing.T) {
			_, err := mail.ResolveMailboxCalendarCorrelation(context.Background(), test.resolver, forwardRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}

	for _, test := range []struct {
		name     string
		resolver mail.CalendarMailboxCorrelationResolver
	}{
		{name: "nil", resolver: nil},
		{name: "typed nil", resolver: typedNilReverse},
	} {
		t.Run("reverse "+test.name, func(t *testing.T) {
			_, err := mail.ResolveCalendarMailboxCorrelation(context.Background(), test.resolver, reverseRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindInvalid)
		})
	}
}

func TestMailboxCalendarCorrelationEntrypointsPreserveCancelledContext(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	forwardRequest := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)
	forwardResult := mailboxCalendarCorrelationResult(t, forwardRequest, fixture.calendarEventID, mail.MailboxCalendarCorrelationPathInboundReceipt)
	reverseResult := calendarMailboxCorrelationResult(t, reverseRequest, fixture.accountID, fixture.messageID, mail.MailboxCalendarCorrelationPathOutboundInvite)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := mail.ResolveMailboxCalendarCorrelation(ctx, &mailboxCalendarCorrelationResolverFake{result: forwardResult}, forwardRequest); !errors.Is(err, context.Canceled) {
		t.Fatalf("ResolveMailboxCalendarCorrelation() error = %v, want context cancellation", err)
	}
	if _, err := mail.ResolveCalendarMailboxCorrelation(ctx, &calendarMailboxCorrelationResolverFake{result: reverseResult}, reverseRequest); !errors.Is(err, context.Canceled) {
		t.Fatalf("ResolveCalendarMailboxCorrelation() error = %v, want context cancellation", err)
	}
}

func TestMailboxCalendarCorrelationEntrypointsRejectInvalidAndMismatchedSuccess(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	forwardRequest := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)
	otherFixture := fixture
	otherFixture.actorID = fixture.otherActorID
	wrongForward := mailboxCalendarCorrelationResult(t, mailboxCalendarCorrelationRequest(t, otherFixture), fixture.calendarEventID, mail.MailboxCalendarCorrelationPathInboundReceipt)
	wrongReverse := calendarMailboxCorrelationResult(t, calendarMailboxCorrelationRequest(t, otherFixture), fixture.accountID, fixture.messageID, mail.MailboxCalendarCorrelationPathOutboundInvite)

	for _, test := range []struct {
		name   string
		result mail.MailboxCalendarCorrelationResult
	}{
		{name: "invalid"},
		{name: "mismatched", result: wrongForward},
	} {
		t.Run("forward "+test.name, func(t *testing.T) {
			_, err := mail.ResolveMailboxCalendarCorrelation(context.Background(), &mailboxCalendarCorrelationResolverFake{result: test.result}, forwardRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindUnavailable)
		})
	}

	for _, test := range []struct {
		name   string
		result mail.CalendarMailboxCorrelationResult
	}{
		{name: "invalid"},
		{name: "mismatched", result: wrongReverse},
	} {
		t.Run("reverse "+test.name, func(t *testing.T) {
			_, err := mail.ResolveCalendarMailboxCorrelation(context.Background(), &calendarMailboxCorrelationResolverFake{result: test.result}, reverseRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindUnavailable)
		})
	}
}

func TestMailboxCalendarCorrelationEntrypointsNormalizeErrors(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	forwardRequest := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)

	for _, test := range []struct {
		name string
		err  error
		want mail.MailboxCalendarCorrelationErrorKind
	}{
		{name: "unknown adapter error", err: errors.New("adapter failure"), want: mail.MailboxCalendarCorrelationErrorKindUnavailable},
		{name: "invalid", err: mail.NewMailboxCalendarCorrelationError(mail.MailboxCalendarCorrelationErrorKindInvalid), want: mail.MailboxCalendarCorrelationErrorKindInvalid},
		{name: "not found", err: mail.NewMailboxCalendarCorrelationError(mail.MailboxCalendarCorrelationErrorKindNotFound), want: mail.MailboxCalendarCorrelationErrorKindNotFound},
		{name: "ambiguous", err: mail.NewMailboxCalendarCorrelationError(mail.MailboxCalendarCorrelationErrorKindAmbiguous), want: mail.MailboxCalendarCorrelationErrorKindAmbiguous},
		{name: "stale", err: mail.NewMailboxCalendarCorrelationError(mail.MailboxCalendarCorrelationErrorKindStale), want: mail.MailboxCalendarCorrelationErrorKindStale},
		{name: "unavailable", err: mail.NewMailboxCalendarCorrelationError(mail.MailboxCalendarCorrelationErrorKindUnavailable), want: mail.MailboxCalendarCorrelationErrorKindUnavailable},
	} {
		t.Run("forward "+test.name, func(t *testing.T) {
			_, err := mail.ResolveMailboxCalendarCorrelation(context.Background(), &mailboxCalendarCorrelationResolverFake{err: test.err}, forwardRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, test.want)
		})
		t.Run("reverse "+test.name, func(t *testing.T) {
			_, err := mail.ResolveCalendarMailboxCorrelation(context.Background(), &calendarMailboxCorrelationResolverFake{err: test.err}, reverseRequest)
			requireMailboxCalendarCorrelationErrorKind(t, err, test.want)
		})
	}
}

func TestMailboxCalendarCorrelationEntrypointsAcceptExactResults(t *testing.T) {
	fixture := calendarCorrelationFixtureValues()
	forwardRequest := mailboxCalendarCorrelationRequest(t, fixture)
	reverseRequest := calendarMailboxCorrelationRequest(t, fixture)

	for _, path := range []mail.MailboxCalendarCorrelationPath{
		mail.MailboxCalendarCorrelationPathInboundReceipt,
		mail.MailboxCalendarCorrelationPathOutboundInvite,
	} {
		t.Run(string(path), func(t *testing.T) {
			wantForward := mailboxCalendarCorrelationResult(t, forwardRequest, fixture.calendarEventID, path)
			gotForward, err := mail.ResolveMailboxCalendarCorrelation(context.Background(), &mailboxCalendarCorrelationResolverFake{result: wantForward}, forwardRequest)
			if err != nil {
				t.Fatalf("ResolveMailboxCalendarCorrelation() error = %v", err)
			}
			if !gotForward.Matches(forwardRequest) || gotForward.CalendarEventID() != fixture.calendarEventID || gotForward.Path() != path {
				t.Fatalf("ResolveMailboxCalendarCorrelation() = %#v, want exact matching result", gotForward)
			}

			wantReverse := calendarMailboxCorrelationResult(t, reverseRequest, fixture.accountID, fixture.messageID, path)
			gotReverse, err := mail.ResolveCalendarMailboxCorrelation(context.Background(), &calendarMailboxCorrelationResolverFake{result: wantReverse}, reverseRequest)
			if err != nil {
				t.Fatalf("ResolveCalendarMailboxCorrelation() error = %v", err)
			}
			if !gotReverse.Matches(reverseRequest) || gotReverse.AccountID() != fixture.accountID ||
				gotReverse.MessageID() != fixture.messageID || gotReverse.Path() != path {
				t.Fatalf("ResolveCalendarMailboxCorrelation() = %#v, want exact matching result", gotReverse)
			}
		})
	}
}
