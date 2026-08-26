package messenger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

const clientNonceFingerprintVersionV1 = 1

type clientNonceFingerprintPayloadV1 struct {
	Version           int                    `json:"version"`
	Producer          TrustedMessageProducer `json:"producer"`
	ChannelID         string                 `json:"channel_id"`
	SenderID          string                 `json:"sender_id"`
	ClientNonce       string                 `json:"client_nonce"`
	MessageType       string                 `json:"message_type"`
	PollID            string                 `json:"poll_id"`
	ContentJSON       json.RawMessage        `json:"content_json"`
	ContentText       string                 `json:"content_text"`
	ThreadID          string                 `json:"thread_id"`
	AlsoSendToChannel bool                   `json:"also_send_to_channel"`
	ForwardedFromID   string                 `json:"forwarded_from_id"`
	VoiceURL          string                 `json:"voice_url"`
	VoiceDurationMs   int                    `json:"voice_duration_ms"`
	Attachments       []AttachmentRef        `json:"attachments"`
	Mentions          []string               `json:"mentions"`
	MentionGroups     []string               `json:"mention_groups"`
}

func buildPostCreateClientNonceFingerprint(request PostCreatePipelineRequest) (string, error) {
	contentJSON, err := json.Marshal(request.Message.ContentJSON)
	if err != nil {
		return "", err
	}
	attachments := append([]AttachmentRef(nil), request.Message.Attachments...)
	if len(attachments) == 0 {
		attachments = nil
	}
	payload := clientNonceFingerprintPayloadV1{
		Version:           clientNonceFingerprintVersionV1,
		Producer:          request.Producer,
		ChannelID:         request.Message.ChannelID.String(),
		SenderID:          request.Message.SenderID.String(),
		ClientNonce:       request.Message.ClientNonce,
		MessageType:       request.Message.MessageType,
		PollID:            request.Message.PollID,
		ContentJSON:       contentJSON,
		ContentText:       request.Message.ContentText,
		ThreadID:          request.Message.ThreadID,
		AlsoSendToChannel: request.Message.AlsoSendToChannel,
		ForwardedFromID:   request.Message.ForwardedFromID,
		VoiceURL:          request.Message.VoiceURL,
		VoiceDurationMs:   request.Message.VoiceDurationMs,
		Attachments:       attachments,
		Mentions:          sortedUniqueStrings(request.Message.Mentions),
		MentionGroups:     sortedUniqueStrings(request.Message.MentionGroups),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func validateCanonicalMessage(requested, canonical *MessageDocument) error {
	if canonical == nil {
		return apperrors.NewInternal("ошибка сохранения сообщения")
	}
	if canonical.DeletedAt != nil {
		return apperrors.NewConflict("client_nonce уже принадлежит удалённому сообщению")
	}

	compatible, err := messageFingerprintMatches(requested, canonical)
	if err != nil {
		return apperrors.NewInternal("ошибка сравнения сообщения")
	}
	if !compatible {
		return apperrors.NewConflict("client_nonce уже использован для другого сообщения")
	}

	return nil
}

func messageFingerprintMatches(requested, canonical *MessageDocument) (bool, error) {
	if requested.ChannelID != canonical.ChannelID ||
		requested.ThreadID != canonical.ThreadID ||
		requested.AlsoSendToChannel != canonical.AlsoSendToChannel ||
		requested.ContentText != canonical.ContentText ||
		requested.VoiceURL != canonical.VoiceURL ||
		requested.VoiceDurationMs != canonical.VoiceDurationMs ||
		!attachmentsEqual(requested.Attachments, canonical.Attachments) ||
		!stringSetsEqual(requested.Mentions, canonical.Mentions) ||
		!stringSetsEqual(requested.MentionGroups, canonical.MentionGroups) {
		return false, nil
	}

	requestedJSON, err := json.Marshal(requested.ContentJSON)
	if err != nil {
		return false, err
	}
	canonicalJSON, err := json.Marshal(canonical.ContentJSON)
	if err != nil {
		return false, err
	}

	return bytes.Equal(requestedJSON, canonicalJSON), nil
}

func attachmentsEqual(left, right []AttachmentRef) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func stringSetsEqual(left, right []string) bool {
	leftSet := sortedUniqueStrings(left)
	rightSet := sortedUniqueStrings(right)
	if len(leftSet) != len(rightSet) {
		return false
	}
	for i := range leftSet {
		if leftSet[i] != rightSet[i] {
			return false
		}
	}
	return true
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	unique := sorted[:0]
	for _, value := range sorted {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}