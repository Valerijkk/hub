package calendar_test

// BE-V4-001/002 / TASK-189: soft-deleted leave/sprint не попадают в календарь.
// Mutation: снять DeletedAtIsNil в ListApprovedLeaveInRange / ListSprintsInRange → FAIL.
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

func TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189(t *testing.T) {
	dsn := "file:cal_leave_trash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := calendar.NewEntRepository(client)

	viewer, err := client.User.Create().
		SetEmail("lv-" + uuid.New().String()[:8] + "@t.com").
		SetName("Viewer").
		Save(ctx)
	require.NoError(t, err)

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	active, err := client.LeaveRequest.Create().
		SetRequesterID(viewer.ID).
		SetLeaveType(leaverequest.LeaveTypeVacation).
		SetStartDate(from.AddDate(0, 0, 2)).
		SetEndDate(from.AddDate(0, 0, 5)).
		SetStatus(leaverequest.StatusApproved).
		Save(ctx)
	require.NoError(t, err)

	trashed, err := client.LeaveRequest.Create().
		SetRequesterID(viewer.ID).
		SetLeaveType(leaverequest.LeaveTypeSick).
		SetStartDate(from.AddDate(0, 0, 10)).
		SetEndDate(from.AddDate(0, 0, 12)).
		SetStatus(leaverequest.StatusApproved).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	got, err := repo.ListApprovedLeaveInRange(ctx, viewer.ID, nil, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)
	for _, lv := range got {
		assert.NotEqual(t, trashed.ID, lv.ID, "BE-V4-001: soft-deleted leave must not appear")
	}
}

func TestListSprintsInRange_ExcludesSoftDeleted_Task189(t *testing.T) {
	dsn := "file:cal_sprint_trash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := calendar.NewEntRepository(client)

	proj, err := client.Project.Create().
		SetName("Cal Sprint Trash").
		SetKey("CST").
		Save(ctx)
	require.NoError(t, err)

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	active, err := client.Sprint.Create().
		SetName("Active").
		SetGoal("keep").
		SetStartDate(from.AddDate(0, 0, 1)).
		SetEndDate(from.AddDate(0, 0, 14)).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	trashed, err := client.Sprint.Create().
		SetName("Trashed").
		SetGoal("gone").
		SetStartDate(from.AddDate(0, 0, 5)).
		SetEndDate(from.AddDate(0, 0, 20)).
		SetProjectID(proj.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	got, err := repo.ListSprintsInRange(ctx, []uuid.UUID{proj.ID}, from, to)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, active.ID, got[0].ID)
	for _, sp := range got {
		assert.NotEqual(t, trashed.ID, sp.ID, "BE-V4-002: soft-deleted sprint must not appear")
	}
}
