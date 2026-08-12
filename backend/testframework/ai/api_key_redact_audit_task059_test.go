//go:build hubtest

package ai_test

// TASK-059 / P3-BE-03 / W5-A1-011, W5-A3-019:
// — Create BYOK: raw JSON без api_key/sk-*, has_api_key=true
// — LogAIChat details: только metadata, без prompt/response
// Mutation: сериализовать APIKey в ProviderResponse / добавить prompt в details → FAIL.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/auditlog"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/ai"
	auditmod "github.com/valerii/hub/backend/internal/audit"
)

func TestUserHandler_CreateProvider_RedactsAPIKey_W5A1011(t *testing.T) {
	app, _ := testUserAIApp(t)
	secret := "sk-never-in-response-task059"

	body, _ := json.Marshal(ai.CreateProviderRequest{
		Name:         "BYOK Secret",
		ProviderType: string(aiprovider.ProviderTypeOpenaiCompatible),
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       secret,
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
	assert.NotContains(t, string(raw), secret)
	assert.NotContains(t, string(raw), `"api_key"`)

	var created ai.ProviderResponse
	require.NoError(t, json.Unmarshal(raw, &created))
	assert.True(t, created.HasAPIKey)
}

func TestAuditRepoLogger_LogAIChat_NoPromptInDetails_W5A3019(t *testing.T) {
	dsn := "file:ai_audit_priv_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("ai-priv-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	logger := ai.NewAuditRepoLogger(auditmod.NewRepository(client))
	providerID := uuid.New()
	require.NoError(t, logger.LogAIChat(ctx, user.ID, providerID, "gpt-4", 10, 20, 50*time.Millisecond, "127.0.0.1", "agent"))

	row, err := client.AuditLog.Query().Where(auditlog.ActionEQ("ai.chat")).Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, row.Details)

	for _, forbidden := range []string{"prompt", "response", "content", "messages"} {
		_, ok := row.Details[forbidden]
		assert.False(t, ok, "audit details must not contain %q", forbidden)
	}
	assert.Equal(t, providerID.String(), row.Details["provider_id"])
	assert.Equal(t, "gpt-4", row.Details["model"])
	assert.Contains(t, row.Details, "tokens_in")
	assert.Contains(t, row.Details, "tokens_out")
	assert.Contains(t, row.Details, "latency_ms")
}
