package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля трекера задач.
type Handler struct {
	service    *TrackerService
	automation *AutomationService
}

// NewHandler создаёт новый обработчик трекера.
func NewHandler(service *TrackerService, automation *AutomationService) *Handler {
	return &Handler{service: service, automation: automation}
}

// RegisterRoutes регистрирует все маршруты трекера под /api/projects/:projectId.
// authMw — middleware аутентификации, rbacSvc используется для проверки прав доступа.
//
// ВАЖНО (Fiber v3): Post(path, handler, ...middleware) — variadic middleware выполняются
// ПЕРЕД handler. RBAC передаётся после handler: route.Post(path, h.Handler, rbac.RequirePermission(...)).
// См. tms/handler.go и .planning/audits/2026-06/FA-RBAC-FIBER3-AUDIT.md (FA-02-02).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	tracker := api.Group("/projects/:projectId", authMw)

	// Задачи
	issues := tracker.Group("/issues")
	issues.Post("/", h.CreateIssue, rbac.RequirePermission(rbac.PermTaskCreate, rbacSvc))
	issues.Get("/", h.ListIssues, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	issues.Get("/:issueKey", h.GetIssue, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	issues.Put("/:issueKey", h.UpdateIssue, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Delete("/:issueKey", h.DeleteIssue, rbac.RequirePermission(rbac.PermTaskDelete, rbacSvc))
	issues.Put("/:issueKey/position", h.MoveIssue, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))

	// JQL-поиск задач
	issues.Post("/jql", h.SearchIssuesJQL, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Связи задач
	issues.Post("/:issueKey/links", h.CreateIssueLink, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Get("/:issueKey/links", h.ListIssueLinks, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	issues.Delete("/:issueKey/links/:linkId", h.DeleteIssueLink, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	// A13-004: batch links для Gantt (до /issues/:issueKey чтобы не конфликтовать)
	tracker.Get("/issue-links", h.ListProjectIssueLinks, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Связи задача ↔ тред мессенджера (PROD-03)
	issues.Get("/:issueKey/message-links", h.ListIssueMessageLinks, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	issues.Post("/:issueKey/message-links", h.CreateIssueMessageLink, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Delete("/:issueKey/message-links/:linkId", h.DeleteIssueMessageLink, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	tracker.Get("/message-links", h.ListThreadIssueLinks, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Комментарии
	issues.Post("/:issueKey/comments", h.CreateComment, rbac.RequirePermission(rbac.PermTaskComment, rbacSvc))
	issues.Get("/:issueKey/comments", h.ListComments, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	issues.Put("/:issueKey/comments/:commentId", h.UpdateComment, rbac.RequirePermission(rbac.PermTaskComment, rbacSvc))
	issues.Delete("/:issueKey/comments/:commentId", h.DeleteComment, rbac.RequirePermission(rbac.PermTaskComment, rbacSvc))
	issues.Post("/:issueKey/comments/:commentId/reactions", h.ToggleReaction, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Активность
	issues.Get("/:issueKey/activity", h.GetActivity, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Учёт времени
	issues.Post("/:issueKey/time", h.LogTime, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Post("/:issueKey/timer/start", h.StartTimer, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Post("/:issueKey/timer/stop", h.StopTimer, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	issues.Get("/:issueKey/time-logs", h.ListTimeLogs, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Доски
	boards := tracker.Group("/boards")
	boards.Get("/", h.ListBoards, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	// W1-A1-006 / P2-BE-12: создание доски — мутация; viewer с одним PermProjectView → 403.
	boards.Post("/", h.CreateBoard, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	boards.Get("/:boardId", h.GetBoardData, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Рабочий процесс
	workflow := tracker.Group("/workflow")
	workflow.Get("/statuses", h.ListWorkflowStatuses, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	workflow.Post("/statuses", h.CreateWorkflowStatus, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	workflow.Put("/statuses/:statusId", h.UpdateWorkflowStatus, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	workflow.Delete("/statuses/:statusId", h.DeleteWorkflowStatus, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	workflow.Get("/transitions", h.ListWorkflowTransitions, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	workflow.Post("/transitions", h.CreateWorkflowTransition, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	workflow.Delete("/transitions/:transitionId", h.DeleteWorkflowTransition, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))

	// Метки
	labels := tracker.Group("/labels")
	labels.Get("/", h.ListLabels, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	labels.Post("/", h.CreateLabel, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	labels.Delete("/:labelId", h.DeleteLabel, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	labels.Get("/groups", h.ListLabelGroups, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	labels.Post("/groups", h.CreateLabelGroup, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))

	// Кастомные поля
	customFields := tracker.Group("/custom-fields")
	customFields.Get("/", h.ListCustomFields, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	customFields.Post("/", h.CreateCustomField, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	customFields.Delete("/:fieldId", h.DeleteCustomField, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))

	// Автоматизация (PROD-01)
	automation := tracker.Group("/automation-rules")
	automation.Get("/", h.ListAutomationRules, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	automation.Post("/", h.CreateAutomationRule, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	automation.Patch("/:ruleId", h.UpdateAutomationRule, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	automation.Delete("/:ruleId", h.DeleteAutomationRule, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

// parseIssueKey извлекает issueKey из параметров пути.
func parseIssueKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("issueKey"))
}
