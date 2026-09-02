package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ MailboxDomainProvisioningRepository = (*MailboxEntRepository)(nil)

// ScanMailboxDomainProvisioning returns a bounded, deterministic list of
// currently due desired domain rows. It performs no provider I/O.
func (repository *MailboxEntRepository) ScanMailboxDomainProvisioning(ctx context.Context, request MailboxDomainProvisioningScanRequest) (MailboxDomainProvisioningScanResult, error) {
	request, ok := NormalizeMailboxDomainProvisioningScanRequest(request)
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !ok {
		return MailboxDomainProvisioningScanResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	result := MailboxDomainProvisioningScanResult{DomainIDs: make([]uuid.UUID, 0, request.BatchSize)}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		domainIDs, err := mailboxEntDomainProvisioningScanDue(ctx, transaction, databaseNow, request.BatchSize)
		if err != nil {
			return err
		}
		result.DomainIDs = append(result.DomainIDs, domainIDs...)
		if !result.ValidFor(request) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		return nil
	})
	if err != nil {
		return MailboxDomainProvisioningScanResult{}, mailboxEntSafeError(err)
	}
	if !result.ValidFor(request) {
		return MailboxDomainProvisioningScanResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return MailboxDomainProvisioningScanResult{
		DomainIDs: append(make([]uuid.UUID, 0, len(result.DomainIDs)), result.DomainIDs...),
	}, nil
}

func mailboxEntDomainProvisioningScanDue(ctx context.Context, transaction *ent.Tx, databaseNow time.Time, limit int) ([]uuid.UUID, error) {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(databaseNow) || limit < 1 || limit > MailboxDomainProvisioningMaximumBatchSize {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	active := []predicate.MailDomain{
		maildomain.DesiredPresentEQ(true),
		maildomain.EnabledEQ(true),
		maildomain.DeletedAtIsNil(),
	}
	pending, err := transaction.MailDomain.Query().
		Where(append(active,
			maildomain.StateEQ(maildomain.State("pending")),
		)...).
		Order(ent.Asc(maildomain.FieldUpdatedAt), ent.Asc(maildomain.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	dnsRequired, err := transaction.MailDomain.Query().
		Where(append(active,
			maildomain.StateEQ(maildomain.State("dns_required")),
		)...).
		Order(ent.Asc(maildomain.FieldUpdatedAt), ent.Asc(maildomain.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	retrying, err := transaction.MailDomain.Query().
		Where(append(active,
			maildomain.StateEQ(maildomain.State("error")),
			maildomain.LastErrorCodeIn(
				maildomain.LastErrorCode(MailboxDomainProvisioningCodeEngineUnavailable),
				maildomain.LastErrorCode(MailboxDomainProvisioningCodeOutcomeUnknown),
			),
			maildomain.NextRetryAtNotNil(),
			maildomain.NextRetryAtLTE(databaseNow),
		)...).
		Order(ent.Asc(maildomain.FieldNextRetryAt), ent.Asc(maildomain.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	expired, err := transaction.MailDomain.Query().
		Where(append(active,
			maildomain.StateEQ(maildomain.State("provisioning")),
			maildomain.LeaseTokenNotNil(),
			maildomain.LeaseUntilNotNil(),
			maildomain.LeaseUntilLTE(databaseNow),
		)...).
		Order(ent.Asc(maildomain.FieldLeaseUntil), ent.Asc(maildomain.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	type candidate struct {
		id    uuid.UUID
		dueAt time.Time
	}
	candidates := make([]candidate, 0, len(pending)+len(dnsRequired)+len(retrying)+len(expired))
	appendCandidates := func(rows []*ent.MailDomain) error {
		for _, row := range rows {
			dueAt, _, due := mailboxEntDomainProvisioningDue(row, databaseNow)
			if !due || row == nil || row.ID == uuid.Nil {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			candidates = append(candidates, candidate{id: row.ID, dueAt: dueAt})
		}
		return nil
	}
	if err := appendCandidates(pending); err != nil {
		return nil, err
	}
	if err := appendCandidates(dnsRequired); err != nil {
		return nil, err
	}
	if err := appendCandidates(retrying); err != nil {
		return nil, err
	}
	if err := appendCandidates(expired); err != nil {
		return nil, err
	}

	sort.SliceStable(candidates, func(left, right int) bool {
		if candidates[left].dueAt.Equal(candidates[right].dueAt) {
			return candidates[left].id.String() < candidates[right].id.String()
		}
		return candidates[left].dueAt.Before(candidates[right].dueAt)
	})

	domainIDs := make([]uuid.UUID, 0, limit)
	seen := make(map[uuid.UUID]struct{}, limit)
	for _, candidate := range candidates {
		if len(domainIDs) == limit {
			break
		}
		if _, duplicate := seen[candidate.id]; duplicate {
			continue
		}
		seen[candidate.id] = struct{}{}
		domainIDs = append(domainIDs, candidate.id)
	}
	return domainIDs, nil
}

// ClaimMailboxDomainProvisioning seals exactly one provider-free domain fence.
func (repository *MailboxEntRepository) ClaimMailboxDomainProvisioning(ctx context.Context, request MailboxDomainProvisioningClaimRequest) (MailboxDomainProvisioningClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !request.Valid() {
		return MailboxDomainProvisioningClaimResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	result := MailboxDomainProvisioningClaimResult{Outcome: MailboxDomainProvisioningNoop}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		domain, err := transaction.MailDomain.Query().
			Where(maildomain.IDEQ(request.DomainID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if domain == nil || domain.ID != request.DomainID || !mailboxEntDomainProvisioningRowValid(domain) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if mailboxEntDomainProvisioningConverged(domain) {
			return nil
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		_, action, due := mailboxEntDomainProvisioningDue(domain, databaseNow)
		if !due {
			return nil
		}
		if domain.AttemptCount < 0 || domain.AttemptCount >= math.MaxInt32 || domain.Version < 1 || domain.Version >= math.MaxInt32 {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		leaseDeadline, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok || !mailboxDomainProvisioningCanonicalSecond(leaseDeadline) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		leaseToken := uuid.New()
		if leaseToken == uuid.Nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		predicates, ok := mailboxEntDomainProvisioningCASPredicates(domain)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		updated, err := transaction.MailDomain.Update().
			Where(predicates...).
			SetState(maildomain.State("provisioning")).
			SetAttemptedGeneration(domain.DesiredGeneration).
			SetAttemptedHashVersion(domain.DesiredHashVersion).
			SetAttemptedHash(append([]byte(nil), domain.DesiredHash...)).
			SetLeaseToken(leaseToken).
			SetLeaseUntil(leaseDeadline).
			SetAttemptCount(domain.AttemptCount + 1).
			ClearNextRetryAt().
			ClearLastErrorCode().
			SetVersion(domain.Version + 1).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil || updated != 1 {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		claim, ok := NewMailboxDomainProvisioningClaim(MailboxDomainProvisioningFence{
			DomainID:           domain.ID,
			OrganizationID:     domain.OrganizationID,
			LeaseToken:         leaseToken,
			LeaseDeadline:      leaseDeadline,
			LeaseDuration:      MailboxEngineCommandLeaseDuration,
			DesiredGeneration:  domain.DesiredGeneration,
			DesiredHashVersion: domain.DesiredHashVersion,
			DesiredHash:        hex.EncodeToString(domain.DesiredHash),
			ProvisioningKey:    domain.EngineProvisioningKey,
			Domain:             domain.Domain,
			DesiredPresent:     true,
			DesiredEnabled:     true,
			Action:             action,
		})
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		result = MailboxDomainProvisioningClaimResult{
			Outcome: MailboxDomainProvisioningClaimed,
			Claim:   claim,
		}
		return nil
	})
	if err != nil {
		return MailboxDomainProvisioningClaimResult{}, mailboxEntSafeError(err)
	}
	if !result.Valid() {
		return MailboxDomainProvisioningClaimResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return result, nil
}

// FinalizeMailboxDomainProvisioning records only an already sealed provider
// observation or retry result. It never calls the provider.
func (repository *MailboxEntRepository) FinalizeMailboxDomainProvisioning(ctx context.Context, request MailboxDomainProvisioningFinalizeRequest) (MailboxDomainProvisioningFinalizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !request.Valid() {
		return MailboxDomainProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	fence := request.Claim.Fence()
	desiredHash, ok := mailboxEntDomainProvisioningFenceHash(fence)
	if !ok {
		return MailboxDomainProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxDomainProvisioningFinalizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		domain, err := transaction.MailDomain.Query().
			Where(maildomain.IDEQ(fence.DomainID)).
			ForUpdate().
			Only(ctx)
		if err != nil || domain == nil || domain.ID != fence.DomainID {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if !mailboxEntDomainProvisioningRowValid(domain) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		if !mailboxEntDomainProvisioningLiveFenceMatches(domain, fence, desiredHash, databaseNow) {
			if !mailboxEntDomainProvisioningSuperseded(domain, fence, desiredHash, databaseNow) {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			predicates, ok := mailboxEntDomainProvisioningCASPredicates(domain)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			updated, err := transaction.MailDomain.Update().
				Where(predicates...).
				SetState(maildomain.State("pending")).
				ClearAttemptedGeneration().
				ClearAttemptedHashVersion().
				ClearAttemptedHash().
				ClearLeaseToken().
				ClearLeaseUntil().
				ClearLastErrorCode().
				ClearNextRetryAt().
				SetVersion(domain.Version + 1).
				SetUpdatedAt(databaseNow).
				Save(ctx)
			if err != nil || updated != 1 {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			result = MailboxDomainProvisioningFinalizeResult{Outcome: MailboxDomainProvisioningFinalizeSuperseded}
			return nil
		}

		predicates, ok := mailboxEntDomainProvisioningCASPredicates(domain)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		switch request.Outcome {
		case MailboxDomainProvisioningFinalizeDNSRequired, MailboxDomainProvisioningFinalizeVerified:
			observed, ok := request.Observed()
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			verification, ok := mailboxEntDomainProvisioningVerificationStatus(observed.Verification)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			dkim, ok := mailboxEntDomainProvisioningDKIMStatus(observed.DKIM)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}

			var state maildomain.State
			switch request.Outcome {
			case MailboxDomainProvisioningFinalizeDNSRequired:
				if observed.State != EngineDomainDNSRequired {
					return mailboxServiceError(MailboxServiceErrorInvalid)
				}
				state = maildomain.State("dns_required")
			case MailboxDomainProvisioningFinalizeVerified:
				if observed.State != EngineDomainVerified {
					return mailboxServiceError(MailboxServiceErrorInvalid)
				}
				state = maildomain.State("verified")
			}
			updated, err := transaction.MailDomain.Update().
				Where(predicates...).
				SetState(state).
				SetEngineDomainID(observed.EngineDomainID).
				SetRequiredDNSRecords(append(RequiredDNSRecords(nil), observed.RequiredDNSRecords...)).
				SetVerificationStatus(verification).
				SetDkimStatus(dkim).
				SetDeployedGeneration(fence.DesiredGeneration).
				SetDeployedHashVersion(fence.DesiredHashVersion).
				SetDeployedHash(append([]byte(nil), desiredHash...)).
				ClearAttemptedGeneration().
				ClearAttemptedHashVersion().
				ClearAttemptedHash().
				ClearLeaseToken().
				ClearLeaseUntil().
				ClearLastErrorCode().
				ClearNextRetryAt().
				SetVersion(domain.Version + 1).
				SetUpdatedAt(databaseNow).
				Save(ctx)
			if err != nil || updated != 1 {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			result = MailboxDomainProvisioningFinalizeResult{Outcome: request.Outcome}

		case MailboxDomainProvisioningFinalizeRetry:
			delay, ok := MailboxDomainProvisioningRetryDelay(int(domain.AttemptCount))
			if !ok {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, delay)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			update := transaction.MailDomain.Update().
				Where(predicates...).
				SetState(maildomain.State("error")).
				ClearLeaseToken().
				ClearLeaseUntil().
				SetLastErrorCode(maildomain.LastErrorCode(request.Code)).
				SetVersion(domain.Version + 1).
				SetUpdatedAt(databaseNow)

			switch request.Code {
			case MailboxDomainProvisioningCodeOutcomeUnknown:
				update.SetNextRetryAt(retryAt)
			case MailboxDomainProvisioningCodeEngineUnavailable:
				update.
					ClearAttemptedGeneration().
					ClearAttemptedHashVersion().
					ClearAttemptedHash().
					SetNextRetryAt(retryAt)
			case MailboxDomainProvisioningCodeApplyRejected, MailboxDomainProvisioningCodeDNSInvalid:
				update.
					ClearAttemptedGeneration().
					ClearAttemptedHashVersion().
					ClearAttemptedHash().
					ClearNextRetryAt()
			default:
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			updated, err := update.Save(ctx)
			if err != nil || updated != 1 {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			result = MailboxDomainProvisioningFinalizeResult{
				Outcome: MailboxDomainProvisioningFinalizeRetry,
				Code:    request.Code,
			}
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return nil
	})
	if err != nil {
		return MailboxDomainProvisioningFinalizeResult{}, mailboxEntSafeError(err)
	}
	if !result.Valid() {
		return MailboxDomainProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return result, nil
}

func mailboxEntDomainProvisioningDue(domain *ent.MailDomain, databaseNow time.Time) (time.Time, MailboxDomainProvisioningAction, bool) {
	if !mailboxEntDomainProvisioningRowValid(domain) || !mailboxEntValidTime(databaseNow) || !domain.DesiredPresent || !domain.Enabled || domain.DeletedAt != nil {
		return time.Time{}, "", false
	}
	switch domain.State {
	case maildomain.State("pending"):
		return domain.UpdatedAt, MailboxDomainProvisioningActionEnsure, true
	case maildomain.State("dns_required"):
		return domain.UpdatedAt, MailboxDomainProvisioningActionReconcileThenEnsure, true
	case maildomain.State("error"):
		if domain.NextRetryAt == nil || domain.NextRetryAt.After(databaseNow) || domain.LastErrorCode == nil {
			return time.Time{}, "", false
		}
		switch MailboxDomainProvisioningCode(*domain.LastErrorCode) {
		case MailboxDomainProvisioningCodeEngineUnavailable:
			return *domain.NextRetryAt, MailboxDomainProvisioningActionEnsure, true
		case MailboxDomainProvisioningCodeOutcomeUnknown:
			return *domain.NextRetryAt, MailboxDomainProvisioningActionReconcileThenEnsure, true
		default:
			return time.Time{}, "", false
		}
	case maildomain.State("provisioning"):
		if domain.LeaseUntil == nil || domain.LeaseUntil.After(databaseNow) {
			return time.Time{}, "", false
		}
		return *domain.LeaseUntil, MailboxDomainProvisioningActionReconcileThenEnsure, true
	default:
		return time.Time{}, "", false
	}
}

func mailboxEntDomainProvisioningConverged(domain *ent.MailDomain) bool {
	if !mailboxEntDomainProvisioningRowValid(domain) || !domain.DesiredPresent || !domain.Enabled || domain.DeletedAt != nil || domain.State != maildomain.State("verified") || domain.EngineDomainID == nil || domain.VerificationStatus != maildomain.VerificationStatus("verified") || domain.DkimStatus != maildomain.DkimStatus("valid") {
		return false
	}
	return mailboxEntDomainProvisioningCurrentDeployedTuple(domain) &&
		domain.AttemptedGeneration == nil &&
		domain.AttemptedHashVersion == nil &&
		mailboxEntNullableBytesAbsent(domain.AttemptedHash) &&
		domain.LeaseToken == nil &&
		domain.LeaseUntil == nil &&
		domain.LastErrorCode == nil &&
		domain.NextRetryAt == nil
}

func mailboxEntDomainProvisioningFenceHash(fence MailboxDomainProvisioningFence) ([]byte, bool) {
	if !fence.Valid() || len(fence.DesiredHash) != hex.EncodedLen(sha256.Size) {
		return nil, false
	}
	decoded, err := hex.DecodeString(fence.DesiredHash)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != fence.DesiredHash {
		return nil, false
	}
	return decoded, true
}

func mailboxEntDomainProvisioningLiveFenceMatches(domain *ent.MailDomain, fence MailboxDomainProvisioningFence, desiredHash []byte, databaseNow time.Time) bool {
	if !mailboxEntDomainProvisioningRowValid(domain) || !fence.Valid() || len(desiredHash) != sha256.Size || !mailboxEntValidTime(databaseNow) {
		return false
	}
	if domain.ID != fence.DomainID ||
		domain.OrganizationID != fence.OrganizationID ||
		domain.Domain != fence.Domain ||
		domain.EngineProvisioningKey != fence.ProvisioningKey ||
		!domain.DesiredPresent ||
		!domain.Enabled ||
		domain.DeletedAt != nil ||
		domain.State != maildomain.State("provisioning") ||
		domain.DesiredGeneration != fence.DesiredGeneration ||
		domain.DesiredHashVersion != fence.DesiredHashVersion ||
		subtle.ConstantTimeCompare(domain.DesiredHash, desiredHash) != 1 ||
		domain.AttemptedGeneration == nil ||
		domain.AttemptedHashVersion == nil ||
		domain.AttemptedHash == nil ||
		*domain.AttemptedGeneration != fence.DesiredGeneration ||
		*domain.AttemptedHashVersion != fence.DesiredHashVersion ||
		subtle.ConstantTimeCompare(*domain.AttemptedHash, desiredHash) != 1 ||
		domain.LeaseToken == nil ||
		*domain.LeaseToken != fence.LeaseToken ||
		domain.LeaseUntil == nil ||
		!domain.LeaseUntil.Equal(fence.LeaseDeadline) ||
		!domain.LeaseUntil.After(databaseNow) ||
		domain.AttemptCount < 1 {
		return false
	}
	return true
}

func mailboxEntDomainProvisioningSuperseded(domain *ent.MailDomain, fence MailboxDomainProvisioningFence, desiredHash []byte, databaseNow time.Time) bool {
	if !mailboxEntDomainProvisioningRowValid(domain) || !fence.Valid() || len(desiredHash) != sha256.Size || !mailboxEntValidTime(databaseNow) {
		return false
	}
	if domain.ID != fence.DomainID ||
		domain.OrganizationID != fence.OrganizationID ||
		domain.Domain != fence.Domain ||
		domain.EngineProvisioningKey != fence.ProvisioningKey ||
		!domain.DesiredPresent ||
		!domain.Enabled ||
		domain.DeletedAt != nil ||
		domain.State != maildomain.State("provisioning") ||
		domain.DesiredGeneration <= fence.DesiredGeneration ||
		domain.AttemptedGeneration == nil ||
		domain.AttemptedHashVersion == nil ||
		domain.AttemptedHash == nil ||
		*domain.AttemptedGeneration != fence.DesiredGeneration ||
		*domain.AttemptedHashVersion != fence.DesiredHashVersion ||
		subtle.ConstantTimeCompare(*domain.AttemptedHash, desiredHash) != 1 ||
		domain.LeaseToken == nil ||
		*domain.LeaseToken != fence.LeaseToken ||
		domain.LeaseUntil == nil ||
		!domain.LeaseUntil.Equal(fence.LeaseDeadline) ||
		!domain.LeaseUntil.After(databaseNow) ||
		domain.AttemptCount < 1 {
		return false
	}
	return true
}

func mailboxEntDomainProvisioningCASPredicates(domain *ent.MailDomain) ([]predicate.MailDomain, bool) {
	if !mailboxEntDomainProvisioningRowValid(domain) {
		return nil, false
	}
	requiredDNSRecords, ok := mailboxEntDomainProvisioningRequiredDNSRecordsPredicate(domain.RequiredDNSRecords)
	if !ok {
		return nil, false
	}
	predicates := []predicate.MailDomain{
		maildomain.IDEQ(domain.ID),
		maildomain.OrganizationIDEQ(domain.OrganizationID),
		maildomain.DomainEQ(domain.Domain),
		maildomain.EngineProvisioningKeyEQ(domain.EngineProvisioningKey),
		maildomain.DesiredPresentEQ(domain.DesiredPresent),
		maildomain.EnabledEQ(domain.Enabled),
		maildomain.IsDefaultEQ(domain.IsDefault),
		maildomain.StateEQ(domain.State),
		maildomain.VerificationStatusEQ(domain.VerificationStatus),
		maildomain.DkimStatusEQ(domain.DkimStatus),
		requiredDNSRecords,
		maildomain.DesiredGenerationEQ(domain.DesiredGeneration),
		maildomain.DesiredHashVersionEQ(domain.DesiredHashVersion),
		maildomain.DesiredHashEQ(domain.DesiredHash),
		maildomain.AttemptCountEQ(domain.AttemptCount),
		maildomain.VersionEQ(domain.Version),
		maildomain.CreatedAtEQ(domain.CreatedAt),
		maildomain.UpdatedAtEQ(domain.UpdatedAt),
	}
	if domain.EngineDomainID == nil {
		predicates = append(predicates, maildomain.EngineDomainIDIsNil())
	} else {
		predicates = append(predicates, maildomain.EngineDomainIDEQ(*domain.EngineDomainID))
	}
	if domain.AttemptedGeneration == nil {
		predicates = append(predicates, maildomain.AttemptedGenerationIsNil(), maildomain.AttemptedHashVersionIsNil(), maildomain.AttemptedHashIsNil())
	} else {
		predicates = append(predicates,
			maildomain.AttemptedGenerationEQ(*domain.AttemptedGeneration),
			maildomain.AttemptedHashVersionEQ(*domain.AttemptedHashVersion),
			maildomain.AttemptedHashEQ(*domain.AttemptedHash),
		)
	}
	if domain.DeployedGeneration == nil {
		predicates = append(predicates, maildomain.DeployedGenerationIsNil(), maildomain.DeployedHashVersionIsNil(), maildomain.DeployedHashIsNil())
	} else {
		predicates = append(predicates,
			maildomain.DeployedGenerationEQ(*domain.DeployedGeneration),
			maildomain.DeployedHashVersionEQ(*domain.DeployedHashVersion),
			maildomain.DeployedHashEQ(*domain.DeployedHash),
		)
	}
	if domain.LastErrorCode == nil {
		predicates = append(predicates, maildomain.LastErrorCodeIsNil())
	} else {
		predicates = append(predicates, maildomain.LastErrorCodeEQ(*domain.LastErrorCode))
	}
	if domain.NextRetryAt == nil {
		predicates = append(predicates, maildomain.NextRetryAtIsNil())
	} else {
		predicates = append(predicates, maildomain.NextRetryAtEQ(*domain.NextRetryAt))
	}
	if domain.LeaseToken == nil {
		predicates = append(predicates, maildomain.LeaseTokenIsNil(), maildomain.LeaseUntilIsNil())
	} else {
		predicates = append(predicates, maildomain.LeaseTokenEQ(*domain.LeaseToken), maildomain.LeaseUntilEQ(*domain.LeaseUntil))
	}
	if domain.DeletedAt == nil {
		predicates = append(predicates, maildomain.DeletedAtIsNil())
	} else {
		predicates = append(predicates, maildomain.DeletedAtEQ(*domain.DeletedAt))
	}
	return predicates, true
}

// Ent does not generate an equality predicate for JSON fields. PostgreSQL
// jsonb equality is safe here because MailDomain's schema itself contains
// jsonb checks; the JSON payload is marshalled as an immutable value and bound
// as a query argument rather than interpolated.
func mailboxEntDomainProvisioningRequiredDNSRecordsPredicate(records RequiredDNSRecords) (predicate.MailDomain, bool) {
	if records == nil || records.Validate() != nil {
		return nil, false
	}
	encoded, err := json.Marshal(records)
	if err != nil || len(encoded) == 0 {
		return nil, false
	}
	return predicate.MailDomain(func(selector *entsql.Selector) {
		selector.Where(entsql.P(func(builder *entsql.Builder) {
			builder.Ident(selector.C(maildomain.FieldRequiredDNSRecords)).
				WriteString(" = ").
				Arg(string(encoded)).
				WriteString("::jsonb")
		}))
	}), true
}

func mailboxEntDomainProvisioningRowValid(domain *ent.MailDomain) bool {
	if domain == nil ||
		domain.ID == uuid.Nil ||
		domain.OrganizationID == uuid.Nil ||
		!validCanonicalRegistryDomain(domain.Domain) ||
		!validSafeText(domain.EngineProvisioningKey, MaxEngineOpaqueIDBytes, true) ||
		!mailboxEntValidTime(domain.CreatedAt) ||
		!mailboxEntValidTime(domain.UpdatedAt) ||
		domain.CreatedAt.After(domain.UpdatedAt) ||
		domain.Version < 1 ||
		domain.Version > math.MaxInt32 ||
		domain.DesiredGeneration < 1 ||
		domain.DesiredHashVersion < 1 ||
		len(domain.DesiredHash) != sha256.Size ||
		domain.AttemptCount < 0 ||
		domain.RequiredDNSRecords == nil ||
		domain.RequiredDNSRecords.Validate() != nil ||
		!mailboxEntDomainProvisioningStateValid(domain.State) ||
		!mailboxEntDomainProvisioningVerificationStatusValid(domain.VerificationStatus) ||
		!mailboxEntDomainProvisioningDKIMStatusValid(domain.DkimStatus) ||
		!mailboxEntDomainProvisioningTupleValid(domain.AttemptedGeneration, domain.AttemptedHashVersion, domain.AttemptedHash, domain.DesiredGeneration) ||
		!mailboxEntDomainProvisioningTupleValid(domain.DeployedGeneration, domain.DeployedHashVersion, domain.DeployedHash, domain.DesiredGeneration) ||
		(domain.LeaseToken == nil) != (domain.LeaseUntil == nil) {
		return false
	}
	if domain.EngineDomainID != nil && !validOpaqueID(*domain.EngineDomainID) {
		return false
	}
	if domain.NextRetryAt != nil && !mailboxEntValidTime(*domain.NextRetryAt) {
		return false
	}
	if domain.LeaseUntil != nil && (!mailboxEntValidTime(*domain.LeaseUntil) || *domain.LeaseToken == uuid.Nil) {
		return false
	}
	if domain.DeletedAt != nil && (!mailboxEntValidTime(*domain.DeletedAt) || domain.State != maildomain.State("deleting")) {
		return false
	}
	if domain.LastErrorCode != nil && !mailboxEntDomainProvisioningErrorCodeValid(*domain.LastErrorCode) {
		return false
	}
	return mailboxEntDomainProvisioningStateInvariant(domain)
}

func mailboxEntDomainProvisioningStateInvariant(domain *ent.MailDomain) bool {
	if domain == nil {
		return false
	}
	attempted := domain.AttemptedGeneration != nil
	leased := domain.LeaseToken != nil
	switch domain.State {
	case maildomain.State("pending"):
		return domain.DesiredPresent && !attempted && !leased && domain.LastErrorCode == nil && domain.NextRetryAt == nil
	case maildomain.State("provisioning"):
		return domain.DesiredPresent && attempted && leased && domain.LastErrorCode == nil && domain.NextRetryAt == nil
	case maildomain.State("dns_required"):
		return domain.DesiredPresent &&
			domain.EngineDomainID != nil &&
			mailboxEntDomainProvisioningCurrentDeployedTuple(domain) &&
			len(domain.RequiredDNSRecords) > 0 &&
			domain.VerificationStatus != maildomain.VerificationStatus("verified") &&
			!attempted &&
			!leased &&
			domain.LastErrorCode == nil &&
			domain.NextRetryAt == nil
	case maildomain.State("verified"):
		return domain.DesiredPresent &&
			domain.EngineDomainID != nil &&
			mailboxEntDomainProvisioningCurrentDeployedTuple(domain) &&
			domain.VerificationStatus == maildomain.VerificationStatus("verified") &&
			domain.DkimStatus == maildomain.DkimStatus("valid") &&
			!attempted &&
			!leased &&
			domain.LastErrorCode == nil &&
			domain.NextRetryAt == nil
	case maildomain.State("error"):
		if !domain.DesiredPresent || leased || domain.LastErrorCode == nil {
			return false
		}
		switch MailboxDomainProvisioningCode(*domain.LastErrorCode) {
		case MailboxDomainProvisioningCodeOutcomeUnknown:
			return attempted && domain.NextRetryAt != nil
		case MailboxDomainProvisioningCodeEngineUnavailable:
			return !attempted && domain.NextRetryAt != nil
		case MailboxDomainProvisioningCodeApplyRejected, MailboxDomainProvisioningCodeDNSInvalid:
			return !attempted && domain.NextRetryAt == nil
		default:
			return false
		}
	case maildomain.State("deleting"):
		if domain.DesiredPresent || domain.Enabled || domain.IsDefault {
			return false
		}
		if leased {
			return attempted && domain.LastErrorCode == nil && domain.NextRetryAt == nil
		}
		if attempted {
			return domain.LastErrorCode != nil &&
				MailboxDomainProvisioningCode(*domain.LastErrorCode) == MailboxDomainProvisioningCodeOutcomeUnknown &&
				domain.NextRetryAt != nil
		}
		return true
	default:
		return false
	}
}

func mailboxEntDomainProvisioningTupleValid(generation *int64, version *int32, hash *[]byte, maximumGeneration int64) bool {
	if maximumGeneration < 1 || (generation == nil) != (version == nil) || (generation == nil) != mailboxEntNullableBytesAbsent(hash) {
		return false
	}
	return generation == nil ||
		(*generation >= 1 &&
			*generation <= maximumGeneration &&
			*version >= 1 &&
			len(*hash) == sha256.Size)
}

func mailboxEntDomainProvisioningCurrentDeployedTuple(domain *ent.MailDomain) bool {
	if domain == nil || domain.DeployedGeneration == nil || domain.DeployedHashVersion == nil || domain.DeployedHash == nil {
		return false
	}
	return *domain.DeployedGeneration == domain.DesiredGeneration &&
		*domain.DeployedHashVersion == domain.DesiredHashVersion &&
		len(*domain.DeployedHash) == sha256.Size &&
		subtle.ConstantTimeCompare(*domain.DeployedHash, domain.DesiredHash) == 1
}

func mailboxEntDomainProvisioningStateValid(state maildomain.State) bool {
	switch state {
	case maildomain.State("pending"),
		maildomain.State("provisioning"),
		maildomain.State("dns_required"),
		maildomain.State("verified"),
		maildomain.State("error"),
		maildomain.State("deleting"):
		return true
	default:
		return false
	}
}

func mailboxEntDomainProvisioningVerificationStatusValid(status maildomain.VerificationStatus) bool {
	switch status {
	case maildomain.VerificationStatus("pending"),
		maildomain.VerificationStatus("verified"),
		maildomain.VerificationStatus("failed"):
		return true
	default:
		return false
	}
}

func mailboxEntDomainProvisioningDKIMStatusValid(status maildomain.DkimStatus) bool {
	switch status {
	case maildomain.DkimStatus("unknown"),
		maildomain.DkimStatus("pending"),
		maildomain.DkimStatus("valid"),
		maildomain.DkimStatus("invalid"):
		return true
	default:
		return false
	}
}

func mailboxEntDomainProvisioningErrorCodeValid(code maildomain.LastErrorCode) bool {
	switch MailboxDomainProvisioningCode(code) {
	case MailboxDomainProvisioningCodeEngineUnavailable,
		MailboxDomainProvisioningCodeApplyRejected,
		MailboxDomainProvisioningCodeOutcomeUnknown,
		MailboxDomainProvisioningCodeDNSInvalid:
		return true
	default:
		return false
	}
}

func mailboxEntDomainProvisioningVerificationStatus(value EngineVerificationState) (maildomain.VerificationStatus, bool) {
	switch value {
	case EngineVerificationUnknown, EngineVerificationPending:
		return maildomain.VerificationStatus("pending"), true
	case EngineVerificationVerified:
		return maildomain.VerificationStatus("verified"), true
	case EngineVerificationFailed:
		return maildomain.VerificationStatus("failed"), true
	default:
		return "", false
	}
}

func mailboxEntDomainProvisioningDKIMStatus(value EngineVerificationState) (maildomain.DkimStatus, bool) {
	switch value {
	case EngineVerificationUnknown:
		return maildomain.DkimStatus("unknown"), true
	case EngineVerificationPending:
		return maildomain.DkimStatus("pending"), true
	case EngineVerificationVerified:
		return maildomain.DkimStatus("valid"), true
	case EngineVerificationFailed:
		return maildomain.DkimStatus("invalid"), true
	default:
		return "", false
	}
}
