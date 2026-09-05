//go:build !qa && !release

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"

	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	devMailDemoInvitationRRULE                     = "FREQ=WEEKLY;INTERVAL=1;BYDAY=TH;COUNT=4"
	devMailDemoInvitationOccurrenceMessageID       = "devmail-message-invitation-occurrence-v1"
	devMailDemoInvitationOccurrenceInternetMessage = "<devmail-invitation-occurrence-v1@hub.local>"
	devMailDemoInvitationOccurrenceSubject         = "Mail invitation occurrence"
	devMailDemoInvitationOccurrenceRecurrenceID    = "2026-09-10T10:00:00Z"
	devMailDemoInvitationOccurrenceBody            = "Mail invitation occurrence."
)

type devMailDemoInvitationOccurrenceFixture struct {
	message mailmod.EngineMessage
	part    mailmod.FakeCalendarPart
}

func seedDevMailDemoInvitationOccurrence(ctx context.Context, engine *mailmod.FakeMailEngine, prepared devMailDemoPrepared, eventID uuid.UUID) error {
	if !validDevMailControlContext(ctx) || engine == nil || eventID == uuid.Nil || !prepared.matches(prepared.organizationID, prepared.pmUserID, prepared.dev1UserID) || !validDevMailSourceResult(prepared.source) {
		return errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	fixture, err := newDevMailDemoInvitationOccurrenceFixture(prepared, eventID)
	if err != nil {
		return err
	}
	engine.PutCalendarPart(fixture.part)
	engine.PutMessage(fixture.message)
	if !readBackDevMailDemoInvitationOccurrence(ctx, engine, fixture) {
		return errDevMailDemoUnavailable
	}
	return nil
}

func newDevMailDemoInvitationOccurrenceFixture(prepared devMailDemoPrepared, eventID uuid.UUID) (devMailDemoInvitationOccurrenceFixture, error) {
	if eventID == uuid.Nil || !prepared.matches(prepared.organizationID, prepared.pmUserID, prepared.dev1UserID) || !validDevMailSourceResult(prepared.source) {
		return devMailDemoInvitationOccurrenceFixture{}, errDevMailDemoInvalid
	}
	account, accountOK := prepared.source.accounts[devMailSourceDev1Email]
	inboxID, inboxOK := account.folderIDs[mailmod.EngineFolderRoleInbox]
	if !accountOK || !inboxOK || !validDevMailControlOpaqueID(account.engineAccountID) || !validDevMailControlOpaqueID(inboxID) {
		return devMailDemoInvitationOccurrenceFixture{}, errDevMailDemoInvalid
	}
	itip := devMailDemoInvitationOccurrenceITIP(eventID)
	locatorHash := sha256.Sum256(itip)
	method := mailmod.MailCalendarMethodRequest
	receivedAt := time.Date(2026, time.September, 4, 10, 0, 0, 0, time.UTC)
	message := mailmod.EngineMessage{
		Projection: mailmod.EngineMessageProjection{
			AccountID:               account.engineAccountID,
			MessageID:               devMailDemoInvitationOccurrenceMessageID,
			Version:                 devMailDemoInvitationOccurrenceMessageID,
			InternetMessageID:       devMailDemoInvitationOccurrenceInternetMessage,
			InReplyTo:               "",
			References:              []string{},
			From:                    mailmod.EngineAddress{Email: devMailSourcePMEmail},
			To:                      []mailmod.EngineAddress{{Email: devMailSourceDev1Email}},
			CC:                      []mailmod.EngineAddress{},
			BCC:                     []mailmod.EngineAddress{},
			ReplyTo:                 []mailmod.EngineAddress{},
			Subject:                 devMailDemoInvitationOccurrenceSubject,
			Snippet:                 devMailDemoInvitationOccurrenceSubject,
			ReceivedAt:              &receivedAt,
			CalendarMethod:          &method,
			AuthenticationClass:     mailmod.MailAuthenticationClassTrustedInternal,
			CalendarPartLocatorHash: append([]byte(nil), locatorHash[:]...),
			SizeBytes:               int64(len(devMailDemoInvitationOccurrenceBody)),
			Priority:                mailmod.EnginePriorityNormal,
			FolderIDs:               []string{inboxID},
			Flags:                   mailmod.EngineMessageFlags{},
			ReservedKeywords:        []mailmod.EngineKeywordMarker{},
			Attachments:             []mailmod.EngineProjectionAttachment{},
			Origin:                  mailmod.EngineProjectionOriginInboundOrUnknown,
		},
		Body: mailmod.EngineMessageBody{Format: mailmod.EngineBodyPlain, PlainBody: devMailDemoInvitationOccurrenceBody},
	}
	part := mailmod.FakeCalendarPart{
		AccountID:   account.engineAccountID,
		MessageID:   devMailDemoInvitationOccurrenceMessageID,
		Version:     devMailDemoInvitationOccurrenceMessageID,
		LocatorHash: locatorHash,
		Method:      method,
		ITIP:        append([]byte(nil), itip...),
	}
	return devMailDemoInvitationOccurrenceFixture{message: message, part: part}, nil
}

func devMailDemoInvitationOccurrenceITIP(eventID uuid.UUID) []byte {
	return []byte("BEGIN:VCALENDAR\r\n" +
		"PRODID:-//Hub//Calendar//EN\r\n" +
		"VERSION:2.0\r\n" +
		"CALSCALE:GREGORIAN\r\n" +
		"METHOD:REQUEST\r\n" +
		"BEGIN:VEVENT\r\n" +
		"UID:hub-" + eventID.String() + "\r\n" +
		"DTSTAMP:20260904T100000Z\r\n" +
		"DTSTART;TZID=Europe/Moscow:20260910T130000\r\n" +
		"DTEND;TZID=Europe/Moscow:20260910T133000\r\n" +
		"RECURRENCE-ID;TZID=Europe/Moscow:20260910T130000\r\n" +
		"SEQUENCE:1\r\n" +
		"STATUS:CONFIRMED\r\n" +
		"SUMMARY:Mail invitation\r\n" +
		"DESCRIPTION:" + devMailDemoInvitationDescription + "\r\n" +
		"LOCATION:" + devMailDemoInvitationLocation + "\r\n" +
		"ORGANIZER:mailto:" + devMailSourcePMEmail + "\r\n" +
		"ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:" + devMailSourceDev1Email + "\r\n" +
		"END:VEVENT\r\n" +
		"END:VCALENDAR\r\n")
}

func readBackDevMailDemoInvitationOccurrence(ctx context.Context, engine *mailmod.FakeMailEngine, expected devMailDemoInvitationOccurrenceFixture) bool {
	if ctx == nil || engine == nil {
		return false
	}
	message, err := engine.GetMessage(ctx, mailmod.EngineMessageReadRequest{
		AccountID:       expected.message.Projection.AccountID,
		MessageID:       expected.message.Projection.MessageID,
		MaxContentBytes: mailmod.MaxEngineMessageBytes,
	})
	if err != nil || !reflect.DeepEqual(message, expected.message) {
		return false
	}
	part, err := engine.ReadCalendarPart(ctx, mailmod.EngineCalendarPartReadRequest{
		AccountID:   expected.part.AccountID,
		MessageID:   expected.part.MessageID,
		Version:     expected.part.Version,
		LocatorHash: expected.part.LocatorHash,
		Method:      expected.part.Method,
		MaxBytes:    mailmod.MaxCalendarPartBytes,
	})
	return err == nil && part.LocatorHash == expected.part.LocatorHash && bytes.Equal(part.ITIP, expected.part.ITIP)
}

func (adapter *devMailDemoAdapter) observeOccurrence(ctx context.Context, prepared devMailDemoPrepared, eventID uuid.UUID) error {
	if ctx == nil || eventID == uuid.Nil {
		return errDevMailDemoUnavailable
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, devMailDemoObserveTimeout)
	defer cancel()
	for {
		err := adapter.observeOccurrenceOnce(deadlineCtx, prepared, eventID)
		if err == nil {
			return nil
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if deadlineErr := deadlineCtx.Err(); deadlineErr != nil {
			return deadlineErr
		}
		if !errors.Is(err, errDevMailDemoPending) {
			return errDevMailDemoUnavailable
		}
		timer := time.NewTimer(devMailDemoObservePoll)
		select {
		case <-deadlineCtx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return deadlineCtx.Err()
		case <-timer.C:
		}
	}
}

func (adapter *devMailDemoAdapter) observeOccurrenceOnce(ctx context.Context, prepared devMailDemoPrepared, eventID uuid.UUID) error {
	if ctx == nil || eventID == uuid.Nil {
		return errDevMailDemoUnavailable
	}
	owner, err := adapter.observeOwner(ctx, prepared, devMailSourceDev1Email, prepared.dev1UserID)
	if err != nil {
		return err
	}
	search, ok := devMailDemoProjectionSearch(devMailDemoInvitationOccurrenceSubject)
	if !ok {
		return errDevMailDemoUnavailable
	}
	threads, err := adapter.observeTargetedThreads(ctx, owner.userID, search)
	if err != nil {
		return err
	}
	cache := make(map[uuid.UUID]mailmod.MailboxThreadDetailProjection, len(threads.Items))
	messageID := uuid.Nil
	for _, thread := range threads.Items {
		detail, err := adapter.observeThreadDetail(ctx, owner.userID, thread.ID, cache)
		if err != nil {
			return err
		}
		for _, message := range detail.Messages {
			if message.Subject != devMailDemoInvitationOccurrenceSubject {
				continue
			}
			if message.ID == uuid.Nil || message.CalendarMethod == nil || *message.CalendarMethod != mailmod.MailCalendarMethodRequest || messageID != uuid.Nil {
				return errDevMailDemoUnavailable
			}
			messageID = message.ID
		}
	}
	if messageID == uuid.Nil {
		return errDevMailDemoPending
	}
	detail, err := adapter.messageSvc.GetMailboxMessage(ctx, owner.userID, messageID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		var readError *mailmod.MailboxMessageReadError
		if errors.As(err, &readError) && readError != nil && (readError.Kind() == mailmod.MailboxMessageReadErrorNotFound || readError.Kind() == mailmod.MailboxMessageReadErrorUnavailable) {
			return errDevMailDemoPending
		}
		return errDevMailDemoUnavailable
	}
	if detail.ID != messageID || detail.ThreadID == uuid.Nil || detail.Subject != devMailDemoInvitationOccurrenceSubject || detail.CalendarMethod == nil || *detail.CalendarMethod != mailmod.MailCalendarMethodRequest || detail.CalendarEventID == nil || *detail.CalendarEventID != eventID || !devMailDemoOccurrenceRSVPTargetDTO(detail.CalendarRSVPTarget, eventID) || detail.PlainBody != devMailDemoInvitationOccurrenceBody || detail.SanitizedHTML != nil || !detail.RemoteImagesBlocked || detail.IsDraft || len(detail.Attachments) != 0 {
		return errDevMailDemoUnavailable
	}
	forwardRequest, err := mailmod.NewMailboxCalendarCorrelationRequest(owner.userID, prepared.organizationID, owner.accountID, messageID)
	if err != nil {
		return errDevMailDemoUnavailable
	}
	forward, err := mailmod.ResolveMailboxCalendarCorrelation(ctx, adapter.mailboxRepo, forwardRequest)
	if err != nil {
		return devMailDemoCorrelationError(err)
	}
	if !forward.Matches(forwardRequest) || forward.Path() != mailmod.MailboxCalendarCorrelationPathInboundReceipt || forward.CalendarEventID() != eventID || !devMailDemoOccurrenceRSVPTargetMatches(forward.CalendarRSVPTarget(), eventID) {
		return errDevMailDemoUnavailable
	}
	return nil
}

func devMailDemoSeriesRSVPTargetMatches(target *mailmod.MailboxCalendarRSVPTarget, eventID uuid.UUID) bool {
	return target != nil && target.CalendarEventID() == eventID && target.Scope() == mailmod.MailboxCalendarRSVPTargetScopeSeries && target.RecurrenceID() == nil && target.AttendeeVersion() > 0 && target.RSVPStatus() == mailmod.MailboxCalendarRSVPStatusNeedsAction
}

func devMailDemoSeriesRSVPTargetDTO(target *mailmod.MailboxMessageCalendarRSVPTarget, eventID uuid.UUID) bool {
	return target != nil && target.EventID == eventID && target.Scope == string(mailmod.MailboxCalendarRSVPTargetScopeSeries) && target.RecurrenceID == nil && target.AttendeeVersion > 0 && target.RSVPStatus == mailmod.MailboxCalendarRSVPStatusNeedsAction
}

func devMailDemoOccurrenceRSVPTargetMatches(target *mailmod.MailboxCalendarRSVPTarget, eventID uuid.UUID) bool {
	if target == nil || target.CalendarEventID() != eventID || target.Scope() != mailmod.MailboxCalendarRSVPTargetScopeOccurrence || target.AttendeeVersion() == 0 || target.RSVPStatus() != mailmod.MailboxCalendarRSVPStatusNeedsAction {
		return false
	}
	recurrenceID := target.RecurrenceID()
	return recurrenceID != nil && *recurrenceID == devMailDemoInvitationOccurrenceRecurrenceID
}

func devMailDemoOccurrenceRSVPTargetDTO(target *mailmod.MailboxMessageCalendarRSVPTarget, eventID uuid.UUID) bool {
	return target != nil && target.EventID == eventID && target.Scope == string(mailmod.MailboxCalendarRSVPTargetScopeOccurrence) && target.RecurrenceID != nil && *target.RecurrenceID == devMailDemoInvitationOccurrenceRecurrenceID && target.AttendeeVersion > 0 && target.RSVPStatus == mailmod.MailboxCalendarRSVPStatusNeedsAction
}
