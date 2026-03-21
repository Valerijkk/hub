package dashboard

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля дашбордов.
type Handler struct {
	service *DashboardService
}

// NewHandler создаёт новый обработчик дашбордов.
func NewHandler(service *DashboardService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты дашбордов под /api/dashboard.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	dash := api.Group("/dashboard", authMw)

	dash.Get("/personal", h.Personal)
	dash.Get("/project/:projectId", h.Project)
	dash.Get("/admin", h.Admin)
	dash.Get("/layout", h.GetLayout)
	dash.Put("/layout", h.SaveLayout)
	dash.Get("/templates", h.ListTemplates)
	dash.Post("/templates/apply", h.ApplyTemplate)
}

// Personal возвращает агрегированные данные персонального дашборда.
func (h *Handler) Personal(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	data, err := h.service.GetPersonalDashboardData(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки данных дашборда")
	}

	return c.JSON(data)
}

// Project возвращает статистику проектного дашборда.
func (h *Handler) Project(c fiber.Ctx) error {
	_, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный ID проекта")
	}

	stats, err := h.service.GetProjectStats(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки статистики проекта")
	}

	return c.JSON(stats)
}

// Admin возвращает статистику админского дашборда.
func (h *Handler) Admin(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	// Проверяем роль администратора
	role := c.Locals("systemRole")
	if role == nil || role != "admin" {
		_ = userID // userID используется для контекста
		return fiber.NewError(fiber.StatusForbidden, "требуются права администратора")
	}

	stats, err := h.service.GetAdminStats(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки админ-статистики")
	}

	return c.JSON(stats)
}

// GetLayout возвращает сохранённую раскладку дашборда.
func (h *Handler) GetLayout(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	dashType := c.Query("type", "personal")
	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "невалидный project_id")
		}
		projectID = &pid
	}

	layout, err := h.service.GetLayout(c.Context(), userID, dashType, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки раскладки")
	}

	if layout == nil {
		return c.JSON(LayoutDTO{
			DashboardType: dashType,
			ProjectID:     projectID,
			Items:         []LayoutItem{},
		})
	}

	return c.JSON(layout)
}

// SaveLayout сохраняет раскладку дашборда.
func (h *Handler) SaveLayout(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req SaveLayoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.DashboardType == "" {
		return fiber.NewError(fiber.StatusBadRequest, "dashboard_type обязателен")
	}

	if err := h.service.SaveLayout(c.Context(), userID, req); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка сохранения раскладки")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListTemplates возвращает список доступных шаблонов дашбордов.
func (h *Handler) ListTemplates(c fiber.Ctx) error {
	return c.JSON(h.service.GetTemplates())
}

// applyTemplateRequest — тело запроса на применение шаблона.
type applyTemplateRequest struct {
	ProjectID string `json:"project_id"`
	Template  string `json:"template"`
}

// ApplyTemplate применяет выбранный шаблон дашборда.
func (h *Handler) ApplyTemplate(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req applyTemplateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.Template == "" {
		return fiber.NewError(fiber.StatusBadRequest, "template обязателен")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный project_id")
	}

	if err := h.service.ApplyTemplate(c.Context(), userID, projectID, req.Template); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
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
