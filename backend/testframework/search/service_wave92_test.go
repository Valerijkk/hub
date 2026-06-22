package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func TestEntRepository_GetSavedFilter(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:search_filter_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("search-user@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Search Proj").
		SetKey("SRCH").
		Save(ctx)
	require.NoError(t, err)

	repo := search.NewEntRepository(client)
	created, err := repo.CreateSavedFilter(ctx, search.CreateSavedFilterInput{
		OwnerID:      u.ID,
		Name:         "My filter",
		FilterConfig: map[string]any{"status": "open"},
		ProjectID:    proj.ID,
	})
	require.NoError(t, err)

	got, err := repo.GetSavedFilter(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "My filter", got.Name)
	assert.Equal(t, u.ID, got.OwnerID)
}

type adminSearchRepo struct {
	mockSearchRepo
	admin bool
}

func (r *adminSearchRepo) IsSystemAdmin(_ context.Context, _ uuid.UUID) (bool, error) {
	return r.admin, nil
}
