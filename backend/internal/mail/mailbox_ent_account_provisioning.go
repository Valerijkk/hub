package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

var _ MailboxAccountProvisioningRepository = (*MailboxEntRepository)(nil)

// Scan returns bounded due provisioning accounts and may atomically materialize
// the first account intent for otherwise eligible active users.
func (repository *MailboxEntRepository) Scan(ctx context.Context, request MailboxAccountProvisioningScanRequest) (MailboxAccountProvisioningScanResult, error) {
	request, ok := NormalizeMailboxAccountProvisioningScanRequest(request)
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !ok {
		return MailboxAccountProvisioningScanResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	result := MailboxAccountProvisioningScanResult{AccountIDs: make([]uuid.UUID, 0, request.BatchSize)}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		due, err := mailboxEntAccountProvisioningScanDue(ctx, transaction, databaseNow, request.BatchSize)
		if err != nil {
			return err
		}
		seen := make(map[uuid.UUID]struct{}, request.BatchSize)
		for _, accountID := range due {
			if len(result.AccountIDs) == request.BatchSize {
				break
			}
			if _, duplicate := seen[accountID]; duplicate {
				continue
			}
			seen[accountID] = struct{}{}
			result.AccountIDs = append(result.AccountIDs, accountID)
		}
		if len(result.AccountIDs) == request.BatchSize {
			return nil
		}

		created, err := mailboxEntAccountProvisioningMaterializeIntents(ctx, transaction, databaseNow, request.BatchSize-len(result.AccountIDs))
		if err != nil {
			return err
		}
		for _, accountID := range created {
			if len(result.AccountIDs) == request.BatchSize {
				break
			}
			if _, duplicate := seen[accountID]; duplicate {
				continue
			}
			seen[accountID] = struct{}{}
			result.AccountIDs = append(result.AccountIDs, accountID)
		}
		if !result.ValidFor(request) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		return nil
	})
	if err != nil {
		return MailboxAccountProvisioningScanResult{}, mailboxEntSafeError(err)
	}
	if !result.ValidFor(request) {
		return MailboxAccountProvisioningScanResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return MailboxAccountProvisioningScanResult{AccountIDs: append(make([]uuid.UUID, 0, len(result.AccountIDs)), result.AccountIDs...)}, nil
}

func mailboxEntAccountProvisioningMaterializeIntents(ctx context.Context, transaction *ent.Tx, databaseNow time.Time, limit int) ([]uuid.UUID, error) {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(databaseNow) || limit < 1 || limit > MailboxAccountProvisioningMaximumBatchSize {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	candidates, err := transaction.User.Query().
		Where(
			user.StatusEQ(user.StatusActive),
			user.HasOrganization(),
			user.Not(user.HasMailAccount()),
		).
		Order(ent.Asc(user.FieldID)).
		Limit(limit).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	accountIDs := make([]uuid.UUID, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.ID == uuid.Nil || candidate.Status != user.StatusActive {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		organizations, err := transaction.Organization.Query().
			Where(organization.HasUsersWith(user.IDEQ(candidate.ID))).
			Order(ent.Asc(organization.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if len(organizations) != 1 || organizations[0] == nil || organizations[0].ID == uuid.Nil {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		organizationRow := organizations[0]

		existing, err := transaction.MailAccount.Query().
			Where(mailaccount.UserIDEQ(candidate.ID)).
			Order(ent.Asc(mailaccount.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if len(existing) > 1 {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if len(existing) == 1 {
			continue
		}

		domains, err := transaction.MailDomain.Query().
			Where(
				maildomain.OrganizationIDEQ(organizationRow.ID),
				maildomain.IsDefaultEQ(true),
				maildomain.DesiredPresentEQ(true),
				maildomain.EnabledEQ(true),
				maildomain.StateEQ(maildomain.State("verified")),
				maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
				maildomain.DkimStatusEQ(maildomain.DkimStatus("valid")),
				maildomain.EngineDomainIDNotNil(),
				maildomain.DeletedAtIsNil(),
			).
			Order(ent.Asc(maildomain.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if len(domains) > 1 {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if len(domains) == 0 || !mailboxEntAccountProvisioningDomainLive(domains[0], organizationRow.ID) {
			continue
		}
		domain := domains[0]

		intentID := mailboxEntAccountProvisioningIntentID(organizationRow.ID, candidate.ID, domain.ID)
		provisioningKey := mailboxEntAccountProvisioningKey(organizationRow.ID, candidate.ID, domain.ID, intentID)
		localPart, err := mailboxEntAccountProvisioningLocalPart(ctx, transaction, domain, candidate.ID, candidate.Email)
		if err != nil {
			return nil, err
		}
		controlHash := mailboxEntAccountProvisioningControlHash(organizationRow.ID, candidate.ID, domain.ID, intentID, localPart)
		if intentID == uuid.Nil || !mailboxAccountProvisioningSafeText(provisioningKey) || !validCanonicalRegistryLocalPart(localPart) || len(controlHash) != sha256.Size {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		account, err := transaction.MailAccount.Create().
			SetID(intentID).
			SetOrganizationID(organizationRow.ID).
			SetUserID(candidate.ID).
			SetMailDomainID(domain.ID).
			SetEngineProvisioningKey(provisioningKey).
			SetLocalPart(localPart).
			SetDesiredEnabled(true).
			SetStatus(mailaccount.Status("provisioning")).
			SetControlGeneration(1).
			SetControlHashVersion(1).
			SetControlHash(controlHash[:]).
			SetProjectionState(mailaccount.ProjectionState("bootstrapping")).
			SetProvisionAttempts(0).
			SetNextProvisionAt(databaseNow).
			SetEventCutoffAt(databaseNow).
			SetSyncScheduleUpdatedAt(databaseNow).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil || account == nil || account.ID != intentID {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		if _, err := transaction.MailIdentity.Create().
			SetMailAccountID(account.ID).
			SetNormalizedAddress(localPart+"@"+domain.Domain).
			SetKind(mailidentity.KindPrimary).
			SetStatus(mailidentity.Status("pending")).
			SetSendAllowed(false).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow).
			Save(ctx); err != nil {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		accountIDs = append(accountIDs, account.ID)
	}
	return accountIDs, nil
}

func mailboxEntAccountProvisioningIntentID(organizationID, userID, domainID uuid.UUID) uuid.UUID {
	if organizationID == uuid.Nil || userID == uuid.Nil || domainID == uuid.Nil {
		return uuid.Nil
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("hub.mail.account.intent.v1\x00"+organizationID.String()+"\x00"+userID.String()+"\x00"+domainID.String()))
}

func mailboxEntAccountProvisioningKey(organizationID, userID, domainID, intentID uuid.UUID) string {
	if organizationID == uuid.Nil || userID == uuid.Nil || domainID == uuid.Nil || intentID == uuid.Nil {
		return ""
	}
	return "hub-mail-account-v1:" + organizationID.String() + ":" + userID.String() + ":" + domainID.String() + ":" + intentID.String()
}

func mailboxEntAccountProvisioningControlHash(organizationID, userID, domainID, intentID uuid.UUID, localPart string) [sha256.Size]byte {
	return sha256.Sum256([]byte("hub.mail.account.control.v1\x00" + organizationID.String() + "\x00" + userID.String() + "\x00" + domainID.String() + "\x00" + intentID.String() + "\x00" + localPart + "\x00enabled=1"))
}

func mailboxEntAccountProvisioningLocalPart(ctx context.Context, transaction *ent.Tx, domain *ent.MailDomain, userID uuid.UUID, email string) (string, error) {
	if ctx == nil || transaction == nil || domain == nil || domain.ID == uuid.Nil || userID == uuid.Nil || !validCanonicalRegistryDomain(domain.Domain) {
		return "", mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	candidates := make([]string, 0, 3)
	if emailLocalPart, ok := mailboxEntAccountProvisioningEmailLocalPart(email); ok {
		candidates = append(candidates, emailLocalPart)
	}
	suffix := hex.EncodeToString(userID[:])
	if len(suffix) != 32 {
		return "", mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	shortCandidate := "u-" + suffix[:12] // u- + suffix[:12]
	fullCandidate := "u-" + suffix // u- + suffix
	for _, candidate := range []string{shortCandidate, fullCandidate} {
		if len(candidates) == 0 || candidates[len(candidates)-1] != candidate {
			candidates = append(candidates, candidate)
		}
	}
	for _, candidate := range candidates {
		if !validCanonicalRegistryLocalPart(candidate) {
			continue
		}
		available, err := mailboxEntAccountProvisioningAddressAvailable(ctx, transaction, domain.ID, candidate, candidate+"@"+domain.Domain)
		if err != nil {
			return "", err
		}
		if available {
			return candidate, nil
		}
	}
	return "", mailboxServiceError(MailboxServiceErrorUnavailable)
}

func mailboxEntAccountProvisioningEmailLocalPart(email string) (string, bool) {
	localPart, domainPart, found := strings.Cut(email, "@")
	if !found || localPart == "" || domainPart == "" || strings.Contains(domainPart, "@") {
		return "", false
	}
	localPart = strings.ToLower(localPart)
	return localPart, validCanonicalRegistryLocalPart(localPart)
}

func mailboxEntAccountProvisioningAddressAvailable(ctx context.Context, transaction *ent.Tx, domainID uuid.UUID, localPart, address string) (bool, error) {
	if ctx == nil || transaction == nil || domainID == uuid.Nil || !validCanonicalRegistryLocalPart(localPart) || !validMailboxBootstrapIdentityExpectedEmail(address) {
		return false, mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.MailDomainIDEQ(domainID),
			mailaccount.LocalPartEQ(localPart),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	identities, err := transaction.MailIdentity.Query().
		Where(mailidentity.NormalizedAddressEQ(address)).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	if len(accounts) > 1 || len(identities) > 1 {
		return false, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return len(accounts) == 0 && len(identities) == 0, nil
}

func mailboxEntAccountProvisioningDomainLive(domain *ent.MailDomain, organizationID uuid.UUID) bool {
	if domain == nil || domain.ID == uuid.Nil || organizationID == uuid.Nil || domain.OrganizationID != organizationID || !domain.IsDefault || !domain.DesiredPresent || !domain.Enabled || domain.DeletedAt != nil || domain.State != maildomain.State("verified") || domain.VerificationStatus != maildomain.VerificationStatus("verified") || domain.DkimStatus != maildomain.DkimStatus("valid") || domain.EngineDomainID == nil || !mailboxAccountProvisioningSafeText(*domain.EngineDomainID) || !validCanonicalRegistryDomain(domain.Domain) {
		return false
	}
	return domain.DesiredGeneration >= 1 && domain.DesiredHashVersion >= 1 && len(domain.DesiredHash) == sha256.Size && domain.DeployedGeneration != nil && domain.DeployedHashVersion != nil && domain.DeployedHash != nil && *domain.DeployedGeneration == domain.DesiredGeneration && *domain.DeployedHashVersion == domain.DesiredHashVersion && len(domain.DeployedHash) == sha256.Size && subtle.ConstantTimeCompare(domain.DeployedHash, domain.DesiredHash) == 1
}

// ClaimMailboxAccountProvisioning seals one provider-free capability for exactly one due account.
func (repository *MailboxEntRepository) ClaimMailboxAccountProvisioning(ctx context.Context, request MailboxAccountProvisioningClaimRequest) (MailboxAccountProvisioningClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !request.Valid() {
		return MailboxAccountProvisioningClaimResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	result := MailboxAccountProvisioningClaimResult{Outcome: MailboxAccountProvisioningNoop}
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(mailaccount.IDEQ(request.AccountID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if account == nil || account.ID != request.AccountID {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if account.Status != mailaccount.Status("provisioning") || mailboxEntAccountProvisioningConverged(account) {
			return nil
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		organizationRow, userRow, domain, identity, err := mailboxEntAccountProvisioningClaimLocks(ctx, transaction, account)
		if err != nil {
			return err
		}
		dueAt, due, validDue := mailboxEntAccountProvisioningDue(account, databaseNow)
		if !validDue {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if !due {
			return nil
		}
		if !mailboxEntAccountProvisioningPreProviderShape(account) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if domain != nil && !mailboxEntAccountProvisioningExactPendingPrimary(identity, account, domain) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if !mailboxEntAccountProvisioningDomainGate(domain, organizationRow, userRow, account) {
			if err := mailboxEntAccountProvisioningSetDomainRequired(ctx, transaction, account, dueAt, databaseNow); err != nil {
				return err
			}
			result = MailboxAccountProvisioningClaimResult{Outcome: MailboxAccountProvisioningLocalGate, Code: MailboxAccountProvisioningCodeDomainRequired}
			return nil
		}
		if !mailboxEntAccountProvisioningExactPendingPrimary(identity, account, domain) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if account.ProvisionAttempts < 0 || account.ProvisionAttempts >= math.MaxInt32 {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		leaseDeadline, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		leaseToken := uuid.New()
		if leaseToken == uuid.Nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		predicates, ok := mailboxEntAccountProvisioningCASPredicates(account)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		updated, err := transaction.MailAccount.Update().
			Where(predicates...).
			SetAttemptedControlGeneration(account.ControlGeneration).
			SetAttemptedControlHashVersion(account.ControlHashVersion).
			SetAttemptedControlHash(append([]byte(nil), account.ControlHash...)).
			SetProvisionLeaseToken(leaseToken).
			SetProvisionLeaseUntil(leaseDeadline).
			SetProvisionAttempts(account.ProvisionAttempts + 1).
			ClearNextProvisionAt().
			ClearProvisionErrorCode().
			SetUpdatedAt(databaseNow).
			Save(ctx)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		if updated != 1 {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		claim, ok := NewMailboxAccountProvisioningClaim(MailboxAccountProvisioningFence{
			AccountID:          account.ID,
			OrganizationID:     organizationRow.ID,
			DomainID:           domain.ID,
			UserID:             userRow.ID,
			LeaseToken:         leaseToken,
			LeaseDeadline:      leaseDeadline,
			LeaseDuration:      MailboxEngineCommandLeaseDuration,
			ControlGeneration:  account.ControlGeneration,
			ControlHashVersion: account.ControlHashVersion,
			ControlHash:        hex.EncodeToString(account.ControlHash),
			ProvisioningKey:    account.EngineProvisioningKey,
			LocalPart:          account.LocalPart,
			Domain:             domain.Domain,
		})
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		result = MailboxAccountProvisioningClaimResult{Outcome: MailboxAccountProvisioningClaimed, Claim: claim}
		return nil
	})
	if err != nil {
		return MailboxAccountProvisioningClaimResult{}, mailboxEntSafeError(err)
	}
	if !result.Valid() {
		return MailboxAccountProvisioningClaimResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return result, nil
}

// Finalize records only the result of an already sealed Registry account call.
func (repository *MailboxEntRepository) Finalize(ctx context.Context, request MailboxAccountProvisioningFinalizeRequest) (MailboxAccountProvisioningFinalizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !request.Valid() {
		return MailboxAccountProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	fence := request.Claim.Fence()
	controlHash, ok := mailboxEntAccountProvisioningFenceHash(fence)
	if !ok {
		return MailboxAccountProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var result MailboxAccountProvisioningFinalizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		account, identity, err := mailboxEntAccountProvisioningFinalizeLocks(ctx, transaction, fence)
		if err != nil {
			return err
		}
		if !mailboxEntAccountProvisioningFinalizeFenceMatches(account, identity, fence, controlHash, databaseNow) {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		predicates, ok := mailboxEntAccountProvisioningCASPredicates(account)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorUnavailable)
		}

		switch request.Outcome {
		case MailboxAccountProvisioningFinalizeSuccess:
			updated, err := transaction.MailAccount.Update().
				Where(predicates...).
				SetEngineAccountID(request.EngineAccountID).
				SetDeployedGeneration(fence.ControlGeneration).
				SetDeployedHashVersion(fence.ControlHashVersion).
				SetDeployedHash(append([]byte(nil), controlHash...)).
				ClearAttemptedControlGeneration().
				ClearAttemptedControlHashVersion().
				ClearAttemptedControlHash().
				ClearProvisionLeaseToken().
				ClearProvisionLeaseUntil().
				ClearNextProvisionAt().
				ClearProvisionErrorCode().
				SetUpdatedAt(databaseNow).
				Save(ctx)
			if err != nil {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			if updated != 1 {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			result = MailboxAccountProvisioningFinalizeResult{Outcome: MailboxAccountProvisioningFinalizeSuccess}
		case MailboxAccountProvisioningFinalizeRetry:
			delay, ok := MailboxAccountProvisioningRetryDelay(int(account.ProvisionAttempts))
			if !ok {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, delay)
			if !ok {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			updated, err := transaction.MailAccount.Update().
				Where(predicates...).
				ClearProvisionLeaseToken().
				ClearProvisionLeaseUntil().
				SetNextProvisionAt(retryAt).
				SetProvisionErrorCode(mailaccount.ProvisionErrorCode(request.Code)).
				SetUpdatedAt(databaseNow).
				Save(ctx)
			if err != nil {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			if updated != 1 {
				return mailboxServiceError(MailboxServiceErrorUnavailable)
			}
			result = MailboxAccountProvisioningFinalizeResult{Outcome: MailboxAccountProvisioningFinalizeRetry, Code: request.Code}
		default:
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return nil
	})
	if err != nil {
		return MailboxAccountProvisioningFinalizeResult{}, mailboxEntSafeError(err)
	}
	if !result.Valid() {
		return MailboxAccountProvisioningFinalizeResult{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return result, nil
}

func mailboxEntAccountProvisioningScanDue(ctx context.Context, transaction *ent.Tx, databaseNow time.Time, limit int) ([]uuid.UUID, error) {
	if ctx == nil || transaction == nil || !mailboxEntValidTime(databaseNow) || limit < 1 || limit > MailboxAccountProvisioningMaximumBatchSize {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	dueRows, err := transaction.MailAccount.Query().
		Where(
			mailaccount.StatusEQ(mailaccount.Status("provisioning")),
			mailaccount.NextProvisionAtNotNil(),
			mailaccount.NextProvisionAtLTE(databaseNow),
			mailaccount.ProvisionLeaseTokenIsNil(),
			mailaccount.ProvisionLeaseUntilIsNil(),
		).
		Order(ent.Asc(mailaccount.FieldNextProvisionAt), ent.Asc(mailaccount.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	expiredRows, err := transaction.MailAccount.Query().
		Where(
			mailaccount.StatusEQ(mailaccount.Status("provisioning")),
			mailaccount.ProvisionLeaseTokenNotNil(),
			mailaccount.ProvisionLeaseUntilNotNil(),
			mailaccount.ProvisionLeaseUntilLTE(databaseNow),
		).
		Order(ent.Asc(mailaccount.FieldProvisionLeaseUntil), ent.Asc(mailaccount.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	type dueAccount struct {
		accountID uuid.UUID
		dueAt     time.Time
	}
	candidates := make([]dueAccount, 0, len(dueRows)+len(expiredRows))
	for _, account := range dueRows {
		if account == nil || account.ID == uuid.Nil || account.NextProvisionAt == nil || !mailboxEntValidTime(*account.NextProvisionAt) || account.NextProvisionAt.After(databaseNow) {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		candidates = append(candidates, dueAccount{accountID: account.ID, dueAt: *account.NextProvisionAt})
	}
	for _, account := range expiredRows {
		if account == nil || account.ID == uuid.Nil || account.ProvisionLeaseToken == nil || *account.ProvisionLeaseToken == uuid.Nil || account.ProvisionLeaseUntil == nil || !mailboxEntValidTime(*account.ProvisionLeaseUntil) || account.ProvisionLeaseUntil.After(databaseNow) {
			return nil, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		candidates = append(candidates, dueAccount{accountID: account.ID, dueAt: *account.ProvisionLeaseUntil})
	}
	sort.SliceStable(candidates, func(left, right int) bool {
		if candidates[left].dueAt.Equal(candidates[right].dueAt) {
			return candidates[left].accountID.String() < candidates[right].accountID.String()
		}
		return candidates[left].dueAt.Before(candidates[right].dueAt)
	})

	accountIDs := make([]uuid.UUID, 0, limit)
	seen := make(map[uuid.UUID]struct{}, limit)
	for _, candidate := range candidates {
		if len(accountIDs) == limit {
			break
		}
		if _, duplicate := seen[candidate.accountID]; duplicate {
			continue
		}
		seen[candidate.accountID] = struct{}{}
		accountIDs = append(accountIDs, candidate.accountID)
	}
	return accountIDs, nil
}

func mailboxEntAccountProvisioningClaimLocks(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount) (*ent.Organization, *ent.User, *ent.MailDomain, *ent.MailIdentity, error) {
	if ctx == nil || transaction == nil || account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil {
		return nil, nil, nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	organizationRow, err := transaction.Organization.Query().
		Where(
			organization.IDEQ(account.OrganizationID),
			organization.HasUsersWith(user.IDEQ(account.UserID)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || organizationRow == nil || organizationRow.ID != account.OrganizationID {
		return nil, nil, nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	userRow, err := transaction.User.Query().
		Where(
			user.IDEQ(account.UserID),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(organizationRow.ID)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || userRow == nil || userRow.ID != account.UserID {
		return nil, nil, nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	domains, err := transaction.MailDomain.Query().
		Where(maildomain.IDEQ(account.MailDomainID)).
		Order(ent.Asc(maildomain.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(domains) > 1 {
		return nil, nil, nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	var domain *ent.MailDomain
	if len(domains) == 1 {
		domain = domains[0]
	}
	identities, err := transaction.MailIdentity.Query().
		Where(mailidentity.MailAccountIDEQ(account.ID)).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(identities) != 1 || identities[0] == nil {
		return nil, nil, nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return organizationRow, userRow, domain, identities[0], nil
}

func mailboxEntAccountProvisioningDue(account *ent.MailAccount, databaseNow time.Time) (time.Time, bool, bool) {
	if account == nil || !mailboxEntValidTime(databaseNow) || (account.ProvisionLeaseToken == nil) != (account.ProvisionLeaseUntil == nil) {
		return time.Time{}, false, false
	}
	if account.ProvisionLeaseToken != nil {
		if *account.ProvisionLeaseToken == uuid.Nil || !mailboxEntValidTime(*account.ProvisionLeaseUntil) {
			return time.Time{}, false, false
		}
		return *account.ProvisionLeaseUntil, !account.ProvisionLeaseUntil.After(databaseNow), true
	}
	if account.NextProvisionAt == nil || !mailboxEntValidTime(*account.NextProvisionAt) {
		return time.Time{}, false, false
	}
	return *account.NextProvisionAt, !account.NextProvisionAt.After(databaseNow), true
}

func mailboxEntAccountProvisioningConverged(account *ent.MailAccount) bool {
	if account == nil || account.Status != mailaccount.Status("provisioning") || !account.DesiredEnabled || account.EngineAccountID == nil || !mailboxAccountProvisioningSafeText(*account.EngineAccountID) || account.ControlGeneration < 1 || account.ControlHashVersion < 1 || len(account.ControlHash) != sha256.Size || account.DeployedGeneration == nil || account.DeployedHashVersion == nil || account.DeployedHash == nil || *account.DeployedGeneration != account.ControlGeneration || *account.DeployedHashVersion != account.ControlHashVersion || len(account.DeployedHash) != sha256.Size || subtle.ConstantTimeCompare(account.DeployedHash, account.ControlHash) != 1 {
		return false
	}
	return account.AttemptedControlGeneration == nil && account.AttemptedControlHashVersion == nil && account.AttemptedControlHash == nil && account.ProvisionLeaseToken == nil && account.ProvisionLeaseUntil == nil && account.NextProvisionAt == nil && account.ProvisionErrorCode == nil
}

func mailboxEntAccountProvisioningPreProviderShape(account *ent.MailAccount) bool {
	if account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil || account.Status != mailaccount.Status("provisioning") || !account.DesiredEnabled || account.ProjectionState != mailaccount.ProjectionState("bootstrapping") || account.EngineAccountID != nil || !mailboxAccountProvisioningSafeText(account.EngineProvisioningKey) || !validCanonicalRegistryLocalPart(account.LocalPart) || account.ControlGeneration < 1 || account.ControlHashVersion < 1 || len(account.ControlHash) != sha256.Size || account.DeployedGeneration != nil || account.DeployedHashVersion != nil || account.DeployedHash != nil || account.PasswordEnrollmentGeneration != 0 || account.PasswordEnrolledEngineAccountID != nil || account.PasswordEnrolledAt != nil || account.PasswordEnrollmentLeaseToken != nil || account.PasswordEnrollmentLeaseUntil != nil || account.ProvisionAttempts < 0 || !mailboxEntValidTime(account.EventCutoffAt) || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return false
	}
	if (account.AttemptedControlGeneration == nil) != (account.AttemptedControlHashVersion == nil) || (account.AttemptedControlGeneration == nil) != (account.AttemptedControlHash == nil) {
		return false
	}
	if account.AttemptedControlGeneration != nil && (*account.AttemptedControlGeneration != account.ControlGeneration || *account.AttemptedControlHashVersion != account.ControlHashVersion || len(account.AttemptedControlHash) != sha256.Size || subtle.ConstantTimeCompare(account.AttemptedControlHash, account.ControlHash) != 1) {
		return false
	}
	if (account.ProvisionLeaseToken == nil) != (account.ProvisionLeaseUntil == nil) || (account.ProvisionLeaseUntil != nil && (!mailboxEntValidTime(*account.ProvisionLeaseUntil) || account.AttemptedControlGeneration == nil)) {
		return false
	}
	if account.NextProvisionAt != nil && !mailboxEntValidTime(*account.NextProvisionAt) {
		return false
	}
	if account.ProvisionErrorCode == nil {
		return true
	}
	return *account.ProvisionErrorCode == mailaccount.ProvisionErrorCode(MailboxAccountProvisioningCodeDomainRequired) || *account.ProvisionErrorCode == mailaccount.ProvisionErrorCode(MailboxAccountProvisioningCodeProvisioningRetry) || *account.ProvisionErrorCode == mailaccount.ProvisionErrorCode(MailboxAccountProvisioningCodeUnavailable)
}

func mailboxEntAccountProvisioningExactPendingPrimary(identity *ent.MailIdentity, account *ent.MailAccount, domain *ent.MailDomain) bool {
	if identity == nil || account == nil || domain == nil || identity.ID == uuid.Nil || identity.MailAccountID != account.ID || identity.Kind != mailidentity.KindPrimary || identity.Status != mailidentity.Status("pending") || identity.SendAllowed || identity.EngineIdentityID != nil {
		return false
	}
	expectedAddress := account.LocalPart + "@" + domain.Domain
	return validMailboxBootstrapIdentityExpectedEmail(expectedAddress) && identity.NormalizedAddress == expectedAddress
}

func mailboxEntAccountProvisioningDomainGate(domain *ent.MailDomain, organizationRow *ent.Organization, userRow *ent.User, account *ent.MailAccount) bool {
	if organizationRow == nil || userRow == nil || account == nil || organizationRow.ID == uuid.Nil || userRow.ID == uuid.Nil || account.OrganizationID != organizationRow.ID || account.UserID != userRow.ID || userRow.Status != user.StatusActive {
		return false
	}
	return mailboxEntAccountProvisioningDomainLive(domain, organizationRow.ID) && domain.ID == account.MailDomainID
}

func mailboxEntAccountProvisioningSetDomainRequired(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, dueAt, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || account == nil || !mailboxEntValidTime(dueAt) || !mailboxEntValidTime(databaseNow) || dueAt.After(databaseNow) {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	predicates, ok := mailboxEntAccountProvisioningCASPredicates(account)
	if !ok {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	update := transaction.MailAccount.Update().
		Where(predicates...).
		ClearAttemptedControlGeneration().
		ClearAttemptedControlHashVersion().
		ClearAttemptedControlHash().
		ClearProvisionLeaseToken().
		ClearProvisionLeaseUntil().
		SetProvisionErrorCode(mailaccount.ProvisionErrorCode(MailboxAccountProvisioningCodeDomainRequired)).
		SetUpdatedAt(databaseNow)
	if account.NextProvisionAt == nil {
		update.SetNextProvisionAt(dueAt)
	}
	updated, err := update.Save(ctx)
	if err != nil || updated != 1 {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return nil
}

func mailboxEntAccountProvisioningCASPredicates(account *ent.MailAccount) ([]predicate.MailAccount, bool) {
	if !mailboxEntAccountProvisioningPreProviderShape(account) {
		return nil, false
	}
	predicates := []predicate.MailAccount{
		mailaccount.IDEQ(account.ID),
		mailaccount.OrganizationIDEQ(account.OrganizationID),
		mailaccount.UserIDEQ(account.UserID),
		mailaccount.MailDomainIDEQ(account.MailDomainID),
		mailaccount.EngineProvisioningKeyEQ(account.EngineProvisioningKey),
		mailaccount.LocalPartEQ(account.LocalPart),
		mailaccount.EngineAccountIDIsNil(),
		mailaccount.PasswordEnrollmentGenerationEQ(0),
		mailaccount.PasswordEnrolledEngineAccountIDIsNil(),
		mailaccount.PasswordEnrolledAtIsNil(),
		mailaccount.PasswordEnrollmentLeaseTokenIsNil(),
		mailaccount.PasswordEnrollmentLeaseUntilIsNil(),
		mailaccount.DesiredEnabledEQ(true),
		mailaccount.StatusEQ(mailaccount.Status("provisioning")),
		mailaccount.ControlGenerationEQ(account.ControlGeneration),
		mailaccount.ControlHashVersionEQ(account.ControlHashVersion),
		mailaccount.ControlHashEQ(account.ControlHash),
		mailaccount.DeployedGenerationIsNil(),
		mailaccount.DeployedHashVersionIsNil(),
		mailaccount.DeployedHashIsNil(),
		mailaccount.ProjectionStateEQ(mailaccount.ProjectionState("bootstrapping")),
		mailaccount.ProvisionAttemptsEQ(account.ProvisionAttempts),
		mailaccount.EventCutoffAtEQ(account.EventCutoffAt),
		mailaccount.SyncScheduleUpdatedAtEQ(account.SyncScheduleUpdatedAt),
	}
	if account.AttemptedControlGeneration == nil {
		predicates = append(predicates, mailaccount.AttemptedControlGenerationIsNil(), mailaccount.AttemptedControlHashVersionIsNil(), mailaccount.AttemptedControlHashIsNil())
	} else {
		predicates = append(predicates,
			mailaccount.AttemptedControlGenerationEQ(*account.AttemptedControlGeneration),
			mailaccount.AttemptedControlHashVersionEQ(*account.AttemptedControlHashVersion),
			mailaccount.AttemptedControlHashEQ(account.AttemptedControlHash),
		)
	}
	if account.ProvisionLeaseToken == nil {
		predicates = append(predicates, mailaccount.ProvisionLeaseTokenIsNil(), mailaccount.ProvisionLeaseUntilIsNil())
	} else {
		predicates = append(predicates, mailaccount.ProvisionLeaseTokenEQ(*account.ProvisionLeaseToken), mailaccount.ProvisionLeaseUntilEQ(*account.ProvisionLeaseUntil))
	}
	if account.NextProvisionAt == nil {
		predicates = append(predicates, mailaccount.NextProvisionAtIsNil())
	} else {
		predicates = append(predicates, mailaccount.NextProvisionAtEQ(*account.NextProvisionAt))
	}
	if account.ProvisionErrorCode == nil {
		predicates = append(predicates, mailaccount.ProvisionErrorCodeIsNil())
	} else {
		predicates = append(predicates, mailaccount.ProvisionErrorCodeEQ(*account.ProvisionErrorCode))
	}
	return predicates, true
}

func mailboxEntAccountProvisioningFinalizeLocks(ctx context.Context, transaction *ent.Tx, fence MailboxAccountProvisioningFence) (*ent.MailAccount, *ent.MailIdentity, error) {
	if ctx == nil || transaction == nil || !fence.Valid() {
		return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	account, err := transaction.MailAccount.Query().
		Where(mailaccount.IDEQ(fence.AccountID)).
		ForUpdate().
		Only(ctx)
	if err != nil || account == nil || account.ID != fence.AccountID {
		return nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	identities, err := transaction.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(fence.AccountID),
			mailidentity.KindEQ(mailidentity.KindPrimary),
			mailidentity.StatusEQ(mailidentity.Status("pending")),
			mailidentity.SendAllowedEQ(false),
			mailidentity.EngineIdentityIDIsNil(),
			mailidentity.NormalizedAddressEQ(fence.LocalPart+"@"+fence.Domain),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil || len(identities) != 1 || identities[0] == nil {
		return nil, nil, mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	return account, identities[0], nil
}

func mailboxEntAccountProvisioningFenceHash(fence MailboxAccountProvisioningFence) ([]byte, bool) {
	if !fence.Valid() || len(fence.ControlHash) != hex.EncodedLen(sha256.Size) {
		return nil, false
	}
	decoded, err := hex.DecodeString(fence.ControlHash)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != fence.ControlHash {
		return nil, false
	}
	return decoded, true
}

func mailboxEntAccountProvisioningFinalizeFenceMatches(account *ent.MailAccount, identity *ent.MailIdentity, fence MailboxAccountProvisioningFence, controlHash []byte, databaseNow time.Time) bool {
	if account == nil || identity == nil || !fence.Valid() || len(controlHash) != sha256.Size || !mailboxEntValidTime(databaseNow) || account.ID != fence.AccountID || account.OrganizationID != fence.OrganizationID || account.UserID != fence.UserID || account.MailDomainID != fence.DomainID || !mailboxEntAccountProvisioningPreProviderShape(account) || account.ControlGeneration != fence.ControlGeneration || account.ControlHashVersion != fence.ControlHashVersion || subtle.ConstantTimeCompare(account.ControlHash, controlHash) != 1 || account.AttemptedControlGeneration == nil || account.AttemptedControlHashVersion == nil || account.AttemptedControlHash == nil || *account.AttemptedControlGeneration != fence.ControlGeneration || *account.AttemptedControlHashVersion != fence.ControlHashVersion || len(account.AttemptedControlHash) != sha256.Size || subtle.ConstantTimeCompare(account.AttemptedControlHash, controlHash) != 1 || account.ProvisionLeaseToken == nil || *account.ProvisionLeaseToken != fence.LeaseToken || account.ProvisionLeaseUntil == nil || !account.ProvisionLeaseUntil.Equal(fence.LeaseDeadline) || !account.ProvisionLeaseUntil.After(databaseNow) || account.EngineProvisioningKey != fence.ProvisioningKey || account.LocalPart != fence.LocalPart || account.ProvisionAttempts < 1 {
		return false
	}
	return mailboxEntAccountProvisioningExactPendingPrimary(identity, account, &ent.MailDomain{ID: fence.DomainID, Domain: fence.Domain})
}
