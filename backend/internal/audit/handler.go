package audit

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/user"
)

// Handler предоставляет HTTP-обработчики для административных эндпоинтов аудита.
type Handler struct {
	service *AuditService
}

// NewHandler создаёт новый обработчик аудита.
func NewHandler(service *AuditService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutesSimple регистрирует маршруты администратора (все admin-only).
func (h *Handler) RegisterRoutesSimple(api fiber.Router, authMw fiber.Handler) {
	admin := api.Group("/admin", authMw, h.requireAdmin())

	admin.Get("/audit", h.ListAuditLogs)
	admin.Get("/limits", h.GetLimits)
	admin.Put("/limits/:key", h.UpdateLimit)
	admin.Get("/trash", h.ListTrash)
	admin.Post("/trash/:type/:id/restore", h.RestoreItem)
	admin.Post("/users/:id/deactivate", h.DeactivateUser)
	admin.Post("/users/:id/reassign", h.ReassignTasks)
	admin.Delete("/users/:id/gdpr", h.DeleteUserGDPR)
}

// requireAdmin — middleware, проверяющий что пользователь является администратором.
func (h *Handler) requireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
		}

		u, err := h.service.db.User.Get(c.Context(), userID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "пользователь не найден")
		}

		if u.SystemRole != user.SystemRoleAdmin {
			return fiber.NewError(fiber.StatusForbidden, "доступ только для администратора")
		}

		return c.Next()
	}
}

// ListAuditLogs обрабатывает GET /api/admin/audit.
func (h *Handler) ListAuditLogs(c fiber.Ctx) error {
	filters := AuditLogFilters{
		Action:     utils.CopyString(c.Query("action")),
		EntityType: utils.CopyString(c.Query("entity_type")),
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err == nil {
			filters.UserID = &id
		}
	}

	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		id, err := uuid.Parse(projectIDStr)
		if err == nil {
			filters.ProjectID = &id
		}
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		t, err := time.Parse(time.RFC3339, dateFrom)
		if err == nil {
			filters.DateFrom = &t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		t, err := time.Parse(time.RFC3339, dateTo)
		if err == nil {
			filters.DateTo = &t
		}
	}

	filters.Page, _ = strconv.Atoi(c.Query("page", "1"))
	filters.PerPage, _ = strconv.Atoi(c.Query("per_page", "50"))

	resp, err := h.service.ListAuditLogs(c.Context(), filters)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(resp)
}

// GetLimits обрабатывает GET /api/admin/limits.
func (h *Handler) GetLimits(c fiber.Ctx) error {
	limits, err := h.service.GetResourceLimits(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(limits)
}

// UpdateLimit обрабатывает PUT /api/admin/limits/:key.
func (h *Handler) UpdateLimit(c fiber.Ctx) error {
	key := utils.CopyString(c.Params("key"))
	if key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "ключ лимита обязателен")
	}

	var req UpdateLimitRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	if err := h.service.UpdateResourceLimit(c.Context(), key, req.Value); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"success": true})
}

// ListTrash обрабатывает GET /api/admin/trash.
func (h *Handler) ListTrash(c fiber.Ctx) error {
	entityType := utils.CopyString(c.Query("type", "issue"))

	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		id, err := uuid.Parse(pidStr)
		if err == nil {
			projectID = &id
		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "50"))

	resp, err := h.service.ListTrash(c.Context(), entityType, projectID, page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(resp)
}

// RestoreItem обрабатывает POST /api/admin/trash/:type/:id/restore.
func (h *Handler) RestoreItem(c fiber.Ctx) error {
	entityType := utils.CopyString(c.Params("type"))
	idStr := utils.CopyString(c.Params("id"))

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID")
	}

	if err := h.service.RestoreItem(c.Context(), entityType, entityID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Элемент восстановлен"})
}

// DeactivateUser обрабатывает POST /api/admin/users/:id/deactivate.
func (h *Handler) DeactivateUser(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	if err := h.service.DeactivateUser(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Пользователь деактивирован"})
}

// ReassignTasks обрабатывает POST /api/admin/users/:id/reassign.
func (h *Handler) ReassignTasks(c fiber.Ctx) error {
	fromIDStr := utils.CopyString(c.Params("id"))
	fromUserID, err := uuid.Parse(fromIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	var req ReassignRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID целевого пользователя")
	}

	count, err := h.service.ReassignTasks(c.Context(), fromUserID, toUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"success": true, "reassigned_count": count})
}

// DeleteUserGDPR обрабатывает DELETE /api/admin/users/:id/gdpr.
// Опциональный query-параметр reassign_to — UUID пользователя для переназначения задач.
func (h *Handler) DeleteUserGDPR(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	var reassignTo *uuid.UUID
	if reassignStr := c.Query("reassign_to"); reassignStr != "" {
		id, err := uuid.Parse(reassignStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный reassign_to ID")
		}
		reassignTo = &id
	}

	if err := h.service.DeleteUserGDPR(c.Context(), userID, reassignTo); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Данные пользователя удалены (GDPR)"})
}
