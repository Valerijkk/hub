package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

func TestHandler_ShareFilter_BadJSON(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/filters/"+uuid.New().String()+"/share", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ShareFilter_Unauthorized(t *testing.T) {
	svc, _, _, _, _ := newSearchServiceTestEnv(t)
	h := search.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodPatch, "/api/filters/"+uuid.New().String()+"/share", bytes.NewReader([]byte(`{"shared":true}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetUserID_UnknownType(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("userID", 42)
		_, err := search.ExportGetUserID(c)
		require.Error(t, err)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDeleteSavedFilter_NotOwner(t *testing.T) {
	svc, repo, ctx, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Private", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{}, CreatedAt: now, UpdatedAt: now,
	}

	err := svc.DeleteSavedFilter(ctx, uuid.New(), filterID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "доступ запрещён")
}

func TestShareFilter_Success(t *testing.T) {
	svc, repo, ctx, ownerID, projectID := newSearchServiceTestEnv(t)
	filterID := uuid.New()
	now := time.Now().UTC()
	repo.filters[filterID] = &ent.SavedFilter{
		ID: filterID, Name: "Team filter", OwnerID: ownerID, ProjectID: projectID,
		FilterConfig: map[string]interface{}{"status": "open"}, CreatedAt: now, UpdatedAt: now,
	}

	updated, err := svc.ShareFilter(ctx, ownerID, filterID, true)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.True(t, updated.Shared)
}

func TestSearch_AllShardTypesWithMembership(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	projectID := uuid.New()
	channelID := uuid.New()

	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?)`,
		projectID.String(), userID.String(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?)`,
		channelID.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{isAdmin: true}, db, nil)
	results, err := svc.Search(context.Background(), "hub", userID,
		[]string{"issue", "test_case", "user", "project", "file", "channel", "message"}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetUserChannelIDs_SkipsInvalidRows(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	validChannel := uuid.New()
	_, err := db.Exec(
		`INSERT INTO channels (id, name, channel_type, deleted_at) VALUES (?, 'valid', 'public', NULL)`,
		validChannel.String(),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO channel_members (channel_id, user_id) VALUES (?, ?), (?, ?)`,
		"not-a-uuid", userID.String(),
		validChannel.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	ids, err := search.ExportGetUserChannelIDs(svc, context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
}

func TestGetUserProjectIDs_SkipsInvalidRows(t *testing.T) {
	db := newScopeTestDB(t)
	userID := uuid.New()
	validProject := uuid.New()
	_, err := db.Exec(
		`INSERT INTO project_members (project_members, user_project_members) VALUES (?, ?), (?, ?)`,
		"bad-project", userID.String(),
		validProject.String(), userID.String(),
	)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	ids, err := search.ExportGetUserProjectIDs(svc, context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, ids, 1)
	assert.Equal(t, validProject, ids[0])
}

type failListFiltersRepo struct {
	mockSearchRepo
}

func (f *failListFiltersRepo) ListSavedFilters(context.Context, uuid.UUID, uuid.UUID) ([]*ent.SavedFilter, error) {
	return nil, errors.New("list failed")
}

func TestListSavedFilters_RepoError(t *testing.T) {
	svc := search.NewSearchService(&failListFiltersRepo{}, newScopeTestDB(t), nil)
	_, err := svc.ListSavedFilters(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "фильтры")
}

type failCreateFilterRepo struct {
	mockSearchRepo
}

func (f *failCreateFilterRepo) CreateSavedFilter(context.Context, search.CreateSavedFilterInput) (*ent.SavedFilter, error) {
	return nil, errors.New("create failed")
}

func TestCreateSavedFilter_RepoError(t *testing.T) {
	svc := search.NewSearchService(&failCreateFilterRepo{}, newScopeTestDB(t), nil)
	_, err := svc.CreateSavedFilter(context.Background(), uuid.New(), search.SavedFilterCreateRequest{
		Name: "X", ProjectID: uuid.New().String(), FilterConfig: map[string]interface{}{},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "создать фильтр")
}

func TestHandler_CreateFilter_ServiceError(t *testing.T) {
	svc := search.NewSearchService(&failCreateFilterRepo{}, newScopeTestDB(t), nil)
	h := search.NewHandler(svc)
	userID := uuid.New()
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body := `{"name":"Broken","filter_config":{},"project_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/filters/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
