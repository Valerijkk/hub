package ai_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

func TestEntRepository_CreateProvider_NilCipherRejectsAPIKey(t *testing.T) {
	// W5-A3-007 / P2-BE-09: cipher==nil → Create fail, plaintext не пишется.
	dsn := "file:ainilcipher_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := ai.NewEntRepository(client, nil)

	_, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "Leak",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.example.com/v1",
		APIKey:       "sk-should-not-persist",
		Models:       []string{"m1"},
		Enabled:      true,
	})
	require.ErrorIs(t, err, ai.ErrEncryptionRequired)

	n, err := client.AIProvider.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "plaintext api key must not be persisted")
}

func TestEntRepository_CreateProvider_EncryptsAPIKey(t *testing.T) {
	dsn := "file:aienc_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := ai.NewEntRepository(client, cipher)

	plain := "sk-wave039-secret"
	created, err := repo.CreateProvider(ctx, ai.CreateProviderInput{
		Name:         "Encrypted",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.example.com/v1",
		APIKey:       plain,
		Models:       []string{"m1"},
		Enabled:      true,
	})
	require.NoError(t, err)

	raw, err := client.AIProvider.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.NotEqual(t, plain, raw.APIKeyEnc)
	assert.NotContains(t, raw.String(), plain)

	got, err := repo.GetProvider(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, plain, got.APIKeyEnc)
}

func TestEntRepository_GetProvider_DecryptFailure(t *testing.T) {
	// W5-A3-006: битый ciphertext → ошибка, не «ключ» в Authorization.
	dsn := "file:aidecrypt_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := ai.NewEntRepository(client, cipher)

	p, err := client.AIProvider.Create().
		SetName("Broken").
		SetProviderType(aiprovider.ProviderTypeOpenaiCompatible).
		SetBaseURL("https://api.example.com/v1").
		SetAPIKeyEnc("not-a-valid-ciphertext!!!!").
		SetModels([]string{"m1"}).
		SetEnabled(true).
		SetScope(aiprovider.ScopeSystem).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.GetProvider(ctx, p.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt api key")
}