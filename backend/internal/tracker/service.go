package tracker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TrackerService реализует бизнес-логику модуля трекера задач.
// Оркестрирует вызовы репозитория, workflow-движка, логирование активности и outbox-события.
type TrackerService struct {
	repo     Repository
	workflow *WorkflowEngine
}

// NewTrackerService создаёт новый сервис трекера.
func NewTrackerService(repo Repository) *TrackerService {
	wf := NewWorkflowEngine(repo)
	return &TrackerService{repo: repo, workflow: wf}
}

// ---------------------------------------------------------------------------
// Issues
// ---------------------------------------------------------------------------

// CreateIssue создаёт новую задачу в проекте.
func (s *TrackerService) CreateIssue(ctx context.Context, projectID, userID uuid.UUID, req CreateIssueRequest) (*IssueResponse, error) {
	input := CreateIssueInput{
		Title:       req.Title,
		IssueType:   req.IssueType,
		Priority:    req.Priority,
		ReporterID:  userID,
		CustomFields: req.CustomFields,
		EstimateMinutes: req.EstimateMinutes,
		DueDate:     req.DueDate,
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
}

// GetIssue возвращает задачу по ключу (например "HUB-1").
func (s *TrackerService) GetIssue(ctx context.Context, issueKey string) (*IssueResponse, error) {
	i, err := s.repo.FindIssueByKey(ctx, issueKey)
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
		id, err := uuid.Parse(req.Assignee)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный assignee ID")
		}
		params.Assignee = &id
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
}

// UpdateIssue обновляет задачу с проверкой workflow-переходов и оптимистичной блокировкой.
func (s *TrackerService) UpdateIssue(ctx context.Context, issueKey string, userID uuid.UUID, req UpdateIssueRequest) (*IssueResponse, error) {
	// Находим задачу
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения задачи")
	}

	// Если меняется статус — валидируем переход
	if req.StatusID != nil {
		newStatusID, err := uuid.Parse(*req.StatusID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный status_id")
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
		DueDate:         req.DueDate,
		Version:         req.Version,
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
func (s *TrackerService) DeleteIssue(ctx context.Context, issueKey string, userID uuid.UUID) error {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	return s.repo.DeleteIssue(ctx, existing.ID)
}

// MoveIssue перемещает задачу в другой статус на Kanban-доске.
func (s *TrackerService) MoveIssue(ctx context.Context, issueKey string, userID uuid.UUID, req MoveIssueRequest) (*IssueResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
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

	// Валидируем переход
	currentStatusID := existing.Edges.Status.ID
	if newStatusID != currentStatusID {
		if err := s.workflow.ValidateTransition(ctx, currentStatusID, newStatusID); err != nil {
			return nil, err
		}
	}

	moved, err := s.repo.MoveIssue(ctx, existing.ID, newStatusID, req.Rank)
	if err != nil {
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

// ---------------------------------------------------------------------------
// Issue links
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
	target, _ := s.repo.FindIssueByID(ctx, targetID)
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

// ---------------------------------------------------------------------------
// Workflow
// ---------------------------------------------------------------------------

// ListWorkflowStatuses возвращает все статусы workflow проекта.
func (s *TrackerService) ListWorkflowStatuses(ctx context.Context, projectID uuid.UUID) ([]WorkflowStatusResponse, error) {
	statuses, err := s.repo.ListWorkflowStatuses(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения статусов")
	}
	result := make([]WorkflowStatusResponse, 0, len(statuses))
	for _, st := range statuses {
		result = append(result, *toWorkflowStatusResponse(st))
	}
	return result, nil
}

// CreateWorkflowStatus создаёт новый статус workflow.
func (s *TrackerService) CreateWorkflowStatus(ctx context.Context, projectID uuid.UUID, req CreateWorkflowStatusRequest) (*WorkflowStatusResponse, error) {
	st, err := s.repo.CreateWorkflowStatus(ctx, projectID, req.Name, req.Category, req.Color, req.SortOrder)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("статус с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания статуса")
	}
	return toWorkflowStatusResponse(st), nil
}

// UpdateWorkflowStatus обновляет статус workflow.
func (s *TrackerService) UpdateWorkflowStatus(ctx context.Context, statusID uuid.UUID, req CreateWorkflowStatusRequest) (*WorkflowStatusResponse, error) {
	st, err := s.repo.UpdateWorkflowStatus(ctx, statusID, &req.Name, &req.Color, &req.SortOrder)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("статус не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления статуса")
	}
	return toWorkflowStatusResponse(st), nil
}

// DeleteWorkflowStatus удаляет статус workflow.
func (s *TrackerService) DeleteWorkflowStatus(ctx context.Context, statusID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowStatus(ctx, statusID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("статус не найден")
		}
		return apperrors.NewInternal("ошибка удаления статуса")
	}
	return nil
}

// ListWorkflowTransitions возвращает все переходы workflow проекта.
func (s *TrackerService) ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]WorkflowTransitionResponse, error) {
	transitions, err := s.repo.ListWorkflowTransitions(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения переходов")
	}
	result := make([]WorkflowTransitionResponse, 0, len(transitions))
	for _, t := range transitions {
		result = append(result, WorkflowTransitionResponse{
			ID:           t.ID.String(),
			FromStatusID: t.Edges.FromStatus.ID.String(),
			ToStatusID:   t.Edges.ToStatus.ID.String(),
		})
	}
	return result, nil
}

// CreateWorkflowTransition создаёт разрешённый переход между статусами.
func (s *TrackerService) CreateWorkflowTransition(ctx context.Context, req CreateWorkflowTransitionRequest) (*WorkflowTransitionResponse, error) {
	fromID, err := uuid.Parse(req.FromStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный from_status_id")
	}
	toID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный to_status_id")
	}

	t, err := s.repo.CreateWorkflowTransition(ctx, fromID, toID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("такой переход уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания перехода")
	}
	return &WorkflowTransitionResponse{
		ID:           t.ID.String(),
		FromStatusID: req.FromStatusID,
		ToStatusID:   req.ToStatusID,
	}, nil
}

// DeleteWorkflowTransition удаляет переход workflow.
func (s *TrackerService) DeleteWorkflowTransition(ctx context.Context, transitionID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowTransition(ctx, transitionID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("переход не найден")
		}
		return apperrors.NewInternal("ошибка удаления перехода")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

// CreateComment добавляет комментарий к задаче.
func (s *TrackerService) CreateComment(ctx context.Context, issueKey string, userID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
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

	// Логируем активность
	_, err = s.repo.LogActivity(ctx, existing.ID, userID, "commented", nil)
	if err != nil {
		slog.Error("не удалось записать активность", "error", err)
	}

	return toCommentResponse(c), nil
}

// ListComments возвращает все комментарии задачи.
func (s *TrackerService) ListComments(ctx context.Context, issueKey string) ([]CommentResponse, error) {
	existing, err := s.repo.FindIssueByKey(ctx, issueKey)
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

// UpdateComment обновляет комментарий.
func (s *TrackerService) UpdateComment(ctx context.Context, commentID uuid.UUID, req UpdateCommentRequest) (*CommentResponse, error) {
	input := UpdateCommentInput{
		BodyJSON: req.BodyJSON,
		BodyText: req.BodyText,
	}
	c, err := s.repo.UpdateComment(ctx, commentID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("комментарий не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления комментария")
	}
	return toCommentResponse(c), nil
}

// DeleteComment удаляет комментарий.
func (s *TrackerService) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	if err := s.repo.DeleteComment(ctx, commentID); err != nil {
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

// ---------------------------------------------------------------------------
// Labels
// ---------------------------------------------------------------------------

// CreateLabel создаёт метку в проекте.
func (s *TrackerService) CreateLabel(ctx context.Context, projectID uuid.UUID, req CreateLabelRequest) (*LabelResponse, error) {
	var groupID *uuid.UUID
	if req.GroupID != "" {
		id, err := uuid.Parse(req.GroupID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный group_id")
		}
		groupID = &id
	}

	l, err := s.repo.CreateLabel(ctx, projectID, req.Name, req.Color, groupID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("метка с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания метки")
	}

	return toLabelResponse(l), nil
}

// ListLabels возвращает все метки проекта.
func (s *TrackerService) ListLabels(ctx context.Context, projectID uuid.UUID) ([]LabelResponse, error) {
	labels, err := s.repo.ListLabels(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения меток")
	}
	result := make([]LabelResponse, 0, len(labels))
	for _, l := range labels {
		result = append(result, *toLabelResponse(l))
	}
	return result, nil
}

// DeleteLabel удаляет метку.
func (s *TrackerService) DeleteLabel(ctx context.Context, labelID uuid.UUID) error {
	if err := s.repo.DeleteLabel(ctx, labelID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("метка не найдена")
		}
		return apperrors.NewInternal("ошибка удаления метки")
	}
	return nil
}

// CreateLabelGroup создаёт группу меток.
func (s *TrackerService) CreateLabelGroup(ctx context.Context, projectID uuid.UUID, req CreateLabelGroupRequest) (*LabelGroupResponse, error) {
	g, err := s.repo.CreateLabelGroup(ctx, projectID, req.Name)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("группа меток с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания группы меток")
	}
	return toLabelGroupResponse(g), nil
}

// ListLabelGroups возвращает все группы меток проекта.
func (s *TrackerService) ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]LabelGroupResponse, error) {
	groups, err := s.repo.ListLabelGroups(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения групп меток")
	}
	result := make([]LabelGroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, *toLabelGroupResponse(g))
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Boards
// ---------------------------------------------------------------------------

// CreateBoard создаёт доску в проекте.
func (s *TrackerService) CreateBoard(ctx context.Context, projectID uuid.UUID, req CreateBoardRequest) (*BoardResponse, error) {
	b, err := s.repo.CreateBoard(ctx, projectID, req.Name, req.Filters)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания доски")
	}
	return toBoardResponse(b), nil
}

// ListBoards возвращает все доски проекта.
func (s *TrackerService) ListBoards(ctx context.Context, projectID uuid.UUID) ([]BoardResponse, error) {
	boards, err := s.repo.ListBoards(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения досок")
	}
	result := make([]BoardResponse, 0, len(boards))
	for _, b := range boards {
		result = append(result, *toBoardResponse(b))
	}
	return result, nil
}

// GetBoardData возвращает данные доски с задачами, сгруппированными по статусам.
func (s *TrackerService) GetBoardData(ctx context.Context, boardID uuid.UUID) (*BoardDataResponse, error) {
	b, err := s.repo.GetBoardData(ctx, boardID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("доска не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения данных доски")
	}
	return toBoardDataResponse(b), nil
}

// ---------------------------------------------------------------------------
// Custom fields
// ---------------------------------------------------------------------------

// CreateCustomField создаёт определение кастомного поля.
func (s *TrackerService) CreateCustomField(ctx context.Context, projectID uuid.UUID, req CreateCustomFieldRequest) (*CustomFieldDefinitionResponse, error) {
	input := CreateCustomFieldInput{
		Name:      req.Name,
		FieldType: req.FieldType,
		Options:   req.Options,
		Required:  req.Required,
	}

	cf, err := s.repo.CreateCustomField(ctx, projectID, input)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("поле с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания кастомного поля")
	}
	return toCustomFieldResponse(cf), nil
}

// ListCustomFields возвращает все определения кастомных полей проекта.
func (s *TrackerService) ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]CustomFieldDefinitionResponse, error) {
	fields, err := s.repo.ListCustomFields(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения кастомных полей")
	}
	result := make([]CustomFieldDefinitionResponse, 0, len(fields))
	for _, f := range fields {
		result = append(result, *toCustomFieldResponse(f))
	}
	return result, nil
}

// DeleteCustomField удаляет определение кастомного поля.
func (s *TrackerService) DeleteCustomField(ctx context.Context, fieldID uuid.UUID) error {
	if err := s.repo.DeleteCustomField(ctx, fieldID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("поле не найдено")
		}
		return apperrors.NewInternal("ошибка удаления кастомного поля")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Activity
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

// ---------------------------------------------------------------------------
// Converters (Ent -> DTO)
// ---------------------------------------------------------------------------

// toIssueResponse конвертирует Ent Issue в DTO.
func toIssueResponse(i *ent.Issue) *IssueResponse {
	resp := &IssueResponse{
		ID:               i.ID.String(),
		IssueNumber:      i.IssueNumber,
		IssueKey:         i.IssueKey,
		Title:            i.Title,
		IssueType:        string(i.IssueType),
		Priority:         string(i.Priority),
		Rank:             i.Rank,
		Version:          i.Version,
		EstimateMinutes:  i.EstimateMinutes,
		TimeSpentMinutes: i.TimeSpentMinutes,
		CustomFields:     i.CustomFields,
		DueDate:          i.DueDate,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}

	if i.DescriptionJSON != nil {
		resp.DescriptionJSON = &i.DescriptionJSON
	}
	if i.DescriptionText != "" {
		resp.DescriptionText = i.DescriptionText
	}

	if i.Edges.Status != nil {
		resp.Status = *toWorkflowStatusResponse(i.Edges.Status)
	}
	if i.Edges.Assignee != nil {
		resp.Assignee = toUserBriefResponse(i.Edges.Assignee)
	}
	if i.Edges.Reporter != nil {
		resp.Reporter = toUserBriefResponse(i.Edges.Reporter)
	}
	if i.Edges.Parent != nil {
		resp.Parent = toIssueBriefResponse(i.Edges.Parent)
	}
	if i.Edges.Children != nil {
		resp.ChildrenCount = len(i.Edges.Children)
	}
	if i.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(i.Edges.Labels))
		for _, l := range i.Edges.Labels {
			labels = append(labels, *toLabelResponse(l))
		}
		resp.Labels = labels
	}

	return resp
}

// toIssueBriefResponse конвертирует Ent Issue в краткий DTO.
func toIssueBriefResponse(i *ent.Issue) *IssueBriefResponse {
	status := ""
	if i.Edges.Status != nil {
		status = i.Edges.Status.Name
	}
	return &IssueBriefResponse{
		ID:        i.ID.String(),
		IssueKey:  i.IssueKey,
		Title:     i.Title,
		IssueType: string(i.IssueType),
		Priority:  string(i.Priority),
		Status:    status,
	}
}

// toUserBriefResponse конвертирует Ent User в краткий DTO.
func toUserBriefResponse(u *ent.User) *UserBriefResponse {
	return &UserBriefResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
	}
}

// toWorkflowStatusResponse конвертирует Ent WorkflowStatus в DTO.
func toWorkflowStatusResponse(st *ent.WorkflowStatus) *WorkflowStatusResponse {
	return &WorkflowStatusResponse{
		ID:        st.ID.String(),
		Name:      st.Name,
		Category:  string(st.Category),
		Color:     st.Color,
		SortOrder: st.SortOrder,
		IsDefault: st.IsDefault,
	}
}

// toCommentResponse конвертирует Ent Comment в DTO.
func toCommentResponse(c *ent.Comment) *CommentResponse {
	resp := &CommentResponse{
		ID:         c.ID.String(),
		BodyJSON:   c.BodyJSON,
		BodyText:   c.BodyText,
		IsInternal: c.IsInternal,
		Mentions:   c.Mentions,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
	if c.Edges.Author != nil {
		resp.Author = *toUserBriefResponse(c.Edges.Author)
	}
	if c.Edges.Reactions != nil {
		resp.Reactions = aggregateReactions(c.Edges.Reactions)
	}
	return resp
}

// aggregateReactions группирует реакции по emoji.
func aggregateReactions(reactions []*ent.Reaction) []ReactionResponse {
	grouped := make(map[string]*ReactionResponse)
	for _, r := range reactions {
		if _, ok := grouped[r.Emoji]; !ok {
			grouped[r.Emoji] = &ReactionResponse{
				Emoji: r.Emoji,
				Users: make([]UserBriefResponse, 0),
			}
		}
		entry := grouped[r.Emoji]
		entry.Count++
		if r.Edges.User != nil {
			entry.Users = append(entry.Users, *toUserBriefResponse(r.Edges.User))
		}
	}

	result := make([]ReactionResponse, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, *v)
	}
	return result
}

// toActivityLogResponse конвертирует Ent ActivityLog в DTO.
func toActivityLogResponse(a *ent.ActivityLog) *ActivityLogResponse {
	resp := &ActivityLogResponse{
		ID:        a.ID.String(),
		Action:    a.Action,
		CreatedAt: a.CreatedAt,
	}

	if a.Changes != nil {
		changes := make([]FieldChange, len(a.Changes))
		for i, c := range a.Changes {
			changes[i] = FieldChange{
				Field:    c.Field,
				OldValue: c.OldValue,
				NewValue: c.NewValue,
			}
		}
		resp.Changes = changes
	}

	if a.Edges.User != nil {
		resp.User = *toUserBriefResponse(a.Edges.User)
	}
	return resp
}

// toTimeLogResponse конвертирует Ent TimeLog в DTO.
func toTimeLogResponse(tl *ent.TimeLog) *TimeLogResponse {
	resp := &TimeLogResponse{
		ID:          tl.ID.String(),
		Minutes:     tl.Minutes,
		Description: tl.Description,
		StartedAt:   tl.StartedAt,
		EndedAt:     tl.EndedAt,
		CreatedAt:   tl.CreatedAt,
	}
	if tl.Edges.User != nil {
		resp.User = *toUserBriefResponse(tl.Edges.User)
	}
	return resp
}

// toLabelResponse конвертирует Ent Label в DTO.
func toLabelResponse(l *ent.Label) *LabelResponse {
	resp := &LabelResponse{
		ID:    l.ID.String(),
		Name:  l.Name,
		Color: l.Color,
	}
	if l.Edges.Group != nil {
		resp.GroupName = l.Edges.Group.Name
	}
	return resp
}

// toLabelGroupResponse конвертирует Ent LabelGroup в DTO.
func toLabelGroupResponse(g *ent.LabelGroup) *LabelGroupResponse {
	resp := &LabelGroupResponse{
		ID:   g.ID.String(),
		Name: g.Name,
	}
	if g.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(g.Edges.Labels))
		for _, l := range g.Edges.Labels {
			labels = append(labels, *toLabelResponse(l))
		}
		resp.Labels = labels
	}
	return resp
}

// toBoardResponse конвертирует Ent Board в DTO.
func toBoardResponse(b *ent.Board) *BoardResponse {
	return &BoardResponse{
		ID:        b.ID.String(),
		Name:      b.Name,
		Filters:   b.Filters,
		IsDefault: b.IsDefault,
		CreatedAt: b.CreatedAt,
	}
}

// toBoardDataResponse конвертирует Ent Board с eager-loaded данными в DTO.
func toBoardDataResponse(b *ent.Board) *BoardDataResponse {
	resp := &BoardDataResponse{
		Board: *toBoardResponse(b),
	}

	if b.Edges.Project != nil && b.Edges.Project.Edges.WorkflowStatuses != nil {
		columns := make([]BoardColumnResponse, 0, len(b.Edges.Project.Edges.WorkflowStatuses))
		for _, st := range b.Edges.Project.Edges.WorkflowStatuses {
			col := BoardColumnResponse{
				Status: *toWorkflowStatusResponse(st),
			}
			if st.Edges.Issues != nil {
				issues := make([]IssueBriefResponse, 0, len(st.Edges.Issues))
				for _, i := range st.Edges.Issues {
					issues = append(issues, *toIssueBriefResponse(i))
				}
				col.Issues = issues
			}
			columns = append(columns, col)
		}
		resp.Columns = columns
	}

	return resp
}

// toCustomFieldResponse конвертирует Ent CustomFieldDefinition в DTO.
func toCustomFieldResponse(cf *ent.CustomFieldDefinition) *CustomFieldDefinitionResponse {
	return &CustomFieldDefinitionResponse{
		ID:        cf.ID.String(),
		Name:      cf.Name,
		FieldType: string(cf.FieldType),
		Options:   cf.Options,
		Required:  cf.Required,
		SortOrder: cf.SortOrder,
	}
}

// toIssueLinkResponse конвертирует Ent IssueLink в DTO с информацией о связанной задаче.
func toIssueLinkResponse(l *ent.IssueLink, related *ent.Issue) *IssueLinkResponse {
	resp := &IssueLinkResponse{
		ID:         l.ID.String(),
		LinkType:   string(l.LinkType),
		CustomType: l.CustomType,
	}
	if related != nil {
		resp.Issue = *toIssueBriefResponse(related)
	}
	return resp
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildFieldChanges сравнивает существующую задачу с запросом обновления и возвращает список изменений.
func buildFieldChanges(existing *ent.Issue, req UpdateIssueRequest) []FieldChange {
	var changes []FieldChange

	if req.Title != nil && *req.Title != existing.Title {
		changes = append(changes, FieldChange{Field: "title", OldValue: existing.Title, NewValue: *req.Title})
	}
	if req.Priority != nil && *req.Priority != string(existing.Priority) {
		changes = append(changes, FieldChange{Field: "priority", OldValue: string(existing.Priority), NewValue: *req.Priority})
	}
	if req.StatusID != nil {
		oldStatus := ""
		if existing.Edges.Status != nil {
			oldStatus = existing.Edges.Status.ID.String()
		}
		if *req.StatusID != oldStatus {
			changes = append(changes, FieldChange{Field: "status", OldValue: oldStatus, NewValue: *req.StatusID})
		}
	}
	if req.AssigneeID != nil {
		oldAssignee := ""
		if existing.Edges.Assignee != nil {
			oldAssignee = existing.Edges.Assignee.ID.String()
		}
		if *req.AssigneeID != oldAssignee {
			changes = append(changes, FieldChange{Field: "assignee", OldValue: oldAssignee, NewValue: *req.AssigneeID})
		}
	}

	return changes
}

// extractPlainText извлекает текст из Plate editor JSON для поискового индекса.
func extractPlainText(descJSON map[string]interface{}) string {
	// Упрощённое извлечение: обходим children и собираем text
	children, ok := descJSON["children"].([]interface{})
	if !ok {
		return ""
	}
	var result string
	for _, child := range children {
		node, ok := child.(map[string]interface{})
		if !ok {
			continue
		}
		if text, ok := node["text"].(string); ok {
			result += text
		}
		// Рекурсивно обходим вложенные children
		if innerChildren, ok := node["children"].([]interface{}); ok {
			for _, inner := range innerChildren {
				if innerNode, ok := inner.(map[string]interface{}); ok {
					if text, ok := innerNode["text"].(string); ok {
						result += text
					}
				}
			}
		}
	}
	return result
}
