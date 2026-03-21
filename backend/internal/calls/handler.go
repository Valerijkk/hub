package calls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// Handler предоставляет HTTP-обработчики для модуля голосовых звонков.
type Handler struct {
	service *CallService
	cfg     config.LiveKitConfig
}

// NewHandler создаёт новый обработчик звонков.
func NewHandler(service *CallService, cfg config.LiveKitConfig) *Handler {
	return &Handler{service: service, cfg: cfg}
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

// webhookEvent представляет структуру webhook-события от LiveKit.
type webhookEvent struct {
	Event       string          `json:"event"`
	Room        *webhookRoom    `json:"room,omitempty"`
	Participant *webhookParticipant `json:"participant,omitempty"`
}

// webhookRoom содержит данные комнаты из webhook.
type webhookRoom struct {
	Name string `json:"name"`
}

// webhookParticipant содержит данные участника из webhook.
type webhookParticipant struct {
	Identity string `json:"identity"`
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
