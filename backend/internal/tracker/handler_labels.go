package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Label handlers
// ---------------------------------------------------------------------------

// ListLabels обрабатывает GET /api/projects/:projectId/labels
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

// CreateLabel обрабатывает POST /api/projects/:projectId/labels
func (h *Handler) CreateLabel(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateLabel(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteLabel обрабатывает DELETE /api/projects/:projectId/labels/:labelId
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

// ListLabelGroups обрабатывает GET /api/projects/:projectId/labels/groups
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

// CreateLabelGroup обрабатывает POST /api/projects/:projectId/labels/groups
func (h *Handler) CreateLabelGroup(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelGroupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateLabelGroup(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
