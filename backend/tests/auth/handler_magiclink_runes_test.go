package auth_test

// Тесты валидации длины magic-link кода по рунам (utf8.RuneCountInString),
// а не по байтам: multibyte-код из 6 рун должен проходить проверку длины,
// а код с иным числом рун — отклоняться с 400 независимо от байтовой длины.
import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/auth"
)

func postVerifyMagicLink(t *testing.T, app *fiber.App, email, code string) *http.Response {
	t.Helper()
	body, err := json.Marshal(auth.VerifyMagicLinkRequest{Email: email, Code: code})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/auth/magic-link/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	return resp
}

// TestVerifyMagicLink_MultibyteCode_PassesLengthCheck: «аааааа» — 6 рун,
// но 12 байт. Проверка длины должна пропустить такой код дальше в сервис
// (до фикса len() == 6 отклонял его как 12-символьный). Код неверный,
// поэтому сервис вернёт 401 — но НЕ 400 от валидации длины.
func TestVerifyMagicLink_MultibyteCode_PassesLengthCheck(t *testing.T) {
	h, _, _, magic := newHandlerTestEnv(t)
	ctx := t.Context()
	email := "runes-" + uuid.New().String()[:6] + "@example.com"

	_, err := magic.SendMagicLink(ctx, email)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/magic-link/verify", h.VerifyMagicLink)

	resp := postVerifyMagicLink(t, app, email, "аааааа")
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"6-рунный multibyte-код должен пройти валидацию длины и дойти до проверки кода: %s", string(respBody))
	assert.NotContains(t, string(respBody), "code must be 6 characters")
}

// TestVerifyMagicLink_ThreeEmojiCode_Rejected: 3 эмодзи — 3 руны (12 байт),
// это не 6 символов — валидация длины должна вернуть 400.
func TestVerifyMagicLink_ThreeEmojiCode_Rejected(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{})
	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)

	resp := postVerifyMagicLink(t, app, "emoji@example.com", "🔥🔥🔥")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestVerifyMagicLink_SixByteThreeRuneCode_Rejected: «ааа» — 3 руны, но
// ровно 6 байт. Старая байтовая проверка len() == 6 ошибочно ПРИНИМАЛА
// такой код; по рунам он должен отклоняться с 400.
func TestVerifyMagicLink_SixByteThreeRuneCode_Rejected(t *testing.T) {
	h := auth.NewHandler(&auth.AuthService{})
	app := testApp(http.MethodPost, "/auth/magic-link/verify", h.VerifyMagicLink)

	resp := postVerifyMagicLink(t, app, "bytes@example.com", "ааа")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
