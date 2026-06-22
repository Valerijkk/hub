package audit_test

// Тесты HTTP-обработчика аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
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

	"github.com/valerii/hub/backend/internal/platform/observability"
)

type adminMockRepo struct {
	mockAuditRepo
	adminID uuid.UUID
}

func (m *adminMockRepo) IsSystemAdmin(_ context.Context, userID uuid.UUID) (bool, error) {
	return userID == m.adminID, nil
}

func setupAuditHandlerTest(t *testing.T, repo audit.Repository, userID uuid.UUID) *fiber.App {
	t.Helper()

	svc := audit.NewAuditService(repo, nil)
	h := audit.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutesSimple(api, authMw)
	return app
}

// TestHandler_requireAdmin_Unauthenticated проверяет HTTP-слой: HTTP-обработчик require админ unauthenticated.
func TestHandler_requireAdmin_Unauthenticated(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	svc := audit.NewAuditService(repo, nil)
	h := audit.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutesSimple(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandler_requireAdmin_Forbidden проверяет отклонение при HTTP-обработчик require админ запрещённый.
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

// TestHandler_ListAuditLogs_AdminOK проверяет HTTP-слой: HTTP-обработчик список audit logs админ ok.
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

	var result audit.AuditLogListResponse
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 10, result.PerPage)
	assert.True(t, repo.lastListFilter.SlowOnly)
}

// TestHandler_ListAuditLogs_QueryFilters проверяет HTTP-слой: HTTP-обработчик список audit logs query filters.
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

// TestHandler_GetLimits_AdminOK проверяет HTTP-слой: HTTP-обработчик получение лимиты админ ok.
func TestHandler_GetLimits_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/limits", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_UpdateLimit_InvalidBody проверяет отклонение при HTTP-обработчик обновление limit некорректный body.
func TestHandler_UpdateLimit_InvalidBody(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/admin/limits/"+audit.LimitFileMaxSizeMB,
		strings.NewReader(`not-json`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_UpdateLimit_AdminOK проверяет HTTP-слой: HTTP-обработчик обновление limit админ ok.
func TestHandler_UpdateLimit_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/admin/limits/"+audit.LimitFileMaxSizeMB,
		strings.NewReader(`{"value":"50"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListTrash_AdminOK проверяет HTTP-слой: HTTP-обработчик список корзина админ ok.
func TestHandler_ListTrash_AdminOK(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/trash?type=issue&page=1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RestoreItem_InvalidID проверяет отклонение при HTTP-обработчик restore item некорректный id.
func TestHandler_RestoreItem_InvalidID(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/trash/issue/not-a-uuid/restore", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_BanUser_InvalidID проверяет отклонение при HTTP-обработчик ban user некорректный id.
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

// TestHandler_BanUser_AdminOK проверяет HTTP-слой: HTTP-обработчик ban user админ ok.
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

// TestHandler_ListOperationMetrics_AdminOK проверяет HTTP-слой: HTTP-обработчик список operation метрики админ ok.
func TestHandler_ListOperationMetrics_AdminOK(t *testing.T) {
	observability.ResetMetrics()
	t.Cleanup(observability.ResetMetrics)

	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/observability/ops", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var result []observability.OperationMetric
	require.NoError(t, json.Unmarshal(body, &result))
}

// Проверка на этапе компиляции: adminMockRepo реализует audit.Repository.
var _ audit.Repository = (*adminMockRepo)(nil)