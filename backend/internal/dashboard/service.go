package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/dashboardlayout"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// projectStatsTTL время жизни кеша статистики проекта.
const projectStatsTTL = 5 * time.Minute

// DashboardService предоставляет агрегированные данные для дашбордов.
type DashboardService struct {
	db      *ent.Client
	rdb     *redis.Client
	mongoDB *mongo.Database
}

// NewDashboardService создаёт новый сервис дашбордов.
func NewDashboardService(db *ent.Client, rdb *redis.Client, mongoDB *mongo.Database) *DashboardService {
	return &DashboardService{db: db, rdb: rdb, mongoDB: mongoDB}
}

// priorityRank возвращает числовой приоритет для сортировки (чем выше — тем важнее).
func priorityRank(p issue.Priority) int {
	switch p {
	case issue.PriorityCritical:
		return 5
	case issue.PriorityHigh:
		return 4
	case issue.PriorityMedium:
		return 3
	case issue.PriorityLow:
		return 2
	case issue.PriorityTrivial:
		return 1
	default:
		return 0
	}
}

// GetPersonalDashboardData возвращает агрегированные данные персонального дашборда.
func (s *DashboardService) GetPersonalDashboardData(ctx context.Context, userID uuid.UUID) (*PersonalData, error) {
	data := &PersonalData{
		MyTasks:       make([]TaskSummary, 0),
		UpcomingCalls: make([]CallSummary, 0),
		RecentRuns:    make([]RunSummary, 0),
		Deadlines:     make([]TaskSummary, 0),
	}

	// Мои задачи: назначены на меня, не завершены, до 20 штук
	issues, err := s.db.Issue.Query().
		Where(
			issue.HasAssigneeWith(user.ID(userID)),
			issue.HasStatusWith(workflowstatus.NameNotIn("done", "closed")),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		Order(ent.Desc(issue.FieldPriority)).
		Limit(20).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки задач", "error", err)
	} else {
		for _, iss := range issues {
			ts := TaskSummary{
				ID:       iss.ID,
				Key:      iss.IssueKey,
				Title:    iss.Title,
				Priority: priorityRank(iss.Priority),
			}
			if iss.Edges.Status != nil {
				ts.Status = iss.Edges.Status.Name
			}
			if iss.DueDate != nil {
				ts.DueDate = iss.DueDate
			}
			data.MyTasks = append(data.MyTasks, ts)
		}
	}

	// Непрочитанные сообщения: Redis HGETALL
	unreadMap, err := s.rdb.HGetAll(ctx, fmt.Sprintf("unread:%s", userID.String())).Result()
	if err != nil && err != redis.Nil {
		slog.Error("ошибка загрузки непрочитанных", "error", err)
	}
	total := 0
	for _, v := range unreadMap {
		var n int
		fmt.Sscanf(v, "%d", &n)
		total += n
	}
	data.UnreadCount = total

	// Предстоящие звонки: последние активные звонки в каналах пользователя
	// (Calls schema не имеет scheduled_at, показываем активные)
	activeCalls, err := s.db.Call.Query().
		Where().
		WithChannel().
		Limit(10).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки звонков", "error", err)
	} else {
		for _, c := range activeCalls {
			cs := CallSummary{
				ID:          c.ID,
				ScheduledAt: c.CreatedAt,
			}
			if c.Edges.Channel != nil {
				cs.ChannelName = c.Edges.Channel.Name
				cs.Title = c.Edges.Channel.Name
			} else {
				cs.Title = c.RoomName
			}
			data.UpcomingCalls = append(data.UpcomingCalls, cs)
		}
	}

	// Последние прогоны тестов
	runs, err := s.db.TestRun.Query().
		Where(testrun.HasCreatedByWith(user.ID(userID))).
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Limit(5).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки прогонов", "error", err)
	} else {
		for _, r := range runs {
			data.RecentRuns = append(data.RecentRuns, RunSummary{
				ID:        r.ID,
				Name:      r.Title,
				Status:    string(r.Status),
				Passed:    r.PassedCount,
				Failed:    r.FailedCount,
				Skipped:   r.SkippedCount,
				CreatedAt: r.CreatedAt,
			})
		}
	}

	// Дедлайны: задачи с due_date в ближайшие 7 дней
	now := time.Now()
	weekLater := now.Add(7 * 24 * time.Hour)
	deadlineIssues, err := s.db.Issue.Query().
		Where(
			issue.HasAssigneeWith(user.ID(userID)),
			issue.DueDateGTE(now),
			issue.DueDateLTE(weekLater),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		Order(ent.Asc(issue.FieldDueDate)).
		Limit(10).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки дедлайнов", "error", err)
	} else {
		for _, iss := range deadlineIssues {
			ts := TaskSummary{
				ID:       iss.ID,
				Key:      iss.IssueKey,
				Title:    iss.Title,
				Priority: priorityRank(iss.Priority),
			}
			if iss.Edges.Status != nil {
				ts.Status = iss.Edges.Status.Name
			}
			if iss.DueDate != nil {
				ts.DueDate = iss.DueDate
			}
			data.Deadlines = append(data.Deadlines, ts)
		}
	}

	// Время за сегодня: сумма time_logs
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	timeLogs, err := s.db.TimeLog.Query().
		Where(
			timelog.HasUserWith(user.ID(userID)),
			timelog.CreatedAtGTE(today),
			timelog.CreatedAtLTE(today.Add(24*time.Hour)),
		).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки тайм-логов", "error", err)
	} else {
		for _, tl := range timeLogs {
			data.TimeToday += tl.Minutes
		}
	}

	return data, nil
}

// GetProjectStats возвращает статистику проекта с кешированием в Redis.
func (s *DashboardService) GetProjectStats(ctx context.Context, projectID uuid.UUID) (*ProjectStats, error) {
	cacheKey := fmt.Sprintf("dashboard:project:%s", projectID.String())

	// Попытка чтения из кеша
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var stats ProjectStats
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	stats := &ProjectStats{
		TasksByStatus: make(map[string]int),
		BurndownData:  make([]DataPoint, 0),
		VelocityData:  make([]DataPoint, 0),
		TestTrends:    make([]DataPoint, 0),
		TeamActivity:  make([]ActivityEntry, 0),
	}

	// Задачи по статусам
	issues, err := s.db.Issue.Query().
		Where(
			issue.HasProjectWith(project.ID(projectID)),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		WithAssignee().
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки задач проекта", "error", err)
	} else {
		now := time.Now()
		for _, iss := range issues {
			statusName := "unknown"
			if iss.Edges.Status != nil {
				statusName = iss.Edges.Status.Name
			}
			stats.TasksByStatus[statusName]++

			// Баги: тип bug, не завершены
			if iss.IssueType == issue.IssueTypeBug && statusName != "done" && statusName != "closed" {
				stats.BugCount++
			}

			// Без исполнителя
			if iss.Edges.Assignee == nil {
				stats.Unassigned++
			}

			// Просроченные
			if iss.DueDate != nil && iss.DueDate.Before(now) && statusName != "done" && statusName != "closed" {
				stats.Overdue++
			}
		}
	}

	// Burndown: завершённые задачи по дням за последние 30 дней
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	completedIssues, err := s.db.Issue.Query().
		Where(
			issue.HasProjectWith(project.ID(projectID)),
			issue.HasStatusWith(workflowstatus.NameIn("done", "closed")),
			issue.UpdatedAtGTE(thirtyDaysAgo),
			issue.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки burndown", "error", err)
	} else {
		// Группируем по дням
		dayCounts := make(map[string]int)
		for _, iss := range completedIssues {
			day := iss.UpdatedAt.Format("01-02")
			dayCounts[day]++
		}
		for d := thirtyDaysAgo; d.Before(now); d = d.AddDate(0, 0, 1) {
			day := d.Format("01-02")
			stats.BurndownData = append(stats.BurndownData, DataPoint{
				Label: day,
				Value: float64(dayCounts[day]),
			})
		}
	}

	// Velocity: завершённые задачи по неделям за последние 10 недель
	for w := 9; w >= 0; w-- {
		weekStart := now.AddDate(0, 0, -7*(w+1))
		weekEnd := now.AddDate(0, 0, -7*w)
		count, err := s.db.Issue.Query().
			Where(
				issue.HasProjectWith(project.ID(projectID)),
				issue.HasStatusWith(workflowstatus.NameIn("done", "closed")),
				issue.UpdatedAtGTE(weekStart),
				issue.UpdatedAtLT(weekEnd),
				issue.DeletedAtIsNil(),
			).
			Count(ctx)
		if err != nil {
			continue
		}
		stats.VelocityData = append(stats.VelocityData, DataPoint{
			Label: fmt.Sprintf("W%d", 10-w),
			Value: float64(count),
		})
	}

	// Тренды тестов: последние 10 прогонов
	testRuns, err := s.db.TestRun.Query().
		Where(testrun.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Limit(10).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки прогонов проекта", "error", err)
	} else {
		for i := len(testRuns) - 1; i >= 0; i-- {
			r := testRuns[i]
			totalR := r.PassedCount + r.FailedCount + r.SkippedCount
			passRate := float64(0)
			if totalR > 0 {
				passRate = float64(r.PassedCount) / float64(totalR) * 100
			}
			stats.TestTrends = append(stats.TestTrends, DataPoint{
				Label: r.Title,
				Value: passRate,
			})
		}
	}

	// Активность команды: последние 20 записей из activity_logs проекта
	logs, err := s.db.ActivityLog.Query().
		Where(activitylog.HasIssueWith(issue.HasProjectWith(project.ID(projectID)))).
		WithUser().
		Order(ent.Desc(activitylog.FieldCreatedAt)).
		Limit(20).
		All(ctx)
	if err != nil {
		slog.Error("ошибка загрузки активности", "error", err)
	} else {
		for _, al := range logs {
			entry := ActivityEntry{
				Action:    al.Action,
				CreatedAt: al.CreatedAt,
			}
			if al.Edges.User != nil {
				name := al.Edges.User.Name
				if name == "" {
					name = al.Edges.User.Email
				}
				entry.UserName = name
			}
			stats.TeamActivity = append(stats.TeamActivity, entry)
		}
	}

	// Кешируем результат
	if data, err := json.Marshal(stats); err == nil {
		s.rdb.Set(ctx, cacheKey, string(data), projectStatsTTL)
	}

	return stats, nil
}

// GetAdminStats возвращает статистику для админского дашборда.
func (s *DashboardService) GetAdminStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{
		ServiceHealth: make(map[string]string),
		TopProjects:   make([]ProjectRank, 0),
	}

	// Пользователи
	totalUsers, err := s.db.User.Query().Count(ctx)
	if err == nil {
		stats.TotalUsers = totalUsers
	}

	// Активные пользователи (последние 30 дней по сессиям)
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	activeUsers, err := s.db.Session.Query().
		Where(session.CreatedAtGTE(thirtyDaysAgo)).
		QueryUser().
		GroupBy(user.FieldID).
		Strings(ctx)
	if err == nil {
		stats.ActiveUsers = len(activeUsers)
	}

	// Проекты
	totalProjects, err := s.db.Project.Query().Count(ctx)
	if err == nil {
		stats.TotalProjects = totalProjects
	}

	// Задачи
	totalIssues, err := s.db.Issue.Query().
		Where(issue.DeletedAtIsNil()).
		Count(ctx)
	if err == nil {
		stats.TotalIssues = totalIssues
	}

	// DAU/WAU/MAU
	dayAgo := now.AddDate(0, 0, -1)
	weekAgo := now.AddDate(0, 0, -7)

	dauUsers, err := s.db.Session.Query().
		Where(session.CreatedAtGTE(dayAgo)).
		QueryUser().
		GroupBy(user.FieldID).
		Strings(ctx)
	if err == nil {
		stats.DAU = len(dauUsers)
	}

	wauUsers, err := s.db.Session.Query().
		Where(session.CreatedAtGTE(weekAgo)).
		QueryUser().
		GroupBy(user.FieldID).
		Strings(ctx)
	if err == nil {
		stats.WAU = len(wauUsers)
	}

	stats.MAU = stats.ActiveUsers

	// Проверка здоровья сервисов
	stats.ServiceHealth["postgresql"] = "healthy"

	if err := s.rdb.Ping(ctx).Err(); err != nil {
		stats.ServiceHealth["redis"] = "unhealthy"
	} else {
		stats.ServiceHealth["redis"] = "healthy"
	}

	if s.mongoDB != nil {
		if err := s.mongoDB.Client().Ping(ctx, nil); err != nil {
			stats.ServiceHealth["mongodb"] = "unhealthy"
		} else {
			stats.ServiceHealth["mongodb"] = "healthy"
		}
	}

	stats.ServiceHealth["nats"] = "unknown"
	stats.ServiceHealth["seaweedfs"] = "unknown"

	// Топ-5 проектов по количеству задач
	projects, err := s.db.Project.Query().
		WithIssues().
		All(ctx)
	if err == nil {
		type pRank struct {
			id    uuid.UUID
			name  string
			count int
		}
		ranks := make([]pRank, 0, len(projects))
		for _, p := range projects {
			ranks = append(ranks, pRank{id: p.ID, name: p.Name, count: len(p.Edges.Issues)})
		}
		// Сортируем по количеству задач
		for i := 0; i < len(ranks); i++ {
			for j := i + 1; j < len(ranks); j++ {
				if ranks[j].count > ranks[i].count {
					ranks[i], ranks[j] = ranks[j], ranks[i]
				}
			}
		}
		limit := 5
		if len(ranks) < limit {
			limit = len(ranks)
		}
		for _, r := range ranks[:limit] {
			stats.TopProjects = append(stats.TopProjects, ProjectRank{
				ID:         r.id,
				Name:       r.name,
				IssueCount: r.count,
			})
		}
	}

	return stats, nil
}

// ---------------------------------------------------------------------------
// Шаблоны дашбордов
// ---------------------------------------------------------------------------

// templatePresets содержит предустановленные раскладки дашбордов для ролей.
var templatePresets = map[string][]LayoutItem{
	"dev": {
		{I: "my-tasks", X: 0, Y: 0, W: 6, H: 4, WidgetType: "my-tasks", Config: map[string]interface{}{"limit": 20}},
		{I: "recent-activity", X: 6, Y: 0, W: 6, H: 4, WidgetType: "recent-activity", Config: map[string]interface{}{"limit": 20}},
		{I: "burndown", X: 0, Y: 4, W: 6, H: 3, WidgetType: "burndown", Config: map[string]interface{}{"days": 30}},
		{I: "velocity", X: 6, Y: 4, W: 6, H: 3, WidgetType: "velocity", Config: map[string]interface{}{"weeks": 10}},
	},
	"qa": {
		{I: "test-runs", X: 0, Y: 0, W: 6, H: 4, WidgetType: "test-runs", Config: map[string]interface{}{"limit": 10}},
		{I: "flaky-tests", X: 6, Y: 0, W: 6, H: 4, WidgetType: "flaky-tests", Config: map[string]interface{}{"limit": 10}},
		{I: "coverage", X: 0, Y: 4, W: 6, H: 3, WidgetType: "coverage", Config: map[string]interface{}{}},
		{I: "my-assigned-runs", X: 6, Y: 4, W: 6, H: 3, WidgetType: "my-assigned-runs", Config: map[string]interface{}{"limit": 10}},
	},
	"manager": {
		{I: "project-stats", X: 0, Y: 0, W: 6, H: 3, WidgetType: "project-stats", Config: map[string]interface{}{}},
		{I: "team-activity", X: 6, Y: 0, W: 6, H: 3, WidgetType: "team-activity", Config: map[string]interface{}{"limit": 20}},
		{I: "overdue", X: 0, Y: 3, W: 6, H: 3, WidgetType: "overdue", Config: map[string]interface{}{"limit": 15}},
		{I: "burndown", X: 6, Y: 3, W: 6, H: 3, WidgetType: "burndown", Config: map[string]interface{}{"days": 30}},
	},
}

// GetTemplates возвращает список доступных шаблонов дашбордов.
// Каждый шаблон — преднастроенный набор виджетов для определённой роли.
func (s *DashboardService) GetTemplates() map[string][]LayoutItem {
	// Возвращаем копию, чтобы вызывающий код не мог изменить пресеты.
	result := make(map[string][]LayoutItem, len(templatePresets))
	for name, items := range templatePresets {
		copied := make([]LayoutItem, len(items))
		copy(copied, items)
		result[name] = copied
	}
	return result
}

// ApplyTemplate применяет шаблон дашборда для пользователя в проекте.
// Перезаписывает существующую раскладку типа "project" для указанного пользователя.
func (s *DashboardService) ApplyTemplate(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, templateName string) error {
	items, ok := templatePresets[templateName]
	if !ok {
		return fmt.Errorf("шаблон %q не найден", templateName)
	}

	req := SaveLayoutRequest{
		DashboardType: "project",
		ProjectID:     &projectID,
		Items:         items,
	}
	return s.SaveLayout(ctx, userID, req)
}

// GetLayout возвращает сохранённую раскладку дашборда пользователя.
func (s *DashboardService) GetLayout(ctx context.Context, userID uuid.UUID, dashboardType string, projectID *uuid.UUID) (*LayoutDTO, error) {
	q := s.db.DashboardLayout.Query().
		Where(
			dashboardlayout.HasUserWith(user.ID(userID)),
			dashboardlayout.DashboardTypeEQ(dashboardlayout.DashboardType(dashboardType)),
		)

	if projectID != nil {
		q = q.Where(dashboardlayout.ProjectID(*projectID))
	} else {
		q = q.Where(dashboardlayout.ProjectIDIsNil())
	}

	layout, err := q.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка загрузки раскладки: %w", err)
	}

	items := make([]LayoutItem, 0, len(layout.Items))
	for _, raw := range layout.Items {
		data, _ := json.Marshal(raw)
		var item LayoutItem
		if json.Unmarshal(data, &item) == nil {
			items = append(items, item)
		}
	}

	return &LayoutDTO{
		DashboardType: dashboardType,
		ProjectID:     projectID,
		Items:         items,
	}, nil
}

// SaveLayout сохраняет или обновляет раскладку дашборда.
func (s *DashboardService) SaveLayout(ctx context.Context, userID uuid.UUID, req SaveLayoutRequest) error {
	// Преобразуем items в []map[string]interface{}
	rawItems := make([]map[string]interface{}, 0, len(req.Items))
	for _, item := range req.Items {
		data, _ := json.Marshal(item)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		rawItems = append(rawItems, m)
	}

	// Попытка найти существующую раскладку
	q := s.db.DashboardLayout.Query().
		Where(
			dashboardlayout.HasUserWith(user.ID(userID)),
			dashboardlayout.DashboardTypeEQ(dashboardlayout.DashboardType(req.DashboardType)),
		)

	if req.ProjectID != nil {
		q = q.Where(dashboardlayout.ProjectID(*req.ProjectID))
	} else {
		q = q.Where(dashboardlayout.ProjectIDIsNil())
	}

	existing, err := q.Only(ctx)
	if err == nil {
		_, err = existing.Update().
			SetItems(rawItems).
			Save(ctx)
		return err
	}

	if !ent.IsNotFound(err) {
		return fmt.Errorf("ошибка поиска раскладки: %w", err)
	}

	creator := s.db.DashboardLayout.Create().
		SetUserID(userID).
		SetDashboardType(dashboardlayout.DashboardType(req.DashboardType)).
		SetItems(rawItems)

	if req.ProjectID != nil {
		creator = creator.SetProjectID(*req.ProjectID)
	}

	_, err = creator.Save(ctx)
	return err
}
