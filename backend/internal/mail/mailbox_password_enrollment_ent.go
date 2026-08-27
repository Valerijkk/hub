package mail

import (
	"context"
	"crypto/sha256"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const mailboxPasswordEnrollmentEntLeaseDuration = 30 * time.Second

var _ MailboxPasswordEnrollmentRepository = (*MailboxEntRepository)(nil)

// Claim acquires the sole durable password-enrollment lease for one active
// owner. It performs no Registry or other external I/O.
func (repository *MailboxEntRepository) Claim(ctx context.Context, actorID, requestedLease uuid.UUID) (MailboxPasswordEnrollmentClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || actorID == uuid.Nil || requestedLease == uuid.Nil {
		return MailboxPasswordEnrollmentClaimResult{}, newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}

	var result MailboxPasswordEnrollmentClaimResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := mailboxPasswordEnrollmentLockOwner(ctx, transaction, actorID)
		if err != nil {
			return err
		}
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentAccountIntegrityValid(account) {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentEligible(account) {
			return mailboxPasswordEnrollmentEntRetryable()
		}

		if account.PasswordEnrolledEngineAccountID != nil {
			if account.PasswordEnrollmentLeaseToken != nil {
				return mailboxPasswordEnrollmentEntUnavailable()
			}
			result = NewMailboxPasswordEnrollmentNoopResult()
			return nil
		}
		if account.PasswordEnrollmentLeaseUntil != nil && account.PasswordEnrollmentLeaseUntil.After(databaseWallClock) {
			result = NewMailboxPasswordEnrollmentBusyResult()
			return nil
		}
		if account.PasswordEnrollmentGeneration == math.MaxInt64 {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		deployedGeneration, deployedHashVersion, deployedHash, valid := mailboxPasswordEnrollmentClaimDeployedTuple(account)
		if !valid {
			return mailboxPasswordEnrollmentEntUnavailable()
		}

		leaseUntil, valid := mailboxEntCheckedTimeAdd(databaseWallClock, mailboxPasswordEnrollmentEntLeaseDuration)
		if !valid {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		nextGeneration := account.PasswordEnrollmentGeneration + 1
		if nextGeneration < 1 {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if err := account.Update().
			SetPasswordEnrollmentGeneration(nextGeneration).
			SetPasswordEnrollmentLeaseToken(requestedLease).
			SetPasswordEnrollmentLeaseUntil(leaseUntil).
			SetUpdatedAt(databaseWallClock).
			Exec(ctx); err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}

		claimed, err := NewMailboxPasswordEnrollmentClaimedResult(MailboxPasswordEnrollmentClaimedInput{
			AccountID:           account.ID,
			OrganizationID:      account.OrganizationID,
			EngineAccountID:     *account.EngineAccountID,
			DeployedGeneration:  deployedGeneration,
			DeployedHashVersion: deployedHashVersion,
			DeployedHash:        deployedHash,
			Generation:          nextGeneration,
			LeaseToken:          requestedLease,
			LeaseUntil:          leaseUntil,
		})
		if err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		result = claimed
		return nil
	})
	if err != nil {
		return MailboxPasswordEnrollmentClaimResult{}, mailboxPasswordEnrollmentEntError(err)
	}
	if result.Outcome() == "" {
		return MailboxPasswordEnrollmentClaimResult{}, mailboxPasswordEnrollmentEntUnavailable()
	}
	return result, nil
}

// Complete marks the exact successful lease as enrolled. Any stale capability
// fails closed and cannot mark a newer account state as complete.
func (repository *MailboxEntRepository) Complete(ctx context.Context, claim MailboxPasswordEnrollmentClaim, observedAt time.Time) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxPasswordEnrollmentClaimValid(claim) || !mailboxEntValidTime(observedAt) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	claim = mailboxPasswordEnrollmentClaimCopy(claim)
	observedAt = observedAt.Round(0).UTC()

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(claim.accountID),
				mailaccount.OrganizationIDEQ(claim.organizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return mailboxPasswordEnrollmentEntRetryable()
			}
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentAccountIntegrityValid(account) {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentCurrentTupleMatches(account, claim, databaseWallClock, true) {
			return mailboxPasswordEnrollmentEntRetryable()
		}

		if err := account.Update().
			SetPasswordEnrolledEngineAccountID(claim.engineAccountID).
			SetPasswordEnrolledAt(observedAt).
			ClearPasswordEnrollmentLeaseToken().
			ClearPasswordEnrollmentLeaseUntil().
			SetUpdatedAt(databaseWallClock).
			Exec(ctx); err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		return nil
	})
	return mailboxPasswordEnrollmentEntError(err)
}

// Release clears only the exact current claim after a failed Registry
// replacement. A stale release deliberately succeeds as a no-op.
func (repository *MailboxEntRepository) Release(ctx context.Context, claim MailboxPasswordEnrollmentClaim, code MailboxPasswordEnrollmentReleaseCode) error {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxPasswordEnrollmentClaimValid(claim) || !validMailboxPasswordEnrollmentReleaseCode(code) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	claim = mailboxPasswordEnrollmentClaimCopy(claim)

	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().
			Where(
				mailaccount.IDEQ(claim.accountID),
				mailaccount.OrganizationIDEQ(claim.organizationID),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		databaseWallClock, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentAccountIntegrityValid(account) {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		if !mailboxPasswordEnrollmentCurrentTupleMatches(account, claim, databaseWallClock, false) {
			return nil
		}

		if err := account.Update().
			ClearPasswordEnrollmentLeaseToken().
			ClearPasswordEnrollmentLeaseUntil().
			SetUpdatedAt(databaseWallClock).
			Exec(ctx); err != nil {
			return mailboxPasswordEnrollmentEntUnavailable()
		}
		return nil
	})
	return mailboxPasswordEnrollmentEntError(err)
}

func mailboxPasswordEnrollmentLockOwner(ctx context.Context, transaction *ent.Tx, actorID uuid.UUID) (*ent.MailAccount, error) {
	if ctx == nil || transaction == nil || actorID == uuid.Nil {
		return nil, mailboxPasswordEnrollmentEntUnavailable()
	}

	organizationRow, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, mailboxPasswordEnrollmentEntUnavailable()
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
		return nil, mailboxPasswordEnrollmentEntUnavailable()
	}
	accounts, err := transaction.MailAccount.Query().
		Where(
			mailaccount.UserIDEQ(actor.ID),
			mailaccount.OrganizationIDEQ(organizationRow.ID),
		).
		Order(ent.Asc(mailaccount.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, mailboxPasswordEnrollmentEntUnavailable()
	}
	if len(accounts) != 1 {
		return nil, mailboxPasswordEnrollmentEntUnavailable()
	}
	account := accounts[0]
	if account == nil || account.ID == uuid.Nil || account.UserID != actor.ID || account.OrganizationID != organizationRow.ID {
		return nil, mailboxPasswordEnrollmentEntUnavailable()
	}
	return account, nil
}

func mailboxPasswordEnrollmentAccountIntegrityValid(account *ent.MailAccount) bool {
	if account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.PasswordEnrollmentGeneration < 0 {
		return false
	}
	if account.EngineAccountID != nil && !validOpaqueID(*account.EngineAccountID) {
		return false
	}
	if !mailboxPasswordEnrollmentDeployedTupleValid(account) {
		return false
	}
	if (account.PasswordEnrolledEngineAccountID == nil) != (account.PasswordEnrolledAt == nil) {
		return false
	}
	if account.PasswordEnrolledEngineAccountID != nil {
		if account.EngineAccountID == nil || *account.PasswordEnrolledEngineAccountID != *account.EngineAccountID || !validOpaqueID(*account.PasswordEnrolledEngineAccountID) || account.PasswordEnrollmentGeneration < 1 || !mailboxEntValidTime(*account.PasswordEnrolledAt) {
			return false
		}
	}
	if (account.PasswordEnrollmentLeaseToken == nil) != (account.PasswordEnrollmentLeaseUntil == nil) {
		return false
	}
	if account.PasswordEnrollmentLeaseToken != nil && (*account.PasswordEnrollmentLeaseToken == uuid.Nil || !mailboxEntValidTime(*account.PasswordEnrollmentLeaseUntil)) {
		return false
	}
	return true
}

func mailboxPasswordEnrollmentDeployedTupleValid(account *ent.MailAccount) bool {
	if account == nil {
		return false
	}
	hasGeneration := account.DeployedGeneration != nil
	hasVersion := account.DeployedHashVersion != nil
	hasHash := account.DeployedHash != nil
	if hasGeneration != hasVersion || hasGeneration != hasHash {
		return false
	}
	return !hasGeneration || (*account.DeployedGeneration >= 1 && *account.DeployedHashVersion >= 1 && len(*account.DeployedHash) == sha256.Size)
}

func mailboxPasswordEnrollmentEligible(account *ent.MailAccount) bool {
	return mailboxPasswordEnrollmentAccountIntegrityValid(account) &&
		account.Status == mailaccount.Status("provisioning") &&
		account.Status != mailaccount.Status("deleting") &&
		account.DesiredEnabled &&
		account.EngineAccountID != nil &&
		validOpaqueID(*account.EngineAccountID) &&
		account.DeployedGeneration != nil &&
		account.DeployedHashVersion != nil &&
		account.DeployedHash != nil
}

func mailboxPasswordEnrollmentCurrentTupleMatches(account *ent.MailAccount, claim MailboxPasswordEnrollmentClaim, databaseWallClock time.Time, requireLiveLease bool) bool {
	if !mailboxPasswordEnrollmentClaimValid(claim) || !mailboxEntValidTime(databaseWallClock) || !mailboxPasswordEnrollmentEligible(account) || !mailboxPasswordEnrollmentClaimDeployedTupleMatches(account, claim) || account.ID != claim.accountID || account.OrganizationID != claim.organizationID || *account.EngineAccountID != claim.engineAccountID || account.PasswordEnrollmentGeneration != claim.generation || account.PasswordEnrolledEngineAccountID != nil || account.PasswordEnrolledAt != nil || account.PasswordEnrollmentLeaseToken == nil || account.PasswordEnrollmentLeaseUntil == nil || *account.PasswordEnrollmentLeaseToken != claim.leaseToken || !account.PasswordEnrollmentLeaseUntil.Equal(claim.leaseUntil) {
		return false
	}
	return !requireLiveLease || account.PasswordEnrollmentLeaseUntil.After(databaseWallClock)
}

func mailboxPasswordEnrollmentClaimDeployedTuple(account *ent.MailAccount) (int64, int32, [sha256.Size]byte, bool) {
	if !mailboxPasswordEnrollmentEligible(account) {
		return 0, 0, [sha256.Size]byte{}, false
	}
	var deployedHash [sha256.Size]byte
	copy(deployedHash[:], *account.DeployedHash)
	return *account.DeployedGeneration, *account.DeployedHashVersion, deployedHash, true
}

func mailboxPasswordEnrollmentClaimDeployedTupleMatches(account *ent.MailAccount, claim MailboxPasswordEnrollmentClaim) bool {
	if account == nil || !mailboxPasswordEnrollmentClaimValid(claim) || account.DeployedGeneration == nil || account.DeployedHashVersion == nil || account.DeployedHash == nil || len(*account.DeployedHash) != sha256.Size {
		return false
	}
	var deployedHash [sha256.Size]byte
	copy(deployedHash[:], *account.DeployedHash)
	return *account.DeployedGeneration == claim.deployedGeneration &&
		*account.DeployedHashVersion == claim.deployedHashVersion &&
		deployedHash == claim.deployedHash
}

func mailboxPasswordEnrollmentEntUnavailable() error {
	return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRepositoryUnavailable)
}

func mailboxPasswordEnrollmentEntRetryable() error {
	return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRetryable)
}

func mailboxPasswordEnrollmentEntError(err error) error {
	if err == nil {
		return nil
	}
	if kind, ok := MailboxPasswordEnrollmentErrorKindOf(err); ok {
		return newMailboxPasswordEnrollmentError(kind)
	}
	return mailboxPasswordEnrollmentEntUnavailable()
}
