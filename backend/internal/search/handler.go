package search

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler обрабатывает HTTP-запросы глобального поиска и управления сохранёнными фильтрами.
type Handler struct {
	svc *SearchService
}

// NewHandler создаёт новый обработчик поиска.
func NewHandler(svc *SearchService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes регистрирует маршруты поиска и сохранённых фильтров.
// LO-08: имя параметра выровнено с остальными модулями (authMw).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	// Fiber v3: inline handler chain api.Get(path, mw, handler) не вызывает middleware.
	// Используем Group для корректного применения auth middleware.
	search := api.Group("/search")
	search.Use(authMw)
	search.Get("", h.search) // GET /api/search?q=...

	filters := api.Group("/filters")
	filters.Use(authMw)
	filters.Post("/", h.createFilter)
	filters.Get("/", h.listFilters)
	filters.Delete("/:id", h.deleteFilter)
	filters.Patch("/:id/share", h.shareFilter)
}

// search godoc
// @Summary Глобальный поиск
// @Description Полнотекстовый поиск по задачам, тест-кейсам и другим сущностям.
// @Tags Поиск
// @Produce json
// @Param q query string true "Поисковый запрос"
// @Param type query string false "Типы сущностей через запятую (issue, test_case)"
// @Param limit query int false "Лимит результатов (по умолчанию 20)"
// @Success 200 {object} SearchResponse "Результаты поиска"
// @Failure 400 {object} fiber.Error "Параметр q обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/search [get]
func (h *Handler) search(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		slog.Error("search: getUserID failed", "error", err, "locals_userID_raw", c.Locals("userID"), "auth_header_present", c.Get("Authorization") != "")
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	q := c.Query("q")
	if q == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр q обязателен")
	}

	var types []string
	if t := c.Query("type"); t != "" {
		types = strings.Split(t, ",")
	}

	limit := 20
	if l, err := strconv.Atoi(c.Query("limit", "20")); err == nil {
		limit = l
	}

	results, err := h.svc.Search(c.Context(), q, userID, types, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка поиска")
	}
	if results == nil {
		results = []SearchResult{}
	}

	return c.JSON(SearchResponse{
		Results: results,
		Total:   len(results),
	})
}

// createFilter godoc
// @Summary Создать сохранённый фильтр
// @Description Создаёт сохранённый фильтр задач для быстрого доступа.
// @Tags Поиск
// @Accept json
// @Produce json
// @Param body body SavedFilterCreateRequest true "Данные фильтра"
// @Success 201 {object} SavedFilterDTO "Фильтр создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/filters [post]
func (h *Handler) createFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	var req SavedFilterCreateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}
	if req.Name == "" || req.ProjectID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name и project_id обязательны")
	}

	filter, err := h.svc.CreateSavedFilter(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(filter)
}

// listFilters godoc
// @Summary Список сохранённых фильтров
// @Description Возвращает сохранённые фильтры пользователя для указанного проекта.
// @Tags Поиск
// @Produce json
// @Param project_id query string true "ID проекта"
// @Success 200 {object} SavedFilterListResponse "Список фильтров"
// @Failure 400 {object} fiber.Error "project_id обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/filters [get]
func (h *Handler) listFilters(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "project_id обязателен")
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project_id")
	}

	filters, err := h.svc.ListSavedFilters(c.Context(), userID, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(SavedFilterListResponse{Filters: filters})
}

// deleteFilter godoc
// @Summary Удалить сохранённый фильтр
// @Description Удаляет сохранённый фильтр пользователя.
// @Tags Поиск
// @Param id path string true "ID фильтра"
// @Success 204 "Фильтр удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/filters/{id} [delete]
func (h *Handler) deleteFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	filterID, err := httputil.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.svc.DeleteSavedFilter(c.Context(), userID, filterID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// shareFilter godoc
// @Summary Поделиться фильтром
// @Description Переключает публичность фильтра — делает его видимым для всех участников проекта.
// @Tags Поиск
// @Accept json
// @Produce json
// @Param id path string true "ID фильтра"
// @Param body body SavedFilterShareRequest true "shared: true/false"
// @Success 200 {object} SavedFilterDTO "Фильтр обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/filters/{id}/share [patch]
func (h *Handler) shareFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}

	filterID, err := httputil.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	var req SavedFilterShareRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
	}

	filter, err := h.svc.ShareFilter(c.Context(), userID, filterID, req.Shared)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(filter)
}

// getUserID извлекает ID пользователя из контекста Fiber (установлен auth middleware).
func getUserID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Locals("userID")
	if raw == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "userID отсутствует в контексте")
	}
	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "userID неизвестного типа")
	}
}
