package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TogglePin pins or unpins a message.
// Returns true if pinned, false if unpinned.
func (s *MessengerService) TogglePin(ctx context.Context, channelID uuid.UUID, messageID string, userID uuid.UUID) (bool, error) {
	existing, err := s.repo.FindPin(ctx, channelID, messageID)
	if err == nil && existing != nil {
		// Unpin
		if err := s.repo.RemovePin(ctx, channelID, messageID); err != nil {
			return false, apperrors.NewInternal("ошибка открепления сообщения")
		}
		return false, nil
	}

	// Pin
	if err := s.repo.AddPin(ctx, channelID, messageID, userID); err != nil {
		return false, apperrors.NewInternal("ошибка закрепления сообщения")
	}
	return true, nil
}

// ListPins returns all pinned messages for a channel.
func (s *MessengerService) ListPins(ctx context.Context, channelID uuid.UUID) ([]PinnedMessageResponse, error) {
	pins, err := s.repo.ListPins(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения закреплённых сообщений")
	}
	result := make([]PinnedMessageResponse, 0, len(pins))
	for _, p := range pins {
		result = append(result, PinnedMessageResponse{
			ID:        p.ID.String(),
			MessageID: p.MessageID,
			ChannelID: p.ChannelID.String(),
			PinnedBy:  p.PinnedBy.String(),
			PinnedAt:  p.PinnedAt,
		})
	}
	return result, nil
}
