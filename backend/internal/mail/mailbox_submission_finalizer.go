package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"sort"
	"time"

	"github.com/google/uuid"
)

const mailboxSubmissionFinalizationMaximum = 50

// MailboxSubmissionFinalizationScanRequest bounds one advisory selection of
// accepted submission identities. It grants no claim or engine capability.
type MailboxSubmissionFinalizationScanRequest struct {
	Limit int
}

// MailboxSubmissionFinalizationScanResult contains only copied durable
// identities selected for later one-at-a-time finalization.
type MailboxSubmissionFinalizationScanResult struct {
	SubmissionIDs []uuid.UUID
}

// MailboxSubmissionFinalizationClaimRequest identifies one accepted durable
// submission whose short finalization lease is requested.
type MailboxSubmissionFinalizationClaimRequest struct {
	SubmissionID uuid.UUID
}

// MailboxSubmissionFinalizationExpectedAttachment is the validated durable
// attachment record supplied only to the claimed-result factory.
type MailboxSubmissionFinalizationExpectedAttachment struct {
	AttachmentID uuid.UUID
	ObjectID     uuid.UUID
	SHA256       []byte
	SizeBytes    int64
	ContentID    string
	Filename     string
	MediaType    string
	Disposition  EngineAttachmentDisposition
}

// MailboxSubmissionFinalizationClaimedInput is the repository-only immutable
// input to the claimed-result factory. The factory seals every provider and
// durable binding before it crosses into the finalizer.
type MailboxSubmissionFinalizationClaimedInput struct {
	SubmissionID              uuid.UUID
	DraftID                   uuid.UUID
	Fence                     MailboxEngineCommandFence
	EngineAccountID           string
	ProviderSubmissionID      string
	ProviderMessageID         string
	EngineMessageID           string
	EngineMaxBytes            int64
	EngineBindingHash         []byte
	DispatchAuthorizationHash []byte
	DispatchAuthorizedAt      time.Time
	ExpectedAttachments       []MailboxSubmissionFinalizationExpectedAttachment
}

type mailboxSubmissionFinalizationExpectedAttachment struct {
	attachmentID uuid.UUID
	objectID     uuid.UUID
	sha256       [sha256.Size]byte
	sizeBytes    int64
	contentID    string
	filename     string
	mediaType    string
	disposition  EngineAttachmentDisposition
}

// MailboxSubmissionFinalizationClaim is a sealed capability minted only by a
// repository through NewMailboxSubmissionFinalizationClaimedResult.
type MailboxSubmissionFinalizationClaim struct {
	submissionID              uuid.UUID
	draftID                   uuid.UUID
	fence                     MailboxEngineCommandFence
	engineAccountID           string
	providerSubmissionID      string
	providerMessageID         string
	engineMessageID           string
	engineMaxBytes            int64
	engineBindingHash         [sha256.Size]byte
	dispatchAuthorizationHash [sha256.Size]byte
	dispatchAuthorizedAt      time.Time
	expectedAttachments       []mailboxSubmissionFinalizationExpectedAttachment
}

// SubmissionID returns the durable identity held by this sealed claim.
func (claim MailboxSubmissionFinalizationClaim) SubmissionID() uuid.UUID {
	if !mailboxSubmissionFinalizationClaimValid(claim) {
		return uuid.Nil
	}
	return claim.submissionID
}

// Fence returns a defensive exact copy of the shared command fence.
func (claim MailboxSubmissionFinalizationClaim) Fence() MailboxEngineCommandFence {
	if !mailboxSubmissionFinalizationClaimValid(claim) {
		return MailboxEngineCommandFence{}
	}
	return mailboxSubmissionClaimFenceCopy(claim.fence)
}

// MailboxSubmissionFinalizationClaimOutcome is the closed result of a short
// repository claim transaction.
type MailboxSubmissionFinalizationClaimOutcome string

const (
	MailboxSubmissionFinalizationClaimOutcomeNoop    MailboxSubmissionFinalizationClaimOutcome = "noop"
	MailboxSubmissionFinalizationClaimOutcomeBusy    MailboxSubmissionFinalizationClaimOutcome = "busy"
	MailboxSubmissionFinalizationClaimOutcomeClaimed MailboxSubmissionFinalizationClaimOutcome = "claimed"
)

// MailboxSubmissionFinalizationClaimResult carries exactly one of noop, busy,
// or a sealed finalization claim.
type MailboxSubmissionFinalizationClaimResult struct {
	outcome   MailboxSubmissionFinalizationClaimOutcome
	claim     MailboxSubmissionFinalizationClaim
	busyUntil *time.Time
}

// NewMailboxSubmissionFinalizationNoopResult returns the canonical no-op
// outcome without a sealed capability or lease observation.
func NewMailboxSubmissionFinalizationNoopResult() MailboxSubmissionFinalizationClaimResult {
	return MailboxSubmissionFinalizationClaimResult{outcome: MailboxSubmissionFinalizationClaimOutcomeNoop}
}

// NewMailboxSubmissionFinalizationBusyResult validates and copies a canonical
// UTC busy observation before it crosses the repository boundary.
func NewMailboxSubmissionFinalizationBusyResult(busyUntil time.Time) (MailboxSubmissionFinalizationClaimResult, error) {
	if !mailboxSubmissionClaimExactTimeValid(busyUntil) {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	copyUntil := busyUntil.Round(0).UTC()
	return MailboxSubmissionFinalizationClaimResult{outcome: MailboxSubmissionFinalizationClaimOutcomeBusy, busyUntil: &copyUntil}, nil
}

// NewMailboxSubmissionFinalizationClaimedResult validates and seals one exact
// accepted-submission tuple without exposing its provider bindings.
func NewMailboxSubmissionFinalizationClaimedResult(input MailboxSubmissionFinalizationClaimedInput) (MailboxSubmissionFinalizationClaimResult, error) {
	claim, ok := mailboxSubmissionFinalizationClaimFromInput(input)
	if !ok {
		return MailboxSubmissionFinalizationClaimResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	return MailboxSubmissionFinalizationClaimResult{outcome: MailboxSubmissionFinalizationClaimOutcomeClaimed, claim: mailboxSubmissionFinalizationClaimCopy(claim)}, nil
}

// Outcome returns the closed claim outcome, or the zero value for malformed
// repository output.
func (result MailboxSubmissionFinalizationClaimResult) Outcome() MailboxSubmissionFinalizationClaimOutcome {
	if !mailboxSubmissionFinalizationClaimResultValid(result) {
		return ""
	}
	return result.outcome
}

// Claim returns a defensive sealed capability only for a valid claimed result.
func (result MailboxSubmissionFinalizationClaimResult) Claim() (MailboxSubmissionFinalizationClaim, bool) {
	if !mailboxSubmissionFinalizationClaimResultValid(result) || result.outcome != MailboxSubmissionFinalizationClaimOutcomeClaimed {
		return MailboxSubmissionFinalizationClaim{}, false
	}
	return mailboxSubmissionFinalizationClaimCopy(result.claim), true
}

// BusyUntil returns a defensive UTC lease observation only for a valid busy
// result. It is advisory and cannot be used as a later claim capability.
func (result MailboxSubmissionFinalizationClaimResult) BusyUntil() (time.Time, bool) {
	if !mailboxSubmissionFinalizationClaimResultValid(result) || result.outcome != MailboxSubmissionFinalizationClaimOutcomeBusy {
		return time.Time{}, false
	}
	return result.busyUntil.Round(0).UTC(), true
}

// MailboxSubmissionFinalizationCompletion is a sealed hydrated message and
// claim pair. Repositories re-lock and revalidate it before their atomic
// durable finalization transaction.
type MailboxSubmissionFinalizationCompletion struct {
	claim   MailboxSubmissionFinalizationClaim
	message EngineMessage
}

// MailboxSubmissionFinalizationRepository owns only advisory selection, short
// accepted claims, one completion boundary, and lease release. It performs no
// provider reads.
type MailboxSubmissionFinalizationRepository interface {
	ScanAcceptedMailboxSubmissions(context.Context, MailboxSubmissionFinalizationScanRequest) (MailboxSubmissionFinalizationScanResult, error)
	ClaimAcceptedMailboxSubmission(context.Context, MailboxSubmissionFinalizationClaimRequest) (MailboxSubmissionFinalizationClaimResult, error)
	CompleteAcceptedMailboxSubmission(context.Context, MailboxSubmissionFinalizationCompletion) error
	ReleaseAcceptedMailboxSubmission(context.Context, MailboxSubmissionFinalizationClaim) error
}

// MailboxSubmissionFinalizationErrorKind is the closed content-free finalizer
// failure family.
type MailboxSubmissionFinalizationErrorKind string

const (
	MailboxSubmissionFinalizationErrorInvalid               MailboxSubmissionFinalizationErrorKind = "invalid"
	MailboxSubmissionFinalizationErrorRepositoryUnavailable MailboxSubmissionFinalizationErrorKind = "repository_unavailable"
	MailboxSubmissionFinalizationErrorFinalizationRequired  MailboxSubmissionFinalizationErrorKind = "finalization_required"
)

const mailboxSubmissionFinalizationErrorText = "mailbox submission finalization error"

// MailboxSubmissionFinalizationError carries no raw engine, provider, or
// durable failure detail.
type MailboxSubmissionFinalizationError struct {
	kind MailboxSubmissionFinalizationErrorKind
}

func (errorValue *MailboxSubmissionFinalizationError) Error() string {
	return mailboxSubmissionFinalizationErrorText
}

// Kind returns a safe finalization error classification.
func (errorValue *MailboxSubmissionFinalizationError) Kind() MailboxSubmissionFinalizationErrorKind {
	if errorValue == nil || !validMailboxSubmissionFinalizationErrorKind(errorValue.kind) {
		return MailboxSubmissionFinalizationErrorInvalid
	}
	return errorValue.kind
}

func newMailboxSubmissionFinalizationError(kind MailboxSubmissionFinalizationErrorKind) *MailboxSubmissionFinalizationError {
	if !validMailboxSubmissionFinalizationErrorKind(kind) {
		kind = MailboxSubmissionFinalizationErrorInvalid
	}
	return &MailboxSubmissionFinalizationError{kind: kind}
}

func validMailboxSubmissionFinalizationErrorKind(kind MailboxSubmissionFinalizationErrorKind) bool {
	switch kind {
	case MailboxSubmissionFinalizationErrorInvalid,
		MailboxSubmissionFinalizationErrorRepositoryUnavailable,
		MailboxSubmissionFinalizationErrorFinalizationRequired:
		return true
	default:
		return false
	}
}

// MailboxSubmissionFinalizer hydrates one accepted provider message and proves
// it matches the immutable durable attachment multiset before completion.
type MailboxSubmissionFinalizer struct {
	repository MailboxSubmissionFinalizationRepository
	engine     MailEngine
}

// NewMailboxSubmissionFinalizer requires non-nil durable and engine ports.
func NewMailboxSubmissionFinalizer(repository MailboxSubmissionFinalizationRepository, engine MailEngine) (*MailboxSubmissionFinalizer, error) {
	if mailboxSubmissionTypedNil(repository) || mailboxSubmissionTypedNil(engine) {
		return nil, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	return &MailboxSubmissionFinalizer{repository: repository, engine: engine}, nil
}

// ScanAccepted returns a copied bounded advisory selection. It has no claim or
// engine capability and callers must finalize each retained identity separately.
func (finalizer *MailboxSubmissionFinalizer) ScanAccepted(ctx context.Context, limit int) (MailboxSubmissionFinalizationScanResult, error) {
	if finalizer == nil || mailboxSubmissionTypedNil(finalizer.repository) || ctx == nil || !mailboxSubmissionFinalizationScanRequestValid(MailboxSubmissionFinalizationScanRequest{Limit: limit}) {
		return MailboxSubmissionFinalizationScanResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{}}, nil
	}
	request := MailboxSubmissionFinalizationScanRequest{Limit: limit}
	result, err := finalizer.repository.ScanAcceptedMailboxSubmissions(ctx, request)
	if err != nil {
		return MailboxSubmissionFinalizationScanResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	if !mailboxSubmissionFinalizationScanResultValid(request, result) {
		return MailboxSubmissionFinalizationScanResult{}, newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	return mailboxSubmissionFinalizationScanResultCopy(result), nil
}

// Finalize claims exactly one accepted submission. Cancellation before its
// claim is a no-op; every failure before completion releases the short claim
// with a cancellation-detached context.
func (finalizer *MailboxSubmissionFinalizer) Finalize(ctx context.Context, submissionID uuid.UUID) error {
	if finalizer == nil || mailboxSubmissionTypedNil(finalizer.repository) || mailboxSubmissionTypedNil(finalizer.engine) || ctx == nil || submissionID == uuid.Nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	if ctx.Err() != nil {
		return nil
	}

	claimResult, err := finalizer.repository.ClaimAcceptedMailboxSubmission(ctx, MailboxSubmissionFinalizationClaimRequest{SubmissionID: submissionID})
	if err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	switch claimResult.Outcome() {
	case MailboxSubmissionFinalizationClaimOutcomeNoop, MailboxSubmissionFinalizationClaimOutcomeBusy:
		return nil
	case MailboxSubmissionFinalizationClaimOutcomeClaimed:
		// The sealed claim is extracted below.
	default:
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	claim, claimed := claimResult.Claim()
	if !claimed {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	if ctx.Err() != nil {
		return finalizer.release(ctx, claim)
	}

	message, err := finalizer.engine.GetMessage(ctx, EngineMessageReadRequest{
		AccountID:       claim.engineAccountID,
		MessageID:       claim.providerMessageID,
		MaxContentBytes: claim.engineMaxBytes,
	})
	if err != nil {
		return finalizer.release(ctx, claim)
	}
	hydrated, ok := finalizer.hydrate(ctx, claim, message)
	if !ok || ctx.Err() != nil {
		return finalizer.release(ctx, claim)
	}
	completion := MailboxSubmissionFinalizationCompletion{claim: claim, message: hydrated}
	if !mailboxSubmissionFinalizationCompletionValid(completion) {
		return finalizer.release(ctx, claim)
	}
	if err := finalizer.repository.CompleteAcceptedMailboxSubmission(ctx, mailboxSubmissionFinalizationCompletionCopy(completion)); err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
	}
	return nil
}

func (finalizer *MailboxSubmissionFinalizer) release(ctx context.Context, claim MailboxSubmissionFinalizationClaim) error {
	if !mailboxSubmissionFinalizationClaimValid(claim) {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorInvalid)
	}
	if err := finalizer.repository.ReleaseAcceptedMailboxSubmission(context.WithoutCancel(ctx), mailboxSubmissionFinalizationClaimCopy(claim)); err != nil {
		return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorRepositoryUnavailable)
	}
	return newMailboxSubmissionFinalizationError(MailboxSubmissionFinalizationErrorFinalizationRequired)
}

func (finalizer *MailboxSubmissionFinalizer) hydrate(ctx context.Context, claim MailboxSubmissionFinalizationClaim, message EngineMessage) (EngineMessage, bool) {
	if ctx.Err() != nil || !mailboxSubmissionFinalizationClaimValid(claim) || !mailboxSubmissionFinalizationMessageMatchesClaim(message, claim) {
		return EngineMessage{}, false
	}

	parts := message.Projection.Attachments
	if len(parts) != len(claim.expectedAttachments) || len(parts) > MaxEngineAttachmentItems {
		return EngineMessage{}, false
	}
	declaredSizes := make(map[string]int64, len(parts))
	var aggregate int64
	for _, part := range parts {
		if ctx.Err() != nil || part.SizeBytes < 1 || part.SizeBytes > MaxEngineAttachmentBytes || part.SizeBytes > claim.engineMaxBytes || aggregate > claim.engineMaxBytes-part.SizeBytes || aggregate > MaxEngineMessageBytes-part.SizeBytes {
			return EngineMessage{}, false
		}
		if prior, found := declaredSizes[part.BlobID]; found && prior != part.SizeBytes {
			return EngineMessage{}, false
		}
		declaredSizes[part.BlobID] = part.SizeBytes
		aggregate += part.SizeBytes
	}
	cache := make(map[string]mailboxSubmissionFinalizationBlob, len(parts))
	normalized := make([]AttachmentPolicyAttachment, len(parts))
	for index, part := range parts {
		if ctx.Err() != nil {
			return EngineMessage{}, false
		}
		blob, found := cache[part.BlobID]
		if found {
			if blob.sizeBytes != part.SizeBytes {
				return EngineMessage{}, false
			}
		} else {
			read, err := finalizer.engine.GetBlob(ctx, EngineBlobReadRequest{AccountID: claim.engineAccountID, BlobID: part.BlobID, MaxBytes: part.SizeBytes})
			if err != nil || read.AccountID != claim.engineAccountID || read.BlobID != part.BlobID || read.SizeBytes != part.SizeBytes || read.SizeBytes < 1 || read.SizeBytes > MaxEngineAttachmentBytes || int64(len(read.Bytes)) != read.SizeBytes {
				return EngineMessage{}, false
			}
			blob = mailboxSubmissionFinalizationBlob{sizeBytes: read.SizeBytes, sha256: sha256.Sum256(read.Bytes)}
			cache[part.BlobID] = blob
		}
		if len(part.SHA256) != 0 && !bytes.Equal(part.SHA256, blob.sha256[:]) {
			return EngineMessage{}, false
		}
		normalized[index] = AttachmentPolicyAttachment{
			Origin:      AttachmentByteOriginEngineMIME,
			Filename:    part.Filename,
			MediaType:   part.MediaType,
			ContentID:   part.ContentID,
			Disposition: part.Disposition,
			ScanStatus:  AttachmentScanStatusClean,
			SizeBytes:   part.SizeBytes,
			SHA256:      blob.sha256[:],
		}
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationSend,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments:        normalized,
	})
	if err != nil || !mailboxSubmissionFinalizationAttachmentMultisetMatches(decision.Attachments, claim.expectedAttachments) {
		return EngineMessage{}, false
	}

	hydrated := cloneMessage(message)
	hydrated.Projection.Origin = EngineProjectionOriginConfirmedOwnerSubmission
	if !mailboxSubmissionFinalizationMessageMatchesClaim(hydrated, claim) {
		return EngineMessage{}, false
	}
	return hydrated, true
}

type mailboxSubmissionFinalizationBlob struct {
	sizeBytes int64
	sha256    [sha256.Size]byte
}

func mailboxSubmissionFinalizationMessageMatchesClaim(message EngineMessage, claim MailboxSubmissionFinalizationClaim) bool {
	if validateEngineMessage(message) != nil || message.Projection.AccountID != claim.engineAccountID || message.Projection.MessageID != claim.engineMessageID || message.Projection.Flags.Draft || message.Projection.SentAt == nil || !syncValidTime(*message.Projection.SentAt) || message.Projection.SizeBytes > claim.engineMaxBytes || int64(messageVariableBytes(message)) > claim.engineMaxBytes {
		return false
	}
	return true
}

func mailboxSubmissionFinalizationAttachmentMultisetMatches(provider []AttachmentPolicyAttachment, expected []mailboxSubmissionFinalizationExpectedAttachment) bool {
	if len(provider) != len(expected) || len(provider) > MaxEngineAttachmentItems {
		return false
	}
	counts := make(map[mailboxSubmissionFinalizationAttachmentKey]int, len(expected))
	for _, attachment := range expected {
		counts[mailboxSubmissionFinalizationAttachmentKey{
			sha256:      attachment.sha256,
			sizeBytes:    attachment.sizeBytes,
			contentID:    attachment.contentID,
			filename:     attachment.filename,
			mediaType:    attachment.mediaType,
			disposition:  attachment.disposition,
		}]++
	}
	for _, attachment := range provider {
		if len(attachment.SHA256) != sha256.Size {
			return false
		}
		var digest [sha256.Size]byte
		copy(digest[:], attachment.SHA256)
		key := mailboxSubmissionFinalizationAttachmentKey{
			sha256:      digest,
			sizeBytes:    attachment.SizeBytes,
			contentID:    attachment.ContentID,
			filename:     attachment.Filename,
			mediaType:    attachment.MediaType,
			disposition:  attachment.Disposition,
		}
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

type mailboxSubmissionFinalizationAttachmentKey struct {
	sha256      [sha256.Size]byte
	sizeBytes    int64
	contentID    string
	filename     string
	mediaType    string
	disposition  EngineAttachmentDisposition
}

func mailboxSubmissionFinalizationScanRequestValid(request MailboxSubmissionFinalizationScanRequest) bool {
	return request.Limit >= 1 && request.Limit <= mailboxSubmissionFinalizationMaximum
}

func mailboxSubmissionFinalizationScanResultValid(request MailboxSubmissionFinalizationScanRequest, result MailboxSubmissionFinalizationScanResult) bool {
	if !mailboxSubmissionFinalizationScanRequestValid(request) || result.SubmissionIDs == nil || len(result.SubmissionIDs) > request.Limit {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.SubmissionIDs))
	for _, submissionID := range result.SubmissionIDs {
		if submissionID == uuid.Nil {
			return false
		}
		if _, duplicate := seen[submissionID]; duplicate {
			return false
		}
		seen[submissionID] = struct{}{}
	}
	return true
}

func mailboxSubmissionFinalizationScanResultCopy(result MailboxSubmissionFinalizationScanResult) MailboxSubmissionFinalizationScanResult {
	result.SubmissionIDs = append([]uuid.UUID(nil), result.SubmissionIDs...)
	return result
}

func mailboxSubmissionFinalizationClaimRequestValid(request MailboxSubmissionFinalizationClaimRequest) bool {
	return request.SubmissionID != uuid.Nil
}

func mailboxSubmissionFinalizationClaimFromInput(input MailboxSubmissionFinalizationClaimedInput) (MailboxSubmissionFinalizationClaim, bool) {
	if input.SubmissionID == uuid.Nil || input.DraftID == uuid.Nil || !mailboxSubmissionClaimFenceValid(input.Fence) || !syncValidOpaque(input.EngineAccountID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(input.ProviderSubmissionID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(input.ProviderMessageID, MaxEngineOpaqueIDBytes) || input.ProviderMessageID != input.EngineMessageID || input.EngineMaxBytes < 1 || input.EngineMaxBytes > MaxEngineMessageBytes || len(input.EngineBindingHash) != sha256.Size || len(input.DispatchAuthorizationHash) != sha256.Size || !mailboxSubmissionClaimExactTimeValid(input.DispatchAuthorizedAt) || input.DispatchAuthorizedAt.Before(time.Unix(0, 0).UTC()) || !input.DispatchAuthorizedAt.Before(input.Fence.LeaseUntil) || input.ExpectedAttachments == nil || len(input.ExpectedAttachments) > MaxEngineAttachmentItems {
		return MailboxSubmissionFinalizationClaim{}, false
	}
	claim := MailboxSubmissionFinalizationClaim{
		submissionID:         input.SubmissionID,
		draftID:              input.DraftID,
		fence:                mailboxSubmissionClaimFenceCopy(input.Fence),
		engineAccountID:      input.EngineAccountID,
		providerSubmissionID: input.ProviderSubmissionID,
		providerMessageID:    input.ProviderMessageID,
		engineMessageID:      input.EngineMessageID,
		engineMaxBytes:       input.EngineMaxBytes,
		dispatchAuthorizedAt: input.DispatchAuthorizedAt.Round(0).UTC(),
	}
	copy(claim.engineBindingHash[:], input.EngineBindingHash)
	copy(claim.dispatchAuthorizationHash[:], input.DispatchAuthorizationHash)
	if claim.engineBindingHash == ([sha256.Size]byte{}) || claim.dispatchAuthorizationHash == ([sha256.Size]byte{}) {
		return MailboxSubmissionFinalizationClaim{}, false
	}
	expected, ok := mailboxSubmissionFinalizationExpectedAttachments(input.ExpectedAttachments, input.EngineMaxBytes)
	if !ok {
		return MailboxSubmissionFinalizationClaim{}, false
	}
	claim.expectedAttachments = expected
	if !mailboxSubmissionFinalizationClaimValid(claim) {
		return MailboxSubmissionFinalizationClaim{}, false
	}
	return claim, true
}

func mailboxSubmissionFinalizationExpectedAttachments(input []MailboxSubmissionFinalizationExpectedAttachment, maximum int64) ([]mailboxSubmissionFinalizationExpectedAttachment, bool) {
	if input == nil || len(input) > MaxEngineAttachmentItems || maximum < 1 || maximum > MaxEngineMessageBytes {
		return nil, false
	}
	result := make([]mailboxSubmissionFinalizationExpectedAttachment, len(input))
	attachmentIDs := make(map[uuid.UUID]struct{}, len(input))
	objectIDs := make(map[uuid.UUID]struct{}, len(input))
	var aggregate int64
	for index, attachment := range input {
		if attachment.AttachmentID == uuid.Nil || attachment.ObjectID == uuid.Nil || attachment.AttachmentID == attachment.ObjectID || len(attachment.SHA256) != sha256.Size || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.SizeBytes > maximum || aggregate > maximum-attachment.SizeBytes || aggregate > MaxEngineMessageBytes-attachment.SizeBytes {
			return nil, false
		}
		if _, duplicate := attachmentIDs[attachment.AttachmentID]; duplicate {
			return nil, false
		}
		if _, duplicate := objectIDs[attachment.ObjectID]; duplicate {
			return nil, false
		}
		decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
			Operation:          AttachmentPolicyOperationSend,
			MaxAttachmentBytes: MaxEngineAttachmentBytes,
			Attachments: []AttachmentPolicyAttachment{{
				Origin:      AttachmentByteOriginUpload,
				Filename:    attachment.Filename,
				MediaType:   attachment.MediaType,
				ContentID:   attachment.ContentID,
				Disposition: attachment.Disposition,
				ScanStatus:  AttachmentScanStatusClean,
				SizeBytes:   attachment.SizeBytes,
				SHA256:      attachment.SHA256,
			}},
		})
		if err != nil || len(decision.Attachments) != 1 {
			return nil, false
		}
		normalized := decision.Attachments[0]
		result[index] = mailboxSubmissionFinalizationExpectedAttachment{
			attachmentID: attachment.AttachmentID,
			objectID:     attachment.ObjectID,
			sizeBytes:    normalized.SizeBytes,
			contentID:    normalized.ContentID,
			filename:     normalized.Filename,
			mediaType:    normalized.MediaType,
			disposition:  normalized.Disposition,
		}
		copy(result[index].sha256[:], normalized.SHA256)
		attachmentIDs[attachment.AttachmentID] = struct{}{}
		objectIDs[attachment.ObjectID] = struct{}{}
		aggregate += attachment.SizeBytes
	}
	sort.Slice(result, func(left, right int) bool { return bytes.Compare(result[left].attachmentID[:], result[right].attachmentID[:]) < 0 })
	return result, true
}

func mailboxSubmissionFinalizationClaimValid(claim MailboxSubmissionFinalizationClaim) bool {
	if claim.submissionID == uuid.Nil || claim.draftID == uuid.Nil || !mailboxSubmissionClaimFenceValid(claim.fence) || !syncValidOpaque(claim.engineAccountID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(claim.providerSubmissionID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(claim.providerMessageID, MaxEngineOpaqueIDBytes) || claim.providerMessageID != claim.engineMessageID || claim.engineMaxBytes < 1 || claim.engineMaxBytes > MaxEngineMessageBytes || claim.engineBindingHash == ([sha256.Size]byte{}) || claim.dispatchAuthorizationHash == ([sha256.Size]byte{}) || !mailboxSubmissionClaimExactTimeValid(claim.dispatchAuthorizedAt) || claim.dispatchAuthorizedAt.Before(time.Unix(0, 0).UTC()) || !claim.dispatchAuthorizedAt.Before(claim.fence.LeaseUntil) || claim.expectedAttachments == nil || len(claim.expectedAttachments) > MaxEngineAttachmentItems {
		return false
	}
	public := make([]MailboxSubmissionFinalizationExpectedAttachment, len(claim.expectedAttachments))
	for index, attachment := range claim.expectedAttachments {
		public[index] = MailboxSubmissionFinalizationExpectedAttachment{
			AttachmentID: attachment.attachmentID,
			ObjectID:     attachment.objectID,
			SHA256:       append([]byte(nil), attachment.sha256[:]...),
			SizeBytes:    attachment.sizeBytes,
			ContentID:    attachment.contentID,
			Filename:     attachment.filename,
			MediaType:    attachment.mediaType,
			Disposition:  attachment.disposition,
		}
	}
	_, ok := mailboxSubmissionFinalizationExpectedAttachments(public, claim.engineMaxBytes)
	return ok
}

func mailboxSubmissionFinalizationClaimCopy(claim MailboxSubmissionFinalizationClaim) MailboxSubmissionFinalizationClaim {
	claim.fence = mailboxSubmissionClaimFenceCopy(claim.fence)
	claim.dispatchAuthorizedAt = claim.dispatchAuthorizedAt.Round(0).UTC()
	claim.expectedAttachments = append([]mailboxSubmissionFinalizationExpectedAttachment(nil), claim.expectedAttachments...)
	return claim
}

func mailboxSubmissionFinalizationClaimZero(claim MailboxSubmissionFinalizationClaim) bool {
	return claim.submissionID == uuid.Nil && claim.draftID == uuid.Nil && claim.fence == (MailboxEngineCommandFence{}) && claim.engineAccountID == "" && claim.providerSubmissionID == "" && claim.providerMessageID == "" && claim.engineMessageID == "" && claim.engineMaxBytes == 0 && claim.engineBindingHash == ([sha256.Size]byte{}) && claim.dispatchAuthorizationHash == ([sha256.Size]byte{}) && claim.dispatchAuthorizedAt.IsZero() && claim.expectedAttachments == nil
}

func mailboxSubmissionFinalizationClaimResultValid(result MailboxSubmissionFinalizationClaimResult) bool {
	switch result.outcome {
	case MailboxSubmissionFinalizationClaimOutcomeNoop:
		return mailboxSubmissionFinalizationClaimZero(result.claim) && result.busyUntil == nil
	case MailboxSubmissionFinalizationClaimOutcomeBusy:
		return mailboxSubmissionFinalizationClaimZero(result.claim) && mailboxSubmissionClaimBusyUntilValid(result.busyUntil)
	case MailboxSubmissionFinalizationClaimOutcomeClaimed:
		return mailboxSubmissionFinalizationClaimValid(result.claim) && result.busyUntil == nil
	default:
		return false
	}
}

func mailboxSubmissionFinalizationCompletionValid(completion MailboxSubmissionFinalizationCompletion) bool {
	return mailboxSubmissionFinalizationClaimValid(completion.claim) && mailboxSubmissionFinalizationMessageMatchesClaim(completion.message, completion.claim) && completion.message.Projection.Origin == EngineProjectionOriginConfirmedOwnerSubmission
}

func mailboxSubmissionFinalizationCompletionCopy(completion MailboxSubmissionFinalizationCompletion) MailboxSubmissionFinalizationCompletion {
	completion.claim = mailboxSubmissionFinalizationClaimCopy(completion.claim)
	completion.message = cloneMessage(completion.message)
	return completion
}
