package dashboard_test
import (
	"github.com/valerii/hub/backend/internal/dashboard"
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func newDashboardTestEnv(t *testing.T) (*dashboard.DashboardService, *ent.Client, context.Context, uuid.UUID) {
	t.Helper()

	dsn := "file:dashboardtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("dash-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	return dashboard.NewDashboardService(dashboard.NewEntRepository(client), rdb, nil), client, ctx, user.ID
}

func TestGetTemplates_ReturnsIndependentCopy(t *testing.T) {
	svc := &dashboard.DashboardService{}
	templates := svc.GetTemplates()

	require.Contains(t, templates, "dev")
	require.Contains(t, templates, "qa")
	require.Contains(t, templates, "manager")
	assert.NotEmpty(t, templates["dev"])

	templates["dev"][0].WidgetType = "mutated"
	again := svc.GetTemplates()
	assert.NotEqual(t, "mutated", again["dev"][0].WidgetType)
}

func TestApplyTemplate_UnknownTemplate(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("Tpl").SetKey("TPL").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	err = svc.ApplyTemplate(ctx, userID, proj.ID, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestSaveLayoutAndGetLayout_RoundTrip(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)

	req := dashboard.SaveLayoutRequest{
		DashboardType: "personal",
		Items: []dashboard.LayoutItem{
			{I: "w1", X: 0, Y: 0, W: 6, H: 4, WidgetType: "my-tasks", Config: map[string]interface{}{"limit": 10}},
		},
	}
	require.NoError(t, svc.SaveLayout(ctx, userID, req))

	layout, err := svc.GetLayout(ctx, userID, "personal", nil)
	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Items, 1)
	assert.Equal(t, "my-tasks", layout.Items[0].WidgetType)
	assert.Equal(t, 10, int(layout.Items[0].Config["limit"].(float64)))
}

func TestGetLayout_NotFoundReturnsNil(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)

	layout, err := svc.GetLayout(ctx, userID, "personal", nil)
	require.NoError(t, err)
	assert.Nil(t, layout)
}

func TestGetProjectStats_CacheHit(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("Cached").SetKey("CCH").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)
	projectID := proj.ID

	cached := &dashboard.ProjectStats{
		TasksByStatus: map[string]int{"todo": 3},
		BugCount:      1,
	}
	data, err := json.Marshal(cached)
	require.NoError(t, err)
	cacheKey := "dashboard:project:" + projectID.String()
	require.NoError(t, dashboard.ExportDashboardServiceRedis(svc).Set(ctx, cacheKey, string(data), dashboard.ExportProjectStatsTTL()).Err())

	stats, err := svc.GetProjectStats(ctx, userID, projectID)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TasksByStatus["todo"])
	assert.Equal(t, 1, stats.BugCount)
}

func TestApplyTemplate_DevPreset(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("DevPreset").SetKey("DVP").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.ApplyTemplate(ctx, userID, proj.ID, "dev"))

	layout, err := svc.GetLayout(ctx, userID, "project", &proj.ID)
	require.NoError(t, err)
	require.NotNil(t, layout)
	assert.Equal(t, "project", layout.DashboardType)
	assert.NotEmpty(t, layout.Items)
}

func TestPriorityRank_Ordering(t *testing.T) {
	assert.Greater(t, dashboard.ExportPriorityRank(issue.PriorityCritical), dashboard.ExportPriorityRank(issue.PriorityHigh))
	assert.Greater(t, dashboard.ExportPriorityRank(issue.PriorityHigh), dashboard.ExportPriorityRank(issue.PriorityMedium))
	assert.Greater(t, dashboard.ExportPriorityRank(issue.PriorityMedium), dashboard.ExportPriorityRank(issue.PriorityLow))
	assert.Equal(t, 0, dashboard.ExportPriorityRank(issue.Priority("unknown")))
}

func TestGetPersonalDashboardData_EmptyUser(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)

	data, err := svc.GetPersonalDashboardData(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, data)
	assert.Empty(t, data.MyTasks)
	assert.Equal(t, 0, data.UnreadCount)
	assert.Empty(t, data.UpcomingCalls)
}

func TestGetPersonalDashboardData_WithUnreadCount(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)
	channelID := uuid.New()

	require.NoError(t, dashboard.ExportDashboardServiceRedis(svc).HSet(ctx, "unread:"+userID.String(), channelID.String(), "3").Err())

	data, err := svc.GetPersonalDashboardData(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 3, data.UnreadCount)
}

func TestGetProjectStats_ComputesFromDB(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)

	status, err := client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.IDEQ(projectID))).
		Only(ctx)
	require.NoError(t, err)

	_, err = client.Issue.Create().
		SetProjectID(projectID).
		SetStatusID(status.ID).
		SetReporterID(userID).
		SetIssueNumber(1).
		SetIssueKey("DASH-1").
		SetTitle("Open bug").
		SetIssueType(issue.IssueTypeBug).
		SetPriority(issue.PriorityHigh).
		Save(ctx)
	require.NoError(t, err)

	stats, err := svc.GetProjectStats(ctx, userID, projectID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.BugCount)
	assert.Equal(t, 1, stats.Unassigned)
	assert.Equal(t, 1, stats.TasksByStatus[status.Name])

	cached, err := rdb.Get(ctx, "dashboard:project:"+projectID.String()).Result()
	require.NoError(t, err)
	assert.Contains(t, cached, `"bug_count":1`)
}

// TestGetProjectStats_NonMemberForbidden — A7-001: GetProjectStats требует
// IsProjectMember до cache/DB. Mutation guard: при снятии проверки stranger
// получит stats (err==nil) через shared Redis cache; при живой — ErrForbidden.
func TestGetProjectStats_NonMemberForbidden(t *testing.T) {
	repo, client, ctx, ownerID, projectID := newDashboardRepoEnv(t)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)

	// Seed cache as if a member already warmed it — IDOR would leak via Redis.
	cached := &dashboard.ProjectStats{BugCount: 99, TasksByStatus: map[string]int{"todo": 7}}
	raw, err := json.Marshal(cached)
	require.NoError(t, err)
	require.NoError(t, rdb.Set(ctx, "dashboard:project:"+projectID.String(), string(raw), dashboard.ExportProjectStatsTTL()).Err())

	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)
	_ = ownerID

	stats, err := svc.GetProjectStats(ctx, stranger.ID, projectID)
	require.Error(t, err)
	assert.ErrorIs(t, err, dashboard.ErrForbidden)
	assert.Nil(t, stats, "non-member must not see project stats (A7-001)")
}

func TestGetAdminStats_BasicCounts(t *testing.T) {
	svc, client, ctx, _ := newDashboardTestEnv(t)

	_, err := client.Project.Create().SetName("Alpha").SetKey("ALP").Save(ctx)
	require.NoError(t, err)

	stats, err := svc.GetAdminStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalUsers, 1)
	assert.GreaterOrEqual(t, stats.TotalProjects, 1)
	assert.Equal(t, "healthy", stats.ServiceHealth["postgresql"])
	assert.Equal(t, "healthy", stats.ServiceHealth["redis"])
}

func TestSaveLayout_UpdatesExisting(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)

	req := dashboard.SaveLayoutRequest{
		DashboardType: "personal",
		Items:         []dashboard.LayoutItem{{I: "w1", WidgetType: "my-tasks", W: 6, H: 4}},
	}
	require.NoError(t, svc.SaveLayout(ctx, userID, req))

	req.Items[0].WidgetType = "deadlines"
	require.NoError(t, svc.SaveLayout(ctx, userID, req))

	layout, err := svc.GetLayout(ctx, userID, "personal", nil)
	require.NoError(t, err)
	require.Len(t, layout.Items, 1)
	assert.Equal(t, "deadlines", layout.Items[0].WidgetType)
}

func TestProvisioner_CreateDefaultDashboard(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("Prov").SetKey("PRV").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	prov := dashboard.NewDashboardProjectProvisioner(svc)
	require.NoError(t, prov.CreateDefaultDashboard(ctx, userID, proj.ID))

	layout, err := svc.GetLayout(ctx, userID, "project", &proj.ID)
	require.NoError(t, err)
	require.NotNil(t, layout)
	assert.NotEmpty(t, layout.Items)
}
