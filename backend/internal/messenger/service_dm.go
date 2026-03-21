package messenger

import (
	"context"
	"log/slog"
	"sort"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// CreateOrGetDM creates or returns an existing DM channel.
// Deduplication via sorted user IDs in dm_key.
func (s *MessengerService) CreateOrGetDM(ctx context.Context, userID, targetUserID uuid.UUID) (*ChannelResponse, error) {
	if userID == targetUserID {
		return nil, apperrors.NewBadRequest("нельзя создать DM с самим собой")
	}

	// Build dm_key from sorted IDs
	ids := []string{userID.String(), targetUserID.String()}
	sort.Strings(ids)
	dmKey := ids[0] + "-" + ids[1]

	// Look for existing DM
	existing, err := s.repo.FindDMChannel(ctx, dmKey)
	if err == nil {
		return toChannelResponse(existing), nil
	}

	// Create new DM
	input := CreateChannelInput{
		ChannelType: "dm",
		CreatedBy:   userID,
		DMKey:       &dmKey,
	}
	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания DM канала")
	}

	// Add both participants
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID, targetUserID}, "member"); err != nil {
		slog.Error("не удалось добавить участников DM", "error", err)
	}

	return toChannelResponse(ch), nil
}
