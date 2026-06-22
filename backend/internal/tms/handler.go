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
//
// ВАЖНО (Fiber v3 порядок выполнения): сигнатура методов роутера —
// Post(path string, handler Handler, middleware ...Handler). Variadic-middleware
// выполняются ПЕРЕД основным handler (router.register строит цепочку как
// [middleware..., handler]). Поэтому RBAC-гард передаётся ПОСЛЕ основного
// обработчика: `route.Post(path, h.Handler, rbac.RequirePermission(...))`.
// Если поменять местами (гард первым аргументом), гард станет финальным
// обработчиком и не выполнится, пока handler сам не вызовет c.Next() — то есть
// проверка прав будет молча пропущена (FA-02-02: IDOR / OWASP A01).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	tms := api.Group("/projects/:projectId/tms", authMw)

	// Тест-кейсы
	cases := tms.Group("/cases")
	cases.Post("/", h.CreateTestCase, rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc))
	cases.Get("/", h.ListTestCases, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	cases.Get("/:caseKey", h.GetTestCase, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	cases.Put("/:caseKey", h.UpdateTestCase, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
	cases.Delete("/:caseKey", h.DeleteTestCase, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))

	// Версии тест-кейсов
	cases.Get("/:caseKey/versions", h.ListTestCaseVersions, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	cases.Post("/:caseKey/rollback/:versionId", h.RollbackTestCase, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))

	// Привязка задач
	cases.Post("/:caseKey/links", h.LinkTestCaseToIssue, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
	cases.Delete("/:caseKey/links/:issueId", h.UnlinkTestCaseFromIssue, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
	cases.Get("/:caseKey/links", h.GetTestCaseLinks, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Папки (suites)
	suites := tms.Group("/suites")
	suites.Post("/", h.CreateSuite, rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc))
	suites.Get("/", h.ListSuites, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	suites.Put("/:suiteId", h.UpdateSuite, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
	suites.Delete("/:suiteId", h.DeleteSuite, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))

	// Компоненты
	components := tms.Group("/components")
	components.Post("/", h.CreateComponent, rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc))
	components.Get("/", h.ListComponents, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	components.Delete("/:componentId", h.DeleteComponent, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))

	// Прогоны
	runs := tms.Group("/runs")
	runs.Post("/", h.CreateTestRun, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))
	runs.Get("/", h.ListTestRuns, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	runs.Get("/:runId", h.GetTestRun, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	runs.Put("/:runId/status", h.UpdateRunStatus, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))
	runs.Post("/:runId/reopen", h.ReopenTestRun, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))

	// Управление прогоном
	runs.Post("/:runId/cases", h.AddTestCasesToRun, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))
	runs.Post("/:runId/assignees", h.AssignTestersToRun, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))
	runs.Post("/:runId/start", h.StartTestRun, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))
	runs.Post("/:runId/complete", h.CompleteTestRun, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))
	runs.Get("/:runId/detail", h.GetTestRunDetail, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Результаты
	runs.Post("/:runId/results", h.CreateTestResult, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))
	runs.Put("/:runId/results/:resultId", h.ExecuteTestResult, rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc))
	runs.Get("/:runId/results", h.ListTestResults, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Импорт Allure в существующий прогон
	runs.Post("/:runId/allure", h.ImportAllureToExistingRun, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))

	// Импорт Allure (новый прогон)
	allure := tms.Group("/allure")
	allure.Post("/import", h.ImportAllureResults, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))

	// Баг из результата прогона
	runs.Post("/:runId/results/:resultId/bug", h.CreateBugFromResult, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))

	// Отчёты
	reports := tms.Group("/reports")
	reports.Get("/history", h.GetTestRunHistory, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/flaky", h.GetFlakyTests, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/coverage", h.GetCoverageReport, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/trends", h.GetTestTrends, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/comparison", h.GetRunComparison, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/traceability", h.GetTraceabilityMatrix, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/defect-categories", h.GetDefectCategories, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	reports.Get("/env-comparison", h.CompareEnvironments, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Импорт/экспорт
	tms.Post("/import", h.ImportTestCases, rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc))
	tms.Get("/export", h.ExportTestCases, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	runs.Get("/:runId/export", h.ExportTestRunReport, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// CI/CD Webhooks — используют тот же authMw (поддержка X-API-Key / JWT).
	// Упрощённый интерфейс для пайплайнов: POST файл или JSON → автоматический прогон.
	// FA-02-02: вебхуки создают прогоны в проекте из пути, поэтому требуют
	// тех же прав, что и /tms/allure/import (PermTestRunCreate) — иначе любой
	// аутентифицированный пользователь мог бы импортировать результаты в чужой
	// проект, зная его UUID (IDOR / OWASP A01).
	webhook := tms.Group("/webhook")
	webhook.Post("/", h.WebhookAllure, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))
	webhook.Post("/junit", h.WebhookJUnit, rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))

	// Тестовые окружения
	envs := api.Group("/projects/:projectId/environments", authMw)
	envs.Post("/", h.CreateEnvironmentHandler, rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc))
	envs.Get("/", h.ListEnvironmentsHandler, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	envs.Put("/:envId", h.UpdateEnvironmentHandler, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
	envs.Delete("/:envId", h.DeleteEnvironmentHandler, rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc))
}

// ---------------------------------------------------------------------------
// Вспомогательные функции TMS handler-слоя
// ---------------------------------------------------------------------------

// parseCaseKey извлекает caseKey из параметров пути.
func parseCaseKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("caseKey"))
}
