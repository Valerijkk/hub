package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/search"
)

// ---------------------------------------------------------------------------
// Issue handlers
// ---------------------------------------------------------------------------

// CreateIssue godoc
// @Summary Создать задачу
// @Description Создаёт новую задачу (issue) в указанном проекте.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateIssueRequest true "Данные задачи"
// @Success 201 {object} IssueResponse "Задача создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues [post]
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateIssue(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssues godoc
// @Summary Список задач проекта
// @Description Возвращает список задач проекта с фильтрацией по статусу, приоритету, исполнителю и т.д.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param status query string false "Фильтр по статусу"
// @Param priority query string false "Фильтр по приоритету"
// @Param assignee_id query string false "Фильтр по исполнителю"
// @Success 200 {object} ListIssuesResponse "Список задач"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues [get]
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

// GetIssue godoc
// @Summary Получить задачу
// @Description Возвращает детальную информацию о задаче по её ключу (например, PROJ-123).
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи (например PROJ-123)"
// @Success 200 {object} IssueResponse "Данные задачи"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 404 {object} fiber.Error "Задача не найдена"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey} [get]
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

// UpdateIssue godoc
// @Summary Обновить задачу
// @Description Обновляет поля задачи: название, описание, статус, приоритет, исполнителя и др.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body UpdateIssueRequest true "Обновляемые поля"
// @Success 200 {object} IssueResponse "Задача обновлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey} [put]
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateIssue(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssue godoc
// @Summary Удалить задачу
// @Description Мягкое удаление задачи (перемещение в корзину).
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 204 "Задача удалена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey} [delete]
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

// MoveIssue godoc
// @Summary Переместить задачу
// @Description Изменяет позицию задачи на доске (статус, порядок сортировки).
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body MoveIssueRequest true "Новая позиция"
// @Success 200 {object} IssueResponse "Задача перемещена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/position [put]
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
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

// SearchIssuesJQL godoc
// @Summary JQL-поиск задач
// @Description Парсит JQL-строку, конвертирует в предикаты и возвращает отфильтрованные задачи.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body jqlRequest true "JQL-запрос"
// @Success 200 {object} map[string]interface{} "Результаты: query, total, results"
// @Failure 400 {object} fiber.Error "Ошибка JQL или неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Ошибка выполнения"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/jql [post]
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
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
