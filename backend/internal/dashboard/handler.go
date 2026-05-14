package dashboard

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
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

// Personal godoc
// @Summary Персональный дашборд
// @Description Возвращает агрегированные данные персонального дашборда пользователя.
// @Tags Дашборды
// @Produce json
// @Success 200 {object} PersonalDashboardData "Данные дашборда"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/dashboard/personal [get]
func (h *Handler) Personal(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	data, err := h.service.GetPersonalDashboardData(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки данных дашборда")
	}

	return c.JSON(data)
}

// Project godoc
// @Summary Проектный дашборд
// @Description Возвращает статистику и метрики указанного проекта.
// @Tags Дашборды
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {object} ProjectStatsData "Статистика проекта"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/dashboard/project/{projectId} [get]
func (h *Handler) Project(c fiber.Ctx) error {
	_, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	stats, err := h.service.GetProjectStats(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки статистики проекта")
	}

	return c.JSON(stats)
}

// Admin godoc
// @Summary Админский дашборд
// @Description Возвращает статистику системы: пользователи, проекты, хранилище. Только для админа.
// @Tags Дашборды
// @Produce json
// @Success 200 {object} AdminStatsData "Админ-статистика"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Требуются права администратора"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/dashboard/admin [get]
func (h *Handler) Admin(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	// Проверяем роль администратора через запрос к БД
	u, uErr := h.service.db.User.Get(c.Context(), userID)
	if uErr != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "пользователь не найден")
	}
	if string(u.SystemRole) != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "требуются права администратора")
	}

	stats, err := h.service.GetAdminStats(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка загрузки админ-статистики")
	}

	return c.JSON(stats)
}

// GetLayout godoc
// @Summary Получить раскладку дашборда
// @Description Возвращает сохранённое расположение виджетов дашборда пользователя.
// @Tags Дашборды
// @Produce json
// @Param type query string false "Тип дашборда (по умолчанию personal)"
// @Param project_id query string false "ID проекта (для проектного дашборда)"
// @Success 200 {object} LayoutDTO "Раскладка"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/dashboard/layout [get]
func (h *Handler) GetLayout(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
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

// SaveLayout godoc
// @Summary Сохранить раскладку дашборда
// @Description Сохраняет пользовательское расположение виджетов.
// @Tags Дашборды
// @Accept json
// @Param body body SaveLayoutRequest true "Раскладка виджетов"
// @Success 204 "Раскладка сохранена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/dashboard/layout [put]
func (h *Handler) SaveLayout(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
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

// ListTemplates godoc
// @Summary Список шаблонов дашбордов
// @Description Возвращает доступные шаблоны раскладок дашбордов.
// @Tags Дашборды
// @Produce json
// @Success 200 {array} DashboardTemplate "Шаблоны"
// @Security BearerAuth
// @Router /api/dashboard/templates [get]
func (h *Handler) ListTemplates(c fiber.Ctx) error {
	return c.JSON(h.service.GetTemplates())
}

// applyTemplateRequest — тело запроса на применение шаблона.
type applyTemplateRequest struct {
	ProjectID string `json:"project_id"`
	Template  string `json:"template"`
}

// ApplyTemplate godoc
// @Summary Применить шаблон дашборда
// @Description Применяет предустановленный шаблон раскладки к проектному дашборду.
// @Tags Дашборды
// @Accept json
// @Param body body applyTemplateRequest true "ID проекта и шаблон"
// @Success 204 "Шаблон применён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/dashboard/templates/apply [post]
func (h *Handler) ApplyTemplate(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
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
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}
}
