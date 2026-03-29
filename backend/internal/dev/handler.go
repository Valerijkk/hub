// Package dev provides dev-endpoints for local development:
// seed demo users, project, issues, boards, labels, test cases and channels.
// Login without Magic Link / Telegram.
// All endpoints available only when DEV_MODE=true.
package dev

import (
	"github.com/gofiber/fiber/v3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
)

// Handler handles dev-endpoints (seed, login, status).
type Handler struct {
	db      *ent.Client
	authSvc *auth.AuthService
}

// NewHandler creates a new Handler for dev-endpoints.
func NewHandler(db *ent.Client, authSvc *auth.AuthService) *Handler {
	return &Handler{db: db, authSvc: authSvc}
}

// RegisterRoutes registers dev routes under /api/dev.
// All routes are guarded by DevMode check -- returns 404 when DevMode=false.
func (h *Handler) RegisterRoutes(api fiber.Router, devMode bool) {
	dev := api.Group("/dev")

	guard := func(c fiber.Ctx) error {
		if !devMode {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not available"})
		}
		return c.Next()
	}

	dev.Use(guard)
	dev.Get("/status", h.Status)
	dev.Post("/seed", h.Seed)
	dev.Post("/login", h.Login)
}

// --- DTOs ---

// DevLoginRequest is the request body for POST /api/dev/login.
type DevLoginRequest struct {
	Email string `json:"email"`
}

// DevUserInfo contains demo user info returned in seed response.
type DevUserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	SystemRole  string `json:"system_role"`
	ProjectRole string `json:"project_role,omitempty"`
	Description string `json:"description"`
}

// --- Handlers ---

// Status returns 200 if dev mode is active. Used by frontend to detect availability.
// @Summary Проверить статус dev-режима
// @Description Возвращает 200 если dev-режим активен. Используется фронтендом для определения доступности
// @Tags Разработка
// @Produce json
// @Success 200 {object} map[string]bool "dev_mode: true"
// @Router /dev/status [get]
func (h *Handler) Status(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"dev_mode": true})
}

// Login performs login without Magic Link -- directly creates a JWT session by email.
// @Summary Dev-вход по email
// @Description Создаёт JWT-сессию напрямую по email без magic-link. Только в dev-режиме
// @Tags Разработка
// @Accept json
// @Produce json
// @Param body body DevLoginRequest true "Email пользователя"
// @Success 200 {object} object "access_token, refresh_token, user"
// @Failure 400 {object} map[string]string "Email не указан"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /dev/login [post]
func (h *Handler) Login(c fiber.Ctx) error {
	var req DevLoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	u, err := h.db.User.Query().Where(user.EmailEQ(req.Email)).Only(c.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "user not found, run /api/dev/seed first")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to find user")
	}

	device := "dev-login"
	ip := c.IP()

	tokens, err := h.authSvc.CreateSession(c.Context(), u.ID, device, ip)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
	}

	return c.JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"user":          auth.UserToResponse(u),
	})
}
