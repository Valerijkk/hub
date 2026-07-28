package audit

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/observability"
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
	admin.Get("/observability/ops", h.ListOperationMetrics)
	admin.Get("/limits", h.GetLimits)
	admin.Put("/limits/:key", h.UpdateLimit)
	admin.Get("/trash", h.ListTrash)
	admin.Post("/trash/:type/:id/restore", h.RestoreItem)
	admin.Post("/users/:id/deactivate", h.DeactivateUser)
	admin.Post("/users/:id/ban", h.BanUser)
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

		isAdmin, err := h.service.IsSystemAdmin(c.Context(), userID)
		if err != nil {
			// P3-BE-08 / W6-A4-011: DB outage → 500; 401 только при отсутствии пользователя.
			if ent.IsNotFound(err) {
				return fiber.NewError(fiber.StatusUnauthorized, "пользователь не найден")
			}
			slog.Error("IsSystemAdmin failed", "error", err, "user_id", userID)
			return fiber.NewError(fiber.StatusInternalServerError, "не удалось проверить права администратора")
		}

		if !isAdmin {
			return fiber.NewError(fiber.StatusForbidden, "доступ только для администратора")
		}

		return c.Next()
	}
}

// ListAuditLogs godoc
// @Summary Журнал аудита
// @Description Возвращает пагинированный список записей аудита с фильтрацией.
// @Tags Аудит
// @Produce json
// @Param action query string false "Фильтр по действию"
// @Param entity_type query string false "Фильтр по типу сущности"
// @Param user_id query string false "Фильтр по ID пользователя"
// @Param project_id query string false "Фильтр по ID проекта"
// @Param date_from query string false "Начало периода (RFC3339)"
// @Param date_to query string false "Конец периода (RFC3339)"
// @Param slow query int false "Только медленные HTTP-запросы (entity_type=http, duration_ms>500); значение 1"
// @Param page query int false "Страница (по умолчанию 1)"
// @Param per_page query int false "Кол-во на странице (по умолчанию 50)"
// @Success 200 {object} AuditLogListResponse "Записи аудита"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/admin/audit [get]
func (h *Handler) ListAuditLogs(c fiber.Ctx) error {
	filters := AuditLogFilters{
		Action:     utils.CopyString(c.Query("action")),
		EntityType: utils.CopyString(c.Query("entity_type")),
		EntityID:   utils.CopyString(c.Query("entity_id")),
		IPAddress:  utils.CopyString(c.Query("ip_address")),
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

	if c.Query("slow") == "1" {
		filters.SlowOnly = true
	}

	filters.Page, _ = strconv.Atoi(c.Query("page", "1"))
	filters.PerPage, _ = strconv.Atoi(c.Query("per_page", "50"))

	resp, err := h.service.ListAuditLogs(c.Context(), filters)
	if err != nil {
		// A2W6-A4-007: не отдаём err.Error() в 500 — детали только в slog.
		slog.Error("ListAuditLogs failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось загрузить журнал аудита")
	}

	return c.JSON(resp)
}

// GetLimits godoc
// @Summary Получить лимиты ресурсов
// @Description Возвращает текущие ограничения ресурсов системы.
// @Tags Аудит
// @Produce json
// @Success 200 {array} ResourceLimitResponse "Лимиты"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/limits [get]
func (h *Handler) GetLimits(c fiber.Ctx) error {
	limits, err := h.service.GetResourceLimits(c.Context())
	if err != nil {
		slog.Error("GetResourceLimits failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось загрузить лимиты")
	}
	return c.JSON(limits)
}

// UpdateLimit godoc
// @Summary Обновить лимит ресурса
// @Description Обновляет значение ограничения ресурса по ключу.
// @Tags Аудит
// @Accept json
// @Produce json
// @Param key path string true "Ключ лимита"
// @Param body body UpdateLimitRequest true "Новое значение"
// @Success 200 {object} map[string]bool "success: true"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/limits/{key} [put]
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
		if errors.Is(err, ErrInvalidResourceLimitKey) {
			return fiber.NewError(fiber.StatusBadRequest, "unknown resource limit key")
		}
		if errors.Is(err, ErrInvalidResourceLimitValue) {
			return fiber.NewError(fiber.StatusBadRequest, "resource limit value must be a positive integer")
		}
		slog.Error("UpdateResourceLimit failed", "error", err, "key", key)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить лимит")
	}

	return c.JSON(fiber.Map{"success": true})
}

// ListTrash godoc
// @Summary Корзина (админ)
// @Description Возвращает удалённые элементы с фильтрацией по типу и проекту.
// @Tags Аудит
// @Produce json
// @Param type query string false "Тип сущности (по умолчанию issue)"
// @Param project_id query string false "ID проекта"
// @Param page query int false "Страница"
// @Param per_page query int false "Кол-во на странице"
// @Success 200 {object} TrashListResponse "Удалённые элементы"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/trash [get]
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
		slog.Error("ListTrash failed", "error", err, "entity_type", entityType)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось загрузить корзину")
	}

	return c.JSON(resp)
}

// RestoreItem godoc
// @Summary Восстановить элемент из корзины
// @Description Восстанавливает удалённый элемент указанного типа.
// @Tags Аудит
// @Produce json
// @Param type path string true "Тип сущности"
// @Param id path string true "ID элемента"
// @Success 200 {object} UserManagementResponse "Элемент восстановлен"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/trash/{type}/{id}/restore [post]
func (h *Handler) RestoreItem(c fiber.Ctx) error {
	entityType := utils.CopyString(c.Params("type"))
	idStr := utils.CopyString(c.Params("id"))

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID")
	}

	if err := h.service.RestoreItem(c.Context(), entityType, entityID); err != nil {
		slog.Error("RestoreItem failed", "error", err, "entity_type", entityType, "entity_id", entityID)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось восстановить элемент")
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Элемент восстановлен"})
}

// DeactivateUser godoc
// @Summary Деактивировать пользователя
// @Description Блокирует учётную запись пользователя.
// @Tags Аудит
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} UserManagementResponse "Пользователь деактивирован"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/users/{id}/deactivate [post]
func (h *Handler) DeactivateUser(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	if err := h.service.DeactivateUser(c.Context(), userID); err != nil {
		if errors.Is(err, ErrCannotModifyFounder) {
			return fiber.NewError(fiber.StatusForbidden, "cannot ban or deactivate founder account")
		}
		slog.Error("DeactivateUser failed", "error", err, "user_id", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось деактивировать пользователя")
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Пользователь деактивирован"})
}

// BanUser godoc
// @Summary Заблокировать пользователя
// @Description Деактивирует учётную запись и фиксирует причину в журнале аудита.
// @Tags Аудит
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param body body BanUserRequest true "Причина блокировки"
// @Success 200 {object} UserManagementResponse "Пользователь заблокирован"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/users/{id}/ban [post]
func (h *Handler) BanUser(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	adminIDStr, ok := c.Locals("userID").(string)
	if !ok || adminIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
	}

	var req BanUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	if err := h.service.BanUser(c.Context(), targetID, adminID, req.Reason, c.IP(), c.Get("User-Agent")); err != nil {
		if errors.Is(err, ErrCannotModifyFounder) {
			return fiber.NewError(fiber.StatusForbidden, "cannot ban or deactivate founder account")
		}
		// Валидация сервиса — безопасные user-facing сообщения; infra — fixed 500.
		msg := err.Error()
		if msg == "укажите причину блокировки" || msg == "нельзя заблокировать самого себя" {
			return fiber.NewError(fiber.StatusBadRequest, msg)
		}
		slog.Error("BanUser failed", "error", err, "target_id", targetID)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось заблокировать пользователя")
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Пользователь заблокирован"})
}

// ReassignTasks godoc
// @Summary Переназначить задачи пользователя
// @Description Переназначает все задачи одного пользователя на другого.
// @Tags Аудит
// @Accept json
// @Produce json
// @Param id path string true "ID исходного пользователя"
// @Param body body ReassignRequest true "ID целевого пользователя"
// @Success 200 {object} map[string]interface{} "success и reassigned_count"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/users/{id}/reassign [post]
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
		slog.Error("ReassignTasks failed", "error", err, "from_user_id", fromUserID, "to_user_id", toUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось переназначить задачи")
	}

	return c.JSON(fiber.Map{"success": true, "reassigned_count": count})
}

// DeleteUserGDPR godoc
// @Summary Удалить данные пользователя (GDPR)
// @Description Полностью удаляет персональные данные пользователя в соответствии с GDPR.
// @Tags Аудит
// @Produce json
// @Param id path string true "ID пользователя"
// @Param reassign_to query string false "ID пользователя для переназначения задач"
// @Success 200 {object} UserManagementResponse "Данные удалены"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/users/{id}/gdpr [delete]
func (h *Handler) DeleteUserGDPR(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID пользователя")
	}

	actorStr, ok := c.Locals("userID").(string)
	if !ok || actorStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	actorID, err := uuid.Parse(actorStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
	}

	var reassignTo *uuid.UUID
	if reassignStr := c.Query("reassign_to"); reassignStr != "" {
		id, err := uuid.Parse(reassignStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный reassign_to ID")
		}
		reassignTo = &id
	}

	if err := h.service.DeleteUserGDPR(c.Context(), actorID, userID, reassignTo); err != nil {
		if errors.Is(err, ErrCannotDeleteFounder) {
			return fiber.NewError(fiber.StatusForbidden, "cannot delete founder account")
		}
		slog.Error("DeleteUserGDPR failed", "error", err, "user_id", userID, "actor_id", actorID)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось удалить данные пользователя")
	}

	return c.JSON(UserManagementResponse{Success: true, Message: "Данные пользователя удалены (GDPR)"})
}

// ListOperationMetrics godoc
// @Summary In-memory метрики Timed()-операций
// @Description Возвращает агрегаты duration/count/error по observability.Timed (OBS-01)
// @Tags Аудит
// @Produce json
// @Success 200 {array} observability.OperationMetric
// @Failure 403 {object} fiber.Error "Только для админа"
// @Security BearerAuth
// @Router /api/admin/observability/ops [get]
func (h *Handler) ListOperationMetrics(c fiber.Ctx) error {
	return c.JSON(observability.Snapshot())
}
