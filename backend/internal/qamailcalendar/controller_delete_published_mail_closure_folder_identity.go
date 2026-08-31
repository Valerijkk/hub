//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailusersettings"
	"github.com/valerii/hub/backend/ent/predicate"
)

// inboundRequestFixturePublishedMailFolderIdentityClosureObservation records
// the frozen slots 10-15 of the published-cleanup Mail closure. It contains
// only comparable values, so it can be compared by the later full-proof
// coordinator without retaining transaction-bound Ent rows.
type inboundRequestFixturePublishedMailFolderIdentityClosureObservation struct {
	roots inboundRequestFixtureRootProof

	childFolderCount int

	messageFolderCount         int
	messageFolderID            uuid.UUID
	messageFolderMailMessageID uuid.UUID
	messageFolderMailFolderID  uuid.UUID

	accountFolderCount         int
	accountFolderID            uuid.UUID
	accountFolderMailAccountID uuid.UUID

	fromIdentityDraftCount        int
	pendingFromIdentityDraftCount int
	defaultIdentitySettingsCount  int
}

func (observation inboundRequestFixturePublishedMailFolderIdentityClosureObservation) valid() bool {
	return observation.roots.valid() &&
		observation.childFolderCount == 0 &&
		observation.messageFolderCount == 1 &&
		observation.messageFolderID == observation.roots.mailMessageFolderID &&
		observation.messageFolderMailMessageID == observation.roots.mailMessageID &&
		observation.messageFolderMailFolderID == observation.roots.mailFolderID &&
		observation.accountFolderCount == 1 &&
		observation.accountFolderID == observation.roots.mailFolderID &&
		observation.accountFolderMailAccountID == observation.roots.mailAccountID &&
		observation.fromIdentityDraftCount == 0 &&
		observation.pendingFromIdentityDraftCount == 0 &&
		observation.defaultIdentitySettingsCount == 0
}

// observeInboundRequestFixturePublishedMailFolderIdentityClosure locks and
// proves the folder/identity half of the Mail closure in the caller-owned
// transaction. It intentionally neither constructs the false-sent proof nor
// reads the message/account closure slots owned by its sibling packet.
func observeInboundRequestFixturePublishedMailFolderIdentityClosure(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
) (inboundRequestFixturePublishedMailFolderIdentityClosureObservation, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!inboundRequestFixturePublishedMailFolderIdentityClosureRootsValid(roots, lockedRoots) {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	childFolders, err := tx.MailFolder.Query().
		Where(
			mailfolder.ParentIDEQ(roots.mailFolderID),
			inboundRequestFixturePublishedMailFolderClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(childFolders) != 0 {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	messageFolders, err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.MailFolderIDEQ(roots.mailFolderID),
			inboundRequestFixturePublishedMailMessageFolderClosureNoWait(),
		).
		Order(ent.Asc(mailmessagefolder.FieldID)).
		All(ctx)
	if err != nil ||
		len(messageFolders) != 1 ||
		messageFolders[0] == nil ||
		messageFolders[0].ID != roots.mailMessageFolderID ||
		messageFolders[0].MailMessageID != roots.mailMessageID ||
		messageFolders[0].MailFolderID != roots.mailFolderID {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	accountFolders, err := tx.MailFolder.Query().
		Where(
			mailfolder.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailFolderClosureNoWait(),
		).
		Order(ent.Asc(mailfolder.FieldID)).
		All(ctx)
	if err != nil ||
		len(accountFolders) != 1 ||
		accountFolders[0] == nil ||
		accountFolders[0].ID != lockedRoots.mail.folder.ID ||
		accountFolders[0].ID != roots.mailFolderID ||
		accountFolders[0].MailAccountID != roots.mailAccountID {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	fromIdentityDrafts, err := tx.MailDraft.Query().
		Where(
			maildraft.FromIdentityIDEQ(roots.mailIdentityID),
			inboundRequestFixturePublishedMailDraftClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(fromIdentityDrafts) != 0 {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	pendingFromIdentityDrafts, err := tx.MailDraft.Query().
		Where(
			maildraft.PendingFromIdentityIDEQ(roots.mailIdentityID),
			inboundRequestFixturePublishedMailDraftClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(pendingFromIdentityDrafts) != 0 {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	defaultIdentitySettings, err := tx.MailUserSettings.Query().
		Where(
			mailusersettings.DefaultIdentityIDEQ(roots.mailIdentityID),
			inboundRequestFixturePublishedMailUserSettingsClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(defaultIdentitySettings) != 0 {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}

	observation := inboundRequestFixturePublishedMailFolderIdentityClosureObservation{
		roots:                         roots,
		childFolderCount:              len(childFolders),
		messageFolderCount:            len(messageFolders),
		messageFolderID:               messageFolders[0].ID,
		messageFolderMailMessageID:    messageFolders[0].MailMessageID,
		messageFolderMailFolderID:     messageFolders[0].MailFolderID,
		accountFolderCount:            len(accountFolders),
		accountFolderID:               accountFolders[0].ID,
		accountFolderMailAccountID:    accountFolders[0].MailAccountID,
		fromIdentityDraftCount:        len(fromIdentityDrafts),
		pendingFromIdentityDraftCount: len(pendingFromIdentityDrafts),
		defaultIdentitySettingsCount:  len(defaultIdentitySettings),
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedMailFolderIdentityClosureObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedMailFolderIdentityClosureRootsValid(
	roots inboundRequestFixtureRootProof,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
) bool {
	if !roots.valid() ||
		lockedRoots.roots != roots ||
		lockedRoots.mail.roots != roots ||
		lockedRoots.account.roots != roots ||
		lockedRoots.owner.roots != roots ||
		lockedRoots.mail.messageFolder == nil ||
		lockedRoots.mail.message == nil ||
		lockedRoots.mail.thread == nil ||
		lockedRoots.mail.folder == nil ||
		lockedRoots.account.account == nil ||
		lockedRoots.account.primaryIdentity == nil ||
		lockedRoots.account.domain == nil ||
		lockedRoots.owner.owner == nil ||
		lockedRoots.owner.organization == nil {
		return false
	}

	return lockedRoots.mail.messageFolder.ID == roots.mailMessageFolderID &&
		lockedRoots.mail.messageFolder.MailMessageID == roots.mailMessageID &&
		lockedRoots.mail.messageFolder.MailFolderID == roots.mailFolderID &&
		lockedRoots.mail.message.ID == roots.mailMessageID &&
		lockedRoots.mail.message.MailAccountID == roots.mailAccountID &&
		lockedRoots.mail.message.MailThreadID == roots.mailThreadID &&
		lockedRoots.mail.thread.ID == roots.mailThreadID &&
		lockedRoots.mail.thread.MailAccountID == roots.mailAccountID &&
		lockedRoots.mail.folder.ID == roots.mailFolderID &&
		lockedRoots.mail.folder.MailAccountID == roots.mailAccountID &&
		lockedRoots.account.account.ID == roots.mailAccountID &&
		lockedRoots.account.account.OrganizationID == roots.organizationID &&
		lockedRoots.account.account.UserID == roots.ownerUserID &&
		lockedRoots.account.account.MailDomainID == roots.mailDomainID &&
		lockedRoots.account.primaryIdentity.ID == roots.mailIdentityID &&
		lockedRoots.account.primaryIdentity.MailAccountID == roots.mailAccountID &&
		lockedRoots.account.domain.ID == roots.mailDomainID &&
		lockedRoots.account.domain.OrganizationID == roots.organizationID &&
		lockedRoots.owner.owner.ID == roots.ownerUserID &&
		lockedRoots.owner.organization.ID == roots.organizationID
}

func inboundRequestFixturePublishedMailFolderClosureNoWait() predicate.MailFolder {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailMessageFolderClosureNoWait() predicate.MailMessageFolder {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailDraftClosureNoWait() predicate.MailDraft {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailUserSettingsClosureNoWait() predicate.MailUserSettings {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
