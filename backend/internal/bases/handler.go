package bases

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики no-code баз (V4-04).
type Handler struct {
	service *Service
}

// NewHandler создаёт bases handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует /api/bases/*.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	g := api.Group("/bases", authMw)
	g.Get("/", h.ListBases)
	g.Post("/", h.CreateBase)
	g.Get("/:baseId", h.GetBase)
	g.Put("/:baseId", h.UpdateBase)
	g.Delete("/:baseId", h.DeleteBase)

	g.Post("/:baseId/tables", h.CreateTable)
	g.Get("/:baseId/tables/:tableId", h.GetTableView)
	g.Post("/:baseId/tables/:tableId/columns", h.CreateColumn)
	g.Post("/:baseId/tables/:tableId/records", h.CreateRecord)
	g.Put("/:baseId/tables/:tableId/records/:recordId", h.UpdateRecord)
	g.Delete("/:baseId/tables/:tableId/records/:recordId", h.DeleteRecord)
}

func (h *Handler) ListBases(c fiber.Ctx) error {
	items, err := h.service.ListBases(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list bases")
	}
	return c.JSON(fiber.Map{"bases": items})
}

func (h *Handler) CreateBase(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateBaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateBase(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) GetBase(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetBase(c.Context(), baseID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) UpdateBase(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	var req UpdateBaseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateBase(c.Context(), baseID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteBase(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteBase(c.Context(), baseID); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CreateTable(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	var req CreateTableRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateTable(c.Context(), baseID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) GetTableView(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	tableID, err := parseTableID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.GetTableView(c.Context(), baseID, tableID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) CreateColumn(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	tableID, err := parseTableID(c)
	if err != nil {
		return err
	}
	var req CreateColumnRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateColumn(c.Context(), baseID, tableID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) CreateRecord(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	tableID, err := parseTableID(c)
	if err != nil {
		return err
	}
	var req CreateRecordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CreateRecord(c.Context(), baseID, tableID, userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) UpdateRecord(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	tableID, err := parseTableID(c)
	if err != nil {
		return err
	}
	recordID, err := parseRecordID(c)
	if err != nil {
		return err
	}
	var req UpdateRecordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.UpdateRecord(c.Context(), baseID, tableID, recordID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteRecord(c fiber.Ctx) error {
	baseID, err := parseBaseID(c)
	if err != nil {
		return err
	}
	tableID, err := parseTableID(c)
	if err != nil {
		return err
	}
	recordID, err := parseRecordID(c)
	if err != nil {
		return err
	}
	if err := h.service.DeleteRecord(c.Context(), baseID, tableID, recordID); err != nil {
		return mapServiceError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseBaseID(c fiber.Ctx) (uuid.UUID, error) {
	return httputil.ParseUUIDParam(c, "baseId")
}

func parseTableID(c fiber.Ctx) (uuid.UUID, error) {
	return httputil.ParseUUIDParam(c, "tableId")
}

func parseRecordID(c fiber.Ctx) (uuid.UUID, error) {
	return httputil.ParseUUIDParam(c, "recordId")
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "not found")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "version conflict")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
}
