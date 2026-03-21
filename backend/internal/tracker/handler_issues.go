package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/search"
)

// ---------------------------------------------------------------------------
// Issue handlers
// ---------------------------------------------------------------------------

// CreateIssue обрабатывает POST /api/projects/:projectId/issues
func (h *Handler) CreateIssue(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateIssue(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssues обрабатывает GET /api/projects/:projectId/issues
func (h *Handler) ListIssues(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req ListIssuesRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректные параметры запроса")
	}

	resp, err := h.service.ListIssues(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetIssue обрабатывает GET /api/projects/:projectId/issues/:issueKey
func (h *Handler) GetIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetIssue(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateIssue обрабатывает PUT /api/projects/:projectId/issues/:issueKey
func (h *Handler) UpdateIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateIssue(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssue обрабатывает DELETE /api/projects/:projectId/issues/:issueKey
func (h *Handler) DeleteIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssue(c.Context(), issueKey, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MoveIssue обрабатывает PUT /api/projects/:projectId/issues/:issueKey/position
func (h *Handler) MoveIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req MoveIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.MoveIssue(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// JQL-поиск
// ---------------------------------------------------------------------------

// jqlRequest — тело запроса для JQL-поиска задач.
type jqlRequest struct {
	Query string `json:"query" validate:"required"`
}

// SearchIssuesJQL обрабатывает POST /api/projects/:projectId/issues/jql.
// Парсит JQL-строку, конвертирует в Ent-предикаты и возвращает отфильтрованные задачи.
func (h *Handler) SearchIssuesJQL(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req jqlRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}
	if req.Query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query обязателен")
	}

	// Парсинг JQL-запроса.
	parsed, err := search.ParseJQL(req.Query)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ошибка JQL: "+err.Error())
	}

	// Конвертация в Ent-предикаты.
	predicates := search.ToEntPredicates(parsed, userID)

	// Базовый запрос по проекту.
	query := h.service.repo.GetDB().Issue.Query().
		Where(issue.HasProjectWith(project.ID(projectID))).
		Where(issue.DeletedAtIsNil())

	// Применяем JQL-предикаты.
	for _, pred := range predicates {
		query = query.Where(pred)
	}

	// Сортировка из JQL.
	if parsed.OrderBy != nil {
		switch parsed.OrderBy.Field {
		case "created":
			query = query.Order(issue.ByCreatedAt())
		case "priority":
			query = query.Order(issue.ByPriority())
		case "updated":
			query = query.Order(issue.ByUpdatedAt())
		}
	}

	// Выполнение запроса с eager-loading связей.
	issues, err := query.
		WithStatus().
		WithAssignee().
		Limit(100).
		All(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка выполнения JQL: "+err.Error())
	}

	// Конвертация в ответ — переиспользуем формат ListIssues.
	_ = projectID // projectID используется для контекста RBAC (проверяется middleware)
	type jqlIssueItem struct {
		Key       string `json:"key"`
		Title     string `json:"title"`
		Type      string `json:"type"`
		Priority  string `json:"priority"`
		Status    string `json:"status,omitempty"`
		Assignee  string `json:"assignee,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]jqlIssueItem, 0, len(issues))
	for _, iss := range issues {
		status := ""
		if iss.Edges.Status != nil {
			status = iss.Edges.Status.Name
		}
		assignee := ""
		if iss.Edges.Assignee != nil {
			if iss.Edges.Assignee.Name != "" {
				assignee = iss.Edges.Assignee.Name
			} else {
				assignee = iss.Edges.Assignee.Email
			}
		}
		result = append(result, jqlIssueItem{
			Key:       iss.IssueKey,
			Title:     iss.Title,
			Type:      string(iss.IssueType),
			Priority:  string(iss.Priority),
			Status:    status,
			Assignee:  assignee,
			CreatedAt: iss.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(fiber.Map{
		"query":   req.Query,
		"total":   len(result),
		"results": result,
	})
}
