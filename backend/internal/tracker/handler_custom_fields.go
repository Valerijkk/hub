package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Custom field handlers
// ---------------------------------------------------------------------------

// ListCustomFields обрабатывает GET /api/projects/:projectId/custom-fields
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

// CreateCustomField обрабатывает POST /api/projects/:projectId/custom-fields
func (h *Handler) CreateCustomField(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateCustomFieldRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateCustomField(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteCustomField обрабатывает DELETE /api/projects/:projectId/custom-fields/:fieldId
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
