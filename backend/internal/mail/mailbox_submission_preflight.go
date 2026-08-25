package mail

import (
	"crypto/sha256"
	"encoding/binary"
)

const (
	submissionPreflightBindingDomain  = "hub.mail.submission.preflight.binding.v1"
	submissionPreflightBindingVersion = uint32(1)
)

func preflightSubmissionForMessage(submission EngineSubmission, message EngineMessage) (EngineSubmissionPreflight, error) {
	if err := validateSubmission(submission); err != nil {
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
	recipientCount, err := preflightSubmissionRecipientCount(projection)
	if err != nil {
		return EngineSubmissionPreflight{}, err
	}
	if projection.SizeBytes > submission.MaxMessageBytes {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorPayloadTooLarge}
	}
	messageBytes := int64(messageVariableBytes(message))
	if messageBytes > submission.MaxMessageBytes {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorPayloadTooLarge}
	}
	binding := submissionPreflightBindingHash(submission, from.Email, recipientCount, messageBytes)
	if binding == [sha256.Size]byte{} {
		return EngineSubmissionPreflight{}, &EngineError{Kind: EngineErrorIndeterminate}
	}
	return EngineSubmissionPreflight{FromEmail: from.Email, RecipientCount: recipientCount, MessageBytes: messageBytes, BindingHash: binding}, nil
}

func preflightSubmissionRecipientCount(projection EngineMessageProjection) (int32, error) {
	collections := [][]EngineAddress{projection.To, projection.CC, projection.BCC}
	count := 0
	seen := make(map[string]struct{}, MaxEngineDraftTotalRecipientItems)
	for _, collection := range collections {
		if len(collection) > MaxEngineDraftRecipientItems {
			return 0, &EngineError{Kind: EngineErrorRecipientLimit}
		}
		for _, recipient := range collection {
			canonical, err := NormalizeAddress(recipient)
			if err != nil || canonical != recipient {
				return 0, &EngineError{Kind: EngineErrorRejectedPolicy}
			}
			if _, duplicate := seen[canonical.Email]; duplicate {
				return 0, &EngineError{Kind: EngineErrorRejectedPolicy}
			}
			seen[canonical.Email] = struct{}{}
			count++
			if count > MaxEngineDraftTotalRecipientItems {
				return 0, &EngineError{Kind: EngineErrorRecipientLimit}
			}
		}
	}
	if count == 0 {
		return 0, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	return int32(count), nil
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

func preflightBindingInt64(frame []byte, field byte, value int64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	return append(append(frame, field), encoded[:]...)
}
