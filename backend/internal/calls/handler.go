package calls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
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

	// Scheduled calls (CALL-02)
	channels := api.Group("/channels", authMw)
	channels.Get("/:channelId/scheduled-calls", h.ListScheduledCalls)

	scheduledCalls := api.Group("/scheduled-calls", authMw)
	scheduledCalls.Put("/:id", h.UpdateScheduledCall)
	scheduledCalls.Delete("/:id", h.DeleteScheduledCall)

	// Webhook без аутентификации (LiveKit вызывает напрямую)
	api.Post("/calls/webhook", h.HandleWebhook)
}

// userIDFromCtx извлекает user ID из Fiber Locals.
func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "некорректный user ID")
	}
	return id, nil
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

// CreateCall обрабатывает POST /api/calls — создание нового звонка.
func (h *Handler) CreateCall(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// GetActiveCall обрабатывает GET /api/calls/channel/:channelId — получение активного звонка.
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

// JoinCall обрабатывает POST /api/calls/:callId/join — присоединение к звонку.
func (h *Handler) JoinCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
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

// LeaveCall обрабатывает POST /api/calls/:callId/leave — выход из звонка.
func (h *Handler) LeaveCall(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.LeaveCall(c.Context(), callID, userID); err != nil {
		slog.Error("ошибка выхода из звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось покинуть звонок")
	}

	return c.SendStatus(fiber.StatusOK)
}

// EndCall обрабатывает POST /api/calls/:callId/end — завершение звонка.
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

// RaiseHand обрабатывает POST /api/calls/:callId/raise-hand — поднять/опустить руку.
func (h *Handler) RaiseHand(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req RaiseHandRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// SendChatMessage обрабатывает POST /api/calls/:callId/chat — отправка сообщения в чат звонка.
func (h *Handler) SendChatMessage(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// ListChatMessages обрабатывает GET /api/calls/:callId/chat — получение сообщений чата звонка.
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

// StartRecording обрабатывает POST /api/calls/:callId/recording/start — начать запись звонка.
func (h *Handler) StartRecording(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := userIDFromCtx(c)
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

// StopRecording обрабатывает POST /api/calls/:callId/recording/stop — остановить запись звонка.
func (h *Handler) StopRecording(c fiber.Ctx) error {
	// Получаем recording_id из тела запроса
	var req struct {
		RecordingID uuid.UUID `json:"recording_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// ListProjectRecordings обрабатывает GET /api/projects/:projectId/recordings — список записей проекта.
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

// GetRecordingDownloadURL обрабатывает GET /api/recordings/:recordingId/download — скачивание записи.
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

// CreateScheduledCall обрабатывает POST /api/projects/:projectId/scheduled-calls — создание запланированного звонка.
func (h *Handler) CreateScheduledCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("projectId"))
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateScheduledCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.scheduledSvc.CreateScheduledCall(c.Context(), projectID, userID, req)
	if err != nil {
		slog.Error("ошибка создания запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось создать запланированный звонок")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListScheduledCalls обрабатывает GET /api/channels/:channelId/scheduled-calls — список запланированных звонков.
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

// UpdateScheduledCall обрабатывает PUT /api/scheduled-calls/:id — обновление запланированного звонка.
func (h *Handler) UpdateScheduledCall(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID")
	}

	var req UpdateScheduledCallRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.scheduledSvc.UpdateScheduledCall(c.Context(), id, req)
	if err != nil {
		slog.Error("ошибка обновления запланированного звонка", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось обновить запланированный звонок")
	}

	return c.JSON(resp)
}

// DeleteScheduledCall обрабатывает DELETE /api/scheduled-calls/:id — удаление запланированного звонка.
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

// HandleWebhook обрабатывает входящие webhook-события от LiveKit.
// Реализует ручную верификацию JWT из Authorization header + SHA256 hash body.
// Не использует webhook.ReceiveWebhookEvent из-за несовместимости с Fiber/fasthttp.
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
