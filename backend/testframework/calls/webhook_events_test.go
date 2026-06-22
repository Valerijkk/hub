package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/call"
)

func TestLiveKitWebhook_RoomStarted_WithService(t *testing.T) {
	fx := newCallsTestFixture(t)
	callEnt := fx.createChannelCall(t, fx.owner)

	cfg := liveKitWebhookCfg()
	h := calls.ExportNewHandlerForTest(fx.svc, nil, nil, cfg, nil)
	app := fiber.New()
	app.Post("/api/calls/webhook", h.HandleWebhook)

	body := []byte(`{"event":"room_started","room":{"name":"` + callEnt.RoomName + `"}}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	updated, err := fx.client.Call.Get(context.Background(), callEnt.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.StartedAt)
}

func TestLiveKitWebhook_ParticipantJoined_InvalidIdentity(t *testing.T) {
	fx := newCallsTestFixture(t)
	callEnt := fx.createChannelCall(t, fx.owner)

	cfg := liveKitWebhookCfg()
	h := calls.ExportNewHandlerForTest(fx.svc, nil, nil, cfg, nil)
	app := fiber.New()
	app.Post("/api/calls/webhook", h.HandleWebhook)

	body := []byte(`{"event":"participant_joined","room":{"name":"` + callEnt.RoomName + `"},"participant":{"identity":"not-a-uuid"}}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestLiveKitWebhook_RoomFinished_WithService(t *testing.T) {
	fx := newCallsTestFixture(t)
	callEnt := fx.createChannelCall(t, fx.owner)

	cfg := liveKitWebhookCfg()
	h := calls.ExportNewHandlerForTest(fx.svc, nil, nil, cfg, nil)
	app := fiber.New()
	app.Post("/api/calls/webhook", h.HandleWebhook)

	body := []byte(`{"event":"room_finished","room":{"name":"` + callEnt.RoomName + `"}}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	updated, err := fx.client.Call.Get(context.Background(), callEnt.ID)
	require.NoError(t, err)
	assert.Equal(t, call.StatusEnded, updated.Status)
}
