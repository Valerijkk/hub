package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Custom field handlers
// ---------------------------------------------------------------------------

// ListCustomFields godoc
// @Summary Список кастомных полей
// @Description Возвращает все пользовательские поля проекта.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} CustomFieldResponse "Кастомные поля"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/custom-fields [get]
func (h *Handler) ListCustomFields(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListCustomFields(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateCustomField godoc
// @Summary Создать кастомное поле
// @Description Создаёт новое пользовательское поле для задач проекта.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateCustomFieldRequest true "Данные поля"
// @Success 201 {object} CustomFieldResponse "Поле создано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/custom-fields [post]
func (h *Handler) CreateCustomField(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateCustomFieldRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateCustomField(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteCustomField godoc
// @Summary Удалить кастомное поле
// @Description Удаляет пользовательское поле и все его значения из задач.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param fieldId path string true "ID поля"
// @Success 204 "Поле удалено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/custom-fields/{fieldId} [delete]
func (h *Handler) DeleteCustomField(c fiber.Ctx) error {
	fieldID, err := httputil.ParseUUIDParam(c, "fieldId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteCustomField(c.Context(), fieldID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
