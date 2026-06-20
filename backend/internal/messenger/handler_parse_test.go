package messenger

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChannelID_Validation(t *testing.T) {
	t.Run("invalid uuid", func(t *testing.T) {
		app := fiber.New()
		app.Get("/channels/:channelId", func(c fiber.Ctx) error {
			_, err := parseChannelID(c)
			return err
		})
		req := httptest.NewRequest(fiber.MethodGet, "/channels/not-a-uuid", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid uuid", func(t *testing.T) {
		id := uuid.New()
		var got uuid.UUID
		app := fiber.New()
		app.Get("/channels/:channelId", func(c fiber.Ctx) error {
			var err error
			got, err = parseChannelID(c)
			return err
		})
		req := httptest.NewRequest(fiber.MethodGet, "/channels/"+id.String(), nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
}

func TestParseMessageID_ReturnsParam(t *testing.T) {
	const msgID = "507f1f77bcf86cd799439011"
	var got string
	app := fiber.New()
	app.Get("/messages/:messageId", func(c fiber.Ctx) error {
		got = parseMessageID(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest(fiber.MethodGet, "/messages/"+msgID, nil)
	_, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, msgID, got)
}

func TestSearchMessages_InvalidChannelID(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.SearchMessages(t.Context(), "hello", "bad-uuid", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel_id")
}

func TestHeartbeat_NilPresenceNoOp(t *testing.T) {
	svc, _, _ := newTestService()
	err := svc.Heartbeat(t.Context(), uuid.New().String())
	require.NoError(t, err)
}

func TestGetPresence_NilServiceReturnsEmpty(t *testing.T) {
	svc, _, _ := newTestService()
	statuses, err := svc.GetPresence(t.Context(), []string{uuid.New().String()})
	require.NoError(t, err)
	assert.Empty(t, statuses)
}

func TestGetUnreadCounts_Empty(t *testing.T) {
	svc, _, _ := newTestService()
	counts, err := svc.GetUnreadCounts(t.Context(), uuid.New().String())
	require.NoError(t, err)
	assert.Empty(t, counts)
}
