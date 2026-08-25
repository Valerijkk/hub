package mail

import (
	"crypto/sha256"
	"encoding/binary"
)

const (
	submissionPreflightBindingDomain        = "hub.mail.submission.preflight.binding.v1"
	submissionPreflightBindingVersion       = uint32(1)
	submissionPreflightPolicyBindingDomain  = "hub.mail.submission.preflight.policy.v1"
	submissionPreflightPolicyBindingVersion = uint32(1)
)

func preflightSubmissionForMessage(request EngineSubmissionPreflightRequest, message EngineMessage) (EngineSubmissionPreflight, error) {
	submission := request.Submission
	if err := validateSubmission(submission); err != nil {
		return EngineSubmissionPreflight{}, err
	}
	internalDomains, effectiveMaxMessageBytes, err := preflightSubmissionPolicy(request)
	if err != nil {
		return EngineSubmissionPreflight{}, err
	}
	projection := message.Projection
	if projection.AccountID != submission.AccountID || projection.MessageID != submission.MessageID || projection.Version != submission.ExpectedVersion || !projection.Flags.Draft || projection.Origin != EngineProjectionOriginConfirmedOwnerDraft {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	from, err := NormalizeAddress(projection.From)
	if err != nil || from != projection.From {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	if len(projection.To) > MaxEngineDraftRecipientItems || len(projection.CC) > MaxEngineDraftRecipientItems || len(projection.BCC) > MaxEngineDraftRecipientItems {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRecipientLimit}
	}
	to, cc, bcc, err := normalizeSendRecipients(projection.To, projection.CC, projection.BCC)
	if err != nil {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	if !preflightSubmissionRecipientsCanonical(projection.To, to) || !preflightSubmissionRecipientsCanonical(projection.CC, cc) || !preflightSubmissionRecipientsCanonical(projection.BCC, bcc) {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	recipientCount := len(to) + len(cc) + len(bcc)
	if recipientCount == 0 {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	if recipientCount > MaxEngineDraftTotalRecipientItems || recipientCount > int(request.MaxRecipients) {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRecipientLimit}
	}
	if !request.ExternalMailEnabled && hasExternalRecipient(internalDomains, to, cc, bcc) {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	if projection.SizeBytes > effectiveMaxMessageBytes {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorPayloadTooLarge}
	}
	messageBytes := int64(messageVariableBytes(message))
	if messageBytes > effectiveMaxMessageBytes {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorPayloadTooLarge}
	}
	binding := submissionPreflightBindingHash(submission, from.Email, int32(recipientCount), messageBytes)
	if binding == [sha256.Size]byte{} {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorIndeterminate}
	}
	policyBinding := submissionPreflightPolicyBindingHash(request)
	if policyBinding == [sha256.Size]byte{} {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorIndeterminate}
	}
	return EngineSubmissionPreflight{FromEmail: from.Email, RecipientCount: int32(recipientCount), MessageBytes: messageBytes, BindingHash: binding, PolicyBindingHash: policyBinding}, nil
}

func preflightSubmissionPolicy(request EngineSubmissionPreflightRequest) (map[string]struct{}, int64, error) {
	if request.PolicyMaxMessageBytes < 1_048_576 || request.PolicyMaxMessageBytes > MaxEngineMessageBytes || request.MaxRecipients < 1 || request.MaxRecipients > MaxEngineDraftTotalRecipientItems {
		return nil, 0, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	internalDomains, err := normalizeInternalDomains(request.InternalDomains)
	if err != nil {
		return nil, 0, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	for index := 1; index < len(request.InternalDomains); index++ {
		if request.InternalDomains[index-1] >= request.InternalDomains[index] {
			return nil, 0, &EngineError{Kind: EngineErrorRejectedPolicy}
		}
	}
	effectiveMaxMessageBytes := request.PolicyMaxMessageBytes
	if request.Submission.MaxMessageBytes < effectiveMaxMessageBytes {
		effectiveMaxMessageBytes = request.Submission.MaxMessageBytes
	}
	return internalDomains, effectiveMaxMessageBytes, nil
}

func preflightSubmissionRecipientsCanonical(original, canonical []EngineAddress) bool {
	if len(original) != len(canonical) {
		return false
	}
	for index := range original {
		if original[index] != canonical[index] {
			return false
		}
	}
	return true
}

func submissionPreflightBindingHash(submission EngineSubmission, fromEmail string, recipientCount int32, messageBytes int64) [sha256.Size]byte {
	frame := make([]byte, 0, len(submissionPreflightBindingDomain)+len(submission.AccountID)+len(submission.DraftKey)+len(submission.MessageID)+len(submission.ExpectedVersion)+len(submission.SubmissionKey)+len(fromEmail)+96)
	frame = preflightBindingString(frame, 1, submissionPreflightBindingDomain)
	frame = preflightBindingUint32(frame, 2, submissionPreflightBindingVersion)
	frame = preflightBindingString(frame, 3, submission.AccountID)
	frame = preflightBindingString(frame, 4, submission.DraftKey)
	frame = preflightBindingString(frame, 5, submission.MessageID)
	frame = preflightBindingString(frame, 6, submission.ExpectedVersion)
	frame = preflightBindingString(frame, 7, submission.SubmissionKey)
	frame = preflightBindingInt64(frame, 8, submission.MaxMessageBytes)
	frame = preflightBindingString(frame, 9, fromEmail)
	frame = preflightBindingInt32(frame, 10, recipientCount)
	frame = preflightBindingInt64(frame, 11, messageBytes)
	return sha256.Sum256(frame)
}

func submissionPreflightPolicyBindingHash(request EngineSubmissionPreflightRequest) [sha256.Size]byte {
	frame := make([]byte, 0, len(submissionPreflightPolicyBindingDomain)+len(request.InternalDomains)*32+64)
	frame = preflightBindingString(frame, 1, submissionPreflightPolicyBindingDomain)
	frame = preflightBindingUint32(frame, 2, submissionPreflightPolicyBindingVersion)
	frame = preflightBindingInt64(frame, 3, request.PolicyMaxMessageBytes)
	frame = preflightBindingInt32(frame, 4, request.MaxRecipients)
	frame = preflightBindingBool(frame, 5, request.ExternalMailEnabled)
	frame = preflightBindingUint32(frame, 6, uint32(len(request.InternalDomains)))
	for _, domain := range request.InternalDomains {
		frame = preflightBindingString(frame, 7, domain)
	}
	return sha256.Sum256(frame)
}

func preflightBindingString(frame []byte, field byte, value string) []byte {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	frame = append(frame, field)
	frame = append(frame, length[:]...)
	return append(frame, value...)
}

func preflightBindingInt32(frame []byte, field byte, value int32) []byte {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], uint32(value))
	return append(append(frame, field), encoded[:]...)
}

func preflightBindingUint32(frame []byte, field byte, value uint32) []byte {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	return append(append(frame, field), encoded[:]...)
}

func preflightBindingBool(frame []byte, field byte, value bool) []byte {
	encoded := byte(0)
	if value {
		encoded = 1
	}
	return append(frame, field, encoded)
}

func preflightBindingInt64(frame []byte, field byte, value int64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	return append(append(frame, field), encoded[:]...)
}
