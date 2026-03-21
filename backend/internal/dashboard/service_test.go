package dashboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLayoutItemSerialization проверяет корректность структуры LayoutItem.
func TestLayoutItemSerialization(t *testing.T) {
	item := LayoutItem{
		I:          "widget-1",
		X:          0,
		Y:          0,
		W:          6,
		H:          4,
		WidgetType: "my-tasks",
		Config:     map[string]interface{}{"limit": float64(10)},
	}

	assert.Equal(t, "widget-1", item.I)
	assert.Equal(t, "my-tasks", item.WidgetType)
	assert.Equal(t, 6, item.W)
	assert.Equal(t, 4, item.H)
	assert.Equal(t, float64(10), item.Config["limit"])
}

// TestPersonalDataDefaults проверяет значения по умолчанию PersonalData.
func TestPersonalDataDefaults(t *testing.T) {
	data := PersonalData{}

	assert.Nil(t, data.MyTasks)
	assert.Equal(t, 0, data.UnreadCount)
	assert.Nil(t, data.UpcomingCalls)
	assert.Nil(t, data.RecentRuns)
	assert.Nil(t, data.Deadlines)
	assert.Equal(t, 0, data.TimeToday)
}

// TestProjectStatsDefaults проверяет значения по умолчанию ProjectStats.
func TestProjectStatsDefaults(t *testing.T) {
	stats := ProjectStats{
		TasksByStatus: make(map[string]int),
	}

	assert.NotNil(t, stats.TasksByStatus)
	assert.Equal(t, 0, stats.BugCount)
	assert.Equal(t, 0, stats.Unassigned)
	assert.Equal(t, 0, stats.Overdue)
	assert.Equal(t, float64(0), stats.Coverage)
}

// TestAdminStatsServiceHealth проверяет корректность инициализации ServiceHealth.
func TestAdminStatsServiceHealth(t *testing.T) {
	stats := AdminStats{
		ServiceHealth: map[string]string{
			"postgresql": "healthy",
			"redis":      "healthy",
			"mongodb":    "healthy",
			"nats":       "unknown",
			"seaweedfs":  "unknown",
		},
	}

	assert.Equal(t, "healthy", stats.ServiceHealth["postgresql"])
	assert.Equal(t, "unknown", stats.ServiceHealth["nats"])
	assert.Len(t, stats.ServiceHealth, 5)
}

// TestSaveLayoutRequestValidation проверяет валидацию запроса на сохранение раскладки.
func TestSaveLayoutRequestValidation(t *testing.T) {
	req := SaveLayoutRequest{
		DashboardType: "personal",
		Items: []LayoutItem{
			{I: "w1", X: 0, Y: 0, W: 6, H: 4, WidgetType: "my-tasks"},
			{I: "w2", X: 6, Y: 0, W: 6, H: 3, WidgetType: "deadlines"},
		},
	}

	assert.Equal(t, "personal", req.DashboardType)
	assert.Nil(t, req.ProjectID)
	assert.Len(t, req.Items, 2)
}

// TestLayoutDTOEmptyItems проверяет, что пустая раскладка корректна.
func TestLayoutDTOEmptyItems(t *testing.T) {
	dto := LayoutDTO{
		DashboardType: "admin",
		Items:         []LayoutItem{},
	}

	assert.Empty(t, dto.Items)
	assert.Equal(t, "admin", dto.DashboardType)
	assert.Nil(t, dto.ProjectID)
}
