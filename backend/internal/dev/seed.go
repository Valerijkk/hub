//go:build !release

package dev

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/board"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
)

// Seed создаёт полный набор демо-данных: пользователи, проект, статусы workflow,
// задачи с иерархией и связями, метки, доски, тест-кейсы и каналы мессенджера.
// Идемпотентно — при повторных вызовах пропускает уже созданные данные.
// @Summary Заполнить БД демо-данными
// @Description Создаёт демо-пользователей, проект, задачи, доски, метки, тест-кейсы и каналы. Идемпотентно
// @Tags Разработка
// @Produce json
// @Success 200 {object} object "Созданные демо-данные"
// @Failure 500 {object} map[string]string "Ошибка сидирования"
// @Router /dev/seed [post]
func (h *Handler) Seed(c fiber.Ctx) error {
	h.seedMu.Lock()
	defer h.seedMu.Unlock()

	ctx := c.Context()

	// Создать или найти пользователей
	createdUsers := make([]DevUserInfo, 0, len(demoUsers))
	userIDs := make(map[string]uuid.UUID) // email -> ID пользователя

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

	// Создать или найти демо-проект
	proj, err := h.findOrCreateDemoProject(ctx)
	if err != nil {
		slog.Error("dev seed: не удалось создать проект", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to seed demo project")
	}

	// Назначить роли в проекте
	partial := false
	for _, du := range demoUsers {
		if !du.AddToProject || du.ProjectRole == "" {
			continue
		}
		uid := userIDs[du.Email]
		if err := h.ensureProjectMember(ctx, proj.ID, uid, du.ProjectRole); err != nil {
			slog.Error("dev seed: не удалось добавить в проект", "email", du.Email, "error", err)
			partial = true
		}
	}

	// --- Расширенное сидирование: статусы, задачи, метки, доски, тесты, каналы ---

	summary := fiber.Map{}

	// Статусы workflow
	statusMap, err := h.seedWorkflowStatuses(ctx, proj.ID)
	if err != nil {
		slog.Error("dev seed: ошибка создания статусов", "error", err)
		partial = true
	} else {
		summary["statuses"] = len(statusMap)
	}

	// Переходы workflow (любой-в-любой для Kanban)
	transCount, err := h.seedWorkflowTransitions(ctx, proj.ID, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания переходов", "error", err)
		partial = true
	} else {
		summary["transitions"] = transCount
	}

	// Метки
	labelCount, err := h.seedLabels(ctx, proj.ID)
	if err != nil {
		slog.Error("dev seed: ошибка создания меток", "error", err)
		partial = true
	} else {
		summary["labels"] = labelCount
	}

	// Задачи
	issueMap, err := h.seedIssues(ctx, proj.ID, userIDs, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания задач", "error", err)
		partial = true
	} else {
		summary["issues"] = len(issueMap)
	}

	// Связи задач
	linkCount, err := h.seedIssueLinks(ctx, issueMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания связей", "error", err)
		partial = true
	} else {
		summary["issue_links"] = linkCount
	}

	// Доски
	boardCount, err := h.seedBoards(ctx, proj.ID, statusMap)
	if err != nil {
		slog.Error("dev seed: ошибка создания досок", "error", err)
		partial = true
	} else {
		summary["boards"] = boardCount
	}

	// Тест-кейсы
	tcCount, err := h.seedTestCases(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания тест-кейсов", "error", err)
		partial = true
	} else {
		summary["test_cases"] = tcCount
	}

	// Тестовые прогоны с результатами
	runCount, err := h.seedTestRuns(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания тестовых прогонов", "error", err)
		partial = true
	} else {
		summary["test_runs"] = runCount
	}

	// Каналы мессенджера
	chCount, err := h.seedChannels(ctx, proj.ID, userIDs)
	if err != nil {
		slog.Error("dev seed: ошибка создания каналов", "error", err)
		partial = true
	} else {
		summary["channels"] = chCount
	}

	// AI-провайдер (локальный Ollama — опционально, для @hub-ai / Tab suggest в dev)
	if aiSeeded, err := h.seedAIProvider(ctx); err != nil {
		slog.Error("dev seed: ошибка AI провайдера", "error", err)
		partial = true
	} else if aiSeeded {
		summary["ai_provider"] = "ollama-dev"
	}

	slog.Info("dev seed: демо-данные созданы",
		"users", len(createdUsers), "project", proj.Key,
		"issues", len(issueMap), "statuses", len(statusMap),
		"partial", partial,
	)

	body := fiber.Map{
		"users":   createdUsers,
		"project": fiber.Map{"id": proj.ID.String(), "name": proj.Name, "key": proj.Key},
		"seeded":  summary,
	}
	// P3-BE-16 / W7-A5-013: частичный seed → non-2xx + partial, не «успешный» 200.
	if partial {
		body["partial"] = true
		body["error"] = "seed incomplete"
		return c.Status(fiber.StatusInternalServerError).JSON(body)
	}
	if h.mailCalendarSeeder != nil {
		copiedUserIDs := make(map[string]uuid.UUID, len(userIDs))
		for email, userID := range userIDs {
			copiedUserIDs[email] = userID
		}
		mailCalendarSummary, err := h.mailCalendarSeeder.SeedMailCalendar(ctx, MailCalendarDemoSeedInput{
			ProjectID: proj.ID,
			UserIDs:   copiedUserIDs,
		})
		if err != nil {
			slog.Error("dev seed: Mail/Calendar demo failed", "error", err)
			body["partial"] = true
			body["error"] = "seed incomplete"
			return c.Status(fiber.StatusInternalServerError).JSON(body)
		}
		if mailCalendarSummary == nil {
			mailCalendarSummary = map[string]any{}
		}
		summary["mail_calendar"] = mailCalendarSummary
	}
	return c.JSON(body)
}

// --- Вспомогательные функции сидирования ---

// seedWorkflowStatuses создаёт статусы workflow для проекта DEMO.
// Возвращает map[statusName] -> statusID. Идемпотентно.
func (h *Handler) seedWorkflowStatuses(ctx context.Context, projectID uuid.UUID) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID)

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

// seedWorkflowTransitions создаёт переходы между всеми статусами (любой-в-любой).
// Идемпотентно — пропускает уже существующие переходы.
func (h *Handler) seedWorkflowTransitions(ctx context.Context, projectID uuid.UUID, statusMap map[string]uuid.UUID) (int, error) {
	if len(statusMap) < 2 {
		return 0, nil
	}

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

	statusIDs := make([]uuid.UUID, 0, len(statusMap))
	for _, id := range statusMap {
		statusIDs = append(statusIDs, id)
	}

	created := 0
	for _, from := range statusIDs {
		for _, to := range statusIDs {
			if from != to {
				_, err := h.db.WorkflowTransition.Create().
					SetFromStatusID(from).
					SetToStatusID(to).
					Save(ctx)
				if err != nil {
					// A3W7-5-002 / W7-A5-013: DB-err → propagate (не slog+continue → silent 200).
					return created, fmt.Errorf("создание перехода workflow: %w", err)
				}
				created++
			}
		}
	}

	return created, nil
}

// seedLabels создаёт метки для проекта DEMO. Идемпотентно.
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
// Возвращает map[issueKey] -> issueID. Идемпотентно.
func (h *Handler) seedIssues(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID, statusMap map[string]uuid.UUID) (map[string]uuid.UUID, error) {
	result := make(map[string]uuid.UUID)

	if len(statusMap) == 0 {
		return result, fmt.Errorf("нет статусов для создания задач")
	}

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

	// Если задачи уже существуют — пропустить (идемпотентно)
	if len(existingKeys) > 0 {
		for _, di := range demoIssues {
			if id, ok := existingKeys[di.Key]; ok {
				result[di.Key] = id
			}
		}
		return result, nil
	}

	reporterID := userIDs["pm@hub.local"]

	// Первый проход: создать все задачи без родителя
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

	// Второй проход: установить родителя для задач с ParentKey
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

// seedIssueLinks создаёт связи задач. Идемпотентно.
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

		exists, err := h.db.IssueLink.Query().
			Where(
				issuelink.HasSourceWith(issue.ID(sourceID)),
				issuelink.HasTargetWith(issue.ID(targetID)),
			).
			Exist(ctx)
		if err != nil {
			// A3W7-5-002: DB-err → propagate.
			return created, fmt.Errorf("проверка связи задач: %w", err)
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
			// A3W7-5-002 / W7-A5-013: DB-err → propagate (не silent skip).
			return created, fmt.Errorf("создание связи задач %s→%s: %w", dl.SourceKey, dl.TargetKey, err)
		}
		created++
	}

	return created, nil
}

// seedBoards создаёт Kanban-доски для проекта DEMO. Идемпотентно.
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

	_, err = h.db.Board.Create().
		SetName("Основная доска").
		SetProjectID(projectID).
		SetIsDefault(true).
		SetFilters(map[string]interface{}{}).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("создание основной доски: %w", err)
	}

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

// seedTestCases создаёт демо тест-кейсы. Идемпотентно.
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

// seedTestRuns создаёт демо тестовые прогоны с результатами. Идемпотентно.
func (h *Handler) seedTestRuns(ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID) (int, error) {
	existingCount, err := h.db.TestRun.Query().
		Where(testrun.HasProjectWith(project.ID(projectID))).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("запрос прогонов: %w", err)
	}
	if existingCount > 0 {
		return existingCount, nil
	}

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

// seedChannels создаёт демо-каналы мессенджера. Идемпотентно.
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
			// A3W7-5-002 / W7-A5-013: DB-err → propagate (не silent skip).
			return created, fmt.Errorf("создание канала %s: %w", dc.Name, err)
		}

		_, err = h.db.ChannelMember.Create().
			SetChannelID(ch.ID).
			SetUserID(pmID).
			SetRole(channelmember.RoleOwner).
			Save(ctx)
		if err != nil {
			return created, fmt.Errorf("добавление owner в канал %s: %w", dc.Name, err)
		}

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
				return created, fmt.Errorf("добавление участника %s в канал %s: %w", email, dc.Name, err)
			}
		}

		created++
	}

	return len(existingNames) + created, nil
}

// seedAIProvider создаёт dev Ollama-провайдер, если ни одного нет (идемпотентно).
func (h *Handler) seedAIProvider(ctx context.Context) (bool, error) {
	n, err := h.db.AIProvider.Query().Count(ctx)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	_, err = h.db.AIProvider.Create().
		SetName("Dev Ollama").
		SetProviderType(aiprovider.ProviderTypeOllama).
		SetBaseURL("http://127.0.0.1:11434/v1").
		SetModels([]string{"llama3.2", "llama3"}).
		SetEnabled(true).
		SetPriority(0).
		SetMaxRpm(120).
		SetRuResident(true).
		SetScope(aiprovider.ScopeSystem).
		Save(ctx)
	if err != nil {
		return false, fmt.Errorf("create ai provider: %w", err)
	}
	return true, nil
}

// --- Базовые вспомогательные функции ---

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
