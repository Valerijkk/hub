package calls

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAuthMiddleware(c fiber.Ctx) error {
	if uid := c.Get("X-Test-User-ID"); uid != "" {
		c.Locals("userID", uid)
	}
	if name := c.Get("X-Test-User-Name"); name != "" {
		c.Locals("userName", name)
	}
	return c.Next()
}

func newCallsValidationApp(h *Handler) *fiber.App {
	app := fiber.New()
	app.Use(testAuthMiddleware)

	app.Post("/calls/guest-join", h.GuestJoinCall)
	app.Post("/calls", h.CreateCall)
	app.Post("/calls/instant", h.CreateInstantCall)
	app.Get("/calls/invite/:code", h.JoinByInviteCode)
	app.Get("/calls/channel/:channelId", h.GetActiveCall)
	app.Post("/calls/:callId/join", h.JoinCall)
	app.Post("/calls/:callId/leave", h.LeaveCall)
	app.Post("/calls/:callId/end", h.EndCall)
	app.Post("/calls/:callId/raise-hand", h.RaiseHand)
	app.Post("/calls/:callId/chat", h.SendChatMessage)
	app.Get("/calls/:callId/chat", h.ListChatMessages)
	app.Get("/calls/conflicts", h.CheckConflicts)

	app.Post("/projects/:projectId/scheduled-calls", h.CreateScheduledCall)
	app.Get("/channels/:channelId/scheduled-calls", h.ListScheduledCalls)
	app.Put("/scheduled-calls/:id", h.UpdateScheduledCall)
	app.Delete("/scheduled-calls/:id", h.DeleteScheduledCall)

	return app
}

func TestUserNameFromCtx_Defaults(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(userNameFromCtx(c))
	})

	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestCreateCall_Validation(t *testing.T) {
	app := newCallsValidationApp(&Handler{})
	userID := uuid.New()

	req := httptest.NewRequest(fiber.MethodPost, "/calls", bytes.NewReader([]byte(`{"channel_id":"00000000-0000-0000-0000-000000000000"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", userID.String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreateCall_InvalidJSON(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodPost, "/calls", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestJoinByInviteCode_EmptyCode(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodGet, "/calls/invite/%20", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetActiveCall_InvalidChannelID(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodGet, "/calls/channel/not-uuid", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGuestJoinCall_Validation(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	tests := []struct {
		name string
		body string
	}{
		{name: "missing invite", body: `{"guest_name":"Guest"}`},
		{name: "missing name", body: `{"invite_code":"abc12345"}`},
		{name: "invalid json", body: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/calls/guest-join", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestSendChatMessage_Validation(t *testing.T) {
	app := newCallsValidationApp(&Handler{})
	callID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{name: "empty text", body: `{"text":""}`},
		{name: "invalid json", body: `{`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callID.String()+"/chat", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User-ID", uuid.New().String())

			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestRaiseHand_InvalidCallID(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodPost, "/calls/bad-id/raise-hand", bytes.NewReader([]byte(`{"raised":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRaiseHand_InvalidJSON(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/raise-hand", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCheckConflicts_InvalidTime(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodGet, "/calls/conflicts?time=not-a-time&duration=30", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestScheduledHandlers_InvalidIDs(t *testing.T) {
	app := newCallsValidationApp(&Handler{scheduledSvc: &ScheduledCallService{}})

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{fiber.MethodPost, "/projects/bad/scheduled-calls", `{}`},
		{fiber.MethodGet, "/channels/bad/scheduled-calls", ""},
		{fiber.MethodPut, "/scheduled-calls/bad", `{}`},
		{fiber.MethodDelete, "/scheduled-calls/bad", ""},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			var body *bytes.Reader
			if tt.body != "" {
				body = bytes.NewReader([]byte(tt.body))
			} else {
				body = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("X-Test-User-ID", uuid.New().String())

			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestListChatMessages_InvalidCallID(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodGet, "/calls/not-uuid/chat", nil)
	req.Header.Set("X-Test-User-ID", uuid.New().String())

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestJoinCall_Unauthenticated(t *testing.T) {
	app := newCallsValidationApp(&Handler{})

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+uuid.New().String()+"/join", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
