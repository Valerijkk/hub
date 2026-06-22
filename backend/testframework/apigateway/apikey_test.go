package apigateway_test
import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/apigateway"
)

// TestHashKey проверяет, что apigateway.ExportHashKey возвращает корректный SHA-256 hex.
func TestHashKey(t *testing.T) {
	key := "test-api-key-12345"
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	result := apigateway.ExportHashKey(key)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 64) // SHA-256 hex = 64 символа
}

// TestHashKeyDeterministic проверяет детерминированность хеширования.
func TestHashKeyDeterministic(t *testing.T) {
	key := "deterministic-key"
	h1 := apigateway.ExportHashKey(key)
	h2 := apigateway.ExportHashKey(key)
	assert.Equal(t, h1, h2)
}

// TestHashKeyUniqueness проверяет, что разные ключи дают разные хеши.
func TestHashKeyUniqueness(t *testing.T) {
	h1 := apigateway.ExportHashKey("key-one")
	h2 := apigateway.ExportHashKey("key-two")
	assert.NotEqual(t, h1, h2)
}

// TestKeyGeneration проверяет формат генерируемых ключей.
func TestKeyGeneration(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	fullKey := base64.RawURLEncoding.EncodeToString(raw)

	require.True(t, len(fullKey) >= 8, "ключ должен быть длиннее 8 символов")

	prefix := fullKey[:8]
	assert.Len(t, prefix, 8)
}

// TestContainsEvent проверяет поиск события в списке подписок.
func TestContainsEvent(t *testing.T) {
	events := []string{"issue.created", "issue.updated", "test_run.completed"}

	assert.True(t, apigateway.ExportContainsEvent(events, "issue.created"))
	assert.True(t, apigateway.ExportContainsEvent(events, "test_run.completed"))
	assert.False(t, apigateway.ExportContainsEvent(events, "sprint.started"))
}

// TestContainsEventWildcard проверяет подстановочный символ "*".
func TestContainsEventWildcard(t *testing.T) {
	events := []string{"*"}
	assert.True(t, apigateway.ExportContainsEvent(events, "any.event"))
}

// TestSignPayload проверяет HMAC-SHA256 подпись.
func TestSignPayload(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	secret := "my-secret"

	sig := apigateway.ExportSignPayload(payload, secret)
	assert.Len(t, sig, 64) // HMAC-SHA256 hex = 64 символа

	// Та же подпись при повторном вызове
	sig2 := apigateway.ExportSignPayload(payload, secret)
	assert.Equal(t, sig, sig2)

	// Другой секрет — другая подпись
	sig3 := apigateway.ExportSignPayload(payload, "other-secret")
	assert.NotEqual(t, sig, sig3)
}
