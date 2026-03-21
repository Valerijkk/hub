package tms

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Suite handlers
// ---------------------------------------------------------------------------

// CreateSuite handles POST /api/projects/:projectId/tms/suites
// @Summary Create test case folder
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body CreateSuiteRequest true "Suite data"
// @Success 201 {object} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites [post]
func (h *Handler) CreateSuite(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateSuiteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateSuite(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListSuites handles GET /api/projects/:projectId/tms/suites
// @Summary Suite tree
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites [get]
func (h *Handler) ListSuites(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListSuites(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateSuite handles PUT /api/projects/:projectId/tms/suites/:suiteId
// @Summary Update suite
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param suiteId path string true "Suite ID"
// @Param body body UpdateSuiteRequest true "Update data"
// @Success 200 {object} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites/{suiteId} [put]
func (h *Handler) UpdateSuite(c fiber.Ctx) error {
	suiteID, err := httputil.ParseUUIDParam(c, "suiteId")
	if err != nil {
		return err
	}

	var req UpdateSuiteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateSuite(c.Context(), suiteID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteSuite handles DELETE /api/projects/:projectId/tms/suites/:suiteId
// @Summary Delete suite
// @Tags TMS
// @Param projectId path string true "Project ID"
// @Param suiteId path string true "Suite ID"
// @Success 204
// @Router /api/projects/{projectId}/tms/suites/{suiteId} [delete]
func (h *Handler) DeleteSuite(c fiber.Ctx) error {
	suiteID, err := httputil.ParseUUIDParam(c, "suiteId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteSuite(c.Context(), suiteID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Component handlers
// ---------------------------------------------------------------------------

// CreateComponent handles POST /api/projects/:projectId/tms/components
// @Summary Create component
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body CreateComponentRequest true "Component data"
// @Success 201 {object} ComponentResponse
// @Router /api/projects/{projectId}/tms/components [post]
func (h *Handler) CreateComponent(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateComponentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateComponent(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListComponents handles GET /api/projects/:projectId/tms/components
// @Summary List components
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} ComponentResponse
// @Router /api/projects/{projectId}/tms/components [get]
func (h *Handler) ListComponents(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListComponents(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteComponent handles DELETE /api/projects/:projectId/tms/components/:componentId
// @Summary Delete component
// @Tags TMS
// @Param projectId path string true "Project ID"
// @Param componentId path string true "Component ID"
// @Success 204
// @Router /api/projects/{projectId}/tms/components/{componentId} [delete]
func (h *Handler) DeleteComponent(c fiber.Ctx) error {
	componentID, err := httputil.ParseUUIDParam(c, "componentId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteComponent(c.Context(), componentID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
