package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ListFilters_WithProjectID(t *testing.T) {
	app, _, projectID := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/filters/?project_id="+projectID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateFilter_EmptyName(t *testing.T) {
	app, _, projectID := setupSearchHandlerApp(t)

	body, _ := json.Marshal(map[string]any{
		"name":          "",
		"project_id":    projectID.String(),
		"filter_config": map[string]string{"status": "open"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/filters/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ShareFilter_InvalidID(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/filters/not-a-uuid/share", bytes.NewReader([]byte(`{"shared":true}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Search_InvalidLimit(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=hub&limit=abc", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSearch_SearchAllEntityTypes(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "hub", userID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSearch_InvalidTypeIgnored(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	results, _, err := svc.Search(context.Background(), "x", userID, []string{"issue", "unknown_type"}, 5)
	require.NoError(t, err)
	assert.Empty(t, results)
}
