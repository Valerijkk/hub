package marketplace

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики marketplace.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler marketplace.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты /api/marketplace.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	mp := api.Group("/marketplace", authMw)
	listings := mp.Group("/listings")
	listings.Get("/", h.ListApproved)
	listings.Get("/mine", h.ListMine)
	listings.Get("/pending", h.ListPending)
	listings.Post("/", h.CreateListing)
	listings.Put("/:listingId", h.UpdateListing)
	listings.Post("/:listingId/approve", h.ApproveListing)
	listings.Post("/:listingId/reject", h.RejectListing)
	listings.Post("/:listingId/sold", h.MarkSold)
	listings.Post("/:listingId/withdraw", h.WithdrawListing)
	listings.Delete("/:listingId", h.DeleteListing)
}

func (h *Handler) ListApproved(c fiber.Ctx) error {
	page, perPage := parsePagination(c)
	category := c.Query("category")
	resp, err := h.service.ListApproved(c.Context(), category, page, perPage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list listings")
	}
	return c.JSON(resp)
}

func (h *Handler) ListMine(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListMine(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list my listings")
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) ListPending(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListPending(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) CreateListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateListing(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) UpdateListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	var req UpdateListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateListing(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) ApproveListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	var req ReviewListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.ApproveListing(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) RejectListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	var req ReviewListingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.RejectListing(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) MarkSold(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	var req VersionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.MarkSold(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) WithdrawListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	var req VersionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.WithdrawListing(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteListing(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseListingID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteListing(c.Context(), userID, id); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseListingID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("listingId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid listing id")
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
		return fiber.NewError(fiber.StatusNotFound, "listing not found")
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	case errors.Is(err, ErrNotOwner):
		return fiber.NewError(fiber.StatusForbidden, "not listing owner")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing input")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "listing version conflict")
	case errors.Is(err, ErrInvalidStatus):
		return fiber.NewError(fiber.StatusBadRequest, "invalid listing status")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "marketplace operation failed")
	}
}
