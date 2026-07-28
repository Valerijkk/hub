package messenger_test

// P2-BE-07: invite-handlers отдают Forbidden→403 без internals и TZ JSON.
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/internal/platform/middleware"
)

func setupMessengerHandlerAppWithErrorHandler(t *testing.T) (*fiber.App, *mockRepository, uuid.UUID) {
	t.Helper()
	svc, repo, _ := newTestService()
	h := messenger.NewHandler(svc)
	userID := uuid.New()

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.RequestID())
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)
	return app, repo, userID
}

// TestCreateInviteLink_ForbiddenHTTP_NoInternals: member → 403, TZ body, без sql/ent.
func TestCreateInviteLink_ForbiddenHTTP_NoInternals(t *testing.T) {
	app, repo, memberID := setupMessengerHandlerAppWithErrorHandler(t)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)

	body := `{"expires_in":"1d","max_uses":5}`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/messenger/channels/"+channelID.String()+"/invites",
		bytes.NewReader([]byte(body)),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Contains(t, got, "error")
	assert.Contains(t, got, "request_id")
	_, hasCode := got["code"]
	_, hasMessage := got["message"]
	assert.False(t, hasCode)
	assert.False(t, hasMessage)
	assert.Len(t, got, 2)

	errStr, _ := got["error"].(string)
	assert.NotContains(t, errStr, "sql:")
	assert.NotContains(t, errStr, "pq:")
	assert.NotContains(t, errStr, "ent:")
	assert.NotContains(t, errStr, "stack")
}

// TestListChannelInvites_ForbiddenHTTP: member → 403 через return err (не 500+err.Error).
func TestListChannelInvites_ForbiddenHTTP(t *testing.T) {
	app, repo, memberID := setupMessengerHandlerAppWithErrorHandler(t)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/messenger/channels/"+channelID.String()+"/invites",
		nil,
	)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
