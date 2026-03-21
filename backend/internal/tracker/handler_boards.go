package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Board handlers
// ---------------------------------------------------------------------------

// ListBoards обрабатывает GET /api/projects/:projectId/boards
func (h *Handler) ListBoards(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListBoards(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateBoard обрабатывает POST /api/projects/:projectId/boards
func (h *Handler) CreateBoard(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateBoardRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateBoard(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetBoardData обрабатывает GET /api/projects/:projectId/boards/:boardId
func (h *Handler) GetBoardData(c fiber.Ctx) error {
	boardID, err := httputil.ParseUUIDParam(c, "boardId")
	if err != nil {
		return err
	}

	swimlane := c.Query("swimlane") // assignee|priority|type

	resp, err := h.service.GetBoardData(c.Context(), boardID, swimlane)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
