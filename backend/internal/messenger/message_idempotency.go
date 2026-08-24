package messenger

import (
	"bytes"
	"encoding/json"
	"sort"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

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