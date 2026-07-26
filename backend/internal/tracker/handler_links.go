package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики связей задач
// ---------------------------------------------------------------------------

// CreateIssueLink — Создать связь между задачами
// @Summary Создать связь между задачами
// @Description Создаёт связь (блокирует, связана с, дублирует) между двумя задачами.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body CreateIssueLinkRequest true "Данные связи"
// @Success 201 {object} IssueLinkResponse "Связь создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/links [post]
func (h *Handler) CreateIssueLink(c fiber.Ctx) error {
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

	var req CreateIssueLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateIssueLink(c.Context(), projectID, issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssueLinks — Список связей задачи
// @Summary Список связей задачи
// @Description Возвращает все связи указанной задачи с другими задачами.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} IssueLinkResponse "Список связей"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/links [get]
func (h *Handler) ListIssueLinks(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListIssueLinks(c.Context(), projectID, issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssueLink — Удалить связь задачи
// @Summary Удалить связь задачи
// @Description Удаляет связь между задачами по ID связи.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param linkId path string true "ID связи"
// @Success 204 "Связь удалена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/links/{linkId} [delete]
func (h *Handler) DeleteIssueLink(c fiber.Ctx) error {
	linkID, err := httputil.ParseUUIDParam(c, "linkId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssueLink(c.Context(), linkID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
