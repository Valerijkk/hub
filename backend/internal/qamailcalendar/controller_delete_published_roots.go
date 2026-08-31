//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

// inboundRequestFixtureLockedPublishedRoots is the private transaction-local
// aggregate of the three ordered root-lock proofs. It retains Ent rows only
// while the caller-owned transaction is open; only roots is value authority.
type inboundRequestFixtureLockedPublishedRoots struct {
	roots   inboundRequestFixtureRootProof
	mail    inboundRequestFixtureLockedMailRoots
	account inboundRequestFixtureLockedAccountRoots
	owner   inboundRequestFixtureLockedOwnerRoots
}

// lockInboundRequestFixturePublishedRoots proves the complete deterministic
// nine-root graph in the frozen lock order: Mail roots, account roots, then
// owner roots. The caller owns transaction isolation, commit, and rollback.
func lockInboundRequestFixturePublishedRoots(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
) (inboundRequestFixtureLockedPublishedRoots, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce {
		return inboundRequestFixtureLockedPublishedRoots{}, false
	}

	mailRoots, locked := lockInboundRequestFixturePublishedMailRoots(
		ctx,
		tx,
		plan,
		persisted,
		authority,
	)
	if !locked {
		return inboundRequestFixtureLockedPublishedRoots{}, false
	}

	accountRoots, locked := lockInboundRequestFixturePublishedAccountRoots(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		mailRoots,
	)
	if !locked {
		return inboundRequestFixtureLockedPublishedRoots{}, false
	}

	ownerRoots, locked := lockInboundRequestFixturePublishedOwnerRoots(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		mailRoots,
		accountRoots,
	)
	if !locked || !inboundRequestFixtureLockedPublishedRootsValid(
		roots,
		plan,
		persisted,
		authority,
		mailRoots,
		accountRoots,
		ownerRoots,
	) {
		return inboundRequestFixtureLockedPublishedRoots{}, false
	}

	return inboundRequestFixtureLockedPublishedRoots{
		roots:   roots,
		mail:    mailRoots,
		account: accountRoots,
		owner:   ownerRoots,
	}, true
}

func inboundRequestFixtureLockedPublishedRootsValid(
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	mailRoots inboundRequestFixtureLockedMailRoots,
	accountRoots inboundRequestFixtureLockedAccountRoots,
	ownerRoots inboundRequestFixtureLockedOwnerRoots,
) bool {
	if !inboundRequestFixtureRootProofMatchesPlan(roots, plan, persisted) ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		mailRoots.roots != roots ||
		accountRoots.roots != roots ||
		ownerRoots.roots != roots ||
		mailRoots.messageFolder == nil ||
		mailRoots.message == nil ||
		mailRoots.thread == nil ||
		mailRoots.folder == nil ||
		accountRoots.account == nil ||
		accountRoots.primaryIdentity == nil ||
		accountRoots.domain == nil ||
		ownerRoots.owner == nil ||
		ownerRoots.organization == nil {
		return false
	}

	return mailRoots.messageFolder.ID == roots.mailMessageFolderID &&
		mailRoots.messageFolder.MailMessageID == roots.mailMessageID &&
		mailRoots.messageFolder.MailFolderID == roots.mailFolderID &&
		mailRoots.message.ID == roots.mailMessageID &&
		mailRoots.message.MailAccountID == roots.mailAccountID &&
		mailRoots.message.MailThreadID == roots.mailThreadID &&
		mailRoots.thread.ID == roots.mailThreadID &&
		mailRoots.thread.MailAccountID == roots.mailAccountID &&
		mailRoots.folder.ID == roots.mailFolderID &&
		mailRoots.folder.MailAccountID == roots.mailAccountID &&
		accountRoots.account.ID == roots.mailAccountID &&
		accountRoots.account.OrganizationID == roots.organizationID &&
		accountRoots.account.UserID == roots.ownerUserID &&
		accountRoots.account.MailDomainID == roots.mailDomainID &&
		accountRoots.primaryIdentity.ID == roots.mailIdentityID &&
		accountRoots.primaryIdentity.MailAccountID == roots.mailAccountID &&
		accountRoots.domain.ID == roots.mailDomainID &&
		accountRoots.domain.OrganizationID == roots.organizationID &&
		ownerRoots.owner.ID == roots.ownerUserID &&
		ownerRoots.organization.ID == roots.organizationID
}
