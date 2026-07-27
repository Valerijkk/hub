package hr_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/hr"
)

// TestHandler_ApproveLeave_VersionZero_A6_006 — HTTP path must reject version<1 with 400.
// Mutation: drop service `req.Version < 1` guard → 200/409 instead of 400.
func TestHandler_ApproveLeave_VersionZero_A6_006(t *testing.T) {
	app, memberID, adminID := setupHRTest(t)
	created := createPendingLeaveHTTP(t, app, memberID, 40)

	body := `{"version":0,"review_note":"nope"}`
	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/approve", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_CancelLeave_BindAndVersion_A6_006 — cancel must 400 on bad JSON and version<1.
// Mutation: drop version<1 → `{"version":0}` cancels successfully (red).
func TestHandler_CancelLeave_BindAndVersion_A6_006(t *testing.T) {
	app, memberID, _ := setupHRTest(t)
	created := createPendingLeaveHTTP(t, app, memberID, 42)

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/cancel", bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", memberID.String())
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("version zero", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/cancel", bytes.NewReader([]byte(`{"version":0}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", memberID.String())
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func createPendingLeaveHTTP(t *testing.T, app *fiber.App, memberID uuid.UUID, dayOffset int) hr.LeaveRequestResponse {
	t.Helper()
	start := time.Now().UTC().AddDate(0, 0, dayOffset).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, dayOffset+1).Format("2006-01-02")
	body := `{"leave_type":"personal","start_date":"` + start + `","end_date":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	return created
}
