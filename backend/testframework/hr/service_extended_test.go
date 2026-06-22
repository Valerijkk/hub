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
)

type stubHRNotifier struct {
	submitted int
	decisions int
}

func (s *stubHRNotifier) NotifyLeaveSubmitted(_ context.Context, _ []uuid.UUID, _ hr.LeaveRequestResponse, _ uuid.UUID) error {
	s.submitted++
	return nil
}

func (s *stubHRNotifier) NotifyLeaveDecision(_ context.Context, _ uuid.UUID, _ bool, _ hr.LeaveRequestResponse, _ uuid.UUID) error {
	s.decisions++
	return nil
}

type stubLeaveExporter struct {
	calls int
}

func (s *stubLeaveExporter) ExportApprovedLeave(_ context.Context, _ hr.LeaveRequestResponse, _ uuid.UUID) error {
	s.calls++
	return nil
}

func newHRServiceEnv(t *testing.T) (*hr.Service, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	return hr.NewService(repo, nil, nil), ctx, memberID, adminID
}

func TestService_ListMyLeaveAndCancel(t *testing.T) {
	svc, ctx, memberID, _ := newHRServiceEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 20).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 22).Format("2006-01-02")

	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeavePersonal,
		StartDate: start,
		EndDate:   end,
		Reason:    "personal",
	})
	require.NoError(t, err)

	list, err := svc.ListMyLeave(ctx, memberID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)

	cancelled, err := svc.CancelLeave(ctx, memberID, created.ID, created.Version)
	require.NoError(t, err)
	assert.Equal(t, hr.StatusCancelled, cancelled.Status)
}

func TestService_CreateLeave_NotifiesAdmins(t *testing.T) {
	repo, ctx, memberID, _ := newHRRepoEnv(t)
	notifier := &stubHRNotifier{}
	svc := hr.NewService(repo, notifier, nil)

	start := time.Now().UTC().AddDate(0, 0, 30).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 31).Format("2006-01-02")
	_, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveVacation,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, notifier.submitted)
}

func TestService_ApproveLeave_WithExporter(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	notifier := &stubHRNotifier{}
	exporter := &stubLeaveExporter{}
	svc := hr.NewService(repo, notifier, exporter)

	start := time.Now().UTC().AddDate(0, 0, 40).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 42).Format("2006-01-02")
	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveVacation,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	approved, err := svc.ApproveLeave(ctx, adminID, created.ID, hr.ReviewLeaveRequest{Version: created.Version, ReviewNote: "enjoy"})
	require.NoError(t, err)
	assert.Equal(t, hr.StatusApproved, approved.Status)
	assert.Equal(t, 1, notifier.decisions)
	assert.Equal(t, 1, exporter.calls)
}

func TestService_RejectLeave_NotFound(t *testing.T) {
	svc, ctx, _, adminID := newHRServiceEnv(t)
	_, err := svc.RejectLeave(ctx, adminID, uuid.New(), hr.ReviewLeaveRequest{Version: 1})
	require.ErrorIs(t, err, hr.ErrNotFound)
}

func TestService_ListPending_Forbidden(t *testing.T) {
	svc, ctx, memberID, _ := newHRServiceEnv(t)
	_, err := svc.ListPending(ctx, memberID)
	require.ErrorIs(t, err, hr.ErrForbidden)
}

func TestDisplayName_Fallbacks(t *testing.T) {
	dsn := "file:hrname_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("only-email@test.com").
		SetDisplayName("Display").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Display", displayName(u))

	u2, err := client.User.Create().
		SetEmail("name-only@test.com").
		SetName("Real Name").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Real Name", displayName(u2))
}

func TestParseDateRange_Invalid(t *testing.T) {
	_, _, err := parseDateRange("bad", "2026-01-01")
	require.ErrorIs(t, err, hr.ErrInvalidInput)
}

func TestIsValidLeaveType(t *testing.T) {
	assert.True(t, isValidLeaveType(hr.LeaveVacation))
	assert.False(t, isValidLeaveType(hr.LeaveType("invalid")))
}
