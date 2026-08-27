package mail

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

const (
	// MailboxAccountProvisioningDefaultBatchSize bounds a normal due scan.
	MailboxAccountProvisioningDefaultBatchSize = 20
	// MailboxAccountProvisioningMaximumBatchSize is the hard due-scan cap.
	MailboxAccountProvisioningMaximumBatchSize = 50
	// MailboxAccountProvisioningDefaultScanInterval is the normal empty-scan delay.
	MailboxAccountProvisioningDefaultScanInterval = time.Minute

	mailboxAccountProvisioningRetryInitialDelay = 30 * time.Second
	mailboxAccountProvisioningRetryMaximumDelay = 5 * time.Minute
)

// MailboxAccountProvisioningCode is a closed content-free provisioning code.
type MailboxAccountProvisioningCode string

const (
	// MailboxAccountProvisioningCodeDomainRequired says domain control must settle first.
	MailboxAccountProvisioningCodeDomainRequired MailboxAccountProvisioningCode = "MAILBOX_DOMAIN_REQUIRED"
	// MailboxAccountProvisioningCodeEngineDisabled says the engine gate is disabled locally.
	MailboxAccountProvisioningCodeEngineDisabled MailboxAccountProvisioningCode = "MAILBOX_ENGINE_DISABLED"
	// MailboxAccountProvisioningCodeProvisioningRetry says provider work may be retried.
	MailboxAccountProvisioningCodeProvisioningRetry MailboxAccountProvisioningCode = "MAILBOX_PROVISIONING_RETRY"
	// MailboxAccountProvisioningCodeUnavailable says the provider result is unavailable.
	MailboxAccountProvisioningCodeUnavailable MailboxAccountProvisioningCode = "MAILBOX_UNAVAILABLE"
)

// MailboxAccountProvisioningScanRequest is one bounded due-account scan request.
type MailboxAccountProvisioningScanRequest struct {
	BatchSize    int
	ScanInterval time.Duration
}

// NormalizeMailboxAccountProvisioningScanRequest supplies the frozen defaults and rejects invalid bounds.
func NormalizeMailboxAccountProvisioningScanRequest(request MailboxAccountProvisioningScanRequest) (MailboxAccountProvisioningScanRequest, bool) {
	if request.BatchSize == 0 {
		request.BatchSize = MailboxAccountProvisioningDefaultBatchSize
	}
	if request.ScanInterval == 0 {
		request.ScanInterval = MailboxAccountProvisioningDefaultScanInterval
	}
	if request.BatchSize < 1 || request.BatchSize > MailboxAccountProvisioningMaximumBatchSize || request.ScanInterval < 0 {
		return MailboxAccountProvisioningScanRequest{}, false
	}
	return request, true
}

// MailboxAccountProvisioningScanResult is the bounded list of due account UUIDs.
type MailboxAccountProvisioningScanResult struct {
	AccountIDs []uuid.UUID
}

// ValidFor reports whether the result preserves the normalized scan bound.
func (result MailboxAccountProvisioningScanResult) ValidFor(request MailboxAccountProvisioningScanRequest) bool {
	normalized, ok := NormalizeMailboxAccountProvisioningScanRequest(request)
	if !ok || len(result.AccountIDs) > normalized.BatchSize {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.AccountIDs))
	for _, accountID := range result.AccountIDs {
		if accountID == uuid.Nil {
			return false
		}
		if _, exists := seen[accountID]; exists {
			return false
		}
		seen[accountID] = struct{}{}
	}
	return true
}

// MailboxAccountProvisioningClaimRequest selects one account for a durable claim.
type MailboxAccountProvisioningClaimRequest struct {
	AccountID uuid.UUID
}

// Valid reports whether the request identifies exactly one account.
func (request MailboxAccountProvisioningClaimRequest) Valid() bool {
	return request.AccountID != uuid.Nil
}

// MailboxAccountProvisioningFence is the exact durable control and lease fence for one provider call.
type MailboxAccountProvisioningFence struct {
	AccountID          uuid.UUID
	OrganizationID     uuid.UUID
	DomainID           uuid.UUID
	UserID             uuid.UUID
	LeaseToken         uuid.UUID
	LeaseDeadline      time.Time
	LeaseDuration      time.Duration
	ControlGeneration  int64
	ControlHashVersion int32
	ControlHash        string
	ProvisioningKey    string
	LocalPart          string
	Domain             string
}

// Valid reports whether every field is safe and exact enough to fence provider work.
func (fence MailboxAccountProvisioningFence) Valid() bool {
	if fence.AccountID == uuid.Nil || fence.OrganizationID == uuid.Nil || fence.DomainID == uuid.Nil || fence.UserID == uuid.Nil || fence.LeaseToken == uuid.Nil {
		return false
	}
	if fence.LeaseDuration != MailboxEngineCommandLeaseDuration || !mailboxAccountProvisioningCanonicalSecond(fence.LeaseDeadline) {
		return false
	}
	if fence.ControlGeneration < 1 || fence.ControlHashVersion < 1 || !mailboxAccountProvisioningSafeText(fence.ControlHash) || !mailboxAccountProvisioningSafeText(fence.ProvisioningKey) {
		return false
	}
	if len(fence.LocalPart) == 0 || len(fence.LocalPart) > 64 || len(fence.Domain) == 0 || len(fence.Domain) > 253 {
		return false
	}
	local := validCanonicalRegistryLocalPart(fence.LocalPart)
	return local && validCanonicalRegistryDomain(fence.Domain)
}

func mailboxAccountProvisioningCanonicalSecond(value time.Time) bool {
	return !value.IsZero() && value == value.UTC().Truncate(time.Second)
}

func mailboxAccountProvisioningSafeText(value string) bool {
	return value != "" && validSafeText(value, MaxEngineOpaqueIDBytes, true)
}

// MailboxAccountProvisioningClaim is immutable provider input derived from one sealed claim fence.
type MailboxAccountProvisioningClaim struct {
	fence MailboxAccountProvisioningFence
}

// NewMailboxAccountProvisioningClaim seals a validated claim fence before engine I/O.
func NewMailboxAccountProvisioningClaim(fence MailboxAccountProvisioningFence) (MailboxAccountProvisioningClaim, bool) {
	if !fence.Valid() {
		return MailboxAccountProvisioningClaim{}, false
	}
	return MailboxAccountProvisioningClaim{fence: fence}, true
}

// Valid reports whether the claimed provider input still carries a valid sealed fence.
func (claim MailboxAccountProvisioningClaim) Valid() bool {
	return claim.fence.Valid()
}

// Fence returns a value copy of the exact sealed claim fence.
func (claim MailboxAccountProvisioningClaim) Fence() MailboxAccountProvisioningFence {
	return claim.fence
}

// AccountSpec returns the exact credential-free provider input for one EnsureAccount call.
func (claim MailboxAccountProvisioningClaim) AccountSpec() AccountSpec {
	if !claim.Valid() {
		return AccountSpec{}
	}
	return AccountSpec{
		ProvisioningKey:   claim.fence.ProvisioningKey,
		LocalPart:         claim.fence.LocalPart,
		Domain:            claim.fence.Domain,
		DesiredGeneration: claim.fence.ControlGeneration,
		DesiredHash:       claim.fence.ControlHash,
		Enabled:           true,
		Primary: AccountIdentitySpec{
			LocalPart: claim.fence.LocalPart,
			Domain:    claim.fence.Domain,
			Enabled:   true,
		},
	}
}

// MailboxAccountProvisioningClaimOutcome is a closed durable claim result.
type MailboxAccountProvisioningClaimOutcome string

const (
	// MailboxAccountProvisioningClaimed carries one provider-call claim.
	MailboxAccountProvisioningClaimed MailboxAccountProvisioningClaimOutcome = "claimed"
	// MailboxAccountProvisioningNoop says durable state already needs no provider call.
	MailboxAccountProvisioningNoop MailboxAccountProvisioningClaimOutcome = "noop"
	// MailboxAccountProvisioningLocalGate says a local prerequisite blocked provider work.
	MailboxAccountProvisioningLocalGate MailboxAccountProvisioningClaimOutcome = "local_gate"
)

// MailboxAccountProvisioningClaimResult is the closed result of one durable claim attempt.
type MailboxAccountProvisioningClaimResult struct {
	Outcome MailboxAccountProvisioningClaimOutcome
	Claim   MailboxAccountProvisioningClaim
	Code    MailboxAccountProvisioningCode
}

// Valid reports whether the result contains only its allowed branch data.
func (result MailboxAccountProvisioningClaimResult) Valid() bool {
	switch result.Outcome {
	case MailboxAccountProvisioningClaimed:
		return result.Claim.Valid() && result.Code == ""
	case MailboxAccountProvisioningNoop:
		return result.Claim == (MailboxAccountProvisioningClaim{}) && result.Code == ""
	case MailboxAccountProvisioningLocalGate:
		return result.Claim == (MailboxAccountProvisioningClaim{}) && validMailboxAccountProvisioningLocalGateCode(result.Code)
	default:
		return false
	}
}

func validMailboxAccountProvisioningLocalGateCode(code MailboxAccountProvisioningCode) bool {
	return code == MailboxAccountProvisioningCodeDomainRequired || code == MailboxAccountProvisioningCodeEngineDisabled
}

// MailboxAccountProvisioningFinalizeOutcome is the closed durable provider finalization branch.
type MailboxAccountProvisioningFinalizeOutcome string

const (
	// MailboxAccountProvisioningFinalizeSuccess records one accepted provider success.
	MailboxAccountProvisioningFinalizeSuccess MailboxAccountProvisioningFinalizeOutcome = "success"
	// MailboxAccountProvisioningFinalizeRetry records one safe provider retry.
	MailboxAccountProvisioningFinalizeRetry MailboxAccountProvisioningFinalizeOutcome = "retry"
)

// MailboxAccountProvisioningFinalizeRequest is one exact claimed fence finalization request.
type MailboxAccountProvisioningFinalizeRequest struct {
	Claim           MailboxAccountProvisioningClaim
	Outcome         MailboxAccountProvisioningFinalizeOutcome
	EngineAccountID string
	Code            MailboxAccountProvisioningCode
	account         EngineAccount
}

// NewMailboxAccountProvisioningFinalizeSuccess seals the only accepted Registry success shape.
func NewMailboxAccountProvisioningFinalizeSuccess(claim MailboxAccountProvisioningClaim, account EngineAccount) (MailboxAccountProvisioningFinalizeRequest, bool) {
	if !claim.Valid() || !mailboxAccountProvisioningRegistrySuccess(claim, account) {
		return MailboxAccountProvisioningFinalizeRequest{}, false
	}
	return MailboxAccountProvisioningFinalizeRequest{
		Claim:           claim,
		Outcome:         MailboxAccountProvisioningFinalizeSuccess,
		EngineAccountID: account.AccountID,
		account:         account,
	}, true
}

func mailboxAccountProvisioningRegistrySuccess(claim MailboxAccountProvisioningClaim, account EngineAccount) bool {
	return mailboxAccountProvisioningSafeText(account.AccountID) &&
		account.State == EngineAccountProvisioning &&
		account.AppliedGeneration == claim.fence.ControlGeneration &&
		account.AppliedHash == claim.fence.ControlHash &&
		account.PrimaryIdentityID == "" &&
		account.AliasIdentityIDs != nil && len(account.AliasIdentityIDs) == 0
}

// NewMailboxAccountProvisioningFinalizeRetry seals a retry against the exact claimed fence.
func NewMailboxAccountProvisioningFinalizeRetry(claim MailboxAccountProvisioningClaim, code MailboxAccountProvisioningCode) (MailboxAccountProvisioningFinalizeRequest, bool) {
	if !claim.Valid() || !validMailboxAccountProvisioningRetryCode(code) {
		return MailboxAccountProvisioningFinalizeRequest{}, false
	}
	return MailboxAccountProvisioningFinalizeRequest{
		Claim:   claim,
		Outcome: MailboxAccountProvisioningFinalizeRetry,
		Code:    code,
	}, true
}

// Valid reports whether the request reuses a valid sealed fence and branch-specific safe data.
func (request MailboxAccountProvisioningFinalizeRequest) Valid() bool {
	if !request.Claim.Valid() {
		return false
	}
	switch request.Outcome {
	case MailboxAccountProvisioningFinalizeSuccess:
		return request.Code == "" && request.EngineAccountID == request.account.AccountID && mailboxAccountProvisioningRegistrySuccess(request.Claim, request.account)
	case MailboxAccountProvisioningFinalizeRetry:
		return request.EngineAccountID == "" && validMailboxAccountProvisioningRetryCode(request.Code)
	default:
		return false
	}
}

// MailboxAccountProvisioningFinalizeResult is the closed result of durable finalization.
type MailboxAccountProvisioningFinalizeResult struct {
	Outcome MailboxAccountProvisioningFinalizeOutcome
	Code    MailboxAccountProvisioningCode
}

// Valid reports whether finalization recorded only its branch-specific safe data.
func (result MailboxAccountProvisioningFinalizeResult) Valid() bool {
	switch result.Outcome {
	case MailboxAccountProvisioningFinalizeSuccess:
		return result.Code == ""
	case MailboxAccountProvisioningFinalizeRetry:
		return validMailboxAccountProvisioningRetryCode(result.Code)
	default:
		return false
	}
}

func validMailboxAccountProvisioningRetryCode(code MailboxAccountProvisioningCode) bool {
	return code == MailboxAccountProvisioningCodeProvisioningRetry || code == MailboxAccountProvisioningCodeUnavailable
}

// MailboxAccountProvisioningRepository is the bounded durable port consumed by a future worker.
type MailboxAccountProvisioningRepository interface {
	Scan(context.Context, MailboxAccountProvisioningScanRequest) (MailboxAccountProvisioningScanResult, error)
	ClaimMailboxAccountProvisioning(context.Context, MailboxAccountProvisioningClaimRequest) (MailboxAccountProvisioningClaimResult, error)
	Finalize(context.Context, MailboxAccountProvisioningFinalizeRequest) (MailboxAccountProvisioningFinalizeResult, error)
}

// MailboxAccountProvisioningProviderErrorCode maps all provider failures to one closed safe retry code.
func MailboxAccountProvisioningProviderErrorCode(err error) MailboxAccountProvisioningCode {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, ErrEngineUnavailable) || errors.Is(err, ErrEngineIndeterminate) || errors.Is(err, ErrEngineRestartRequired) {
		return MailboxAccountProvisioningCodeUnavailable
	}
	return MailboxAccountProvisioningCodeProvisioningRetry
}

// MailboxAccountProvisioningRetryDelay returns a capped retry delay without terminal attempts.
func MailboxAccountProvisioningRetryDelay(attempt int) (time.Duration, bool) {
	if attempt < 1 || attempt >= math.MaxInt32 {
		return 0, false
	}
	if attempt >= 5 {
		return mailboxAccountProvisioningRetryMaximumDelay, true
	}
	return mailboxAccountProvisioningRetryInitialDelay * time.Duration(1<<(attempt-1)), true
}
