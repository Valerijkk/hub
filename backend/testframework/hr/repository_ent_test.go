package hr_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/hr"
)

func newHRRepoEnv(t *testing.T) (*hr.EntRepository, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:hrrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	admin, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	return hr.NewEntRepository(client), ctx, member.ID, admin.ID
}

func TestEntRepository_ListLeaveByRequesterAndCancel(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 5)
	end := start.AddDate(0, 0, 2)

	created, err := repo.CreateLeaveRequest(ctx, memberID, leaverequest.LeaveTypeVacation, start, end, "trip")
	require.NoError(t, err)

	list, err := repo.ListLeaveByRequester(ctx, memberID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)

	cancelled, err := repo.CancelLeave(ctx, created.ID, memberID, created.Version)
	require.NoError(t, err)
	assert.Equal(t, leaverequest.StatusCancelled, cancelled.Status)

	_, err = repo.CancelLeave(ctx, created.ID, memberID, cancelled.Version)
	require.Error(t, err)

	adminIDs, err := repo.ListSystemAdminIDs(ctx)
	require.NoError(t, err)
	assert.Contains(t, adminIDs, adminID)
}

// TestEntRepository_ListApprovedLeaveInRange_ViewerScope_W7A2009 — не org-wide.
// Mutation: убрать HasRequesterWith scope → stranger leave в списке.
func TestEntRepository_ListApprovedLeaveInRange_ViewerScope_W7A2009(t *testing.T) {
	dsn := "file:hr_scope_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := hr.NewEntRepository(client)

	viewer, err := client.User.Create().
		SetEmail("v-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	teammate, err := client.User.Create().
		SetEmail("m-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().
		SetEmail("s-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	admin, err := client.User.Create().
		SetEmail("a-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	shared, err := client.Project.Create().SetName("Shared").SetKey("HRS").Save(ctx)
	require.NoError(t, err)
	foreign, err := client.Project.Create().SetName("Foreign").SetKey("HRF").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(viewer.ID).SetProjectID(shared.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(teammate.ID).SetProjectID(shared.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(stranger.ID).SetProjectID(foreign.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)

	start := time.Now().UTC().AddDate(0, 0, 10)
	end := start.AddDate(0, 0, 3)
	from := start.AddDate(0, 0, -1)
	to := end.AddDate(0, 0, 1)

	own, err := repo.CreateLeaveRequest(ctx, viewer.ID, leaverequest.LeaveTypePersonal, start, end, "")
	require.NoError(t, err)
	_, err = repo.ApproveLeave(ctx, own.ID, admin.ID, "ok", own.Version)
	require.NoError(t, err)

	mate, err := repo.CreateLeaveRequest(ctx, teammate.ID, leaverequest.LeaveTypeVacation, start, end, "")
	require.NoError(t, err)
	_, err = repo.ApproveLeave(ctx, mate.ID, admin.ID, "ok", mate.Version)
	require.NoError(t, err)

	str, err := repo.CreateLeaveRequest(ctx, stranger.ID, leaverequest.LeaveTypeSick, start, end, "")
	require.NoError(t, err)
	_, err = repo.ApproveLeave(ctx, str.ID, admin.ID, "ok", str.Version)
	require.NoError(t, err)

	got, err := repo.ListApprovedLeaveInRange(ctx, viewer.ID, []uuid.UUID{shared.ID}, from, to)
	require.NoError(t, err)
	require.Len(t, got, 2)
	ids := map[uuid.UUID]bool{}
	for _, lv := range got {
		ids[lv.ID] = true
		require.NotNil(t, lv.Edges.Requester)
		assert.NotEqual(t, stranger.ID, lv.Edges.Requester.ID, "W7-A2-009: foreign leave must not leak")
	}
	assert.True(t, ids[own.ID])
	assert.True(t, ids[mate.ID])
}

func TestEntRepository_IsSystemAdmin(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)

	ok, err := repo.IsSystemAdmin(ctx, adminID)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.IsSystemAdmin(ctx, memberID)
	require.NoError(t, err)
	assert.False(t, ok)
}
