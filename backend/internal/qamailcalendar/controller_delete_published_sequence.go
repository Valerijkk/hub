//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
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

func deleteInboundRequestFixturePublishedSequenceExact(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
	mailClosure inboundRequestFixturePublishedMailClosure,
	calendar inboundRequestFixtureCalendarProof,
	external inboundRequestFixturePublishedExternalProof,
	outbox inboundRequestFixturePublishedProjectOutboxProof,
) bool {
	roots := lockedRoots.roots
	if ctx == nil ||
		tx == nil ||
		!inboundRequestFixtureRootProofMatchesPlan(roots, plan, persisted) ||
		!engineAuthority.valid() ||
		engineAuthority.nonce != plan.nonce ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			roots,
			plan,
			persisted,
			engineAuthority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) ||
		!mailClosure.valid() ||
		mailClosure.roots != roots ||
		!calendar.valid() ||
		!external.valid(roots, calendar) ||
		!outbox.valid(roots, calendar, engineAuthority) {
		return false
	}
	if _, valid := newInboundRequestFixtureCompleteGraphProof(
		roots,
		mailClosure.falseSent,
		calendar,
	); !valid {
		return false
	}

	affected, err := tx.OutboxEvent.Delete().
		Where(
			outboxevent.IDEQ(outbox.p3OutboxID),
			outboxevent.OriginKindEQ("mail_message"),
			outboxevent.OriginIDEQ(roots.mailMessageID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxMailReceivedSubject),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.OutboxEvent.Delete().
		Where(
			outboxevent.IDEQ(outbox.p4OutboxID),
			outboxevent.OriginKindEQ("calendar_event"),
			outboxevent.OriginIDEQ(calendar.eventID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.OutboxEvent.Delete().
		Where(
			outboxevent.IDEQ(outbox.p5OutboxID),
			outboxevent.OriginKindEQ("calendar_attendee"),
			outboxevent.OriginIDEQ(calendar.attendeeID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	if !deleteInboundRequestFixturePublishedCalendarExact(
		ctx,
		tx,
		plan,
		persisted,
		roots,
		calendar,
	) {
		return false
	}

	affected, err = tx.MailMessageFolder.Delete().
		Where(
			mailmessagefolder.IDEQ(roots.mailMessageFolderID),
			mailmessagefolder.MailMessageIDEQ(roots.mailMessageID),
			mailmessagefolder.MailFolderIDEQ(roots.mailFolderID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailMessage.Delete().
		Where(
			mailmessage.IDEQ(roots.mailMessageID),
			mailmessage.MailAccountIDEQ(roots.mailAccountID),
			mailmessage.MailThreadIDEQ(roots.mailThreadID),
			mailmessage.EngineMessageIDEQ(engineAuthority.messageID),
			mailmessage.ReceivedEventIDEQ(roots.envelopeEventID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailThread.Delete().
		Where(
			mailthread.IDEQ(roots.mailThreadID),
			mailthread.MailAccountIDEQ(roots.mailAccountID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailFolder.Delete().
		Where(
			mailfolder.IDEQ(roots.mailFolderID),
			mailfolder.MailAccountIDEQ(roots.mailAccountID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailIdentity.Delete().
		Where(
			mailidentity.IDEQ(roots.mailIdentityID),
			mailidentity.MailAccountIDEQ(roots.mailAccountID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailAccount.Delete().
		Where(
			mailaccount.IDEQ(roots.mailAccountID),
			mailaccount.OrganizationIDEQ(roots.organizationID),
			mailaccount.UserIDEQ(roots.ownerUserID),
			mailaccount.MailDomainIDEQ(roots.mailDomainID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.MailDomain.Delete().
		Where(
			maildomain.IDEQ(roots.mailDomainID),
			maildomain.OrganizationIDEQ(roots.organizationID),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.User.Delete().
		Where(
			user.IDEQ(roots.ownerUserID),
			user.HasOrganizationWith(organization.IDEQ(roots.organizationID)),
		).
		Exec(ctx)
	if err != nil || affected != 1 {
		return false
	}

	affected, err = tx.Organization.Delete().
		Where(
			organization.IDEQ(roots.organizationID),
			organization.SlugEQ(plan.organizationSlug),
		).
		Exec(ctx)
	return err == nil && affected == 1
}
