package messenger

import (
	"context"
	"log/slog"
	"sort"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// CreateOrGetDM создаёт или возвращает существующий DM-канал.
// Дедупликация через отсортированные user ID в dm_key — один DM на пару пользователей.
func (s *MessengerService) CreateOrGetDM(ctx context.Context, userID, targetUserID uuid.UUID) (*ChannelResponse, error) {
	if userID == targetUserID {
		return nil, apperrors.NewBadRequest("нельзя создать DM с самим собой")
	}

	// dm_key из отсортированных ID — симметричный ключ для пары A-B и B-A.
	ids := []string{userID.String(), targetUserID.String()}
	sort.Strings(ids)
	dmKey := ids[0] + "-" + ids[1]

	// Ищем существующий DM по dm_key
	existing, err := s.repo.FindDMChannel(ctx, dmKey)
	if err == nil {
		if existing != nil {
			return toChannelResponse(existing), nil
		}
	} else if !ent.IsNotFound(err) {
		return nil, apperrors.NewInternal("ошибка поиска DM канала")
	}

	// Создаём новый DM
	input := CreateChannelInput{
		ChannelType: "dm",
		CreatedBy:   userID,
		DMKey:       &dmKey,
	}
	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания DM канала")
	}

	// A2W3-2-002: оба участника обязательны — иначе orphan DM без members.
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID, targetUserID}, "member"); err != nil {
		slog.Error("не удалось добавить участников DM", "error", err, "channel_id", ch.ID)
		if delErr := s.repo.DeleteChannel(ctx, ch.ID); delErr != nil {
			slog.Error("не удалось удалить orphan DM", "error", delErr, "channel_id", ch.ID)
		}
		return nil, apperrors.NewInternal("ошибка создания DM канала")
	}

	s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, ch.ID, []uuid.UUID{userID, targetUserID})
	return toChannelResponse(ch), nil
}
