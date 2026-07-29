//go:build hubtest

package notification_test

// TASK-188 / ASN-002: Telegram link-code confirm rate limit (как HI-04).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/notification"
)

// TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002 —
// после ExportTGLinkCodeVerifyMax() (=5) неудач → ErrTelegramConfirmLimited;
// HTTP confirm → 429.
// Mutation: убрать fails-check / recordLinkCodeFailure в ConfirmLinkCode → RED.
func TestTelegramBind_ConfirmLinkCode_BruteForceLockout_ASN002(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	dsn := "file:tg_asn002_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("asn002@" + uuid.New().String()[:8] + ".local").
		Save(ctx)
	require.NoError(t, err)

	repo := notification.NewRepository(client)
	svc := notification.NewTelegramBindService(repo, rdb, nil, "123456789:aaaabbbbaaaabbbbaaaabbbbaaaabbbbccc", "")
	require.NotNil(t, svc)

	// Валидный код в Redis — цель брутфорса (как magic-link).
	require.NoError(t, mr.Set(notification.ExportTGLinkCodePrefix()+"654321", "99887766"))

	max := notification.ExportTGLinkCodeVerifyMax()
	for i := 0; i < max; i++ {
		err := svc.ConfirmLinkCode(ctx, u.ID, "000000")
		require.Error(t, err, "попытка %d должна провалиться", i+1)
		assert.True(t, errors.Is(err, notification.ErrTelegramInvalidCode),
			"попытка %d: ожидался ErrTelegramInvalidCode, got %v", i+1, err)
	}

	err = svc.ConfirmLinkCode(ctx, u.ID, "000000")
	require.Error(t, err)
	assert.True(t, errors.Is(err, notification.ErrTelegramConfirmLimited),
		"6-я попытка должна вернуть ErrTelegramConfirmLimited, got %v", err)

	// Даже правильный код во время lockout — отказ.
	err = svc.ConfirmLinkCode(ctx, u.ID, "654321")
	require.Error(t, err)
	assert.True(t, errors.Is(err, notification.ErrTelegramConfirmLimited))

	failsRaw, err := mr.Get(notification.ExportTGLinkCodeFailsPrefix() + u.ID.String())
	require.NoError(t, err)
	failsInt, err := strconv.Atoi(failsRaw)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, failsInt, max)

	// HTTP: confirm после lockout → 429.
	h := notification.NewHandler(notification.NewNotificationService(repo, nil, nil, nil), svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", u.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body, err := json.Marshal(notification.TelegramConfirmRequest{Code: "000000"})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/telegram/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
