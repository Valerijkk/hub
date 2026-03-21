package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/search"
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

// parseIssueKey извлекает issueKey из параметров пути.
func parseIssueKey(c fiber.Ctx) string {
	return utils.CopyString(c.Params("issueKey"))
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

// ---------------------------------------------------------------------------
// Issue handlers
// ---------------------------------------------------------------------------

// CreateIssue обрабатывает POST /api/projects/:projectId/issues
func (h *Handler) CreateIssue(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateIssue(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssues обрабатывает GET /api/projects/:projectId/issues
func (h *Handler) ListIssues(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req ListIssuesRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректные параметры запроса")
	}

	resp, err := h.service.ListIssues(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetIssue обрабатывает GET /api/projects/:projectId/issues/:issueKey
func (h *Handler) GetIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetIssue(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateIssue обрабатывает PUT /api/projects/:projectId/issues/:issueKey
func (h *Handler) UpdateIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateIssue(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssue обрабатывает DELETE /api/projects/:projectId/issues/:issueKey
func (h *Handler) DeleteIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssue(c.Context(), issueKey, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MoveIssue обрабатывает PUT /api/projects/:projectId/issues/:issueKey/position
func (h *Handler) MoveIssue(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req MoveIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.MoveIssue(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Issue link handlers
// ---------------------------------------------------------------------------

// CreateIssueLink обрабатывает POST /api/projects/:projectId/issues/:issueKey/links
func (h *Handler) CreateIssueLink(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateIssueLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateIssueLink(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssueLinks обрабатывает GET /api/projects/:projectId/issues/:issueKey/links
func (h *Handler) ListIssueLinks(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListIssueLinks(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssueLink обрабатывает DELETE /api/projects/:projectId/issues/:issueKey/links/:linkId
func (h *Handler) DeleteIssueLink(c fiber.Ctx) error {
	linkID, err := parseUUIDParam(c, "linkId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssueLink(c.Context(), linkID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Comment handlers
// ---------------------------------------------------------------------------

// CreateComment обрабатывает POST /api/projects/:projectId/issues/:issueKey/comments
func (h *Handler) CreateComment(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateComment(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListComments обрабатывает GET /api/projects/:projectId/issues/:issueKey/comments
func (h *Handler) ListComments(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListComments(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateComment обрабатывает PUT /api/projects/:projectId/issues/:issueKey/comments/:commentId
func (h *Handler) UpdateComment(c fiber.Ctx) error {
	commentID, err := parseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}

	var req UpdateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateComment(c.Context(), commentID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteComment обрабатывает DELETE /api/projects/:projectId/issues/:issueKey/comments/:commentId
func (h *Handler) DeleteComment(c fiber.Ctx) error {
	commentID, err := parseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteComment(c.Context(), commentID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleReaction обрабатывает POST /api/projects/:projectId/issues/:issueKey/comments/:commentId/reactions
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
	commentID, err := parseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	added, err := h.service.ToggleReaction(c.Context(), commentID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// ---------------------------------------------------------------------------
// Activity handler
// ---------------------------------------------------------------------------

// GetActivity обрабатывает GET /api/projects/:projectId/issues/:issueKey/activity
func (h *Handler) GetActivity(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetActivity(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Time tracking handlers
// ---------------------------------------------------------------------------

// LogTime обрабатывает POST /api/projects/:projectId/issues/:issueKey/time
func (h *Handler) LogTime(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req LogTimeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.LogTime(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StartTimer обрабатывает POST /api/projects/:projectId/issues/:issueKey/timer/start
func (h *Handler) StartTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StartTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StopTimer обрабатывает POST /api/projects/:projectId/issues/:issueKey/timer/stop
func (h *Handler) StopTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StopTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTimeLogs обрабатывает GET /api/projects/:projectId/issues/:issueKey/time-logs
func (h *Handler) ListTimeLogs(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListTimeLogs(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Board handlers
// ---------------------------------------------------------------------------

// ListBoards обрабатывает GET /api/projects/:projectId/boards
func (h *Handler) ListBoards(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListBoards(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateBoard обрабатывает POST /api/projects/:projectId/boards
func (h *Handler) CreateBoard(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateBoardRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateBoard(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetBoardData обрабатывает GET /api/projects/:projectId/boards/:boardId
func (h *Handler) GetBoardData(c fiber.Ctx) error {
	boardID, err := parseUUIDParam(c, "boardId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetBoardData(c.Context(), boardID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Workflow handlers
// ---------------------------------------------------------------------------

// ListWorkflowStatuses обрабатывает GET /api/projects/:projectId/workflow/statuses
func (h *Handler) ListWorkflowStatuses(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListWorkflowStatuses(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateWorkflowStatus обрабатывает POST /api/projects/:projectId/workflow/statuses
func (h *Handler) CreateWorkflowStatus(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateWorkflowStatus(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateWorkflowStatus обрабатывает PUT /api/projects/:projectId/workflow/statuses/:statusId
func (h *Handler) UpdateWorkflowStatus(c fiber.Ctx) error {
	statusID, err := parseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	var req CreateWorkflowStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateWorkflowStatus(c.Context(), statusID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteWorkflowStatus обрабатывает DELETE /api/projects/:projectId/workflow/statuses/:statusId
func (h *Handler) DeleteWorkflowStatus(c fiber.Ctx) error {
	statusID, err := parseUUIDParam(c, "statusId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowStatus(c.Context(), statusID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListWorkflowTransitions обрабатывает GET /api/projects/:projectId/workflow/transitions
func (h *Handler) ListWorkflowTransitions(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListWorkflowTransitions(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateWorkflowTransition обрабатывает POST /api/projects/:projectId/workflow/transitions
func (h *Handler) CreateWorkflowTransition(c fiber.Ctx) error {
	var req CreateWorkflowTransitionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateWorkflowTransition(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteWorkflowTransition обрабатывает DELETE /api/projects/:projectId/workflow/transitions/:transitionId
func (h *Handler) DeleteWorkflowTransition(c fiber.Ctx) error {
	transitionID, err := parseUUIDParam(c, "transitionId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteWorkflowTransition(c.Context(), transitionID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Label handlers
// ---------------------------------------------------------------------------

// ListLabels обрабатывает GET /api/projects/:projectId/labels
func (h *Handler) ListLabels(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListLabels(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateLabel обрабатывает POST /api/projects/:projectId/labels
func (h *Handler) CreateLabel(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateLabel(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteLabel обрабатывает DELETE /api/projects/:projectId/labels/:labelId
func (h *Handler) DeleteLabel(c fiber.Ctx) error {
	labelID, err := parseUUIDParam(c, "labelId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteLabel(c.Context(), labelID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListLabelGroups обрабатывает GET /api/projects/:projectId/labels/groups
func (h *Handler) ListLabelGroups(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListLabelGroups(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateLabelGroup обрабатывает POST /api/projects/:projectId/labels/groups
func (h *Handler) CreateLabelGroup(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateLabelGroupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateLabelGroup(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ---------------------------------------------------------------------------
// Custom field handlers
// ---------------------------------------------------------------------------

// ListCustomFields обрабатывает GET /api/projects/:projectId/custom-fields
func (h *Handler) ListCustomFields(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListCustomFields(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateCustomField обрабатывает POST /api/projects/:projectId/custom-fields
func (h *Handler) CreateCustomField(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateCustomFieldRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateCustomField(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteCustomField обрабатывает DELETE /api/projects/:projectId/custom-fields/:fieldId
func (h *Handler) DeleteCustomField(c fiber.Ctx) error {
	fieldID, err := parseUUIDParam(c, "fieldId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteCustomField(c.Context(), fieldID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// JQL-поиск
// ---------------------------------------------------------------------------

// jqlRequest — тело запроса для JQL-поиска задач.
type jqlRequest struct {
	Query string `json:"query" validate:"required"`
}

// SearchIssuesJQL обрабатывает POST /api/projects/:projectId/issues/jql.
// Парсит JQL-строку, конвертирует в Ent-предикаты и возвращает отфильтрованные задачи.
func (h *Handler) SearchIssuesJQL(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req jqlRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}
	if req.Query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query обязателен")
	}

	// Парсинг JQL-запроса.
	parsed, err := search.ParseJQL(req.Query)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ошибка JQL: "+err.Error())
	}

	// Конвертация в Ent-предикаты.
	predicates := search.ToEntPredicates(parsed, userID)

	// Базовый запрос по проекту.
	query := h.service.repo.GetDB().Issue.Query().
		Where(issue.HasProjectWith(project.ID(projectID))).
		Where(issue.DeletedAtIsNil())

	// Применяем JQL-предикаты.
	for _, pred := range predicates {
		query = query.Where(pred)
	}

	// Сортировка из JQL.
	if parsed.OrderBy != nil {
		switch parsed.OrderBy.Field {
		case "created":
			query = query.Order(issue.ByCreatedAt())
		case "priority":
			query = query.Order(issue.ByPriority())
		case "updated":
			query = query.Order(issue.ByUpdatedAt())
		}
	}

	// Выполнение запроса с eager-loading связей.
	issues, err := query.
		WithStatus().
		WithAssignee().
		Limit(100).
		All(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка выполнения JQL: "+err.Error())
	}

	// Конвертация в ответ — переиспользуем формат ListIssues.
	_ = projectID // projectID используется для контекста RBAC (проверяется middleware)
	type jqlIssueItem struct {
		Key       string `json:"key"`
		Title     string `json:"title"`
		Type      string `json:"type"`
		Priority  string `json:"priority"`
		Status    string `json:"status,omitempty"`
		Assignee  string `json:"assignee,omitempty"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]jqlIssueItem, 0, len(issues))
	for _, iss := range issues {
		status := ""
		if iss.Edges.Status != nil {
			status = iss.Edges.Status.Name
		}
		assignee := ""
		if iss.Edges.Assignee != nil {
			if iss.Edges.Assignee.Name != "" {
				assignee = iss.Edges.Assignee.Name
			} else {
				assignee = iss.Edges.Assignee.Email
			}
		}
		result = append(result, jqlIssueItem{
			Key:       iss.IssueKey,
			Title:     iss.Title,
			Type:      string(iss.IssueType),
			Priority:  string(iss.Priority),
			Status:    status,
			Assignee:  assignee,
			CreatedAt: iss.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(fiber.Map{
		"query":   req.Query,
		"total":   len(result),
		"results": result,
	})
}
