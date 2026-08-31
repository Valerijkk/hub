//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

// inboundRequestFixtureLockedOwnerRoots contains only rows locked by the
// caller-owned transaction. It must remain transaction-local and must not be
// retained in a Controller record.
type inboundRequestFixtureLockedOwnerRoots struct {
	roots        inboundRequestFixtureRootProof
	owner        *ent.User
	organization *ent.Organization
}

// lockInboundRequestFixturePublishedOwnerRoots locks and proves the final
// User -> Organization roots of the deterministic inbound fixture. It also
// locks the complete organization User set and accepts only the one fixture
// owner, so a later root delete cannot silently acquire another member.
// The caller owns SERIALIZABLE isolation, commit, and rollback.
func lockInboundRequestFixturePublishedOwnerRoots(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	mailRoots inboundRequestFixtureLockedMailRoots,
	accountRoots inboundRequestFixtureLockedAccountRoots,
) (inboundRequestFixtureLockedOwnerRoots, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePublishedOwnerRootsInputValid(
			plan,
			persisted,
			authority,
			roots,
			mailRoots,
			accountRoots,
		) {
		return inboundRequestFixtureLockedOwnerRoots{}, false
	}

	owner, err := tx.User.Query().
		Where(
			user.IDEQ(roots.ownerUserID),
			user.HasOrganizationWith(organization.IDEQ(roots.organizationID)),
			inboundRequestFixturePublishedOwnerUserNoWait(),
		).
		Only(ctx)
	if err != nil || !inboundRequestFixturePublishedOwnerValid(owner, roots, plan) {
		return inboundRequestFixtureLockedOwnerRoots{}, false
	}

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(roots.organizationID),
			organization.HasUsersWith(user.IDEQ(roots.ownerUserID)),
			inboundRequestFixturePublishedOwnerOrganizationNoWait(),
		).
		Only(ctx)
	if err != nil || !inboundRequestFixturePublishedOrganizationValid(organizationRow, roots, plan) {
		return inboundRequestFixtureLockedOwnerRoots{}, false
	}

	organizationUsers, err := tx.User.Query().
		Where(
			user.HasOrganizationWith(organization.IDEQ(roots.organizationID)),
			inboundRequestFixturePublishedOwnerUserNoWait(),
		).
		Order(ent.Asc(user.FieldID)).
		All(ctx)
	if err != nil ||
		len(organizationUsers) != 1 ||
		organizationUsers[0] == nil ||
		organizationUsers[0].ID != roots.ownerUserID ||
		!inboundRequestFixturePublishedOwnerValid(organizationUsers[0], roots, plan) {
		return inboundRequestFixtureLockedOwnerRoots{}, false
	}

	return inboundRequestFixtureLockedOwnerRoots{
		roots:        roots,
		owner:        owner,
		organization: organizationRow,
	}, true
}

func inboundRequestFixturePublishedOwnerRootsInputValid(
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	roots inboundRequestFixtureRootProof,
	mailRoots inboundRequestFixtureLockedMailRoots,
	accountRoots inboundRequestFixtureLockedAccountRoots,
) bool {
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	if err != nil ||
		plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		!roots.valid() ||
		mailRoots.roots != roots ||
		accountRoots.roots != roots ||
		mailRoots.messageFolder == nil ||
		mailRoots.message == nil ||
		mailRoots.thread == nil ||
		mailRoots.folder == nil ||
		accountRoots.account == nil ||
		accountRoots.primaryIdentity == nil ||
		accountRoots.domain == nil {
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
		accountRoots.domain.OrganizationID == roots.organizationID
}

func inboundRequestFixturePublishedOwnerValid(
	owner *ent.User,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return owner != nil &&
		owner.ID == roots.ownerUserID &&
		owner.Version == 1 &&
		owner.Email == plan.ownerAddress &&
		owner.Name == "QA Owner" &&
		owner.DisplayName == "QA Owner" &&
		owner.AvatarURL == "" &&
		owner.TelegramID == nil &&
		owner.Phone == "" &&
		owner.Position == "" &&
		owner.Department == "" &&
		owner.Timezone == "UTC" &&
		owner.Language == "ru" &&
		owner.Birthday == nil &&
		owner.SystemRole == user.SystemRoleMember &&
		owner.Status == user.StatusActive &&
		owner.PendingSince == nil &&
		owner.OnboardingCompleted &&
		owner.QuietHoursStart == nil &&
		owner.QuietHoursEnd == nil &&
		owner.CustomFields == nil &&
		owner.CreatedAt.Equal(plan.fixedAt) &&
		owner.UpdatedAt.Equal(plan.fixedAt)
}

func inboundRequestFixturePublishedOrganizationValid(
	organizationRow *ent.Organization,
	roots inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
) bool {
	return organizationRow != nil &&
		organizationRow.ID == roots.organizationID &&
		organizationRow.Version == 1 &&
		organizationRow.Name == "QA Mail Calendar "+plan.nonce &&
		organizationRow.Slug == plan.organizationSlug &&
		organizationRow.InviteToken == plan.correlationID &&
		organizationRow.CreatedAt.Equal(plan.fixedAt) &&
		organizationRow.UpdatedAt.Equal(plan.fixedAt)
}

func inboundRequestFixturePublishedOwnerUserNoWait() predicate.User {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedOwnerOrganizationNoWait() predicate.Organization {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
