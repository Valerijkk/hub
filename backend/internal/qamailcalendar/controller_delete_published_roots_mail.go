//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixtureLockedMailRoots contains rows locked by the caller's
// transaction. It must not escape that transaction or be retained in a
// Controller record.
type inboundRequestFixtureLockedMailRoots struct {
	roots         inboundRequestFixtureRootProof
	messageFolder *ent.MailMessageFolder
	message       *ent.MailMessage
	thread        *ent.MailThread
	folder        *ent.MailFolder
}

// lockInboundRequestFixturePublishedMailRoots locks and proves only the
// deterministic MailMessageFolder -> MailMessage -> MailThread -> MailFolder
// fixture roots. The caller owns transaction isolation, commit, and rollback.
func lockInboundRequestFixturePublishedMailRoots(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
) (inboundRequestFixtureLockedMailRoots, bool) {
	roots, rootsValid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!rootsValid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!engineAuthority.valid() ||
		engineAuthority.nonce != plan.nonce {
		return inboundRequestFixtureLockedMailRoots{}, false
	}

	messageFolderNoWait := predicate.MailMessageFolder(func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	})
	messageNoWait := predicate.MailMessage(func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	})
	threadNoWait := predicate.MailThread(func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	})
	folderNoWait := predicate.MailFolder(func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	})

	messageFolder, err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.IDEQ(roots.mailMessageFolderID),
			mailmessagefolder.MailMessageIDEQ(roots.mailMessageID),
			mailmessagefolder.MailFolderIDEQ(roots.mailFolderID),
			messageFolderNoWait,
		).
		Only(ctx)
	if err != nil ||
		messageFolder == nil ||
		messageFolder.ID != roots.mailMessageFolderID ||
		messageFolder.MailMessageID != roots.mailMessageID ||
		messageFolder.MailFolderID != roots.mailFolderID ||
		messageFolder.EngineState != nil ||
		messageFolder.LastSeenSyncEpoch != 0 ||
		!messageFolder.ProjectedAt.Equal(plan.fixedAt) ||
		!messageFolder.CreatedAt.Equal(plan.fixedAt) ||
		!messageFolder.UpdatedAt.Equal(plan.fixedAt) {
		return inboundRequestFixtureLockedMailRoots{}, false
	}

	message, err := tx.MailMessage.Query().
		Where(
			mailmessage.IDEQ(roots.mailMessageID),
			mailmessage.MailAccountIDEQ(roots.mailAccountID),
			mailmessage.MailThreadIDEQ(roots.mailThreadID),
			messageNoWait,
		).
		Only(ctx)
	if err != nil ||
		message == nil ||
		message.ID != roots.mailMessageID ||
		message.MailAccountID != roots.mailAccountID ||
		message.MailThreadID != roots.mailThreadID ||
		!engineAuthority.matchesMessageVersion(message.EngineMessageID, message.EngineVersion) ||
		message.InternetMessageID == nil ||
		*message.InternetMessageID != "<"+plan.itipUID+"@example.test>" ||
		message.InReplyTo != nil ||
		message.References == nil ||
		len(message.References) != 0 ||
		message.From.Email != plan.organizerAddress ||
		message.From.Name != "" ||
		message.To == nil ||
		len(message.To) != 1 ||
		message.To[0].Email != plan.ownerAddress ||
		message.To[0].Name != "" ||
		message.Cc == nil ||
		len(message.Cc) != 0 ||
		message.Bcc == nil ||
		len(message.Bcc) != 0 ||
		message.ReplyTo == nil ||
		len(message.ReplyTo) != 0 ||
		message.Subject != plan.subject ||
		message.Snippet != plan.snippet ||
		message.SentAt != nil ||
		message.ReceivedAt == nil ||
		!message.ReceivedAt.Equal(plan.fixedAt) ||
		message.SizeBytes != int64(len(plan.snippet)) ||
		message.Priority != mailmessage.Priority("normal") ||
		message.HasAttachments ||
		message.IsDraft ||
		message.IsSeen ||
		message.IsAnswered ||
		message.IsFlagged ||
		message.IsJunk ||
		message.EngineState != nil ||
		message.ProjectionOrigin != mailmessage.ProjectionOrigin("inbound_or_unknown") ||
		message.FlagsSyncedAt == nil ||
		!message.FlagsSyncedAt.Equal(plan.fixedAt) ||
		message.CalendarMethod == nil ||
		*message.CalendarMethod != mailmessage.CalendarMethod("request") ||
		message.CalendarPartLocatorHash == nil ||
		!bytes.Equal(*message.CalendarPartLocatorHash, plan.calendarLocatorHash[:]) ||
		message.AuthenticationClass != mailmessage.AuthenticationClass("trusted_internal") ||
		message.LastSeenSyncEpoch != 0 ||
		message.ReceivedEventID == nil ||
		*message.ReceivedEventID != roots.envelopeEventID ||
		message.DeletedAt != nil ||
		!message.CreatedAt.Equal(plan.fixedAt) ||
		!message.UpdatedAt.Equal(plan.fixedAt) {
		return inboundRequestFixtureLockedMailRoots{}, false
	}

	thread, err := tx.MailThread.Query().
		Where(
			mailthread.IDEQ(roots.mailThreadID),
			mailthread.MailAccountIDEQ(roots.mailAccountID),
			threadNoWait,
		).
		Only(ctx)
	if err != nil ||
		thread == nil ||
		thread.ID != roots.mailThreadID ||
		thread.MailAccountID != roots.mailAccountID ||
		thread.ThreadKeyVersion != 1 ||
		!bytes.Equal(thread.ThreadKey, plan.threadKey[:]) ||
		thread.NormalizedSubject != plan.normalizedSubject ||
		thread.ParticipantAddresses == nil ||
		len(thread.ParticipantAddresses) != 2 ||
		thread.ParticipantAddresses[0] != plan.organizerAddress ||
		thread.ParticipantAddresses[1] != plan.ownerAddress ||
		!thread.LastMessageAt.Equal(plan.fixedAt) ||
		thread.MessageCount != 1 ||
		thread.UnreadCount != 1 ||
		thread.IsPinned ||
		thread.HasAttachments ||
		thread.Snippet != plan.snippet ||
		thread.LastSeenSyncEpoch != 0 ||
		thread.DeletedAt != nil ||
		!thread.CreatedAt.Equal(plan.fixedAt) ||
		!thread.UpdatedAt.Equal(plan.fixedAt) {
		return inboundRequestFixtureLockedMailRoots{}, false
	}

	folder, err := tx.MailFolder.Query().
		Where(
			mailfolder.IDEQ(roots.mailFolderID),
			mailfolder.MailAccountIDEQ(roots.mailAccountID),
			folderNoWait,
		).
		Only(ctx)
	if err != nil ||
		folder == nil ||
		folder.ID != roots.mailFolderID ||
		folder.MailAccountID != roots.mailAccountID ||
		folder.EngineFolderID == nil ||
		!inboundRequestFixturePublishedFolderIDMatchesAuthority(engineAuthority, *folder.EngineFolderID) ||
		folder.ParentID != nil ||
		folder.Role != nil ||
		folder.Name != plan.folderName ||
		folder.NormalizedName != plan.normalizedFolderName ||
		folder.Color != nil ||
		folder.SortOrder != 0 ||
		folder.UnreadCount != 0 ||
		folder.TotalCount != 0 ||
		folder.ControlState != mailfolder.ControlStateActive ||
		folder.ControlReceiptID != nil ||
		folder.LastSeenSyncEpoch != 0 ||
		folder.Version != 1 ||
		folder.DeletedAt != nil ||
		!folder.CreatedAt.Equal(plan.fixedAt) ||
		!folder.UpdatedAt.Equal(plan.fixedAt) {
		return inboundRequestFixtureLockedMailRoots{}, false
	}

	return inboundRequestFixtureLockedMailRoots{
		roots:         roots,
		messageFolder: messageFolder,
		message:       message,
		thread:        thread,
		folder:        folder,
	}, true
}

func inboundRequestFixturePublishedFolderIDMatchesAuthority(
	authority inboundRequestFixturePublishedEngineAuthority,
	folderID string,
) bool {
	switch {
	case authority.mappedMode():
		return folderID == authority.mapped.folderID
	case authority.maplessMode():
		return folderID != ""
	default:
		return false
	}
}
