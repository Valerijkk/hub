package audit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type adminMockRepo struct {
	mockAuditRepo
	adminID uuid.UUID
}

func (m *adminMockRepo) IsSystemAdmin(_ context.Context, userID uuid.UUID) (bool, error) {
	return userID == m.adminID, nil
}

func setupAuditHandlerTest(t *testing.T, repo Repository, userID uuid.UUID) *fiber.App {
	t.Helper()

	svc := NewAuditService(repo, nil)
	h := NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutesSimple(api, authMw)
	return app
}

func TestHandler_requireAdmin_Unauthenticated(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	svc := NewAuditService(repo, nil)
	h := NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutesSimple(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_requireAdmin_Forbidden(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	nonAdmin := uuid.New()
	app := setupAuditHandlerTest(t, repo, nonAdmin)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_ListAuditLogs_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit?slow=1&page=2&per_page=10", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result AuditLogListResponse
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.True(t, repo.lastListFilter.SlowOnly)
}

func TestHandler_ListAuditLogs_QueryFilters(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	dateFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	dateTo := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	url := "/api/admin/audit?action=login&entity_type=http&entity_id=e1&ip_address=10.0" +
		"&user_id=" + userID.String() +
		"&project_id=" + projectID.String() +
		"&date_from=" + dateFrom +
		"&date_to=" + dateTo

	req := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "login", repo.lastListFilter.Action)
	assert.Equal(t, "http", repo.lastListFilter.EntityType)
	assert.Equal(t, "e1", repo.lastListFilter.EntityID)
	assert.Equal(t, "10.0", repo.lastListFilter.IPAddress)
	require.NotNil(t, repo.lastListFilter.UserID)
	assert.Equal(t, userID, *repo.lastListFilter.UserID)
	require.NotNil(t, repo.lastListFilter.ProjectID)
	assert.Equal(t, projectID, *repo.lastListFilter.ProjectID)
}

func TestHandler_GetLimits_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/limits", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateLimit_InvalidBody(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/admin/limits/"+LimitFileMaxSizeMB,
		strings.NewReader(`not-json`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateLimit_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/admin/limits/"+LimitFileMaxSizeMB,
		strings.NewReader(`{"value":"50"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_ListTrash_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/trash?type=issue&page=1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RestoreItem_InvalidID(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/trash/issue/not-a-uuid/restore", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_BanUser_InvalidID(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/bad-id/ban",
		strings.NewReader(`{"reason":"spam"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_BanUser_AdminOK(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+targetID.String()+"/ban",
		strings.NewReader(`{"reason":"policy violation"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, repo.deactivateIDs, targetID)
}

// Ensure adminMockRepo satisfies Repository at compile time.
var _ Repository = (*adminMockRepo)(nil)