//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedMailClosureMessageObservation is the private,
// comparable result of C2a. The later false-sent shard owns drafts and
// submissions; this shard proves only the original inbound message cardinality.
type inboundRequestFixturePublishedMailClosureMessageObservation struct {
	originalInboundMessageCount int
}

func (observation inboundRequestFixturePublishedMailClosureMessageObservation) valid() bool {
	return observation.originalInboundMessageCount == 1
}

// lockInboundRequestFixturePublishedMailClosureMessage observes the first nine
// frozen Mail closure slots inside the caller-owned transaction. It neither
// mutates rows nor constructs the final false-sent proof.
func lockInboundRequestFixturePublishedMailClosureMessage(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
) (inboundRequestFixturePublishedMailClosureMessageObservation, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		lockedRoots.roots != roots ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			roots,
			plan,
			persisted,
			authority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	attachments, err := tx.MailAttachment.Query().
		Where(
			mailattachment.MailMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureAttachmentNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(attachments) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	draftsReply, err := tx.MailDraft.Query().
		Where(
			maildraft.ReplyToMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureDraftNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(draftsReply) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	draftsForward, err := tx.MailDraft.Query().
		Where(
			maildraft.ForwardMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureDraftNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(draftsForward) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	draftsPendingReply, err := tx.MailDraft.Query().
		Where(
			maildraft.PendingReplyToMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureDraftNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(draftsPendingReply) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	draftsPendingForward, err := tx.MailDraft.Query().
		Where(
			maildraft.PendingForwardMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureDraftNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(draftsPendingForward) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	messageFolders, err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.MailMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureMessageFolderNoWait(),
		).
		Order(ent.Asc(mailmessagefolder.FieldID)).
		All(ctx)
	if err != nil ||
		len(messageFolders) != 1 ||
		!inboundRequestFixturePublishedMailClosureMessageFolderMatchesLocked(
			messageFolders[0],
			lockedRoots.mail.messageFolder,
			roots,
		) {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	messageLabels, err := tx.MailMessageLabel.Query().
		Where(
			mailmessagelabel.MailMessageIDEQ(roots.mailMessageID),
			inboundRequestFixturePublishedMailClosureMessageLabelNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(messageLabels) != 0 {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	messagesByThread, err := tx.MailMessage.Query().
		Where(
			mailmessage.MailThreadIDEQ(roots.mailThreadID),
			inboundRequestFixturePublishedMailClosureMessageNoWait(),
		).
		Order(ent.Asc(mailmessage.FieldID)).
		All(ctx)
	if err != nil ||
		len(messagesByThread) != 1 ||
		!inboundRequestFixturePublishedMailClosureOriginalInboundMessageMatchesLocked(
			messagesByThread[0],
			lockedRoots.mail.message,
			roots,
		) {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	messagesByAccount, err := tx.MailMessage.Query().
		Where(
			mailmessage.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailClosureMessageNoWait(),
		).
		Order(ent.Asc(mailmessage.FieldID)).
		All(ctx)
	if err != nil ||
		len(messagesByAccount) != 1 ||
		!inboundRequestFixturePublishedMailClosureOriginalInboundMessageMatchesLocked(
			messagesByAccount[0],
			lockedRoots.mail.message,
			roots,
		) {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}

	observation := inboundRequestFixturePublishedMailClosureMessageObservation{
		originalInboundMessageCount: len(messagesByAccount),
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedMailClosureMessageObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedMailClosureMessageFolderMatchesLocked(
	value *ent.MailMessageFolder,
	locked *ent.MailMessageFolder,
	roots inboundRequestFixtureRootProof,
) bool {
	return value != nil &&
		locked != nil &&
		value.ID == roots.mailMessageFolderID &&
		value.ID == locked.ID &&
		value.MailMessageID == roots.mailMessageID &&
		value.MailMessageID == locked.MailMessageID &&
		value.MailFolderID == roots.mailFolderID &&
		value.MailFolderID == locked.MailFolderID &&
		value.EngineState == nil &&
		locked.EngineState == nil &&
		value.LastSeenSyncEpoch == locked.LastSeenSyncEpoch &&
		value.ProjectedAt.Equal(locked.ProjectedAt) &&
		value.CreatedAt.Equal(locked.CreatedAt) &&
		value.UpdatedAt.Equal(locked.UpdatedAt)
}

func inboundRequestFixturePublishedMailClosureOriginalInboundMessageMatchesLocked(
	value *ent.MailMessage,
	locked *ent.MailMessage,
	roots inboundRequestFixtureRootProof,
) bool {
	return value != nil &&
		locked != nil &&
		value.ID == roots.mailMessageID &&
		value.ID == locked.ID &&
		value.MailAccountID == roots.mailAccountID &&
		value.MailAccountID == locked.MailAccountID &&
		value.MailThreadID == roots.mailThreadID &&
		value.MailThreadID == locked.MailThreadID &&
		value.EngineMessageID == locked.EngineMessageID &&
		value.EngineVersion == locked.EngineVersion &&
		value.ReceivedEventID != nil &&
		*value.ReceivedEventID == roots.envelopeEventID &&
		locked.ReceivedEventID != nil &&
		*locked.ReceivedEventID == roots.envelopeEventID &&
		value.ProjectionOrigin == mailmessage.ProjectionOrigin("inbound_or_unknown") &&
		locked.ProjectionOrigin == mailmessage.ProjectionOrigin("inbound_or_unknown") &&
		value.ProjectionOrigin == locked.ProjectionOrigin &&
		!value.IsDraft &&
		!locked.IsDraft &&
		value.SentAt == nil &&
		locked.SentAt == nil &&
		value.ReceivedAt != nil &&
		locked.ReceivedAt != nil &&
		value.ReceivedAt.Equal(*locked.ReceivedAt) &&
		value.DeletedAt == nil &&
		locked.DeletedAt == nil
}

func inboundRequestFixturePublishedMailClosureAttachmentNoWait() predicate.MailAttachment {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailClosureDraftNoWait() predicate.MailDraft {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailClosureMessageFolderNoWait() predicate.MailMessageFolder {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailClosureMessageLabelNoWait() predicate.MailMessageLabel {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailClosureMessageNoWait() predicate.MailMessage {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
