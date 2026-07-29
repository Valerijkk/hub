package tracker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Связи задач
// ---------------------------------------------------------------------------

// CreateIssueLink создаёт связь между двумя задачами.
func (s *TrackerService) CreateIssueLink(ctx context.Context, projectID uuid.UUID, issueKey string, userID uuid.UUID, req CreateIssueLinkRequest) (*IssueLinkResponse, error) {
	source, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("исходная задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	targetID, err := uuid.Parse(req.TargetIssueID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный target_issue_id")
	}

	// A3-011-target: цель только в том же проекте; NotFound не раскрывает существование.
	target, err := s.repo.FindIssueInProject(ctx, projectID, targetID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("целевая задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения целевой задачи")
	}

	link, err := s.repo.CreateIssueLink(ctx, source.ID, targetID, req.LinkType, req.CustomType)
	if err != nil {
		// A3W1-2-005: AppError пробрасываем; сырой repo/sql → Internal/Conflict.
		var ae *apperrors.AppError
		if errors.As(err, &ae) {
			return nil, ae
		}
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("связь этого типа между этими задачами уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания связи")
	}

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, source.ID, userID, "linked", []FieldChange{
		{Field: "link", NewValue: req.LinkType + ":" + req.TargetIssueID},
	})
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	return toIssueLinkResponse(link, target), nil
}

// ListIssueLinks возвращает все связи задачи.
func (s *TrackerService) ListIssueLinks(ctx context.Context, projectID uuid.UUID, issueKey string) ([]IssueLinkResponse, error) {
	source, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	links, err := s.repo.ListIssueLinks(ctx, source.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения связей")
	}

	result := make([]IssueLinkResponse, 0, len(links))
	for _, l := range links {
		// Определяем связанную задачу (target или source)
		var related *ent.Issue
		if l.Edges.Target != nil {
			related = l.Edges.Target
		} else if l.Edges.Source != nil {
			related = l.Edges.Source
		}
		result = append(result, *toIssueLinkResponse(l, related))
	}
	return result, nil
}

// ListProjectIssueLinks — A13-004: batch связей проекта (1 запрос вместо N+1 на Gantt).
func (s *TrackerService) ListProjectIssueLinks(ctx context.Context, projectID uuid.UUID) ([]ProjectIssueLinkResponse, error) {
	links, err := s.repo.ListProjectIssueLinks(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения связей проекта")
	}
	result := make([]ProjectIssueLinkResponse, 0, len(links))
	for _, l := range links {
		item := ProjectIssueLinkResponse{
			ID:         l.ID.String(),
			LinkType:   string(l.LinkType),
			CustomType: l.CustomType,
		}
		if l.Edges.Source != nil {
			item.SourceIssueKey = l.Edges.Source.IssueKey
		}
		if l.Edges.Target != nil {
			item.TargetIssueKey = l.Edges.Target.IssueKey
		}
		result = append(result, item)
	}
	return result, nil
}

// DeleteIssueLink удаляет связь задачи, резолвя source через projectID+issueKey (A3-011-link).
func (s *TrackerService) DeleteIssueLink(ctx context.Context, projectID uuid.UUID, issueKey string, linkID uuid.UUID) error {
	source, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	if err := s.repo.DeleteIssueLink(ctx, source.ID, linkID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("связь не найдена")
		}
		return apperrors.NewInternal("ошибка удаления связи")
	}
	return nil
}
