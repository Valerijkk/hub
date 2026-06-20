package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_LoginTelegram_InvalidBody(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram", h.LoginTelegram)

	req := httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_LoginTelegram_InvalidInitData(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram", h.LoginTelegram)

	body, _ := json.Marshal(TelegramAuthRequest{InitData: "bad-data"})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_LoginTelegram_Success(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram", h.LoginTelegram)

	botToken := "bot"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	userJSON := `{"id":424242,"first_name":"Tele","username":"tguser"}`
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "q1")
	params.Set("hash", computeTelegramHash(t, botToken, params))

	body, _ := json.Marshal(TelegramAuthRequest{InitData: params.Encode()})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	count, err := svc.db.User.Query().Count(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestHandler_LoginTelegramCode_NoBot(t *testing.T) {
	h, _, _ := newHandlerTestEnv(t)
	app := testApp(http.MethodPost, "/auth/telegram-code", h.LoginTelegramCode)

	body, _ := json.Marshal(map[string]string{"code": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestHandler_LoginTelegramCode_MissingCode(t *testing.T) {
	h := &Handler{service: &AuthService{}, tgBot: &TelegramBotAuth{}}
	app := testApp(http.MethodPost, "/auth/telegram-code", h.LoginTelegramCode)

	req := httptest.NewRequest(http.MethodPost, "/auth/telegram-code", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthService_SetRegistrationMode(t *testing.T) {
	_, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))
	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "invite_only", mode)
}
