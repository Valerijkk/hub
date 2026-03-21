package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля трекера задач.
type Handler struct {
	service *TrackerService
}

// NewHandler создаёт новый обработчик трекера.
func NewHandler(service *TrackerService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты трекера под /api/projects/:projectId.
// authMw — middleware аутентификации, rbacSvc используется для проверки прав доступа.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	tracker := api.Group("/projects/:projectId", authMw)

	// Issues
	issues := tracker.Group("/issues")
	issues.Post("/", rbac.RequirePermission(rbac.PermTaskCreate, rbacSvc), h.CreateIssue)
	issues.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListIssues)
	issues.Get("/:issueKey", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetIssue)
	issues.Put("/:issueKey", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.UpdateIssue)
	issues.Delete("/:issueKey", rbac.RequirePermission(rbac.PermTaskDelete, rbacSvc), h.DeleteIssue)
	issues.Put("/:issueKey/position", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.MoveIssue)

	// JQL-поиск задач
	issues.Post("/jql", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.SearchIssuesJQL)

	// Issue links
	issues.Post("/:issueKey/links", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.CreateIssueLink)
	issues.Get("/:issueKey/links", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListIssueLinks)
	issues.Delete("/:issueKey/links/:linkId", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.DeleteIssueLink)

	// Comments
	issues.Post("/:issueKey/comments", rbac.RequirePermission(rbac.PermTaskComment, rbacSvc), h.CreateComment)
	issues.Get("/:issueKey/comments", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListComments)
	issues.Put("/:issueKey/comments/:commentId", rbac.RequirePermission(rbac.PermTaskComment, rbacSvc), h.UpdateComment)
	issues.Delete("/:issueKey/comments/:commentId", rbac.RequirePermission(rbac.PermTaskComment, rbacSvc), h.DeleteComment)
	issues.Post("/:issueKey/comments/:commentId/reactions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ToggleReaction)

	// Activity
	issues.Get("/:issueKey/activity", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetActivity)

	// Time tracking
	issues.Post("/:issueKey/time", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.LogTime)
	issues.Post("/:issueKey/timer/start", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.StartTimer)
	issues.Post("/:issueKey/timer/stop", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.StopTimer)
	issues.Get("/:issueKey/time-logs", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTimeLogs)

	// Boards
	boards := tracker.Group("/boards")
	boards.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListBoards)
	boards.Post("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.CreateBoard)
	boards.Get("/:boardId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetBoardData)

	// Workflow
	workflow := tracker.Group("/workflow")
	workflow.Get("/statuses", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListWorkflowStatuses)
	workflow.Post("/statuses", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateWorkflowStatus)
	workflow.Put("/statuses/:statusId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.UpdateWorkflowStatus)
	workflow.Delete("/statuses/:statusId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.DeleteWorkflowStatus)
	workflow.Get("/transitions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListWorkflowTransitions)
	workflow.Post("/transitions", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateWorkflowTransition)
	workflow.Delete("/transitions/:transitionId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.DeleteWorkflowTransition)

	// Labels
	labels := tracker.Group("/labels")
	labels.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListLabels)
	labels.Post("/", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.CreateLabel)
	labels.Delete("/:labelId", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.DeleteLabel)
	labels.Get("/groups", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListLabelGroups)
	labels.Post("/groups", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.CreateLabelGroup)

	// Custom Fields
	customFields := tracker.Group("/custom-fields")
	customFields.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListCustomFields)
	customFields.Post("/", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateCustomField)
	customFields.Delete("/:fieldId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.DeleteCustomField)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseIssueKey извлекает issueKey из параметров пути.
func parseIssueKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("issueKey"))
}
