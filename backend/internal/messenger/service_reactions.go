package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ToggleReaction добавляет или удаляет реакцию (toggle).
// Возвращает true, если реакция добавлена, false — если удалена.
func (s *MessengerService) ToggleReaction(ctx context.Context, messageID string, userID uuid.UUID, emoji string) (bool, error) {
	exists, err := s.repo.FindReaction(ctx, messageID, userID.String(), emoji)
	if err != nil {
		return false, apperrors.NewInternal("ошибка проверки реакции")
	}

	msg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return false, apperrors.NewNotFound("сообщение не найдено")
	}

	if exists {
		if err := s.repo.RemoveReaction(ctx, messageID, userID.String(), emoji); err != nil {
			return false, apperrors.NewInternal("ошибка удаления реакции")
		}
	} else {
		if err := s.repo.AddReaction(ctx, messageID, userID.String(), emoji); err != nil {
			return false, apperrors.NewInternal("ошибка добавления реакции")
		}
	}

	added := !exists
	s.nats.PublishJSON(fmt.Sprintf(SubjectReactionToggled, msg.ChannelID.String()), ReactionToggledEvent{
		MessageID: messageID,
		ChannelID: msg.ChannelID,
		UserID:    userID,
		Emoji:     emoji,
		Added:     added,
	})

	return added, nil
}
