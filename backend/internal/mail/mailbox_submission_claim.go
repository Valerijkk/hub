package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	mailboxSubmissionPolicyBindingDomain  = "hub.mail.submission.policy.binding.v1"
	mailboxSubmissionPolicyBindingVersion = uint32(1)
)

// ErrMailboxSubmissionPolicyBindingInvalid is returned when a durable policy
// snapshot cannot safely bind a submission claim.
var ErrMailboxSubmissionPolicyBindingInvalid = errors.New("mailbox submission policy binding invalid")

// MailboxSubmissionPolicyDomain is one versioned, verified internal mail
// domain. Its identifier and version make a delete-and-recreate cycle visible
// to the durable policy binding.
type MailboxSubmissionPolicyDomain struct {
	ID      uuid.UUID
	Version int64
	Domain  string
}

// MailboxSubmissionPolicyBindingInput contains the full durable policy
// snapshot used to create a submission claim. It contains no provider tuple
// and no recipient addresses.
type MailboxSubmissionPolicyBindingInput struct {
	OrganizationID           uuid.UUID
	PolicyID                 uuid.UUID
	PolicyVersion            int64
	PolicyMaxMessageBytes    int64
	PolicyMaxRecipients      int32
	ExternalMailEnabled      bool
	VerifiedEnabledDomains []MailboxSubmissionPolicyDomain
}

// MailboxSubmissionPolicyBinding is an immutable, versioned policy snapshot.
// The provider reads a semantic request derived by PreflightRequest; this
// full hash additionally binds database identities and versions against ABA.
type MailboxSubmissionPolicyBinding struct {
	organizationID      uuid.UUID
	policyID            uuid.UUID
	policyVersion       int64
	maxMessageBytes     int64
	maxRecipients       int32
	externalMailEnabled bool
	domains             []MailboxSubmissionPolicyDomain
	hash                [sha256.Size]byte
}

// NewMailboxSubmissionPolicyBinding validates, canonicalizes, and copies a
// durable policy snapshot before it may participate in a submission claim.
func NewMailboxSubmissionPolicyBinding(input MailboxSubmissionPolicyBindingInput) (MailboxSubmissionPolicyBinding, error) {
	domains, ok := canonicalMailboxSubmissionPolicyDomains(input.VerifiedEnabledDomains)
	if !ok || input.OrganizationID == uuid.Nil || input.PolicyID == uuid.Nil || input.PolicyVersion < 1 ||
		input.PolicyMaxMessageBytes < 1_048_576 || input.PolicyMaxMessageBytes > MaxEngineMessageBytes ||
		input.PolicyMaxRecipients < 1 || input.PolicyMaxRecipients > MaxEngineDraftTotalRecipientItems {
		return MailboxSubmissionPolicyBinding{}, ErrMailboxSubmissionPolicyBindingInvalid
	}

	binding := MailboxSubmissionPolicyBinding{
		organizationID:      input.OrganizationID,
		policyID:            input.PolicyID,
		policyVersion:       input.PolicyVersion,
		maxMessageBytes:     input.PolicyMaxMessageBytes,
		maxRecipients:       input.PolicyMaxRecipients,
		externalMailEnabled: input.ExternalMailEnabled,
		domains:             domains,
	}
	hash, ok := mailboxSubmissionPolicyBindingHash(binding)
	if !ok || hash == [sha256.Size]byte{} {
		return MailboxSubmissionPolicyBinding{}, ErrMailboxSubmissionPolicyBindingInvalid
	}
	binding.hash = hash
	return mailboxSubmissionPolicyBindingCopy(binding), nil
}

// MaxMessageBytes returns the policy cap safely bound into this snapshot.
func (binding MailboxSubmissionPolicyBinding) MaxMessageBytes() int64 {
	if !mailboxSubmissionPolicyBindingValid(binding) {
		return 0
	}
	return binding.maxMessageBytes
}

// MaxRecipients returns the policy recipient cap safely bound into this snapshot.
func (binding MailboxSubmissionPolicyBinding) MaxRecipients() int32 {
	if !mailboxSubmissionPolicyBindingValid(binding) {
		return 0
	}
	return binding.maxRecipients
}

// ExternalMailEnabled reports the policy toggle safely bound into this snapshot.
func (binding MailboxSubmissionPolicyBinding) ExternalMailEnabled() bool {
	return mailboxSubmissionPolicyBindingValid(binding) && binding.externalMailEnabled
}

// InternalDomains returns a defensive, canonical copy of verified internal domains.
func (binding MailboxSubmissionPolicyBinding) InternalDomains() []string {
	if !mailboxSubmissionPolicyBindingValid(binding) {
		return nil
	}
	domains := make([]string, len(binding.domains))
	for index, domain := range binding.domains {
		domains[index] = strings.Clone(domain.Domain)
	}
	return domains
}

// Hash returns the full versioned durable policy binding hash.
func (binding MailboxSubmissionPolicyBinding) Hash() [sha256.Size]byte {
	if !mailboxSubmissionPolicyBindingValid(binding) {
		return [sha256.Size]byte{}
	}
	return binding.hash
}

// PreflightRequest derives the provider-safe semantic policy request for one
// exact submission. The provider tuple stays only inside EngineSubmission.
func (binding MailboxSubmissionPolicyBinding) PreflightRequest(submission EngineSubmission) (EngineSubmissionPreflightRequest, bool) {
	if !mailboxSubmissionPolicyBindingValid(binding) || validateSubmission(submission) != nil {
		return EngineSubmissionPreflightRequest{}, false
	}
	return EngineSubmissionPreflightRequest{
		Submission:             submission,
		PolicyMaxMessageBytes:  binding.maxMessageBytes,
		MaxRecipients:          binding.maxRecipients,
		ExternalMailEnabled:    binding.externalMailEnabled,
		InternalDomains:        binding.InternalDomains(),
	}, true
}

// MailboxSubmissionClaimRequest asks the repository to atomically claim one
// queued durable submission. The repository chooses the lease and fences.
type MailboxSubmissionClaimRequest struct {
	SubmissionID uuid.UUID
}

// MailboxSubmissionClaim is a sealed capability minted only by the
// repository. It retains the exact provider submission privately for the
// future dispatcher, while public accessors expose only durable-safe data.
type MailboxSubmissionClaim struct {
	submissionID uuid.UUID
	fence        MailboxEngineCommandFence
	policy       MailboxSubmissionPolicyBinding
	submission   EngineSubmission
}

// SubmissionID returns the durable submission identity held by this claim.
func (claim MailboxSubmissionClaim) SubmissionID() uuid.UUID {
	if !mailboxSubmissionClaimValid(claim) {
		return uuid.Nil
	}
	return claim.submissionID
}

// Fence returns the exact shared account/submission lease tuple. A repository
// must match it against both durable rows before an authorization transition.
func (claim MailboxSubmissionClaim) Fence() MailboxEngineCommandFence {
	if !mailboxSubmissionClaimValid(claim) {
		return MailboxEngineCommandFence{}
	}
	return mailboxSubmissionClaimFenceCopy(claim.fence)
}

// Policy returns a defensive copy of the claim's immutable policy snapshot.
func (claim MailboxSubmissionClaim) Policy() MailboxSubmissionPolicyBinding {
	if !mailboxSubmissionClaimValid(claim) {
		return MailboxSubmissionPolicyBinding{}
	}
	return mailboxSubmissionPolicyBindingCopy(claim.policy)
}

// MailboxSubmissionClaimResult is either one sealed claim or a busy lease
// observation. Claimed results never carry BusyUntil; busy results never carry
// a usable Claim.
type MailboxSubmissionClaimResult struct {
	claimed   bool
	claim     MailboxSubmissionClaim
	busyUntil *time.Time
}

// Claimed reports whether this result carries one valid sealed claim.
func (result MailboxSubmissionClaimResult) Claimed() bool {
	return mailboxSubmissionClaimResultValid(result) && result.claimed
}

// Claim returns a defensive sealed capability only for a valid claimed result.
func (result MailboxSubmissionClaimResult) Claim() (MailboxSubmissionClaim, bool) {
	if !mailboxSubmissionClaimResultValid(result) || !result.claimed {
		return MailboxSubmissionClaim{}, false
	}
	return mailboxSubmissionClaimCopy(result.claim), true
}

// BusyUntil returns a defensive UTC lease observation only for a valid busy result.
func (result MailboxSubmissionClaimResult) BusyUntil() (time.Time, bool) {
	if !mailboxSubmissionClaimResultValid(result) || result.claimed {
		return time.Time{}, false
	}
	return result.busyUntil.Round(0).UTC(), true
}

// MailboxSubmissionAuthorizeRequest binds a sealed claim to the read-only
// provider preflight. The repository must revalidate both policy bindings and
// both durable fences in one authorization CAS before provider submission.
type MailboxSubmissionAuthorizeRequest struct {
	Claim     MailboxSubmissionClaim
	Preflight EngineSubmissionPreflight
}

// MailboxSubmissionAuthorizeOutcome is the closed durable authorization result.
type MailboxSubmissionAuthorizeOutcome string

const (
	MailboxSubmissionAuthorizeOutcomeAuthorized  MailboxSubmissionAuthorizeOutcome = "authorized"
	MailboxSubmissionAuthorizeOutcomeRateLimited MailboxSubmissionAuthorizeOutcome = "rate_limited"
	MailboxSubmissionAuthorizeOutcomeRequeued    MailboxSubmissionAuthorizeOutcome = "requeued"
)

// MailboxSubmissionDispatchAuthorization is a sealed capability issued only
// after the durable authorizing-to-submitting CAS has succeeded.
type MailboxSubmissionDispatchAuthorization struct {
	submissionID uuid.UUID
	fence        MailboxEngineCommandFence
	authorizedAt time.Time
}

// SubmissionID returns the durable submission identity authorized for dispatch.
func (authorization MailboxSubmissionDispatchAuthorization) SubmissionID() uuid.UUID {
	if !mailboxSubmissionDispatchAuthorizationValid(authorization) {
		return uuid.Nil
	}
	return authorization.submissionID
}

// Fence returns the exact shared lease tuple authorized for dispatch.
func (authorization MailboxSubmissionDispatchAuthorization) Fence() MailboxEngineCommandFence {
	if !mailboxSubmissionDispatchAuthorizationValid(authorization) {
		return MailboxEngineCommandFence{}
	}
	return mailboxSubmissionClaimFenceCopy(authorization.fence)
}

// AuthorizedAt returns a canonical UTC copy of the authorization instant.
func (authorization MailboxSubmissionDispatchAuthorization) AuthorizedAt() time.Time {
	if !mailboxSubmissionDispatchAuthorizationValid(authorization) {
		return time.Time{}
	}
	return authorization.authorizedAt.Round(0).UTC()
}

// MailboxSubmissionAuthorizeResult contains either a dispatch capability, a
// rate-limit retry instant, or a durable requeue result.
type MailboxSubmissionAuthorizeResult struct {
	outcome       MailboxSubmissionAuthorizeOutcome
	authorization MailboxSubmissionDispatchAuthorization
	retryAt       *time.Time
}

// Outcome returns the closed durable outcome, or the zero value for a
// malformed result.
func (result MailboxSubmissionAuthorizeResult) Outcome() MailboxSubmissionAuthorizeOutcome {
	if !mailboxSubmissionAuthorizeResultValid(result) {
		return ""
	}
	return result.outcome
}

// Authorization returns a sealed dispatch capability only after a valid
// authorization transition.
func (result MailboxSubmissionAuthorizeResult) Authorization() (MailboxSubmissionDispatchAuthorization, bool) {
	if !mailboxSubmissionAuthorizeResultValid(result) || result.outcome != MailboxSubmissionAuthorizeOutcomeAuthorized {
		return MailboxSubmissionDispatchAuthorization{}, false
	}
	return mailboxSubmissionDispatchAuthorizationCopy(result.authorization), true
}

// RetryAt returns a defensive UTC retry instant only for a valid rate-limit result.
func (result MailboxSubmissionAuthorizeResult) RetryAt() (time.Time, bool) {
	if !mailboxSubmissionAuthorizeResultValid(result) || result.outcome != MailboxSubmissionAuthorizeOutcomeRateLimited {
		return time.Time{}, false
	}
	return result.retryAt.Round(0).UTC(), true
}

// MailboxSubmissionClaimRepository owns the short transactional claim and
// authorization CAS boundaries. It performs no provider I/O.
type MailboxSubmissionClaimRepository interface {
	ClaimMailboxSubmission(context.Context, MailboxSubmissionClaimRequest) (MailboxSubmissionClaimResult, error)
	AuthorizeMailboxSubmission(context.Context, MailboxSubmissionAuthorizeRequest) (MailboxSubmissionAuthorizeResult, error)
}

func canonicalMailboxSubmissionPolicyDomains(source []MailboxSubmissionPolicyDomain) ([]MailboxSubmissionPolicyDomain, bool) {
	if len(source) < 1 || len(source) > 100 {
		return nil, false
	}
	domains := append([]MailboxSubmissionPolicyDomain(nil), source...)
	seenIDs := make(map[uuid.UUID]struct{}, len(domains))
	seenDomains := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		if !mailboxSubmissionPolicyDomainValid(domain) {
			return nil, false
		}
		if _, exists := seenIDs[domain.ID]; exists {
			return nil, false
		}
		if _, exists := seenDomains[domain.Domain]; exists {
			return nil, false
		}
		seenIDs[domain.ID] = struct{}{}
		seenDomains[domain.Domain] = struct{}{}
	}
	sort.Slice(domains, func(left, right int) bool {
		if domains[left].Domain != domains[right].Domain {
			return domains[left].Domain < domains[right].Domain
		}
		return bytes.Compare(domains[left].ID[:], domains[right].ID[:]) < 0
	})
	return domains, true
}

func mailboxSubmissionPolicyDomainValid(domain MailboxSubmissionPolicyDomain) bool {
	if domain.ID == uuid.Nil || domain.Version < 1 || len(domain.Domain) < 1 || len(domain.Domain) > 253 || !isLowerASCII(domain.Domain) {
		return false
	}
	canonical, err := NormalizeAddress(EngineAddress{Email: "domain-check@" + domain.Domain})
	return err == nil && canonical.Email == "domain-check@"+domain.Domain
}

func mailboxSubmissionPolicyBindingValid(binding MailboxSubmissionPolicyBinding) bool {
	if !mailboxSubmissionPolicyBindingFieldsValid(binding) || binding.hash == [sha256.Size]byte{} {
		return false
	}
	expected, ok := mailboxSubmissionPolicyBindingHash(binding)
	return ok && subtle.ConstantTimeCompare(binding.hash[:], expected[:]) == 1
}

func mailboxSubmissionPolicyBindingFieldsValid(binding MailboxSubmissionPolicyBinding) bool {
	if binding.organizationID == uuid.Nil || binding.policyID == uuid.Nil || binding.policyVersion < 1 ||
		binding.maxMessageBytes < 1_048_576 || binding.maxMessageBytes > MaxEngineMessageBytes ||
		binding.maxRecipients < 1 || binding.maxRecipients > MaxEngineDraftTotalRecipientItems {
		return false
	}
	canonical, ok := canonicalMailboxSubmissionPolicyDomains(binding.domains)
	if !ok || len(canonical) != len(binding.domains) {
		return false
	}
	for index := range canonical {
		if canonical[index] != binding.domains[index] {
			return false
		}
	}
	return true
}

// mailboxSubmissionPolicyBindingHash is SHA-256 over V1's exact binary
// encoding: frame(domain), U32BE(version), organization UUID16, policy UUID16,
// U64BE(policy version), U64BE(message cap), U64BE(recipient cap), U8(toggle),
// U32BE(domain count), then ordered UUID16/U64BE/frame(domain) records.
func mailboxSubmissionPolicyBindingHash(binding MailboxSubmissionPolicyBinding) ([sha256.Size]byte, bool) {
	if !mailboxSubmissionPolicyBindingFieldsValid(binding) {
		return [sha256.Size]byte{}, false
	}
	encoded := make([]byte, 0, len(mailboxSubmissionPolicyBindingDomain)+len(binding.domains)*48+96)
	var ok bool
	encoded, ok = mailboxSubmissionPolicyBindingAppendFrame(encoded, mailboxSubmissionPolicyBindingDomain)
	if !ok {
		return [sha256.Size]byte{}, false
	}
	var version [4]byte
	binary.BigEndian.PutUint32(version[:], mailboxSubmissionPolicyBindingVersion)
	encoded = append(encoded, version[:]...)
	encoded = append(encoded, binding.organizationID[:]...)
	encoded = append(encoded, binding.policyID[:]...)
	for _, value := range []int64{binding.policyVersion, binding.maxMessageBytes, int64(binding.maxRecipients)} {
		var encodedValue [8]byte
		binary.BigEndian.PutUint64(encodedValue[:], uint64(value))
		encoded = append(encoded, encodedValue[:]...)
	}
	if binding.externalMailEnabled {
		encoded = append(encoded, 1)
	} else {
		encoded = append(encoded, 0)
	}
	var count [4]byte
	binary.BigEndian.PutUint32(count[:], uint32(len(binding.domains)))
	encoded = append(encoded, count[:]...)
	for _, domain := range binding.domains {
		encoded = append(encoded, domain.ID[:]...)
		var domainVersion [8]byte
		binary.BigEndian.PutUint64(domainVersion[:], uint64(domain.Version))
		encoded = append(encoded, domainVersion[:]...)
		encoded, ok = mailboxSubmissionPolicyBindingAppendFrame(encoded, domain.Domain)
		if !ok {
			return [sha256.Size]byte{}, false
		}
	}
	return sha256.Sum256(encoded), true
}

func mailboxSubmissionPolicyBindingAppendFrame(encoded []byte, value string) ([]byte, bool) {
	if uint64(len(value)) > uint64(^uint32(0)) {
		return nil, false
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	encoded = append(encoded, length[:]...)
	return append(encoded, value...), true
}

func mailboxSubmissionPolicyBindingCopy(binding MailboxSubmissionPolicyBinding) MailboxSubmissionPolicyBinding {
	binding.domains = append([]MailboxSubmissionPolicyDomain(nil), binding.domains...)
	for index := range binding.domains {
		binding.domains[index].Domain = strings.Clone(binding.domains[index].Domain)
	}
	return binding
}

func mailboxSubmissionClaimValid(claim MailboxSubmissionClaim) bool {
	return claim.submissionID != uuid.Nil && mailboxSubmissionClaimFenceValid(claim.fence) &&
		mailboxSubmissionPolicyBindingValid(claim.policy) && validateSubmission(claim.submission) == nil
}

func mailboxSubmissionClaimFenceValid(fence MailboxEngineCommandFence) bool {
	return fence.AccountID != uuid.Nil && fence.LeaseID != uuid.Nil && fence.MutationEpoch >= 1 &&
		mailboxSubmissionClaimExactTimeValid(fence.LeaseUntil)
}

func mailboxSubmissionClaimExactTimeValid(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0 && value == value.Round(0)
}

func mailboxSubmissionClaimFenceCopy(fence MailboxEngineCommandFence) MailboxEngineCommandFence {
	fence.LeaseUntil = fence.LeaseUntil.Round(0).UTC()
	return fence
}

func mailboxSubmissionClaimCopy(claim MailboxSubmissionClaim) MailboxSubmissionClaim {
	claim.fence = mailboxSubmissionClaimFenceCopy(claim.fence)
	claim.policy = mailboxSubmissionPolicyBindingCopy(claim.policy)
	return claim
}

func mailboxSubmissionClaimResultValid(result MailboxSubmissionClaimResult) bool {
	if result.claimed {
		return mailboxSubmissionClaimValid(result.claim) && result.busyUntil == nil
	}
	return mailboxSubmissionClaimZero(result.claim) && mailboxSubmissionClaimBusyUntilValid(result.busyUntil)
}

func mailboxSubmissionClaimZero(claim MailboxSubmissionClaim) bool {
	return claim.submissionID == uuid.Nil && claim.fence == (MailboxEngineCommandFence{}) &&
		mailboxSubmissionPolicyBindingZero(claim.policy) && claim.submission == (EngineSubmission{})
}

func mailboxSubmissionPolicyBindingZero(binding MailboxSubmissionPolicyBinding) bool {
	return binding.organizationID == uuid.Nil && binding.policyID == uuid.Nil && binding.policyVersion == 0 &&
		binding.maxMessageBytes == 0 && binding.maxRecipients == 0 && !binding.externalMailEnabled &&
		len(binding.domains) == 0 && binding.hash == [sha256.Size]byte{}
}

func mailboxSubmissionClaimBusyUntilValid(value *time.Time) bool {
	return value != nil && mailboxSubmissionClaimExactTimeValid(*value)
}

func mailboxSubmissionAuthorizeRequestValid(request MailboxSubmissionAuthorizeRequest) bool {
	return mailboxSubmissionClaimValid(request.Claim) && mailboxSubmissionPreflightValid(request.Preflight)
}

func mailboxSubmissionPreflightValid(preflight EngineSubmissionPreflight) bool {
	if preflight.RecipientCount < 1 || preflight.RecipientCount > MaxEngineDraftTotalRecipientItems ||
		preflight.MessageBytes < 1 || preflight.MessageBytes > MaxEngineMessageBytes ||
		preflight.BindingHash == [sha256.Size]byte{} || preflight.PolicyBindingHash == [sha256.Size]byte{} {
		return false
	}
	canonical, err := NormalizeAddress(EngineAddress{Email: preflight.FromEmail})
	return err == nil && canonical.Email == preflight.FromEmail
}

func mailboxSubmissionDispatchAuthorizationValid(authorization MailboxSubmissionDispatchAuthorization) bool {
	return authorization.submissionID != uuid.Nil && mailboxSubmissionClaimFenceValid(authorization.fence) &&
		mailboxSubmissionClaimExactTimeValid(authorization.authorizedAt)
}

func mailboxSubmissionAuthorizeResultValid(result MailboxSubmissionAuthorizeResult) bool {
	switch result.outcome {
	case MailboxSubmissionAuthorizeOutcomeAuthorized:
		return mailboxSubmissionDispatchAuthorizationValid(result.authorization) && result.retryAt == nil
	case MailboxSubmissionAuthorizeOutcomeRateLimited:
		return result.authorization == (MailboxSubmissionDispatchAuthorization{}) && mailboxSubmissionClaimBusyUntilValid(result.retryAt)
	case MailboxSubmissionAuthorizeOutcomeRequeued:
		return result.authorization == (MailboxSubmissionDispatchAuthorization{}) && result.retryAt == nil
	default:
		return false
	}
}

func mailboxSubmissionDispatchAuthorizationCopy(authorization MailboxSubmissionDispatchAuthorization) MailboxSubmissionDispatchAuthorization {
	authorization.fence = mailboxSubmissionClaimFenceCopy(authorization.fence)
	authorization.authorizedAt = authorization.authorizedAt.Round(0).UTC()
	return authorization
}
