//go:build !release

// Package dev предоставляет dev-эндпоинты для локальной разработки:
// сидирование демо-пользователей, проекта, задач, досок, меток, тест-кейсов и каналов.
// Вход без Magic Link (только DEV). Telegram — notification bind, не login (ADR-AUTH-01).
// Все эндпоинты доступны только при DEV_MODE=true.
//
// ME-09: этот файл НЕ компилируется в release-сборках (`go build -tags release`).
// В release-бинарнике заглушка в handler_release.go предоставляет no-op
// Handler, чей RegisterRoutes ничего не делает — поэтому /api/dev/* просто
// не существует, независимо от переменной DEV_MODE. Это делает опасные
// пути /api/dev/login и /api/dev/seed *невозможными для включения в рантайме*.
package dev

import (
	"sync"

	"github.com/gofiber/fiber/v3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
)

// Handler обрабатывает dev-эндпоинты (seed, login, status).
type Handler struct {
	db                 *ent.Client
	authSvc            *auth.AuthService
	mailCalendarSeeder MailCalendarDemoSeeder
	seedMu             sync.Mutex
}

// NewHandler создаёт новый Handler для dev-эндпоинтов.
func NewHandler(db *ent.Client, authSvc *auth.AuthService, mailCalendarSeeder MailCalendarDemoSeeder) *Handler {
	return &Handler{db: db, authSvc: authSvc, mailCalendarSeeder: mailCalendarSeeder}
}

// RegisterRoutes регистрирует dev-маршруты под /api/dev.
// Все маршруты защищены проверкой DevMode — возвращает 404 при DevMode=false.
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

// --- DTO ---

// DevLoginRequest — тело запроса для POST /api/dev/login.
type DevLoginRequest struct {
	Email string `json:"email"`
}

// DevUserInfo содержит информацию о демо-пользователе в ответе seed.
type DevUserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	SystemRole  string `json:"system_role"`
	ProjectRole string `json:"project_role,omitempty"`
	Description string `json:"description"`
}

// --- Обработчики ---

// Status возвращает 200, если dev-режим активен. Используется фронтендом для проверки доступности.
// @Summary Проверить статус dev-режима
// @Description Возвращает 200 если dev-режим активен. Используется фронтендом для определения доступности
// @Tags Разработка
// @Produce json
// @Success 200 {object} map[string]bool "dev_mode: true"
// @Router /dev/status [get]
func (h *Handler) Status(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"dev_mode": true})
}

// Login выполняет вход без Magic Link — напрямую создаёт JWT-сессию по email.
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

	// P3-BE-16 / W7-A5-012: не обходить StatusActive (pending/deactivated → 403).
	if u.Status != user.StatusActive {
		return fiber.NewError(fiber.StatusForbidden, "user is not active")
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
