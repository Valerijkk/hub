package tms

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля TMS (управление тестированием).
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

	// Test cases
	cases := tms.Group("/cases")
	cases.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateTestCase)
	cases.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCases)
	cases.Get("/:caseKey", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestCase)
	cases.Put("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateTestCase)
	cases.Delete("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteTestCase)

	// Test case versions
	cases.Get("/:caseKey/versions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCaseVersions)
	cases.Post("/:caseKey/rollback/:versionId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.RollbackTestCase)

	// Issue linking
	cases.Post("/:caseKey/links", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.LinkTestCaseToIssue)
	cases.Delete("/:caseKey/links/:issueId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UnlinkTestCaseFromIssue)
	cases.Get("/:caseKey/links", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestCaseLinks)

	// Suites
	suites := tms.Group("/suites")
	suites.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateSuite)
	suites.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListSuites)
	suites.Put("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateSuite)
	suites.Delete("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteSuite)

	// Components
	components := tms.Group("/components")
	components.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateComponent)
	components.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListComponents)
	components.Delete("/:componentId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteComponent)

	// Runs
	runs := tms.Group("/runs")
	runs.Post("/", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.CreateTestRun)
	runs.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestRuns)
	runs.Get("/:runId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestRun)
	runs.Put("/:runId/status", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.UpdateRunStatus)
	runs.Post("/:runId/reopen", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.ReopenTestRun)

	// Run management
	runs.Post("/:runId/cases", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.AddTestCasesToRun)
	runs.Post("/:runId/assignees", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.AssignTestersToRun)
	runs.Post("/:runId/start", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.StartTestRun)
	runs.Post("/:runId/complete", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.CompleteTestRun)
	runs.Get("/:runId/detail", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestRunDetail)

	// Results
	runs.Post("/:runId/results", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.CreateTestResult)
	runs.Put("/:runId/results/:resultId", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.ExecuteTestResult)
	runs.Get("/:runId/results", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestResults)

	// Allure import to existing run
	runs.Post("/:runId/allure", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.ImportAllureToExistingRun)

	// Allure import (new run)
	allure := tms.Group("/allure")
	allure.Post("/import", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.ImportAllureResults)

	// Bug from result
	runs.Post("/:runId/results/:resultId/bug", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.CreateBugFromResult)

	// Reports
	reports := tms.Group("/reports")
	reports.Get("/history", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestRunHistory)
	reports.Get("/flaky", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetFlakyTests)
	reports.Get("/coverage", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetCoverageReport)
	reports.Get("/trends", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestTrends)
	reports.Get("/comparison", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetRunComparison)
	reports.Get("/traceability", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTraceabilityMatrix)
	reports.Get("/defect-categories", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetDefectCategories)
	reports.Get("/env-comparison", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.CompareEnvironments)

	// Import/export
	tms.Post("/import", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.ImportTestCases)
	tms.Get("/export", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ExportTestCases)
	runs.Get("/:runId/export", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ExportTestRunReport)

	// Test environments
	envs := api.Group("/projects/:projectId/environments", authMw)
	envs.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateEnvironmentHandler)
	envs.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListEnvironmentsHandler)
	envs.Put("/:envId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateEnvironmentHandler)
	envs.Delete("/:envId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteEnvironmentHandler)
}

// ---------------------------------------------------------------------------
// TMS-specific helpers
// ---------------------------------------------------------------------------

// parseCaseKey extracts caseKey from path params.
func parseCaseKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("caseKey"))
}
