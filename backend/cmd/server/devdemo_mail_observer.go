//go:build !qa && !release

package main

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

type devMailDemoOwner struct {
	userID    uuid.UUID
	accountID uuid.UUID
}

func (adapter *devMailDemoAdapter) observeOnce(ctx context.Context, prepared devMailDemoPrepared, pmEventID uuid.UUID) error {
	if ctx == nil {
		return errDevMailDemoUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if pmEventID == uuid.Nil || !validDevMailControlResult(prepared.control) || !validDevMailSourceResult(prepared.source) {
		return errDevMailDemoUnavailable
	}
	plan, err := newDevMailSourcePlan(prepared.control)
	if err != nil || !validDevMailSourcePlan(plan) || !devMailDemoSourceResultEqual(plan.result, prepared.source) {
		return errDevMailDemoUnavailable
	}
	pmOwner, err := adapter.observeOwner(ctx, prepared, devMailSourcePMEmail, prepared.pmUserID)
	if err != nil {
		return err
	}
	dev1Owner, err := adapter.observeOwner(ctx, prepared, devMailSourceDev1Email, prepared.dev1UserID)
	if err != nil {
		return err
	}
	if pmOwner.accountID == dev1Owner.accountID {
		return errDevMailDemoUnavailable
	}
	if _, err := adapter.observeFolders(ctx, pmOwner.userID); err != nil {
		return err
	}
	dev1Folders, err := adapter.observeFolders(ctx, dev1Owner.userID)
	if err != nil {
		return err
	}
	if err := adapter.observeStaticSource(ctx, prepared, plan, dev1Owner, dev1Folders); err != nil {
		return err
	}
	pmEvent, err := adapter.readPMEvent(ctx, prepared, pmEventID)
	if err != nil {
		return err
	}
	pmMessageID, err := adapter.observePMInvitation(ctx, prepared, pmOwner, pmEventID)
	if err != nil {
		return err
	}
	dev1EventID, dev1MessageID, err := adapter.observeDev1Invitation(ctx, prepared, dev1Owner, pmEvent)
	if err != nil {
		return err
	}
	if pmMessageID == dev1MessageID || pmEventID == dev1EventID {
		return errDevMailDemoUnavailable
	}
	return adapter.observeInvitationSemantics(ctx, prepared, pmEvent, dev1EventID)
}

func (adapter *devMailDemoAdapter) observeOwner(ctx context.Context, prepared devMailDemoPrepared, email string, userID uuid.UUID) (devMailDemoOwner, error) {
	owner, err := adapter.mailboxSvc.ResolveOwner(ctx, userID)
	if err != nil {
		return devMailDemoOwner{}, devMailDemoMailboxError(err)
	}
	account, exists := prepared.control.accounts[email]
	if !exists || userID == uuid.Nil || owner.State() != mailmod.MailboxOwnerActive || owner.OrganizationID() != prepared.organizationID {
		return devMailDemoOwner{}, errDevMailDemoUnavailable
	}
	accountID, accountOK := owner.AccountID()
	address, addressOK := owner.PrimaryAddress()
	state, stateOK := owner.ProjectionState()
	if !accountOK || accountID == uuid.Nil || accountID != account.accountID || !addressOK || address != email {
		return devMailDemoOwner{}, errDevMailDemoUnavailable
	}
	if !stateOK || state != mailmod.MailboxProjectionReady {
		return devMailDemoOwner{}, errDevMailDemoPending
	}
	return devMailDemoOwner{userID: userID, accountID: accountID}, nil
}

// observeFolders returns the exact public durable role-to-folder UUID map.
func (adapter *devMailDemoAdapter) observeFolders(ctx context.Context, actorID uuid.UUID) (map[mailmod.EngineFolderRole]uuid.UUID, error) {
	projection, err := adapter.projectionSvc.ListMailboxFolders(ctx, actorID)
	if err != nil {
		return nil, devMailDemoMailboxError(err)
	}
	if projection.ProjectionState != mailmod.MailboxProjectionReady || projection.ProjectionAsOf == nil {
		return nil, errDevMailDemoPending
	}
	if len(projection.Items) != 6 {
		return nil, errDevMailDemoUnavailable
	}
	required := map[mailmod.EngineFolderRole]struct{}{
		mailmod.EngineFolderRoleInbox: {}, mailmod.EngineFolderRoleSent: {}, mailmod.EngineFolderRoleDrafts: {},
		mailmod.EngineFolderRoleArchive: {}, mailmod.EngineFolderRoleTrash: {}, mailmod.EngineFolderRoleJunk: {},
	}
	roles := make(map[mailmod.EngineFolderRole]uuid.UUID, len(required))
	ids := make(map[uuid.UUID]struct{}, len(required))
	for _, item := range projection.Items {
		if item.ID == uuid.Nil || item.Role == nil || item.State != "active" {
			return nil, errDevMailDemoUnavailable
		}
		role := mailmod.EngineFolderRole(*item.Role)
		if _, valid := required[role]; !valid {
			return nil, errDevMailDemoUnavailable
		}
		if _, duplicate := roles[role]; duplicate {
			return nil, errDevMailDemoUnavailable
		}
		if _, duplicate := ids[item.ID]; duplicate {
			return nil, errDevMailDemoUnavailable
		}
		roles[role], ids[item.ID] = item.ID, struct{}{}
	}
	if len(roles) != len(required) {
		return nil, errDevMailDemoUnavailable
	}
	return roles, nil
}

// observeStaticSource proves the immutable source via only owner-gated public
// MailboxFolderProjectionService reads. It never reads body, BCC, engine IDs,
// or Ent rows after rebuilding the source plan.
func (adapter *devMailDemoAdapter) observeStaticSource(ctx context.Context, prepared devMailDemoPrepared, plan devMailSourcePlan, owner devMailDemoOwner, durableFolders map[mailmod.EngineFolderRole]uuid.UUID) error {
	sourceAccount, sourceOK := plan.result.accounts[devMailSourceDev1Email]
	controlAccount, controlOK := prepared.control.accounts[devMailSourceDev1Email]
	if !sourceOK || !controlOK || sourceAccount.engineAccountID != controlAccount.engineAccountID || controlAccount.accountID != owner.accountID || len(durableFolders) != 6 {
		return errDevMailDemoUnavailable
	}
	detailCache := make(map[uuid.UUID]mailmod.MailboxThreadDetailProjection, plan.result.threadsTotal)
	messageIDs := make(map[uuid.UUID]struct{}, len(plan.messages))
	threadIDs := make(map[uuid.UUID]struct{}, plan.result.threadsTotal)
	attachments := 0
	for index := range plan.messages {
		expected := plan.messages[index]
		email, role, ok := devMailDemoSourceMessageFolderRole(plan, expected)
		if !ok || email != devMailSourceDev1Email || expected.Projection.AccountID != sourceAccount.engineAccountID {
			return errDevMailDemoUnavailable
		}
		folderID, ok := durableFolders[role]
		if !ok || folderID == uuid.Nil {
			return errDevMailDemoUnavailable
		}
		search, ok := devMailDemoProjectionSearch(expected.Projection.Subject)
		if !ok {
			return errDevMailDemoUnavailable
		}
		threads, err := adapter.observeTargetedThreads(ctx, owner.userID, search)
		if err != nil {
			return err
		}
		matches := make(map[uuid.UUID]uuid.UUID, 1)
		for _, thread := range threads.Items {
			detail, err := adapter.observeThreadDetail(ctx, owner.userID, thread.ID, detailCache)
			if err != nil {
				return err
			}
			for _, actual := range detail.Messages {
				if !devMailDemoSourceProjectionMessageMatches(actual, expected, folderID) {
					continue
				}
				if previous, exists := matches[actual.ID]; exists && previous != thread.ID {
					return errDevMailDemoUnavailable
				}
				matches[actual.ID] = thread.ID
			}
		}
		if len(matches) == 0 {
			return errDevMailDemoPending
		}
		if len(matches) != 1 {
			return errDevMailDemoUnavailable
		}
		for messageID, threadID := range matches {
			if messageID == uuid.Nil || threadID == uuid.Nil {
				return errDevMailDemoUnavailable
			}
			if _, duplicate := messageIDs[messageID]; duplicate {
				return errDevMailDemoUnavailable
			}
			messageIDs[messageID], threadIDs[threadID] = struct{}{}, struct{}{}
			if len(expected.Projection.Attachments) != 0 {
				attachments++
			}
		}
	}
	if len(messageIDs) != prepared.source.messagesTotal || len(threadIDs) != prepared.source.threadsTotal || attachments != prepared.source.attachmentsTotal || attachments != 1 {
		return errDevMailDemoUnavailable
	}
	return nil
}

func devMailDemoSourceMessageFolderRole(plan devMailSourcePlan, expected mailmod.EngineMessage) (string, mailmod.EngineFolderRole, bool) {
	if len(expected.Projection.FolderIDs) != 1 {
		return "", "", false
	}
	for email, account := range plan.result.accounts {
		if account.engineAccountID != expected.Projection.AccountID {
			continue
		}
		for role, id := range account.folderIDs {
			if id != expected.Projection.FolderIDs[0] {
				continue
			}
			for _, folder := range plan.folders {
				if folder.AccountID == account.engineAccountID && folder.FolderID == id && folder.Role == role {
					return email, role, true
				}
			}
		}
	}
	return "", "", false
}

func (adapter *devMailDemoAdapter) observeTargetedThreads(ctx context.Context, actorID uuid.UUID, search string) (mailmod.MailboxThreadProjection, error) {
	projection, err := adapter.projectionSvc.ListMailboxThreads(ctx, actorID, mailmod.MailboxThreadProjectionQuery{Limit: 100, Search: &search})
	if err != nil {
		return mailmod.MailboxThreadProjection{}, devMailDemoMailboxError(err)
	}
	if projection.ProjectionState != mailmod.MailboxProjectionReady || projection.ProjectionAsOf == nil {
		return mailmod.MailboxThreadProjection{}, errDevMailDemoPending
	}
	if len(projection.Items) > 100 || projection.NextCursor != nil {
		return mailmod.MailboxThreadProjection{}, errDevMailDemoUnavailable
	}
	seen := make(map[uuid.UUID]struct{}, len(projection.Items))
	for _, item := range projection.Items {
		if item.ID == uuid.Nil {
			return mailmod.MailboxThreadProjection{}, errDevMailDemoUnavailable
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return mailmod.MailboxThreadProjection{}, errDevMailDemoUnavailable
		}
		seen[item.ID] = struct{}{}
	}
	return projection, nil
}

func (adapter *devMailDemoAdapter) observeThreadDetail(ctx context.Context, actorID, threadID uuid.UUID, cache map[uuid.UUID]mailmod.MailboxThreadDetailProjection) (mailmod.MailboxThreadDetailProjection, error) {
	if detail, exists := cache[threadID]; exists {
		return detail, nil
	}
	detail, err := adapter.projectionSvc.GetMailboxThreadDetail(ctx, actorID, mailmod.MailboxThreadDetailProjectionQuery{ThreadID: threadID, Limit: 100})
	if err != nil {
		return mailmod.MailboxThreadDetailProjection{}, devMailDemoMailboxError(err)
	}
	if detail.ProjectionState != mailmod.MailboxProjectionReady || detail.ProjectionAsOf == nil {
		return mailmod.MailboxThreadDetailProjection{}, errDevMailDemoPending
	}
	if detail.Thread.ID != threadID || len(detail.Messages) > 100 || detail.NextCursor != nil {
		return mailmod.MailboxThreadDetailProjection{}, errDevMailDemoUnavailable
	}
	seen := make(map[uuid.UUID]struct{}, len(detail.Messages))
	for _, message := range detail.Messages {
		if message.ID == uuid.Nil || message.ThreadID != threadID {
			return mailmod.MailboxThreadDetailProjection{}, errDevMailDemoUnavailable
		}
		if _, duplicate := seen[message.ID]; duplicate {
			return mailmod.MailboxThreadDetailProjection{}, errDevMailDemoUnavailable
		}
		seen[message.ID] = struct{}{}
	}
	cache[threadID] = detail
	return detail, nil
}

func devMailDemoSourceProjectionMessageMatches(actual mailmod.MailboxThreadDetailProjectionMessage, expected mailmod.EngineMessage, durableFolderID uuid.UUID) bool {
	projection := expected.Projection
	if actual.ID == uuid.Nil || actual.ThreadID == uuid.Nil || projection.ReceivedAt == nil || actual.ReceivedAt == nil || actual.CalendarMethod != nil || actual.IsDraft ||
		actual.Subject != projection.Subject || actual.Snippet != projection.Snippet || actual.From.Email != projection.From.Email || !actual.ReceivedAt.Equal(*projection.ReceivedAt) ||
		actual.Priority != string(projection.Priority) || actual.SizeBytes != projection.SizeBytes || actual.Seen != projection.Flags.Seen || actual.Flagged != projection.Flags.Flagged ||
		actual.Answered != projection.Flags.Answered || actual.Junk != projection.Flags.Junk || actual.HasAttachments != (len(projection.Attachments) != 0) ||
		len(actual.FolderIDs) != 1 || actual.FolderIDs[0] != durableFolderID {
		return false
	}
	return devMailDemoProjectionAddressesMatch(actual.To, projection.To) && devMailDemoProjectionAddressesMatch(actual.CC, projection.CC)
}

func devMailDemoProjectionAddressesMatch(actual []mailmod.MailboxThreadDetailProjectionAddress, expected []mailmod.EngineAddress) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range expected {
		if actual[index].Email != expected[index].Email {
			return false
		}
	}
	return true
}

func devMailDemoProjectionSearch(subject string) (string, bool) {
	search, err := mailmod.NormalizeMailboxThreadProjectionSubjectSearch(subject)
	return search, err == nil
}

func (adapter *devMailDemoAdapter) readPMEvent(ctx context.Context, prepared devMailDemoPrepared, eventID uuid.UUID) (calendarmod.CalendarEventDetail, error) {
	event, err := adapter.eventReadSvc.GetCalendarEvent(ctx, prepared.pmUserID, eventID)
	if err != nil {
		return calendarmod.CalendarEventDetail{}, errDevMailDemoUnavailable
	}
	if event.ID != eventID || event.Title == "" {
		return calendarmod.CalendarEventDetail{}, errDevMailDemoUnavailable
	}
	return event, nil
}

func (adapter *devMailDemoAdapter) observePMInvitation(ctx context.Context, prepared devMailDemoPrepared, owner devMailDemoOwner, eventID uuid.UUID) (uuid.UUID, error) {
	reverseRequest, err := mailmod.NewCalendarMailboxCorrelationRequest(owner.userID, prepared.organizationID, eventID)
	if err != nil {
		return uuid.Nil, errDevMailDemoUnavailable
	}
	reverse, err := mailmod.ResolveCalendarMailboxCorrelation(ctx, adapter.mailboxRepo, reverseRequest)
	if err != nil {
		return uuid.Nil, devMailDemoCorrelationError(err)
	}
	if !reverse.Matches(reverseRequest) || reverse.AccountID() != owner.accountID || reverse.MessageID() == uuid.Nil || reverse.Path() != mailmod.MailboxCalendarCorrelationPathOutboundInvite {
		return uuid.Nil, errDevMailDemoUnavailable
	}
	forwardRequest, err := mailmod.NewMailboxCalendarCorrelationRequest(owner.userID, prepared.organizationID, owner.accountID, reverse.MessageID())
	if err != nil {
		return uuid.Nil, errDevMailDemoUnavailable
	}
	forward, err := mailmod.ResolveMailboxCalendarCorrelation(ctx, adapter.mailboxRepo, forwardRequest)
	if err != nil {
		return uuid.Nil, devMailDemoCorrelationError(err)
	}
	if !forward.Matches(forwardRequest) || forward.CalendarEventID() != eventID || forward.Path() != mailmod.MailboxCalendarCorrelationPathOutboundInvite {
		return uuid.Nil, errDevMailDemoUnavailable
	}
	return reverse.MessageID(), nil
}

func (adapter *devMailDemoAdapter) observeDev1Invitation(ctx context.Context, prepared devMailDemoPrepared, owner devMailDemoOwner, pmEvent calendarmod.CalendarEventDetail) (uuid.UUID, uuid.UUID, error) {
	search, ok := devMailDemoProjectionSearch(pmEvent.Title)
	if !ok {
		return uuid.Nil, uuid.Nil, errDevMailDemoUnavailable
	}
	threads, err := adapter.observeTargetedThreads(ctx, owner.userID, search)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	cache := make(map[uuid.UUID]mailmod.MailboxThreadDetailProjection, len(threads.Items))
	var eventID, messageID uuid.UUID
	for _, thread := range threads.Items {
		detail, err := adapter.observeThreadDetail(ctx, owner.userID, thread.ID, cache)
		if err != nil {
			return uuid.Nil, uuid.Nil, err
		}
		for _, message := range detail.Messages {
			if message.CalendarMethod == nil || *message.CalendarMethod != mailmod.MailCalendarMethodRequest {
				continue
			}
			request, err := mailmod.NewMailboxCalendarCorrelationRequest(owner.userID, prepared.organizationID, owner.accountID, message.ID)
			if err != nil {
				return uuid.Nil, uuid.Nil, errDevMailDemoUnavailable
			}
			forward, err := mailmod.ResolveMailboxCalendarCorrelation(ctx, adapter.mailboxRepo, request)
			if err != nil {
				return uuid.Nil, uuid.Nil, devMailDemoCorrelationError(err)
			}
			if !forward.Matches(request) || forward.Path() != mailmod.MailboxCalendarCorrelationPathInboundReceipt || forward.CalendarEventID() == uuid.Nil || eventID != uuid.Nil {
				return uuid.Nil, uuid.Nil, errDevMailDemoUnavailable
			}
			reverseRequest, err := mailmod.NewCalendarMailboxCorrelationRequest(owner.userID, prepared.organizationID, forward.CalendarEventID())
			if err != nil {
				return uuid.Nil, uuid.Nil, errDevMailDemoUnavailable
			}
			reverse, err := mailmod.ResolveCalendarMailboxCorrelation(ctx, adapter.mailboxRepo, reverseRequest)
			if err != nil {
				return uuid.Nil, uuid.Nil, devMailDemoCorrelationError(err)
			}
			if !reverse.Matches(reverseRequest) || reverse.Path() != mailmod.MailboxCalendarCorrelationPathInboundReceipt || reverse.AccountID() != owner.accountID || reverse.MessageID() != message.ID {
				return uuid.Nil, uuid.Nil, errDevMailDemoUnavailable
			}
			eventID, messageID = forward.CalendarEventID(), message.ID
		}
	}
	if eventID == uuid.Nil || messageID == uuid.Nil {
		return uuid.Nil, uuid.Nil, errDevMailDemoPending
	}
	return eventID, messageID, nil
}

func (adapter *devMailDemoAdapter) observeInvitationSemantics(ctx context.Context, prepared devMailDemoPrepared, pmEvent calendarmod.CalendarEventDetail, dev1EventID uuid.UUID) error {
	dev1Event, err := adapter.eventReadSvc.GetCalendarEvent(ctx, prepared.dev1UserID, dev1EventID)
	if err != nil {
		return errDevMailDemoUnavailable
	}
	if !devMailDemoInvitationEqual(pmEvent, dev1Event) || !dev1Event.CanRSVP || dev1Event.ViewerAttendee == nil || dev1Event.ViewerAttendee.UserID == nil || *dev1Event.ViewerAttendee.UserID != prepared.dev1UserID || dev1Event.ViewerAttendee.Role != calendarmod.CalendarEventAttendeeRoleRequired {
		return errDevMailDemoUnavailable
	}
	return nil
}

func devMailDemoInvitationEqual(pmEvent, dev1Event calendarmod.CalendarEventDetail) bool {
	if pmEvent.ID == uuid.Nil || dev1Event.ID == uuid.Nil || pmEvent.ID == dev1Event.ID || pmEvent.Title != dev1Event.Title || pmEvent.Description != dev1Event.Description || pmEvent.Location != dev1Event.Location || pmEvent.AllDay != dev1Event.AllDay || pmEvent.Timezone != dev1Event.Timezone || pmEvent.OrganizerEmail != dev1Event.OrganizerEmail || pmEvent.OrganizerEmail != devMailSourcePMEmail || pmEvent.AllDay || pmEvent.StartsAt == nil || pmEvent.EndsAt == nil || dev1Event.StartsAt == nil || dev1Event.EndsAt == nil {
		return false
	}
	return pmEvent.StartsAt.Location() == time.UTC && pmEvent.EndsAt.Location() == time.UTC && dev1Event.StartsAt.Location() == time.UTC && dev1Event.EndsAt.Location() == time.UTC && pmEvent.StartsAt.Equal(*dev1Event.StartsAt) && pmEvent.EndsAt.Equal(*dev1Event.EndsAt)
}

func devMailDemoMailboxError(err error) error {
	if errors.Is(err, mailmod.ErrMailboxServiceNotFound) || errors.Is(err, mailmod.ErrMailboxServiceUnavailable) {
		return errDevMailDemoPending
	}
	return errDevMailDemoUnavailable
}

func devMailDemoCorrelationError(err error) error {
	kind, known := mailmod.MailboxCalendarCorrelationErrorKindOf(err)
	if !known {
		return errDevMailDemoUnavailable
	}
	switch kind {
	case mailmod.MailboxCalendarCorrelationErrorKindNotFound,
		mailmod.MailboxCalendarCorrelationErrorKindStale,
		mailmod.MailboxCalendarCorrelationErrorKindUnavailable:
		return errDevMailDemoPending
	default:
		return errDevMailDemoUnavailable
	}
}
