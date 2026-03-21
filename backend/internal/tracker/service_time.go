package tracker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Time tracking
// ---------------------------------------------------------------------------

// LogTime записывает время на задачу.
func (s *TrackerService) LogTime(ctx context.Context, issueKey string, userID uuid.UUID, req LogTimeRequest) (*TimeLogResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	tl, err := s.repo.LogTime(ctx, existing.ID, userID, req.Minutes, req.Description)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка записи времени")
	}

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, existing.ID, userID, "time_logged", []FieldChange{
		{Field: "minutes", NewValue: req.Minutes},
	})
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	return toTimeLogResponse(tl), nil
}

// StartTimer запускает таймер на задаче.
func (s *TrackerService) StartTimer(ctx context.Context, issueKey string, userID uuid.UUID) (*TimeLogResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	tl, err := s.repo.StartTimer(ctx, existing.ID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка запуска таймера")
	}

	return toTimeLogResponse(tl), nil
}

// StopTimer останавливает таймер на задаче.
func (s *TrackerService) StopTimer(ctx context.Context, issueKey string, userID uuid.UUID) (*TimeLogResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	tl, err := s.repo.StopTimer(ctx, existing.ID, userID)
	if err != nil {
		return nil, err
	}

	return toTimeLogResponse(tl), nil
}

// ListTimeLogs возвращает все записи времени для задачи.
func (s *TrackerService) ListTimeLogs(ctx context.Context, issueKey string) ([]TimeLogResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	logs, err := s.repo.ListTimeLogs(ctx, existing.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения записей времени")
	}

	result := make([]TimeLogResponse, 0, len(logs))
	for _, l := range logs {
		result = append(result, *toTimeLogResponse(l))
	}
	return result, nil
}
