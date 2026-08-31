//go:build qa && !release

package qamailcalendar

import (
	"context"
	"database/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendarinboundmessagereceipt"
	"github.com/valerii/hub/backend/ent/calendarinboundobjectstate"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/calendarreplydelivery"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/ent/user"
)

type inboundRequestFixturePublishedCommitUnknownObservation uint8

const (
	inboundRequestFixturePublishedCommitUnknownObservationInvalid inboundRequestFixturePublishedCommitUnknownObservation = iota
	inboundRequestFixturePublishedCommitUnknownObservationExactPresent
	inboundRequestFixturePublishedCommitUnknownObservationExactAbsent
	inboundRequestFixturePublishedCommitUnknownObservationPartial
)

func observeInboundRequestFixturePublishedCommitUnknown(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	expectedCleanup cleanupDomainProof,
) inboundRequestFixturePublishedCommitUnknownObservation {
	if ctx == nil ||
		tx == nil ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		!expectedCleanup.valid() ||
		!expectedCleanup.present {
		return inboundRequestFixturePublishedCommitUnknownObservationInvalid
	}

	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if !valid {
		return inboundRequestFixturePublishedCommitUnknownObservationInvalid
	}
	calendar, valid := newInboundRequestFixtureCalendarProof(
		expectedCleanup.inboundMessageReceiptID,
		expectedCleanup.inboundObjectStateID,
		expectedCleanup.eventID,
		expectedCleanup.attendeeID,
		expectedCleanup.replyDeliveryID,
		expectedCleanup.mutationReceiptID,
	)
	if !valid {
		return inboundRequestFixturePublishedCommitUnknownObservationInvalid
	}
	cleanup, valid := newCleanupDomainProof(calendar)
	if !valid || cleanup != expectedCleanup {
		return inboundRequestFixturePublishedCommitUnknownObservationInvalid
	}

	proof, _, observed := observeInboundRequestFixturePublishedFullProof(
		ctx,
		tx,
		plan,
		persisted,
		authority,
	)
	if observed {
		if proof.valid(plan, persisted) && proof.cleanup == expectedCleanup {
			return inboundRequestFixturePublishedCommitUnknownObservationExactPresent
		}
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	messageFolder, err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.IDEQ(roots.mailMessageFolderID),
			inboundRequestFixturePublishedMailClosureMessageFolderNoWait(),
		).
		Only(ctx)
	if err == nil && messageFolder != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	message, err := tx.MailMessage.Query().
		Where(
			mailmessage.IDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureMessageNoWait(),
		).
		Only(ctx)
	if err == nil && message != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	thread, err := tx.MailThread.Query().
		Where(
			mailthread.IDEQ(roots.mailThreadID),
			inboundRequestFixturePublishedMailThreadAccountClosureNoWait(),
		).
		Only(ctx)
	if err == nil && thread != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	folder, err := tx.MailFolder.Query().
		Where(
			mailfolder.IDEQ(roots.mailFolderID),
			inboundRequestFixturePublishedMailFolderClosureNoWait(),
		).
		Only(ctx)
	if err == nil && folder != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	identity, err := tx.MailIdentity.Query().
		Where(
			mailidentity.IDEQ(roots.mailIdentityID),
			inboundRequestFixturePublishedMailIdentityNoWait(),
		).
		Only(ctx)
	if err == nil && identity != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	account, err := tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailAccountNoWait(),
		).
		Only(ctx)
	if err == nil && account != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	domain, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(roots.mailDomainID),
			inboundRequestFixturePublishedMailDomainNoWait(),
		).
		Only(ctx)
	if err == nil && domain != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	owner, err := tx.User.Query().
		Where(
			user.IDEQ(roots.ownerUserID),
			inboundRequestFixturePublishedOwnerUserNoWait(),
		).
		Only(ctx)
	if err == nil && owner != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(roots.organizationID),
			inboundRequestFixturePublishedOwnerOrganizationNoWait(),
		).
		Only(ctx)
	if err == nil && organizationRow != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	replyDelivery, err := tx.CalendarReplyDelivery.Query().
		Where(
			calendarreplydelivery.IDEQ(calendar.replyDeliveryID),
			inboundRequestFixturePublishedCalendarReplyDeliveryNoWait(),
		).
		Only(ctx)
	if err == nil && replyDelivery != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	mutationReceipt, err := tx.CalendarMutationReceipt.Query().
		Where(
			calendarmutationreceipt.IDEQ(calendar.mutationReceiptID),
			inboundRequestFixturePublishedCalendarMutationReceiptNoWait(),
		).
		Only(ctx)
	if err == nil && mutationReceipt != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	inboundMessageReceipt, err := tx.CalendarInboundMessageReceipt.Query().
		Where(
			calendarinboundmessagereceipt.IDEQ(calendar.inboundMessageReceiptID),
			inboundRequestFixturePublishedCalendarReceiptNoWait(),
		).
		Only(ctx)
	if err == nil && inboundMessageReceipt != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	inboundObjectState, err := tx.CalendarInboundObjectState.Query().
		Where(
			calendarinboundobjectstate.IDEQ(calendar.inboundObjectStateID),
			inboundRequestFixturePublishedCalendarObjectStateNoWait(),
		).
		Only(ctx)
	if err == nil && inboundObjectState != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	attendee, err := tx.CalendarAttendee.Query().
		Where(
			calendarattendee.IDEQ(calendar.attendeeID),
			inboundRequestFixturePublishedCalendarAttendeeNoWait(),
		).
		Only(ctx)
	if err == nil && attendee != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	event, err := tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(calendar.eventID),
			inboundRequestFixturePublishedCalendarEventNoWait(),
		).
		Only(ctx)
	if err == nil && event != nil {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}
	if !ent.IsNotFound(err) {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	p3Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("mail_message"),
			outboxevent.OriginIDEQ(roots.mailMessageID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxMailReceivedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p3Rows) != 0 {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	p4Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("calendar_event"),
			outboxevent.OriginIDEQ(calendar.eventID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p4Rows) != 0 {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	p5Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("calendar_attendee"),
			outboxevent.OriginIDEQ(calendar.attendeeID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p5Rows) != 0 {
		return inboundRequestFixturePublishedCommitUnknownObservationPartial
	}

	return inboundRequestFixturePublishedCommitUnknownObservationExactAbsent
}

func (c *Controller) reconcileInboundRequestFixturePublishedCleanupCommitUnknown(
	ctx context.Context,
	expected inboundRequestFixtureRecord,
) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(expected.nonce) {
		return errInboundRequestFixtureLifecycle
	}

	canonicalPlan, err := newInboundRequestFixturePlan(expected.nonce)
	if err != nil ||
		expected.plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(expected.persisted, expected.plan) ||
		expected.state != inboundRequestFixtureStateCleanupCommitUnknown ||
		!inboundRequestFixtureStateAcceptsCleanupDomainProof(expected.state, expected.cleanupProof) ||
		!expected.cleanupProof.present ||
		expected.engineFixture.accountID == "" ||
		expected.engineFixture.primaryIdentityID == "" ||
		expected.engineFixture.folderID == "" ||
		expected.engineFixture.messageID == "" ||
		expected.engineFixture.version == "" {
		return errInboundRequestFixtureLifecycle
	}

	calendar, valid := newInboundRequestFixtureCalendarProof(
		expected.cleanupProof.inboundMessageReceiptID,
		expected.cleanupProof.inboundObjectStateID,
		expected.cleanupProof.eventID,
		expected.cleanupProof.attendeeID,
		expected.cleanupProof.replyDeliveryID,
		expected.cleanupProof.mutationReceiptID,
	)
	cleanup, valid := newCleanupDomainProof(calendar)
	if !valid || cleanup != expected.cleanupProof {
		return errInboundRequestFixtureLifecycle
	}

	authority, valid := newInboundRequestFixturePublishedMappedEngineAuthority(
		expected.plan,
		expected.engineFixture,
	)
	if !valid {
		return errInboundRequestFixtureLifecycle
	}

	tx, err := c.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return errInboundRequestFixtureLifecycle
	}
	finished := false
	commitAttempted := false
	defer func() {
		if !finished && !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	observation := observeInboundRequestFixturePublishedCommitUnknown(
		ctx,
		tx,
		expected.plan,
		expected.persisted,
		authority,
		expected.cleanupProof,
	)
	switch observation {
	case inboundRequestFixturePublishedCommitUnknownObservationExactPresent:
		commitAttempted = true
		if err := tx.Commit(); err != nil {
			return errInboundRequestFixtureLifecycle
		}
		finished = true
		if _, restored := c.restoreInboundRequestFixturePublishedAfterCommitUnknownPresent(
			expected.nonce,
			expected,
		); !restored {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	case inboundRequestFixturePublishedCommitUnknownObservationExactAbsent:
		commitAttempted = true
		if err := tx.Commit(); err != nil {
			return errInboundRequestFixtureLifecycle
		}
		finished = true
		claimed, claimedOK := c.claimInboundRequestFixturePublishedCleanupEngineFromCommitUnknown(
			expected.nonce,
			expected,
		)
		if !claimedOK {
			return errInboundRequestFixtureLifecycle
		}
		_, err := c.finalizeInboundRequestFixturePublishedCleanupEngine(claimed)
		return err
	default:
		return errInboundRequestFixtureLifecycle
	}
}
