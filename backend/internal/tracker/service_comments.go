package tracker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Комментарии
// ---------------------------------------------------------------------------

// CreateComment добавляет комментарий к задаче.
func (s *TrackerService) CreateComment(ctx context.Context, projectID uuid.UUID, issueKey string, userID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	input := CreateCommentInput{
		BodyJSON:   req.BodyJSON,
		BodyText:   req.BodyText,
		IsInternal: req.IsInternal,
		Mentions:   req.Mentions,
	}

	c, err := s.repo.CreateComment(ctx, existing.ID, userID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания комментария")
	}

	if !req.IsInternal {
		s.mirrorIssueCommentsToThreads(ctx, existing.ID, userID, req.BodyText)
	}

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, existing.ID, userID, "commented", nil)
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	return toCommentResponse(c), nil
}

// ListComments возвращает все комментарии задачи.
func (s *TrackerService) ListComments(ctx context.Context, projectID uuid.UUID, issueKey string) ([]CommentResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	comments, err := s.repo.ListComments(ctx, existing.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения комментариев")
	}

	result := make([]CommentResponse, 0, len(comments))
	for _, c := range comments {
		result = append(result, *toCommentResponse(c))
	}
	return result, nil
}

// UpdateComment обновляет комментарий автора на задаче (A3-011).
func (s *TrackerService) UpdateComment(ctx context.Context, projectID uuid.UUID, issueKey string, userID, commentID uuid.UUID, req UpdateCommentRequest) (*CommentResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}
	input := UpdateCommentInput{
		BodyJSON: req.BodyJSON,
		BodyText: req.BodyText,
	}
	c, err := s.repo.UpdateComment(ctx, existing.ID, userID, commentID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("комментарий не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления комментария")
	}
	return toCommentResponse(c), nil
}

// DeleteComment удаляет комментарий автора на задаче (A3-011).
func (s *TrackerService) DeleteComment(ctx context.Context, projectID uuid.UUID, issueKey string, userID, commentID uuid.UUID) error {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	if err := s.repo.DeleteComment(ctx, existing.ID, userID, commentID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("комментарий не найден")
		}
		return apperrors.NewInternal("ошибка удаления комментария")
	}
	return nil
}

// ToggleReaction добавляет или удаляет реакцию на комментарий.
func (s *TrackerService) ToggleReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) (bool, error) {
	return s.repo.ToggleReaction(ctx, commentID, userID, emoji)
}
