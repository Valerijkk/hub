package tms

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики прогонов тестов
// ---------------------------------------------------------------------------

// CreateTestRun обрабатывает POST /api/projects/:projectId/tms/runs
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateTestRun(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTestRuns обрабатывает GET /api/projects/:projectId/tms/runs
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

// GetTestRun обрабатывает GET /api/projects/:projectId/tms/runs/:runId
// @Summary Get test run
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId} [get]
func (h *Handler) GetTestRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetTestRun(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateRunStatus обрабатывает PUT /api/projects/:projectId/tms/runs/:runId/status
// @Summary Update run status
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body UpdateRunStatusRequest true "New status"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/status [put]
func (h *Handler) UpdateRunStatus(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req UpdateRunStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateRunStatus(c.Context(), projectID, runID, req.Status)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ReopenTestRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/reopen
// @Summary Reopen test run
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/reopen [post]
func (h *Handler) ReopenTestRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.ReopenTestRun(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Обработчики результатов тестов
// ---------------------------------------------------------------------------

// CreateTestResult обрабатывает POST /api/projects/:projectId/tms/runs/:runId/results
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
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req CreateTestResultRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateTestResult(c.Context(), projectID, runID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ExecuteTestResult обрабатывает PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId
// @Summary Execute test result (update status, steps, comment, files)
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param resultId path string true "Result ID"
// @Param body body UpdateTestResultRequest true "Execution data"
// @Success 200 {object} TestResultResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/results/{resultId} [put]
func (h *Handler) ExecuteTestResult(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
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
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.ExecuteTestResult(c.Context(), projectID, runID, resultID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTestResults обрабатывает GET /api/projects/:projectId/tms/runs/:runId/results
// @Summary List run results
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {array} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results [get]
func (h *Handler) ListTestResults(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.ListTestResults(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Обработчики расширенных операций прогона
// ---------------------------------------------------------------------------

// AddTestCasesToRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/cases
// @Summary Add test cases to run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body AddTestCasesToRunRequest true "List of test case IDs"
// @Success 200 {object} TestRunDetailResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/cases [post]
func (h *Handler) AddTestCasesToRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req AddTestCasesToRunRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.AddTestCasesToRun(c.Context(), projectID, runID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AssignTestersToRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/assignees
// @Summary Assign testers to run
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param body body AssignTestersRequest true "List of user IDs"
// @Success 200 {object} TestRunDetailResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/assignees [post]
func (h *Handler) AssignTestersToRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	var req AssignTestersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.AssignTestersToRun(c.Context(), projectID, runID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// StartTestRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/start
// @Summary Start test run (created -> in_progress)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/start [post]
func (h *Handler) StartTestRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.StartTestRun(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CompleteTestRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/complete
// @Summary Complete test run (in_progress -> completed)
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/complete [post]
func (h *Handler) CompleteTestRun(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.CompleteTestRun(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTestRunDetail обрабатывает GET /api/projects/:projectId/tms/runs/:runId/detail
// @Summary Get run details with results and assignees
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Success 200 {object} TestRunDetailResponse
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/detail [get]
func (h *Handler) GetTestRunDetail(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	runID, err := httputil.ParseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetTestRunDetail(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
