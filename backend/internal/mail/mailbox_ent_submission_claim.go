package mail

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/mailboxpolicy"
	"github.com/valerii/hub/backend/ent/mailsendratebucket"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

var _ MailboxSubmissionClaimRepository = (*MailboxEntRepository)(nil)

// ClaimMailboxSubmission atomically advances one due queued submission to the
// short authorizing lease. It intentionally never calls BeginEngineCommand:
// the account and submission fences are minted together in one transaction.
func (repository *MailboxEntRepository) ClaimMailboxSubmission(ctx context.Context, request MailboxSubmissionClaimRequest) (MailboxSubmissionClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || request.SubmissionID == uuid.Nil {
		return MailboxSubmissionClaimResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	// This non-locking lookup is deliberately limited to the submission owner;
	// every mutable decision is made after the ordered transaction locks below.
	lookup, err := repository.client.MailSubmission.Query().Where(mailsubmission.IDEQ(request.SubmissionID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return MailboxSubmissionClaimResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
		}
		return MailboxSubmissionClaimResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if lookup == nil || lookup.ID != request.SubmissionID || lookup.MailAccountID == uuid.Nil {
		return MailboxSubmissionClaimResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}

	var result MailboxSubmissionClaimResult
	err = repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft -> MailboxPolicy -> MailDomain.
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(lookup.MailAccountID)).ForUpdate().Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		if kind := mailboxEntSubmissionClaimAccountKind(account); kind != "" {
			return NewMailboxSubmissionRepositoryError(kind)
		}

		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(request.SubmissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		if !mailboxEntSubmissionRowValid(submission, account.ID) || submission.ID != request.SubmissionID {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if submission.MailDraftID == nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		policy, err := transaction.MailboxPolicy.Query().
			Where(mailboxpolicy.OrganizationIDEQ(account.OrganizationID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		domains, err := transaction.MailDomain.Query().
			Where(
				maildomain.OrganizationIDEQ(account.OrganizationID),
				maildomain.DesiredPresentEQ(true),
				maildomain.EnabledEQ(true),
				maildomain.StateEQ(maildomain.State("verified")),
				maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
				maildomain.DeletedAtIsNil(),
			).
			Order(ent.Asc(maildomain.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionClaimRepositoryError(err)
		}
		if busyUntil, busy, valid := mailboxEntSubmissionClaimBusyUntil(account); !valid {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		} else if busy {
			result = MailboxSubmissionClaimResult{busyUntil: &busyUntil}
			return nil
		}
		if MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusQueued || submission.NextAttemptAt == nil ||
			submission.NextAttemptAt.After(databaseNow) || submission.AttemptCount < 0 || submission.AttemptCount > 7 {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}
		binding, ok := mailboxEntSubmissionBindingFromRow(submission)
		if !ok || !mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		policyBinding, err := mailboxEntSubmissionClaimPolicyBinding(account, policy, domains)
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, MailboxEngineCommandLeaseDuration)
		if !ok {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
		if !ok {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		leaseID := uuid.New()
		fence := MailboxEngineCommandFence{AccountID: account.ID, LeaseID: leaseID, MutationEpoch: nextEpoch, LeaseUntil: leaseUntil}
		if !mailboxSubmissionClaimFenceValid(fence) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}

		if err := account.Update().
			SetMutationLeaseToken(leaseID).
			SetMutationLeaseUntil(leaseUntil).
			SetMutationEpoch(nextEpoch).
			SetSyncScheduleUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		if err := submission.Update().
			SetStatus(mailsubmission.Status("authorizing")).
			SetCommandLeaseToken(leaseID).
			SetCommandLeaseExpiresAt(leaseUntil).
			SetCommandFence(nextEpoch).
			ClearDispatchAuthorizedAt().
			ClearDispatchAuthorizationHash().
			ClearNextAttemptAt().
			SetAttemptCount(submission.AttemptCount + 1).
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}

		engineSubmission := mailboxEntSubmissionEngineSubmission(binding)
		claim := MailboxSubmissionClaim{
			submissionID: submission.ID,
			fence:        fence,
			policy:       policyBinding,
			submission:   engineSubmission,
		}
		if !mailboxSubmissionClaimValid(claim) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = MailboxSubmissionClaimResult{claimed: true, claim: mailboxSubmissionClaimCopy(claim)}
		return nil
	})
	if err != nil {
		return MailboxSubmissionClaimResult{}, mailboxEntSubmissionClaimRepositoryError(err)
	}
	return result, nil
}

// AuthorizeMailboxSubmission performs the pre-dispatch authorization CAS. It
// consumes only the sealed claim and read-only provider preflight result.
func (repository *MailboxEntRepository) AuthorizeMailboxSubmission(ctx context.Context, request MailboxSubmissionAuthorizeRequest) (MailboxSubmissionAuthorizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionAuthorizeRequestValid(request) {
		return MailboxSubmissionAuthorizeResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	claim := mailboxSubmissionClaimCopy(request.Claim)
	preflight := request.Preflight

	var result MailboxSubmissionAuthorizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		// Lock order: MailAccount -> MailSubmission -> MailDraft -> User -> MailIdentity
		// -> MailAttachment -> MailObject -> MailboxPolicy -> MailDomain -> rate buckets.
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(claim.fence.AccountID)).ForUpdate().Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(claim.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		if submission.MailDraftID == nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}

		terminalCode := MailboxSubmissionErrorCode("")
		actor, err := transaction.User.Query().
			Where(
				user.IDEQ(account.UserID),
				user.StatusEQ(user.StatusActive),
				user.HasOrganizationWith(organization.IDEQ(account.OrganizationID)),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				terminalCode = MailboxSubmissionErrorCodePolicy
			} else {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
			}
		}
		if terminalCode == "" && (actor == nil || actor.ID != account.UserID) {
			terminalCode = MailboxSubmissionErrorCodePolicy
		}

		var identity *ent.MailIdentity
		if terminalCode == "" {
			identity, err = transaction.MailIdentity.Query().
				Where(
					mailidentity.IDEQ(draft.FromIdentityID),
					mailidentity.MailAccountIDEQ(account.ID),
					mailidentity.StatusEQ(mailidentity.Status("active")),
					mailidentity.SendAllowedEQ(true),
					mailidentity.KindIn(mailidentity.Kind("primary"), mailidentity.Kind("alias")),
				).
				ForUpdate().
				Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					terminalCode = MailboxSubmissionErrorCodeIdentityUnavailable
				} else {
					return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
				}
			}
		}

		var attachments []MailboxSubmissionCleanAttachment
		if terminalCode == "" {
			var clean bool
			attachments, clean, err = mailboxEntSubmissionLockCleanAttachments(ctx, transaction, account, draft)
			if err != nil {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
			}
			if !clean {
				terminalCode = MailboxSubmissionErrorCodeAttachmentNotClean
			}
		}

		var policy *ent.MailboxPolicy
		var domains []*ent.MailDomain
		if terminalCode == "" {
			policy, err = transaction.MailboxPolicy.Query().
				Where(mailboxpolicy.OrganizationIDEQ(account.OrganizationID)).
				ForUpdate().
				Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					terminalCode = MailboxSubmissionErrorCodePolicy
				} else {
					return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
				}
			}
		}
		if terminalCode == "" {
			domains, err = transaction.MailDomain.Query().
				Where(
					maildomain.OrganizationIDEQ(account.OrganizationID),
					maildomain.DesiredPresentEQ(true),
					maildomain.EnabledEQ(true),
					maildomain.StateEQ(maildomain.State("verified")),
					maildomain.VerificationStatusEQ(maildomain.VerificationStatus("verified")),
					maildomain.DeletedAtIsNil(),
				).
				Order(ent.Asc(maildomain.FieldID)).
				ForUpdate().
				All(ctx)
			if err != nil {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
			}
		}

		var organizationBucket, userBucket *ent.MailSendRateBucket
		if terminalCode == "" {
			organizationBucket, err = mailboxEntSubmissionLockRateBucket(ctx, transaction, mailsendratebucket.Scope("organization_instance"), account.OrganizationID)
			if err != nil {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
			}
			userBucket, err = mailboxEntSubmissionLockRateBucket(ctx, transaction, mailsendratebucket.Scope("user"), account.UserID)
			if err != nil {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
			}
		}

		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionClaimRepositoryError(err)
		}
		if !mailboxEntSubmissionAuthorizingFenceMatches(account, submission, claim, databaseNow) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}
		if terminalCode != "" {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, terminalCode); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(terminalCode)
			return nil
		}

		binding, ok := mailboxEntSubmissionBindingFromRow(submission)
		if !ok {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeRejected); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeRejected)
			return nil
		}
		persistedSubmission := mailboxEntSubmissionEngineSubmission(binding)
		if persistedSubmission != claim.submission || !mailboxEntSubmissionClaimDraftMatches(draft, account.ID, *submission.MailDraftID, binding) {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeRejected); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeRejected)
			return nil
		}
		if !mailboxEntSubmissionIdentityMatchesPreflight(identity, account.ID, preflight.FromEmail) {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeIdentityUnavailable); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeIdentityUnavailable)
			return nil
		}
		preflightBinding := submissionPreflightBindingHash(persistedSubmission, preflight.FromEmail, preflight.RecipientCount, preflight.MessageBytes)
		if subtle.ConstantTimeCompare(preflight.BindingHash[:], preflightBinding[:]) != 1 {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeRejected); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeRejected)
			return nil
		}
		freshPolicy, policyErr := mailboxEntSubmissionClaimPolicyBinding(account, policy, domains)
		if policyErr != nil {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodePolicy); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodePolicy)
			return nil
		}
		freshRequest, valid := freshPolicy.PreflightRequest(persistedSubmission)
		if !valid {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodePolicy); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodePolicy)
			return nil
		}
		semanticBinding := submissionPreflightPolicyBindingHash(freshRequest)
		if subtle.ConstantTimeCompare(preflight.PolicyBindingHash[:], semanticBinding[:]) != 1 {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodePolicy); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodePolicy)
			return nil
		}
		claimPolicyHash := claim.policy.Hash()
		freshPolicyHash := freshPolicy.Hash()
		if subtle.ConstantTimeCompare(claimPolicyHash[:], freshPolicyHash[:]) != 1 {
			if err := mailboxEntSubmissionRequeue(ctx, account, submission, draft, claim, databaseNow, submission.AttemptCount-1, databaseNow); err != nil {
				return err
			}
			result = mailboxEntSubmissionRequeuedResult(databaseNow)
			return nil
		}

		cleanAttachmentProof, err := MailboxSubmissionCleanAttachmentProofHash(attachments)
		if err != nil {
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeAttachmentNotClean); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeAttachmentNotClean)
			return nil
		}
		organizationState, organizationMissing, valid := mailboxEntSubmissionRateBucketState(organizationBucket, mailsendratebucket.Scope("organization_instance"), account.OrganizationID)
		if !valid {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		userState, userMissing, valid := mailboxEntSubmissionRateBucketState(userBucket, mailsendratebucket.Scope("user"), account.UserID)
		if !valid {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		// Exactly one policy evaluator call decides both buckets all-or-nothing.
		rateDecision, err := EvaluateMailboxSubmissionRateLimit(MailboxSubmissionRateLimitInput{
			Now:                  databaseNow,
			OrganizationCapacity: policy.InstanceSendRatePerMinute,
			UserCapacity:         policy.SendRatePerMinute,
			Organization:         organizationState,
			User:                 userState,
		})
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if err := mailboxEntSubmissionPersistRateBucket(ctx, transaction, organizationBucket, organizationMissing, mailsendratebucket.Scope("organization_instance"), account.OrganizationID, rateDecision.Organization, databaseNow); err != nil {
			return err
		}
		if err := mailboxEntSubmissionPersistRateBucket(ctx, transaction, userBucket, userMissing, mailsendratebucket.Scope("user"), account.UserID, rateDecision.User, databaseNow); err != nil {
			return err
		}
		if !rateDecision.Allowed {
			if rateDecision.RetryAt == nil || !mailboxEntValidTime(*rateDecision.RetryAt) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			if err := mailboxEntSubmissionRequeue(ctx, account, submission, draft, claim, databaseNow, submission.AttemptCount-1, *rateDecision.RetryAt); err != nil {
				return err
			}
			result = mailboxEntSubmissionRateLimitedResult(*rateDecision.RetryAt)
			return nil
		}

		authorizationHash, err := MailboxSubmissionDispatchAuthorizationHash(MailboxSubmissionDispatchAuthorizationHashInput{
			AccountID:                  account.ID,
			SubmissionID:               submission.ID,
			LeaseID:                    claim.fence.LeaseID,
			CommandFence:               claim.fence.MutationEpoch,
			LeaseUntil:                 claim.fence.LeaseUntil,
			AuthorizedAt:               databaseNow,
			EngineBindingHash:          binding.engineBindingHash,
			PreflightBindingHash:       preflight.BindingHash,
			PreflightPolicyBindingHash: preflight.PolicyBindingHash,
			DurablePolicyBindingHash:   freshPolicyHash,
			AttachmentCleanProofHash:   cleanAttachmentProof,
		})
		if err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if err := submission.Update().
			SetStatus(mailsubmission.Status("submitting")).
			SetDispatchAuthorizedAt(databaseNow).
			SetDispatchAuthorizationHash(append([]byte(nil), authorizationHash[:]...)).
			SetUpdatedAt(databaseNow).
			Exec(ctx); err != nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		authorization := MailboxSubmissionDispatchAuthorization{
			submissionID:      submission.ID,
			fence:             claim.fence,
			authorizedAt:      databaseNow,
			authorizationHash: authorizationHash,
		}
		if !mailboxSubmissionDispatchAuthorizationValid(authorization) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		result = MailboxSubmissionAuthorizeResult{outcome: MailboxSubmissionAuthorizeOutcomeAuthorized, authorization: mailboxSubmissionDispatchAuthorizationCopy(authorization)}
		return nil
	})
	if err != nil {
		return MailboxSubmissionAuthorizeResult{}, mailboxEntSubmissionClaimRepositoryError(err)
	}
	return result, nil
}

// ResolveMailboxSubmissionPreflightFailure resolves a read-only provider
// preflight result while the same dual lease is still live. It never reclaims
// stale state and performs no provider I/O.
func (repository *MailboxEntRepository) ResolveMailboxSubmissionPreflightFailure(ctx context.Context, request MailboxSubmissionPreflightFailureRequest) (MailboxSubmissionAuthorizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxSubmissionPreflightFailureRequestValid(request) {
		return MailboxSubmissionAuthorizeResult{}, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	claim := mailboxSubmissionClaimCopy(request.Claim)

	var result MailboxSubmissionAuthorizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		account, err := transaction.MailAccount.Query().Where(mailaccount.IDEQ(claim.fence.AccountID)).ForUpdate().Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		submission, err := transaction.MailSubmission.Query().
			Where(mailsubmission.IDEQ(claim.submissionID), mailsubmission.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		if submission.MailDraftID == nil {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		draft, err := transaction.MailDraft.Query().
			Where(maildraft.IDEQ(*submission.MailDraftID), maildraft.MailAccountIDEQ(account.ID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxEntSubmissionClaimLookupError(err)
		}
		databaseNow, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxEntSubmissionClaimRepositoryError(err)
		}
		if !mailboxEntSubmissionAuthorizingFenceMatches(account, submission, claim, databaseNow) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
		}

		switch request.Disposition {
		case MailboxSubmissionPreflightFailureDispositionFailed:
			if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, request.FailureCode); err != nil {
				return err
			}
			result = mailboxEntSubmissionFailedResult(request.FailureCode)
		case MailboxSubmissionPreflightFailureDispositionRetry:
			if submission.AttemptCount == 8 {
				if err := mailboxEntSubmissionTerminal(ctx, account, submission, draft, claim, databaseNow, MailboxSubmissionErrorCodeRetryExhausted); err != nil {
					return err
				}
				result = mailboxEntSubmissionFailedResult(MailboxSubmissionErrorCodeRetryExhausted)
				return nil
			}
			retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, time.Duration(request.RetryAfterSeconds)*time.Second)
			if !ok {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			if err := mailboxEntSubmissionRequeue(ctx, account, submission, draft, claim, databaseNow, submission.AttemptCount, retryAt); err != nil {
				return err
			}
			result = mailboxEntSubmissionRequeuedResult(retryAt)
		case MailboxSubmissionPreflightFailureDispositionRateLimited:
			retryAt, ok := mailboxEntCheckedTimeAdd(databaseNow, time.Duration(request.RetryAfterSeconds)*time.Second)
			if !ok || submission.AttemptCount < 1 {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
			}
			if err := mailboxEntSubmissionRequeue(ctx, account, submission, draft, claim, databaseNow, submission.AttemptCount-1, retryAt); err != nil {
				return err
			}
			result = mailboxEntSubmissionRateLimitedResult(retryAt)
		default:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		return nil
	})
	if err != nil {
		return MailboxSubmissionAuthorizeResult{}, mailboxEntSubmissionClaimRepositoryError(err)
	}
	return result, nil
}

func mailboxEntSubmissionClaimLookupError(err error) error {
	if ent.IsNotFound(err) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorNotFound)
	}
	return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
}

func mailboxEntSubmissionClaimAccountKind(account *ent.MailAccount) MailboxSubmissionRepositoryErrorKind {
	if account == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.UserID == uuid.Nil || account.MailDomainID == uuid.Nil ||
		account.Status != mailaccount.StatusActive || !account.DesiredEnabled || account.EngineAccountID == nil ||
		!syncValidOpaque(*account.EngineAccountID, MaxEngineOpaqueIDBytes) || account.MutationEpoch < 0 || !mailboxEntValidTime(account.SyncScheduleUpdatedAt) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	if _, _, valid := mailboxEntSyncLeaseFence(account); !valid ||
		(account.MutationLeaseToken == nil) != (account.MutationLeaseUntil == nil) ||
		(account.MutationLeaseUntil != nil && !mailboxEntValidTime(*account.MutationLeaseUntil)) {
		return MailboxSubmissionRepositoryErrorInvalid
	}
	return ""
}

// mailboxEntSubmissionClaimBusyUntil observes any stored lease, including an
// expired one. Expiry belongs to the dedicated stale-recovery path; claimers
// must never steal it. The furthest stored boundary is conservative if both
// scheduler and command leases are present.
func mailboxEntSubmissionClaimBusyUntil(account *ent.MailAccount) (time.Time, bool, bool) {
	if account == nil {
		return time.Time{}, false, false
	}
	var busyUntil *time.Time
	if account.MutationLeaseToken != nil {
		if account.MutationLeaseUntil == nil || *account.MutationLeaseToken == uuid.Nil || !mailboxEntValidTime(*account.MutationLeaseUntil) {
			return time.Time{}, false, false
		}
		value := *account.MutationLeaseUntil
		busyUntil = &value
	}
	_, hasSyncLease, syncValid := mailboxEntSyncLeaseFence(account)
	if !syncValid {
		return time.Time{}, false, false
	}
	if hasSyncLease {
		if account.SyncLeaseUntil == nil || !mailboxEntValidTime(*account.SyncLeaseUntil) {
			return time.Time{}, false, false
		}
		if busyUntil == nil || busyUntil.Before(*account.SyncLeaseUntil) {
			value := *account.SyncLeaseUntil
			busyUntil = &value
		}
	}
	if busyUntil == nil {
		return time.Time{}, false, true
	}
	return busyUntil.Round(0).UTC(), true, true
}

func mailboxEntSubmissionClaimDraftMatches(draft *ent.MailDraft, accountID, draftID uuid.UUID, binding mailboxEngineSubmissionBinding) bool {
	return draft != nil && mailboxEntDraftReadRowValid(draft, accountID, draftID) && draft.ID == draftID && draft.MailAccountID == accountID &&
		MailboxDraftStatus(draft.Status) == MailboxDraftStatusSubmitting && draft.Revision >= 1 && draft.EngineDraftKey == binding.engineDraftKey &&
		draft.EngineMessageID != nil && draft.EngineVersion != nil && *draft.EngineMessageID == binding.engineMessageID &&
		*draft.EngineVersion == binding.engineExpectedVersion && binding.engineAccountID != ""
}

func mailboxEntSubmissionEngineSubmission(binding mailboxEngineSubmissionBinding) EngineSubmission {
	return EngineSubmission{
		AccountID:       binding.engineAccountID,
		DraftKey:        binding.engineDraftKey,
		MessageID:       binding.engineMessageID,
		ExpectedVersion: binding.engineExpectedVersion,
		SubmissionKey:   binding.engineSubmissionKey,
		MaxMessageBytes: binding.engineMaxMessageBytes,
	}
}

func mailboxEntSubmissionClaimPolicyBinding(account *ent.MailAccount, policy *ent.MailboxPolicy, domains []*ent.MailDomain) (MailboxSubmissionPolicyBinding, error) {
	if account == nil || policy == nil || account.ID == uuid.Nil || account.OrganizationID == uuid.Nil || account.MailDomainID == uuid.Nil ||
		policy.ID == uuid.Nil || policy.OrganizationID != account.OrganizationID || policy.Version < 1 ||
		policy.MaxMessageBytes < 1_048_576 || policy.MaxMessageBytes > MaxEngineMessageBytes || policy.MaxRecipients < 1 ||
		policy.MaxRecipients > MaxEngineDraftTotalRecipientItems || policy.SendRatePerMinute < 1 || policy.SendRatePerMinute > mailboxSubmissionRateMaximumCapacity ||
		policy.InstanceSendRatePerMinute < 1 || policy.InstanceSendRatePerMinute > mailboxSubmissionRateMaximumCapacity {
		return MailboxSubmissionPolicyBinding{}, ErrMailboxSubmissionPolicyBindingInvalid
	}
	boundDomains := make([]MailboxSubmissionPolicyDomain, 0, len(domains))
	hasAccountDomain := false
	for _, domain := range domains {
		if domain == nil || domain.ID == uuid.Nil || domain.OrganizationID != account.OrganizationID || domain.Version < 1 || !domain.DesiredPresent || !domain.Enabled ||
			domain.State != maildomain.State("verified") || domain.VerificationStatus != maildomain.VerificationStatus("verified") || domain.DeletedAt != nil {
			return MailboxSubmissionPolicyBinding{}, ErrMailboxSubmissionPolicyBindingInvalid
		}
		boundDomains = append(boundDomains, MailboxSubmissionPolicyDomain{ID: domain.ID, Version: int64(domain.Version), Domain: domain.Domain})
		hasAccountDomain = hasAccountDomain || domain.ID == account.MailDomainID
	}
	if !hasAccountDomain {
		return MailboxSubmissionPolicyBinding{}, ErrMailboxSubmissionPolicyBindingInvalid
	}
	return NewMailboxSubmissionPolicyBinding(MailboxSubmissionPolicyBindingInput{
		OrganizationID:           account.OrganizationID,
		PolicyID:                 policy.ID,
		PolicyVersion:            int64(policy.Version),
		PolicyMaxMessageBytes:    policy.MaxMessageBytes,
		PolicyMaxRecipients:      policy.MaxRecipients,
		ExternalMailEnabled:      policy.ExternalMailEnabled,
		VerifiedEnabledDomains: boundDomains,
	})
}

func mailboxEntSubmissionAuthorizingFenceMatches(account *ent.MailAccount, submission *ent.MailSubmission, claim MailboxSubmissionClaim, databaseNow time.Time) bool {
	if account == nil || submission == nil || !mailboxSubmissionClaimValid(claim) || !mailboxEntValidTime(databaseNow) ||
		account.ID != claim.fence.AccountID || submission.ID != claim.submissionID || submission.MailAccountID != account.ID ||
		MailboxSubmissionStatus(submission.Status) != MailboxSubmissionStatusAuthorizing || submission.AttemptCount < 1 || submission.AttemptCount > 8 ||
		account.MutationLeaseToken == nil || account.MutationLeaseUntil == nil || submission.CommandLeaseToken == nil || submission.CommandLeaseExpiresAt == nil ||
		*account.MutationLeaseToken != claim.fence.LeaseID || *submission.CommandLeaseToken != claim.fence.LeaseID ||
		!account.MutationLeaseUntil.Equal(claim.fence.LeaseUntil) || !submission.CommandLeaseExpiresAt.Equal(claim.fence.LeaseUntil) ||
		account.MutationEpoch != claim.fence.MutationEpoch || submission.CommandFence != claim.fence.MutationEpoch ||
		!mailboxEntValidTime(*account.MutationLeaseUntil) || !mailboxEntValidTime(*submission.CommandLeaseExpiresAt) ||
		!account.MutationLeaseUntil.After(databaseNow) || !submission.CommandLeaseExpiresAt.After(databaseNow) ||
		submission.NextAttemptAt != nil || submission.DispatchAuthorizedAt != nil || submission.DispatchAuthorizationHash != nil ||
		submission.ProviderSubmissionID != nil || submission.ProviderMessageID != nil || submission.ResultMailMessageID != nil || submission.LastErrorCode != nil {
		return false
	}
	return true
}

func mailboxEntSubmissionIdentityMatchesPreflight(identity *ent.MailIdentity, accountID uuid.UUID, fromEmail string) bool {
	if identity == nil || identity.ID == uuid.Nil || identity.MailAccountID != accountID || !identity.SendAllowed ||
		identity.Status != mailidentity.Status("active") || identity.Kind != mailidentity.Kind("primary") && identity.Kind != mailidentity.Kind("alias") {
		return false
	}
	canonical, err := NormalizeAddress(EngineAddress{Email: identity.NormalizedAddress})
	return err == nil && canonical.Email == identity.NormalizedAddress && subtle.ConstantTimeCompare([]byte(identity.NormalizedAddress), []byte(fromEmail)) == 1
}

func mailboxEntSubmissionLockCleanAttachments(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, draft *ent.MailDraft) ([]MailboxSubmissionCleanAttachment, bool, error) {
	if ctx == nil || transaction == nil || account == nil || draft == nil || account.ID == uuid.Nil || draft.ID == uuid.Nil || draft.DraftClientID == uuid.Nil {
		return nil, false, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	rows, err := transaction.MailAttachment.Query().
		Where(mailattachment.Or(mailattachment.MailDraftIDEQ(draft.ID), mailattachment.PendingMailDraftIDEQ(draft.ID))).
		Order(ent.Asc(mailattachment.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, false, err
	}
	if len(rows) > maxMailboxSubmissionCleanAttachments {
		return nil, false, nil
	}
	objectIDs := make([]uuid.UUID, 0, len(rows))
	objectsSeen := make(map[uuid.UUID]struct{}, len(rows))
	for index, attachment := range rows {
		if !mailboxEntSubmissionCleanAttachmentRow(attachment, draft) || index > 0 && !syncLessUUID(rows[index-1].ID, attachment.ID) {
			return nil, false, nil
		}
		objectID := *attachment.StorageObjectID
		if _, found := objectsSeen[objectID]; !found {
			objectsSeen[objectID] = struct{}{}
			objectIDs = append(objectIDs, objectID)
		}
	}
	sort.Slice(objectIDs, func(left, right int) bool { return syncLessUUID(objectIDs[left], objectIDs[right]) })
	objects := make(map[uuid.UUID]*ent.MailObject, len(objectIDs))
	for _, objectID := range objectIDs {
		object, err := transaction.MailObject.Query().Where(mailobject.IDEQ(objectID)).ForUpdate().Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, false, nil
			}
			return nil, false, err
		}
		objects[objectID] = object
	}
	proof := make([]MailboxSubmissionCleanAttachment, 0, len(rows))
	for _, attachment := range rows {
		object := objects[*attachment.StorageObjectID]
		if !mailboxEntSubmissionCleanObjectRow(object, account, draft, attachment) {
			return nil, false, nil
		}
		var digest [sha256.Size]byte
		copy(digest[:], attachment.SHA256)
		proof = append(proof, MailboxSubmissionCleanAttachment{ID: attachment.ID, SHA256: digest})
	}
	return proof, true, nil
}

func mailboxEntSubmissionCleanAttachmentRow(attachment *ent.MailAttachment, draft *ent.MailDraft) bool {
	return attachment != nil && draft != nil && attachment.ID != uuid.Nil && attachment.MailMessageID == nil && attachment.MailDraftID != nil &&
		*attachment.MailDraftID == draft.ID && attachment.PendingMailDraftID == nil && attachment.DraftClientID != nil &&
		*attachment.DraftClientID == draft.DraftClientID && attachment.StorageObjectID != nil && *attachment.StorageObjectID != uuid.Nil &&
		attachment.EngineBlobID == nil && attachment.SizeBytes >= 1 && attachment.SizeBytes <= MaxEngineAttachmentBytes &&
		attachment.ScanStatus == mailattachment.ScanStatus("clean") && len(attachment.SHA256) == sha256.Size &&
		attachment.ScanLeaseToken == nil && attachment.ScanLeaseUntil == nil && attachment.ScanTerminalAt != nil &&
		mailboxEntDraftAttachmentProofSHA256Valid(attachment.SHA256) && mailboxEntApplyAttachmentOptionalTextValid(attachment.ContentID) &&
		mailboxEntApplyAttachmentTextValid(attachment.Filename, false) && mailboxEntApplyAttachmentTextValid(attachment.MediaType, false) &&
		mailboxEntApplyAttachmentDispositionValid(attachment.Disposition) &&
		(attachment.Disposition != mailattachment.Disposition(EngineAttachmentDispositionInline) || attachment.ContentID != nil)
}

func mailboxEntSubmissionCleanObjectRow(object *ent.MailObject, account *ent.MailAccount, draft *ent.MailDraft, attachment *ent.MailAttachment) bool {
	return object != nil && account != nil && draft != nil && attachment != nil && attachment.StorageObjectID != nil && object.ID == *attachment.StorageObjectID &&
		object.DeletionID != uuid.Nil && object.OrganizationID == account.OrganizationID && object.MailAccountID == account.ID &&
		object.UploaderUserID == account.UserID && object.DraftClientID == draft.DraftClientID && object.UploadState == mailobject.UploadState("confirmed") &&
		object.ExpectedSizeBytes == attachment.SizeBytes && object.ActualSizeBytes != nil && *object.ActualSizeBytes == object.ExpectedSizeBytes &&
		len(object.SHA256) == sha256.Size && subtle.ConstantTimeCompare(object.SHA256, attachment.SHA256) == 1 &&
		object.NextDeleteAt == nil && object.DeleteLeaseToken == nil && object.DeleteLeaseUntil == nil && object.DeleteTerminalAt == nil
}

func mailboxEntSubmissionLockRateBucket(ctx context.Context, transaction *ent.Tx, scope mailsendratebucket.Scope, ownerID uuid.UUID) (*ent.MailSendRateBucket, error) {
	if ctx == nil || transaction == nil || ownerID == uuid.Nil {
		return nil, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	query := transaction.MailSendRateBucket.Query().Where(mailsendratebucket.ScopeEQ(scope))
	switch scope {
	case mailsendratebucket.Scope("organization_instance"):
		query = query.Where(mailsendratebucket.OrganizationIDEQ(ownerID), mailsendratebucket.UserIDIsNil())
	case mailsendratebucket.Scope("user"):
		query = query.Where(mailsendratebucket.UserIDEQ(ownerID), mailsendratebucket.OrganizationIDIsNil())
	default:
		return nil, NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	row, err := query.ForUpdate().Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row, nil
}

func mailboxEntSubmissionRateBucketState(row *ent.MailSendRateBucket, scope mailsendratebucket.Scope, ownerID uuid.UUID) (MailboxSubmissionRateBucketState, bool, bool) {
	if row == nil {
		return MailboxSubmissionRateBucketState{}, true, true
	}
	if row.ID == uuid.Nil || row.Scope != scope || row.Tokens < 0 || row.Tokens > mailboxSubmissionRateMaximumCapacity || row.Version < 1 ||
		!mailboxEntValidTime(row.LastRefillAt) || row.RefillRemainderNanos < 0 || row.RefillRemainderNanos >= mailboxSubmissionRateMinuteNanos {
		return MailboxSubmissionRateBucketState{}, false, false
	}
	switch scope {
	case mailsendratebucket.Scope("organization_instance"):
		if row.OrganizationID == nil || *row.OrganizationID != ownerID || row.UserID != nil {
			return MailboxSubmissionRateBucketState{}, false, false
		}
	case mailsendratebucket.Scope("user"):
		if row.UserID == nil || *row.UserID != ownerID || row.OrganizationID != nil {
			return MailboxSubmissionRateBucketState{}, false, false
		}
	default:
		return MailboxSubmissionRateBucketState{}, false, false
	}
	return MailboxSubmissionRateBucketState{Tokens: row.Tokens, LastRefillAt: row.LastRefillAt, RefillRemainderNanos: row.RefillRemainderNanos, Version: row.Version}, false, true
}

func mailboxEntSubmissionPersistRateBucket(ctx context.Context, transaction *ent.Tx, row *ent.MailSendRateBucket, missing bool, scope mailsendratebucket.Scope, ownerID uuid.UUID, final MailboxSubmissionRateBucketState, databaseNow time.Time) error {
	if ctx == nil || transaction == nil || ownerID == uuid.Nil || !mailboxEntValidTime(databaseNow) || final.Version < 1 ||
		final.Tokens < 0 || final.Tokens > mailboxSubmissionRateMaximumCapacity || !mailboxEntValidTime(final.LastRefillAt) ||
		final.RefillRemainderNanos < 0 || final.RefillRemainderNanos >= mailboxSubmissionRateMinuteNanos {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if missing {
		if row != nil || final.Version != 1 {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		create := transaction.MailSendRateBucket.Create().
			SetScope(scope).
			SetTokens(final.Tokens).
			SetLastRefillAt(final.LastRefillAt).
			SetRefillRemainderNanos(final.RefillRemainderNanos).
			SetVersion(final.Version).
			SetCreatedAt(databaseNow).
			SetUpdatedAt(databaseNow)
		switch scope {
		case mailsendratebucket.Scope("organization_instance"):
			create.SetOrganizationID(ownerID)
		case mailsendratebucket.Scope("user"):
			create.SetUserID(ownerID)
		default:
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		if _, err := create.Save(ctx); err != nil {
			if ent.IsConstraintError(err) {
				return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
			}
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
		}
		return nil
	}
	if row == nil || final.Version < row.Version || final.Version > row.Version+1 {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	original := MailboxSubmissionRateBucketState{Tokens: row.Tokens, LastRefillAt: row.LastRefillAt, RefillRemainderNanos: row.RefillRemainderNanos, Version: row.Version}
	if final.Version == row.Version {
		if !mailboxSubmissionRateEqualPersistedState(original, final) {
			return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
		}
		return nil
	}
	if final.Version != row.Version+1 {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	updated, err := transaction.MailSendRateBucket.Update().
		Where(mailsendratebucket.IDEQ(row.ID), mailsendratebucket.VersionEQ(row.Version)).
		SetTokens(final.Tokens).
		SetLastRefillAt(final.LastRefillAt).
		SetRefillRemainderNanos(final.RefillRemainderNanos).
		SetVersion(final.Version).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if updated != 1 {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	return nil
}

func mailboxEntSubmissionDraftSubmittingMatches(account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft) bool {
	return account != nil && submission != nil && draft != nil && account.ID != uuid.Nil && submission.MailAccountID == account.ID &&
		submission.MailDraftID != nil && *submission.MailDraftID != uuid.Nil && draft.ID == *submission.MailDraftID &&
		draft.MailAccountID == account.ID && MailboxDraftStatus(draft.Status) == MailboxDraftStatusSubmitting
}

func mailboxEntSubmissionTerminal(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, claim MailboxSubmissionClaim, databaseNow time.Time, code MailboxSubmissionErrorCode) error {
	if !mailboxSubmissionAuthorizeFailureCodeValid(&code) || !mailboxEntSubmissionAuthorizingFenceMatches(account, submission, claim, databaseNow) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	if !mailboxEntSubmissionDraftSubmittingMatches(account, submission, draft) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := submission.Update().
		SetStatus(mailsubmission.Status("failed")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearDispatchAuthorizedAt().
		ClearDispatchAuthorizationHash().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		ClearNextAttemptAt().
		SetLastErrorCode(mailsubmission.LastErrorCode(code)).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := draft.Update().
		SetStatus(maildraft.Status("active")).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntSubmissionRequeue(ctx context.Context, account *ent.MailAccount, submission *ent.MailSubmission, draft *ent.MailDraft, claim MailboxSubmissionClaim, databaseNow time.Time, attempts int32, retryAt time.Time) error {
	if attempts < 0 || attempts > 7 || !mailboxEntValidTime(retryAt) || retryAt.Before(databaseNow) ||
		!mailboxEntSubmissionAuthorizingFenceMatches(account, submission, claim, databaseNow) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorVersionConflict)
	}
	if !mailboxEntSubmissionDraftSubmittingMatches(account, submission, draft) {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	nextEpoch, ok := mailboxEntNextEpoch(account.MutationEpoch)
	if !ok {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorInvalid)
	}
	if err := submission.Update().
		SetStatus(mailsubmission.Status("queued")).
		ClearCommandLeaseToken().
		ClearCommandLeaseExpiresAt().
		ClearDispatchAuthorizedAt().
		ClearDispatchAuthorizationHash().
		ClearProviderSubmissionID().
		ClearProviderMessageID().
		ClearResultMailMessageID().
		ClearLastErrorCode().
		SetNextAttemptAt(retryAt).
		SetAttemptCount(attempts).
		SetUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	if err := account.Update().
		ClearMutationLeaseToken().
		ClearMutationLeaseUntil().
		SetMutationEpoch(nextEpoch).
		SetSyncScheduleUpdatedAt(databaseNow).
		Exec(ctx); err != nil {
		return NewMailboxSubmissionRepositoryError(MailboxSubmissionRepositoryErrorUnavailable)
	}
	return nil
}

func mailboxEntSubmissionFailedResult(code MailboxSubmissionErrorCode) MailboxSubmissionAuthorizeResult {
	copy := code
	return MailboxSubmissionAuthorizeResult{outcome: MailboxSubmissionAuthorizeOutcomeFailed, failureCode: &copy}
}

func mailboxEntSubmissionRequeuedResult(retryAt time.Time) MailboxSubmissionAuthorizeResult {
	copy := retryAt.Round(0).UTC()
	return MailboxSubmissionAuthorizeResult{outcome: MailboxSubmissionAuthorizeOutcomeRequeued, retryAt: &copy}
}

func mailboxEntSubmissionRateLimitedResult(retryAt time.Time) MailboxSubmissionAuthorizeResult {
	copy := retryAt.Round(0).UTC()
	return MailboxSubmissionAuthorizeResult{outcome: MailboxSubmissionAuthorizeOutcomeRateLimited, retryAt: &copy}
}
