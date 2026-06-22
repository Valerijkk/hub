package tms

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики импорта Allure
// ---------------------------------------------------------------------------

// ImportAllureResults обрабатывает POST /api/projects/:projectId/tms/allure/import
// @Summary Import Allure results (create automatic run)
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body ImportAllureRequest true "Allure results"
// @Success 201 {object} TestRunDetailResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/allure/import [post]
func (h *Handler) ImportAllureResults(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ImportAllureRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.ImportAllureResults(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ImportAllureToExistingRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/allure
// @Summary Add Allure results to existing run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body ImportAllureRequest true "Allure results"
// @Success 200 {object} TestRunDetailResponse
// @Failure 400 {object} apperr.AppError
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/allure [post]
func (h *Handler) ImportAllureToExistingRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ImportAllureRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.ImportAllureToExistingRun(c.Context(), projectID, runID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Обработчики импорта/экспорта файлов
// ---------------------------------------------------------------------------

// ImportTestCases обрабатывает POST /api/projects/:projectId/tms/import
// @Summary Import test cases from file
// @Tags TMS
// @Accept multipart/form-data
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} ImportResult
// @Router /api/projects/{projectId}/tms/import [post]
func (h *Handler) ImportTestCases(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ImportTestCases(c, projectID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ExportTestCases обрабатывает GET /api/projects/:projectId/tms/export
// @Summary Export test cases to file
// @Tags TMS
// @Produce application/octet-stream
// @Param projectId path string true "Project ID"
// @Param format query string false "Format (csv, json, xml, excel)"
// @Success 200 {file} binary
// @Router /api/projects/{projectId}/tms/export [get]
func (h *Handler) ExportTestCases(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	format := c.Query("format", "csv")

	reader, filename, err := h.service.ExportTestCases(c.Context(), projectID, format)
	if err != nil {
		return err
	}

	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.SendStream(reader)
}

// ExportTestRunReport обрабатывает GET /api/projects/:projectId/tms/runs/:runId/export
// @Summary Export run report
// @Tags TMS
// @Produce application/octet-stream
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param format query string false "Format (pdf, csv, json)"
// @Success 200 {file} binary
// @Router /api/projects/{projectId}/tms/runs/{runId}/export [get]
func (h *Handler) ExportTestRunReport(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	format := c.Query("format", "csv")

	reader, filename, err := h.service.ExportTestRunReport(c.Context(), runID, format)
	if err != nil {
		return err
	}

	c.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Set("Content-Type", "application/octet-stream")
	return c.SendStream(reader)
}
