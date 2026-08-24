package mail

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
)

// mailboxDraftAttachmentOpen is deliberately package-private: only the draft
// materializer may turn an approved storage reference into a byte stream.
type mailboxDraftAttachmentOpen func(context.Context) (io.ReadCloser, error)

type mailboxDraftAttachmentBinding struct {
	AccountID  string
	DraftKey   string
	MutationID string
}

type mailboxDraftAttachmentSourceSpec struct {
	Attachment EngineDraftAttachment
	SizeBytes  int64
	ScanStatus AttachmentScanStatus
	Open       mailboxDraftAttachmentOpen
}

type mailboxDraftAttachmentSource struct {
	attachment EngineDraftAttachment
	sizeBytes  int64
	scanStatus AttachmentScanStatus
	open       mailboxDraftAttachmentOpen
	consumed   atomic.Bool
}

// mailboxDraftAttachmentCapability is an unexported, context-scoped proof
// that the public attachment metadata is backed by approved stream sources.
// It intentionally contains neither object locators nor a global lookup key.
type mailboxDraftAttachmentCapability struct {
	binding            mailboxDraftAttachmentBinding
	maxAttachmentBytes int64
	sources            []*mailboxDraftAttachmentSource
}

type mailboxDraftAttachmentCapabilityContextKey struct{}

func newMailboxDraftAttachmentCapability(binding mailboxDraftAttachmentBinding, maximum int64, specs []mailboxDraftAttachmentSourceSpec) (*mailboxDraftAttachmentCapability, error) {
	if !mailboxDraftAttachmentBindingValid(binding) || maximum < 1 || maximum > MaxEngineAttachmentBytes || len(specs) == 0 || len(specs) > MaxEngineAttachmentItems {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}

	sources := make([]*mailboxDraftAttachmentSource, len(specs))
	policyAttachments := make([]AttachmentPolicyAttachment, len(specs))
	seenKeys := make(map[string]struct{}, len(specs))
	seenCIDs := make(map[string]struct{}, len(specs))
	for index, spec := range specs {
		attachment := cloneMailboxDraftAttachment(spec.Attachment)
		if spec.Open == nil || len(attachment.Bytes) != 0 || !validOpaqueID(attachment.AttachmentKey) || !mailboxDraftContentIDValid(attachment.ContentID) || mailboxDraftAttachmentSizeInvalid(spec.SizeBytes, maximum) || attachment.SHA256 == ([32]byte{}) || mailboxDraftAttachmentScanStatusInvalid(spec.ScanStatus) {
			return nil, fakeEngineError(EngineErrorRejectedPolicy)
		}
		if _, duplicate := seenKeys[attachment.AttachmentKey]; duplicate {
			return nil, fakeEngineError(EngineErrorRejectedPolicy)
		}
		seenKeys[attachment.AttachmentKey] = struct{}{}
		if attachment.ContentID != "" {
			if _, duplicate := seenCIDs[attachment.ContentID]; duplicate {
				return nil, fakeEngineError(EngineErrorRejectedPolicy)
			}
			seenCIDs[attachment.ContentID] = struct{}{}
		}
		if attachment.Disposition == EngineAttachmentDispositionInline && attachment.ContentID == "" {
			return nil, fakeEngineError(EngineErrorRejectedPolicy)
		}
		policyAttachments[index] = AttachmentPolicyAttachment{
			Origin:      AttachmentByteOriginUpload,
			Filename:    attachment.Filename,
			MediaType:   attachment.MediaType,
			ContentID:   attachment.ContentID,
			Disposition: attachment.Disposition,
			ScanStatus:  spec.ScanStatus,
			SizeBytes:   spec.SizeBytes,
			SHA256:      append([]byte(nil), attachment.SHA256[:]...),
		}
		sources[index] = &mailboxDraftAttachmentSource{attachment: attachment, sizeBytes: spec.SizeBytes, scanStatus: spec.ScanStatus, open: spec.Open}
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{Operation: AttachmentPolicyOperationSend, MaxAttachmentBytes: maximum, Attachments: policyAttachments})
	if err != nil || len(decision.Attachments) != len(sources) {
		return nil, mailboxDraftAttachmentPolicyEngineError(err)
	}
	for index, normalized := range decision.Attachments {
		attachment := sources[index].attachment
		if normalized.Filename != attachment.Filename || normalized.MediaType != attachment.MediaType || normalized.ContentID != attachment.ContentID || normalized.Disposition != attachment.Disposition || normalized.ScanStatus != sources[index].scanStatus || normalized.SizeBytes != sources[index].sizeBytes || !bytes.Equal(normalized.SHA256, attachment.SHA256[:]) {
			return nil, fakeEngineError(EngineErrorRejectedPolicy)
		}
	}
	return &mailboxDraftAttachmentCapability{binding: binding, maxAttachmentBytes: maximum, sources: sources}, nil
}

func mailboxDraftAttachmentCapabilityContext(ctx context.Context, capability *mailboxDraftAttachmentCapability) context.Context {
	if ctx == nil || capability == nil {
		return ctx
	}
	return context.WithValue(ctx, mailboxDraftAttachmentCapabilityContextKey{}, capability)
}

func mailboxDraftAttachmentCapabilityFromContext(ctx context.Context, draft EngineDraft) (*mailboxDraftAttachmentCapability, bool, error) {
	if ctx == nil {
		return nil, false, nil
	}
	value := ctx.Value(mailboxDraftAttachmentCapabilityContextKey{})
	if value == nil {
		return nil, false, nil
	}
	capability, ok := value.(*mailboxDraftAttachmentCapability)
	if !ok || capability == nil || !mailboxDraftAttachmentCapabilityMatches(capability, draft) {
		return nil, true, fakeEngineError(EngineErrorRejectedPolicy)
	}
	return capability, true, nil
}

func mailboxDraftAttachmentCapabilityMatches(capability *mailboxDraftAttachmentCapability, draft EngineDraft) bool {
	if capability == nil || !mailboxDraftAttachmentBindingValid(capability.binding) || capability.binding.AccountID != draft.AccountID || capability.binding.DraftKey != draft.DraftKey || capability.binding.MutationID != draft.MutationID || capability.maxAttachmentBytes < 1 || capability.maxAttachmentBytes > MaxEngineAttachmentBytes || len(capability.sources) == 0 || len(capability.sources) != len(draft.Message.Attachments) || len(capability.sources) > MaxEngineAttachmentItems {
		return false
	}
	seenKeys := make(map[string]struct{}, len(capability.sources))
	seenCIDs := make(map[string]struct{}, len(capability.sources))
	for index, source := range capability.sources {
		if source == nil || source.open == nil || source.sizeBytes < 1 || source.sizeBytes > capability.maxAttachmentBytes || source.sizeBytes > MaxEngineAttachmentBytes || source.scanStatus != AttachmentScanStatusClean || source.attachment.SHA256 == ([32]byte{}) || len(source.attachment.Bytes) != 0 || !mailboxDraftAttachmentMetadataEqual(source.attachment, draft.Message.Attachments[index]) || !mailboxDraftContentIDValid(source.attachment.ContentID) {
			return false
		}
		if _, duplicate := seenKeys[source.attachment.AttachmentKey]; duplicate {
			return false
		}
		seenKeys[source.attachment.AttachmentKey] = struct{}{}
		if source.attachment.ContentID != "" {
			if _, duplicate := seenCIDs[source.attachment.ContentID]; duplicate {
				return false
			}
			seenCIDs[source.attachment.ContentID] = struct{}{}
		}
		if source.attachment.Disposition == EngineAttachmentDispositionInline && source.attachment.ContentID == "" {
			return false
		}
	}
	return true
}

func (source *mailboxDraftAttachmentSource) openOnce(ctx context.Context) (io.ReadCloser, error) {
	if source == nil || ctx == nil || source.open == nil || !source.consumed.CompareAndSwap(false, true) {
		return nil, fakeEngineError(EngineErrorConflict)
	}
	reader, err := source.open(ctx)
	if err != nil || reader == nil {
		return nil, fakeEngineError(EngineErrorUnavailable)
	}
	return reader, nil
}

func mailboxDraftAttachmentBindingValid(binding mailboxDraftAttachmentBinding) bool {
	return validAccountID(binding.AccountID) && validSafeText(binding.DraftKey, MaxEngineOpaqueIDBytes, true) && validSafeText(binding.MutationID, MaxEngineOpaqueIDBytes, true)
}

func mailboxDraftAttachmentSizeInvalid(size, maximum int64) bool {
	return size < 1 || size > maximum || size > MaxEngineAttachmentBytes
}

func mailboxDraftAttachmentScanStatusInvalid(status AttachmentScanStatus) bool {
	return status != AttachmentScanStatusClean
}

func mailboxDraftAttachmentMetadataEqual(left, right EngineDraftAttachment) bool {
	return left.AttachmentKey == right.AttachmentKey && left.ContentID == right.ContentID && left.Filename == right.Filename && left.MediaType == right.MediaType && left.Disposition == right.Disposition && left.SHA256 == right.SHA256 && len(right.Bytes) == 0
}

func cloneMailboxDraftAttachment(value EngineDraftAttachment) EngineDraftAttachment {
	value.Bytes = append([]byte(nil), value.Bytes...)
	return value
}

func mailboxDraftAttachmentPolicyEngineError(err error) error {
	if policy, ok := err.(*AttachmentPolicyError); ok && policy.Kind == AttachmentErrorTooLarge {
		return fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return fakeEngineError(EngineErrorRejectedPolicy)
}
