package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Workflow handlers
// ---------------------------------------------------------------------------

// ListWorkflowStatuses обрабатывает GET /api/projects/:projectId/workflow/statuses
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

// CreateWorkflowStatus обрабатывает POST /api/projects/:projectId/workflow/statuses
func (h *Handler) CreateWorkflowStatus(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateWorkflowStatus(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateWorkflowStatus обрабатывает PUT /api/projects/:projectId/workflow/statuses/:statusId
func (h *Handler) UpdateWorkflowStatus(c fiber.Ctx) error {
	statusID, err := httputil.ParseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateWorkflowStatus(c.Context(), statusID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteWorkflowStatus обрабатывает DELETE /api/projects/:projectId/workflow/statuses/:statusId
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

// ListWorkflowTransitions обрабатывает GET /api/projects/:projectId/workflow/transitions
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

// CreateWorkflowTransition обрабатывает POST /api/projects/:projectId/workflow/transitions
func (h *Handler) CreateWorkflowTransition(c fiber.Ctx) error {
	var req CreateWorkflowTransitionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateWorkflowTransition(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteWorkflowTransition обрабатывает DELETE /api/projects/:projectId/workflow/transitions/:transitionId
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
