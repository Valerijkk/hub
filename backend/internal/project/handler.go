package project

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// Handler provides HTTP handlers for the project module.
type Handler struct {
	service *ProjectService
}

// NewHandler creates a new project handler.
func NewHandler(service *ProjectService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all project-related routes.
// authMw is the authentication middleware.
// requirePerm is a factory that returns a permission-checking middleware.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	projects := api.Group("/projects", authMw)

	// Project CRUD
	projects.Post("/", h.CreateProject)
	projects.Get("/", h.ListProjects)
	projects.Get("/trash", h.ListTrash) // before :projectId to avoid conflict
	projects.Get("/:projectId", h.GetProject)
	projects.Put("/:projectId", h.UpdateProject)
	projects.Post("/:projectId/archive", h.ArchiveProject)
	projects.Post("/:projectId/unarchive", h.UnarchiveProject)
	projects.Delete("/:projectId", h.DeleteProject)
	projects.Post("/:projectId/restore", h.RestoreProject)

	// Members
	projects.Get("/:projectId/members", h.ListMembers)
	projects.Post("/:projectId/members", h.AddMember)
	projects.Put("/:projectId/members/:userId", h.UpdateMemberRole)
	projects.Delete("/:projectId/members/:userId", h.RemoveMember)

	// Invites
	projects.Post("/:projectId/invites", h.CreateInvite)
	projects.Get("/:projectId/invites", h.ListInvites)
	projects.Delete("/:projectId/invites/:inviteId", h.RevokeInvite)

	// Join (public-ish, only auth required)
	api.Post("/invites/join", authMw, h.JoinByInvite)

	// Teams
	teams := api.Group("/teams", authMw)
	teams.Post("/", h.CreateTeam)
	teams.Get("/", h.ListTeams)
	teams.Post("/:teamId/members", h.AddTeamMember)
	teams.Delete("/:teamId/members/:userId", h.RemoveTeamMember)
	teams.Post("/:teamId/projects/:projectId", h.AddTeamToProject)
}

// userIDFromCtx extracts user ID from Fiber Locals, using CopyString for safety.
func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}
	return id, nil
}

// parseProjectID extracts and validates the projectId path parameter.
func parseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID required")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
	}
	return id, nil
}

// CreateProject handles POST /api/projects
func (h *Handler) CreateProject(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateProjectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.CreateProject(c.Context(), userID, req.Name, req.Key, req.Description)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListProjects handles GET /api/projects
func (h *Handler) ListProjects(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	includeArchived := c.Query("include_archived") == "true"
	resp, err := h.service.ListProjects(c.Context(), userID, includeArchived)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetProject handles GET /api/projects/:projectId
func (h *Handler) GetProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateProject handles PUT /api/projects/:projectId
func (h *Handler) UpdateProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req UpdateProjectRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.UpdateProject(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ArchiveProject handles POST /api/projects/:projectId/archive
func (h *Handler) ArchiveProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ArchiveProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UnarchiveProject handles POST /api/projects/:projectId/unarchive
func (h *Handler) UnarchiveProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.UnarchiveProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteProject handles DELETE /api/projects/:projectId
func (h *Handler) DeleteProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteProject(c.Context(), projectID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Project moved to trash"})
}

// RestoreProject handles POST /api/projects/:projectId/restore
func (h *Handler) RestoreProject(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.RestoreProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTrash handles GET /api/projects/trash
func (h *Handler) ListTrash(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListTrash(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListMembers handles GET /api/projects/:projectId/members
func (h *Handler) ListMembers(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListMembers(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddMember handles POST /api/projects/:projectId/members
func (h *Handler) AddMember(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req AddMemberRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	memberUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	resp, err := h.service.AddMember(c.Context(), projectID, memberUUID, req.Role)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateMemberRole handles PUT /api/projects/:projectId/members/:userId
func (h *Handler) UpdateMemberRole(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	userIDStr := utils.CopyString(c.Params("userId"))
	memberUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.UpdateMemberRole(c.Context(), projectID, memberUUID, req.Role)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// RemoveMember handles DELETE /api/projects/:projectId/members/:userId
func (h *Handler) RemoveMember(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	userIDStr := utils.CopyString(c.Params("userId"))
	memberUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	if err := h.service.RemoveMember(c.Context(), projectID, memberUUID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// CreateInvite handles POST /api/projects/:projectId/invites
func (h *Handler) CreateInvite(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.CreateInvite(c.Context(), projectID, userID, req.Role, req.MaxUses, req.ExpiresAt)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListInvites handles GET /api/projects/:projectId/invites
func (h *Handler) ListInvites(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListInvites(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// RevokeInvite handles DELETE /api/projects/:projectId/invites/:inviteId
func (h *Handler) RevokeInvite(c fiber.Ctx) error {
	inviteIDStr := utils.CopyString(c.Params("inviteId"))
	inviteUUID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid invite ID")
	}

	if err := h.service.RevokeInvite(c.Context(), inviteUUID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// JoinByInvite handles POST /api/invites/join
func (h *Handler) JoinByInvite(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req JoinByInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.JoinByInvite(c.Context(), req.Code, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateTeam handles POST /api/teams.
func (h *Handler) CreateTeam(c fiber.Ctx) error {
	var req CreateTeamRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "team name is required")
	}
	resp, err := h.service.CreateTeam(c.Context(), req.Name, req.Description)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTeams handles GET /api/teams.
func (h *Handler) ListTeams(c fiber.Ctx) error {
	resp, err := h.service.ListTeams(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

// AddTeamMember handles POST /api/teams/:teamId/members.
func (h *Handler) AddTeamMember(c fiber.Ctx) error {
	teamIDStr := utils.CopyString(c.Params("teamId"))
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid team ID")
	}
	var req AddTeamMemberRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}
	if err := h.service.AddTeamMember(c.Context(), teamID, userID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "member added to team"})
}

// RemoveTeamMember handles DELETE /api/teams/:teamId/members/:userId.
func (h *Handler) RemoveTeamMember(c fiber.Ctx) error {
	teamIDStr := utils.CopyString(c.Params("teamId"))
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid team ID")
	}
	userIDStr := utils.CopyString(c.Params("userId"))
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}
	if err := h.service.RemoveTeamMember(c.Context(), teamID, userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// AddTeamToProject handles POST /api/teams/:teamId/projects/:projectId.
func (h *Handler) AddTeamToProject(c fiber.Ctx) error {
	teamIDStr := utils.CopyString(c.Params("teamId"))
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid team ID")
	}
	projectID, err := parseProjectID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
	}
	if err := h.service.AddTeamToProject(c.Context(), teamID, projectID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "team added to project"})
}
