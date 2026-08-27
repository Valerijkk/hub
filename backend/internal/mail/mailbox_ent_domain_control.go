package mail

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"entgo.io/ent"
	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const mailDomainControlReplayLifetime = 30 * 24 * time.Hour

var _ MailDomainControlRepository = (*MailboxEntRepository)(nil)

// CreateMailDomain records the complete pending desired state together with its
// idempotent 202 receipt. It intentionally performs no provider or DNS I/O.
func (repository *MailboxEntRepository) CreateMailDomain(ctx context.Context, request mailDomainControlCreateRepositoryRequest) (MailDomainControlCreateResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !validMailDomainControlCreateRepositoryRequest(request) {
		return MailDomainControlCreateResult{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}

	var result MailDomainControlCreateResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *appent.Tx) error {
		organizationID, err := mailDomainControlResolveActorOrganizationForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailDomainControlEntError(err)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailDomainControlEntError(err)
		}

		receipts, err := transaction.MailMutationReceipt.Query().
			Where(
				mailmutationreceipt.OrganizationIDEQ(organizationID),
				mailmutationreceipt.ActorUserIDEQ(request.ActorID),
				mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(mailDomainControlOperation)),
				mailmutationreceipt.IdempotencyKeyHashEQ(request.IdempotencyKeyHash[:]...),
			).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if len(receipts) > 1 {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if len(receipts) == 1 {
			return mailDomainControlReplayCreate(ctx, transaction, databaseNow, organizationID, request, receipts[0], &result)
		}
		return mailDomainControlCreateFresh(ctx, transaction, databaseNow, organizationID, request, &result)
	})
	if err != nil {
		return MailDomainControlCreateResult{}, mailDomainControlEntError(err)
	}
	return mailDomainControlCreateResultCopy(result), nil
}

// ListMailDomains returns only bounded safe fields from the actor organization.
func (repository *MailboxEntRepository) ListMailDomains(ctx context.Context, request MailDomainControlListRequest) ([]MailDomainControlView, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.ActorID == uuid.Nil {
		return nil, mailDomainControlError(MailDomainControlErrorInvalid)
	}

	var result []MailDomainControlView
	err := repository.withMailboxEntTransaction(ctx, func(transaction *appent.Tx) error {
		organizationID, err := mailDomainControlResolveActorOrganizationForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailDomainControlEntError(err)
		}
		if _, err := mailboxEntDatabaseWallClock(ctx, transaction); err != nil {
			return mailDomainControlEntError(err)
		}
		rows, err := transaction.MailDomain.Query().
			Where(
				maildomain.OrganizationIDEQ(organizationID),
				maildomain.DeletedAtIsNil(),
			).
			Order(appent.Asc(maildomain.FieldDomain), appent.Asc(maildomain.FieldID)).
			Limit(mailDomainControlMaximumList + 1).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if len(rows) > mailDomainControlMaximumList {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		result = make([]MailDomainControlView, 0, len(rows))
		for _, row := range rows {
			view, ok := mailDomainControlViewFromEnt(row, organizationID)
			if !ok {
				return mailDomainControlError(MailDomainControlErrorUnavailable)
			}
			result = append(result, view)
		}
		return nil
	})
	if err != nil {
		return nil, mailDomainControlEntError(err)
	}
	return append([]MailDomainControlView(nil), result...), nil
}

// SetDefaultMailDomain changes a default only after locking the candidate and
// the prior organization default. No provider call is part of this mutation.
func (repository *MailboxEntRepository) SetDefaultMailDomain(ctx context.Context, request MailDomainControlSetDefaultRequest) (MailDomainControlView, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.ActorID == uuid.Nil || request.DomainID == uuid.Nil {
		return MailDomainControlView{}, mailDomainControlError(MailDomainControlErrorInvalid)
	}

	var result MailDomainControlView
	err := repository.withMailboxEntTransaction(ctx, func(transaction *appent.Tx) error {
		organizationID, err := mailDomainControlResolveActorOrganizationForUpdate(ctx, transaction, request.ActorID)
		if err != nil {
			return mailDomainControlEntError(err)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailDomainControlEntError(err)
		}

		candidates, err := transaction.MailDomain.Query().
			Where(
				maildomain.IDEQ(request.DomainID),
				maildomain.OrganizationIDEQ(organizationID),
				maildomain.DesiredPresentEQ(true),
				maildomain.EnabledEQ(true),
				maildomain.StateEQ(maildomain.State("verified")),
				maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
				maildomain.DKIMStatusEQ(maildomain.DKIMStatus("valid")),
				maildomain.DeletedAtIsNil(),
			).
			ForUpdate().
			All(ctx)
		if err != nil {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if len(candidates) != 1 || !mailDomainControlVerifiedEntRowValid(candidates[0], organizationID) {
			return mailDomainControlError(MailDomainControlErrorConflict)
		}
		candidate := candidates[0]

		defaults, err := transaction.MailDomain.Query().
			Where(
				maildomain.OrganizationIDEQ(organizationID),
				maildomain.IsDefaultEQ(true),
				maildomain.DeletedAtIsNil(),
			).
			Order(appent.Asc(maildomain.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil || len(defaults) > 1 {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if len(defaults) == 1 && !mailDomainControlVerifiedEntRowValid(defaults[0], organizationID) {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}

		if candidate.IsDefault {
			if len(defaults) != 1 || defaults[0].ID != candidate.ID {
				return mailDomainControlError(MailDomainControlErrorUnavailable)
			}
			view, ok := mailDomainControlViewFromEnt(candidate, organizationID)
			if !ok {
				return mailDomainControlError(MailDomainControlErrorUnavailable)
			}
			result = view
			return nil
		}

		if len(defaults) == 1 {
			updated, err := transaction.MailDomain.Update().
				Where(
					maildomain.IDEQ(defaults[0].ID),
					maildomain.OrganizationIDEQ(organizationID),
					maildomain.IsDefaultEQ(true),
					maildomain.DesiredPresentEQ(true),
					maildomain.EnabledEQ(true),
					maildomain.StateEQ(maildomain.State("verified")),
					maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
					maildomain.DKIMStatusEQ(maildomain.DKIMStatus("valid")),
					maildomain.DeletedAtIsNil(),
				).
				SetIsDefault(false).
				SetUpdatedAt(databaseNow).
				Save(ctx)
			if err != nil {
				return mailDomainControlError(MailDomainControlErrorUnavailable)
			}
			if updated != 1 {
				return mailDomainControlError(MailDomainControlErrorConflict)
			}
		}

		updated, err := transaction.MailDomain.Update().
			Where(
				maildomain.IDEQ(candidate.ID),
				maildomain.OrganizationIDEQ(organizationID),
				maildomain.IsDefaultEQ(false),
				maildomain.DesiredPresentEQ(true),
				maildomain.EnabledEQ(true),
				maildomain.StateEQ(maildomain.State("verified")),
				maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
				maildomain.DKIMStatusEQ(maildomain.DKIMStatus("valid")),
				maildomain.DeletedAtIsNil(),
			).
			SetIsDefault(true).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		if updated != 1 {
			return mailDomainControlError(MailDomainControlErrorConflict)
		}
		view, ok := mailDomainControlViewFromEnt(candidate, organizationID)
		if !ok || view.IsDefault {
			return mailDomainControlError(MailDomainControlErrorUnavailable)
		}
		view.IsDefault = true
		result = view
		return nil
	})
	if err != nil {
		return MailDomainControlView{}, mailDomainControlEntError(err)
	}
	return mailDomainControlViewCopy(result), nil
}

func mailDomainControlResolveActorOrganizationForUpdate(ctx context.Context, transaction *appent.Tx, actorID uuid.UUID) (uuid.UUID, error) {
	if ctx == nil || transaction == nil || actorID == uuid.Nil {
		return uuid.Nil, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	actor, err := transaction.User.Query().
		Where(user.IDEQ(actorID), user.StatusEQ(user.StatusActive)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return uuid.Nil, mailDomainControlError(MailDomainControlErrorNotFound)
		}
		return uuid.Nil, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	organizations, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actor.ID), user.StatusEQ(user.StatusActive))).
		Order(appent.Asc(organization.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return uuid.Nil, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	if len(organizations) != 1 || organizations[0] == nil || organizations[0].ID == uuid.Nil {
		return uuid.Nil, mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	return organizations[0].ID, nil
}

func mailDomainControlReplayCreate(ctx context.Context, transaction *appent.Tx, databaseNow time.Time, organizationID uuid.UUID, request mailDomainControlCreateRepositoryRequest, receipt *appent.MailMutationReceipt, result *MailDomainControlCreateResult) error {
	if result == nil {
		return mailDomainControlError(MailDomainControlErrorInvalid)
	}
	if kind := mailDomainControlReceiptReplayKind(receipt, databaseNow, organizationID, request); kind != "" {
		return mailDomainControlError(kind)
	}
	domain, err := transaction.MailDomain.Query().
		Where(maildomain.IDEQ(*receipt.TargetID), maildomain.OrganizationIDEQ(organizationID), maildomain.DeletedAtIsNil()).
		ForUpdate().
		Only(ctx)
	if err != nil || !mailDomainControlReplayTargetValid(domain, organizationID, request) {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	view, ok := mailDomainControlViewFromEnt(domain, organizationID)
	if !ok {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	*result = MailDomainControlCreateResult{Domain: view, Replayed: true}
	return nil
}

func mailDomainControlCreateFresh(ctx context.Context, transaction *appent.Tx, databaseNow time.Time, organizationID uuid.UUID, request mailDomainControlCreateRepositoryRequest, result *MailDomainControlCreateResult) error {
	if result == nil || organizationID == uuid.Nil || !mailboxEntValidTime(databaseNow) || !validMailDomainControlCreateRepositoryRequest(request) {
		return mailDomainControlError(MailDomainControlErrorInvalid)
	}
	existing, err := transaction.MailDomain.Query().
		Where(maildomain.DomainEQ(request.CanonicalDomain), maildomain.DeletedAtIsNil()).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	if len(existing) != 0 {
		return mailDomainControlError(MailDomainControlErrorConflict)
	}
	replayUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, mailDomainControlReplayLifetime)
	if !ok {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	receipt, err := transaction.MailMutationReceipt.Create().
		SetOrganizationID(organizationID).
		SetActorUserID(request.ActorID).
		SetOperation(mailmutationreceipt.Operation(mailDomainControlOperation)).
		SetKeySource(mailmutationreceipt.KeySource("client")).
		SetState(mailmutationreceipt.State("completed")).
		SetIdempotencyKeyHash(append([]byte(nil), request.IdempotencyKeyHash[:]...)).
		SetRequestHash(append([]byte(nil), request.RequestHash[:]...)).
		SetFingerprintVersion(request.FingerprintVersion).
		SetTargetKind(mailmutationreceipt.TargetKind("domain")).
		SetTargetID(request.DomainID).
		SetOriginalHTTPStatus(202).
		SetTerminalAt(databaseNow).
		SetReplayUntil(replayUntil).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || mailDomainControlReceiptReplayKind(receipt, databaseNow, organizationID, request) != "" {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	domain, err := transaction.MailDomain.Create().
		SetID(request.DomainID).
		SetOrganizationID(organizationID).
		SetDomain(request.CanonicalDomain).
		SetEngineProvisioningKey(request.ProvisioningKey).
		SetDesiredPresent(true).
		SetEnabled(true).
		SetIsDefault(false).
		SetState(maildomain.State("pending")).
		SetVerificationStatus(maildomain.VerificationStatus("pending")).
		SetDKIMStatus(maildomain.DKIMStatus("unknown")).
		SetRequiredDNSRecords(RequiredDNSRecords{}).
		SetDesiredGeneration(1).
		SetDesiredHashVersion(mailDomainControlFingerprintVersion).
		SetDesiredHash(append([]byte(nil), request.DesiredHash[:]...)).
		SetAttemptCount(0).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || !mailDomainControlPendingReplayRowValid(domain, organizationID, request) {
		if ent.IsConstraintError(err) {
			return mailDomainControlError(MailDomainControlErrorConflict)
		}
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	view, ok := mailDomainControlViewFromEnt(domain, organizationID)
	if !ok {
		return mailDomainControlError(MailDomainControlErrorUnavailable)
	}
	*result = MailDomainControlCreateResult{Domain: view}
	return nil
}

func mailDomainControlReceiptReplayKind(receipt *appent.MailMutationReceipt, databaseNow time.Time, organizationID uuid.UUID, request mailDomainControlCreateRepositoryRequest) MailDomainControlErrorKind {
	if receipt == nil || organizationID == uuid.Nil || !mailboxEntValidTime(databaseNow) || receipt.ID == uuid.Nil || receipt.OrganizationID != organizationID || receipt.ActorUserID != request.ActorID || receipt.MailAccountID != nil || receipt.Operation != mailmutationreceipt.Operation(mailDomainControlOperation) || receipt.KeySource != mailmutationreceipt.KeySource("client") || !bytes.Equal(receipt.IdempotencyKeyHash, request.IdempotencyKeyHash[:]) {
		return MailDomainControlErrorUnavailable
	}
	if !bytes.Equal(receipt.RequestHash, request.RequestHash[:]) {
		return MailDomainControlErrorConflict
	}
	if receipt.FingerprintVersion != request.FingerprintVersion || receipt.State != mailmutationreceipt.State("completed") || receipt.TargetKind == nil || *receipt.TargetKind != mailmutationreceipt.TargetKind("domain") || receipt.TargetID == nil || *receipt.TargetID != request.DomainID || receipt.OriginalHTTPStatus == nil || *receipt.OriginalHTTPStatus != 202 || receipt.SafeCode != nil || receipt.BulkAction != nil || receipt.ArgumentKind != nil || receipt.ArgumentID != nil || receipt.ReconcileAttempts != 0 || receipt.NextReconcileAt != nil || receipt.ReconcileLeaseToken != nil || receipt.ReconcileLeaseUntil != nil || receipt.TerminalAt == nil || receipt.ReplayUntil == nil || receipt.CompactedAt != nil || !mailboxEntValidTime(*receipt.TerminalAt) || !mailboxEntValidTime(*receipt.ReplayUntil) {
		return MailDomainControlErrorUnavailable
	}
	if !databaseNow.Before(*receipt.ReplayUntil) {
		return MailDomainControlErrorConflict
	}
	expectedReplayUntil, ok := mailboxEntCheckedTimeAdd(*receipt.TerminalAt, mailDomainControlReplayLifetime)
	if !ok || !receipt.ReplayUntil.Equal(expectedReplayUntil) || !mailboxEntValidTime(receipt.CreatedAt) || !mailboxEntValidTime(receipt.UpdatedAt) {
		return MailDomainControlErrorUnavailable
	}
	return ""
}

func mailDomainControlPendingReplayRowValid(domain *appent.MailDomain, organizationID uuid.UUID, request mailDomainControlCreateRepositoryRequest) bool {
	if !mailDomainControlEntRowValid(domain, organizationID) || domain.ID != request.DomainID || domain.Domain != request.CanonicalDomain || domain.EngineProvisioningKey != request.ProvisioningKey || domain.EngineDomainID != nil || !domain.DesiredPresent || !domain.Enabled || domain.IsDefault || domain.State != maildomain.State("pending") || domain.VerificationStatus != maildomain.VerificationStatus("pending") || domain.DKIMStatus != maildomain.DKIMStatus("unknown") || len(domain.RequiredDNSRecords) != 0 || domain.DesiredGeneration != 1 || domain.DesiredHashVersion != request.FingerprintVersion || subtle.ConstantTimeCompare(domain.DesiredHash, request.DesiredHash[:]) != 1 || domain.AttemptedGeneration != nil || domain.AttemptedHashVersion != nil || domain.AttemptedHash != nil || domain.DeployedGeneration != nil || domain.DeployedHashVersion != nil || domain.DeployedHash != nil || domain.LastErrorCode != nil || domain.NextRetryAt != nil || domain.AttemptCount != 0 || domain.LeaseToken != nil || domain.LeaseUntil != nil || domain.DeletedAt != nil {
		return false
	}
	return domain.CreatedAt.Equal(domain.UpdatedAt)
}

// mailDomainControlReplayTargetValid keeps receipt replay read-only while
// allowing a legitimately advanced target row to be returned as its safe view.
func mailDomainControlReplayTargetValid(domain *appent.MailDomain, organizationID uuid.UUID, request mailDomainControlCreateRepositoryRequest) bool {
	return mailDomainControlEntRowValid(domain, organizationID) && domain.ID == request.DomainID && domain.Domain == request.CanonicalDomain && domain.EngineProvisioningKey == request.ProvisioningKey && domain.DeletedAt == nil
}

func mailDomainControlVerifiedEntRowValid(domain *appent.MailDomain, organizationID uuid.UUID) bool {
	if !mailDomainControlEntRowValid(domain, organizationID) || !domain.DesiredPresent || !domain.Enabled || domain.State != maildomain.State("verified") || domain.VerificationStatus != maildomain.VerificationStatus("verified") || domain.DKIMStatus != maildomain.DKIMStatus("valid") || domain.DeletedAt != nil || domain.EngineDomainID == nil || !validMailboxOAuthOpaque(*domain.EngineDomainID, MaxEngineOpaqueIDBytes) || domain.DeployedGeneration == nil || domain.DeployedHashVersion == nil || domain.DeployedHash == nil || *domain.DeployedGeneration != domain.DesiredGeneration || *domain.DeployedHashVersion != domain.DesiredHashVersion || subtle.ConstantTimeCompare(domain.DeployedHash, domain.DesiredHash) != 1 || domain.AttemptedGeneration != nil || domain.AttemptedHashVersion != nil || domain.AttemptedHash != nil || domain.LastErrorCode != nil || domain.NextRetryAt != nil || domain.LeaseToken != nil || domain.LeaseUntil != nil {
		return false
	}
	return true
}

func mailDomainControlEntRowValid(domain *appent.MailDomain, organizationID uuid.UUID) bool {
	if domain == nil || organizationID == uuid.Nil || domain.ID == uuid.Nil || domain.OrganizationID != organizationID || !mailboxEntValidTime(domain.CreatedAt) || !mailboxEntValidTime(domain.UpdatedAt) || domain.CreatedAt.After(domain.UpdatedAt) || domain.DesiredGeneration < 1 || domain.DesiredHashVersion < 1 || !validMailDomainControlDigest(domain.DesiredHash) || domain.AttemptCount < 0 || domain.RequiredDNSRecords == nil || domain.RequiredDNSRecords.Validate() != nil {
		return false
	}
	canonicalDomain, ok := canonicalMailDomainControlDomain(domain.Domain)
	if !ok || canonicalDomain != domain.Domain || domain.EngineProvisioningKey == "" || !validMailDomainControlState(string(domain.State)) || !validMailDomainControlVerificationStatus(string(domain.VerificationStatus)) || !validMailDomainControlDKIMStatus(string(domain.DKIMStatus)) || !mailDomainControlTupleValid(domain.AttemptedGeneration, domain.AttemptedHashVersion, domain.AttemptedHash, domain.DesiredGeneration) || !mailDomainControlTupleValid(domain.DeployedGeneration, domain.DeployedHashVersion, domain.DeployedHash, domain.DesiredGeneration) || (domain.LeaseToken == nil) != (domain.LeaseUntil == nil) {
		return false
	}
	if domain.EngineDomainID != nil && !validMailboxOAuthOpaque(*domain.EngineDomainID, MaxEngineOpaqueIDBytes) {
		return false
	}
	if domain.NextRetryAt != nil && !mailboxEntValidTime(*domain.NextRetryAt) || domain.LeaseUntil != nil && (!mailboxEntValidTime(*domain.LeaseUntil) || *domain.LeaseToken == uuid.Nil) || domain.DeletedAt != nil && !mailboxEntValidTime(*domain.DeletedAt) {
		return false
	}
	return domain.LastErrorCode == nil || mailboxSetupStatusDomainErrorCodeValid(*domain.LastErrorCode)
}

func mailDomainControlTupleValid(generation *int64, version *int32, hash []byte, maximumGeneration int64) bool {
	if (generation == nil) != (version == nil) || (generation == nil) != (hash == nil) {
		return false
	}
	return generation == nil || *generation >= 1 && *generation <= maximumGeneration && *version >= 1 && validMailDomainControlDigest(hash)
}

func mailDomainControlViewFromEnt(domain *appent.MailDomain, organizationID uuid.UUID) (MailDomainControlView, bool) {
	if !mailDomainControlEntRowValid(domain, organizationID) || domain.DeletedAt != nil {
		return MailDomainControlView{}, false
	}
	view := MailDomainControlView{
		ID:                 domain.ID,
		Domain:             domain.Domain,
		State:              string(domain.State),
		VerificationStatus: string(domain.VerificationStatus),
		DKIMStatus:         string(domain.DKIMStatus),
		RequiredDNSRecords: append(RequiredDNSRecords(nil), domain.RequiredDNSRecords...),
		IsDefault:          domain.IsDefault,
	}
	return view, validMailDomainControlView(view)
}

func mailDomainControlEntError(err error) error {
	var controlError *MailDomainControlError
	if errors.As(err, &controlError) && controlError != nil {
		return mailDomainControlError(controlError.Kind)
	}
	var serviceError *MailboxServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		switch serviceError.Kind {
		case MailboxServiceErrorInvalid:
			return mailDomainControlError(MailDomainControlErrorInvalid)
		case MailboxServiceErrorNotFound:
			return mailDomainControlError(MailDomainControlErrorNotFound)
		}
	}
	return mailDomainControlError(MailDomainControlErrorUnavailable)
}
