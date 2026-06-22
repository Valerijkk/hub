package dashboard

import (
	"time"

	"github.com/google/uuid"
)

// LayoutItem представляет один виджет в раскладке дашборда.
type LayoutItem struct {
	I          string                 `json:"i"`           // Уникальный идентификатор виджета
	X          int                    `json:"x"`           // Горизонтальная позиция
	Y          int                    `json:"y"`           // Вертикальная позиция
	W          int                    `json:"w"`           // Ширина в колонках
	H          int                    `json:"h"`           // Высота в строках
	WidgetType string                 `json:"widget_type"` // Тип виджета (my-tasks, burndown и т.д.)
	Config     map[string]interface{} `json:"config"`      // Дополнительная конфигурация виджета
}

// LayoutDTO описывает раскладку дашборда для передачи клиенту.
type LayoutDTO struct {
	DashboardType string     `json:"dashboard_type"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty"`
	Items         []LayoutItem `json:"items"`
}

// SaveLayoutRequest запрос на сохранение раскладки дашборда.
type SaveLayoutRequest struct {
	DashboardType string       `json:"dashboard_type"`
	ProjectID     *uuid.UUID   `json:"project_id,omitempty"`
	Items         []LayoutItem `json:"items"`
}

// --- Персональный дашборд ---

// TaskSummary краткая информация о задаче для виджетов дашборда.
type TaskSummary struct {
	ID       uuid.UUID `json:"id"`
	Key      string    `json:"key"`
	Title    string    `json:"title"`
	Status   string    `json:"status"`
	Priority int       `json:"priority"`
	DueDate  *time.Time `json:"due_date,omitempty"`
}

// CallSummary краткая информация о звонке.
type CallSummary struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ScheduledAt time.Time `json:"scheduled_at"`
	ChannelName string    `json:"channel_name"`
}

// RunSummary краткая информация о прогоне тестов.
type RunSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Passed    int       `json:"passed"`
	Failed    int       `json:"failed"`
	Skipped   int       `json:"skipped"`
	CreatedAt time.Time `json:"created_at"`
}

// PersonalData агрегированные данные персонального дашборда.
type PersonalData struct {
	MyTasks       []TaskSummary `json:"my_tasks"`
	UnreadCount   int           `json:"unread_count"`
	UpcomingCalls []CallSummary `json:"upcoming_calls"`
	RecentRuns    []RunSummary  `json:"recent_runs"`
	Deadlines     []TaskSummary `json:"deadlines"`
	TimeToday     int           `json:"time_today"` // Минуты за сегодня
}

// --- Проектный дашборд ---

// DataPoint точка данных для графиков.
type DataPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// ActivityEntry запись активности в проекте.
type ActivityEntry struct {
	UserName  string    `json:"user_name"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectStats агрегированные данные проектного дашборда.
type ProjectStats struct {
	TasksByStatus map[string]int  `json:"tasks_by_status"`
	BurndownData  []DataPoint     `json:"burndown_data"`
	VelocityData  []DataPoint     `json:"velocity_data"`
	BugCount      int             `json:"bug_count"`
	TeamActivity  []ActivityEntry `json:"team_activity"`
	TestTrends    []DataPoint     `json:"test_trends"`
	Coverage      float64         `json:"coverage"`
	FlakyCount    int             `json:"flaky_count"`
	Unassigned    int             `json:"unassigned"`
	Overdue       int             `json:"overdue"`
}

// --- Админский дашборд ---

// ProjectRank рейтинг проекта по количеству задач.
type ProjectRank struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IssueCount int       `json:"issue_count"`
}

// AdminStats агрегированные данные админского дашборда.
type AdminStats struct {
	TotalUsers    int               `json:"total_users"`
	ActiveUsers   int               `json:"active_users"`
	TotalProjects int               `json:"total_projects"`
	TotalIssues   int               `json:"total_issues"`
	DAU           int               `json:"dau"`
	WAU           int               `json:"wau"`
	MAU           int               `json:"mau"`
	StorageUsed   int64             `json:"storage_used"`
	ServiceHealth map[string]string `json:"service_health"`
	TopProjects   []ProjectRank     `json:"top_projects"`
}
