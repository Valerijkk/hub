//go:build !release

package dev

import (
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// --- Demo users ---

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

// --- Demo workflow statuses ---

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

// --- Demo issues ---

type demoIssue struct {
	Key           string
	IssueType     issue.IssueType
	Title         string
	StatusName    string
	AssigneeEmail string
	Priority      issue.Priority
	ParentKey     string // parent issue key (empty = no parent)
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

// --- Demo issue links ---

type demoLink struct {
	SourceKey string
	TargetKey string
	LinkType  issuelink.LinkType
}

var demoLinks = []demoLink{
	{"DEMO-4", "DEMO-2", issuelink.LinkTypeBlocks},
	{"DEMO-12", "DEMO-8", issuelink.LinkTypeRelatesTo},
}

// --- Demo labels ---

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

// --- Demo test cases ---

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

// --- Demo channels ---

type demoChannel struct {
	Name         string
	Description  string
	Type         channel.ChannelType
	IsReadonly   bool
	MemberEmails []string // member emails (empty = all project members)
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

// --- Demo test runs ---

type demoTestRun struct {
	Title        string
	Status       testrun.Status
	RunType      testrun.RunType
	TotalCount   int
	PassedCount  int
	FailedCount  int
	SkippedCount int
	BrokenCount  int
	UnknownCount int
	Results      []demoTestResult // results per test case
}

type demoTestResult struct {
	CaseIdx    int // index into demoTestCases
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
