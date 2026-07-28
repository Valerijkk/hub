package apigateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/apigateway"
	"github.com/valerii/hub/backend/internal/rbac"
)

// webhookEnv — общая инфраструктура для тестов webhook CRUD с RBAC (APIGW-01).
type webhookEnv struct {
	client     *ent.Client
	rbacSvc    *rbac.RBACService
	webhookSvc *apigateway.WebhookService
	apiKeySvc  *apigateway.APIKeyService
	ctx        context.Context
	projectID  uuid.UUID
	ownerID    uuid.UUID
}

// newWebhookEnv создаёт ent-клиент, RBAC-сервис, проект и владельца-участника.
func newWebhookEnv(t *testing.T) *webhookEnv {
	t.Helper()
	dsn := "file:apigwwebhook_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx := context.Background()
	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))

	proj, err := client.Project.Create().
		SetName("Webhook Project").
		SetKey("WHK").
		Save(ctx)
	require.NoError(t, err)

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	repo := apigateway.NewEntRepository(client, apigatewayTestCipher(t))
	webhookSvc := apigateway.NewWebhookService(repo)
	apiKeySvc := apigateway.NewAPIKeyService(repo)
	t.Cleanup(apiKeySvc.Close)

	return &webhookEnv{
		client:     client,
		rbacSvc:    rbacSvc,
		webhookSvc: webhookSvc,
		apiKeySvc:  apiKeySvc,
		ctx:        ctx,
		projectID:  proj.ID,
		ownerID:    owner.ID,
	}
}

// buildApp регистрирует маршруты с auth-middleware, подставляющим actingUserID.
// Если actingUserID == uuid.Nil, userID не выставляется (имитация анонима).
func (e *webhookEnv) buildApp(t *testing.T, actingUserID uuid.UUID) *fiber.App {
	t.Helper()
	h := apigateway.NewHandler(e.apiKeySvc, e.webhookSvc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		if actingUserID != uuid.Nil {
			c.Locals("userID", actingUserID.String())
		}
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, e.rbacSvc)
	t.Cleanup(func() { _ = app.Shutdown() })
	return app
}

func TestHandler_CreateEndpoint_Validation(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)

	t.Run("invalid project id", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "https://example.com/hook",
			Events: []string{"issue.created"},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/projects/not-uuid/webhooks/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unsafe url rejected", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "file:///etc/passwd",
			Events: []string{"issue.created"},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/webhooks/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_CreateEndpoint_Success(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)

	body, _ := json.Marshal(apigateway.CreateWebhookRequest{
		URL:    "https://example.com/hub-webhook",
		Events: []string{"issue.created", "sprint.completed"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/webhooks/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_ListEndpoints_Empty(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+env.projectID.String()+"/webhooks/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var eps []apigateway.WebhookResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&eps))
	assert.Empty(t, eps)
}

func TestHandler_DeleteEndpoint_NotFound(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, env.ownerID)

	req := httptest.NewRequest(http.MethodDelete, "/api/webhooks/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHandler_Webhook_IDOR_ProjectScoped — APIGW-01: пользователь без членства
// в проекте не может создавать/просматривать вебхуки проекта (403).
func TestHandler_Webhook_IDOR_ProjectScoped(t *testing.T) {
	env := newWebhookEnv(t)

	stranger, err := env.client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@example.com").
		Save(env.ctx)
	require.NoError(t, err)

	app := env.buildApp(t, stranger.ID)

	t.Run("create forbidden for non-member", func(t *testing.T) {
		body, _ := json.Marshal(apigateway.CreateWebhookRequest{
			URL:    "https://attacker.example.com/steal",
			Events: []string{"*"},
		})
		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/webhooks/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("list forbidden for non-member", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/projects/"+env.projectID.String()+"/webhooks/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

// TestHandler_Webhook_IDOR_ByEndpointID — APIGW-01: чужой эндпоинт по ID нельзя
// обновить/удалить/прочитать доставки без членства в его проекте (403).
func TestHandler_Webhook_IDOR_ByEndpointID(t *testing.T) {
	env := newWebhookEnv(t)

	// Владелец создаёт легитимный эндпоинт.
	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	stranger, err := env.client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@example.com").
		Save(env.ctx)
	require.NoError(t, err)

	app := env.buildApp(t, stranger.ID)

	t.Run("update forbidden", func(t *testing.T) {
		newURL := "https://attacker.example.com/hook"
		body, _ := json.Marshal(apigateway.UpdateWebhookRequest{URL: &newURL})
		req := httptest.NewRequest(http.MethodPut, "/api/webhooks/"+created.ID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("delete forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/webhooks/"+created.ID, nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("deliveries forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/webhooks/"+created.ID+"/deliveries", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

// TestHandler_Webhook_Unauthenticated — без userID все CRUD-маршруты дают 401.
func TestHandler_Webhook_Unauthenticated(t *testing.T) {
	env := newWebhookEnv(t)
	app := env.buildApp(t, uuid.Nil)

	body, _ := json.Marshal(apigateway.CreateWebhookRequest{
		URL:    "https://example.com/hook",
		Events: []string{"issue.created"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/webhooks/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebhookService_ListEndpoints_ReturnsCreated(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	list, err := env.webhookSvc.ListEndpoints(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)
}

func TestWebhookService_UpdateEndpoint_InvalidURL(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	badURL := "file:///etc/passwd"
	_, err = env.webhookSvc.UpdateEndpoint(env.ctx, uuid.MustParse(created.ID), apigateway.UpdateWebhookRequest{
		URL: &badURL,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apigateway.ErrInvalidInput)
}

func TestWebhookService_ListDeliveries_Empty(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	deliveries, err := env.webhookSvc.ListDeliveries(env.ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Empty(t, deliveries)
}

// TestWebhookService_EndpointProjectID — APIGW-01: резолвинг проекта по ID
// эндпоинта; для несуществующего ID возвращается not-found.
func TestWebhookService_EndpointProjectID(t *testing.T) {
	env := newWebhookEnv(t)

	created, err := env.webhookSvc.CreateEndpoint(env.ctx, env.projectID, apigateway.CreateWebhookRequest{
		URL:    "https://hooks.example.com/events",
		Events: []string{"issue.updated"},
	})
	require.NoError(t, err)

	pid, err := env.webhookSvc.EndpointProjectID(env.ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, env.projectID, pid)

	_, err = env.webhookSvc.EndpointProjectID(env.ctx, uuid.New())
	require.Error(t, err)
	assert.True(t, apigateway.IsNotFound(err))
}
