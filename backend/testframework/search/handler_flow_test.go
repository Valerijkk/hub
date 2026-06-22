package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSearchHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	svc, _, _, userID, projectID := newSearchServiceTestEnv(t)
	h := search.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, userID, projectID
}

func TestHandler_Search_MissingQuery(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Search_Success(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=hub&limit=5", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body search.SearchResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotNil(t, body.Results)
}

func TestHandler_Search_Unauthorized(t *testing.T) {
	svc, _, _, _, _ := newSearchServiceTestEnv(t)
	h := search.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_FilterCRUDFlow(t *testing.T) {
	app, _, projectID := setupSearchHandlerApp(t)

	createBody := `{"name":"Open bugs","filter_config":{"status":"open"},"project_id":"` + projectID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/filters/", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created search.SavedFilterDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Open bugs", created.Name)

	req = httptest.NewRequest(http.MethodGet, "/api/filters/?project_id="+projectID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var list search.SavedFilterListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	require.Len(t, list.Filters, 1)

	shareBody := `{"shared":true}`
	req = httptest.NewRequest(http.MethodPatch, "/api/filters/"+created.ID+"/share", bytes.NewReader([]byte(shareBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/filters/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestHandler_CreateFilter_Validation(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/filters/", bytes.NewReader([]byte(`{"name":"","project_id":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListFilters_MissingProjectID(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/filters/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
