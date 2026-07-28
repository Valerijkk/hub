package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newAPIKeyRepo(t *testing.T) (apigateway.APIKeyRepository, uuid.UUID, context.Context) {
	t.Helper()
	dsn := "file:apigwrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("repo-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	return apigateway.NewEntRepository(client), u.ID, ctx
}

func TestEntRepository_CreateAndListAPIKeys(t *testing.T) {
	repo, userID, ctx := newAPIKeyRepo(t)

	ak, err := repo.CreateAPIKey(ctx, userID, "ci", "hash123", "prefix12", nil)
	require.NoError(t, err)
	assert.Equal(t, "ci", ak.Name)

	list, err := repo.ListAPIKeysByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, ak.ID, list[0].ID)
}

func TestEntRepository_WebhookEndpointCRUD(t *testing.T) {
	dsn := "file:whrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := apigateway.NewEntRepository(client)
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Webhook Proj").SetKey("PR").Save(ctx)
	require.NoError(t, err)

	ep, err := repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", "secret", []string{"*"})
	require.NoError(t, err)

	list, err := repo.ListWebhookEndpointsByProject(ctx, proj.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	active, err := repo.ListActiveWebhookEndpoints(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	require.NotNil(t, active[0].Edges.Project, "ListActiveWebhookEndpoints must WithProject()")
	assert.Equal(t, proj.ID, active[0].Edges.Project.ID)

	isActive := false
	updated, err := repo.UpdateWebhookEndpoint(ctx, ep.ID, apigateway.UpdateWebhookEndpointInput{IsActive: &isActive})
	require.NoError(t, err)
	assert.False(t, updated.IsActive)

	require.NoError(t, repo.DeleteWebhookEndpoint(ctx, ep.ID))
}
