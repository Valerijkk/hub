package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики досок
// ---------------------------------------------------------------------------

// ListBoards — Список досок проекта
// @Summary Список досок проекта
// @Description Возвращает все канбан-доски указанного проекта.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} BoardResponse "Список досок"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/boards [get]
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

// CreateBoard — Создать доску
// @Summary Создать доску
// @Description Создаёт новую канбан-доску в проекте.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateBoardRequest true "Данные доски"
// @Success 201 {object} BoardResponse "Доска создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/boards [post]
func (h *Handler) CreateBoard(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateBoardRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateBoard(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetBoardData — Данные доски
// @Summary Данные доски
// @Description Возвращает колонки и задачи доски с опциональной группировкой по swimlane.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param boardId path string true "ID доски"
// @Param swimlane query string false "Группировка: assignee, priority, type"
// @Success 200 {object} BoardDataResponse "Данные доски"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /projects/{projectId}/boards/{boardId} [get]
func (h *Handler) GetBoardData(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	boardID, err := httputil.ParseUUIDParam(c, "boardId")
	if err != nil {
		return err
	}

	swimlane := c.Query("swimlane") // assignee|priority|type

	resp, err := h.service.GetBoardData(c.Context(), projectID, boardID, swimlane)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
