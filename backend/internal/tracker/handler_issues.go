package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики задач
// ---------------------------------------------------------------------------

// CreateIssue — Создать задачу
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

// ListIssues — Список задач проекта
// @Summary Список задач проекта
// @Description Курсорная пагинация (NFR-14): cursor=issue_key, limit 1..100. Offset не поддерживается.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param cursor query string false "issue_key последней строки предыдущей страницы"
// @Param limit query int false "Размер страницы (1..100, default 50)"
// @Param status query string false "Фильтр по статусу"
// @Param priority query string false "Фильтр по приоритету"
// @Param assignee query string false "Фильтр по исполнителю (uuid | none | unassigned)"
// @Param issue_type query string false "Фильтр по типу задачи"
// @Param search query string false "Полнотекстовый поиск"
// @Param order_by query string false "created_at | priority | updated_at"
// @Success 200 {object} PaginatedResponse[IssueResponse] "Список задач (items, next_cursor, has_more, total)"
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

// GetIssue — Получить задачу
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
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetIssue(c.Context(), projectID, issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateIssue — Обновить задачу
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
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
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

	resp, err := h.service.UpdateIssue(c.Context(), projectID, issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssue — Удалить задачу
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
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssue(c.Context(), projectID, issueKey, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MoveIssue — Переместить задачу
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
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
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

	resp, err := h.service.MoveIssue(c.Context(), projectID, issueKey, userID, req)
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

// SearchIssuesJQL — JQL-поиск задач
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

	// W1-A1-008 / P3-BE-07: JQL только через service (не GetDB из handler).
	resp, err := h.service.SearchIssuesJQL(c.Context(), projectID, userID, req.Query)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}
