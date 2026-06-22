package tracker

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Активность
// ---------------------------------------------------------------------------

// GetActivity возвращает лог активности задачи.
func (s *TrackerService) GetActivity(ctx context.Context, issueKey string) ([]ActivityLogResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	activities, err := s.repo.ListActivities(ctx, existing.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения активности")
	}

	result := make([]ActivityLogResponse, 0, len(activities))
	for _, a := range activities {
		result = append(result, *toActivityLogResponse(a))
	}
	return result, nil
}
