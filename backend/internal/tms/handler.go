package tms

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля TMS.
type Handler struct {
	service *TMSService
}

// NewHandler создаёт новый обработчик TMS.
func NewHandler(service *TMSService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты TMS под /api/projects/:projectId/tms.
// authMw — middleware аутентификации, rbacSvc используется для проверки прав доступа.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	tms := api.Group("/projects/:projectId/tms", authMw)

	// Тест-кейсы
	cases := tms.Group("/cases")
	cases.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateTestCase)
	cases.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCases)
	cases.Get("/:caseKey", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestCase)
	cases.Put("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateTestCase)
	cases.Delete("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteTestCase)

	// Версии тест-кейсов
	cases.Get("/:caseKey/versions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCaseVersions)
	cases.Post("/:caseKey/rollback/:versionId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.RollbackTestCase)

	// Привязка к задачам
	cases.Post("/:caseKey/links", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.LinkTestCaseToIssue)
	cases.Delete("/:caseKey/links/:issueId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UnlinkTestCaseFromIssue)

	// Папки/сьюты
	suites := tms.Group("/suites")
	suites.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateSuite)
	suites.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListSuites)
	suites.Put("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateSuite)
	suites.Delete("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteSuite)

	// Компоненты
	components := tms.Group("/components")
	components.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateComponent)
	components.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListComponents)
	components.Delete("/:componentId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteComponent)

	// Прогоны
	runs := tms.Group("/runs")
	runs.Post("/", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.CreateTestRun)
	runs.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestRuns)
	runs.Get("/:runId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestRun)
	runs.Put("/:runId/status", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.UpdateRunStatus)
	runs.Post("/:runId/reopen", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.ReopenTestRun)

	// Результаты
	runs.Post("/:runId/results", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.CreateTestResult)
	runs.Put("/:runId/results/:resultId", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.UpdateTestResult)
	runs.Get("/:runId/results", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestResults)
}

// ---------------------------------------------------------------------------
// Хелперы
// ---------------------------------------------------------------------------

// userIDFromCtx извлекает user ID из Fiber Locals.
func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
	}
	return id, nil
}

// parseProjectID извлекает и валидирует projectId из параметров пути.
func parseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	return id, nil
}

// parseUUIDParam парсит UUID из параметра пути.
func parseUUIDParam(c fiber.Ctx, name string) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params(name))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный "+name)
	}
	return id, nil
}

// parseCaseKey извлекает caseKey из параметров пути.
func parseCaseKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("caseKey"))
}

// ---------------------------------------------------------------------------
// Тест-кейсы
// ---------------------------------------------------------------------------

// CreateTestCase обрабатывает POST /api/projects/:projectId/tms/cases
// @Summary Создать тест-кейс
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateTestCaseRequest true "Данные тест-кейса"
// @Success 201 {object} TestCaseResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/cases [post]
func (h *Handler) CreateTestCase(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateTestCaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateTestCase(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTestCases обрабатывает GET /api/projects/:projectId/tms/cases
// @Summary Список тест-кейсов
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param offset query int false "Смещение"
// @Param limit query int false "Лимит"
// @Param suite_id query string false "ID папки"
// @Param status query string false "Статус"
// @Param priority query string false "Приоритет"
// @Param search query string false "Поиск по названию"
// @Success 200 {object} PaginatedResponse[TestCaseResponse]
// @Router /api/projects/{projectId}/tms/cases [get]
func (h *Handler) ListTestCases(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
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
// @Summary Получить тест-кейс
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} errors.AppError
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
// @Summary Обновить тест-кейс
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Param body body UpdateTestCaseRequest true "Данные обновления"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey} [put]
func (h *Handler) UpdateTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateTestCaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateTestCase(c.Context(), caseKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteTestCase обрабатывает DELETE /api/projects/:projectId/tms/cases/:caseKey
// @Summary Удалить тест-кейс
// @Tags TMS
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Success 204
// @Failure 404 {object} errors.AppError
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
// Версии
// ---------------------------------------------------------------------------

// ListTestCaseVersions обрабатывает GET /api/projects/:projectId/tms/cases/:caseKey/versions
// @Summary Список версий тест-кейса
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
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
// @Summary Откатить тест-кейс к версии
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Param versionId path string true "ID версии"
// @Success 200 {object} TestCaseResponse
// @Failure 404 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/rollback/{versionId} [post]
func (h *Handler) RollbackTestCase(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	versionID, err := parseUUIDParam(c, "versionId")
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
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
// Привязка к задачам
// ---------------------------------------------------------------------------

// LinkTestCaseToIssue обрабатывает POST /api/projects/:projectId/tms/cases/:caseKey/links
// @Summary Привязать тест-кейс к задаче
// @Tags TMS
// @Accept json
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Param body body LinkIssueRequest true "ID задачи"
// @Success 204
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/links [post]
func (h *Handler) LinkTestCaseToIssue(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}

	var req LinkIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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
// @Summary Отвязать тест-кейс от задачи
// @Tags TMS
// @Param projectId path string true "ID проекта"
// @Param caseKey path string true "Ключ тест-кейса"
// @Param issueId path string true "ID задачи"
// @Success 204
// @Router /api/projects/{projectId}/tms/cases/{caseKey}/links/{issueId} [delete]
func (h *Handler) UnlinkTestCaseFromIssue(c fiber.Ctx) error {
	caseKey := parseCaseKey(c)
	if caseKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "case key обязателен")
	}
	issueID, err := parseUUIDParam(c, "issueId")
	if err != nil {
		return err
	}

	if err := h.service.UnlinkTestCaseFromIssue(c.Context(), caseKey, issueID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Папки/сьюты
// ---------------------------------------------------------------------------

// CreateSuite обрабатывает POST /api/projects/:projectId/tms/suites
// @Summary Создать папку тест-кейсов
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateSuiteRequest true "Данные папки"
// @Success 201 {object} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites [post]
func (h *Handler) CreateSuite(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
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

// ListSuites обрабатывает GET /api/projects/:projectId/tms/suites
// @Summary Дерево папок тест-кейсов
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites [get]
func (h *Handler) ListSuites(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListSuites(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateSuite обрабатывает PUT /api/projects/:projectId/tms/suites/:suiteId
// @Summary Обновить папку
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param suiteId path string true "ID папки"
// @Param body body UpdateSuiteRequest true "Данные обновления"
// @Success 200 {object} SuiteResponse
// @Router /api/projects/{projectId}/tms/suites/{suiteId} [put]
func (h *Handler) UpdateSuite(c fiber.Ctx) error {
	suiteID, err := parseUUIDParam(c, "suiteId")
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

// DeleteSuite обрабатывает DELETE /api/projects/:projectId/tms/suites/:suiteId
// @Summary Удалить папку
// @Tags TMS
// @Param projectId path string true "ID проекта"
// @Param suiteId path string true "ID папки"
// @Success 204
// @Router /api/projects/{projectId}/tms/suites/{suiteId} [delete]
func (h *Handler) DeleteSuite(c fiber.Ctx) error {
	suiteID, err := parseUUIDParam(c, "suiteId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteSuite(c.Context(), suiteID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Компоненты
// ---------------------------------------------------------------------------

// CreateComponent обрабатывает POST /api/projects/:projectId/tms/components
// @Summary Создать компонент
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateComponentRequest true "Данные компонента"
// @Success 201 {object} ComponentResponse
// @Router /api/projects/{projectId}/tms/components [post]
func (h *Handler) CreateComponent(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
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

// ListComponents обрабатывает GET /api/projects/:projectId/tms/components
// @Summary Список компонентов
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} ComponentResponse
// @Router /api/projects/{projectId}/tms/components [get]
func (h *Handler) ListComponents(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListComponents(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteComponent обрабатывает DELETE /api/projects/:projectId/tms/components/:componentId
// @Summary Удалить компонент
// @Tags TMS
// @Param projectId path string true "ID проекта"
// @Param componentId path string true "ID компонента"
// @Success 204
// @Router /api/projects/{projectId}/tms/components/{componentId} [delete]
func (h *Handler) DeleteComponent(c fiber.Ctx) error {
	componentID, err := parseUUIDParam(c, "componentId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteComponent(c.Context(), componentID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Прогоны
// ---------------------------------------------------------------------------

// CreateTestRun обрабатывает POST /api/projects/:projectId/tms/runs
// @Summary Создать прогон тестов
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateTestRunRequest true "Данные прогона"
// @Success 201 {object} TestRunResponse
// @Router /api/projects/{projectId}/tms/runs [post]
func (h *Handler) CreateTestRun(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
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

// ListTestRuns обрабатывает GET /api/projects/:projectId/tms/runs
// @Summary Список прогонов
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param offset query int false "Смещение"
// @Param limit query int false "Лимит"
// @Param status query string false "Статус"
// @Success 200 {object} PaginatedResponse[TestRunResponse]
// @Router /api/projects/{projectId}/tms/runs [get]
func (h *Handler) ListTestRuns(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
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
// @Summary Получить прогон
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Success 200 {object} TestRunResponse
// @Failure 404 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId} [get]
func (h *Handler) GetTestRun(c fiber.Ctx) error {
	runID, err := parseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetTestRun(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateRunStatus обрабатывает PUT /api/projects/:projectId/tms/runs/:runId/status
// @Summary Обновить статус прогона
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Param body body UpdateRunStatusRequest true "Новый статус"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/status [put]
func (h *Handler) UpdateRunStatus(c fiber.Ctx) error {
	runID, err := parseUUIDParam(c, "runId")
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

// ReopenTestRun обрабатывает POST /api/projects/:projectId/tms/runs/:runId/reopen
// @Summary Переоткрыть прогон
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Success 200 {object} TestRunResponse
// @Failure 400 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/runs/{runId}/reopen [post]
func (h *Handler) ReopenTestRun(c fiber.Ctx) error {
	runID, err := parseUUIDParam(c, "runId")
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
// Результаты
// ---------------------------------------------------------------------------

// CreateTestResult обрабатывает POST /api/projects/:projectId/tms/runs/:runId/results
// @Summary Создать результат теста
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Param body body CreateTestResultRequest true "Данные результата"
// @Success 201 {object} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results [post]
func (h *Handler) CreateTestResult(c fiber.Ctx) error {
	runID, err := parseUUIDParam(c, "runId")
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

// UpdateTestResult обрабатывает PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId
// @Summary Обновить результат теста
// @Tags TMS
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Param resultId path string true "ID результата"
// @Param body body UpdateTestResultRequest true "Данные обновления"
// @Success 200 {object} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results/{resultId} [put]
func (h *Handler) UpdateTestResult(c fiber.Ctx) error {
	resultID, err := parseUUIDParam(c, "resultId")
	if err != nil {
		return err
	}

	var req UpdateTestResultRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateTestResult(c.Context(), resultID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTestResults обрабатывает GET /api/projects/:projectId/tms/runs/:runId/results
// @Summary Список результатов прогона
// @Tags TMS
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path string true "ID прогона"
// @Success 200 {array} TestResultResponse
// @Router /api/projects/{projectId}/tms/runs/{runId}/results [get]
func (h *Handler) ListTestResults(c fiber.Ctx) error {
	runID, err := parseUUIDParam(c, "runId")
	if err != nil {
		return err
	}

	resp, err := h.service.ListTestResults(c.Context(), runID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
