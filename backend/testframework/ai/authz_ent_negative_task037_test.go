//go:build hubtest

package ai_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// TASK-037 / P2-BE-02 / W5-A1-006 — BYOK foreign-owner denial на Ent + HTTP.
// Mutation: GetUserProvider без HasOwnerWith → stranger PATCH 200.

func TestGetUserProvider_RejectsForeignOwner_Ent_TASK037(t *testing.T) {
	repo, client, ctx, ownerID := newAIRepoEnv(t)

	stranger, err := client.User.Create().
		SetEmail("ai-str-" + uuid.New().String()[:8] + "@test.com").
		SetName("Stranger").
		Save(ctx)
	require.NoError(t, err)

	owned, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "Owner BYOK",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-owner-secret",
		Models:       []string{"gpt-4o-mini"},
		Enabled:      true,
		Scope:        "user",
		OwnerUserID:  &ownerID,
	})
	require.NoError(t, err)

	_, err = repo.GetUserProvider(ctx, stranger.ID, owned.ID)
	require.True(t, ent.IsNotFound(err), "foreign owner must 404, got %v", err)

	got, err := repo.GetUserProvider(ctx, ownerID, owned.ID)
	require.NoError(t, err)
	assert.Equal(t, owned.ID, got.ID)
	assert.Empty(t, got.APIKeyEnc, "API key must never leave GetUserProvider")
}

func TestUserHandler_UpdateProvider_ForeignOwner_TASK037(t *testing.T) {
	dsn := "file:ai_byok_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := t.Context()
	owner, err := client.User.Create().
		SetEmail("byok-owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)
	attacker, err := client.User.Create().
		SetEmail("byok-atk-" + uuid.New().String()[:8] + "@test.com").
		SetName("Attacker").
		Save(ctx)
	require.NoError(t, err)

	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := ai.NewEntRepository(client, cipher)
	svc := ai.NewService(repo, &mockAdapter{}, nil)

	created, err := svc.CreateUserProvider(ctx, owner.ID, ai.CreateProviderRequest{
		Name:         "Secret",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-never-leak",
		Models:       []string{"gpt-4o-mini"},
	})
	require.NoError(t, err)

	h := ai.NewUserHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", attacker.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	enabled := false
	patch, _ := json.Marshal(ai.UpdateProviderRequest{Enabled: &enabled})
	req := httptest.NewRequest(http.MethodPatch, "/api/ai/my/providers/"+created.ID, bytes.NewReader(patch))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	raw, _ := json.Marshal(body)
	assert.NotContains(t, string(raw), "sk-never-leak")
	assert.NotContains(t, string(raw), "api_key")

	// Owner provider untouched.
	pid := uuid.MustParse(created.ID)
	still, err := repo.GetUserProvider(ctx, owner.ID, pid)
	require.NoError(t, err)
	assert.True(t, still.Enabled)
}
