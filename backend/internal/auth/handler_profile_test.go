package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateProfile_Success(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("prof-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Put("/auth/me", h.UpdateProfile, RequireAuth(svc))

	name := "Updated Name"
	body, _ := json.Marshal(UpdateProfileRequest{Name: &name})
	req := httptest.NewRequest(http.MethodPut, "/auth/me", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), name)
}

func TestHandler_CompleteOnboarding_Success(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("onb-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/me/onboarding", h.CompleteOnboarding, RequireAuth(svc))

	display := "Dev User"
	body, _ := json.Marshal(OnboardingCompleteRequest{
		ProfileData: UpdateProfileRequest{DisplayName: &display},
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/me/onboarding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RevokeSession_Success(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("rev-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	sessions, err := svc.ListSessions(ctx, u.ID)
	require.NoError(t, err)
	require.NotEmpty(t, sessions)

	app := fiber.New()
	app.Delete("/auth/sessions/:sessionId", h.RevokeSession, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodDelete, "/auth/sessions/"+sessions[0].ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RevokeSession_InvalidID(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("revbad-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Delete("/auth/sessions/:sessionId", h.RevokeSession, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodDelete, "/auth/sessions/not-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("badprof-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Put("/auth/me", h.UpdateProfile, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPut, "/auth/me", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CompleteOnboarding_InvalidJSON(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	u, err := svc.db.User.Create().SetEmail("badonb-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, u.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/me/onboarding", h.CompleteOnboarding, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/auth/me/onboarding", bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
