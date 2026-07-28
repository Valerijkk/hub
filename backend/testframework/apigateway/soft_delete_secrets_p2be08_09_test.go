package apigateway_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/apigateway"
)

func TestSoftDeleteAPIKey_RevokeKeepsRow(t *testing.T) {
	// W4-A5-007 / P2-BE-08: DeleteAPIKey → deleted_at; hash lookup fails.
	dsn := "file:apiksoft_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("softkey-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	ak, err := repo.CreateAPIKey(ctx, u.ID, "ci", "hash-soft-delete", "prefix12", nil)
	require.NoError(t, err)

	require.NoError(t, repo.DeleteAPIKey(ctx, ak.ID))

	_, err = repo.GetAPIKeyByHashWithUser(ctx, "hash-soft-delete")
	require.Error(t, err)

	list, err := repo.ListAPIKeysByUser(ctx, u.ID)
	require.NoError(t, err)
	assert.Empty(t, list)

	raw, err := client.APIKey.Get(ctx, ak.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)
}

func TestWebhookSecret_EncryptedAtRest(t *testing.T) {
	// W4-A5-008 / P2-BE-09: secret at-rest encrypted; Sensitive() hides String().
	dsn := "file:apigwsec_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("WH").SetKey("WH1").Save(ctx)
	require.NoError(t, err)

	cipher := apigatewayTestCipher(t)
	repo := apigateway.NewEntRepository(client, cipher)
	plain := "whsec_p2be09_plaintext"

	ep, err := repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", plain, []string{"issue.created"})
	require.NoError(t, err)

	raw, err := client.WebhookEndpoint.Get(ctx, ep.ID)
	require.NoError(t, err)
	assert.NotEqual(t, plain, raw.Secret)
	assert.NotContains(t, raw.String(), plain)

	active, err := repo.ListActiveWebhookEndpoints(ctx)
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, plain, active[0].Secret)

	require.NoError(t, repo.DeleteWebhookEndpoint(ctx, ep.ID))
	active, err = repo.ListActiveWebhookEndpoints(ctx)
	require.NoError(t, err)
	assert.Empty(t, active)

	raw2, err := client.WebhookEndpoint.Get(ctx, ep.ID)
	require.NoError(t, err)
	require.NotNil(t, raw2.DeletedAt)
}

func TestWebhookSecret_NilCipherRejectsCreate(t *testing.T) {
	dsn := "file:apigwnil_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("WH").SetKey("WH2").Save(ctx)
	require.NoError(t, err)

	repo := apigateway.NewEntRepository(client, nil)
	_, err = repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", "secret", []string{"issue.created"})
	require.ErrorIs(t, err, apigateway.ErrEncryptionRequired)
}
