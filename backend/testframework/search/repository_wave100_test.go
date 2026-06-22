package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEntRepository_IsSystemAdmin_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:adm_nf_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	repo := search.NewEntRepository(client)
	ok, err := repo.IsSystemAdmin(context.Background(), uuid.New())
	require.Error(t, err)
	assert.False(t, ok)
}

func TestEntRepository_IsSystemAdmin_Founder(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:adm_fnd_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	founder, err := client.User.Create().
		SetEmail("founder-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole("founder").
		Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	ok, err := repo.IsSystemAdmin(ctx, founder.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestEntRepository_ListSavedFilters_IncludesShared(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:sf_shared_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("owner@test.com").Save(ctx)
	require.NoError(t, err)
	other, err := client.User.Create().SetEmail("other@test.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("Shared Proj").SetKey("SHR").Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	created, err := repo.CreateSavedFilter(ctx, search.CreateSavedFilterInput{
		Name: "Team bugs", FilterConfig: map[string]interface{}{"type": "bug"},
		OwnerID: owner.ID, ProjectID: proj.ID,
	})
	require.NoError(t, err)
	_, err = repo.UpdateSavedFilterShared(ctx, created.ID, true)
	require.NoError(t, err)

	list, err := repo.ListSavedFilters(ctx, other.ID, proj.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.True(t, list[0].Shared)
}

type brokenShareRepo struct {
	*statefulSearchRepo
}

func (b *brokenShareRepo) GetSavedFilter(context.Context, uuid.UUID) (*ent.SavedFilter, error) {
	return nil, errors.New("db unavailable")
}

func TestHandler_ShareFilter_ServiceError(t *testing.T) {
	repo := &brokenShareRepo{statefulSearchRepo: newStatefulSearchRepo()}
	svc := search.NewSearchService(repo, newScopeTestDB(t), nil)
	h := search.NewHandler(svc)
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(fiber.MethodPatch, "/api/filters/"+uuid.New().String()+"/share", bytes.NewReader([]byte(`{"shared":true}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestSearch_SearchAllTypesWithMembership(t *testing.T) {
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
	results, err := svc.Search(context.Background(), "hub", userID, nil, 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}
