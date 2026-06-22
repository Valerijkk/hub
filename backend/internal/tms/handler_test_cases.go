package tms

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики тест-кейсов
// ---------------------------------------------------------------------------

// CreateTestCase обрабатывает POST /api/projects/:projectId/tms/cases
// @Summary Create test case
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body CreateTestCaseRequest true "Test case data"
// @Success 201 {object} TestCaseResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/cases [post]
func (h *Handler) CreateTestCase(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateTestCaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateTestCase(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTestCases обрабатывает GET /api/projects/:projectId/tms/cases
// @Summary List test cases
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Param suite_id query string false "Suite ID"
// @Param status query string false "Status"
// @Param priority query string false "Priority"
// @Param search query string false "Search by title"
// @Success 200 {object} PaginatedResponse[TestCaseResponse]
// @Router /api/projects/{projectId}/tms/cases [get]
func (h *Handler) ListTestCases(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req ListTestCasesRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректные параметры запроса")
	}

	resp, err := h.service.ListTestCases(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetTestCase обрабатывает GET /api/projects/:projectId/tms/cases/:caseKey
// @Summary Get test case
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey} [get]
func (h *Handler) GetTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	resp, err := h.service.GetTestCase(c.Context(), caseKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateTestCase обрабатывает PUT /api/projects/:projectId/tms/cases/:caseKey
// @Summary Update test case
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Param body body UpdateTestCaseRequest true "Update data"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey} [put]
func (h *Handler) UpdateTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateTestCaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateTestCase(c.Context(), caseKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteTestCase обрабатывает DELETE /api/projects/:projectId/tms/cases/:caseKey
// @Summary Delete test case
// @Tags TMS
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Success 204
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey} [delete]
func (h *Handler) DeleteTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	if err := h.service.DeleteTestCase(c.Context(), caseKey); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Обработчики версий тест-кейсов
// ---------------------------------------------------------------------------

// ListTestCaseVersions обрабатывает GET /api/projects/:projectId/tms/cases/:caseKey/versions
// @Summary List test case versions
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Success 200 {array} TestCaseVersionResponse
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/versions [get]
func (h *Handler) ListTestCaseVersions(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	resp, err := h.service.ListTestCaseVersions(c.Context(), caseKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// RollbackTestCase обрабатывает POST /api/projects/:projectId/tms/cases/:caseKey/rollback/:versionId
// @Summary Rollback test case to version
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Param versionId path string true "Version ID"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/rollback/{versionId} [post]
func (h *Handler) RollbackTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	versionID, err := httputil.ParseUUIDParam(c, "versionId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.RollbackTestCase(c.Context(), caseKey, versionID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Обработчики связей тест-кейс↔задача
// ---------------------------------------------------------------------------

// LinkTestCaseToIssue обрабатывает POST /api/projects/:projectId/tms/cases/:caseKey/links
// @Summary Link test case to issue
// @Tags TMS
// @Accept json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Param body body LinkIssueRequest true "Issue ID"
// @Success 204
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/links [post]
func (h *Handler) LinkTestCaseToIssue(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	var req LinkIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	issueID, err := uuid.Parse(req.IssueID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный issue_id")
	}

	if err := h.service.LinkTestCaseToIssue(c.Context(), caseKey, issueID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UnlinkTestCaseFromIssue обрабатывает DELETE /api/projects/:projectId/tms/cases/:caseKey/links/:issueId
// @Summary Unlink test case from issue
// @Tags TMS
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Param issueId path string true "Issue ID"
// @Success 204
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/links/{issueId} [delete]
func (h *Handler) UnlinkTestCaseFromIssue(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	issueID, err := httputil.ParseUUIDParam(c, "issueId")
	if err != nil {
		return err
	}

	if err := h.service.UnlinkTestCaseFromIssue(c.Context(), caseKey, issueID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetTestCaseLinks обрабатывает GET /api/projects/:projectId/tms/cases/:caseKey/links
// @Summary Get linked issues for test case
// @Tags TMS
// @Produce json
// @Param projectId path string true "Project ID"
// @Param caseKey path string true "Case key"
// @Success 200 {array} IssueBriefResponse
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/links [get]
func (h *Handler) GetTestCaseLinks(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	resp, err := h.service.GetTestCaseLinks(c.Context(), caseKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateBugFromResult обрабатывает POST /api/projects/:projectId/tms/runs/:runId/results/:resultId/bug
// @Summary Create bug from failed test result
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param runId path string true "Run ID"
// @Param resultId path string true "Result ID"
// @Param body body CreateBugRequest true "Bug data"
// @Success 201 {object} IssueBriefResponse
// @Failure 400 {object} apperr.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/results/{resultId}/bug [post]
func (h *Handler) CreateBugFromResult(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
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

	var req CreateBugRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateBugFromFailedResult(c.Context(), projectID, userID, resultID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
