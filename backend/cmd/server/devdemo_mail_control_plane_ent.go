//go:build !qa && !release

package main

import (
	"bytes"
	"context"
	"time"

	entgo "entgo.io/ent"
	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailboxpolicy"
	"github.com/valerii/hub/backend/ent/mailusersettings"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

func reconcileDevMailControlPlaneTransaction(
	ctx context.Context,
	client *appent.Client,
	plan devMailControlPlan,
	engineDomainID string,
	providerAccounts map[string]devMailProviderAccount,
) (devMailControlPlaneResult, error) {
	if !validDevMailControlContext(ctx) || client == nil || !validDevMailControlPlan(plan) || !validDevMailControlOpaqueID(engineDomainID) || len(providerAccounts) != len(plan.accounts) {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	for _, accountPlan := range plan.accounts {
		provider, ok := providerAccounts[accountPlan.email]
		if !ok || !validDevMailProviderAccountHandles(provider) {
			return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageProvider, devMailControlFailureUnavailable)
		}
		if err := verifyDevMailControlUser(ctx, client, accountPlan); err != nil {
			return devMailControlPlaneResult{}, err
		}
	}

	domain, domainCreated, err := reconcileDevMailControlDomain(ctx, client, plan.domain, engineDomainID, plan.now)
	if err != nil {
		return devMailControlPlaneResult{}, err
	}
	if err := verifyDevMailControlDefaultDomain(ctx, client, plan.domain); err != nil {
		return devMailControlPlaneResult{}, err
	}
	policyCreated, err := reconcileDevMailControlPolicy(ctx, client, plan.policy, plan.now)
	if err != nil {
		return devMailControlPlaneResult{}, err
	}

	result := devMailControlPlaneResult{
		accounts:        make(map[string]devMailControlPlaneAccount, len(plan.accounts)),
		domainsTotal:    1,
		accountsTotal:   len(plan.accounts),
		identitiesTotal: len(plan.accounts),
		settingsTotal:   len(plan.accounts),
		policiesTotal:   1,
	}
	if domainCreated {
		result.domainsCreated = 1
	} else {
		result.domainsReplayed = 1
	}
	if policyCreated {
		result.policiesCreated = 1
	} else {
		result.policiesReplayed = 1
	}

	for _, accountPlan := range plan.accounts {
		provider := providerAccounts[accountPlan.email]
		account, accountCreated, err := reconcileDevMailControlAccount(ctx, client, plan.domain, accountPlan, provider.accountID, plan.now)
		if err != nil {
			return devMailControlPlaneResult{}, err
		}
		identity, identityCreated, err := reconcileDevMailControlIdentity(ctx, client, account, accountPlan, provider.primaryIdentityID, plan.now)
		if err != nil {
			return devMailControlPlaneResult{}, err
		}
		settingsCreated, err := reconcileDevMailControlSettings(ctx, client, accountPlan, account, identity, plan.now)
		if err != nil {
			return devMailControlPlaneResult{}, err
		}
		if accountCreated {
			result.accountsCreated++
		} else {
			result.accountsReplayed++
		}
		if identityCreated {
			result.identitiesCreated++
		} else {
			result.identitiesReplayed++
		}
		if settingsCreated {
			result.settingsCreated++
		} else {
			result.settingsReplayed++
		}
		result.accounts[accountPlan.email] = devMailControlPlaneAccount{
			accountID:               account.ID,
			engineAccountID:         provider.accountID,
			enginePrimaryIdentityID: provider.primaryIdentityID,
		}
	}
	if domain == nil || !validDevMailControlResult(result) {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageResult, devMailControlFailureUnavailable)
	}
	return result, nil
}

func verifyDevMailControlUser(ctx context.Context, client *appent.Client, plan devMailControlAccountPlan) error {
	rows, err := client.User.Query().
		Where(
			user.IDEQ(plan.userID),
			user.EmailEQ(plan.email),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(plan.organizationID)),
		).
		ForUpdate().
		All(ctx)
	if err != nil {
		return newDevMailControlError(devMailControlStageUsers, devMailControlFailureUnavailable)
	}
	if len(rows) != 1 || rows[0] == nil || rows[0].ID != plan.userID || rows[0].Email != plan.email || rows[0].Status != user.StatusActive {
		return newDevMailControlError(devMailControlStageUsers, devMailControlFailureConflict)
	}
	return nil
}

func reconcileDevMailControlDomain(ctx context.Context, client *appent.Client, plan devMailControlDomainPlan, engineDomainID string, now time.Time) (*appent.MailDomain, bool, error) {
	rows, err := client.MailDomain.Query().
		Where(entgo.Or(
			maildomain.IDEQ(plan.id),
			maildomain.DomainEQ(plan.domain),
			maildomain.EngineProvisioningKeyEQ(plan.provisioningKey),
		)).
		Order(entgo.Asc(maildomain.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, false, newDevMailControlError(devMailControlStageDomain, devMailControlFailureUnavailable)
	}
	switch len(rows) {
	case 0:
		domain, err := client.MailDomain.Create().
			SetID(plan.id).
			SetOrganizationID(plan.organizationID).
			SetDomain(plan.domain).
			SetEngineProvisioningKey(plan.provisioningKey).
			SetEngineDomainID(engineDomainID).
			SetDesiredPresent(true).
			SetEnabled(true).
			SetIsDefault(true).
			SetState(maildomain.State("verified")).
			SetVerificationStatus(maildomain.VerificationStatus("verified")).
			SetDkimStatus(maildomain.DkimStatus("valid")).
			SetRequiredDNSRecords(mailmod.RequiredDNSRecords{}).
			SetDesiredGeneration(devMailControlGeneration).
			SetDesiredHashVersion(devMailControlHashVersion).
			SetDesiredHash(append([]byte(nil), plan.desiredHash[:]...)).
			SetDeployedGeneration(devMailControlGeneration).
			SetDeployedHashVersion(devMailControlHashVersion).
			SetDeployedHash(append([]byte(nil), plan.desiredHash[:]...)).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			return nil, false, devMailControlTransactionFailure(devMailControlStageDomain, err)
		}
		if !validDevMailControlDomainRow(domain, plan, engineDomainID) {
			return nil, false, newDevMailControlError(devMailControlStageDomain, devMailControlFailureUnavailable)
		}
		return domain, true, nil
	case 1:
		if !validDevMailControlDomainRow(rows[0], plan, engineDomainID) {
			return nil, false, newDevMailControlError(devMailControlStageDomain, devMailControlFailureConflict)
		}
		return rows[0], false, nil
	default:
		return nil, false, newDevMailControlError(devMailControlStageDomain, devMailControlFailureConflict)
	}
}

func verifyDevMailControlDefaultDomain(ctx context.Context, client *appent.Client, plan devMailControlDomainPlan) error {
	rows, err := client.MailDomain.Query().
		Where(
			maildomain.OrganizationIDEQ(plan.organizationID),
			maildomain.IsDefaultEQ(true),
			maildomain.DeletedAtIsNil(),
		).
		Order(entgo.Asc(maildomain.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return newDevMailControlError(devMailControlStageDomain, devMailControlFailureUnavailable)
	}
	if len(rows) != 1 || rows[0] == nil || rows[0].ID != plan.id {
		return newDevMailControlError(devMailControlStageDomain, devMailControlFailureConflict)
	}
	return nil
}

func reconcileDevMailControlPolicy(ctx context.Context, client *appent.Client, plan devMailControlPolicyPlan, now time.Time) (bool, error) {
	rows, err := client.MailboxPolicy.Query().
		Where(entgo.Or(
			mailboxpolicy.IDEQ(plan.id),
			mailboxpolicy.OrganizationIDEQ(plan.organizationID),
		)).
		Order(entgo.Asc(mailboxpolicy.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, newDevMailControlError(devMailControlStagePolicy, devMailControlFailureUnavailable)
	}
	switch len(rows) {
	case 0:
		policy, err := client.MailboxPolicy.Create().
			SetID(plan.id).
			SetOrganizationID(plan.organizationID).
			SetExternalMailEnabled(false).
			SetMaxMessageBytes(devMailControlMaxMessageBytes).
			SetMaxAttachmentBytes(devMailControlMaxAttachmentBytes).
			SetMaxRecipients(devMailControlMaxRecipients).
			SetSendRatePerMinute(devMailControlSendRatePerMinute).
			SetInstanceSendRatePerMinute(devMailControlInstanceSendRateMinute).
			SetTrashRetentionDays(devMailControlTrashRetentionDays).
			SetJunkRetentionDays(devMailControlJunkRetentionDays).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			return false, devMailControlTransactionFailure(devMailControlStagePolicy, err)
		}
		if !validDevMailControlPolicyRow(policy, plan) {
			return false, newDevMailControlError(devMailControlStagePolicy, devMailControlFailureUnavailable)
		}
		return true, nil
	case 1:
		if !validDevMailControlPolicyRow(rows[0], plan) {
			return false, newDevMailControlError(devMailControlStagePolicy, devMailControlFailureConflict)
		}
		return false, nil
	default:
		return false, newDevMailControlError(devMailControlStagePolicy, devMailControlFailureConflict)
	}
}

func reconcileDevMailControlAccount(ctx context.Context, client *appent.Client, domain devMailControlDomainPlan, plan devMailControlAccountPlan, engineAccountID string, now time.Time) (*appent.MailAccount, bool, error) {
	rows, err := client.MailAccount.Query().
		Where(entgo.Or(
			mailaccount.IDEQ(plan.id),
			mailaccount.UserIDEQ(plan.userID),
			mailaccount.EngineProvisioningKeyEQ(plan.provisioningKey),
			mailaccount.EngineAccountIDEQ(engineAccountID),
			entgo.And(mailaccount.MailDomainIDEQ(domain.id), mailaccount.LocalPartEQ(plan.localPart)),
		)).
		Order(entgo.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, false, newDevMailControlError(devMailControlStageAccount, devMailControlFailureUnavailable)
	}
	switch len(rows) {
	case 0:
		account, err := client.MailAccount.Create().
			SetID(plan.id).
			SetOrganizationID(plan.organizationID).
			SetUserID(plan.userID).
			SetMailDomainID(domain.id).
			SetEngineProvisioningKey(plan.provisioningKey).
			SetLocalPart(plan.localPart).
			SetEngineAccountID(engineAccountID).
			SetDesiredEnabled(true).
			SetStatus(mailaccount.Status("active")).
			SetControlGeneration(devMailControlGeneration).
			SetControlHashVersion(devMailControlHashVersion).
			SetControlHash(append([]byte(nil), plan.controlHash[:]...)).
			SetDeployedGeneration(devMailControlGeneration).
			SetDeployedHashVersion(devMailControlHashVersion).
			SetDeployedHash(append([]byte(nil), plan.controlHash[:]...)).
			SetProjectionState(mailaccount.ProjectionState("bootstrapping")).
			SetNextSyncAt(now).
			SetSyncScheduleUpdatedAt(now).
			SetEventCutoffAt(now).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			return nil, false, devMailControlTransactionFailure(devMailControlStageAccount, err)
		}
		if !validDevMailControlAccountRow(account, domain, plan, engineAccountID, now) {
			return nil, false, newDevMailControlError(devMailControlStageAccount, devMailControlFailureUnavailable)
		}
		return account, true, nil
	case 1:
		if !validDevMailControlAccountRow(rows[0], domain, plan, engineAccountID, now) {
			return nil, false, newDevMailControlError(devMailControlStageAccount, devMailControlFailureConflict)
		}
		return rows[0], false, nil
	default:
		return nil, false, newDevMailControlError(devMailControlStageAccount, devMailControlFailureConflict)
	}
}

func reconcileDevMailControlIdentity(ctx context.Context, client *appent.Client, account *appent.MailAccount, plan devMailControlAccountPlan, engineIdentityID string, now time.Time) (*appent.MailIdentity, bool, error) {
	if account == nil || account.ID != plan.id {
		return nil, false, newDevMailControlError(devMailControlStageIdentity, devMailControlFailureInvalid)
	}
	rows, err := client.MailIdentity.Query().
		Where(entgo.Or(
			mailidentity.IDEQ(plan.identityID),
			mailidentity.NormalizedAddressEQ(plan.email),
			entgo.And(mailidentity.MailAccountIDEQ(account.ID), mailidentity.KindEQ(mailidentity.Kind("primary"))),
			entgo.And(mailidentity.MailAccountIDEQ(account.ID), mailidentity.EngineIdentityIDEQ(engineIdentityID)),
		)).
		Order(entgo.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, false, newDevMailControlError(devMailControlStageIdentity, devMailControlFailureUnavailable)
	}
	switch len(rows) {
	case 0:
		identity, err := client.MailIdentity.Create().
			SetID(plan.identityID).
			SetMailAccountID(account.ID).
			SetNormalizedAddress(plan.email).
			SetKind(mailidentity.Kind("primary")).
			SetStatus(mailidentity.Status("active")).
			SetSendAllowed(true).
			SetEngineIdentityID(engineIdentityID).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			return nil, false, devMailControlTransactionFailure(devMailControlStageIdentity, err)
		}
		if !validDevMailControlIdentityRow(identity, account.ID, plan, engineIdentityID) {
			return nil, false, newDevMailControlError(devMailControlStageIdentity, devMailControlFailureUnavailable)
		}
		return identity, true, nil
	case 1:
		if !validDevMailControlIdentityRow(rows[0], account.ID, plan, engineIdentityID) {
			return nil, false, newDevMailControlError(devMailControlStageIdentity, devMailControlFailureConflict)
		}
		return rows[0], false, nil
	default:
		return nil, false, newDevMailControlError(devMailControlStageIdentity, devMailControlFailureConflict)
	}
}

func reconcileDevMailControlSettings(ctx context.Context, client *appent.Client, plan devMailControlAccountPlan, account *appent.MailAccount, identity *appent.MailIdentity, now time.Time) (bool, error) {
	if account == nil || identity == nil || account.ID != plan.id || identity.ID != plan.identityID {
		return false, newDevMailControlError(devMailControlStageSettings, devMailControlFailureInvalid)
	}
	rows, err := client.MailUserSettings.Query().
		Where(entgo.Or(
			mailusersettings.IDEQ(plan.settingsID),
			mailusersettings.MailAccountIDEQ(account.ID),
			mailusersettings.DefaultIdentityIDEQ(identity.ID),
		)).
		Order(entgo.Asc(mailusersettings.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, newDevMailControlError(devMailControlStageSettings, devMailControlFailureUnavailable)
	}
	switch len(rows) {
	case 0:
		settings, err := client.MailUserSettings.Create().
			SetID(plan.settingsID).
			SetMailAccountID(account.ID).
			SetDefaultIdentityID(identity.ID).
			SetRemoteImages(mailusersettings.RemoteImages("blocked")).
			SetComposeFormat(mailusersettings.ComposeFormat("rich")).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
		if err != nil {
			return false, devMailControlTransactionFailure(devMailControlStageSettings, err)
		}
		if !validDevMailControlSettingsRow(settings, account.ID, identity.ID, plan.settingsID) {
			return false, newDevMailControlError(devMailControlStageSettings, devMailControlFailureUnavailable)
		}
		return true, nil
	case 1:
		if !validDevMailControlSettingsRow(rows[0], account.ID, identity.ID, plan.settingsID) {
			return false, newDevMailControlError(devMailControlStageSettings, devMailControlFailureConflict)
		}
		return false, nil
	default:
		return false, newDevMailControlError(devMailControlStageSettings, devMailControlFailureConflict)
	}
}

func validDevMailControlDomainRow(row *appent.MailDomain, plan devMailControlDomainPlan, engineDomainID string) bool {
	return row != nil &&
		row.ID == plan.id &&
		row.OrganizationID == plan.organizationID &&
		row.Domain == plan.domain &&
		row.EngineProvisioningKey == plan.provisioningKey &&
		row.EngineDomainID != nil && *row.EngineDomainID == engineDomainID &&
		row.DesiredPresent && row.Enabled && row.IsDefault && row.DeletedAt == nil &&
		row.State == maildomain.State("verified") &&
		row.VerificationStatus == maildomain.VerificationStatus("verified") &&
		row.DkimStatus == maildomain.DkimStatus("valid") &&
		len(row.RequiredDNSRecords) == 0 &&
		row.DesiredGeneration == devMailControlGeneration &&
		row.DesiredHashVersion == devMailControlHashVersion &&
		bytes.Equal(row.DesiredHash, plan.desiredHash[:]) &&
		row.DeployedGeneration != nil && *row.DeployedGeneration == devMailControlGeneration &&
		row.DeployedHashVersion != nil && *row.DeployedHashVersion == devMailControlHashVersion &&
		row.DeployedHash != nil && bytes.Equal(row.DeployedHash, plan.desiredHash[:]) &&
		row.AttemptedGeneration == nil && row.AttemptedHashVersion == nil && row.AttemptedHash == nil &&
		row.LastErrorCode == nil && row.NextRetryAt == nil && row.LeaseToken == nil && row.LeaseUntil == nil
}

func validDevMailControlAccountRow(row *appent.MailAccount, domain devMailControlDomainPlan, plan devMailControlAccountPlan, engineAccountID string, now time.Time) bool {
	if row == nil || !validDevMailControlNow(now) || row.ID != plan.id || row.OrganizationID != plan.organizationID || row.UserID != plan.userID || row.MailDomainID != domain.id || row.EngineProvisioningKey != plan.provisioningKey || row.LocalPart != plan.localPart || row.EngineAccountID == nil || *row.EngineAccountID != engineAccountID || !row.DesiredEnabled || row.Status != mailaccount.Status("active") || row.ControlGeneration != devMailControlGeneration || row.ControlHashVersion != devMailControlHashVersion || !bytes.Equal(row.ControlHash, plan.controlHash[:]) || row.DeployedGeneration == nil || *row.DeployedGeneration != devMailControlGeneration || row.DeployedHashVersion == nil || *row.DeployedHashVersion != devMailControlHashVersion || row.DeployedHash == nil || !bytes.Equal(row.DeployedHash, plan.controlHash[:]) || row.ProvisionErrorCode != nil || row.SyncErrorCode != nil {
		return false
	}
	switch row.ProjectionState {
	case mailaccount.ProjectionState("bootstrapping"):
		return row.NextSyncAt != nil && !row.NextSyncAt.After(now) &&
			row.SyncCursor == nil && row.RebuildCheckpoint == nil && row.RebuildBaseCursor == nil && row.LastSyncedAt == nil &&
			row.ProvisionAttempts == 0 && row.SyncAttempts == 0 &&
			row.SyncLeaseToken == nil && row.SyncLeaseUntil == nil &&
			row.SyncLeaseSyncEpoch == nil && row.SyncLeaseMutationEpoch == nil && row.SyncLeasePhase == nil &&
			row.SyncLeaseCursor == nil && row.SyncLeaseBaseCursor == nil && row.SyncLeaseCheckpoint == nil &&
			row.SyncLeaseResultHashVersion == nil && row.SyncLeaseResultHash == nil &&
			row.SyncLeaseAppliedPhase == nil && row.SyncLeaseAppliedNextAt == nil && row.SyncLeaseAppliedAt == nil
	case mailaccount.ProjectionState("ready"):
		return row.NextSyncAt != nil && row.SyncAttempts == 0 &&
			row.SyncCursor != nil && *row.SyncCursor != "" && row.LastSyncedAt != nil &&
			row.RebuildCheckpoint == nil && row.RebuildBaseCursor == nil
	default:
		return false
	}
}

func validDevMailControlIdentityRow(row *appent.MailIdentity, accountID uuid.UUID, plan devMailControlAccountPlan, engineIdentityID string) bool {
	return row != nil &&
		row.ID == plan.identityID &&
		row.MailAccountID == accountID &&
		row.NormalizedAddress == plan.email &&
		row.Kind == mailidentity.Kind("primary") &&
		row.Status == mailidentity.Status("active") &&
		row.SendAllowed &&
		row.EngineIdentityID != nil && *row.EngineIdentityID == engineIdentityID
}

func validDevMailControlSettingsRow(row *appent.MailUserSettings, accountID, identityID, settingsID uuid.UUID) bool {
	return row != nil &&
		row.ID == settingsID &&
		row.MailAccountID == accountID &&
		row.DefaultIdentityID == identityID &&
		row.DefaultSignatureID == nil &&
		row.RemoteImages == mailusersettings.RemoteImages("blocked") &&
		row.ComposeFormat == mailusersettings.ComposeFormat("rich")
}

func validDevMailControlPolicyRow(row *appent.MailboxPolicy, plan devMailControlPolicyPlan) bool {
	return row != nil &&
		row.ID == plan.id &&
		row.OrganizationID == plan.organizationID &&
		row.Version >= 1 &&
		!row.ExternalMailEnabled &&
		row.MaxMessageBytes == devMailControlMaxMessageBytes &&
		row.MaxAttachmentBytes == devMailControlMaxAttachmentBytes &&
		row.MaxRecipients == devMailControlMaxRecipients &&
		row.SendRatePerMinute == devMailControlSendRatePerMinute &&
		row.InstanceSendRatePerMinute == devMailControlInstanceSendRateMinute &&
		row.TrashRetentionDays == devMailControlTrashRetentionDays &&
		row.JunkRetentionDays == devMailControlJunkRetentionDays
}

func devMailControlTransactionFailure(stage devMailControlStage, err error) error {
	if appent.IsConstraintError(err) {
		return newDevMailControlError(stage, devMailControlFailureConflict)
	}
	return newDevMailControlError(stage, devMailControlFailureUnavailable)
}
