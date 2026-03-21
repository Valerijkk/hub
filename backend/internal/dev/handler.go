// Package dev предоставляет dev-эндпоинты для локальной разработки:
// seed демо-пользователей, проекта, задач, досок, меток, тест-кейсов и каналов.
// Вход без Magic Link / Telegram.
// Все эндпоинты доступны только при DEV_MODE=true.
package dev

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/board"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
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
	Email        string
	Name         string
	DisplayName  string
	SystemRole   user.SystemRole
	ProjectRole  projectmember.ProjectRole
	Description  string
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

// --- Демо workflow-статусы ---

type demoStatus struct {
	Name     string
	Color    string
	Category workflowstatus.Category
	Order    int
}

var demoStatuses = []demoStatus{
	{Name: "Бэклог", Color: "#94A3B8", Category: workflowstatus.CategoryTodo, Order: 1},
	{Name: "К выполнению", Color: "#3B82F6", Category: workflowstatus.CategoryTodo, Order: 2},
	{Name: "В работе", Color: "#F59E0B", Category: workflowstatus.CategoryInProgress, Order: 3},
	{Name: "На ревью", Color: "#8B5CF6", Category: workflowstatus.CategoryInProgress, Order: 4},
	{Name: "Тестирование", Color: "#10B981", Category: workflowstatus.CategoryInProgress, Order: 5},
	{Name: "Готово", Color: "#22C55E", Category: workflowstatus.CategoryDone, Order: 6},
}

// --- Демо задачи ---

type demoIssue struct {
	Key          string
	IssueType    issue.IssueType
	Title        string
	StatusName   string
	AssigneeEmail string
	Priority     issue.Priority
	ParentKey    string // ключ родительской задачи (пустой — нет)
}

var demoIssues = []demoIssue{
	{"DEMO-1", issue.IssueTypeEpic, "Авторизация и регистрация", "В работе", "pm@hub.local", issue.PriorityHigh, ""},
	{"DEMO-2", issue.IssueTypeStory, "Magic Link вход по email", "На ревью", "dev1@hub.local", issue.PriorityHigh, "DEMO-1"},
	{"DEMO-3", issue.IssueTypeTask, "Реализовать отправку кода через Resend", "Готово", "dev1@hub.local", issue.PriorityMedium, "DEMO-2"},
	{"DEMO-4", issue.IssueTypeBug, "Код не приходит на Gmail", "К выполнению", "dev2@hub.local", issue.PriorityCritical, ""},
	{"DEMO-5", issue.IssueTypeStory, "Telegram авторизация", "В работе", "dev2@hub.local", issue.PriorityHigh, ""},
	{"DEMO-6", issue.IssueTypeSubTask, "Верификация HMAC подписи", "Тестирование", "dev2@hub.local", issue.PriorityMedium, "DEMO-5"},
	{"DEMO-7", issue.IssueTypeEpic, "Трекер задач", "Бэклог", "pm@hub.local", issue.PriorityHigh, ""},
	{"DEMO-8", issue.IssueTypeImprovement, "Ускорить загрузку Kanban доски", "К выполнению", "dev1@hub.local", issue.PriorityMedium, ""},
	{"DEMO-9", issue.IssueTypeTechDebt, "Рефакторинг RBAC мидлвары", "Бэклог", "dev2@hub.local", issue.PriorityLow, ""},
	{"DEMO-10", issue.IssueTypeSpike, "Исследовать WebRTC для звонков", "Готово", "dev1@hub.local", issue.PriorityMedium, ""},
	{"DEMO-11", issue.IssueTypeTask, "Настроить CI/CD пайплайн", "Готово", "dev2@hub.local", issue.PriorityHigh, ""},
	{"DEMO-12", issue.IssueTypeBug, "Sidebar не сворачивается на мобилке", "К выполнению", "dev1@hub.local", issue.PriorityMedium, ""},
	{"DEMO-13", issue.IssueTypeStory, "Загрузка файлов drag & drop", "В работе", "dev1@hub.local", issue.PriorityMedium, ""},
	{"DEMO-14", issue.IssueTypeRelease, "Релиз v0.1.0-alpha", "Бэклог", "pm@hub.local", issue.PriorityHigh, ""},
	{"DEMO-15", issue.IssueTypeTask, "Написать тесты для auth модуля", "Тестирование", "tester1@hub.local", issue.PriorityMedium, ""},
}

// --- Демо связи ---

type demoLink struct {
	SourceKey string
	TargetKey string
	LinkType  issuelink.LinkType
}

var demoLinks = []demoLink{
	{"DEMO-4", "DEMO-2", issuelink.LinkTypeBlocks},
	{"DEMO-12", "DEMO-8", issuelink.LinkTypeRelatesTo},
}

// --- Демо метки ---

type demoLabel struct {
	Name  string
	Color string
}

var demoLabels = []demoLabel{
	{"frontend", "#3B82F6"},
	{"backend", "#10B981"},
	{"bug", "#EF4444"},
	{"urgent", "#F97316"},
	{"docs", "#94A3B8"},
}

// --- Демо тест-кейсы ---

type demoTestCase struct {
	CaseKey     string
	Title       string
	Description string
	Steps       []schema.TestStep
}

var demoTestCases = []demoTestCase{
	{
		CaseKey: "TC-DEMO-1", Title: "Вход по Magic Link",
		Description: "Проверка входа через email magic link",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Открыть страницу /login", ExpectedResult: "Отображается форма входа"},
			{Order: 2, Action: "Ввести email и нажать 'Получить ссылку'", ExpectedResult: "Показано сообщение об отправке"},
			{Order: 3, Action: "Перейти по ссылке из email", ExpectedResult: "Выполнен вход, редирект на /dashboard"},
		},
	},
	{
		CaseKey: "TC-DEMO-2", Title: "Вход через Telegram",
		Description: "Проверка входа через Telegram бота",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Нажать 'Войти через Telegram'", ExpectedResult: "Открывается Telegram виджет"},
			{Order: 2, Action: "Подтвердить вход в Telegram", ExpectedResult: "Выполнен вход, редирект на /dashboard"},
		},
	},
	{
		CaseKey: "TC-DEMO-3", Title: "Создание проекта",
		Description: "Проверка создания нового проекта",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Перейти в /projects и нажать 'Создать проект'", ExpectedResult: "Открыта форма создания"},
			{Order: 2, Action: "Заполнить название и ключ, нажать 'Создать'", ExpectedResult: "Проект создан, открыта страница проекта"},
		},
	},
	{
		CaseKey: "TC-DEMO-4", Title: "Перемещение задачи на Kanban",
		Description: "Проверка drag & drop задачи между колонками",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Открыть Kanban доску проекта", ExpectedResult: "Колонки со статусами отображаются"},
			{Order: 2, Action: "Перетащить задачу из 'К выполнению' в 'В работе'", ExpectedResult: "Задача переместилась, статус обновлён"},
		},
	},
	{
		CaseKey: "TC-DEMO-5", Title: "Загрузка файла",
		Description: "Проверка загрузки файла к задаче",
		Steps: []schema.TestStep{
			{Order: 1, Action: "Открыть задачу и перетащить файл в зону загрузки", ExpectedResult: "Файл загружен, отображается в списке"},
		},
	},
}

// --- Демо каналы ---

type demoChannel struct {
	Name        string
	Description string
	Type        channel.ChannelType
	IsReadonly  bool
	// MemberEmails — email-ы участников (пустой = все участники проекта)
	MemberEmails []string
}

var demoChannels = []demoChannel{
	{
		Name: "demo-general", Description: "Общий канал проекта",
		Type: channel.ChannelTypePublic, IsReadonly: false,
		MemberEmails: []string{"pm@hub.local", "dev1@hub.local", "dev2@hub.local", "tester1@hub.local", "tester2@hub.local"},
	},
	{
		Name: "demo-dev", Description: "Канал разработчиков (приватный)",
		Type: channel.ChannelTypePrivate, IsReadonly: false,
		MemberEmails: []string{"dev1@hub.local", "dev2@hub.local"},
	},
	{
		Name: "demo-bugs", Description: "Канал багов (только чтение, bot)",
		Type: channel.ChannelTypePublic, IsReadonly: true,
		MemberEmails: []string{"pm@hub.local", "dev1@hub.local", "dev2@hub.local", "tester1@hub.local"},
	},
}

// --- Демо тестовые прогоны ---

// demoTestRun описывает демо-прогон тестов для seed.
type demoTestRun struct {
	Title       string
	Status      testrun.Status
	RunType     testrun.RunType
	TotalCount  int
	PassedCount int
	FailedCount int
	SkippedCount int
	BrokenCount int
	UnknownCount int
	// Results — результаты для каждого тест-кейса (по индексу из demoTestCases).
	Results []demoTestResult
}

// demoTestResult описывает результат теста в демо-прогоне.
type demoTestResult struct {
	CaseIdx    int                 // индекс тест-кейса в demoTestCases
	Status     testresult.Status
	DurationMs int64
	Comment    string
}

var demoTestRuns = []demoTestRun{
	{
		Title: "Регрессия v0.1.0", Status: testrun.StatusCompleted, RunType: testrun.RunTypeManual,
		TotalCount: 5, PassedCount: 3, FailedCount: 1, SkippedCount: 1,
		Results: []demoTestResult{
			{CaseIdx: 0, Status: testresult.StatusPassed, DurationMs: 12000, Comment: "Вход выполнен успешно"},
			{CaseIdx: 1, Status: testresult.StatusPassed, DurationMs: 8000, Comment: "Telegram верификация пройдена"},
			{CaseIdx: 2, Status: testresult.StatusPassed, DurationMs: 15000, Comment: "Проект создан корректно"},
			{CaseIdx: 3, Status: testresult.StatusFailed, DurationMs: 23000, Comment: "Карточка не переместилась — drag event не срабатывает"},
			{CaseIdx: 4, Status: testresult.StatusSkipped, DurationMs: 0, Comment: "Пропущено — SeaweedFS не запущен"},
		},
	},
	{
		Title: "Smoke тест Sprint 1", Status: testrun.StatusInProgress, RunType: testrun.RunTypeManual,
		TotalCount: 3, PassedCount: 1, FailedCount: 0, SkippedCount: 0, UnknownCount: 2,
		Results: []demoTestResult{
			{CaseIdx: 0, Status: testresult.StatusPassed, DurationMs: 9500, Comment: "Быстрая проверка входа — ОК"},
			{CaseIdx: 1, Status: testresult.StatusUnknown, DurationMs: 0, Comment: ""},
			{CaseIdx: 2, Status: testresult.StatusUnknown, DurationMs: 0, Comment: ""},
		},
	},
}

// --- Handlers ---

// Status возвращает 200 если dev-режим активен. Используется фронтендом для определения доступности.
func (h *Handler) Status(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"dev_mode": true})
}

// Seed создаёт полный набор демо-данных: пользователей, проект, workflow-статусы,
// задачи с иерархией и связями, метки, доски, тест-кейсы и каналы мессенджера.
// Идемпотентный — при повторном вызове пропускает уже созданные данные.
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

	// --- Расширенный seed: статусы, задачи, метки, доски, тесты, каналы ---

	summary := fiber.Map{}

	// Workflow-статусы
	statusMap, err := h.seedWorkflowStatuses(ctx, proj.ID)
	if err != nil {
		slog.Error("dev seed: ошибка создания статусов", "error", err)
	} else {
		summary["statuses"] = len(statusMap)
	}

	// Workflow-переходы (any-to-any для Kanban)
	transCount, err := h.seedWorkflowTransitions(ctx, proj.ID, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания переходов", "error", err)
	} else {
		summary["transitions"] = transCount
	}

	// Метки
	labelCount, err := h.seedLabels(ctx, proj.ID)
	if err != nil {
		slog.Error("dev seed: ошибка создания меток", "error", err)
	} else {
		summary["labels"] = labelCount
	}

	// Задачи
	issueMap, err := h.seedIssues(ctx, proj.ID, userIDs, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания задач", "error", err)
	} else {
		summary["issues"] = len(issueMap)
	}

	// Связи между задачами
	linkCount, err := h.seedIssueLinks(ctx, issueMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания связей", "error", err)
	} else {
		summary["issue_links"] = linkCount
	}

	// Доски
	boardCount, err := h.seedBoards(ctx, proj.ID, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания досок", "error", err)
	} else {
		summary["boards"] = boardCount
	}

	// Тест-кейсы
	tcCount, err := h.seedTestCases(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания тест-кейсов", "error", err)
	} else {
		summary["test_cases"] = tcCount
	}

	// Тестовые прогоны с результатами
	runCount, err := h.seedTestRuns(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания тестовых прогонов", "error", err)
	} else {
		summary["test_runs"] = runCount
	}

	// Каналы мессенджера
	chCount, err := h.seedChannels(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания каналов", "error", err)
	} else {
		summary["channels"] = chCount
	}

	slog.Info("dev seed: демо-данные созданы",
		"users", len(createdUsers), "project", proj.Key,
		"issues", len(issueMap), "statuses", len(statusMap),
	)

	return c.JSON(fiber.Map{
		"users":   createdUsers,
		"project": fiber.Map{"id": proj.ID.String(), "name": proj.Name, "key": proj.Key},
		"seeded":  summary,
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

// --- Seed helpers ---

// seedWorkflowStatuses создаёт workflow-статусы для проекта DEMO.
// Возвращает map[statusName] -> statusID. Идемпотентный.
func (h *Handler) seedWorkflowStatuses(ctx context.Context, projectID uuid.UUID) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID)

	// Проверяем существующие статусы
	existing, err := h.db.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("запрос статусов: %w", err)
	}

	existingNames := make(map[string]uuid.UUID)
	for _, st := range existing {
		existingNames[st.Name] = st.ID
	}

	for _, ds := range demoStatuses {
		if id, ok := existingNames[ds.Name]; ok {
			result[ds.Name] = id
			continue
		}

		st, err := h.db.WorkflowStatus.Create().
			SetName(ds.Name).
			SetColor(ds.Color).
			SetCategory(ds.Category).
			SetSortOrder(ds.Order).
			SetProjectID(projectID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("создание статуса %q: %w", ds.Name, err)
		}
		result[ds.Name] = st.ID
	}

	return result, nil
}

// seedWorkflowTransitions создаёт переходы между всеми статусами (any-to-any).
// Позволяет перемещать задачи между любыми колонками Kanban-доски.
// Идемпотентный — пропускает уже существующие переходы.
func (h *Handler) seedWorkflowTransitions(ctx context.Context, projectID uuid.UUID, statusMap map[string]uuid.UUID) (int, error) {
	if len(statusMap) < 2 {
		return 0, nil
	}

	// Проверяем, есть ли уже переходы для этого проекта
	existingCount, err := h.db.WorkflowTransition.Query().
		Where(
			workflowtransition.HasFromStatusWith(
				workflowstatus.HasProjectWith(project.ID(projectID)),
			),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос переходов: %w", err)
	}
	if existingCount > 0 {
		return existingCount, nil
	}

	// Собираем все ID статусов
	statusIDs := make([]uuid.UUID, 0, len(statusMap))
	for _, id := range statusMap {
		statusIDs = append(statusIDs, id)
	}

	// Создаём переходы между всеми статусами (any-to-any для Kanban)
	created := 0
	for _, from := range statusIDs {
		for _, to := range statusIDs {
			if from != to {
				_, err := h.db.WorkflowTransition.Create().
					SetFromStatusID(from).
					SetToStatusID(to).
					Save(ctx)
				if err != nil {
					slog.Error("dev seed: ошибка создания перехода", "from", from, "to", to, "error", err)
					continue
				}
				created++
			}
		}
	}

	return created, nil
}

// seedLabels создаёт метки для проекта DEMO. Идемпотентный.
func (h *Handler) seedLabels(ctx context.Context, projectID uuid.UUID) (int, error) {
	existing, err := h.db.Label.Query().
		Where(label.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос меток: %w", err)
	}

	existingNames := make(map[string]bool)
	for _, l := range existing {
		existingNames[l.Name] = true
	}

	created := 0
	for _, dl := range demoLabels {
		if existingNames[dl.Name] {
			continue
		}

		_, err := h.db.Label.Create().
			SetName(dl.Name).
			SetColor(dl.Color).
			SetProjectID(projectID).
			Save(ctx)
		if err != nil {
			return created, fmt.Errorf("создание метки %q: %w", dl.Name, err)
		}
		created++
	}

	return len(existingNames) + created, nil
}

// seedIssues создаёт демо-задачи для проекта DEMO.
// Возвращает map[issueKey] -> issueID. Идемпотентный.
func (h *Handler) seedIssues(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID, statusMap map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID)

	if len(statusMap) == 0 {
		return result, fmt.Errorf("нет статусов для создания задач")
	}

	// Проверяем существующие задачи проекта
	existingIssues, err := h.db.Issue.Query().
		Where(issue.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("запрос задач: %w", err)
	}

	existingKeys := make(map[string]uuid.UUID)
	for _, iss := range existingIssues {
		existingKeys[iss.IssueKey] = iss.ID
	}

	// Если задачи уже существуют — пропускаем (идемпотентность)
	if len(existingKeys) > 0 {
		for _, di := range demoIssues {
			if id, ok := existingKeys[di.Key]; ok {
				result[di.Key] = id
			}
		}
		return result, nil
	}

	// Получаем PM как reporter по умолчанию
	reporterID := userIDs["pm@hub.local"]

	// Первый проход: создаём все задачи без parent
	for i, di := range demoIssues {
		statusID, ok := statusMap[di.StatusName]
		if !ok {
			slog.Warn("dev seed: статус не найден для задачи", "key", di.Key, "status", di.StatusName)
			continue
		}

		assigneeID, ok := userIDs[di.AssigneeEmail]
		if !ok {
			slog.Warn("dev seed: assignee не найден", "key", di.Key, "email", di.AssigneeEmail)
			continue
		}

		builder := h.db.Issue.Create().
			SetIssueNumber(i + 1).
			SetIssueKey(di.Key).
			SetTitle(di.Title).
			SetIssueType(di.IssueType).
			SetPriority(di.Priority).
			SetProjectID(projectID).
			SetStatusID(statusID).
			SetAssigneeID(assigneeID).
			SetReporterID(reporterID)

		created, err := builder.Save(ctx)
		if err != nil {
			return result, fmt.Errorf("создание задачи %s: %w", di.Key, err)
		}
		result[di.Key] = created.ID
	}

	// Второй проход: устанавливаем parent для задач с ParentKey
	for _, di := range demoIssues {
		if di.ParentKey == "" {
			continue
		}
		childID, ok := result[di.Key]
		if !ok {
			continue
		}
		parentID, ok := result[di.ParentKey]
		if !ok {
			slog.Warn("dev seed: parent не найден", "child", di.Key, "parent", di.ParentKey)
			continue
		}
		err := h.db.Issue.UpdateOneID(childID).SetParentID(parentID).Exec(ctx)
		if err != nil {
			slog.Error("dev seed: ошибка установки parent", "child", di.Key, "parent", di.ParentKey, "error", err)
		}
	}

	return result, nil
}

// seedIssueLinks создаёт связи между задачами. Идемпотентный.
func (h *Handler) seedIssueLinks(ctx context.Context, issueMap map[string]uuid.UUID) (int, error) {
	if len(issueMap) == 0 {
		return 0, nil
	}

	created := 0
	for _, dl := range demoLinks {
		sourceID, ok := issueMap[dl.SourceKey]
		if !ok {
			continue
		}
		targetID, ok := issueMap[dl.TargetKey]
		if !ok {
			continue
		}

		// Проверяем, не существует ли уже
		exists, err := h.db.IssueLink.Query().
			Where(
				issuelink.HasSourceWith(issue.ID(sourceID)),
				issuelink.HasTargetWith(issue.ID(targetID)),
			).
			Exist(ctx)
		if err != nil {
			slog.Error("dev seed: ошибка проверки связи", "error", err)
			continue
		}
		if exists {
			created++
			continue
		}

		_, err = h.db.IssueLink.Create().
			SetSourceID(sourceID).
			SetTargetID(targetID).
			SetLinkType(dl.LinkType).
			Save(ctx)
		if err != nil {
			slog.Error("dev seed: ошибка создания связи", "source", dl.SourceKey, "target", dl.TargetKey, "error", err)
			continue
		}
		created++
	}

	return created, nil
}

// seedBoards создаёт Kanban-доски для проекта DEMO. Идемпотентный.
func (h *Handler) seedBoards(ctx context.Context, projectID uuid.UUID, statusMap map[string]uuid.UUID) (int, error) {
	existing, err := h.db.Board.Query().
		Where(board.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос досок: %w", err)
	}
	if len(existing) > 0 {
		return len(existing), nil
	}

	// Доска 1: Основная (все статусы)
	_, err = h.db.Board.Create().
		SetName("Основная доска").
		SetProjectID(projectID).
		SetIsDefault(true).
		SetFilters(map[string]interface{}{}).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("создание основной доски: %w", err)
	}

	// Доска 2: Спринт 1 (только определённые статусы)
	sprintStatusIDs := make([]string, 0)
	for _, name := range []string{"К выполнению", "В работе", "На ревью", "Готово"} {
		if id, ok := statusMap[name]; ok {
			sprintStatusIDs = append(sprintStatusIDs, id.String())
		}
	}

	filters := map[string]interface{}{}
	if len(sprintStatusIDs) > 0 {
		filters["status"] = sprintStatusIDs
	}

	_, err = h.db.Board.Create().
		SetName("Спринт 1").
		SetProjectID(projectID).
		SetFilters(filters).
		Save(ctx)
	if err != nil {
		return 1, fmt.Errorf("создание доски спринта: %w", err)
	}

	return 2, nil
}

// seedTestCases создаёт демо тест-кейсы. Идемпотентный.
func (h *Handler) seedTestCases(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID) (int, error) {
	existing, err := h.db.TestCase.Query().
		Where(testcase.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос тест-кейсов: %w", err)
	}
	if len(existing) > 0 {
		return len(existing), nil
	}

	testerID := userIDs["tester1@hub.local"]

	for i, tc := range demoTestCases {
		_, err := h.db.TestCase.Create().
			SetCaseNumber(i + 1).
			SetCaseKey(tc.CaseKey).
			SetTitle(tc.Title).
			SetDescription(tc.Description).
			SetCaseType(testcase.CaseTypeClassic).
			SetPriority(testcase.PriorityHigh).
			SetStatus(testcase.StatusActive).
			SetSteps(tc.Steps).
			SetProjectID(projectID).
			SetCreatedByID(testerID).
			Save(ctx)
		if err != nil {
			return i, fmt.Errorf("создание тест-кейса %s: %w", tc.CaseKey, err)
		}
	}

	return len(demoTestCases), nil
}

// seedTestRuns создаёт демо-прогоны тестов с результатами. Идемпотентный.
// Для каждого прогона создаёт TestRun и набор TestResult, привязанных к тест-кейсам.
func (h *Handler) seedTestRuns(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID) (int, error) {
	// Проверяем, есть ли уже прогоны для проекта
	existingCount, err := h.db.TestRun.Query().
		Where(testrun.HasProjectWith(project.ID(projectID))).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос прогонов: %w", err)
	}
	if existingCount > 0 {
		return existingCount, nil
	}

	// Загружаем тест-кейсы проекта для привязки результатов
	cases, err := h.db.TestCase.Query().
		Where(testcase.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос тест-кейсов для прогонов: %w", err)
	}
	if len(cases) == 0 {
		return 0, nil
	}

	testerID := userIDs["tester1@hub.local"]
	now := time.Now()

	created := 0
	for _, dr := range demoTestRuns {
		// Определяем временные метки в зависимости от статуса
		startedAt := now.Add(-2 * time.Hour)
		var completedAt *time.Time
		if dr.Status == testrun.StatusCompleted {
			t := now.Add(-1 * time.Hour)
			completedAt = &t
		}

		run, err := h.db.TestRun.Create().
			SetTitle(dr.Title).
			SetStatus(dr.Status).
			SetRunType(dr.RunType).
			SetTotalCount(dr.TotalCount).
			SetPassedCount(dr.PassedCount).
			SetFailedCount(dr.FailedCount).
			SetSkippedCount(dr.SkippedCount).
			SetBrokenCount(dr.BrokenCount).
			SetUnknownCount(dr.UnknownCount).
			SetStartedAt(startedAt).
			SetNillableCompletedAt(completedAt).
			SetProjectID(projectID).
			SetCreatedByID(testerID).
			Save(ctx)
		if err != nil {
			return created, fmt.Errorf("создание прогона %q: %w", dr.Title, err)
		}

		// Создаём результаты для прогона
		for _, res := range dr.Results {
			if res.CaseIdx >= len(cases) {
				continue
			}
			tc := cases[res.CaseIdx]

			builder := h.db.TestResult.Create().
				SetStatus(res.Status).
				SetComment(res.Comment).
				SetRunID(run.ID).
				SetTestCaseID(tc.ID).
				SetExecutedByID(testerID)

			if res.DurationMs > 0 {
				builder.SetDurationMs(res.DurationMs)
			}

			// Для завершённых результатов ставим временные метки
			if res.Status != testresult.StatusUnknown {
				resultStart := startedAt.Add(time.Duration(res.CaseIdx) * 30 * time.Second)
				resultEnd := resultStart.Add(time.Duration(res.DurationMs) * time.Millisecond)
				builder.SetStartedAt(resultStart).SetCompletedAt(resultEnd)
			}

			_, err := builder.Save(ctx)
			if err != nil {
				slog.Error("dev seed: ошибка создания результата теста",
					"run", dr.Title, "case", tc.CaseKey, "error", err)
			}
		}

		created++
	}

	return created, nil
}

// seedChannels создаёт демо-каналы мессенджера. Идемпотентный.
func (h *Handler) seedChannels(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID) (int, error) {
	existing, err := h.db.Channel.Query().
		Where(channel.ProjectIDEQ(projectID)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос каналов: %w", err)
	}

	existingNames := make(map[string]bool)
	for _, ch := range existing {
		existingNames[ch.Name] = true
	}

	pmID := userIDs["pm@hub.local"]
	created := 0

	for _, dc := range demoChannels {
		if existingNames[dc.Name] {
			created++
			continue
		}

		ch, err := h.db.Channel.Create().
			SetChannelType(dc.Type).
			SetName(dc.Name).
			SetDescription(dc.Description).
			SetIsReadonly(dc.IsReadonly).
			SetProjectID(projectID).
			SetCreatedBy(pmID).
			Save(ctx)
		if err != nil {
			slog.Error("dev seed: ошибка создания канала", "name", dc.Name, "error", err)
			continue
		}

		// Добавляем создателя как owner
		_, err = h.db.ChannelMember.Create().
			SetChannelID(ch.ID).
			SetUserID(pmID).
			SetRole(channelmember.RoleOwner).
			Save(ctx)
		if err != nil {
			slog.Error("dev seed: ошибка добавления owner в канал", "name", dc.Name, "error", err)
		}

		// Добавляем участников
		for _, email := range dc.MemberEmails {
			uid, ok := userIDs[email]
			if !ok || uid == pmID {
				continue
			}
			_, err = h.db.ChannelMember.Create().
				SetChannelID(ch.ID).
				SetUserID(uid).
				SetRole(channelmember.RoleMember).
				Save(ctx)
			if err != nil {
				slog.Error("dev seed: ошибка добавления участника в канал", "name", dc.Name, "email", email, "error", err)
			}
		}

		created++
	}

	return len(existingNames) + created, nil
}

// --- Core helpers ---

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
