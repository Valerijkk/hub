package crm

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики CRM-lite.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler CRM.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты /api/crm.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	crmGroup := api.Group("/crm", authMw)
	crmGroup.Get("/contacts", h.ListContacts)
	crmGroup.Post("/contacts", h.CreateContact)
	crmGroup.Put("/contacts/:contactId", h.UpdateContact)
	crmGroup.Delete("/contacts/:contactId", h.DeleteContact)

	crmGroup.Get("/pipeline", h.GetPipeline)
	crmGroup.Post("/deals", h.CreateDeal)
	crmGroup.Put("/deals/:dealId", h.UpdateDeal)
	crmGroup.Delete("/deals/:dealId", h.DeleteDeal)
}

func (h *Handler) ListContacts(c fiber.Ctx) error {
	items, err := h.service.ListContacts(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list contacts")
	}
	return c.JSON(fiber.Map{"contacts": items})
}

func (h *Handler) CreateContact(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateContactRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateContact(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) UpdateContact(c fiber.Ctx) error {
	id, err := parseContactID(c)
	if err != nil {
		return err
	}
	var req UpdateContactRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateContact(c.Context(), id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteContact(c fiber.Ctx) error {
	id, err := parseContactID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteContact(c.Context(), id); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetPipeline(c fiber.Ctx) error {
	resp, err := h.service.GetPipeline(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get pipeline")
	}
	return c.JSON(resp)
}

func (h *Handler) CreateDeal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateDealRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateDeal(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) UpdateDeal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := parseDealID(c)
	if err != nil {
		return err
	}
	var req UpdateDealRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateDeal(c.Context(), userID, id, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteDeal(c fiber.Ctx) error {
	id, err := parseDealID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteDeal(c.Context(), id); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseContactID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("contactId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid contact id")
	}
	return id, nil
}

func parseDealID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("dealId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid deal id")
	}
	return id, nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "crm entity not found")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid crm input")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "crm version conflict")
	case errors.Is(err, ErrProjectForbidden):
		return fiber.NewError(fiber.StatusForbidden, "project access required")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "crm operation failed")
	}
}
