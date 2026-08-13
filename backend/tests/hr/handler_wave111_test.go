package hr_test
import (
	"github.com/valerii/hub/backend/internal/hr"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_RejectLeave_InvalidJSON(t *testing.T) {
	app, memberID, adminID := setupHRTest(t)

	start := time.Now().UTC().AddDate(0, 0, 20).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 21).Format("2006-01-02")
	body := `{"leave_type":"vacation","start_date":"` + start + `","end_date":"` + end + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/reject", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateLeave_InvalidJSON(t *testing.T) {
	app, memberID, _ := setupHRTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_RejectLeave_Forbidden(t *testing.T) {
	app, memberID, _ := setupHRTest(t)

	start := time.Now().UTC().AddDate(0, 0, 30).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 31).Format("2006-01-02")
	body := `{"leave_type":"sick","start_date":"` + start + `","end_date":"` + end + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	rejectBody := []byte(`{"version":` + strconv.Itoa(created.Version) + `}`)
	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/reject", bytes.NewReader(rejectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandler_CancelLeave_InvalidLeaveID(t *testing.T) {
	app, memberID, _ := setupHRTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave/not-a-uuid/cancel", bytes.NewReader([]byte(`{"version":1}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
