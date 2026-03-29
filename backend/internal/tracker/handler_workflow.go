package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Workflow handlers
// ---------------------------------------------------------------------------

// ListWorkflowStatuses godoc
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

// CreateWorkflowStatus godoc
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

// UpdateWorkflowStatus godoc
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
	statusID, err := httputil.ParseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateWorkflowStatus(c.Context(), statusID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteWorkflowStatus godoc
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
	statusID, err := httputil.ParseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowStatus(c.Context(), statusID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListWorkflowTransitions godoc
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

// CreateWorkflowTransition godoc
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
	var req CreateWorkflowTransitionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateWorkflowTransition(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteWorkflowTransition godoc
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
	transitionID, err := httputil.ParseUUIDParam(c, "transitionId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowTransition(c.Context(), transitionID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
