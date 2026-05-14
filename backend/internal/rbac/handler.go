package rbac

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
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
// @Summary Создать кастомную роль
// @Description Создаёт новую роль с набором разрешений в проекте
// @Tags Роли и права
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param body body CustomRoleRequest true "Название и разрешения роли"
// @Success 201 {object} object "Созданная роль"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 500 {object} map[string]string "Ошибка создания"
// @Router /projects/{projectId}/roles [post]
func (h *CustomRoleHandler) CreateRole(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
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
// @Summary Получить список ролей проекта
// @Description Возвращает все кастомные роли с их разрешениями
// @Tags Роли и права
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Success 200 {array} object "Список ролей"
// @Failure 400 {object} map[string]string "Некорректный ID проекта"
// @Failure 500 {object} map[string]string "Ошибка получения"
// @Router /projects/{projectId}/roles [get]
func (h *CustomRoleHandler) ListRoles(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	roles, err := h.svc.ListCustomRoles(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения ролей")
	}
	return c.JSON(roles)
}

// UpdateRole обрабатывает PUT /api/projects/:projectId/roles/:roleId.
// @Summary Обновить кастомную роль
// @Description Обновляет название и/или разрешения роли
// @Tags Роли и права
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param roleId path string true "ID роли"
// @Param body body CustomRoleRequest true "Новые данные роли"
// @Success 200 {object} object "Обновлённая роль"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 500 {object} map[string]string "Ошибка обновления"
// @Router /projects/{projectId}/roles/{roleId} [put]
func (h *CustomRoleHandler) UpdateRole(c fiber.Ctx) error {
	roleID, err := httputil.ParseUUIDParam(c, "roleId")
	if err != nil {
		return err
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
// @Summary Удалить кастомную роль
// @Description Удаляет роль из проекта. Пользователи с этой ролью потеряют её
// @Tags Роли и права
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param roleId path string true "ID роли"
// @Success 200 {object} map[string]string "Роль удалена"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 500 {object} map[string]string "Ошибка удаления"
// @Router /projects/{projectId}/roles/{roleId} [delete]
func (h *CustomRoleHandler) DeleteRole(c fiber.Ctx) error {
	roleID, err := httputil.ParseUUIDParam(c, "roleId")
	if err != nil {
		return err
	}

	if err := h.svc.DeleteCustomRole(c.Context(), roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка удаления роли")
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

// AssignRole обрабатывает POST /api/projects/:projectId/roles/:roleId/assign/:userId.
// @Summary Назначить роль пользователю
// @Description Назначает кастомную роль участнику проекта
// @Tags Роли и права
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param roleId path string true "ID роли"
// @Param userId path string true "ID пользователя"
// @Success 200 {object} map[string]string "Роль назначена"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 500 {object} map[string]string "Ошибка назначения"
// @Router /projects/{projectId}/roles/{roleId}/assign/{userId} [post]
func (h *CustomRoleHandler) AssignRole(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	roleID, err := httputil.ParseUUIDParam(c, "roleId")
	if err != nil {
		return err
	}
	userID, err := httputil.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	if err := h.svc.AssignCustomRole(c.Context(), projectID, userID, roleID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка назначения роли")
	}
	return c.JSON(fiber.Map{"status": "assigned"})
}

// UnassignRole обрабатывает DELETE /api/projects/:projectId/roles/:roleId/assign/:userId.
// @Summary Снять роль с пользователя
// @Description Снимает кастомную роль с участника проекта
// @Tags Роли и права
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param roleId path string true "ID роли"
// @Param userId path string true "ID пользователя"
// @Success 200 {object} map[string]string "Роль снята"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 500 {object} map[string]string "Ошибка снятия роли"
// @Router /projects/{projectId}/roles/{roleId}/assign/{userId} [delete]
func (h *CustomRoleHandler) UnassignRole(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	if err := h.svc.ClearCustomRole(c.Context(), projectID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка снятия роли")
	}
	return c.JSON(fiber.Map{"status": "unassigned"})
}
