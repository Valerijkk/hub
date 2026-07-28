//go:build hubtest

package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	_ "github.com/mattn/go-sqlite3"
)

// A2W5-A1-003: Create/Update не отдают encrypt/base_url internals.
func TestMapProviderMutateError_StableMessages_TASK113(t *testing.T) {
	status, msg := ai.ExportMapProviderMutateError(ai.ErrEncryptionRequired)
	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Equal(t, "encryption required", msg)

	status, msg = ai.ExportMapProviderMutateError(errors.Join(errors.New("wrap"), safehttp.ErrBlockedAddress))
	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Equal(t, "invalid base_url", msg)
	assert.NotContains(t, msg, "blocked")

	status, msg = ai.ExportMapProviderMutateError(errors.New("encrypt api key: cipher exploded secret=xyz"))
	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Equal(t, "invalid provider configuration", msg)
	assert.NotContains(t, msg, "cipher")
	assert.NotContains(t, msg, "secret")
}

// A2W5-A1-004: TestProvider JSON.error без upstream URL/auth.
func TestSanitizeProviderTestError_NoUpstreamLeak_TASK113(t *testing.T) {
	raw := errors.New(`POST https://api.evil.example/v1/chat: 401 Authorization: Bearer sk-secret`)
	got := ai.ExportSanitizeProviderTestError(raw)
	assert.Equal(t, "provider test failed", got)
	assert.NotContains(t, got, "evil.example")
	assert.NotContains(t, got, "sk-secret")
	assert.NotContains(t, got, "Bearer")

	got = ai.ExportSanitizeProviderTestError(errors.New("invalid provider URL: blocked private address"))
	assert.Equal(t, "invalid provider URL", got)
	assert.NotContains(t, got, "blocked private")
}

func TestAdminHandler_CreateProvider_NoAPIKeyInJSON_TASK113(t *testing.T) {
	// A2W5-A1-005: Create с sk-test → raw key не в JSON, только has_api_key.
	app, _ := testAdminAIApp(t, true)
	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "DeepSeek",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.deepseek.com/v1",
		APIKey:       "sk-test-leak-me-please",
		Models:       []string{"deepseek-chat"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "sk-test-leak-me-please")
	assert.NotContains(t, string(raw), `"api_key"`)

	var created ai.ProviderResponse
	require.NoError(t, json.Unmarshal(raw, &created))
	assert.True(t, created.HasAPIKey)
}

func TestUserHandler_CreateProvider_NoAPIKeyInJSON_TASK113(t *testing.T) {
	app, _ := testUserAIApp(t)
	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "BYOK",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-user-secret-key",
		Models:       []string{"gpt-4o-mini"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ai/my/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "sk-user-secret-key")
	assert.NotContains(t, string(raw), `"api_key"`)
}

func TestAdminHandler_CreateProvider_InvalidBaseURL_NoInternals_TASK113(t *testing.T) {
	// A2W5-A1-003: EntRepository валидирует BaseURL; handler мапит без internals.
	// (mockRepo намеренно пропускает SSRF — offline Ollama fixtures в других тестах.)
	dsn := "file:ai_a1003_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := ai.NewEntRepository(client, cipher)
	svc := ai.NewService(repo, &mockAdapter{}, nil)

	userID := uuid.New()
	h := ai.NewAdminHandler(svc, func(_ fiber.Ctx, id uuid.UUID) (bool, error) {
		return id == userID, nil
	})
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "SSRFy",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "http://127.0.0.1:9/v1",
		Models:       []string{"m"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(string(raw)), "invalid base_url")
	assert.NotContains(t, string(raw), "blocked")
	assert.NotContains(t, string(raw), "127.0.0.1")
}

func TestService_TestProvider_SanitizedError_TASK113(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{providers: []*ent.AIProvider{{
		ID: id, Name: "FailProv", BaseURL: "https://api.example.com/v1",
		Models: []string{"m"}, Enabled: true, Scope: aiprovider.ScopeSystem,
		APIKeyEnc: "sk-live",
	}}}
	svc := ai.NewService(repo, &mockAdapter{err: errors.New("dial https://api.example.com: Authorization Bearer sk-live")}, nil)
	resp, err := svc.TestProvider(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.OK)
	assert.Equal(t, "provider test failed", resp.Error)
	assert.NotContains(t, resp.Error, "api.example.com")
	assert.NotContains(t, resp.Error, "sk-live")
	assert.NotContains(t, resp.Error, "Bearer")
}
