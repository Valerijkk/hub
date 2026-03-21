package rbac

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// CustomRoleHandler обрабатывает HTTP-запросы для управления кастомными ролями проекта.
type CustomRoleHandler struct {
	svc *RBACService
}

// NewCustomRoleHandler создаёт новый обработчик кастомных ролей.
func NewCustomRoleHandler(svc *RBACService) *CustomRoleHandler {
	return &CustomRoleHandler{svc: svc}
}

// RegisterRoutes регистрирует маршруты для кастомных ролей проекта.
func (h *CustomRoleHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	roles := api.Group("/projects/:projectId/roles", authMw)

	roles.Post("/", RequirePermission(PermProjectManageRoles, h.svc), h.CreateRole)
	roles.Get("/", RequirePermission(PermProjectView, h.svc), h.ListRoles)
	roles.Put("/:roleId", RequirePermission(PermProjectManageRoles, h.svc), h.UpdateRole)
	roles.Delete("/:roleId", RequirePermission(PermProjectManageRoles, h.svc), h.DeleteRole)
	roles.Post("/:roleId/assign/:userId", RequirePermission(PermProjectManageRoles, h.svc), h.AssignRole)
	roles.Delete("/:roleId/assign/:userId", RequirePermission(PermProjectManageRoles, h.svc), h.UnassignRole)
}

// CreateRole обрабатывает POST /api/projects/:projectId/roles.
func (h *CustomRoleHandler) CreateRole(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}

	var req CustomRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	role, err := h.svc.CreateCustomRole(c.Context(), projectID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка создания роли")
	}
	return c.Status(fiber.StatusCreated).JSON(role)
}

// ListRoles обрабатывает GET /api/projects/:projectId/roles.
func (h *CustomRoleHandler) ListRoles(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}

	roles, err := h.svc.ListCustomRoles(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения ролей")
	}
	return c.JSON(roles)
}

// UpdateRole обрабатывает PUT /api/projects/:projectId/roles/:roleId.
func (h *CustomRoleHandler) UpdateRole(c fiber.Ctx) error {
	roleID, err := uuid.Parse(c.Params("roleId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный role ID")
	}

	var req CustomRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	role, err := h.svc.UpdateCustomRole(c.Context(), roleID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка обновления роли")
	}
	return c.JSON(role)
}

// DeleteRole обрабатывает DELETE /api/projects/:projectId/roles/:roleId.
func (h *CustomRoleHandler) DeleteRole(c fiber.Ctx) error {
	roleID, err := uuid.Parse(c.Params("roleId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный role ID")
	}

	if err := h.svc.DeleteCustomRole(c.Context(), roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка удаления роли")
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// AssignRole обрабатывает POST /api/projects/:projectId/roles/:roleId/assign/:userId.
func (h *CustomRoleHandler) AssignRole(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	roleID, err := uuid.Parse(c.Params("roleId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный role ID")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный user ID")
	}

	if err := h.svc.AssignCustomRole(c.Context(), projectID, userID, roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка назначения роли")
	}
	return c.JSON(fiber.Map{"status": "assigned"})
}

// UnassignRole обрабатывает DELETE /api/projects/:projectId/roles/:roleId/assign/:userId.
func (h *CustomRoleHandler) UnassignRole(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный user ID")
	}

	if err := h.svc.ClearCustomRole(c.Context(), projectID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка снятия роли")
	}
	return c.JSON(fiber.Map{"status": "unassigned"})
}
