package tracker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// ---------------------------------------------------------------------------
// Задачи
// ---------------------------------------------------------------------------

// CreateIssue создаёт новую задачу в проекте.
func (s *TrackerService) CreateIssue(ctx context.Context, projectID, userID uuid.UUID, req CreateIssueRequest) (*IssueResponse, error) {
	return observability.Timed(ctx, "tracker.CreateIssue", func(ctx context.Context) (*IssueResponse, error) {
		input := CreateIssueInput{
			Title:           req.Title,
			IssueType:       req.IssueType,
			Priority:        req.Priority,
			ReporterID:      userID,
			CustomFields:    req.CustomFields,
			EstimateMinutes: req.EstimateMinutes,
			StoryPoints:     req.StoryPoints,
			DueDate:         req.DueDate,
		}

		// Конвертируем чеклист из DTO в JSONB-формат
		if len(req.Checklist) > 0 {
			cl := make([]map[string]interface{}, len(req.Checklist))
			for i, item := range req.Checklist {
				cl[i] = map[string]interface{}{"id": item.ID, "text": item.Text, "checked": item.Checked}
			}
			input.Checklist = cl
		}

		if req.DescriptionJSON != nil {
			input.DescriptionJSON = *req.DescriptionJSON
		}

		if req.AssigneeID != "" {
			id, err := uuid.Parse(req.AssigneeID)
			if err != nil {
				return nil, apperrors.NewBadRequest("некорректный assignee_id")
			}
			input.AssigneeID = &id
		}
		if req.ParentID != "" {
			id, err := uuid.Parse(req.ParentID)
			if err != nil {
				return nil, apperrors.NewBadRequest("некорректный parent_id")
			}
			input.ParentID = &id
		}
		if req.StatusID != "" {
			id, err := uuid.Parse(req.StatusID)
			if err != nil {
				return nil, apperrors.NewBadRequest("некорректный status_id")
			}
			input.StatusID = &id
		}
		if len(req.Labels) > 0 {
			labelIDs := make([]uuid.UUID, 0, len(req.Labels))
			for _, l := range req.Labels {
				id, err := uuid.Parse(l)
				if err != nil {
					return nil, apperrors.NewBadRequest("некорректный label ID: " + l)
				}
				labelIDs = append(labelIDs, id)
			}
			input.Labels = labelIDs
		}

		// P1-BE-03: StatusID / ParentID / Labels должны принадлежать тому же projectID.
		if input.StatusID != nil {
			if _, err := s.repo.FindWorkflowStatusInProject(ctx, projectID, *input.StatusID); err != nil {
				if ent.IsNotFound(err) {
					return nil, apperrors.NewNotFound("статус не найден")
				}
				return nil, apperrors.NewInternal("ошибка проверки статуса")
			}
		}
		if input.ParentID != nil {
			if _, err := s.repo.FindIssueInProject(ctx, projectID, *input.ParentID); err != nil {
				if ent.IsNotFound(err) {
					return nil, apperrors.NewNotFound("родительская задача не найдена")
				}
				return nil, apperrors.NewInternal("ошибка проверки родителя")
			}
		}
		if len(input.Labels) > 0 {
			n, err := s.repo.CountLabelsInProject(ctx, projectID, input.Labels)
			if err != nil {
				return nil, apperrors.NewInternal("ошибка проверки меток")
			}
			if n != len(input.Labels) {
				return nil, apperrors.NewBadRequest("метка не принадлежит проекту")
			}
		}

		created, err := s.repo.CreateIssue(ctx, projectID, input)
		if err != nil {
			return nil, err
		}

		// Логируем активность
		_, err = s.repo.LogActivity(ctx, created.ID, userID, "created", nil)
		if err != nil {
			slog.Error("не удалось записать активность", "error", err)
		}

		return toIssueResponse(created), nil
	})
}

// CreateIssueFromPromotion создаёт задачу из промоушена сообщения мессенджера.
// Реализует messenger.IssuePromoter — RBAC-проверка проектного членства уже
// выполнена вызывающим хендлером мессенджера (он получает /messages/:id и
// заранее не знает projectID), поэтому здесь дополнительной проверки нет —
// CreateIssue в репозитории сам упадёт с FK-нарушением, если projectID
// невалидный. Для надёжной защиты от cross-project promotion вызывающий
// код должен предварительно проверить право пользователя через RBACService.
//
// Возвращает issueKey (например "HUB-432") и issue.ID для построения backlink.
func (s *TrackerService) CreateIssueFromPromotion(
	ctx context.Context,
	projectID, userID uuid.UUID,
	title, issueType string,
	descriptionJSON map[string]interface{},
) (string, string, error) {
	req := CreateIssueRequest{
		Title:     title,
		IssueType: issueType,
		Priority:  "medium",
	}
	if descriptionJSON != nil {
		req.DescriptionJSON = &descriptionJSON
	}
	resp, err := s.CreateIssue(ctx, projectID, userID, req)
	if err != nil {
		return "", "", err
	}
	return resp.IssueKey, resp.ID, nil
}

// GetIssue возвращает задачу по ключу внутри projectID (например "HUB-1").
func (s *TrackerService) GetIssue(ctx context.Context, projectID uuid.UUID, issueKey string) (*IssueResponse, error) {
	i, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}
	return toIssueResponse(i), nil
}

// ListIssues возвращает пагинированный список задач проекта.
func (s *TrackerService) ListIssues(ctx context.Context, projectID uuid.UUID, req ListIssuesRequest) (*PaginatedResponse[IssueResponse], error) {
	return observability.Timed(ctx, "tracker.ListIssues", func(ctx context.Context) (*PaginatedResponse[IssueResponse], error) {
		params := ListIssuesParams{
			Cursor:    req.Cursor,
			Limit:     req.Limit,
			Status:    req.Status,
			Priority:  req.Priority,
			IssueType: req.IssueType,
			Search:    req.Search,
			OrderBy:   req.OrderBy,
		}
		if req.Assignee != "" {
			switch req.Assignee {
			case "none", "unassigned":
				params.UnassignedOnly = true
			default:
				id, err := uuid.Parse(req.Assignee)
				if err != nil {
					return nil, apperrors.NewBadRequest("некорректный assignee ID")
				}
				params.Assignee = &id
			}
		}

		items, total, err := s.repo.ListIssues(ctx, projectID, params)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка получения списка задач")
		}

		responses := make([]IssueResponse, 0, len(items))
		for _, item := range items {
			responses = append(responses, *toIssueResponse(item))
		}

		var nextCursor string
		if len(items) > 0 {
			nextCursor = items[len(items)-1].IssueKey
		}

		limit := req.Limit
		if limit <= 0 || limit > 100 {
			limit = 50
		}

		return &PaginatedResponse[IssueResponse]{
			Items:      responses,
			NextCursor: nextCursor,
			HasMore:    len(items) == limit,
			Total:      total,
		}, nil
	})
}

// UpdateIssue обновляет задачу с проверкой workflow-переходов и оптимистичной блокировкой.
func (s *TrackerService) UpdateIssue(ctx context.Context, projectID uuid.UUID, issueKey string, userID uuid.UUID, req UpdateIssueRequest) (*IssueResponse, error) {
	// Находим задачу
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	// Если меняется статус — валидируем переход и принадлежность статуса проекту (W1-A2-009).
	if req.StatusID != nil {
		newStatusID, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный status_id")
		}
		if existing.Edges.Status == nil {
			return nil, apperrors.NewBadRequest("задача не имеет статуса")
		}
		if _, err := s.repo.FindWorkflowStatusInProject(ctx, projectID, newStatusID); err != nil {
			if ent.IsNotFound(err) {
				return nil, apperrors.NewNotFound("статус не найден")
			}
			return nil, apperrors.NewInternal("ошибка проверки статуса")
		}
		currentStatusID := existing.Edges.Status.ID
		if newStatusID != currentStatusID {
			if err := s.workflow.ValidateTransition(ctx, currentStatusID, newStatusID); err != nil {
				return nil, err
			}
		}
	}

	// Собираем изменения для лога активности
	changes := buildFieldChanges(existing, req)

	// Формируем input для обновления
	input := UpdateIssueInput{
		Title:           req.Title,
		DescriptionJSON: req.DescriptionJSON,
		Priority:        req.Priority,
		CustomFields:    req.CustomFields,
		EstimateMinutes: req.EstimateMinutes,
		StoryPoints:     req.StoryPoints,
		DueDate:         req.DueDate,
		Version:         req.Version,
	}

	// Конвертируем чеклист из DTO в JSONB-формат
	if req.Checklist != nil {
		cl := make([]map[string]interface{}, len(*req.Checklist))
		for i, item := range *req.Checklist {
			cl[i] = map[string]interface{}{"id": item.ID, "text": item.Text, "checked": item.Checked}
		}
		input.Checklist = &cl
	}

	if req.DescriptionJSON != nil {
		text := extractPlainText(*req.DescriptionJSON)
		input.DescriptionText = &text
	}
	if req.AssigneeID != nil {
		id, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный assignee_id")
		}
		input.AssigneeID = &id
	}
	if req.StatusID != nil {
		id, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный status_id")
		}
		input.StatusID = &id
	}
	if req.Labels != nil {
		labelIDs := make([]uuid.UUID, 0, len(*req.Labels))
		for _, l := range *req.Labels {
			id, err := uuid.Parse(l)
			if err != nil {
				return nil, apperrors.NewBadRequest("некорректный label ID")
			}
			labelIDs = append(labelIDs, id)
		}
		input.Labels = &labelIDs
	}

	updated, err := s.repo.UpdateIssue(ctx, existing.ID, input)
	if err != nil {
		return nil, err
	}

	// Логируем активность
	if len(changes) > 0 {
		_, err = s.repo.LogActivity(ctx, existing.ID, userID, "updated", changes)
		if err != nil {
			slog.Error("не удалось записать активность", "error", err)
		}
	}

	return toIssueResponse(updated), nil
}

// DeleteIssue удаляет задачу по ключу.
func (s *TrackerService) DeleteIssue(ctx context.Context, projectID uuid.UUID, issueKey string, userID uuid.UUID) error {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	return s.repo.DeleteIssue(ctx, existing.ID)
}

// MoveIssue перемещает задачу в другой статус на Kanban-доске.
func (s *TrackerService) MoveIssue(ctx context.Context, projectID uuid.UUID, issueKey string, userID uuid.UUID, req MoveIssueRequest) (*IssueResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, projectID, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	newStatusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный status_id")
	}

	// Валидируем переход и что статус ∈ project (W1-A2-009).
	if existing.Edges.Status == nil {
		return nil, apperrors.NewBadRequest("задача не имеет статуса")
	}
	if _, err := s.repo.FindWorkflowStatusInProject(ctx, projectID, newStatusID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("статус не найден")
		}
		return nil, apperrors.NewInternal("ошибка проверки статуса")
	}
	currentStatusID := existing.Edges.Status.ID
	if newStatusID != currentStatusID {
		if err := s.workflow.ValidateTransition(ctx, currentStatusID, newStatusID); err != nil {
			return nil, err
		}
		// Проверяем WIP-лимит целевого статуса (KANB-02)
		if s.wipChecker != nil {
			if err := s.wipChecker.CheckWIPLimit(ctx, newStatusID); err != nil {
				return nil, err
			}
		}
	}

	moved, err := s.repo.MoveIssue(ctx, existing.ID, newStatusID, req.Rank, existing.Version)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, apperrors.NewInternal("ошибка перемещения задачи")
	}

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, existing.ID, userID, "status_changed", []FieldChange{
		{Field: "status", OldValue: currentStatusID.String(), NewValue: newStatusID.String()},
	})
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	return toIssueResponse(moved), nil
}
