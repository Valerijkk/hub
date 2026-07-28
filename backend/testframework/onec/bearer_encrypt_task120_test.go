package onec_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/integrations/onec"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// TestService_BearerTokenEncryptedAtRest_W7A3007 — BearerToken в system_setting ciphertext.
// Mutation: saveConfig без encryptToken → raw содержит plaintext → FAIL.
func TestService_BearerTokenEncryptedAtRest_W7A3007(t *testing.T) {
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)

	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, cipher)

	const plaintext = "plaintext-1c-bearer-secret-W7A3007"
	enabled := true
	url := "https://1c.example.local/webhook"
	field := onec.EmployeeIDEmail
	token := plaintext

	resp, err := svc.UpdateSettings(context.Background(), onec.UpdateSettingsRequest{
		Enabled:         &enabled,
		WebhookURL:      &url,
		BearerToken:     &token,
		EmployeeIDField: &field,
	})
	require.NoError(t, err)
	assert.True(t, resp.BearerTokenSet)

	raw := store.data[onec.ExportSettingsKey()]
	require.NotEmpty(t, raw)
	assert.NotContains(t, raw, plaintext, "W7-A3-007: plaintext token must not sit in system_setting JSON")

	cfg, err := onec.ExportDecodeConfig(raw)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, cfg.BearerToken)
	assert.NotEmpty(t, cfg.BearerToken)

	// Ciphertext должен расшифровываться тем же ключом.
	dec, ok := cipher.MaybeDecryptString(cfg.BearerToken)
	require.True(t, ok)
	assert.Equal(t, plaintext, dec)

	got, err := svc.GetSettings(context.Background())
	require.NoError(t, err)
	assert.True(t, got.BearerTokenSet)
}

// TestService_LegacyPlaintextTokenStillLoads_W7A3007 — MaybeDecrypt: legacy plaintext читается.
func TestService_LegacyPlaintextTokenStillLoads_W7A3007(t *testing.T) {
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)

	store := &memStore{data: map[string]string{}}
	const legacy = "legacy-plain-token"
	raw, err := onec.ExportEncodeConfig(onec.Config{
		Enabled:         true,
		WebhookURL:      "https://1c.example.local/hook",
		BearerToken:     legacy,
		EmployeeIDField: onec.EmployeeIDEmail,
	})
	require.NoError(t, err)
	store.data[onec.ExportSettingsKey()] = raw
	assert.True(t, strings.Contains(raw, legacy))

	svc := onec.NewService(store, stubUsers{}, cipher)
	got, err := svc.GetSettings(context.Background())
	require.NoError(t, err)
	assert.True(t, got.BearerTokenSet)
}
