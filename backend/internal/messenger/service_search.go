package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// SearchMessages выполняет полнотекстовый поиск по сообщениям в каналах,
// где состоит userID. При явном channel_id — дополнительно проверяет membership.
func (s *MessengerService) SearchMessages(ctx context.Context, userID uuid.UUID, query, channelID string, limit int) ([]MessageResponse, error) {
	var scope []uuid.UUID
	if channelID != "" {
		parsed, err := uuid.Parse(channelID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный channel_id")
		}
		ok, err := s.repo.IsMember(ctx, parsed, userID)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки членства")
		}
		if !ok {
			return nil, apperrors.NewForbidden("нет доступа к каналу")
		}
		scope = []uuid.UUID{parsed}
	} else {
		channels, err := s.repo.ListUserChannels(ctx, userID)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка получения каналов")
		}
		if len(channels) == 0 {
			return []MessageResponse{}, nil
		}
		scope = make([]uuid.UUID, 0, len(channels))
		for _, ch := range channels {
			scope = append(scope, ch.ID)
		}
	}

	docs, err := s.repo.SearchMessages(ctx, query, scope, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка поиска сообщений")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}
