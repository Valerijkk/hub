package calls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля голосовых звонков.
type Handler struct {
	service      *CallService
	recSvc       *RecordingService
	scheduledSvc *ScheduledCallService
	cfg          config.LiveKitConfig
	rbacSvc      *rbac.RBACService
}

// NewHandler создаёт новый обработчик звонков.
// rbacSvc используется для проверки прав доступа к записям звонков (CR-02).
// Может быть nil — в этом случае проверки прав не выполняются (только в тестах).
func NewHandler(service *CallService, recSvc *RecordingService, scheduledSvc *ScheduledCallService, cfg config.LiveKitConfig, rbacSvc *rbac.RBACService) *Handler {
	return &Handler{service: service, recSvc: recSvc, scheduledSvc: scheduledSvc, cfg: cfg, rbacSvc: rbacSvc}
}

// RegisterRoutes регистрирует все маршруты звонков под /api/calls.
// authMw — middleware аутентификации. Webhook-эндпоинт регистрируется без аутентификации.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	// Гостевой доступ — без аутентификации
	api.Post("/calls/guest-join", h.GuestJoinCall)

	// Остальные endpoints — с аутентификацией
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Webhook handler
// ---------------------------------------------------------------------------

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
	EgressID    string              `json:"egress_id"`
	FileResults []webhookFileResult `json:"file_results,omitempty"`
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

	// Проверяем API key (issuer).
	// ME-11: НЕ логируем h.cfg.APIKey в открытом виде — это секрет, который
	// попадает в log streams и сторонние log-aggregators. Вместо этого пишем
	// только присланный issuer (привнесённый атакующим) и короткий sha256-
	// префикс ожидаемого ключа — этого достаточно, чтобы оператор сопоставил
	// конфиг с инцидентом, не утекая сам секрет.
	iss, _ := claims["iss"].(string)
	if iss != h.cfg.APIKey {
		expectedSum := sha256.Sum256([]byte(h.cfg.APIKey))
		expectedHash := hex.EncodeToString(expectedSum[:])[:8]
		slog.Error("несовпадение API key в webhook", "got", iss, "expected_hash", expectedHash)
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
