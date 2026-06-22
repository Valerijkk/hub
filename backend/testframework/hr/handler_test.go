package hr_test
import (
	"github.com/valerii/hub/backend/internal/hr"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func setupHRTest(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:hr?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("dev@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	admin, err := client.User.Create().
		SetEmail("hr-admin@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := hr.NewEntRepository(client)
	svc := hr.NewService(repo, nil, nil)
	h := hr.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")

	authMw := func(c fiber.Ctx) error {
		uid := c.Get("X-Test-User-ID")
		if uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, member.ID, admin.ID
}

func TestHandler_LeaveApproveFlow(t *testing.T) {
	app, memberID, adminID := setupHRTest(t)

	start := time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 10).Format("2006-01-02")
	body := `{"leave_type":"vacation","start_date":"` + start + `","end_date":"` + end + `","reason":"Family trip"}`

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, hr.StatusPending, created.Status)

	req = httptest.NewRequest(http.MethodGet, "/api/hr/leave/pending", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	approveBody, err := json.Marshal(hr.ReviewLeaveRequest{Version: created.Version})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/approve", bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var approved hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&approved))
	assert.Equal(t, hr.StatusApproved, approved.Status)
}

func TestHandler_LeaveRejectFlow(t *testing.T) {
	app, memberID, adminID := setupHRTest(t)

	start := time.Now().UTC().AddDate(0, 0, 14).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 16).Format("2006-01-02")
	body := `{"leave_type":"sick","start_date":"` + start + `","end_date":"` + end + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	rejectBody, err := json.Marshal(hr.ReviewLeaveRequest{Version: created.Version})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/reject", bytes.NewReader(rejectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var rejected hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rejected))
	assert.Equal(t, hr.StatusRejected, rejected.Status)
}

func TestHandler_CreateLeave_OverlapRejected(t *testing.T) {
	app, memberID, _ := setupHRTest(t)
	start := time.Now().UTC().AddDate(0, 0, 3).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02")
	body := `{"start_date":"` + start + `","end_date":"` + end + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_ListMyLeaveAndCancel(t *testing.T) {
	app, memberID, _ := setupHRTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/hr/leave", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	start := time.Now().UTC().AddDate(0, 0, 50).Format("2006-01-02")
	end := time.Now().UTC().AddDate(0, 0, 52).Format("2006-01-02")
	body := `{"leave_type":"other","start_date":"` + start + `","end_date":"` + end + `"}`
	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created hr.LeaveRequestResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	cancelBody, _ := json.Marshal(map[string]int{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/hr/leave/"+created.ID.String()+"/cancel", bytes.NewReader(cancelBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListPending_Forbidden(t *testing.T) {
	app, memberID, _ := setupHRTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/hr/leave/pending", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
