package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureSearchIndexes_FailsOnNonPostgres(t *testing.T) {
	db := newScopeTestDB(t)
	err := search.EnsureSearchIndexes(context.Background(), db)
	require.Error(t, err)
}

func TestSearch_AdminUserShard(t *testing.T) {
	db := newScopeTestDB(t)
	svc := search.NewSearchService(&mockSearchRepo{isAdmin: true}, db, nil)

	results, _, err := svc.Search(context.Background(), "admin-user", uuid.New(), []string{"user"}, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestHandler_Wave93_GetUserID_UUIDLocal(t *testing.T) {
	svc, _, _, userID, projectID := newSearchServiceTestEnv(t)
	h := search.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(fiber.MethodGet, "/api/search?q=hub&limit=5", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(fiber.MethodGet, "/api/filters/?project_id="+projectID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_Wave93_DeleteFilterBadID(t *testing.T) {
	app, _, _ := setupSearchHandlerApp(t)

	req := httptest.NewRequest(fiber.MethodDelete, "/api/filters/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
