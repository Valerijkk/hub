package hr_test
import (
	"github.com/valerii/hub/backend/internal/hr"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_RejectLeave_RequiresVersion_A6_006(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	notifier := &stubHRNotifier{}
	svc := hr.NewService(repo, notifier, nil)

	start := time.Now().UTC().AddDate(0, 0, 45).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 47).Format("2006-01-02")
	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveVacation,
		StartDate: start,
		EndDate:   end,
		Reason:    "declined trip",
	})
	require.NoError(t, err)

	_, err = svc.RejectLeave(ctx, adminID, created.ID, hr.ReviewLeaveRequest{
		ReviewNote: "team capacity",
		// Version omitted / 0 — A6-006
	})
	require.ErrorIs(t, err, hr.ErrInvalidInput)

	rejected, err := svc.RejectLeave(ctx, adminID, created.ID, hr.ReviewLeaveRequest{
		ReviewNote: "team capacity",
		Version:    created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, hr.StatusRejected, rejected.Status)
	assert.Equal(t, 1, notifier.decisions)
}

func TestService_ApproveLeave_RequiresVersion_A6_006(t *testing.T) {
	repo, ctx, memberID, adminID := newHRRepoEnv(t)
	svc := hr.NewService(repo, nil, nil)

	start := time.Now().UTC().AddDate(0, 0, 55).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 57).Format("2006-01-02")
	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeavePersonal,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	_, err = svc.ApproveLeave(ctx, adminID, created.ID, hr.ReviewLeaveRequest{
		ReviewNote: "approved",
	})
	require.ErrorIs(t, err, hr.ErrInvalidInput)

	approved, err := svc.ApproveLeave(ctx, adminID, created.ID, hr.ReviewLeaveRequest{
		ReviewNote: "approved",
		Version:    created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, hr.StatusApproved, approved.Status)
}

func TestService_ApproveLeave_ForbiddenForMember(t *testing.T) {
	svc, ctx, memberID, _ := newHRServiceEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 60).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 61).Format("2006-01-02")

	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveSick,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	_, err = svc.ApproveLeave(ctx, memberID, created.ID, hr.ReviewLeaveRequest{Version: created.Version})
	require.ErrorIs(t, err, hr.ErrForbidden)
}

func TestService_RejectLeave_ForbiddenForMember(t *testing.T) {
	svc, ctx, memberID, _ := newHRServiceEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 65).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 66).Format("2006-01-02")

	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveOther,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	_, err = svc.RejectLeave(ctx, memberID, created.ID, hr.ReviewLeaveRequest{Version: created.Version})
	require.ErrorIs(t, err, hr.ErrForbidden)
}

func TestService_ApproveLeave_NotFound(t *testing.T) {
	svc, ctx, _, adminID := newHRServiceEnv(t)
	_, err := svc.ApproveLeave(ctx, adminID, uuid.New(), hr.ReviewLeaveRequest{Version: 1})
	require.ErrorIs(t, err, hr.ErrNotFound)
}
