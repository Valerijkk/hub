package intranet

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики intranet.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler intranet.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты /api/intranet.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	intra := api.Group("/intranet", authMw)
	intra.Get("/directory", h.ListDirectory)
	intra.Get("/directory/users/:userId", h.GetDirectoryUser)
	kudos := intra.Group("/kudos")
	kudos.Get("/feed", h.ListKudosFeed)
	kudos.Get("/received", h.ListMyReceivedKudos)
	kudos.Post("/", h.CreateKudo)
	coffee := intra.Group("/coffee")
	coffee.Get("/enrollment", h.GetCoffeeEnrollment)
	coffee.Put("/enrollment", h.SetCoffeeEnrollment)
	coffee.Get("/match", h.GetMyCoffeeMatch)
	coffee.Get("/stats", h.GetCoffeeStats)
	coffee.Get("/rounds", h.ListCoffeeRounds)
	coffee.Post("/run", h.RunCoffeeMatching)
	news := intra.Group("/news")
	news.Get("/", h.ListNews)
	news.Get("/admin", h.ListNewsAdmin)
	news.Post("/", h.CreateNews)
	news.Get("/:newsId", h.GetNews)
	news.Put("/:newsId", h.UpdateNews)
	news.Delete("/:newsId", h.DeleteNews)
}

// ListDirectory godoc
// @Summary Справочник сотрудников
// @Description Дерево отделов с карточками активных сотрудников (MOD-02)
// @Tags Intranet
// @Produce json
// @Success 200 {object} DirectoryResponse
// @Security BearerAuth
// @Router /api/intranet/directory [get]
func (h *Handler) ListDirectory(c fiber.Ctx) error {
	resp, err := h.service.ListDirectory(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list directory")
	}
	return c.JSON(resp)
}

// GetDirectoryUser godoc
// @Summary Карточка сотрудника
// @Tags Intranet
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} DirectoryUserResponse
// @Security BearerAuth
// @Router /api/intranet/directory/users/{userId} [get]
func (h *Handler) GetDirectoryUser(c fiber.Ctx) error {
	id, err := parseUserID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetDirectoryUser(c.Context(), id)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

// ListNews godoc
// @Summary Лента новостей
// @Description Опубликованные новости компании (pinned first)
// @Tags Intranet
// @Produce json
// @Param page query int false "Страница"
// @Param per_page query int false "На странице"
// @Success 200 {object} NewsListResponse
// @Security BearerAuth
// @Router /api/intranet/news [get]
func (h *Handler) ListNews(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	resp, err := h.service.ListNews(c.Context(), page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list news")
	}
	return c.JSON(resp)
}

// ListNewsAdmin godoc
// @Summary Все новости (admin)
// @Tags Intranet
// @Produce json
// @Success 200 {object} NewsListResponse
// @Security BearerAuth
// @Router /api/intranet/news/admin [get]
func (h *Handler) ListNewsAdmin(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	page, perPage := parsePagination(c)
	resp, err := h.service.ListNewsAdmin(c.Context(), userID, page, perPage)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

// GetNews godoc
// @Summary Получить новость
// @Tags Intranet
// @Produce json
// @Param newsId path string true "News ID"
// @Success 200 {object} NewsPostResponse
// @Security BearerAuth
// @Router /api/intranet/news/{newsId} [get]
func (h *Handler) GetNews(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseNewsID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetNews(c.Context(), userID, id, false)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

// CreateNews godoc
// @Summary Создать новость (admin)
// @Tags Intranet
// @Accept json
// @Produce json
// @Param body body CreateNewsRequest true "News"
// @Success 201 {object} NewsPostResponse
// @Security BearerAuth
// @Router /api/intranet/news [post]
func (h *Handler) CreateNews(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateNewsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateNews(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateNews godoc
// @Summary Обновить новость (admin)
// @Tags Intranet
// @Accept json
// @Produce json
// @Param newsId path string true "News ID"
// @Param body body UpdateNewsRequest true "News"
// @Success 200 {object} NewsPostResponse
// @Security BearerAuth
// @Router /api/intranet/news/{newsId} [put]
func (h *Handler) UpdateNews(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseNewsID(c)
	if err != nil {
		return err
	}
	var req UpdateNewsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateNews(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

// DeleteNews godoc
// @Summary Удалить новость (admin, soft)
// @Tags Intranet
// @Param newsId path string true "News ID"
// @Success 204
// @Security BearerAuth
// @Router /api/intranet/news/{newsId} [delete]
func (h *Handler) DeleteNews(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseNewsID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteNews(c.Context(), userID, id); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListKudosFeed godoc
// @Summary Публичная лента kudos
// @Tags Intranet
// @Produce json
// @Success 200 {object} KudoListResponse
// @Security BearerAuth
// @Router /api/intranet/kudos/feed [get]
func (h *Handler) ListKudosFeed(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	resp, err := h.service.ListKudosFeed(c.Context(), page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list kudos feed")
	}
	return c.JSON(resp)
}

// ListMyReceivedKudos godoc
// @Summary Мои полученные kudos
// @Tags Intranet
// @Produce json
// @Success 200 {object} KudoListResponse
// @Security BearerAuth
// @Router /api/intranet/kudos/received [get]
func (h *Handler) ListMyReceivedKudos(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	page, perPage := parsePagination(c)
	resp, err := h.service.ListMyReceivedKudos(c.Context(), userID, page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list received kudos")
	}
	return c.JSON(resp)
}

// CreateKudo godoc
// @Summary Отправить kudos коллеге
// @Tags Intranet
// @Accept json
// @Produce json
// @Param body body CreateKudoRequest true "Kudos"
// @Success 201 {object} KudoResponse
// @Security BearerAuth
// @Router /api/intranet/kudos [post]
func (h *Handler) CreateKudo(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateKudoRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateKudo(c.Context(), userID, req)
	if err != nil {
		return mapKudoServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) GetCoffeeEnrollment(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetCoffeeEnrollment(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get coffee enrollment")
	}
	return c.JSON(resp)
}

func (h *Handler) SetCoffeeEnrollment(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req struct {
		Active bool `json:"active"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.SetCoffeeEnrollment(c.Context(), userID, req.Active)
	if err != nil {
		return mapCoffeeServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) GetMyCoffeeMatch(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetMyCoffeeMatch(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get coffee match")
	}
	return c.JSON(resp)
}

func (h *Handler) GetCoffeeStats(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetCoffeeStats(c.Context(), userID)
	if err != nil {
		return mapCoffeeServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) ListCoffeeRounds(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListCoffeeRounds(c.Context(), userID)
	if err != nil {
		return mapCoffeeServiceError(err)
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) RunCoffeeMatching(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.RunCoffeeMatching(c.Context(), userID)
	if err != nil {
		return mapCoffeeServiceError(err)
	}
	return c.JSON(resp)
}

func parseNewsID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("newsId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid news id")
	}
	return id, nil
}

func parseUserID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("userId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}
	return id, nil
}

func parsePagination(c fiber.Ctx) (page, perPage int) {
	page = 1
	perPage = 25
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}
	return page, perPage
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "news post not found")
	case errors.Is(err, ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid news post input")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "news post version conflict")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "intranet operation failed")
	}
}

func mapKudoServiceError(err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, "recipient not found")
	case errors.Is(err, ErrSelfKudo):
		return fiber.NewError(fiber.StatusBadRequest, "cannot send kudos to yourself")
	case errors.Is(err, ErrInvalidKudo):
		return fiber.NewError(fiber.StatusBadRequest, "invalid kudos message")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "kudos operation failed")
	}
}

func mapCoffeeServiceError(err error) error {
	switch {
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	case errors.Is(err, ErrCoffeeNotEnough):
		return fiber.NewError(fiber.StatusBadRequest, "not enough participants")
	case errors.Is(err, ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "coffee operation failed")
	}
}
