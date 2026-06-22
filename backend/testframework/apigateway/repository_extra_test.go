package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEntRepository_CountAndDeleteAPIKey(t *testing.T) {
	repo, userID, ctx := newAPIKeyRepo(t)

	ak, err := repo.CreateAPIKey(ctx, userID, "del", "hash999", "pre999", nil)
	require.NoError(t, err)

	count, err := repo.CountAPIKeyForUser(ctx, userID, ak.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	require.NoError(t, repo.DeleteAPIKey(ctx, ak.ID))

	count, err = repo.CountAPIKeyForUser(ctx, userID, ak.ID)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestEntRepository_CreateWebhookDeliveryAndList(t *testing.T) {
	dsn := "file:whdel_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := apigateway.NewEntRepository(client)
	ctx := t.Context()

	proj, err := client.Project.Create().SetName("Del Proj").SetKey("DEL").Save(ctx)
	require.NoError(t, err)

	ep, err := repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", "secret", []string{"*"})
	require.NoError(t, err)

	err = repo.CreateWebhookDelivery(ctx, apigateway.CreateWebhookDeliveryInput{
		EventType:    "issue.created",
		Payload:      map[string]interface{}{"id": "1"},
		StatusCode:   200,
		ResponseBody: "ok",
		EndpointID:   ep.ID,
	})
	require.NoError(t, err)

	deliveries, err := repo.ListWebhookDeliveries(ctx, ep.ID, 10)
	require.NoError(t, err)
	require.Len(t, deliveries, 1)
	assert.Equal(t, "issue.created", deliveries[0].EventType)
}
