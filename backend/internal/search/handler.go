package search

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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
func (h *Handler) RegisterRoutes(api fiber.Router, requireAuth fiber.Handler) {
	api.Get("/search", requireAuth, h.search)

	filters := api.Group("/filters", requireAuth)
	filters.Post("/", h.createFilter)
	filters.Get("/", h.listFilters)
	filters.Delete("/:id", h.deleteFilter)
	filters.Patch("/:id/share", h.shareFilter)
}

// search обрабатывает GET /api/search?q=...&type=issue,test_case&limit=20.
func (h *Handler) search(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
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

// createFilter обрабатывает POST /api/filters.
func (h *Handler) createFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
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

// listFilters обрабатывает GET /api/filters?project_id=...
func (h *Handler) listFilters(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
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

// deleteFilter обрабатывает DELETE /api/filters/:id.
func (h *Handler) deleteFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	filterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный id фильтра")
	}

	if err := h.svc.DeleteSavedFilter(c.Context(), userID, filterID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// shareFilter обрабатывает PATCH /api/filters/:id/share.
func (h *Handler) shareFilter(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	filterID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный id фильтра")
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
	idStr, ok := c.Locals("user_id").(string)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user_id отсутствует")
	}
	return uuid.Parse(idStr)
}
