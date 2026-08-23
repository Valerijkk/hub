package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики комментариев
// ---------------------------------------------------------------------------

// CreateComment — Создать комментарий
// @Summary Создать комментарий
// @Description Добавляет комментарий к задаче.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body CreateCommentRequest true "Текст комментария"
// @Success 201 {object} CommentResponse "Комментарий создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/comments [post]
func (h *Handler) CreateComment(c fiber.Ctx) error {
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

	var req CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateComment(c.Context(), projectID, issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListComments — Список комментариев задачи
// @Summary Список комментариев задачи
// @Description Возвращает все комментарии указанной задачи.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} CommentResponse "Список комментариев"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/comments [get]
func (h *Handler) ListComments(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListComments(c.Context(), projectID, issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateComment — Обновить комментарий
// @Summary Обновить комментарий
// @Description Редактирует текст существующего комментария.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param commentId path string true "ID комментария"
// @Param body body UpdateCommentRequest true "Новый текст"
// @Success 200 {object} CommentResponse "Комментарий обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/comments/{commentId} [put]
func (h *Handler) UpdateComment(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateComment(c.Context(), projectID, issueKey, userID, commentID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteComment — Удалить комментарий
// @Summary Удалить комментарий
// @Description Удаляет комментарий к задаче.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param commentId path string true "ID комментария"
// @Success 204 "Комментарий удалён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/comments/{commentId} [delete]
func (h *Handler) DeleteComment(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteComment(c.Context(), projectID, issueKey, userID, commentID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleReaction — Переключить реакцию на комментарий
// @Summary Переключить реакцию на комментарий
// @Description Добавляет или убирает эмодзи-реакцию на комментарий задачи.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param commentId path string true "ID комментария"
// @Param body body CreateReactionRequest true "Эмодзи реакции"
// @Success 200 {object} map[string]bool "added: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/comments/{commentId}/reactions [post]
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	added, err := h.service.ToggleReaction(c.Context(), projectID, issueKey, commentID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// ---------------------------------------------------------------------------
// Обработчик активности
// ---------------------------------------------------------------------------

// GetActivity — История активности задачи
// @Summary История активности задачи
// @Description Возвращает историю изменений задачи (создание, обновления полей, комментарии).
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} ActivityLogResponse "История активности"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/activity [get]
func (h *Handler) GetActivity(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetActivity(c.Context(), projectID, issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
