package ai

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Handler — пользовательские AI routes.
type Handler struct {
	svc       *Service
	botUserID uuid.UUID
}

// NewHandler создаёт user AI handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SetBotUserID — system user id @hub-ai для GET /ai/bot.
func (h *Handler) SetBotUserID(id uuid.UUID) {
	h.botUserID = id
}

// RegisterRoutes монтирует /ai/* под auth API.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rateLimitMw ...fiber.Handler) {
	ai := api.Group("/ai", authMw)
	for _, mw := range rateLimitMw {
		ai.Use(mw)
	}
	ai.Get("/models", h.ListModels)
	ai.Get("/usage/today", h.GetUsageToday)
	ai.Get("/preferences", h.GetPreferences)
	ai.Patch("/preferences", h.UpdatePreferences)
	ai.Get("/bot", h.GetBotInfo)
	ai.Post("/jql/assist", h.AssistJQL)
	ai.Post("/composer/suggest", h.SuggestComposer)
	ai.Post("/v1/chat/completions", h.ChatCompletions)
}

// GetBotInfo GET /api/ai/bot — системный пользователь @hub-ai.
func (h *Handler) GetBotInfo(c fiber.Ctx) error {
	if h.botUserID == uuid.Nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "bot not configured")
	}
	return c.JSON(BotInfoResponse{
		UserID:      h.botUserID.String(),
		DisplayName: BotDisplayName,
		Email:       BotEmail,
	})
}

// ListModels GET /api/ai/models
func (h *Handler) ListModels(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.svc.ListModels(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list models")
	}
	return c.JSON(resp)
}

// GetPreferences GET /api/ai/preferences
func (h *Handler) GetPreferences(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.svc.GetUserPreference(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load preferences")
	}
	return c.JSON(resp)
}

// UpdatePreferences PATCH /api/ai/preferences
func (h *Handler) UpdatePreferences(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var req UpdateUserPreferenceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.SaveUserPreference(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrModelNotAllowed) {
			return fiber.NewError(fiber.StatusBadRequest, "model not available")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save preferences")
	}
	return c.JSON(resp)
}

// GetUsageToday GET /api/ai/usage/today
func (h *Handler) GetUsageToday(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	resp, err := h.svc.GetUsageToday(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load usage")
	}
	return c.JSON(resp)
}

// AssistJQL POST /api/ai/jql/assist
func (h *Handler) AssistJQL(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var req JQLAssistRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.AssistJQL(c.Context(), userID, req)
	if err != nil {
		// A2W5-A1-001 / TASK-085: те же sentinel, что SuggestComposer (не всё → 502).
		if errors.Is(err, ErrNoProvider) || errors.Is(err, ErrProviderDisabled) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrProviderRPMUnavailable) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrDailyQuotaExceeded) {
			c.Set("Retry-After", secondsUntilMidnightUTC())
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrProviderRateLimited) {
			c.Set("Retry-After", "60")
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrExternalAIConsentRequired) {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
		if errors.Is(err, ErrAllProvidersFailed) {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		if errors.Is(err, ErrContentPolicyViolation) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		// A3W5-1-002: пустой query — 400, не catch-all 502.
		if errors.Is(err, ErrQueryRequired) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(resp)
}

// SuggestComposer POST /api/ai/composer/suggest — Tab-автодополнение (M3-024).
func (h *Handler) SuggestComposer(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var req ComposerSuggestRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.SuggestComposer(c.Context(), userID, req, c.IP(), string(c.Request().Header.UserAgent()))
	if err != nil {
		if errors.Is(err, ErrNoProvider) || errors.Is(err, ErrProviderDisabled) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrProviderRPMUnavailable) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrDailyQuotaExceeded) {
			c.Set("Retry-After", secondsUntilMidnightUTC())
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrProviderRateLimited) {
			c.Set("Retry-After", "60")
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrExternalAIConsentRequired) {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
		if errors.Is(err, ErrAllProvidersFailed) {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		if errors.Is(err, ErrContentPolicyViolation) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		// P3-BE-08 / W5-A1-008: sentinel + errors.Is, не strings.Contains на err.Error().
		if errors.Is(err, ErrInvalidPrefix) || errors.Is(err, ErrEmptySuggestion) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	return c.JSON(resp)
}

// ChatCompletions POST /api/ai/v1/chat/completions
func (h *Handler) ChatCompletions(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var req ChatCompletionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Stream {
		if err := h.svc.StartChatStream(c.Context(), userID, req, c.IP(), string(c.Request().Header.UserAgent())); err != nil {
			if errors.Is(err, ErrDailyQuotaExceeded) {
				c.Set("Retry-After", secondsUntilMidnightUTC())
				return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
			}
			if errors.Is(err, ErrProviderRateLimited) {
				c.Set("Retry-After", "60")
				return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
			}
			if errors.Is(err, ErrProviderRPMUnavailable) {
				return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
			}
			if errors.Is(err, ErrExternalAIConsentRequired) {
				return fiber.NewError(fiber.StatusForbidden, err.Error())
			}
			if errors.Is(err, ErrContentPolicyViolation) {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			// A3W5-1-001: stream preflight — те же sentinel, что non-stream (не 400 + err.Error()).
			if errors.Is(err, ErrNoProvider) || errors.Is(err, ErrProviderDisabled) {
				return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
			}
			if errors.Is(err, ErrAllProvidersFailed) {
				return fiber.NewError(fiber.StatusBadGateway, err.Error())
			}
			return fiber.NewError(fiber.StatusBadGateway, "provider error")
		}
		return c.Status(fiber.StatusAccepted).JSON(StreamStartResponse{
			StreamID: req.StreamID,
			Status:   "streaming",
		})
	}
	resp, err := h.svc.ChatCompletion(c.Context(), userID, req, c.IP(), string(c.Request().Header.UserAgent()))
	if err != nil {
		if errors.Is(err, ErrNoProvider) || errors.Is(err, ErrProviderDisabled) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrProviderRPMUnavailable) {
			return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
		}
		if errors.Is(err, ErrDailyQuotaExceeded) {
			c.Set("Retry-After", secondsUntilMidnightUTC())
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrProviderRateLimited) {
			c.Set("Retry-After", "60")
			return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
		}
		if errors.Is(err, ErrExternalAIConsentRequired) {
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
		if errors.Is(err, ErrAllProvidersFailed) {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		if errors.Is(err, ErrContentPolicyViolation) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusBadGateway, "provider error")
	}
	return c.JSON(resp)
}

func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid user id")
	}
	return userID, nil
}

func secondsUntilMidnightUTC() string {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	sec := int(midnight.Sub(now).Seconds())
	if sec < 1 {
		sec = 1
	}
	return fmt.Sprintf("%d", sec)
}
