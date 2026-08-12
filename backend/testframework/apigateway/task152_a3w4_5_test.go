//go:build hubtest

package apigateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/apigateway"
)

// A3W4-5-001: soft-deleted project → ListActiveWebhookEndpoints пуст (0 outbound jobs).
// Mutation: убрать HasProjectWith(project.DeletedAtIsNil()) → тест краснеет.
func TestListActiveWebhookEndpoints_SoftDeletedProject_Excluded(t *testing.T) {
	dsn := "file:whsoftdel_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	ctx := context.Background()

	alive, err := client.Project.Create().SetName("Alive").SetKey("ALV").Save(ctx)
	require.NoError(t, err)
	dead, err := client.Project.Create().SetName("Dead").SetKey("DED").Save(ctx)
	require.NoError(t, err)

	_, err = repo.CreateWebhookEndpoint(ctx, alive.ID, "https://example.com/alive", "s1", []string{"issue.created"})
	require.NoError(t, err)
	_, err = repo.CreateWebhookEndpoint(ctx, dead.ID, "https://example.com/dead", "s2", []string{"issue.created"})
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = client.Project.UpdateOneID(dead.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	active, err := repo.ListActiveWebhookEndpoints(ctx)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.NotNil(t, active[0].Edges.Project)
	assert.Equal(t, alive.ID, active[0].Edges.Project.ID)
}

// A3W4-5-002: stale version → 409; свежий version → 200 + version++.
// Mutation: Update без VersionEQ → stale проходит (last-write-wins).
func TestUpdateWebhookEndpoint_VersionEQ_ConflictAndSuccess(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/v",
		Events: []string{"issue.created"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, created.Version)

	newURL := "https://hooks.example.com/v2"
	updated, err := env.webhookSvc.UpdateEndpoint(env.ctx, uuid.MustParse(created.ID), apigateway.UpdateWebhookRequest{
		URL:     &newURL,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.Equal(t, newURL, updated.URL)

	staleURL := "https://hooks.example.com/stale"
	_, err = env.webhookSvc.UpdateEndpoint(env.ctx, uuid.MustParse(created.ID), apigateway.UpdateWebhookRequest{
		URL:     &staleURL,
		Version: 1, // stale
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrVersionConflict)
	assert.Equal(t, fiber.StatusConflict, apigateway.MapToFiber(err).(*fiber.Error).Code)

	app := env.buildApp(t, env.ownerID)
	body, _ := json.Marshal(apigateway.UpdateWebhookRequest{URL: &staleURL, Version: 1})
	req := httptest.NewRequest(http.MethodPut, "/api/webhooks/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	freshURL := "https://hooks.example.com/v3"
	body, _ = json.Marshal(apigateway.UpdateWebhookRequest{URL: &freshURL, Version: 2})
	req = httptest.NewRequest(http.MethodPut, "/api/webhooks/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got apigateway.WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, 3, got.Version)
	assert.Equal(t, freshURL, got.URL)
}

// A3W4-5-003 (trim): sprint.completed не в allowlist — Create/Update отклоняют.
// Mutation: вернуть "sprint.completed" в allowedWebhookEvents → тест краснеет.
func TestValidateWebhookEvents_SprintCompleted_Rejected(t *testing.T) {
	env := newWebhookEnv(t)

	_, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/sprint",
		Events: []string{"sprint.completed"},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrInvalidInput)
	assert.Contains(t, err.Error(), "sprint.completed")

	// Matching algorithm остаётся (legacy endpoints), но allowlist закрыт.
	assert.True(t, apigateway.ExportContainsEvent(
		[]string{"sprint.completed"}, "sprint."+uuid.New().String()+".completed"))
}

// A3W4-5-004: name >255 рун → 400, не 500.
// Mutation: убрать rune MaxLen guard → длинное имя → 500 (Ent).
func TestCreateAPIKey_NameMaxLen_400(t *testing.T) {
	h, _, userID := newHandlerTestEnv(t)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	})
	app.Post("/user/api-keys", h.CreateAPIKey)

	longName := strings.Repeat("я", 256)
	require.Equal(t, 256, utf8.RuneCountInString(longName))

	body, _ := json.Marshal(apigateway.CreateAPIKeyRequest{Name: longName})
	req := httptest.NewRequest(http.MethodPost, "/user/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Ent MaxLen(255) считает байты; ASCII на границе 255 — валидный happy-path.
	okName := strings.Repeat("a", 255)
	body, _ = json.Marshal(apigateway.CreateAPIKeyRequest{Name: okName})
	req = httptest.NewRequest(http.MethodPost, "/user/api-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// Repo UpdateWebhookEndpoint: VersionEQ на слое Ent (не только service).
func TestEntRepository_UpdateWebhookEndpoint_VersionEQ(t *testing.T) {
	dsn := "file:whver_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	ctx := context.Background()

	_, err := client.User.Create().
		SetEmail("ver-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("Ver Proj").SetKey("VR").Save(ctx)
	require.NoError(t, err)

	ep, err := repo.CreateWebhookEndpoint(ctx, proj.ID, "https://example.com/hook", "secret", []string{"*"})
	require.NoError(t, err)
	assert.Equal(t, 1, ep.Version)

	isActive := false
	updated, err := repo.UpdateWebhookEndpoint(ctx, ep.ID, apigateway.UpdateWebhookEndpointInput{
		IsActive: &isActive,
		Version:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.False(t, updated.IsActive)

	_, err = repo.UpdateWebhookEndpoint(ctx, ep.ID, apigateway.UpdateWebhookEndpointInput{
		IsActive: &isActive,
		Version:  1,
	})
	require.ErrorIs(t, err, apigateway.ErrVersionConflict)
}
