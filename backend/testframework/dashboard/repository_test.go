package dashboard_test
import (
	"github.com/valerii/hub/backend/internal/dashboard"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/dashboardlayout"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func newDashboardRepoEnv(t *testing.T) (*dashboard.EntRepository, *ent.Client, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:dashrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("dashrepo-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Dash Project").
		SetKey("DASH").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return dashboard.NewEntRepository(client), client, ctx, user.ID, proj.ID
}

func TestEntRepository_SaveAndGetLayout(t *testing.T) {
	repo, _, ctx, userID, projectID := newDashboardRepoEnv(t)

	req := dashboard.SaveLayoutRequest{
		DashboardType: "project",
		ProjectID:     &projectID,
		Items: []dashboard.LayoutItem{
			{I: "w1", X: 0, Y: 0, W: 6, H: 4, WidgetType: "my-tasks"},
		},
	}
	require.NoError(t, repo.SaveDashboardLayout(ctx, userID, req))

	layout, err := repo.GetDashboardLayout(ctx, userID, "project", &projectID)
	require.NoError(t, err)
	require.NotNil(t, layout)
	assert.Equal(t, dashboardlayout.DashboardTypeProject, layout.DashboardType)
}

func TestEntRepository_CountUsers(t *testing.T) {
	repo, _, ctx, _, _ := newDashboardRepoEnv(t)
	count, err := repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
}

func TestEntRepository_ListMyOpenTasks_Empty(t *testing.T) {
	repo, _, ctx, userID, _ := newDashboardRepoEnv(t)
	tasks, err := repo.ListMyOpenTasks(ctx, userID, 10)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestIsNotFound_Dashboard(t *testing.T) {
	repo, _, ctx, userID, _ := newDashboardRepoEnv(t)
	_, err := repo.GetDashboardLayout(ctx, userID, "personal", nil)
	assert.True(t, dashboard.IsNotFound(err))
}

// dashboardlayoutType removed — use ent enum directly.

func TestEntRepository_CountDistinctUsersFromSessionsSince(t *testing.T) {
	repo, client, ctx, userID, _ := newDashboardRepoEnv(t)

	before, err := repo.CountDistinctUsersFromSessionsSince(ctx, time.Now().AddDate(-1, 0, 0))
	require.NoError(t, err)

	_, err = client.Session.Create().
		SetUserID(userID).
		SetRefreshToken("hash-" + uuid.New().String()).
		SetDevice("browser").
		SetIPAddress("127.0.0.1").
		SetExpiresAt(time.Now().Add(24 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	after, err := repo.CountDistinctUsersFromSessionsSince(ctx, time.Now().AddDate(-1, 0, 0))
	require.NoError(t, err)
	assert.Equal(t, before+1, after)
}
