package messenger_test

// Тесты HTTP-обработчика мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseChannelID_Validation проверяет сценарий: разбор channel id валидация.
func TestParseChannelID_Validation(t *testing.T) {
	t.Run("invalid uuid", func(t *testing.T) {
		app := fiber.New()
		app.Get("/channels/:channelId", func(c fiber.Ctx) error {
			_, err := messenger.ExportParseChannelID(c)
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
			got, err = messenger.ExportParseChannelID(c)
			return err
		})
		req := httptest.NewRequest(fiber.MethodGet, "/channels/"+id.String(), nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
}

// TestParseMessageID_ReturnsParam проверяет сценарий: разбор message id returns param.
func TestParseMessageID_ReturnsParam(t *testing.T) {
	const msgID = "507f1f77bcf86cd799439011"
	var got string
	app := fiber.New()
	app.Get("/messages/:messageId", func(c fiber.Ctx) error {
		got = messenger.ExportParseMessageID(c)
		return c.SendStatus(200)
	})
	req := httptest.NewRequest(fiber.MethodGet, "/messages/"+msgID, nil)
	_, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, msgID, got)
}

// TestSearchMessages_InvalidChannelID проверяет отклонение при search сообщения некорректный channel id.
func TestSearchMessages_InvalidChannelID(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.SearchMessages(t.Context(), uuid.New(), "hello", "bad-uuid", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel_id")
}

// TestHeartbeat_NilPresenceNoOp проверяет сценарий: heartbeat nil присутствие no op.
func TestHeartbeat_NilPresenceNoOp(t *testing.T) {
	svc, _, _ := newTestService()
	err := svc.Heartbeat(t.Context(), uuid.New().String())
	require.NoError(t, err)
}

// TestGetPresence_NilServiceReturnsEmpty проверяет бизнес-логику: получение присутствие nil сервис returns пустой.
func TestGetPresence_NilServiceReturnsEmpty(t *testing.T) {
	svc, _, _ := newTestService()
	statuses, err := svc.GetPresence(t.Context(), []string{uuid.New().String()})
	require.NoError(t, err)
	assert.Empty(t, statuses)
}

// TestGetUnreadCounts_Empty проверяет сценарий: получение непрочитанные counts пустой.
func TestGetUnreadCounts_Empty(t *testing.T) {
	svc, _, _ := newTestService()
	counts, err := svc.GetUnreadCounts(t.Context(), uuid.New().String())
	require.NoError(t, err)
	assert.Empty(t, counts)
}
