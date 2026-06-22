package tracker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Связи задач
// ---------------------------------------------------------------------------

// CreateIssueLink создаёт связь между двумя задачами.
func (s *TrackerService) CreateIssueLink(ctx context.Context, issueKey string, userID uuid.UUID, req CreateIssueLinkRequest) (*IssueLinkResponse, error) {
	source, err := s.repo.FindIssueByKey(ctx, issueKey)
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

	link, err := s.repo.CreateIssueLink(ctx, source.ID, targetID, req.LinkType, req.CustomType)
	if err != nil {
		return nil, err
	}

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, source.ID, userID, "linked", []FieldChange{
		{Field: "link", NewValue: req.LinkType + ":" + req.TargetIssueID},
	})
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	// Подгружаем target для ответа
	target, err := s.repo.FindIssueByID(ctx, targetID)
	if err != nil {
		slog.Error("не удалось подгрузить target для ответа", "error", err)
	}
	return toIssueLinkResponse(link, target), nil
}

// ListIssueLinks возвращает все связи задачи.
func (s *TrackerService) ListIssueLinks(ctx context.Context, issueKey string) ([]IssueLinkResponse, error) {
	source, err := s.repo.FindIssueByKey(ctx, issueKey)
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

// DeleteIssueLink удаляет связь между задачами.
func (s *TrackerService) DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error {
	return s.repo.DeleteIssueLink(ctx, linkID)
}
