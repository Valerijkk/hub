package hr_test
import (
	"github.com/valerii/hub/backend/internal/hr"
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

func TestEntRepository_ListApprovedLeaveInRange(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 10)
	end := start.AddDate(0, 0, 3)

	created, err := repo.CreateLeaveRequest(ctx, memberID, leaverequest.LeaveTypeSick, start, end, "")
	require.NoError(t, err)

	_, err = repo.ApproveLeave(ctx, created.ID, adminID, "ok", created.Version)
	require.NoError(t, err)

	from := start.AddDate(0, 0, -1)
	to := end.AddDate(0, 0, 1)
	approved, err := repo.ListApprovedLeaveInRange(ctx, from, to)
	require.NoError(t, err)
	require.Len(t, approved, 1)
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
