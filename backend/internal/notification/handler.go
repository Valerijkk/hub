package notification

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Handler предоставляет HTTP-обработчики для модуля уведомлений.
type Handler struct {
	service *NotificationService
}

// NewHandler создаёт новый обработчик уведомлений.
func NewHandler(service *NotificationService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты уведомлений под /api/notifications.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	notifs := api.Group("/notifications", authMw)

	notifs.Get("/", h.List)
	notifs.Get("/unread-count", h.UnreadCount)
	notifs.Patch("/:id/read", h.MarkRead)
	notifs.Patch("/read-all", h.MarkAllRead)
	notifs.Get("/preferences", h.GetPreferences)
	notifs.Put("/preferences", h.UpdatePreference)
}

// List возвращает пагинированный список уведомлений пользователя.
// Query params: page, per_page, is_read (optional).
func (h *Handler) List(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var isRead *bool
	if v := c.Query("is_read"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			isRead = &b
		}
	}

	result, err := h.service.ListNotifications(c.Context(), userID, page, perPage, isRead)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения уведомлений")
	}

	return c.JSON(result)
}

// UnreadCount возвращает количество непрочитанных уведомлений.
func (h *Handler) UnreadCount(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	count, err := h.service.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка подсчёта уведомлений")
	}

	return c.JSON(UnreadCountResponse{Count: count})
}

// MarkRead помечает одно уведомление как прочитанное.
func (h *Handler) MarkRead(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный ID уведомления")
	}

	if err := h.service.MarkRead(c.Context(), notifID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка отметки прочитанным")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MarkAllRead помечает все уведомления пользователя как прочитанные.
func (h *Handler) MarkAllRead(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	if err := h.service.MarkAllRead(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка отметки всех прочитанными")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPreferences возвращает настройки уведомлений пользователя.
func (h *Handler) GetPreferences(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	prefs, err := h.service.GetPreferences(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения настроек")
	}

	return c.JSON(prefs)
}

// UpdatePreference создаёт или обновляет настройку уведомлений.
func (h *Handler) UpdatePreference(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req PreferenceUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.NotificationType == "" || req.Channel == "" {
		return fiber.NewError(fiber.StatusBadRequest, "notification_type и channel обязательны")
	}

	pref, err := h.service.UpdatePreference(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка обновления настройки")
	}

	return c.JSON(pref)
}

// getUserID извлекает ID пользователя из контекста Fiber.
func getUserID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Locals("userID")
	if raw == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}
}
