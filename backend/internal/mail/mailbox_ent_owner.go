package mail

import (
	"context"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

// ResolveMailboxOwner resolves one actor-owned mailbox snapshot from locked state.
func (repository *MailboxEntRepository) ResolveMailboxOwner(ctx context.Context, actorID uuid.UUID) (MailboxOwnerRecord, error) {
	if repository == nil || repository.client == nil || ctx == nil || actorID == uuid.Nil {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var record MailboxOwnerRecord
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		resolved, err := mailboxEntResolveOwnerForUpdate(ctx, transaction, actorID)
		if err != nil {
			return err
		}
		record = mailboxEntOwnerRecordCopy(resolved)
		return nil
	})
	if err != nil {
		return MailboxOwnerRecord{}, mailboxEntSafeError(err)
	}
	return mailboxEntOwnerRecordCopy(record), nil
}

func mailboxEntResolveOwnerForUpdate(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID) (MailboxOwnerRecord, error) {
	if transaction == nil || actorID == uuid.Nil {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	organizationRow, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxEntOwnerLookupError(err)
	}

	actor, err := transaction.User.Query().
		Where(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationRow.ID)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxEntOwnerLookupError(err)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.UserIDEQ(actor.ID),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxEntSafeError(err)
	}
	if len(accounts) == 0 {
		return MailboxOwnerRecord{
			ActorID:        actor.ID,
			OrganizationID: organizationRow.ID,
			State:          MailboxOwnerNotConfigured,
		}, nil
	}
	if len(accounts) != 1 {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	account := accounts[0]
	if account.ID == uuid.Nil || account.UserID != actor.ID || account.OrganizationID != organizationRow.ID {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	if MailboxOwnerState(account.Status) == MailboxOwnerDeleting {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorDeleting)
	}

	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(account.ID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return MailboxOwnerRecord{}, mailboxEntSafeError(err)
	}
	if len(identities) != 1 {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	identity := identities[0]
	if identity.ID == uuid.Nil || identity.MailAccountID != account.ID || identity.Kind != mailidentity.KindPrimary {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	return mailboxEntOwnerRecordFromRows(actor.ID, organizationRow.ID, account, identity)
}

func mailboxEntOwnerRecordFromRows(actorID, organizationID uuid.UUID, account *ent.MailAccount, identity *ent.MailIdentity) (MailboxOwnerRecord, error) {
	if actorID == uuid.Nil || organizationID == uuid.Nil || account == nil || identity == nil || account.ID == uuid.Nil || account.UserID != actorID || account.OrganizationID != organizationID {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	state := MailboxOwnerState(account.Status)
	projectionState := MailboxProjectionState(account.ProjectionState)
	if !validMailboxOwnerState(state) || state == MailboxOwnerNotConfigured || state == MailboxOwnerDeleting || !validMailboxProjectionState(projectionState) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	if identity.MailAccountID != account.ID || identity.Kind != mailidentity.KindPrimary || !mailboxEntPrimaryIdentityStateValid(identity) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	engineAccountID := mailboxEntStringPointer(account.EngineAccountID)
	if engineAccountID != nil && !validMailboxOwnerOpaque(*engineAccountID) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	primaryAddress := identity.NormalizedAddress
	if !validMailboxOwnerPrimaryAddress(primaryAddress) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	recoveryCode := mailboxEntProvisionRecoveryCode(account)
	if recoveryCode != nil && !validMailboxOwnerRecoveryCode(*recoveryCode) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	if !mailboxEntAccountStateValid(state, account.DesiredEnabled, engineAccountID, recoveryCode, identity) {
		return MailboxOwnerRecord{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}

	return mailboxEntOwnerRecordCopy(MailboxOwnerRecord{
		ActorID:         actorID,
		OrganizationID:  organizationID,
		AccountID:       account.ID,
		EngineAccountID: engineAccountID,
		State:           state,
		ProjectionState: projectionState,
		PrimaryAddress:  mailboxEntStringPointer(&primaryAddress),
		RecoveryCode:    recoveryCode,
	}), nil
}

func mailboxEntOwnerLookupError(err error) error {
	if ent.IsNotFound(err) {
		return mailboxServiceError(MailboxServiceErrorNotFound)
	}
	return mailboxEntSafeError(err)
}

func mailboxEntPrimaryIdentityStateValid(identity *ent.MailIdentity) bool {
	if identity == nil || identity.ID == uuid.Nil || identity.MailAccountID == uuid.Nil || identity.Kind != mailidentity.KindPrimary {
		return false
	}
	switch identity.Status {
	case mailidentity.StatusPending, mailidentity.StatusActive, mailidentity.StatusError:
		return true
	default:
		return false
	}
}

func mailboxEntProvisionRecoveryCode(account *ent.MailAccount) *MailboxOwnerRecoveryCode {
	if account == nil || account.ProvisionErrorCode == nil {
		return nil
	}
	code := MailboxOwnerRecoveryCode(*account.ProvisionErrorCode)
	return mailboxEntRecoveryCodePointer(&code)
}

func mailboxEntAccountStateValid(state MailboxOwnerState, desiredEnabled bool, engineAccountID *string, recoveryCode *MailboxOwnerRecoveryCode, identity *ent.MailIdentity) bool {
	if identity == nil {
		return false
	}

	switch state {
	case MailboxOwnerProvisioning:
		return true
	case MailboxOwnerActive:
		return desiredEnabled && engineAccountID != nil && recoveryCode == nil && identity.Status == mailidentity.StatusActive && identity.EngineIdentityID != nil
	case MailboxOwnerSuspended:
		return !desiredEnabled && engineAccountID != nil && recoveryCode == nil && identity.Status == mailidentity.StatusActive && identity.EngineIdentityID != nil
	case MailboxOwnerError:
		return recoveryCode != nil
	default:
		return false
	}
}
