package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// SearchMessages выполняет полнотекстовый поиск по заранее авторизованной области каналов.
func (s *MessengerService) SearchMessages(ctx context.Context, userID uuid.UUID, query, channelID string, limit int) ([]MessageResponse, error) {
	var scope []uuid.UUID
	if channelID != "" {
		parsed, err := uuid.Parse(channelID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный channel_id")
		}

		_, scoped, accessErr := s.resolveChannelAction(ctx, userID, parsed, ChannelActionReadTimeline)
		if accessErr != nil {
			return nil, accessErr
		}
		if !scoped {
			ok, membershipErr := s.repo.IsMember(ctx, parsed, userID)
			if membershipErr != nil {
				return nil, apperrors.NewInternal("ошибка проверки членства")
			}
			if !ok {
				return nil, apperrors.NewForbidden("нет доступа к каналу")
			}
		}
		scope = []uuid.UUID{parsed}
	} else {
		if s.messageSearchScope == nil {
			return []MessageResponse{}, apperrors.NewForbidden("область поиска недоступна")
		}
		authorized, err := s.messageSearchScope.ResolveAuthorizedMessageSearchScope(ctx, userID)
		if err != nil {
			return []MessageResponse{}, apperrors.NewForbidden("область поиска недоступна")
		}
		if authorized.TenantID == uuid.Nil || authorized.UserID != userID || len(authorized.Targets) == 0 {
			return []MessageResponse{}, apperrors.NewForbidden("область поиска недоступна")
		}

		scope = make([]uuid.UUID, 0, len(authorized.Targets))
		seen := make(map[uuid.UUID]struct{}, len(authorized.Targets))
		for _, target := range authorized.Targets {
			if target.ChannelID == uuid.Nil {
				return []MessageResponse{}, apperrors.NewForbidden("область поиска недоступна")
			}
			if _, duplicate := seen[target.ChannelID]; duplicate {
				continue
			}
			seen[target.ChannelID] = struct{}{}
			scope = append(scope, target.ChannelID)
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
