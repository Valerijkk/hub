package calls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler предоставляет HTTP-обработчики для модуля голосовых звонков.
type Handler struct {
	service      *CallService
	recSvc       *RecordingService
	scheduledSvc *ScheduledCallService
	cfg          config.LiveKitConfig
}

// NewHandler создаёт новый обработчик звонков.
func NewHandler(service *CallService, recSvc *RecordingService, scheduledSvc *ScheduledCallService, cfg config.LiveKitConfig) *Handler {
	return &Handler{service: service, recSvc: recSvc, scheduledSvc: scheduledSvc, cfg: cfg}
}

// RegisterRoutes регистрирует все маршруты звонков под /api/calls.
// authMw — middleware аутентификации. Webhook-эндпоинт регистрируется без аутентификации.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	calls := api.Group("/calls", authMw)
	calls.Post("/", h.CreateCall)
	calls.Post("/instant", h.CreateInstantCall)
	calls.Get("/invite/:code", h.JoinByInviteCode)
	calls.Get("/channel/:channelId", h.GetActiveCall)
	calls.Post("/:callId/join", h.JoinCall)
	calls.Post("/:callId/leave", h.LeaveCall)
	calls.Post("/:callId/end", h.EndCall)
	calls.Post("/:callId/raise-hand", h.RaiseHand)
	calls.Post("/:callId/chat", h.SendChatMessage)
	calls.Get("/:callId/chat", h.ListChatMessages)

	// Recording endpoints (CALL-04, CALL-05)
	calls.Post("/:callId/recording/start", h.StartRecording)
	calls.Post("/:callId/recording/stop", h.StopRecording)

	// Project recordings (CALL-05)
	projects := api.Group("/projects", authMw)
	projects.Get("/:projectId/recordings", h.ListProjectRecordings)
	projects.Post("/:projectId/scheduled-calls", h.CreateScheduledCall)

	// Recording download
	recordings := api.Group("/recordings", authMw)
	recordings.Get("/:recordingId/download", h.GetRecordingDownloadURL)

	// Conflict check
	calls.Get("/conflicts", h.CheckConflicts)

	// Scheduled calls (CALL-02)
	channels := api.Group("/channels", authMw)
	channels.Get("/:channelId/scheduled-calls", h.ListScheduledCalls)

	scheduledCalls := api.Group("/scheduled-calls", authMw)
	scheduledCalls.Put("/:id", h.UpdateScheduledCall)
	scheduledCalls.Delete("/:id", h.DeleteScheduledCall)

	// Webhook без аутентификации (LiveKit вызывает напрямую)
	api.Post("/calls/webhook", h.HandleWebhook)
}

// parseCallID извлекает и валидирует callId из параметров пути.
func parseCallID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("callId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "call ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный call ID")
	}
	return id, nil
}

// userNameFromCtx извлекает имя пользователя из Fiber Locals.
func userNameFromCtx(c fiber.Ctx) string {
	name, _ := c.Locals("userName").(string)
	if name == "" {
		name = "user"
	}
	return name
}

// CreateCall godoc
// @Summary Создать звонок
// @Description Создаёт новый голосовой/видеозвонок в указанном канале.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param body body CreateCallRequest true "Данные звонка (channel_id)"
// @Success 201 {object} CallResponse "Звонок создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls [post]
func (h *Handler) CreateCall(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if req.ChannelID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "channel_id обязателен")
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.CreateCall(c.Context(), req.ChannelID, userID, userName)
	if err != nil {
		slog.Error("ошибка создания звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// CreateInstantCall godoc
// @Summary Создать быстрый звонок
// @Description Создаёт мгновенный звонок без привязки к каналу. Генерирует invite_code для ссылки.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param body body object false "Опциональный title"
// @Success 201 {object} CallResponse "Звонок создан"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/instant [post]
func (h *Handler) CreateInstantCall(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Title string `json:"title"`
	}
	_ = c.Bind().JSON(&req)

	userName := userNameFromCtx(c)
	resp, err := h.service.CreateInstantCall(c.Context(), userID, userName, req.Title)
	if err != nil {
		slog.Error("ошибка создания instant звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// JoinByInviteCode godoc
// @Summary Присоединиться по invite-коду
// @Description Присоединяет пользователя к звонку по пригласительному коду.
// @Tags Звонки
// @Produce json
// @Param code path string true "Пригласительный код"
// @Success 200 {object} CallResponse "Данные звонка с токеном"
// @Failure 400 {object} fiber.Error "Код не указан"
// @Failure 404 {object} fiber.Error "Звонок не найден"
// @Security BearerAuth
// @Router /api/calls/invite/{code} [get]
func (h *Handler) JoinByInviteCode(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	code := utils.CopyString(c.Params("code"))
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invite code обязателен")
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.JoinByInviteCode(c.Context(), code, userID, userName)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "звонок не найден или уже завершён")
	}

	return c.JSON(resp)
}

// GetActiveCall godoc
// @Summary Получить активный звонок канала
// @Description Возвращает текущий активный звонок в указанном канале (или null).
// @Tags Звонки
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {object} CallResponse "Активный звонок или null"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/channel/{channelId} [get]
func (h *Handler) GetActiveCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("channelId"))
	channelID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel ID")
	}

	resp, err := h.service.GetActiveCallByChannel(c.Context(), channelID)
	if err != nil {
		slog.Error("ошибка получения активного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка поиска звонка")
	}

	if resp == nil {
		return c.JSON(fiber.Map{"call": nil})
	}

	return c.JSON(resp)
}

// JoinCall godoc
// @Summary Присоединиться к звонку
// @Description Присоединяет текущего пользователя к звонку и возвращает LiveKit токен.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Success 200 {object} JoinCallResponse "Токен для подключения"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/join [post]
func (h *Handler) JoinCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	userName := userNameFromCtx(c)
	resp, err := h.service.JoinCall(c.Context(), callID, userID, userName)
	if err != nil {
		slog.Error("ошибка присоединения к звонку", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось присоединиться к звонку")
	}

	return c.JSON(resp)
}

// LeaveCall godoc
// @Summary Покинуть звонок
// @Description Текущий пользователь покидает звонок.
// @Tags Звонки
// @Param callId path string true "ID звонка"
// @Success 200 "Пользователь покинул звонок"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/leave [post]
func (h *Handler) LeaveCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.LeaveCall(c.Context(), callID, userID); err != nil {
		slog.Error("ошибка выхода из звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось покинуть звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// EndCall godoc
// @Summary Завершить звонок
// @Description Завершает звонок для всех участников.
// @Tags Звонки
// @Param callId path string true "ID звонка"
// @Success 200 "Звонок завершён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/end [post]
func (h *Handler) EndCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}

	if err := h.service.EndCall(c.Context(), callID); err != nil {
		slog.Error("ошибка завершения звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось завершить звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// RaiseHand godoc
// @Summary Поднять/опустить руку
// @Description Переключает состояние "поднятой руки" участника в звонке.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param callId path string true "ID звонка"
// @Param body body RaiseHandRequest true "raised: true/false"
// @Success 200 {object} map[string]bool "raised: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 404 {object} fiber.Error "Звонок не найден"
// @Security BearerAuth
// @Router /api/calls/{callId}/raise-hand [post]
func (h *Handler) RaiseHand(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req RaiseHandRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	callID, err := parseCallID(c)
	if err != nil {
		return err
	}

	callEnt, err := h.service.repo.GetCall(c.Context(), callID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "звонок не найден")
	}

	if err := h.service.SetRaiseHand(c.Context(), callEnt.RoomName, userID.String(), req.Raised); err != nil {
		slog.Error("ошибка поднятия руки", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить статус руки")
	}

	return c.JSON(fiber.Map{"raised": req.Raised})
}

// SendChatMessage godoc
// @Summary Отправить сообщение в чат звонка
// @Description Отправляет текстовое сообщение в чат текущего звонка.
// @Tags Звонки
// @Accept json
// @Produce json
// @Param callId path string true "ID звонка"
// @Param body body object true "text: текст сообщения"
// @Success 201 {object} CallChatMessage "Сообщение отправлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/chat [post]
func (h *Handler) SendChatMessage(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.Text == "" {
		return fiber.NewError(fiber.StatusBadRequest, "текст сообщения обязателен")
	}

	userName := userNameFromCtx(c)
	msg, err := h.service.SendChatMessage(c.Context(), callID.String(), userID.String(), userName, req.Text)
	if err != nil {
		slog.Error("ошибка отправки сообщения чата звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось отправить сообщение")
	}

	return c.Status(fiber.StatusCreated).JSON(msg)
}

// ListChatMessages godoc
// @Summary Список сообщений чата звонка
// @Description Возвращает сообщения чата звонка с пагинацией по лимиту.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Param limit query int false "Лимит сообщений (по умолчанию 50, макс 200)"
// @Success 200 {object} map[string][]CallChatMessage "messages: массив сообщений"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/chat [get]
func (h *Handler) ListChatMessages(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	msgs, err := h.service.ListChatMessages(c.Context(), callID.String(), limit)
	if err != nil {
		slog.Error("ошибка получения сообщений чата звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить сообщения")
	}

	if msgs == nil {
		msgs = []CallChatMessage{}
	}

	return c.JSON(fiber.Map{"messages": msgs})
}

// ---------------------------------------------------------------------------
// Recording handlers (CALL-04, CALL-05)
// ---------------------------------------------------------------------------

// StartRecording godoc
// @Summary Начать запись звонка
// @Description Запускает запись текущего звонка через LiveKit Egress.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Success 201 {object} RecordingResponse "Запись начата"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/recording/start [post]
func (h *Handler) StartRecording(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.recSvc.StartRecording(c.Context(), callID, userID)
	if err != nil {
		slog.Error("ошибка начала записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось начать запись")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StopRecording godoc
// @Summary Остановить запись звонка
// @Description Останавливает активную запись звонка.
// @Tags Звонки
// @Accept json
// @Param callId path string true "ID звонка"
// @Param body body object true "recording_id: ID записи"
// @Success 200 "Запись остановлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/recording/stop [post]
func (h *Handler) StopRecording(c fiber.Ctx) error {
	// Получаем recording_id из тела запроса
	var req struct {
		RecordingID uuid.UUID `json:"recording_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.RecordingID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "recording_id обязателен")
	}

	if err := h.recSvc.StopRecording(c.Context(), req.RecordingID); err != nil {
		slog.Error("ошибка остановки записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось остановить запись")
	}

	return c.SendStatus(fiber.StatusOK)
}

// ListProjectRecordings godoc
// @Summary Список записей проекта
// @Description Возвращает пагинированный список записей звонков проекта.
// @Tags Звонки
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Param limit query int false "Лимит (по умолчанию 20, макс 100)"
// @Success 200 {object} map[string]interface{} "recordings и total"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/projects/{projectId}/recordings [get]
func (h *Handler) ListProjectRecordings(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("projectId"))
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}

	offset := 0
	limit := 20
	if o := c.Query("offset"); o != "" {
		if parsed, parseErr := strconv.Atoi(o); parseErr == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	recordings, total, err := h.recSvc.ListProjectRecordings(c.Context(), projectID, offset, limit)
	if err != nil {
		slog.Error("ошибка получения записей проекта", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить записи")
	}

	return c.JSON(fiber.Map{
		"recordings": recordings,
		"total":      total,
	})
}

// GetRecordingDownloadURL godoc
// @Summary Скачать запись
// @Description Возвращает presigned URL для скачивания записи звонка.
// @Tags Звонки
// @Produce json
// @Param recordingId path string true "ID записи"
// @Success 200 {object} RecordingDownloadResponse "URL для скачивания"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/recordings/{recordingId}/download [get]
func (h *Handler) GetRecordingDownloadURL(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("recordingId"))
	recordingID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный recording ID")
	}

	resp, err := h.recSvc.GetRecordingDownloadURL(c.Context(), recordingID)
	if err != nil {
		slog.Error("ошибка получения URL записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить URL записи")
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// ScheduledCall handlers (CALL-02)
// ---------------------------------------------------------------------------

// CreateScheduledCall godoc
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
		slog.Error("ошибка создания запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать запланированный звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListScheduledCalls godoc
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

	calls, err := h.scheduledSvc.ListScheduledCalls(c.Context(), channelID)
	if err != nil {
		slog.Error("ошибка получения запланированных звонков", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить запланированные звонки")
	}

	return c.JSON(fiber.Map{"scheduled_calls": calls})
}

// UpdateScheduledCall godoc
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

	var req UpdateScheduledCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.scheduledSvc.UpdateScheduledCall(c.Context(), id, req)
	if err != nil {
		slog.Error("ошибка обновления запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить запланированный звонок")
	}

	return c.JSON(resp)
}

// DeleteScheduledCall godoc
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

	if err := h.scheduledSvc.DeleteScheduledCall(c.Context(), id); err != nil {
		slog.Error("ошибка удаления запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось удалить запланированный звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// CheckConflicts godoc
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

	resp, err := h.scheduledSvc.CheckConflicts(c.Context(), requestedTime, duration, participantIDs)
	if err != nil {
		slog.Error("ошибка проверки конфликтов", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось проверить конфликты")
	}

	return c.JSON(resp)
}

// webhookEvent представляет структуру webhook-события от LiveKit.
type webhookEvent struct {
	Event       string              `json:"event"`
	Room        *webhookRoom        `json:"room,omitempty"`
	Participant *webhookParticipant `json:"participant,omitempty"`
	EgressInfo  *webhookEgressInfo  `json:"egress_info,omitempty"`
}

// webhookRoom содержит данные комнаты из webhook.
type webhookRoom struct {
	Name string `json:"name"`
}

// webhookParticipant содержит данные участника из webhook.
type webhookParticipant struct {
	Identity string `json:"identity"`
}

// webhookEgressInfo содержит данные egress из webhook.
type webhookEgressInfo struct {
	EgressID    string               `json:"egress_id"`
	FileResults []webhookFileResult  `json:"file_results,omitempty"`
}

// webhookFileResult содержит данные файла из egress.
type webhookFileResult struct {
	Filename string  `json:"filename"`
	Size     int64   `json:"size"`
	Duration float64 `json:"duration"`
}

// HandleWebhook godoc
// @Summary Webhook от LiveKit
// @Description Обрабатывает входящие события от LiveKit (room_started, participant_joined и др.).
// @Description Верификация JWT из Authorization header + SHA256 hash body.
// @Tags Звонки
// @Accept json
// @Param body body object true "Webhook event от LiveKit"
// @Success 200 "Событие обработано"
// @Failure 400 {object} fiber.Error "Неверное тело"
// @Failure 401 {object} fiber.Error "Невалидный webhook"
// @Router /api/calls/webhook [post]
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
	// Копируем тело запроса (Fiber/fasthttp переиспользует контекст)
	body := make([]byte, len(c.Body()))
	copy(body, c.Body())

	// Верифицируем webhook
	authHeader := utils.CopyString(c.Get("Authorization"))
	if !h.verifyWebhookSignature(authHeader, body) {
		return fiber.NewError(fiber.StatusUnauthorized, "невалидный webhook")
	}

	// Парсим событие
	var event webhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело webhook")
	}

	ctx := c.Context()

	switch event.Event {
	case "room_started":
		if event.Room != nil {
			if err := h.service.OnRoomStarted(ctx, event.Room.Name); err != nil {
				slog.Error("ошибка обработки room_started", "error", err)
			}
		}
	case "room_finished":
		if event.Room != nil {
			if err := h.service.OnRoomFinished(ctx, event.Room.Name); err != nil {
				slog.Error("ошибка обработки room_finished", "error", err)
			}
		}
	case "participant_joined":
		if event.Room != nil && event.Participant != nil {
			if err := h.service.OnParticipantJoined(ctx, event.Room.Name, event.Participant.Identity); err != nil {
				slog.Error("ошибка обработки participant_joined", "error", err)
			}
		}
	case "participant_left":
		if event.Room != nil && event.Participant != nil {
			if err := h.service.OnParticipantLeft(ctx, event.Room.Name, event.Participant.Identity); err != nil {
				slog.Error("ошибка обработки participant_left", "error", err)
			}
		}
	case "egress_ended":
		if event.EgressInfo != nil && h.recSvc != nil {
			var fileSize int64
			var duration float64
			if len(event.EgressInfo.FileResults) > 0 {
				fileSize = event.EgressInfo.FileResults[0].Size
				duration = event.EgressInfo.FileResults[0].Duration
			}
			if err := h.recSvc.OnEgressEnded(ctx, event.EgressInfo.EgressID, fileSize, duration); err != nil {
				slog.Error("ошибка обработки egress_ended", "egress_id", event.EgressInfo.EgressID, "error", err)
			}
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

// verifyWebhookSignature проверяет подпись webhook-запроса от LiveKit.
// LiveKit подписывает webhook JWT токеном, где body hash в claims.
func (h *Handler) verifyWebhookSignature(authHeader string, body []byte) bool {
	if authHeader == "" {
		return false
	}

	// Парсим JWT без верификации чтобы получить claims
	token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
		// Проверяем что используется HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(h.cfg.APISecret), nil
	})
	if err != nil {
		slog.Error("невалидный JWT webhook", "error", err)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return false
	}

	// Проверяем API key (issuer)
	iss, _ := claims["iss"].(string)
	if iss != h.cfg.APIKey {
		slog.Error("несовпадение API key в webhook", "expected", h.cfg.APIKey, "got", iss)
		return false
	}

	// Проверяем SHA256 hash тела запроса
	expectedHash, _ := claims["sha256"].(string)
	if expectedHash != "" {
		hash := sha256.Sum256(body)
		actualHash := base64.StdEncoding.EncodeToString(hash[:])
		if !hmac.Equal([]byte(expectedHash), []byte(actualHash)) {
			slog.Error("несовпадение hash тела webhook")
			return false
		}
	}

	return true
}
