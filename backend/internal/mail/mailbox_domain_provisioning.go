package mail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

const (
	// MailboxDomainProvisioningDefaultBatchSize bounds one normal due scan.
	MailboxDomainProvisioningDefaultBatchSize = 20
	// MailboxDomainProvisioningMaximumBatchSize is the immutable scan cap.
	MailboxDomainProvisioningMaximumBatchSize = 20
	// MailboxDomainProvisioningDefaultScanInterval is the normal empty-scan delay.
	MailboxDomainProvisioningDefaultScanInterval = time.Minute

	mailboxDomainProvisioningRetryInitialDelay = 30 * time.Second
	mailboxDomainProvisioningRetryMaximumDelay = 5 * time.Minute
)

// MailboxDomainProvisioningCode is a closed, schema-compatible retry code.
type MailboxDomainProvisioningCode string

const (
	MailboxDomainProvisioningCodeEngineUnavailable MailboxDomainProvisioningCode = "MAIL_DOMAIN_ENGINE_UNAVAILABLE"
	MailboxDomainProvisioningCodeApplyRejected     MailboxDomainProvisioningCode = "MAIL_DOMAIN_APPLY_REJECTED"
	MailboxDomainProvisioningCodeOutcomeUnknown    MailboxDomainProvisioningCode = "MAIL_DOMAIN_OUTCOME_UNKNOWN"
	MailboxDomainProvisioningCodeDNSInvalid        MailboxDomainProvisioningCode = "MAIL_DOMAIN_DNS_INVALID"
)

func validMailboxDomainProvisioningCode(code MailboxDomainProvisioningCode) bool {
	switch code {
	case MailboxDomainProvisioningCodeEngineUnavailable,
		MailboxDomainProvisioningCodeApplyRejected,
		MailboxDomainProvisioningCodeOutcomeUnknown,
		MailboxDomainProvisioningCodeDNSInvalid:
		return true
	default:
		return false
	}
}

// MailboxDomainProvisioningScanRequest selects a bounded due-domain batch.
type MailboxDomainProvisioningScanRequest struct {
	BatchSize    int
	ScanInterval time.Duration
}

// NormalizeMailboxDomainProvisioningScanRequest supplies fixed defaults and
// rejects values outside the bounded worker contract.
func NormalizeMailboxDomainProvisioningScanRequest(request MailboxDomainProvisioningScanRequest) (MailboxDomainProvisioningScanRequest, bool) {
	if request.BatchSize == 0 {
		request.BatchSize = MailboxDomainProvisioningDefaultBatchSize
	}
	if request.ScanInterval == 0 {
		request.ScanInterval = MailboxDomainProvisioningDefaultScanInterval
	}
	if request.BatchSize < 1 || request.BatchSize > MailboxDomainProvisioningMaximumBatchSize || request.ScanInterval < 0 {
		return MailboxDomainProvisioningScanRequest{}, false
	}
	return request, true
}

// MailboxDomainProvisioningScanResult is a bounded, duplicate-free domain list.
type MailboxDomainProvisioningScanResult struct {
	DomainIDs []uuid.UUID
}

// Clone returns an independent copy of the selected identifiers.
func (result MailboxDomainProvisioningScanResult) Clone() MailboxDomainProvisioningScanResult {
	return MailboxDomainProvisioningScanResult{DomainIDs: append([]uuid.UUID(nil), result.DomainIDs...)}
}

// ValidFor reports whether the scan result respects the normalized request.
func (result MailboxDomainProvisioningScanResult) ValidFor(request MailboxDomainProvisioningScanRequest) bool {
	normalized, ok := NormalizeMailboxDomainProvisioningScanRequest(request)
	if !ok || len(result.DomainIDs) > normalized.BatchSize {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.DomainIDs))
	for _, domainID := range result.DomainIDs {
		if domainID == uuid.Nil {
			return false
		}
		if _, exists := seen[domainID]; exists {
			return false
		}
		seen[domainID] = struct{}{}
	}
	return true
}

// MailboxDomainProvisioningAction is sealed before provider I/O.
type MailboxDomainProvisioningAction string

const (
	MailboxDomainProvisioningActionEnsure              MailboxDomainProvisioningAction = "ensure"
	MailboxDomainProvisioningActionReconcileThenEnsure MailboxDomainProvisioningAction = "reconcile_then_ensure"
)

func validMailboxDomainProvisioningAction(action MailboxDomainProvisioningAction) bool {
	return action == MailboxDomainProvisioningActionEnsure ||
		action == MailboxDomainProvisioningActionReconcileThenEnsure
}

// MailboxDomainProvisioningClaimRequest selects one durable domain claim.
type MailboxDomainProvisioningClaimRequest struct {
	DomainID uuid.UUID
}

// Valid reports whether the request identifies exactly one domain.
func (request MailboxDomainProvisioningClaimRequest) Valid() bool {
	return request.DomainID != uuid.Nil
}

// MailboxDomainProvisioningFence is the exact desired-state and lease fence.
// Only desired present+enabled domains may reach provider work.
type MailboxDomainProvisioningFence struct {
	DomainID           uuid.UUID
	OrganizationID     uuid.UUID
	LeaseToken         uuid.UUID
	LeaseDeadline      time.Time
	LeaseDuration      time.Duration
	DesiredGeneration  int64
	DesiredHashVersion int32
	DesiredHash        string
	ProvisioningKey    string
	Domain             string
	DesiredPresent     bool
	DesiredEnabled     bool
	Action             MailboxDomainProvisioningAction
}

// Valid reports whether the fence is safe for exactly one provider command.
func (fence MailboxDomainProvisioningFence) Valid() bool {
	if fence.DomainID == uuid.Nil || fence.OrganizationID == uuid.Nil || fence.LeaseToken == uuid.Nil {
		return false
	}
	if fence.LeaseDuration != MailboxEngineCommandLeaseDuration || !mailboxDomainProvisioningCanonicalSecond(fence.LeaseDeadline) {
		return false
	}
	if fence.DesiredGeneration < 1 || fence.DesiredHashVersion < 1 ||
		!mailboxDomainProvisioningDesiredHashValid(fence.DesiredHash) ||
		!validSafeText(fence.ProvisioningKey, MaxEngineOpaqueIDBytes, true) ||
		!validCanonicalRegistryDomain(fence.Domain) ||
		!fence.DesiredPresent || !fence.DesiredEnabled ||
		!validMailboxDomainProvisioningAction(fence.Action) {
		return false
	}
	return true
}

func mailboxDomainProvisioningCanonicalSecond(value time.Time) bool {
	return !value.IsZero() && value == value.UTC().Truncate(time.Second)
}

// mailboxDomainProvisioningDesiredHashValid accepts exactly the canonical
// lowercase hexadecimal encoding of a SHA-256 digest.
func mailboxDomainProvisioningDesiredHashValid(value string) bool {
	if len(value) != hex.EncodedLen(sha256.Size) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}

// MailboxDomainProvisioningClaim seals one validated fence before provider I/O.
type MailboxDomainProvisioningClaim struct {
	fence MailboxDomainProvisioningFence
}

// NewMailboxDomainProvisioningClaim seals a durable claim fence.
func NewMailboxDomainProvisioningClaim(fence MailboxDomainProvisioningFence) (MailboxDomainProvisioningClaim, bool) {
	if !fence.Valid() {
		return MailboxDomainProvisioningClaim{}, false
	}
	return MailboxDomainProvisioningClaim{fence: fence}, true
}

// Valid reports whether the sealed fence remains valid.
func (claim MailboxDomainProvisioningClaim) Valid() bool {
	return claim.fence.Valid()
}

// Fence returns a value copy of the sealed durable fence.
func (claim MailboxDomainProvisioningClaim) Fence() MailboxDomainProvisioningFence {
	return claim.fence
}

// DomainSpec returns the only allowed provider desired state.
func (claim MailboxDomainProvisioningClaim) DomainSpec() DomainSpec {
	if !claim.Valid() {
		return DomainSpec{}
	}
	return DomainSpec{
		ProvisioningKey:   claim.fence.ProvisioningKey,
		Domain:            claim.fence.Domain,
		DesiredGeneration: claim.fence.DesiredGeneration,
		DesiredHash:       claim.fence.DesiredHash,
		Present:           true,
		Enabled:           true,
	}
}

// Action returns the sealed provider action.
func (claim MailboxDomainProvisioningClaim) Action() MailboxDomainProvisioningAction {
	if !claim.Valid() {
		return ""
	}
	return claim.fence.Action
}

// MailboxDomainProvisioningClaimOutcome is a closed durable claim branch.
type MailboxDomainProvisioningClaimOutcome string

const (
	MailboxDomainProvisioningClaimed MailboxDomainProvisioningClaimOutcome = "claimed"
	MailboxDomainProvisioningNoop    MailboxDomainProvisioningClaimOutcome = "noop"
)

// MailboxDomainProvisioningClaimResult is one durable claim result.
type MailboxDomainProvisioningClaimResult struct {
	Outcome MailboxDomainProvisioningClaimOutcome
	Claim   MailboxDomainProvisioningClaim
}

// Valid reports whether branch-specific claim data is exact.
func (result MailboxDomainProvisioningClaimResult) Valid() bool {
	switch result.Outcome {
	case MailboxDomainProvisioningClaimed:
		return result.Claim.Valid()
	case MailboxDomainProvisioningNoop:
		return result.Claim == (MailboxDomainProvisioningClaim{})
	default:
		return false
	}
}

// MailboxDomainProvisioner is the future worker provider port and owns I/O
// outside durable transactions. No current implementation is asserted against
// this interface: MailboxRegistryProvisioningEngine remains an EnsureDomain
// adapter until it independently implements GetDomain.
type MailboxDomainProvisioner interface {
	GetDomain(context.Context, string) (EngineDomainState, error)
	EnsureDomain(context.Context, DomainSpec) (EngineDomainState, error)
}

// MailboxDomainProvisioningObserved is a validated provider observation.
type MailboxDomainProvisioningObserved struct {
	EngineDomainID     string
	State              EngineDomainStateKind
	RequiredDNSRecords RequiredDNSRecords
	Verification       EngineVerificationState
	DKIM               EngineVerificationState
}

// Clone returns an independent copy, including DNS records.
func (observed MailboxDomainProvisioningObserved) Clone() MailboxDomainProvisioningObserved {
	cloned := observed
	cloned.RequiredDNSRecords = make(RequiredDNSRecords, len(observed.RequiredDNSRecords))
	copy(cloned.RequiredDNSRecords, observed.RequiredDNSRecords)
	return cloned
}

// Valid accepts only provider-observed DNS-required and fully verified states.
func (observed MailboxDomainProvisioningObserved) Valid() bool {
	if !validOpaqueID(observed.EngineDomainID) ||
		!validMailboxDomainProvisioningVerification(observed.Verification) ||
		!validMailboxDomainProvisioningVerification(observed.DKIM) ||
		observed.RequiredDNSRecords == nil ||
		observed.RequiredDNSRecords.Validate() != nil {
		return false
	}
	switch observed.State {
	case EngineDomainDNSRequired:
		return len(observed.RequiredDNSRecords) > 0 && observed.Verification != EngineVerificationVerified
	case EngineDomainVerified:
		return observed.Verification == EngineVerificationVerified && observed.DKIM == EngineVerificationVerified
	default:
		return false
	}
}

func validMailboxDomainProvisioningVerification(value EngineVerificationState) bool {
	switch value {
	case EngineVerificationUnknown, EngineVerificationPending, EngineVerificationVerified, EngineVerificationFailed:
		return true
	default:
		return false
	}
}

func mailboxDomainProvisioningObservedFromEngine(state EngineDomainState) (MailboxDomainProvisioningObserved, bool) {
	if !state.Applied {
		return MailboxDomainProvisioningObserved{}, false
	}
	records := make(RequiredDNSRecords, len(state.DNSRecords))
	for index, record := range state.DNSRecords {
		records[index] = RequiredDNSRecord{
			Name:  record.Name,
			Type:  record.Type,
			Value: record.Value,
			TTL:   record.TTL,
		}
	}
	observed := MailboxDomainProvisioningObserved{
		EngineDomainID:     state.DomainID,
		State:              state.State,
		RequiredDNSRecords: records,
		Verification:       state.Verification,
		DKIM:               state.DKIM,
	}
	return observed, observed.Valid()
}

// MailboxDomainProvisioningFinalizeOutcome is a closed durable finalization branch.
type MailboxDomainProvisioningFinalizeOutcome string

const (
	MailboxDomainProvisioningFinalizeDNSRequired MailboxDomainProvisioningFinalizeOutcome = "dns_required"
	MailboxDomainProvisioningFinalizeVerified    MailboxDomainProvisioningFinalizeOutcome = "verified"
	MailboxDomainProvisioningFinalizeRetry       MailboxDomainProvisioningFinalizeOutcome = "retry"
	// MailboxDomainProvisioningFinalizeSuperseded is an output-only durable CAS
	// result; it is never a request outcome.
	MailboxDomainProvisioningFinalizeSuperseded MailboxDomainProvisioningFinalizeOutcome = "superseded"
)

// MailboxDomainProvisioningFinalizeRequest is one sealed provider outcome.
type MailboxDomainProvisioningFinalizeRequest struct {
	Claim   MailboxDomainProvisioningClaim
	Outcome MailboxDomainProvisioningFinalizeOutcome
	Code    MailboxDomainProvisioningCode

	observed MailboxDomainProvisioningObserved
}

// NewMailboxDomainProvisioningFinalizeDNSRequired seals a DNS-required observation.
func NewMailboxDomainProvisioningFinalizeDNSRequired(claim MailboxDomainProvisioningClaim, state EngineDomainState) (MailboxDomainProvisioningFinalizeRequest, bool) {
	request, ok := mailboxDomainProvisioningFinalizeObserved(claim, state)
	if !ok || request.Outcome != MailboxDomainProvisioningFinalizeDNSRequired {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	return request, true
}

// NewMailboxDomainProvisioningFinalizeVerified seals a fully verified observation.
func NewMailboxDomainProvisioningFinalizeVerified(claim MailboxDomainProvisioningClaim, state EngineDomainState) (MailboxDomainProvisioningFinalizeRequest, bool) {
	request, ok := mailboxDomainProvisioningFinalizeObserved(claim, state)
	if !ok || request.Outcome != MailboxDomainProvisioningFinalizeVerified {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	return request, true
}

func mailboxDomainProvisioningFinalizeObserved(claim MailboxDomainProvisioningClaim, state EngineDomainState) (MailboxDomainProvisioningFinalizeRequest, bool) {
	if !claim.Valid() {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	observed, ok := mailboxDomainProvisioningObservedFromEngine(state)
	if !ok {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	var outcome MailboxDomainProvisioningFinalizeOutcome
	switch observed.State {
	case EngineDomainDNSRequired:
		outcome = MailboxDomainProvisioningFinalizeDNSRequired
	case EngineDomainVerified:
		outcome = MailboxDomainProvisioningFinalizeVerified
	default:
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	return MailboxDomainProvisioningFinalizeRequest{
		Claim:    claim,
		Outcome:  outcome,
		observed: observed.Clone(),
	}, true
}

// NewMailboxDomainProvisioningFinalizeRetry seals a schema-compatible retry.
func NewMailboxDomainProvisioningFinalizeRetry(claim MailboxDomainProvisioningClaim, code MailboxDomainProvisioningCode) (MailboxDomainProvisioningFinalizeRequest, bool) {
	if !claim.Valid() || !validMailboxDomainProvisioningCode(code) {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	return MailboxDomainProvisioningFinalizeRequest{
		Claim:   claim,
		Outcome: MailboxDomainProvisioningFinalizeRetry,
		Code:    code,
	}, true
}

// Valid reports whether the finalization has exact branch-specific data.
func (request MailboxDomainProvisioningFinalizeRequest) Valid() bool {
	if !request.Claim.Valid() {
		return false
	}
	switch request.Outcome {
	case MailboxDomainProvisioningFinalizeDNSRequired:
		return request.Code == "" && request.observed.Valid() && request.observed.State == EngineDomainDNSRequired
	case MailboxDomainProvisioningFinalizeVerified:
		return request.Code == "" && request.observed.Valid() && request.observed.State == EngineDomainVerified
	case MailboxDomainProvisioningFinalizeRetry:
		return validMailboxDomainProvisioningCode(request.Code) &&
			request.observed.EngineDomainID == "" &&
			request.observed.State == "" &&
			request.observed.RequiredDNSRecords == nil &&
			request.observed.Verification == "" &&
			request.observed.DKIM == ""
	default:
		return false
	}
}

// Observed returns an independent provider observation for successful branches.
func (request MailboxDomainProvisioningFinalizeRequest) Observed() (MailboxDomainProvisioningObserved, bool) {
	if !request.Valid() || request.Outcome == MailboxDomainProvisioningFinalizeRetry {
		return MailboxDomainProvisioningObserved{}, false
	}
	return request.observed.Clone(), true
}

// MailboxDomainProvisioningFinalizeResult is the closed durable write result.
type MailboxDomainProvisioningFinalizeResult struct {
	Outcome MailboxDomainProvisioningFinalizeOutcome
	Code    MailboxDomainProvisioningCode
}

// Valid reports whether the repository recorded the requested branch.
func (result MailboxDomainProvisioningFinalizeResult) Valid() bool {
	switch result.Outcome {
	case MailboxDomainProvisioningFinalizeDNSRequired, MailboxDomainProvisioningFinalizeVerified, MailboxDomainProvisioningFinalizeSuperseded:
		return result.Code == ""
	case MailboxDomainProvisioningFinalizeRetry:
		return validMailboxDomainProvisioningCode(result.Code)
	default:
		return false
	}
}

// MailboxDomainProvisioningRepository owns bounded scan, durable claim and
// exact finalization. It performs no provider I/O.
type MailboxDomainProvisioningRepository interface {
	ScanMailboxDomainProvisioning(context.Context, MailboxDomainProvisioningScanRequest) (MailboxDomainProvisioningScanResult, error)
	ClaimMailboxDomainProvisioning(context.Context, MailboxDomainProvisioningClaimRequest) (MailboxDomainProvisioningClaimResult, error)
	FinalizeMailboxDomainProvisioning(context.Context, MailboxDomainProvisioningFinalizeRequest) (MailboxDomainProvisioningFinalizeResult, error)
}

// ExecuteMailboxDomainProvisioning performs the sealed provider action and
// always returns a finalizable safe retry after a provider panic.
func ExecuteMailboxDomainProvisioning(ctx context.Context, provider MailboxDomainProvisioner, claim MailboxDomainProvisioningClaim) (request MailboxDomainProvisioningFinalizeRequest, ok bool) {
	if ctx == nil || provider == nil || !claim.Valid() {
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
	retry := func(code MailboxDomainProvisioningCode) (MailboxDomainProvisioningFinalizeRequest, bool) {
		return NewMailboxDomainProvisioningFinalizeRetry(claim, code)
	}
	defer func() {
		if recover() != nil {
			request, ok = retry(MailboxDomainProvisioningCodeOutcomeUnknown)
		}
	}()
	if ctx.Err() != nil {
		return retry(MailboxDomainProvisioningCodeEngineUnavailable)
	}
	switch claim.Action() {
	case MailboxDomainProvisioningActionEnsure:
		return mailboxDomainProvisioningEnsure(ctx, provider, claim)
	case MailboxDomainProvisioningActionReconcileThenEnsure:
		state, err := provider.GetDomain(ctx, claim.Fence().Domain)
		if err != nil || ctx.Err() != nil {
			return retry(MailboxDomainProvisioningCodeOutcomeUnknown)
		}
		switch state.State {
		case EngineDomainAbsent:
			if state.DomainID != "" || state.Applied || len(state.DNSRecords) != 0 ||
				state.Verification != EngineVerificationUnknown || state.DKIM != EngineVerificationUnknown {
				return retry(MailboxDomainProvisioningCodeDNSInvalid)
			}
			return mailboxDomainProvisioningEnsure(ctx, provider, claim)
		case EngineDomainDNSRequired:
			if request, ok := NewMailboxDomainProvisioningFinalizeDNSRequired(claim, state); ok {
				return request, true
			}
			return retry(MailboxDomainProvisioningCodeDNSInvalid)
		case EngineDomainVerified:
			if request, ok := NewMailboxDomainProvisioningFinalizeVerified(claim, state); ok {
				return request, true
			}
			return retry(MailboxDomainProvisioningCodeDNSInvalid)
		case EngineDomainRejected:
			return retry(MailboxDomainProvisioningCodeApplyRejected)
		case EngineDomainIndeterminate:
			return retry(MailboxDomainProvisioningCodeOutcomeUnknown)
		default:
			return retry(MailboxDomainProvisioningCodeEngineUnavailable)
		}
	default:
		return MailboxDomainProvisioningFinalizeRequest{}, false
	}
}

func mailboxDomainProvisioningEnsure(ctx context.Context, provider MailboxDomainProvisioner, claim MailboxDomainProvisioningClaim) (MailboxDomainProvisioningFinalizeRequest, bool) {
	retry := func(code MailboxDomainProvisioningCode) (MailboxDomainProvisioningFinalizeRequest, bool) {
		return NewMailboxDomainProvisioningFinalizeRetry(claim, code)
	}
	state, err := provider.EnsureDomain(ctx, claim.DomainSpec())
	if err != nil || ctx.Err() != nil {
		return retry(mailboxDomainProvisioningEnsureErrorCode(err))
	}
	switch state.State {
	case EngineDomainDNSRequired:
		if request, ok := NewMailboxDomainProvisioningFinalizeDNSRequired(claim, state); ok {
			return request, true
		}
	case EngineDomainVerified:
		if request, ok := NewMailboxDomainProvisioningFinalizeVerified(claim, state); ok {
			return request, true
		}
	case EngineDomainRejected:
		return retry(MailboxDomainProvisioningCodeApplyRejected)
	}
	// Ensure may have reached the provider even when the returned observation
	// is malformed or indeterminate; retain the attempted tuple for reconciliation.
	return retry(MailboxDomainProvisioningCodeOutcomeUnknown)
}

func mailboxDomainProvisioningGetErrorCode(err error) MailboxDomainProvisioningCode {
	if errors.Is(err, ErrEngineRejectedPolicy) ||
		errors.Is(err, ErrEngineUnauthorizedAccount) ||
		errors.Is(err, ErrEnginePayloadTooLarge) {
		return MailboxDomainProvisioningCodeApplyRejected
	}
	return MailboxDomainProvisioningCodeEngineUnavailable
}

func mailboxDomainProvisioningEnsureErrorCode(err error) MailboxDomainProvisioningCode {
	if errors.Is(err, ErrEngineRejectedPolicy) ||
		errors.Is(err, ErrEngineUnauthorizedAccount) ||
		errors.Is(err, ErrEnginePayloadTooLarge) {
		return MailboxDomainProvisioningCodeApplyRejected
	}
	// Cancellation, timeout, unavailable and indeterminate errors may occur
	// after dispatch, therefore they must be reconciled before another ensure.
	return MailboxDomainProvisioningCodeOutcomeUnknown
}

// MailboxDomainProvisioningRetryDelay returns a bounded retry delay.
func MailboxDomainProvisioningRetryDelay(attempt int) (time.Duration, bool) {
	if attempt < 1 || attempt >= math.MaxInt32 {
		return 0, false
	}
	if attempt >= 5 {
		return mailboxDomainProvisioningRetryMaximumDelay, true
	}
	return mailboxDomainProvisioningRetryInitialDelay * time.Duration(1<<(attempt-1)), true
}
