package accounting

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики бухгалтерии (V4-03, только админ).
type Handler struct {
	service *Service
}

// NewHandler создаёт handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует /api/accounting/*.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	g := api.Group("/accounting", authMw)
	g.Get("/accounts", h.ListAccounts)
	g.Post("/accounts", h.CreateAccount)
	g.Get("/journal", h.ListJournal)
	g.Post("/journal", h.CreateJournal)
	g.Post("/journal/:entryId/post", h.PostJournal)
	g.Get("/trial-balance", h.TrialBalance)
	g.Get("/invoices", h.ListInvoices)
	g.Post("/invoices", h.CreateInvoice)
	g.Post("/invoices/:invoiceId/issue", h.IssueInvoice)
	g.Post("/invoices/:invoiceId/pay", h.MarkInvoicePaid)
	g.Get("/income-statement", h.IncomeStatement)
}

func (h *Handler) ListAccounts(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListAccounts(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"accounts": items})
}

func (h *Handler) CreateAccount(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateAccountRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.service.CreateAccount(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) ListJournal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListJournal(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"entries": items})
}

func (h *Handler) CreateJournal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateJournalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.service.CreateJournal(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) PostJournal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	entryID, err := httputil.ParseUUIDParam(c, "entryId")
	if err != nil {
		return err
	}
	resp, err := h.service.PostJournal(c.Context(), userID, entryID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) TrialBalance(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	rows, err := h.service.TrialBalance(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"rows": rows})
}

func (h *Handler) ListInvoices(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListInvoices(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"invoices": items})
}

func (h *Handler) CreateInvoice(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateInvoiceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.service.CreateInvoice(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) IssueInvoice(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	invoiceID, err := httputil.ParseUUIDParam(c, "invoiceId")
	if err != nil {
		return err
	}
	resp, err := h.service.IssueInvoice(c.Context(), userID, invoiceID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) MarkInvoicePaid(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	invoiceID, err := httputil.ParseUUIDParam(c, "invoiceId")
	if err != nil {
		return err
	}
	resp, err := h.service.MarkInvoicePaid(c.Context(), userID, invoiceID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) IncomeStatement(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.service.IncomeStatement(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "not found")
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "admin only")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	case errors.Is(err, ErrUnbalanced):
		return fiber.NewError(fiber.StatusBadRequest, "journal not balanced")
	case errors.Is(err, ErrAlreadyPosted):
		return fiber.NewError(fiber.StatusConflict, "already posted")
	case errors.Is(err, ErrInvoiceExists):
		return fiber.NewError(fiber.StatusConflict, "invoice exists for deal")
	case errors.Is(err, ErrInvalidInvoice):
		return fiber.NewError(fiber.StatusBadRequest, "invalid invoice state")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
}
