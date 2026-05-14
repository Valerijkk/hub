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

// Handler предоставляет HTTP-обработчики для всех эндпоинтов аутентификации.
type Handler struct {
	service *AuthService
	db      *ent.Client
	tgBot   *TelegramBotAuth
}

// NewHandler создаёт новый обработчик аутентификации.
func NewHandler(service *AuthService, db *ent.Client, tgBot *TelegramBotAuth) *Handler {
	return &Handler{service: service, db: db, tgBot: tgBot}
}

// RegisterRoutes регистрирует все маршруты аутентификации на указанном Fiber-роутере.
// LO-08: имя параметра выровнено с остальными модулями (authMw).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	auth := api.Group("/auth")

	// Публичные эндпоинты (без аутентификации).
	auth.Post("/magic-link", h.SendMagicLink)
	auth.Post("/magic-link/verify", h.VerifyMagicLink)
	auth.Post("/telegram", h.LoginTelegram)
	auth.Post("/telegram-code", h.LoginTelegramCode)
	auth.Post("/refresh", h.RefreshToken)

	// Защищённые эндпоинты (требуют аутентификации).
	authGroup := auth.Group("", authMw)
	authGroup.Post("/logout", h.Logout)
	authGroup.Get("/me", h.GetMe)
	authGroup.Put("/me", h.UpdateProfile)
	authGroup.Post("/me/onboarding", h.CompleteOnboarding)
	authGroup.Get("/sessions", h.ListSessions)
	authGroup.Delete("/sessions/:sessionId", h.RevokeSession)

	// Эндпоинты только для администраторов.
	admin := api.Group("/admin", authMw)
	admin.Get("/registration-mode", h.GetRegistrationMode)
	admin.Put("/registration-mode", h.SetRegistrationMode)
}

// ---- Публичные обработчики ----

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

	code, err := h.service.magicLink.SendMagicLink(c.Context(), req.Email)
	if err != nil {
		if err == ErrMagicLinkRateLimited {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests, try again later")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send magic link")
	}

	resp := fiber.Map{"message": "Magic link sent"}
	if h.service.magicLink.IsDevMode() {
		resp["code"] = code
	}
	return c.JSON(resp)
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
		if err == ErrMagicLinkVerifyLimited {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many verification attempts, try again later")
		}
		if err == ErrRegistrationInviteOnly {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}

	return c.JSON(tokens)
}

// LoginTelegram обрабатывает POST /api/auth/telegram — вход через Telegram Web App.
// @Summary Вход через Telegram Web App
// @Description Аутентификация через initData из Telegram Mini App
// @Tags auth
// @Accept json
// @Produce json
// @Param body body TelegramAuthRequest true "initData из Telegram Web App"
// @Success 200 {object} TokenPairResponse "JWT-токены"
// @Failure 401 {object} map[string]string "Ошибка аутентификации Telegram"
// @Failure 403 {object} map[string]string "Регистрация только по приглашению"
// @Router /auth/telegram [post]
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

// LoginTelegramCode обрабатывает POST /api/auth/telegram-code — вход по коду из Telegram-бота.
// @Summary Вход по коду из Telegram-бота
// @Description Аутентификация по одноразовому коду, полученному от Telegram-бота
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object true "code — одноразовый код"
// @Success 200 {object} TokenPairResponse "JWT-токены"
// @Failure 400 {object} map[string]string "Код не указан"
// @Failure 401 {object} map[string]string "Неверный или просроченный код"
// @Failure 403 {object} map[string]string "Регистрация только по приглашению"
// @Failure 503 {object} map[string]string "Telegram-бот не настроен"
// @Router /auth/telegram-code [post]
func (h *Handler) LoginTelegramCode(c fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind().JSON(&req); err != nil || req.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code is required")
	}

	if h.tgBot == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "telegram bot not configured")
	}

	tgUser, err := h.tgBot.VerifyCode(c.Context(), req.Code, c.IP())
	if err != nil {
		if err == ErrTelegramVerifyLimited {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many verification attempts, try again later")
		}
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	device := utils.CopyString(c.Get("User-Agent"))
	ip := c.IP()

	tokens, err := h.service.LoginWithTelegramUser(c.Context(), tgUser, device, ip)
	if err != nil {
		if err == ErrRegistrationInviteOnly {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
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

// UpdateProfile обрабатывает PUT /api/auth/me — обновление профиля текущего пользователя.
// @Summary Обновить профиль пользователя
// @Description Обновляет имя, аватар и другие поля профиля текущего пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateProfileRequest true "Поля для обновления"
// @Success 200 {object} UserResponse "Обновлённый профиль"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Router /auth/me [put]
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

// CompleteOnboarding обрабатывает POST /api/auth/me/onboarding — завершение онбординга.
// @Summary Завершить онбординг пользователя
// @Description Сохраняет данные онбординга (имя, роль, предпочтения) и отмечает его пройденным
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body OnboardingCompleteRequest true "Данные онбординга"
// @Success 200 {object} UserResponse "Профиль после онбординга"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Router /auth/me/onboarding [post]
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

// ListSessions обрабатывает GET /api/auth/sessions — список активных сессий пользователя.
// @Summary Получить список активных сессий
// @Description Возвращает все активные сессии текущего пользователя с пометкой текущей
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SessionResponse "Список сессий"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Router /auth/sessions [get]
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

// RevokeSession обрабатывает DELETE /api/auth/sessions/:sessionId — отзыв сессии.
// @Summary Отозвать сессию
// @Description Завершает указанную сессию пользователя (можно завершить любую свою сессию)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "ID сессии"
// @Success 200 {object} map[string]string "Сессия отозвана"
// @Failure 400 {object} map[string]string "Некорректный ID сессии"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 404 {object} map[string]string "Сессия не найдена"
// @Router /auth/sessions/{sessionId} [delete]
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

// GetRegistrationMode обрабатывает GET /api/admin/registration-mode — получение режима регистрации.
// @Summary Получить режим регистрации
// @Description Возвращает текущий режим регистрации: open или invite_only (только для админов)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Текущий режим"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Router /admin/registration-mode [get]
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

// SetRegistrationMode обрабатывает PUT /api/admin/registration-mode — установка режима регистрации.
// @Summary Установить режим регистрации
// @Description Переключает режим регистрации между open и invite_only (только для админов)
// @Tags Администрирование
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SetRegistrationModeRequest true "Новый режим: open или invite_only"
// @Success 200 {object} map[string]string "Установленный режим"
// @Failure 400 {object} map[string]string "Некорректный режим"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Router /admin/registration-mode [put]
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

// sessionIDFromCtx извлекает ID сессии из контекста Fiber.
func sessionIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("sessionID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "session not found")
	}
	return uuid.Parse(idStr)
}

// requireAdmin проверяет, что текущий пользователь имеет system_role "admin".
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
