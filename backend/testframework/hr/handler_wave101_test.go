package hr_test
import (
	"github.com/valerii/hub/backend/internal/hr"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_MapServiceError(t *testing.T) {
	assert.Equal(t, fiber.StatusNotFound, mapServiceError(hr.ErrNotFound).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusForbidden, mapServiceError(hr.ErrForbidden).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusBadRequest, mapServiceError(hr.ErrInvalidInput).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusConflict, mapServiceError(hr.ErrVersionConflict).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusInternalServerError, mapServiceError(assert.AnError).(*fiber.Error).Code)
}

func TestService_CancelLeave_VersionZero(t *testing.T) {
	svc, ctx, memberID, _ := newHRServiceEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 10).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 12).Format("2006-01-02")

	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeavePersonal,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	cancelled, err := svc.CancelLeave(ctx, memberID, created.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, hr.StatusCancelled, cancelled.Status)
}

func TestService_CancelLeave_ForbiddenForOtherUser(t *testing.T) {
	svc, ctx, memberID, adminID := newHRServiceEnv(t)
	start := time.Now().UTC().AddDate(0, 0, 15).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 16).Format("2006-01-02")

	created, err := svc.CreateLeave(ctx, memberID, hr.CreateLeaveRequest{
		LeaveType: hr.LeaveSick,
		StartDate: start,
		EndDate:   end,
	})
	require.NoError(t, err)

	_, err = svc.CancelLeave(ctx, adminID, created.ID, 0)
	require.ErrorIs(t, err, hr.ErrForbidden)
}

func TestHandler_ApproveLeave_NotFound(t *testing.T) {
	app, _, adminID := setupHRTest(t)

	body := `{"version":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+uuid.New().String()+"/approve", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
