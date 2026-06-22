package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики меток
// ---------------------------------------------------------------------------

// ListLabels — Список меток проекта
// @Summary Список меток проекта
// @Description Возвращает все метки (labels) указанного проекта.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} LabelResponse "Список меток"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/labels [get]
func (h *Handler) ListLabels(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListLabels(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateLabel — Создать метку
// @Summary Создать метку
// @Description Создаёт новую метку для задач проекта.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateLabelRequest true "Данные метки"
// @Success 201 {object} LabelResponse "Метка создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/labels [post]
func (h *Handler) CreateLabel(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateLabel(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteLabel — Удалить метку
// @Summary Удалить метку
// @Description Удаляет метку из проекта.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param labelId path string true "ID метки"
// @Success 204 "Метка удалена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/labels/{labelId} [delete]
func (h *Handler) DeleteLabel(c fiber.Ctx) error {
	labelID, err := httputil.ParseUUIDParam(c, "labelId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteLabel(c.Context(), labelID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListLabelGroups — Список групп меток
// @Summary Список групп меток
// @Description Возвращает группы меток проекта для организации по категориям.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} LabelGroupResponse "Группы меток"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/labels/groups [get]
func (h *Handler) ListLabelGroups(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListLabelGroups(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateLabelGroup — Создать группу меток
// @Summary Создать группу меток
// @Description Создаёт новую группу для организации меток проекта.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateLabelGroupRequest true "Данные группы"
// @Success 201 {object} LabelGroupResponse "Группа создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/labels/groups [post]
func (h *Handler) CreateLabelGroup(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelGroupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateLabelGroup(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
