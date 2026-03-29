package tracker

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Time tracking handlers
// ---------------------------------------------------------------------------

// LogTime godoc
// @Summary Залогировать время
// @Description Записывает затраченное время на задачу.
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body LogTimeRequest true "Данные о времени"
// @Success 201 {object} TimeLogResponse "Время залогировано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/time [post]
func (h *Handler) LogTime(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req LogTimeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.LogTime(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StartTimer godoc
// @Summary Запустить таймер
// @Description Запускает таймер отслеживания времени для задачи.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 201 {object} TimerResponse "Таймер запущен"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/timer/start [post]
func (h *Handler) StartTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StartTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StopTimer godoc
// @Summary Остановить таймер
// @Description Останавливает таймер и записывает затраченное время.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {object} TimeLogResponse "Таймер остановлен, время записано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/timer/stop [post]
func (h *Handler) StopTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StopTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTimeLogs godoc
// @Summary Список записей времени
// @Description Возвращает все записи затраченного времени на задачу.
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} TimeLogResponse "Записи времени"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/time-logs [get]
func (h *Handler) ListTimeLogs(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListTimeLogs(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
