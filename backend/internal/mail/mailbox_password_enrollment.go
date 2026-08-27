package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	`github.com/valerii/hub/backend/internal/mail/stalwart`
)

const (
	mailboxPasswordEnrollmentLeaseSafetyMargin = time.Second
	mailboxPasswordEnrollmentCleanupTimeout   = 5 * time.Second
	mailboxPasswordEnrollmentMinimumBytes     = 12
	mailboxPasswordEnrollmentMaximumBytes     = 256
)

// RegistryPasswordPort performs the one authorized Registry password
// replacement. Its caller retains ownership of the supplied buffer.
type RegistryPasswordPort interface {
	SetUserPassword(context.Context, string, []byte) error
}

// MailboxPasswordEnrollmentResult is the only successful public enrollment
// outcome. It deliberately carries no account, lease, or provider detail.
type MailboxPasswordEnrollmentResult string

const (
	MailboxPasswordEnrollmentResultEnrolled        MailboxPasswordEnrollmentResult = "enrolled"
	MailboxPasswordEnrollmentResultAlreadyEnrolled MailboxPasswordEnrollmentResult = "already_enrolled"
	MailboxPasswordEnrollmentResultBusy            MailboxPasswordEnrollmentResult = "busy"
)

// MailboxPasswordEnrollmentClaimOutcome is the closed result of an enrollment
// lease request.
type MailboxPasswordEnrollmentClaimOutcome string

const (
	MailboxPasswordEnrollmentClaimOutcomeNoop    MailboxPasswordEnrollmentClaimOutcome = "noop"
	MailboxPasswordEnrollmentClaimOutcomeBusy    MailboxPasswordEnrollmentClaimOutcome = "busy"
	MailboxPasswordEnrollmentClaimOutcomeClaimed MailboxPasswordEnrollmentClaimOutcome = "claimed"
)

// MailboxPasswordEnrollmentClaimedInput is accepted only by the claimed-result
// factory at the repository boundary. The service never accepts these values
// from a caller and the sealed claim exposes none of them.
type MailboxPasswordEnrollmentClaimedInput struct {
	AccountID           uuid.UUID
	OrganizationID      uuid.UUID
	EngineAccountID     string
	DeployedGeneration  int64
	DeployedHashVersion int32
	DeployedHash        [sha256.Size]byte
	Generation          int64
	LeaseToken          uuid.UUID
	LeaseUntil          time.Time
}

// MailboxPasswordEnrollmentClaim is a sealed lease capability. Account and
// deployment identifiers stay private so they cannot escape in public results
// or errors.
type MailboxPasswordEnrollmentClaim struct {
	accountID           uuid.UUID
	organizationID      uuid.UUID
	engineAccountID     string
	deployedGeneration  int64
	deployedHashVersion int32
	deployedHash        [sha256.Size]byte
	generation          int64
	leaseToken          uuid.UUID
	leaseUntil          time.Time
}

// MailboxPasswordEnrollmentClaimResult carries exactly one noop, busy, or
// sealed claimed outcome.
type MailboxPasswordEnrollmentClaimResult struct {
	outcome MailboxPasswordEnrollmentClaimOutcome
	claim   MailboxPasswordEnrollmentClaim
}

// NewMailboxPasswordEnrollmentNoopResult constructs the canonical durable
// already-enrolled outcome.
func NewMailboxPasswordEnrollmentNoopResult() MailboxPasswordEnrollmentClaimResult {
	return MailboxPasswordEnrollmentClaimResult{outcome: MailboxPasswordEnrollmentClaimOutcomeNoop}
}

// NewMailboxPasswordEnrollmentBusyResult constructs the canonical live-lease
// outcome without exposing lease metadata.
func NewMailboxPasswordEnrollmentBusyResult() MailboxPasswordEnrollmentClaimResult {
	return MailboxPasswordEnrollmentClaimResult{outcome: MailboxPasswordEnrollmentClaimOutcomeBusy}
}

// NewMailboxPasswordEnrollmentClaimedResult validates and seals the exact
// account, organization, Registry identity, deployed configuration,
// generation, and lease tuple.
func NewMailboxPasswordEnrollmentClaimedResult(input MailboxPasswordEnrollmentClaimedInput) (MailboxPasswordEnrollmentClaimResult, error) {
	claim := mailboxPasswordEnrollmentClaimFromInput(input)
	if !mailboxPasswordEnrollmentClaimValid(claim) {
		return MailboxPasswordEnrollmentClaimResult{}, newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	return MailboxPasswordEnrollmentClaimResult{
		outcome: MailboxPasswordEnrollmentClaimOutcomeClaimed,
		claim:   mailboxPasswordEnrollmentClaimCopy(claim),
	}, nil
}

// Outcome returns the closed claim result, or the zero value for malformed
// repository output.
func (result MailboxPasswordEnrollmentClaimResult) Outcome() MailboxPasswordEnrollmentClaimOutcome {
	if !mailboxPasswordEnrollmentClaimResultValid(result) {
		return ""
	}
	return result.outcome
}

// Claim returns only a defensive sealed capability for a valid claimed result.
func (result MailboxPasswordEnrollmentClaimResult) Claim() (MailboxPasswordEnrollmentClaim, bool) {
	if !mailboxPasswordEnrollmentClaimResultValid(result) || result.outcome != MailboxPasswordEnrollmentClaimOutcomeClaimed {
		return MailboxPasswordEnrollmentClaim{}, false
	}
	return mailboxPasswordEnrollmentClaimCopy(result.claim), true
}

// MailboxPasswordEnrollmentReleaseCode is the closed post-claim conclusion
// following an unsuccessful Registry mutation attempt.
type MailboxPasswordEnrollmentReleaseCode string

const (
	MailboxPasswordEnrollmentReleaseCodeTransient     MailboxPasswordEnrollmentReleaseCode = "transient"
	MailboxPasswordEnrollmentReleaseCodeIndeterminate MailboxPasswordEnrollmentReleaseCode = "indeterminate"
	MailboxPasswordEnrollmentReleaseCodeRejected      MailboxPasswordEnrollmentReleaseCode = "rejected"
)

// MailboxPasswordEnrollmentRepository owns durable claim fencing, exact
// completion, and release. It never performs Registry I/O.
type MailboxPasswordEnrollmentRepository interface {
	Claim(context.Context, uuid.UUID, uuid.UUID) (MailboxPasswordEnrollmentClaimResult, error)
	Complete(context.Context, MailboxPasswordEnrollmentClaim, time.Time) error
	Release(context.Context, MailboxPasswordEnrollmentClaim, MailboxPasswordEnrollmentReleaseCode) error
}

// MailboxPasswordEnrollmentErrorKind is the closed content-free error family
// for enrollment failures.
type MailboxPasswordEnrollmentErrorKind string

const (
	MailboxPasswordEnrollmentErrorInvalid               MailboxPasswordEnrollmentErrorKind = "invalid"
	MailboxPasswordEnrollmentErrorRepositoryUnavailable MailboxPasswordEnrollmentErrorKind = "repository_unavailable"
	MailboxPasswordEnrollmentErrorRetryable             MailboxPasswordEnrollmentErrorKind = "retryable"
)

const mailboxPasswordEnrollmentErrorText = "mailbox password enrollment error"

// MailboxPasswordEnrollmentError never retains or exposes an account ID,
// Registry response, secret, lease, or panic value.
type MailboxPasswordEnrollmentError struct {
	kind MailboxPasswordEnrollmentErrorKind
}

func (errorValue *MailboxPasswordEnrollmentError) Error() string {
	return mailboxPasswordEnrollmentErrorText
}

// Kind returns a stable safe classification.
func (errorValue *MailboxPasswordEnrollmentError) Kind() MailboxPasswordEnrollmentErrorKind {
	if errorValue == nil || !validMailboxPasswordEnrollmentErrorKind(errorValue.kind) {
		return MailboxPasswordEnrollmentErrorInvalid
	}
	return errorValue.kind
}

// MailboxPasswordEnrollmentErrorKindOf extracts a valid closed error kind.
func MailboxPasswordEnrollmentErrorKindOf(err error) (MailboxPasswordEnrollmentErrorKind, bool) {
	var enrollmentError *MailboxPasswordEnrollmentError
	if !errors.As(err, &enrollmentError) || enrollmentError == nil || enrollmentError.Kind() == MailboxPasswordEnrollmentErrorInvalid {
		return MailboxPasswordEnrollmentErrorInvalid, false
	}
	return enrollmentError.Kind(), true
}

func newMailboxPasswordEnrollmentError(kind MailboxPasswordEnrollmentErrorKind) *MailboxPasswordEnrollmentError {
	if !validMailboxPasswordEnrollmentErrorKind(kind) {
		kind = MailboxPasswordEnrollmentErrorInvalid
	}
	return &MailboxPasswordEnrollmentError{kind: kind}
}

func validMailboxPasswordEnrollmentErrorKind(kind MailboxPasswordEnrollmentErrorKind) bool {
	switch kind {
	case MailboxPasswordEnrollmentErrorInvalid,
		MailboxPasswordEnrollmentErrorRepositoryUnavailable,
		MailboxPasswordEnrollmentErrorRetryable:
		return true
	default:
		return false
	}
}

// MailboxPasswordEnrollmentService performs a one-shot replacement under one
// short durable claim. Replacement is idempotent, so an indeterminate outcome
// releases the marker for a later caller resubmission instead of inventing an
// enrolled-unknown state.
type MailboxPasswordEnrollmentService struct {
	repository MailboxPasswordEnrollmentRepository
	passwords  RegistryPasswordPort
}

// NewMailboxPasswordEnrollmentService requires exact durable and Registry
// ports. Construction performs neither durable nor provider I/O.
func NewMailboxPasswordEnrollmentService(repository MailboxPasswordEnrollmentRepository, passwords RegistryPasswordPort) (*MailboxPasswordEnrollmentService, error) {
	if mailboxSubmissionTypedNil(repository) || mailboxSubmissionTypedNil(passwords) {
		return nil, newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	return &MailboxPasswordEnrollmentService{repository: repository, passwords: passwords}, nil
}

// EnrollPassword replaces a claimed user's Registry password exactly once. It
// copies the caller-owned secret once, clears that owned buffer on every exit,
// and returns only closed public results or content-free typed errors.
func (service *MailboxPasswordEnrollmentService) EnrollPassword(ctx context.Context, actorID uuid.UUID, password []byte) (MailboxPasswordEnrollmentResult, error) {
	if service == nil || mailboxSubmissionTypedNil(service.repository) || mailboxSubmissionTypedNil(service.passwords) || ctx == nil || actorID == uuid.Nil || !validMailboxPasswordEnrollmentSecret(password) {
		return "", newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	if ctx.Err() != nil {
		return "", newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRetryable)
	}

	ownedPassword := append([]byte(nil), password...)
	defer clearMailboxPasswordEnrollmentSecret(ownedPassword)

	claimResult, claimFailed := service.safeClaim(ctx, actorID, uuid.New())
	if claimFailed {
		return "", newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRepositoryUnavailable)
	}
	switch claimResult.Outcome() {
	case MailboxPasswordEnrollmentClaimOutcomeNoop:
		return MailboxPasswordEnrollmentResultAlreadyEnrolled, nil
	case MailboxPasswordEnrollmentClaimOutcomeBusy:
		return MailboxPasswordEnrollmentResultBusy, nil
	case MailboxPasswordEnrollmentClaimOutcomeClaimed:
		// The sealed capability is extracted below.
	default:
		return "", newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	claim, claimed := claimResult.Claim()
	if !claimed {
		return "", newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	if ctx.Err() != nil {
		return "", service.release(ctx, claim, MailboxPasswordEnrollmentReleaseCodeTransient)
	}

	passwordContext, cancel, valid := mailboxPasswordEnrollmentExternalContext(ctx, claim.leaseUntil)
	if !valid {
		return "", service.release(ctx, claim, MailboxPasswordEnrollmentReleaseCodeTransient)
	}
	defer cancel()
	if passwordContext.Err() != nil {
		return "", service.release(ctx, claim, MailboxPasswordEnrollmentReleaseCodeTransient)
	}

	providerErr, providerPanic := service.safeSetUserPassword(passwordContext, claim.engineAccountID, ownedPassword)
	if providerPanic {
		return "", service.release(ctx, claim, MailboxPasswordEnrollmentReleaseCodeIndeterminate)
	}
	if providerErr != nil {
		return "", service.release(ctx, claim, mailboxPasswordEnrollmentReleaseCode(providerErr))
	}
	if err := service.complete(ctx, claim, time.Now().UTC().Truncate(time.Second)); err != nil {
		return "", err
	}
	return MailboxPasswordEnrollmentResultEnrolled, nil
}

func (service *MailboxPasswordEnrollmentService) complete(ctx context.Context, claim MailboxPasswordEnrollmentClaim, observedAt time.Time) error {
	if !mailboxPasswordEnrollmentClaimValid(claim) || !mailboxEntValidTime(observedAt) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	completeContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), mailboxPasswordEnrollmentCleanupTimeout)
	defer cancel()
	if service.safeComplete(completeContext, mailboxPasswordEnrollmentClaimCopy(claim), observedAt.Round(0).UTC()) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRetryable)
	}
	return nil
}

func (service *MailboxPasswordEnrollmentService) release(ctx context.Context, claim MailboxPasswordEnrollmentClaim, code MailboxPasswordEnrollmentReleaseCode) error {
	if !mailboxPasswordEnrollmentClaimValid(claim) || !validMailboxPasswordEnrollmentReleaseCode(code) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorInvalid)
	}
	releaseContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), mailboxPasswordEnrollmentCleanupTimeout)
	defer cancel()
	if service.safeRelease(releaseContext, mailboxPasswordEnrollmentClaimCopy(claim), code) {
		return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRepositoryUnavailable)
	}
	return newMailboxPasswordEnrollmentError(MailboxPasswordEnrollmentErrorRetryable)
}

func (service *MailboxPasswordEnrollmentService) safeClaim(ctx context.Context, actorID, leaseToken uuid.UUID) (result MailboxPasswordEnrollmentClaimResult, failed bool) {
	defer func() {
		if recover() != nil {
			result = MailboxPasswordEnrollmentClaimResult{}
			failed = true
		}
	}()
	result, err := service.repository.Claim(ctx, actorID, leaseToken)
	return result, err != nil
}

func (service *MailboxPasswordEnrollmentService) safeSetUserPassword(ctx context.Context, engineAccountID string, password []byte) (resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			resultErr = nil
			panicked = true
		}
	}()
	return service.passwords.SetUserPassword(ctx, engineAccountID, password), false
}

func (service *MailboxPasswordEnrollmentService) safeComplete(ctx context.Context, claim MailboxPasswordEnrollmentClaim, observedAt time.Time) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = true
		}
	}()
	return service.repository.Complete(ctx, claim, observedAt) != nil
}

func (service *MailboxPasswordEnrollmentService) safeRelease(ctx context.Context, claim MailboxPasswordEnrollmentClaim, code MailboxPasswordEnrollmentReleaseCode) (failed bool) {
	defer func() {
		if recover() != nil {
			failed = true
		}
	}()
	return service.repository.Release(ctx, claim, code) != nil
}

func mailboxPasswordEnrollmentExternalContext(ctx context.Context, leaseUntil time.Time) (context.Context, context.CancelFunc, bool) {
	if ctx == nil || ctx.Err() != nil || !mailboxEntValidTime(leaseUntil) {
		return nil, nil, false
	}
	deadline := leaseUntil.Add(-mailboxPasswordEnrollmentLeaseSafetyMargin)
	if !deadline.After(time.Now().UTC()) {
		return nil, nil, false
	}
	passwordContext, cancel := context.WithDeadline(ctx, deadline)
	if passwordContext.Err() != nil {
		cancel()
		return nil, nil, false
	}
	return passwordContext, cancel, true
}

func mailboxPasswordEnrollmentClaimFromInput(input MailboxPasswordEnrollmentClaimedInput) MailboxPasswordEnrollmentClaim {
	return MailboxPasswordEnrollmentClaim{
		accountID:           input.AccountID,
		organizationID:      input.OrganizationID,
		engineAccountID:     input.EngineAccountID,
		deployedGeneration:  input.DeployedGeneration,
		deployedHashVersion: input.DeployedHashVersion,
		deployedHash:        input.DeployedHash,
		generation:          input.Generation,
		leaseToken:          input.LeaseToken,
		leaseUntil:          input.LeaseUntil.Round(0).UTC(),
	}
}

func mailboxPasswordEnrollmentClaimValid(claim MailboxPasswordEnrollmentClaim) bool {
	return claim.accountID != uuid.Nil &&
		claim.organizationID != uuid.Nil &&
		validOpaqueID(claim.engineAccountID) &&
		claim.deployedGeneration >= 1 &&
		claim.deployedHashVersion >= 1 &&
		claim.generation >= 1 &&
		claim.leaseToken != uuid.Nil &&
		mailboxEntValidTime(claim.leaseUntil)
}

func mailboxPasswordEnrollmentClaimCopy(claim MailboxPasswordEnrollmentClaim) MailboxPasswordEnrollmentClaim {
	claim.leaseUntil = claim.leaseUntil.Round(0).UTC()
	return claim
}

func mailboxPasswordEnrollmentClaimZero(claim MailboxPasswordEnrollmentClaim) bool {
	return claim.accountID == uuid.Nil &&
		claim.organizationID == uuid.Nil &&
		claim.engineAccountID == "" &&
		claim.deployedGeneration == 0 &&
		claim.deployedHashVersion == 0 &&
		claim.deployedHash == ([sha256.Size]byte{}) &&
		claim.generation == 0 &&
		claim.leaseToken == uuid.Nil &&
		claim.leaseUntil.IsZero()
}

func mailboxPasswordEnrollmentClaimResultValid(result MailboxPasswordEnrollmentClaimResult) bool {
	switch result.outcome {
	case MailboxPasswordEnrollmentClaimOutcomeNoop, MailboxPasswordEnrollmentClaimOutcomeBusy:
		return mailboxPasswordEnrollmentClaimZero(result.claim)
	case MailboxPasswordEnrollmentClaimOutcomeClaimed:
		return mailboxPasswordEnrollmentClaimValid(result.claim)
	default:
		return false
	}
}

func validMailboxPasswordEnrollmentReleaseCode(code MailboxPasswordEnrollmentReleaseCode) bool {
	switch code {
	case MailboxPasswordEnrollmentReleaseCodeTransient,
		MailboxPasswordEnrollmentReleaseCodeIndeterminate,
		MailboxPasswordEnrollmentReleaseCodeRejected:
		return true
	default:
		return false
	}
}

func mailboxPasswordEnrollmentReleaseCode(err error) MailboxPasswordEnrollmentReleaseCode {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return MailboxPasswordEnrollmentReleaseCodeTransient
	}
	var engineError *EngineError
	if errors.As(err, &engineError) && engineError != nil {
		switch engineError.Kind {
		case EngineErrorRejectedPolicy, EngineErrorUnauthorizedAccount, EngineErrorNotFound, EngineErrorConflict, EngineErrorPayloadTooLarge, EngineErrorIdentityUnavailable:
			return MailboxPasswordEnrollmentReleaseCodeRejected
		case EngineErrorUnavailable, EngineErrorRestartRequired, EngineErrorRateLimited:
			return MailboxPasswordEnrollmentReleaseCodeTransient
		default:
			return MailboxPasswordEnrollmentReleaseCodeIndeterminate
		}
	}
	if providerKind, known := stalwartProviderErrorKind(err); known {
		switch providerKind {
		case stalwart.ErrorInvalidConfiguration, stalwart.ErrorInvalidRequest, stalwart.ErrorUnauthorized, stalwart.ErrorNotFound, stalwart.ErrorPayloadTooLarge, stalwart.ErrorUnsupported:
			return MailboxPasswordEnrollmentReleaseCodeRejected
		case stalwart.ErrorRestartRequired, stalwart.ErrorRateLimited:
			return MailboxPasswordEnrollmentReleaseCodeTransient
		default:
			return MailboxPasswordEnrollmentReleaseCodeIndeterminate
		}
	}
	return MailboxPasswordEnrollmentReleaseCodeIndeterminate
}

func validMailboxPasswordEnrollmentSecret(password []byte) bool {
	if len(password) < mailboxPasswordEnrollmentMinimumBytes || len(password) > mailboxPasswordEnrollmentMaximumBytes || !utf8.Valid(password) {
		return false
	}
	for len(password) != 0 {
		character, size := utf8.DecodeRune(password)
		if size == 0 || character == utf8.RuneError && size == 1 || unicode.IsControl(character) || (!unicode.IsGraphic(character) && !unicode.IsSpace(character)) {
			return false
		}
		password = password[size:]
	}
	return true
}

func clearMailboxPasswordEnrollmentSecret(password []byte) {
	for index := range password {
		password[index] = 0
	}
}
