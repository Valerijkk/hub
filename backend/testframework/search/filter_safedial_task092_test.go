package search_test

// TASK-092 / A2W5-A4-001/002/004: filter HTTP 403/404 + Meili NewSafeHTTPClient.
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/search"
)

// TestHandler_DeleteFilter_NotOwner_Returns403 — mutation: убрать
// errors.Is(ErrFilterForbidden) в deleteFilter → снова 500.
func TestHandler_DeleteFilter_NotOwner_Returns403(t *testing.T) {
	svc, repo, _, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Private", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	stranger := uuid.New()
	h := search.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", stranger.String())
		return c.Next()
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/filters/"+filterID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestHandler_ShareFilter_NotOwner_Returns403 — mutation: убрать
// errors.Is(ErrFilterForbidden) в shareFilter → снова 500.
func TestHandler_ShareFilter_NotOwner_Returns403(t *testing.T) {
	svc, repo, _, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Team", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	stranger := uuid.New()
	h := search.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", stranger.String())
		return c.Next()
	})

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/filters/"+filterID.String()+"/share",
		bytes.NewReader([]byte(`{"shared":true}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_DeleteFilter_NotFound_Returns404(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/filters/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDeleteSavedFilter_NotOwner_ErrFilterForbidden(t *testing.T) {
	svc, repo, ctx, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Private", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	err := svc.DeleteSavedFilter(ctx, uuid.New(), filterID)
	require.ErrorIs(t, err, search.ErrFilterForbidden)
}

func TestShareFilter_NotOwner_ErrFilterForbidden(t *testing.T) {
	svc, repo, ctx, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "X", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	_, err := svc.ShareFilter(ctx, uuid.New(), filterID, true)
	require.ErrorIs(t, err, search.ErrFilterForbidden)
}

// TestNewMeiliIndexerFromEnv_UsesSafeHTTPClient_Task092 — mutation: вернуть
// &http.Client{Timeout:30s} без SafeDial → DialContext nil / metadata dial OK.
func TestNewMeiliIndexerFromEnv_UsesSafeHTTPClient_Task092(t *testing.T) {
	t.Setenv("MEILI_HOST", "https://meili.example.com")
	t.Setenv("MEILI_MASTER_KEY", "x")
	idx := search.NewMeiliIndexerFromEnv()
	require.NotNil(t, idx)

	client := search.ExportMeiliIndexerClient(idx)
	require.NotNil(t, client)
	require.NotNil(t, client.Transport, "NewSafeHTTPClient задаёт Transport с SafeDial")

	rt, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "ожидается *http.Transport от NewSafeHTTPClient")
	require.NotNil(t, rt.DialContext)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := rt.DialContext(ctx, "tcp", "169.254.169.254:80")
	require.Error(t, err)
	assert.ErrorIs(t, err, safehttp.ErrBlockedAddress)
}
