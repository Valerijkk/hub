package project

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики модуля project.
type Handler struct {
	service *ProjectService
}

// NewHandler создаёт обработчик project.
func NewHandler(service *ProjectService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты project.
// authMw — middleware аутентификации.
// rbacSvc — проверка прав на маршрутах с :projectId (Fiber v3: RBAC после handler в сигнатуре).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	projects := api.Group("/projects", authMw)

	// CRUD проектов
	projects.Post("/", h.CreateProject)
	projects.Get("/", h.ListProjects)
	projects.Get("/trash", h.ListTrash) // до :projectId, чтобы избежать конфликта маршрутов
	projects.Get("/:projectId", h.GetProject, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	projects.Put("/:projectId", h.UpdateProject, rbac.RequirePermission(rbac.PermProjectEdit, rbacSvc))
	projects.Post("/:projectId/archive", h.ArchiveProject, rbac.RequirePermission(rbac.PermProjectEdit, rbacSvc))
	projects.Post("/:projectId/unarchive", h.UnarchiveProject, rbac.RequirePermission(rbac.PermProjectEdit, rbacSvc))
	projects.Delete("/:projectId", h.DeleteProject, rbac.RequirePermission(rbac.PermProjectDelete, rbacSvc))
	projects.Post("/:projectId/restore", h.RestoreProject, rbac.RequirePermission(rbac.PermProjectDelete, rbacSvc))
	projects.Delete("/:projectId/permanent", h.PermanentDeleteProject, rbac.RequirePermission(rbac.PermProjectDelete, rbacSvc))

	// Участники
	projects.Get("/:projectId/members", h.ListMembers, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	projects.Post("/:projectId/members", h.AddMember, rbac.RequirePermission(rbac.PermProjectManageMembers, rbacSvc))
	projects.Put("/:projectId/members/:userId", h.UpdateMemberRole, rbac.RequirePermission(rbac.PermProjectManageMembers, rbacSvc))
	projects.Delete("/:projectId/members/:userId", h.RemoveMember, rbac.RequirePermission(rbac.PermProjectManageMembers, rbacSvc))

	// Приглашения
	projects.Post("/:projectId/invites", h.CreateInvite, rbac.RequirePermission(rbac.PermInviteCreate, rbacSvc))
	projects.Get("/:projectId/invites", h.ListInvites, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	projects.Delete("/:projectId/invites/:inviteId", h.RevokeInvite, rbac.RequirePermission(rbac.PermInviteManage, rbacSvc))

	// Join — только auth, без project RBAC (invite-код сам определяет проект).
	api.Post("/invites/join", authMw, h.JoinByInvite)

	// Команды (орг-уровень без :projectId — auth; привязка к проекту — RBAC).
	teams := api.Group("/teams", authMw)
	teams.Post("/", h.CreateTeam)
	teams.Get("/", h.ListTeams)
	teams.Post("/:teamId/members", h.AddTeamMember)
	teams.Delete("/:teamId/members/:userId", h.RemoveTeamMember)
	teams.Post("/:teamId/projects/:projectId", h.AddTeamToProject, rbac.RequirePermission(rbac.PermTeamManage, rbacSvc))
}

// CreateProject — Создать проект
// @Summary Создать проект
// @Description Создаёт новый проект. Создатель автоматически становится владельцем.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param body body CreateProjectRequest true "Данные проекта"
// @Success 201 {object} ProjectResponse "Проект создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects [post]
func (h *Handler) CreateProject(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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

// ListProjects — Список проектов
// @Summary Список проектов
// @Description Возвращает проекты, в которых состоит текущий пользователь.
// @Tags Проекты
// @Produce json
// @Param include_archived query bool false "Включить архивные проекты"
// @Success 200 {array} ProjectResponse "Список проектов"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects [get]
func (h *Handler) ListProjects(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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

// GetProject — Получить проект
// @Summary Получить проект
// @Description Возвращает детальную информацию о проекте.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} ProjectResponse "Данные проекта"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 404 {object} fiber.Error "Проект не найден"
// @Security BearerAuth
// @Router /api/projects/{projectId} [get]
func (h *Handler) GetProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetProject(c.Context(), projectID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateProject — Обновить проект
// @Summary Обновить проект
// @Description Обновляет название, описание и настройки проекта.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body UpdateProjectRequest true "Обновляемые данные"
// @Success 200 {object} ProjectResponse "Проект обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId} [put]
func (h *Handler) UpdateProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
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

// ArchiveProject — Архивировать проект
// @Summary Архивировать проект
// @Description Переводит проект в архивный статус. Задачи становятся только для чтения.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} ProjectResponse "Проект архивирован"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/archive [post]
func (h *Handler) ArchiveProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ArchiveProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UnarchiveProject — Разархивировать проект
// @Summary Разархивировать проект
// @Description Восстанавливает проект из архива.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} ProjectResponse "Проект разархивирован"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/unarchive [post]
func (h *Handler) UnarchiveProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.UnarchiveProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteProject — Удалить проект
// @Summary Удалить проект
// @Description Перемещает проект в корзину (мягкое удаление).
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} map[string]string "Проект перемещён в корзину"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId} [delete]
func (h *Handler) DeleteProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteProject(c.Context(), projectID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Project moved to trash"})
}

// RestoreProject — Восстановить проект
// @Summary Восстановить проект
// @Description Восстанавливает проект из корзины.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} ProjectResponse "Проект восстановлен"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/restore [post]
func (h *Handler) RestoreProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.RestoreProject(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// PermanentDeleteProject — Жёсткое удаление проекта из корзины
// @Summary Удалить проект навсегда
// @Description Безвозвратно удаляет проект (hard-delete). TZ: Admin/GDPR.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 204 "Удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 404 {object} fiber.Error "Не найден"
// @Security BearerAuth
// @Router /api/projects/{projectId}/permanent [delete]
func (h *Handler) PermanentDeleteProject(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	if err := h.service.PermanentDelete(c.Context(), projectID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListTrash — Корзина проектов
// @Summary Корзина проектов
// @Description Возвращает список удалённых проектов пользователя.
// @Tags Проекты
// @Produce json
// @Success 200 {array} ProjectResponse "Удалённые проекты"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/trash [get]
func (h *Handler) ListTrash(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListTrash(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListMembers — Список участников проекта
// @Summary Список участников проекта
// @Description Возвращает всех участников проекта с их ролями.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} MemberResponse "Участники проекта"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/members [get]
func (h *Handler) ListMembers(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListMembers(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddMember — Добавить участника в проект
// @Summary Добавить участника в проект
// @Description Добавляет пользователя в проект с указанной ролью.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body AddMemberRequest true "ID пользователя и роль"
// @Success 201 {object} MemberResponse "Участник добавлен"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/members [post]
func (h *Handler) AddMember(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
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

// UpdateMemberRole — Изменить роль участника
// @Summary Изменить роль участника
// @Description Обновляет роль участника в проекте.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param userId path string true "ID пользователя"
// @Param body body UpdateMemberRoleRequest true "Новая роль"
// @Success 200 {object} MemberResponse "Роль обновлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/members/{userId} [put]
func (h *Handler) UpdateMemberRole(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
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

// RemoveMember — Удалить участника из проекта
// @Summary Удалить участника из проекта
// @Description Удаляет пользователя из проекта.
// @Tags Проекты
// @Param projectId path string true "ID проекта"
// @Param userId path string true "ID пользователя"
// @Success 204 "Участник удалён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/members/{userId} [delete]
func (h *Handler) RemoveMember(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
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

// CreateInvite — Создать приглашение в проект
// @Summary Создать приглашение в проект
// @Description Генерирует инвайт-ссылку для вступления в проект.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateInviteRequest true "Параметры приглашения"
// @Success 201 {object} InviteResponse "Приглашение создано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/invites [post]
func (h *Handler) CreateInvite(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	userID, err := httputil.UserIDFromCtx(c)
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

// ListInvites — Список приглашений проекта
// @Summary Список приглашений проекта
// @Description Возвращает все активные приглашения указанного проекта.
// @Tags Проекты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} InviteResponse "Список приглашений"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/invites [get]
func (h *Handler) ListInvites(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListInvites(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// RevokeInvite — Отозвать приглашение
// @Summary Отозвать приглашение
// @Description Деактивирует приглашение в проект.
// @Tags Проекты
// @Param projectId path string true "ID проекта"
// @Param inviteId path string true "ID приглашения"
// @Success 204 "Приглашение отозвано"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/invites/{inviteId} [delete]
func (h *Handler) RevokeInvite(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
	}
	inviteIDStr := utils.CopyString(c.Params("inviteId"))
	inviteUUID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid invite ID")
	}

	if err := h.service.RevokeInvite(c.Context(), projectID, inviteUUID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// JoinByInvite — Вступить по приглашению
// @Summary Вступить по приглашению
// @Description Вступает в проект по инвайт-коду.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param body body JoinByInviteRequest true "Код приглашения"
// @Success 200 {object} ProjectResponse "Вступление в проект"
// @Failure 400 {object} fiber.Error "Неверный код"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/invites/join [post]
func (h *Handler) JoinByInvite(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
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

// CreateTeam — Создать команду
// @Summary Создать команду
// @Description Создаёт новую команду.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param body body CreateTeamRequest true "Данные команды"
// @Success 201 {object} TeamResponse "Команда создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/teams [post]
func (h *Handler) CreateTeam(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateTeamRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "team name is required")
	}
	resp, err := h.service.CreateTeam(c.Context(), userID, req.Name, req.Description)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListTeams — Список команд
// @Summary Список команд
// @Description Возвращает все команды.
// @Tags Проекты
// @Produce json
// @Success 200 {array} TeamResponse "Список команд"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/teams [get]
func (h *Handler) ListTeams(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.ListTeams(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

// AddTeamMember — Добавить участника в команду
// @Summary Добавить участника в команду
// @Description Добавляет пользователя в команду.
// @Tags Проекты
// @Accept json
// @Produce json
// @Param teamId path string true "ID команды"
// @Param body body AddTeamMemberRequest true "ID пользователя"
// @Success 201 {object} map[string]string "Участник добавлен"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/teams/{teamId}/members [post]
func (h *Handler) AddTeamMember(c fiber.Ctx) error {
	actorID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
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
	if err := h.service.AddTeamMember(c.Context(), teamID, actorID, userID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "member added to team"})
}

// RemoveTeamMember — Удалить участника из команды
// @Summary Удалить участника из команды
// @Description Удаляет пользователя из команды.
// @Tags Проекты
// @Param teamId path string true "ID команды"
// @Param userId path string true "ID пользователя"
// @Success 204 "Участник удалён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/teams/{teamId}/members/{userId} [delete]
func (h *Handler) RemoveTeamMember(c fiber.Ctx) error {
	actorID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
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
	if err := h.service.RemoveTeamMember(c.Context(), teamID, actorID, userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// AddTeamToProject — Привязать команду к проекту
// @Summary Привязать команду к проекту
// @Description Добавляет команду в проект — все участники команды получают доступ.
// @Tags Проекты
// @Produce json
// @Param teamId path string true "ID команды"
// @Param projectId path string true "ID проекта"
// @Success 200 {object} map[string]string "Команда добавлена в проект"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/teams/{teamId}/projects/{projectId} [post]
func (h *Handler) AddTeamToProject(c fiber.Ctx) error {
	teamIDStr := utils.CopyString(c.Params("teamId"))
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid team ID")
	}
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
	}
	actorID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	if err := h.service.AddTeamToProject(c.Context(), teamID, projectID, actorID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "team added to project"})
}
