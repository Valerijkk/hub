package messenger

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TogglePin закрепляет или открепляет сообщение.
// Возвращает true, если закреплено, false — если откреплено.
// Требуется membership; сообщение должно принадлежать каналу (P1-BE-16).
func (s *MessengerService) TogglePin(ctx context.Context, channelID uuid.UUID, messageID string, userID uuid.UUID) (bool, error) {
	_, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionPin)
	if accessErr != nil {
		return false, accessErr
	}
	if !scoped {
		isMember, membershipErr := s.repo.IsMember(ctx, channelID, userID)
		if membershipErr != nil {
			return false, apperrors.NewInternal("ошибка проверки участия")
		}
		if !isMember {
			return false, apperrors.NewForbidden("вы не являетесь участником канала")
		}
	}

	msg, err := s.repo.FindMessageByID(ctx, messageID)
	if err != nil {
		return false, apperrors.NewNotFound("сообщение не найдено")
	}
	if msg.ChannelID != channelID {
		return false, apperrors.NewNotFound("сообщение не найдено")
	}

	existing, err := s.repo.FindPin(ctx, channelID, messageID)
	if err == nil && existing != nil {
		if err := s.repo.RemovePin(ctx, channelID, messageID); err != nil {
			return false, apperrors.NewInternal("ошибка открепления сообщения")
		}
		return false, nil
	} else if err != nil && !ent.IsNotFound(err) {
		return false, apperrors.NewInternal("ошибка поиска закреплённого сообщения")
	}

	if err := s.repo.AddPin(ctx, channelID, messageID, userID); err != nil {
		return false, apperrors.NewInternal("ошибка закрепления сообщения")
	}
	return true, nil
}

// ListPins возвращает все закреплённые сообщения канала.
// HI-01: закрепы видят только участники канала.
func (s *MessengerService) ListPins(ctx context.Context, channelID, userID uuid.UUID) ([]PinnedMessageResponse, error) {
	_, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionListPins)
	if accessErr != nil {
		return nil, accessErr
	}
	if !scoped {
		isMember, membershipErr := s.repo.IsMember(ctx, channelID, userID)
		if membershipErr != nil {
			return nil, apperrors.NewInternal("ошибка проверки участия")
		}
		if !isMember {
			return nil, apperrors.NewForbidden("вы не являетесь участником канала")
		}
	}

	pins, err := s.repo.ListPins(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения закреплённых сообщений")
	}
	result := make([]PinnedMessageResponse, 0, len(pins))
	for _, p := range pins {
		if p == nil || p.ChannelID != channelID {
			continue
		}
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
