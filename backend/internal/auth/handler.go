package auth

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"

	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler предоставляет HTTP-обработчики для всех эндпоинтов аутентификации.
//
// ARCH-H1: handler не держит прямую зависимость от *ent.Client — все обращения
// к БД идут через AuthService (и его субсервисы). Это сохраняет слойную
// архитектуру handler → service → repository (db) и упрощает unit-тесты.
type Handler struct {
	service *AuthService
}

// NewHandler создаёт новый обработчик аутентификации.
func NewHandler(service *AuthService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты аутентификации на указанном Fiber-роутере.
// LO-08: имя параметра выровнено с остальными модулями (authMw).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	auth := api.Group("/auth")

	// Публичные эндпоинты (без аутентификации).
	auth.Post("/magic-link", h.SendMagicLink)
	auth.Post("/magic-link/verify", h.VerifyMagicLink)
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
	admin.Get("/users/pending", h.ListPendingUsers)
	admin.Post("/users/:userId/approve", h.ApprovePendingUser)
	admin.Post("/users/:userId/reject", h.RejectPendingUser)
	admin.Get("/organization/invite", h.GetOrgInvite)
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

	// invite_only: нужен существующий пользователь, валидный invite или setup founder.
	if err := h.service.Login.CanSendMagicLink(c.Context(), req.Email, req.InviteToken); err != nil {
		if err == ErrInvalidInviteToken || err == ErrRegistrationInviteOnly {
			return fiber.NewError(fiber.StatusForbidden, "registration is by invite only")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "registration check failed")
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
	if req.Code == "" || utf8.RuneCountInString(req.Code) != 6 {
		return fiber.NewError(fiber.StatusBadRequest, "code must be 6 characters")
	}

	device := utils.CopyString(c.Get("User-Agent"))
	ip := c.IP()

	tokens, err := h.service.LoginWithMagicLink(c.Context(), req.Email, req.Code, req.InviteToken, device, ip)
	if err != nil {
		if err == ErrMagicLinkInvalid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired code")
		}
		if err == ErrMagicLinkVerifyLimited {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many verification attempts, try again later")
		}
		if err == ErrRegistrationInviteOnly || err == ErrInvalidInviteToken {
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
// @Success 200 {object} TokenPairResponse "Новые JWT-токены"
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

	// Копируем sessionId из Fiber ctx до async (Pitfall 1).
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

// ListPendingUsers возвращает пользователей, ожидающих одобрения admin (ORG-06).
// @Summary Список пользователей в очереди Pending
// @Description Возвращает пользователей со статусом pending, ожидающих одобрения администратора (ORG-06)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Success 200 {array} PendingUserResponse "Очередь pending"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Router /admin/users/pending [get]
func (h *Handler) ListPendingUsers(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}
	users, err := h.service.Org.ListPendingUsers(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list pending users")
	}
	resp := make([]PendingUserResponse, 0, len(users))
	for _, u := range users {
		item := PendingUserResponse{ID: u.ID.String(), Email: u.Email}
		if u.PendingSince != nil {
			item.PendingSince = *u.PendingSince
		}
		resp = append(resp, item)
	}
	return c.JSON(resp)
}

// ApprovePendingUser активирует pending-пользователя (ORG-06).
// @Summary Одобрить pending-пользователя
// @Description Активирует пользователя из очереди pending (ORG-06)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Param userId path string true "UUID пользователя"
// @Success 200 {object} map[string]string "user approved"
// @Failure 400 {object} map[string]string "Некорректный UUID"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /admin/users/{userId}/approve [post]
func (h *Handler) ApprovePendingUser(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}
	userID, err := uuid.Parse(utils.CopyString(c.Params("userId")))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}
	if err := h.service.Org.ApprovePendingUser(c.Context(), userID); err != nil {
		if err == ErrUserNotFound {
			return fiber.NewError(fiber.StatusNotFound, "pending user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to approve user")
	}
	return c.JSON(fiber.Map{"message": "user approved"})
}

// RejectPendingUser удаляет pending-регистрацию (ORG-06).
// @Summary Отклонить pending-пользователя
// @Description Удаляет регистрацию пользователя из очереди pending (ORG-06)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Param userId path string true "UUID пользователя"
// @Success 200 {object} map[string]string "user rejected"
// @Failure 400 {object} map[string]string "Некорректный UUID"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /admin/users/{userId}/reject [post]
func (h *Handler) RejectPendingUser(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}
	userID, err := uuid.Parse(utils.CopyString(c.Params("userId")))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}
	if err := h.service.Org.RejectPendingUser(c.Context(), userID); err != nil {
		if err == ErrUserNotFound {
			return fiber.NewError(fiber.StatusNotFound, "pending user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to reject user")
	}
	return c.JSON(fiber.Map{"message": "user rejected"})
}

// GetOrgInvite возвращает invite-токен org для sharing (admin/founder, ORG-01).
// @Summary Получить org invite token
// @Description Возвращает invite-токен организации для invite-only регистрации (ORG-01)
// @Tags Администрирование
// @Produce json
// @Security BearerAuth
// @Success 200 {object} OrgInviteResponse "Invite token и URL"
// @Failure 401 {object} map[string]string "Не аутентифицирован"
// @Failure 403 {object} map[string]string "Нет прав администратора"
// @Failure 404 {object} map[string]string "Организация не настроена"
// @Router /admin/organization/invite [get]
func (h *Handler) GetOrgInvite(c fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}
	org, err := h.service.Org.GetOrganization(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "organization not configured")
	}
	return c.JSON(OrgInviteResponse{
		InviteToken: org.InviteToken,
		InviteURL:   "/login?invite=" + org.InviteToken,
	})
}

// ---- Вспомогательные функции ----

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

	u, err := h.service.GetMe(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user not found")
	}

	if !IsSystemAdmin(u.SystemRole) {
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	}

	return nil
}
