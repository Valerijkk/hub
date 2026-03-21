package auth

import (
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler is the HTTP handler for all auth-related endpoints.
type Handler struct {
	service *AuthService
	db      *ent.Client
}

// NewHandler creates a new auth Handler.
func NewHandler(service *AuthService, db *ent.Client) *Handler {
	return &Handler{service: service, db: db}
}

// RegisterRoutes registers all auth routes on the given Fiber router.
func (h *Handler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) {
	auth := api.Group("/auth")

	// Public endpoints (no auth middleware).
	auth.Post("/magic-link", h.SendMagicLink)
	auth.Post("/magic-link/verify", h.VerifyMagicLink)
	auth.Post("/telegram", h.LoginTelegram)
	auth.Post("/refresh", h.RefreshToken)

	// Protected endpoints.
	authGroup := auth.Group("", authMiddleware)
	authGroup.Post("/logout", h.Logout)
	authGroup.Get("/me", h.GetMe)
	authGroup.Put("/me", h.UpdateProfile)
	authGroup.Post("/me/onboarding", h.CompleteOnboarding)
	authGroup.Get("/sessions", h.ListSessions)
	authGroup.Delete("/sessions/:sessionId", h.RevokeSession)

	// Admin-only endpoints.
	admin := api.Group("/admin", authMiddleware)
	admin.Get("/registration-mode", h.GetRegistrationMode)
	admin.Put("/registration-mode", h.SetRegistrationMode)
}

// ---- Public handlers ----

// SendMagicLink обрабатывает отправку magic-link для аутентификации.
// @Summary Отправить magic-link на email
// @Description Отправляет одноразовый код на указанный email для входа в систему
// @Tags auth
// @Accept json
// @Produce json
// @Param body body MagicLinkRequest true "Email пользователя"
// @Success 200 {object} map[string]string "Magic link sent"
// @Failure 400 {object} map[string]string "Некорректный email"
// @Failure 403 {object} map[string]string "Регистрация только по приглашению"
// @Failure 429 {object} map[string]string "Слишком много запросов"
// @Router /auth/magic-link [post]
func (h *Handler) SendMagicLink(c fiber.Ctx) error {
	var req MagicLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid email format")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check registration mode: if invite_only and user does not exist, reject.
	mode, _ := h.service.GetRegistrationMode(c.Context())
	if mode == "invite_only" {
		exists, _ := h.db.User.Query().Where(user.EmailEQ(req.Email)).Exist(c.Context())
		if !exists {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
	}

	if err := h.service.magicLink.SendMagicLink(c.Context(), req.Email); err != nil {
		if err == ErrMagicLinkRateLimited {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests, try again later")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send magic link")
	}

	return c.JSON(fiber.Map{"message": "Magic link sent"})
}

// VerifyMagicLink проверяет одноразовый код и выдаёт JWT-токены.
// @Summary Верификация magic-link кода
// @Description Проверяет 6-значный код, создаёт сессию и возвращает access/refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param body body VerifyMagicLinkRequest true "Email и код верификации"
// @Success 200 {object} TokenPairResponse "JWT-токены"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Неверный или просроченный код"
// @Router /auth/magic-link/verify [post]
func (h *Handler) VerifyMagicLink(c fiber.Ctx) error {
	var req VerifyMagicLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Code == "" || len(req.Code) != 6 {
		return fiber.NewError(fiber.StatusBadRequest, "code must be 6 characters")
	}

	device := utils.CopyString(c.Get("User-Agent"))
	ip := c.IP()

	tokens, err := h.service.LoginWithMagicLink(c.Context(), req.Email, req.Code, device, ip)
	if err != nil {
		if err == ErrMagicLinkInvalid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired code")
		}
		if err == ErrRegistrationInviteOnly {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}

	return c.JSON(tokens)
}

// LoginTelegram handles POST /api/auth/telegram.
func (h *Handler) LoginTelegram(c fiber.Ctx) error {
	var req TelegramAuthRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	device := utils.CopyString(c.Get("User-Agent"))
	ip := c.IP()

	tokens, err := h.service.LoginWithTelegram(c.Context(), req.InitData, device, ip)
	if err != nil {
		if err == ErrTelegramAuthExpired {
			return fiber.NewError(fiber.StatusUnauthorized, "telegram auth expired")
		}
		if err == ErrRegistrationInviteOnly {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
		return fiber.NewError(fiber.StatusUnauthorized, "telegram auth failed")
	}

	return c.JSON(tokens)
}

// RefreshToken обновляет пару JWT-токенов по refresh-токену.
// @Summary Обновить JWT-токены
// @Description Принимает refresh-токен и возвращает новую пару access/refresh
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RefreshRequest true "Refresh-токен"
// @Success 200 {object} TokenPair "Новые JWT-токены"
// @Failure 401 {object} map[string]string "Невалидный refresh-токен"
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	tokens, err := h.service.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired refresh token")
	}

	return c.JSON(tokens)
}

// ---- Protected handlers ----

// Logout завершает текущую сессию пользователя.
// @Summary Выход из системы
// @Description Инвалидирует текущую сессию и refresh-токен
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Logged out"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Router /auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	sessionID, err := sessionIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid session")
	}

	if err := h.service.Logout(c.Context(), sessionID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "logout failed")
	}

	return c.JSON(fiber.Map{"message": "Logged out"})
}

// GetMe возвращает профиль текущего аутентифицированного пользователя.
// @Summary Получить профиль текущего пользователя
// @Description Возвращает данные профиля по JWT-токену
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse "Профиль пользователя"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Router /auth/me [get]
func (h *Handler) GetMe(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	u, err := h.service.GetMe(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	return c.JSON(UserToResponse(u))
}

// UpdateProfile handles PUT /api/auth/me.
func (h *Handler) UpdateProfile(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	var req UpdateProfileRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	u, err := h.service.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "update profile failed")
	}

	return c.JSON(UserToResponse(u))
}

// CompleteOnboarding handles POST /api/auth/me/onboarding.
func (h *Handler) CompleteOnboarding(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	var req OnboardingCompleteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	u, err := h.service.CompleteOnboarding(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "onboarding failed")
	}

	return c.JSON(UserToResponse(u))
}

// ListSessions handles GET /api/auth/sessions.
func (h *Handler) ListSessions(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	currentSessionID, _ := sessionIDFromCtx(c)

	sessions, err := h.service.ListSessions(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "list sessions failed")
	}

	result := make([]SessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = SessionResponse{
			ID:           s.ID.String(),
			Device:       s.Device,
			IPAddress:    s.IPAddress,
			LastActiveAt: s.LastActiveAt,
			CreatedAt:    s.CreatedAt,
			IsCurrent:    s.ID == currentSessionID,
		}
	}

	return c.JSON(result)
}

// RevokeSession handles DELETE /api/auth/sessions/:sessionId.
func (h *Handler) RevokeSession(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	// Copy string from Fiber ctx (Pitfall 1).
	sessionIDStr := utils.CopyString(c.Params("sessionId"))
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid session ID")
	}

	if err := h.service.RevokeSession(c.Context(), userID, sessionID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "session not found")
	}

	return c.JSON(fiber.Map{"message": "Session revoked"})
}

// ---- Admin handlers ----

// GetRegistrationMode handles GET /api/admin/registration-mode.
func (h *Handler) GetRegistrationMode(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	mode, err := h.service.GetRegistrationMode(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get registration mode")
	}

	return c.JSON(fiber.Map{"mode": mode})
}

// SetRegistrationMode handles PUT /api/admin/registration-mode.
func (h *Handler) SetRegistrationMode(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	var req SetRegistrationModeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Mode != "open" && req.Mode != "invite_only" {
		return fiber.NewError(fiber.StatusBadRequest, "mode must be 'open' or 'invite_only'")
	}

	if err := h.service.SetRegistrationMode(c.Context(), req.Mode); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to set registration mode")
	}

	return c.JSON(fiber.Map{"mode": req.Mode})
}

// ---- Helpers ----

func sessionIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("sessionID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "session not found")
	}
	return uuid.Parse(idStr)
}

// requireAdmin checks that the current user has system_role "admin".
func (h *Handler) requireAdmin(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}

	u, err := h.db.User.Get(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}

	if u.SystemRole != user.SystemRoleAdmin {
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	}

	return nil
}
