package tms

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Test run handlers
// ---------------------------------------------------------------------------

// CreateTestRun handles POST /api/projects/:projectId/tms/runs
// @Summary Create test run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body CreateTestRunRequest true "Run data"
// @Success 201 {object} TestRunResponse
// @Router /api/projects/{projectId}/tms/runs [post]
func (h *Handler) CreateTestRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateTestRunRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateTestRun(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTestRuns handles GET /api/projects/:projectId/tms/runs
// @Summary List test runs
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Param status query string false "Status"
// @Success 200 {object} PaginatedResponse[TestRunResponse]
// @Router /api/projects/{projectId}/tms/runs [get]
func (h *Handler) ListTestRuns(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req ListTestRunsRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректные параметры запроса")
	}

	resp, err := h.service.ListTestRuns(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTestRun handles GET /api/projects/:projectId/tms/runs/:runId
// @Summary Get test run
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 404 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId} [get]
func (h *Handler) GetTestRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetTestRun(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateRunStatus handles PUT /api/projects/:projectId/tms/runs/:runId/status
// @Summary Update run status
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body UpdateRunStatusRequest true "New status"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/status [put]
func (h *Handler) UpdateRunStatus(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req UpdateRunStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateRunStatus(c.Context(), runID, req.Status)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ReopenTestRun handles POST /api/projects/:projectId/tms/runs/:runId/reopen
// @Summary Reopen test run
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/reopen [post]
func (h *Handler) ReopenTestRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.ReopenTestRun(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Test result handlers
// ---------------------------------------------------------------------------

// CreateTestResult handles POST /api/projects/:projectId/tms/runs/:runId/results
// @Summary Create test result
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body CreateTestResultRequest true "Result data"
// @Success 201 {object} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results [post]
func (h *Handler) CreateTestResult(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req CreateTestResultRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateTestResult(c.Context(), runID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ExecuteTestResult handles PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId
// @Summary Execute test result (update status, steps, comment, files)
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param resultId path string true "Result ID"
// @Param body body UpdateTestResultRequest true "Execution data"
// @Success 200 {object} TestResultResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/results/{resultId} [put]
func (h *Handler) ExecuteTestResult(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}
	resultID, err := httputil.ParseUUIDParam(c, "resultId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateTestResultRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.ExecuteTestResult(c.Context(), runID, resultID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTestResults handles GET /api/projects/:projectId/tms/runs/:runId/results
// @Summary List run results
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {array} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results [get]
func (h *Handler) ListTestResults(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.ListTestResults(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Extended run operation handlers
// ---------------------------------------------------------------------------

// AddTestCasesToRun handles POST /api/projects/:projectId/tms/runs/:runId/cases
// @Summary Add test cases to run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body AddTestCasesToRunRequest true "List of test case IDs"
// @Success 200 {object} TestRunDetailResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/cases [post]
func (h *Handler) AddTestCasesToRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req AddTestCasesToRunRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.AddTestCasesToRun(c.Context(), runID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AssignTestersToRun handles POST /api/projects/:projectId/tms/runs/:runId/assignees
// @Summary Assign testers to run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body AssignTestersRequest true "List of user IDs"
// @Success 200 {object} TestRunDetailResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/assignees [post]
func (h *Handler) AssignTestersToRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req AssignTestersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.AssignTestersToRun(c.Context(), runID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// StartTestRun handles POST /api/projects/:projectId/tms/runs/:runId/start
// @Summary Start test run (created -> in_progress)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/start [post]
func (h *Handler) StartTestRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.StartTestRun(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CompleteTestRun handles POST /api/projects/:projectId/tms/runs/:runId/complete
// @Summary Complete test run (in_progress -> completed)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/complete [post]
func (h *Handler) CompleteTestRun(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.CompleteTestRun(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTestRunDetail handles GET /api/projects/:projectId/tms/runs/:runId/detail
// @Summary Get run details with results and assignees
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunDetailResponse
// @Failure 404 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/detail [get]
func (h *Handler) GetTestRunDetail(c fiber.Ctx) error {
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetTestRunDetail(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
