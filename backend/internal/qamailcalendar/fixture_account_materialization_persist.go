//go:build qa && !release

package qamailcalendar

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/user"
)

var errAccountMaterializationFixturePersist = errors.New("qa mail account materialization fixture persistence failed")

type accountMaterializationFixturePersisted struct {
	organizationID uuid.UUID
	ownerUserID    uuid.UUID
	mailDomainID   uuid.UUID
}

func persistAccountMaterializationFixture(
	ctx context.Context,
	client *ent.Client,
	plan accountMaterializationFixturePlan,
) (accountMaterializationFixturePersisted, error) {
	if ctx == nil || client == nil || !accountMaterializationFixturePlanMatchesCanonical(plan) {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}

	tx, err := client.Tx(ctx)
	if err != nil || tx == nil {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.Organization.Create().
		SetID(plan.organizationID).
		SetName("QA Mail Materialization " + plan.nonce).
		SetSlug(plan.organizationSlug).
		SetInviteToken("qa-mail-materialization-invite-" + plan.nonce).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}

	if _, err := tx.User.Create().
		SetID(plan.ownerUserID).
		SetEmail(plan.ownerAddress).
		SetName("QA Owner").
		SetDisplayName("QA Owner").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOrganizationID(plan.organizationID).
		SetTimezone("UTC").
		SetLanguage("ru").
		SetOnboardingCompleted(true).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}

	if _, err := tx.MailDomain.Create().
		SetID(plan.mailDomainID).
		SetOrganizationID(plan.organizationID).
		SetDomain(plan.domain).
		SetEngineProvisioningKey(plan.domainProvisioningKey).
		SetEngineDomainID(plan.engineDomainID).
		SetDesiredPresent(true).
		SetEnabled(true).
		SetIsDefault(true).
		SetState(maildomain.State("verified")).
		SetVerificationStatus(maildomain.VerificationStatus("verified")).
		SetDkimStatus(maildomain.DkimStatus("valid")).
		SetDesiredGeneration(1).
		SetDesiredHashVersion(1).
		SetDesiredHash(append([]byte(nil), plan.domainControlHash[:]...)).
		SetDeployedGeneration(1).
		SetDeployedHashVersion(1).
		SetDeployedHash(append([]byte(nil), plan.domainControlHash[:]...)).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx); err != nil {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}

	accountCount, err := tx.MailAccount.Query().
		Where(mailaccount.UserIDEQ(plan.ownerUserID)).
		Count(ctx)
	if err != nil || accountCount != 0 {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}

	if err := tx.Commit(); err != nil {
		return accountMaterializationFixturePersisted{}, errAccountMaterializationFixturePersist
	}
	committed = true

	return accountMaterializationFixturePersisted{
		organizationID: plan.organizationID,
		ownerUserID:    plan.ownerUserID,
		mailDomainID:   plan.mailDomainID,
	}, nil
}
