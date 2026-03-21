package sprint

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля спринтов.
type Handler struct {
	service *SprintService
	poker   *PokerService
}

// NewHandler создаёт новый обработчик спринтов.
func NewHandler(service *SprintService, poker *PokerService) *Handler {
	return &Handler{service: service, poker: poker}
}

// RegisterRoutes регистрирует все маршруты модуля спринтов.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	sprints := api.Group("/projects/:projectId/sprints", authMw)

	sprints.Post("/", rbac.RequirePermission(rbac.PermTaskCreate, rbacSvc), h.CreateSprint)
	sprints.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListSprints)
	sprints.Get("/:sprintId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetSprint)
	sprints.Put("/:sprintId", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.UpdateSprint)
	sprints.Post("/:sprintId/start", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.StartSprint)
	sprints.Post("/:sprintId/complete", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.CompleteSprint)
	sprints.Put("/:sprintId/retro", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.UpdateRetroNotes)
	sprints.Get("/:sprintId/burndown", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetBurndown)
	sprints.Post("/:sprintId/poker", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.StartPokerSession)
	sprints.Post("/:sprintId/issues", rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc), h.AddIssuesToSprint)

	// Velocity на уровне проекта
	velocity := api.Group("/projects/:projectId/velocity", authMw)
	velocity.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetVelocity)

	// Poker маршруты без projectId
	poker := api.Group("/sprints/poker", authMw)
	poker.Post("/:sessionId/vote", h.CastVote)
	poker.Post("/:sessionId/reveal", h.RevealVotes)
	poker.Get("/:sessionId", h.GetPokerSession)
}

// ---------------------------------------------------------------------------
// Sprint handlers
// ---------------------------------------------------------------------------

// CreateSprint обрабатывает POST /api/projects/:projectId/sprints.
func (h *Handler) CreateSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateSprint(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListSprints обрабатывает GET /api/projects/:projectId/sprints.
func (h *Handler) ListSprints(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListSprints(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetSprint обрабатывает GET /api/projects/:projectId/sprints/:sprintId.
func (h *Handler) GetSprint(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetSprint(c.Context(), sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateSprint обрабатывает PUT /api/projects/:projectId/sprints/:sprintId.
func (h *Handler) UpdateSprint(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	var req UpdateSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateSprint(c.Context(), sprintID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// StartSprint обрабатывает POST /api/projects/:projectId/sprints/:sprintId/start.
func (h *Handler) StartSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.StartSprint(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CompleteSprint обрабатывает POST /api/projects/:projectId/sprints/:sprintId/complete.
func (h *Handler) CompleteSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.CompleteSprint(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateRetroNotes обрабатывает PUT /api/projects/:projectId/sprints/:sprintId/retro.
func (h *Handler) UpdateRetroNotes(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	var req RetroNotesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateRetroNotes(c.Context(), sprintID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetBurndown обрабатывает GET /api/projects/:projectId/sprints/:sprintId/burndown.
func (h *Handler) GetBurndown(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetBurndown(c.Context(), sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetVelocity обрабатывает GET /api/projects/:projectId/velocity.
func (h *Handler) GetVelocity(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetVelocity(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddIssuesToSprint обрабатывает POST /api/projects/:projectId/sprints/:sprintId/issues.
func (h *Handler) AddIssuesToSprint(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	var req MoveIssueToSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	issueIDs := make([]uuid.UUID, 0, len(req.IssueIDs))
	for _, idStr := range req.IssueIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный issue_id: "+idStr)
		}
		issueIDs = append(issueIDs, id)
	}

	if err := h.service.AddIssuesToSprint(c.Context(), sprintID, issueIDs); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Poker handlers
// ---------------------------------------------------------------------------

// StartPokerSession обрабатывает POST /api/projects/:projectId/sprints/:sprintId/poker.
func (h *Handler) StartPokerSession(c fiber.Ctx) error {
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	var req StartPokerRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	issueID, err := uuid.Parse(req.IssueID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный issue_id")
	}

	resp, err := h.poker.StartPokerSession(c.Context(), sprintID, issueID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// CastVote обрабатывает POST /api/sprints/poker/:sessionId/vote.
func (h *Handler) CastVote(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req VoteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	if err := h.poker.CastVote(c.Context(), sessionID, userID, req.Value); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RevealVotes обрабатывает POST /api/sprints/poker/:sessionId/reveal.
func (h *Handler) RevealVotes(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	resp, err := h.poker.RevealVotes(c.Context(), sessionID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetPokerSession обрабатывает GET /api/sprints/poker/:sessionId.
func (h *Handler) GetPokerSession(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	resp, err := h.poker.GetSession(c.Context(), sessionID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

