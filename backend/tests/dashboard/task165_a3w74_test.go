//go:build hubtest

package dashboard_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/dashboard"
)

// aggregateFailRepo — IsProjectMember OK, Aggregate → err (A3W7-4-001).
type aggregateFailRepo struct {
	*dashboard.EntRepository
}

func (r *aggregateFailRepo) AggregateProjectIssueStats(ctx context.Context, projectID uuid.UUID) (*dashboard.ProjectIssueAgg, error) {
	return nil, errors.New("injected aggregate failure")
}

func (r *aggregateFailRepo) IsProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return true, nil
}

// TestGetProjectStats_AggregateFailClosed_A3W74_001 — critical Aggregate err → error, не 200 empty.
// Mutation: slog + return stats, nil при Aggregate err → require.Error FAIL.
func TestGetProjectStats_AggregateFailClosed_A3W74_001(t *testing.T) {
	repo, _, ctx, userID, projectID := newDashboardRepoEnv(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(&aggregateFailRepo{EntRepository: repo}, rdb, nil)
	stats, err := svc.GetProjectStats(ctx, userID, projectID)
	require.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "aggregate project issue stats")
}

// TestGetAdminStats_CountsFailClosed_A3W74_001 — CountUsers/counts err → error, не silent zeros.
// Mutation: if err == nil { stats.TotalUsers = … } → require.Error FAIL.
func TestGetAdminStats_CountsFailClosed_A3W74_001(t *testing.T) {
	svc, client, ctx, _ := newDashboardTestEnv(t)
	require.NoError(t, client.Close())

	stats, err := svc.GetAdminStats(ctx)
	require.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "count users")
}

// TestGetLayout_StrangerProjectForbidden_A3W74_003 — project_id без membership → ErrForbidden.
// Mutation: убрать IsProjectMember из GetLayout → stranger OK (err==nil).
func TestGetLayout_StrangerProjectForbidden_A3W74_003(t *testing.T) {
	svc, client, ctx, _ := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("LayoutGet").SetKey("LGT").Save(ctx)
	require.NoError(t, err)

	stranger, err := client.User.Create().
		SetEmail("getlayout-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	layout, err := svc.GetLayout(ctx, stranger.ID, "project", &proj.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, dashboard.ErrForbidden))
	assert.Nil(t, layout)

	_, err = client.ProjectMember.Create().
		SetUserID(stranger.ID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	layout, err = svc.GetLayout(ctx, stranger.ID, "project", &proj.ID)
	require.NoError(t, err)
	assert.Nil(t, layout) // нет сохранённой раскладки — nil OK
}

// TestAggregateProjectCompletedByDay_Bounded_A3W74_005 — burndown path без .All() issues.
// Mutation: вернуть List*.All() без Count/aggregate → метод отсутствует / hygiene FAIL.
func TestAggregateProjectCompletedByDay_Bounded_A3W74_005(t *testing.T) {
	repo, client, ctx, userID, projectID := newDashboardRepoEnv(t)

	done, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		iss, err := client.Issue.Create().
			SetProjectID(projectID).
			SetStatusID(done.ID).
			SetReporterID(userID).
			SetIssueNumber(200 + i).
			SetIssueKey(fmt.Sprintf("BRN-%d", i)).
			SetTitle("done").
			SetIssueType(issue.IssueTypeTask).
			SetPriority(issue.PriorityMedium).
			Save(ctx)
		require.NoError(t, err)
		require.NoError(t, client.Issue.UpdateOneID(iss.ID).SetUpdatedAt(now.Add(-time.Duration(i)*time.Hour)).Exec(ctx))
	}

	since := now.Add(-48 * time.Hour)
	counts, err := repo.AggregateProjectCompletedByDay(ctx, projectID, since, now.Add(time.Hour))
	require.NoError(t, err)
	total := 0
	for _, c := range counts {
		total += c
	}
	assert.Equal(t, 5, total)
	// Bounded: ключи = календарные дни диапазона, не N issues.
	assert.LessOrEqual(t, len(counts), 4)
}

// TestGetPersonalDashboard_StillFailVisible_AfterTask165 — регрессия TASK-119.
func TestGetPersonalDashboard_StillFailVisible_AfterTask165(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	require.NoError(t, client.Close())
	_, err := svc.GetPersonalDashboardData(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load my tasks")
}

// TestGetAdminStats_CountsFailClosed_BeforeTop_A3W74_001 — counts fail-closed (клиент закрыт).
func TestGetAdminStats_CountsFailClosed_BeforeTop_A3W74_001(t *testing.T) {
	repo, client, ctx, _, _ := newDashboardRepoEnv(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)
	require.NoError(t, client.Close())

	_, err = svc.GetAdminStats(ctx)
	require.Error(t, err)
}
