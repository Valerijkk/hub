package tms

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// apperr используется в swag @Failure аннотациях.
var _ *apperr.AppError
// ---------------------------------------------------------------------------
// Обработчики отчётов TMS
// ---------------------------------------------------------------------------

// GetTestRunHistory обрабатывает GET /api/projects/:projectId/tms/reports/history
// @Summary Test run history
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {array} TestRunSummary
// @Router /api/projects/{projectId}/tms/reports/history [get]
func (h *Handler) GetTestRunHistory(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	resp, err := h.service.GetTestRunHistory(c.Context(), projectID, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetFlakyTests обрабатывает GET /api/projects/:projectId/tms/reports/flaky
// @Summary Flaky tests list
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} FlakyTestReport
// @Router /api/projects/{projectId}/tms/reports/flaky [get]
func (h *Handler) GetFlakyTests(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetFlakyTests(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetCoverageReport обрабатывает GET /api/projects/:projectId/tms/reports/coverage
// @Summary Issue coverage by test cases report
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} CoverageReport
// @Router /api/projects/{projectId}/tms/reports/coverage [get]
func (h *Handler) GetCoverageReport(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetCoverageReport(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTestTrends обрабатывает GET /api/projects/:projectId/tms/reports/trends
// @Summary Test run trends (pass/fail/skip per run)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param limit query int false "Number of recent runs (default 20)"
// @Success 200 {array} TrendPoint
// @Router /api/projects/{projectId}/tms/reports/trends [get]
func (h *Handler) GetTestTrends(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	resp, err := h.service.GetTestTrends(c.Context(), projectID, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetRunComparison обрабатывает GET /api/projects/:projectId/tms/reports/comparison
// @Summary Compare two test runs
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param run1 query string true "First run ID"
// @Param run2 query string true "Second run ID"
// @Success 200 {object} RunComparison
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/reports/comparison [get]
func (h *Handler) GetRunComparison(c fiber.Ctx) error {
	run1Str := c.Query("run1")
	run2Str := c.Query("run2")
	if run1Str == "" || run2Str == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметры run1 и run2 обязательны")
	}

	runID1, err := uuid.Parse(run1Str)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный run1")
	}
	runID2, err := uuid.Parse(run2Str)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный run2")
	}

	resp, err := h.service.GetRunComparison(c.Context(), runID1, runID2)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTraceabilityMatrix обрабатывает GET /api/projects/:projectId/tms/reports/traceability
// @Summary Issue-test case traceability matrix
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} TraceabilityMatrix
// @Router /api/projects/{projectId}/tms/reports/traceability [get]
func (h *Handler) GetTraceabilityMatrix(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetTraceabilityMatrix(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CompareEnvironments обрабатывает GET /api/projects/:projectId/tms/reports/env-comparison
// @Summary Compare test results between two environments
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param env1 query string true "First environment ID"
// @Param env2 query string true "Second environment ID"
// @Success 200 {object} EnvironmentComparisonResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/reports/env-comparison [get]
func (h *Handler) CompareEnvironments(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	env1Str := c.Query("env1")
	env2Str := c.Query("env2")
	if env1Str == "" || env2Str == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметры env1 и env2 обязательны")
	}

	envID1, err := uuid.Parse(env1Str)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный env1")
	}
	envID2, err := uuid.Parse(env2Str)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный env2")
	}

	resp, err := h.service.CompareEnvironments(c.Context(), projectID, envID1, envID2)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetDefectCategories обрабатывает GET /api/projects/:projectId/tms/reports/defect-categories
// @Summary Defect categories (group failed tests by error)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} DefectCategory
// @Router /api/projects/{projectId}/tms/reports/defect-categories [get]
func (h *Handler) GetDefectCategories(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetDefectCategories(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
