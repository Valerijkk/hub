package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики рабочего процесса
// ---------------------------------------------------------------------------

// ListWorkflowStatuses — Список статусов workflow
// @Summary Список статусов workflow
// @Description Возвращает все статусы рабочего процесса проекта.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} WorkflowStatusResponse "Список статусов"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/statuses [get]
func (h *Handler) ListWorkflowStatuses(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListWorkflowStatuses(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateWorkflowStatus — Создать статус workflow
// @Summary Создать статус workflow
// @Description Создаёт новый статус в рабочем процессе проекта.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateWorkflowStatusRequest true "Данные статуса"
// @Success 201 {object} WorkflowStatusResponse "Статус создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/statuses [post]
func (h *Handler) CreateWorkflowStatus(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateWorkflowStatus(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateWorkflowStatus — Обновить статус workflow
// @Summary Обновить статус workflow
// @Description Обновляет название, цвет или категорию статуса.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param statusId path string true "ID статуса"
// @Param body body CreateWorkflowStatusRequest true "Обновляемые данные"
// @Success 200 {object} WorkflowStatusResponse "Статус обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/statuses/{statusId} [put]
func (h *Handler) UpdateWorkflowStatus(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	statusID, err := httputil.ParseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateWorkflowStatus(c.Context(), projectID, statusID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteWorkflowStatus — Удалить статус workflow
// @Summary Удалить статус workflow
// @Description Удаляет статус из рабочего процесса проекта.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param statusId path string true "ID статуса"
// @Success 204 "Статус удалён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/statuses/{statusId} [delete]
func (h *Handler) DeleteWorkflowStatus(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	statusID, err := httputil.ParseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowStatus(c.Context(), projectID, statusID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListWorkflowTransitions — Список переходов workflow
// @Summary Список переходов workflow
// @Description Возвращает все допустимые переходы между статусами.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} WorkflowTransitionResponse "Список переходов"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/transitions [get]
func (h *Handler) ListWorkflowTransitions(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListWorkflowTransitions(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateWorkflowTransition — Создать переход workflow
// @Summary Создать переход workflow
// @Description Создаёт новый допустимый переход между статусами.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateWorkflowTransitionRequest true "Данные перехода"
// @Success 201 {object} WorkflowTransitionResponse "Переход создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/transitions [post]
func (h *Handler) CreateWorkflowTransition(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateWorkflowTransitionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateWorkflowTransition(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteWorkflowTransition — Удалить переход workflow
// @Summary Удалить переход workflow
// @Description Удаляет переход между статусами.
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param transitionId path string true "ID перехода"
// @Success 204 "Переход удалён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/workflow/transitions/{transitionId} [delete]
func (h *Handler) DeleteWorkflowTransition(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	transitionID, err := httputil.ParseUUIDParam(c, "transitionId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowTransition(c.Context(), projectID, transitionID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
