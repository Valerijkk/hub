package auth_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/auth"
)

var errGetUserSimulated = errors.New("simulated db failure")

type failingGetUserRepo struct {
	auth.Repository
}

func (r *failingGetUserRepo) GetUser(_ context.Context, _ uuid.UUID) (*ent.User, error) {
	return nil, errGetUserSimulated
}

// TestRequireAuth_DBErrorFailsClosed проверяет, что ошибка GetUser из БД
// возвращает 503 (fail-closed), а не пропускает запрос.
func TestRequireAuth_DBErrorFailsClosed(t *testing.T) {
	svc, client := newAuthMiddlewareEnv(t)
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("failclosed-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	baseRepo := auth.NewEntRepository(client)
	failSvc := auth.NewAuthService(
		&failingGetUserRepo{Repository: baseRepo},
		auth.ExportAuthServiceRedis(svc),
		nil,
		"mw-secret",
		15*time.Minute,
		30*24*time.Hour,
		auth.ExportAuthServiceMagicLink(svc),
		"bot",
		nil,
	)

	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, auth.RequireAuth(failSvc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}

// TestRequireAuth_InvalidUserIDClaim проверяет, что JWT с некорректным uid
// возвращает 401 (fail-closed), а не пропускает запрос.
func TestRequireAuth_InvalidUserIDClaim(t *testing.T) {
	svc, _ := newAuthMiddlewareEnv(t)

	now := time.Now().UTC()
	claims := auth.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "hub",
		},
		UserID:    "not-a-valid-uuid",
		SessionID: uuid.New().String(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("mw-secret"))
	require.NoError(t, err)

	app := fiber.New()
	terminal := func(c fiber.Ctx) error { return c.SendStatus(200) }
	app.Get("/", terminal, auth.RequireAuth(svc))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
