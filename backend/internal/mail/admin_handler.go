package mail

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// AdminHandler — admin API для настроек hub-mail (основной Hub server).
// W7-A1-005: без *ent.Client — user-list через SettingsStore (ActiveUserEmailSource).
type AdminHandler struct {
	svc       *AdminService
	store     SettingsStore
	sender    BroadcastSender
	isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)
}

// NewAdminHandler создаёт admin handler.
// store должен реализовывать ActiveUserEmailSource для Broadcast (EntSettingsStore).
func NewAdminHandler(svc *AdminService, store SettingsStore, sender BroadcastSender, isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)) *AdminHandler {
	return &AdminHandler{svc: svc, store: store, sender: sender, isAdminFn: isAdminFn}
}

// RegisterRoutes регистрирует /admin/mail/* под auth + admin middleware.
func (h *AdminHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	admin := api.Group("/admin/mail", authMw, h.requireAdmin())
	admin.Get("/settings", h.GetSettings)
	admin.Put("/settings", h.UpdateSettings)
	admin.Post("/broadcast", h.Broadcast)
}

func (h *AdminHandler) requireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user id")
		}
		isAdmin, err := h.isAdminFn(c, userID)
		if err != nil || !isAdmin {
			return fiber.NewError(fiber.StatusForbidden, "admin only")
		}
		return c.Next()
	}
}

// GetSettings GET /admin/mail/settings
// @Summary Получить настройки hub-mail
// @Description Возвращает URL hub-mail, from address и лимиты (Phase 15.2, MAIL-01)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SettingsResponse "Настройки почты"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Admin only"
// @Router /admin/mail/settings [get]
func (h *AdminHandler) GetSettings(c fiber.Ctx) error {
	resp, err := h.svc.GetSettings(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load settings")
	}
	return c.JSON(resp)
}

// UpdateSettings PUT /admin/mail/settings
// @Summary Обновить настройки hub-mail
// @Description Сохраняет hub_mail_url, from_alerts, alerts_enabled, daily_limit (Phase 15.2)
// @Tags Администрирование
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateSettingsRequest true "Настройки"
// @Success 200 {object} SettingsResponse "Обновлённые настройки"
// @Failure 400 {object} map[string]string "invalid body"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Admin only"
// @Router /admin/mail/settings [put]
func (h *AdminHandler) UpdateSettings(c fiber.Ctx) error {
	var req UpdateSettingsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.UpdateSettings(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save settings")
	}
	return c.JSON(resp)
}

// Broadcast POST /admin/mail/broadcast
// @Summary Admin email broadcast
// @Tags Администрирование
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body BroadcastRequest true "Broadcast"
// @Success 200 {object} BroadcastResponse
// @Router /admin/mail/broadcast [post]
func (h *AdminHandler) Broadcast(c fiber.Ctx) error {
	var req BroadcastRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	users, ok := h.store.(ActiveUserEmailSource)
	if !ok || users == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "broadcast user source unavailable")
	}
	resp, err := h.svc.Broadcast(c.Context(), users, h.sender, req)
	if err != nil {
		if errors.Is(err, ErrInvalidRequest) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "broadcast failed")
	}
	return c.JSON(resp)
}
