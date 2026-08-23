package calendar_test

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
	"github.com/valerii/hub/backend/internal/calendar"
)

// TestListApprovedLeaveInRange_ScopedToSharedProjects_A6_005 — org-wide leave
// must not leak to a JWT with no shared project membership.
func TestListApprovedLeaveInRange_ScopedToSharedProjects_A6_005(t *testing.T) {
	dsn := "file:cal_a6005_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := calendar.NewEntRepository(client)

	viewer, err := client.User.Create().SetEmail("v-" + uuid.New().String()[:8] + "@t.com").SetName("Viewer").Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().SetEmail("s-" + uuid.New().String()[:8] + "@t.com").SetName("Stranger").Save(ctx)
	require.NoError(t, err)
	teammate, err := client.User.Create().SetEmail("m-" + uuid.New().String()[:8] + "@t.com").SetName("Teammate").Save(ctx)
	require.NoError(t, err)

	shared, err := client.Project.Create().SetName("Shared").SetKey("SH").Save(ctx)
	require.NoError(t, err)
	foreign, err := client.Project.Create().SetName("Foreign").SetKey("FG").Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().SetUserID(viewer.ID).SetProjectID(shared.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(teammate.ID).SetProjectID(shared.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(stranger.ID).SetProjectID(foreign.ID).SetProjectRole("developer").Save(ctx)
	require.NoError(t, err)

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	_, err = client.LeaveRequest.Create().
		SetRequesterID(stranger.ID).
		SetLeaveType(leaverequest.LeaveTypeSick).
		SetStartDate(from.AddDate(0, 0, 2)).
		SetEndDate(from.AddDate(0, 0, 5)).
		SetStatus(leaverequest.StatusApproved).
		Save(ctx)
	require.NoError(t, err)
	mateLeave, err := client.LeaveRequest.Create().
		SetRequesterID(teammate.ID).
		SetLeaveType(leaverequest.LeaveTypeVacation).
		SetStartDate(from.AddDate(0, 0, 10)).
		SetEndDate(from.AddDate(0, 0, 14)).
		SetStatus(leaverequest.StatusApproved).
		Save(ctx)
	require.NoError(t, err)
	ownLeave, err := client.LeaveRequest.Create().
		SetRequesterID(viewer.ID).
		SetLeaveType(leaverequest.LeaveTypePersonal).
		SetStartDate(from.AddDate(0, 0, 20)).
		SetEndDate(from.AddDate(0, 0, 21)).
		SetStatus(leaverequest.StatusApproved).
		Save(ctx)
	require.NoError(t, err)

	projectIDs, err := repo.ListUserProjectIDs(ctx, viewer.ID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{shared.ID}, projectIDs)

	got, err := repo.ListApprovedLeaveInRange(ctx, viewer.ID, projectIDs, from, to)
	require.NoError(t, err)
	require.Len(t, got, 2)
	ids := make(map[uuid.UUID]bool, len(got))
	for _, lv := range got {
		ids[lv.ID] = true
		require.NotNil(t, lv.Edges.Requester)
		assert.NotEqual(t, stranger.ID, lv.Edges.Requester.ID, "A6-005: foreign leave must not leak")
	}
	assert.True(t, ids[ownLeave.ID], "own leave visible")
	assert.True(t, ids[mateLeave.ID], "shared-project teammate leave visible")
}
