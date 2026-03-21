package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// SearchMessages performs full-text search across messages.
func (s *MessengerService) SearchMessages(ctx context.Context, query, channelID string, limit int) ([]MessageResponse, error) {
	var chID *uuid.UUID
	if channelID != "" {
		parsed, err := uuid.Parse(channelID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный channel_id")
		}
		chID = &parsed
	}

	docs, err := s.repo.SearchMessages(ctx, query, chID, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка поиска сообщений")
	}
	result := make([]MessageResponse, 0, len(docs))
	for _, doc := range docs {
		result = append(result, *toMessageResponse(doc))
	}
	return result, nil
}
