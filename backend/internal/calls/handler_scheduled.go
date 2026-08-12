package calls

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики запланированных звонков (CALL-02)
// ---------------------------------------------------------------------------

// CreateScheduledCall — Создать запланированный звонок
// @Summary Создать запланированный звонок
// @Description Создаёт запланированный звонок в проекте с указанным временем и участниками.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateScheduledCallRequest true "Данные звонка"
// @Success 201 {object} ScheduledCallResponse "Звонок запланирован"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/projects/{projectId}/scheduled-calls [post]
func (h *Handler) CreateScheduledCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("projectId"))
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateScheduledCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.scheduledSvc.CreateScheduledCall(c.Context(), projectID, userID, req)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для создания запланированного звонка")
		}
		slog.Error("ошибка создания запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать запланированный звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListScheduledCalls — Список запланированных звонков
// @Summary Список запланированных звонков
// @Description Возвращает запланированные звонки для указанного канала.
// @Tags Звонки
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {object} map[string][]ScheduledCallResponse "scheduled_calls"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/channels/{channelId}/scheduled-calls [get]
func (h *Handler) ListScheduledCalls(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("channelId"))
	channelID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	calls, err := h.scheduledSvc.ListScheduledCalls(c.Context(), channelID, userID)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для просмотра запланированных звонков")
		}
		slog.Error("ошибка получения запланированных звонков", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить запланированные звонки")
	}

	return c.JSON(fiber.Map{"scheduled_calls": calls})
}

// UpdateScheduledCall — Обновить запланированный звонок
// @Summary Обновить запланированный звонок
// @Description Обновляет время, название или участников запланированного звонка.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param id path string true "ID запланированного звонка"
// @Param body body UpdateScheduledCallRequest true "Обновляемые данные"
// @Success 200 {object} ScheduledCallResponse "Звонок обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/scheduled-calls/{id} [put]
func (h *Handler) UpdateScheduledCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req UpdateScheduledCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.scheduledSvc.UpdateScheduledCall(c.Context(), id, userID, req)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для изменения запланированного звонка")
		}
		if errors.Is(err, ErrCallNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "запланированный звонок не найден")
		}
		slog.Error("ошибка обновления запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить запланированный звонок")
	}

	return c.JSON(resp)
}

// DeleteScheduledCall — Удалить запланированный звонок
// @Summary Удалить запланированный звонок
// @Description Удаляет запланированный звонок.
// @Tags Звонки
// @Param id path string true "ID запланированного звонка"
// @Success 200 "Звонок удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/scheduled-calls/{id} [delete]
func (h *Handler) DeleteScheduledCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.scheduledSvc.DeleteScheduledCall(c.Context(), id, userID); err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для удаления запланированного звонка")
		}
		if errors.Is(err, ErrCallNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "запланированный звонок не найден")
		}
		slog.Error("ошибка удаления запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось удалить запланированный звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// CheckConflicts — Проверить конфликты расписания
// @Summary Проверить конфликты расписания
// @Description Проверяет, есть ли пересечения в расписании участников на указанное время.
// @Tags Звонки
// @Produce json
// @Param time query string true "Время начала (ISO 8601)"
// @Param duration query int false "Длительность в минутах (по умолчанию 60)"
// @Param participant_ids query string false "ID участников через запятую"
// @Success 200 {object} ConflictResponse "Результат проверки конфликтов"
// @Failure 400 {object} fiber.Error "Неверные параметры"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/conflicts [get]
func (h *Handler) CheckConflicts(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	timeStr := c.Query("time")
	if timeStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр time обязателен")
	}
	requestedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный формат time (ожидается ISO 8601)")
	}

	durationStr := c.Query("duration", "60")
	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный duration")
	}

	participantIDsStr := c.Query("participant_ids")
	if participantIDsStr == "" {
		return c.JSON(ConflictResponse{HasConflict: false, Conflicts: []ConflictDetail{}})
	}

	idParts := strings.Split(participantIDsStr, ",")
	participantIDs := make([]uuid.UUID, 0, len(idParts))
	for _, idStr := range idParts {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			continue
		}
		participantIDs = append(participantIDs, id)
	}

	if len(participantIDs) == 0 {
		return c.JSON(ConflictResponse{HasConflict: false, Conflicts: []ConflictDetail{}})
	}

	resp, err := h.scheduledSvc.CheckConflicts(c.Context(), userID, requestedTime, duration, participantIDs)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "нет доступа к календарю участника")
		}
		slog.Error("ошибка проверки конфликтов", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось проверить конфликты")
	}

	return c.JSON(resp)
}
