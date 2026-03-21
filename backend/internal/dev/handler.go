// Package dev предоставляет dev-эндпоинты для локальной разработки:
// seed демо-пользователей и вход без Magic Link / Telegram.
// Все эндпоинты доступны только при DEV_MODE=true.
package dev

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
)

// Handler обрабатывает dev-эндпоинты (seed, login, status).
type Handler struct {
	db      *ent.Client
	authSvc *auth.AuthService
}

// NewHandler создаёт новый Handler для dev-эндпоинтов.
func NewHandler(db *ent.Client, authSvc *auth.AuthService) *Handler {
	return &Handler{db: db, authSvc: authSvc}
}

// RegisterRoutes регистрирует dev-маршруты в группе /api/dev.
// Все маршруты защищены проверкой DevMode — при DevMode=false возвращается 404.
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

// DevLoginRequest — тело запроса для POST /api/dev/login.
type DevLoginRequest struct {
	Email string `json:"email"`
}

// DevUserInfo — информация о демо-пользователе в ответе seed.
type DevUserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	SystemRole  string `json:"system_role"`
	ProjectRole string `json:"project_role,omitempty"`
	Description string `json:"description"`
}

// --- Демо-пользователи ---

type demoUser struct {
	Email       string
	Name        string
	DisplayName string
	SystemRole  user.SystemRole
	ProjectRole projectmember.ProjectRole
	Description string
	AddToProject bool
}

var demoUsers = []demoUser{
	{
		Email: "admin@hub.local", Name: "Админ Хабов", DisplayName: "Админ",
		SystemRole: user.SystemRoleAdmin, ProjectRole: "",
		Description: "Полный доступ", AddToProject: false,
	},
	{
		Email: "pm@hub.local", Name: "Пётр Менеджеров", DisplayName: "Пётр",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleOwner,
		Description: "Project Manager", AddToProject: true,
	},
	{
		Email: "dev1@hub.local", Name: "Дмитрий Кодов", DisplayName: "Дмитрий",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleDeveloper,
		Description: "Frontend разработчик", AddToProject: true,
	},
	{
		Email: "dev2@hub.local", Name: "Анна Бэкендова", DisplayName: "Анна",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleDeveloper,
		Description: "Backend разработчик", AddToProject: true,
	},
	{
		Email: "tester1@hub.local", Name: "Тестер Тестеров", DisplayName: "Тестер",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleTester,
		Description: "QA инженер", AddToProject: true,
	},
	{
		Email: "tester2@hub.local", Name: "Мария Автотестова", DisplayName: "Мария",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleTester,
		Description: "Автоматизатор", AddToProject: true,
	},
	{
		Email: "viewer@hub.local", Name: "Наблюдатель Смотров", DisplayName: "Наблюдатель",
		SystemRole: user.SystemRoleMember, ProjectRole: projectmember.ProjectRoleViewer,
		Description: "Только просмотр", AddToProject: true,
	},
	{
		Email: "newbie@hub.local", Name: "Новичок Первовходов", DisplayName: "Новичок",
		SystemRole: user.SystemRoleMember, ProjectRole: "",
		Description: "Не прошёл онбординг", AddToProject: false,
	},
}

// --- Handlers ---

// Status возвращает 200 если dev-режим активен. Используется фронтендом для определения доступности.
func (h *Handler) Status(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"dev_mode": true})
}

// Seed создаёт демо-пользователей и проект "Hub Demo" (DEMO).
// Если пользователь уже существует — пропускает. Идемпотентный эндпоинт.
func (h *Handler) Seed(c fiber.Ctx) error {
	ctx := c.Context()

	// Создаём или находим пользователей
	createdUsers := make([]DevUserInfo, 0, len(demoUsers))
	userIDs := make(map[string]uuid.UUID) // email -> user ID

	for _, du := range demoUsers {
		u, err := h.findOrCreateDemoUser(ctx, du)
		if err != nil {
			slog.Error("dev seed: не удалось создать пользователя", "email", du.Email, "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to seed user: "+du.Email)
		}
		userIDs[du.Email] = u.ID

		info := DevUserInfo{
			ID:          u.ID.String(),
			Email:       u.Email,
			Name:        u.Name,
			SystemRole:  string(u.SystemRole),
			Description: du.Description,
		}
		if du.ProjectRole != "" {
			info.ProjectRole = string(du.ProjectRole)
		}
		createdUsers = append(createdUsers, info)
	}

	// Создаём или находим демо-проект
	proj, err := h.findOrCreateDemoProject(ctx)
	if err != nil {
		slog.Error("dev seed: не удалось создать проект", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to seed demo project")
	}

	// Назначаем роли в проекте
	for _, du := range demoUsers {
		if !du.AddToProject || du.ProjectRole == "" {
			continue
		}
		uid := userIDs[du.Email]
		if err := h.ensureProjectMember(ctx, proj.ID, uid, du.ProjectRole); err != nil {
			slog.Error("dev seed: не удалось добавить в проект", "email", du.Email, "error", err)
		}
	}

	slog.Info("dev seed: демо-данные созданы", "users", len(createdUsers), "project", proj.Key)

	return c.JSON(fiber.Map{
		"users":   createdUsers,
		"project": fiber.Map{"id": proj.ID.String(), "name": proj.Name, "key": proj.Key},
	})
}

// Login выполняет вход без Magic Link — напрямую создаёт JWT-сессию по email.
func (h *Handler) Login(c fiber.Ctx) error {
	var req DevLoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	// Ищем пользователя по email
	u, err := h.db.User.Query().Where(user.EmailEQ(req.Email)).Only(c.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "user not found, run /api/dev/seed first")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to find user")
	}

	// Создаём сессию напрямую (без Magic Link)
	device := "dev-login"
	ip := c.IP()

	tokens, err := h.authSvc.CreateSession(c.Context(), u.ID, device, ip)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
	}

	// Возвращаем токены + пользователя (как обычный login)
	return c.JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"user":          auth.UserToResponse(u),
	})
}

// --- Helpers ---

// findOrCreateDemoUser ищет пользователя по email, создаёт если не найден.
func (h *Handler) findOrCreateDemoUser(ctx context.Context, du demoUser) (*ent.User, error) {
	u, err := h.db.User.Query().Where(user.EmailEQ(du.Email)).Only(ctx)
	if err == nil {
		return u, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	u, err = h.db.User.Create().
		SetEmail(du.Email).
		SetName(du.Name).
		SetDisplayName(du.DisplayName).
		SetSystemRole(du.SystemRole).
		SetOnboardingCompleted(du.Email != "newbie@hub.local").
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// findOrCreateDemoProject ищет проект DEMO, создаёт если не найден.
func (h *Handler) findOrCreateDemoProject(ctx context.Context) (*ent.Project, error) {
	p, err := h.db.Project.Query().Where(project.KeyEQ("DEMO")).Only(ctx)
	if err == nil {
		return p, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	p, err = h.db.Project.Create().
		SetName("Hub Demo").
		SetKey("DEMO").
		SetDescription("Демо-проект для разработки и тестирования").
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ensureProjectMember добавляет пользователя в проект с указанной ролью, если ещё не добавлен.
func (h *Handler) ensureProjectMember(ctx context.Context, projectID, userID uuid.UUID, role projectmember.ProjectRole) error {
	exists, err := h.db.ProjectMember.Query().
		Where(
			projectmember.HasUserWith(user.ID(userID)),
			projectmember.HasProjectWith(project.ID(projectID)),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = h.db.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(projectID).
		SetProjectRole(role).
		Save(ctx)
	return err
}
