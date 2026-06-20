package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func newAuthMiddlewareEnv(t *testing.T) (*AuthService, *ent.Client) {
	t.Helper()
	dsn := "file:authtest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := NewMagicLinkService(rdb, nil, "test@example.com")
	svc := NewAuthService(client, rdb, nil, "mw-secret", 15*time.Minute, 30*24*time.Hour, magic, "bot", nil)
	return svc, client
}

func TestUserStatusCacheKey(t *testing.T) {
	uid := uuid.New()
	assert.Equal(t, "user:status:"+uid.String(), UserStatusCacheKey(uid))
}

func TestInvalidateUserStatusCache_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() {
		InvalidateUserStatusCache(context.Background(), nil, uuid.New())
	})
}

func TestRequireAuth_ValidToken(t *testing.T) {
	svc, client := newAuthMiddlewareEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("mw-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := func(c fiber.Ctx) error {
		return c.SendString(c.Locals("userID").(string))
	}
	app.Get("/me", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	svc, _ := newAuthMiddlewareEnv(t)
	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_DeactivatedUser(t *testing.T) {
	svc, client := newAuthMiddlewareEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("deact-" + uuid.New().String()[:6] + "@example.com").
		SetStatus(user.StatusDeactivated).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAuthService_FacadeDelegates(t *testing.T) {
	svc, client := newAuthMiddlewareEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("facade-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "open", mode)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	claims, err := svc.ParseAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, u.ID.String(), claims.UserID)

	got, err := svc.GetMe(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)

	exists, err := svc.UserExistsByEmail(ctx, u.Email)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRequireAuth_CacheHitDeactivated(t *testing.T) {
	svc, client := newAuthMiddlewareEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("cache-deact-" + uuid.New().String()[:6] + "@example.com").
		SetStatus(user.StatusDeactivated).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	require.NoError(t, svc.redis.Set(ctx, UserStatusCacheKey(u.ID), userStatusCacheDeactivated, userStatusCacheTTL).Err())

	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	svc, _ := newAuthMiddlewareEnv(t)
	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer not-valid-jwt")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_MalformedAuthHeader(t *testing.T) {
	svc, _ := newAuthMiddlewareEnv(t)
	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Token abc")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
