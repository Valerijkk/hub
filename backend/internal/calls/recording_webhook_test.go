package calls

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// LiveKit webhook signature verification — guard test for handler.go:219.
//
// Контракт (см. verifyWebhookSignature):
//  1. Authorization header содержит JWT, подписанный API Secret (HMAC).
//  2. Claim "iss" == APIKey.
//  3. Claim "sha256" == base64(sha256(body)) (если присутствует).
//
// Тесты гоняют тот же путь, что и production handler: создают подписанный
// JWT, вызывают HandleWebhook через httptest, проверяют 200/401.

// liveKitWebhookCfg возвращает фиксированную тестовую LiveKit-конфигурацию.
func liveKitWebhookCfg() config.LiveKitConfig {
	return config.LiveKitConfig{
		URL:       "ws://livekit-stub:7880",
		APIKey:    "test-api-key",
		APISecret: "test-api-secret-1234567890",
	}
}

// signLiveKitWebhook собирает JWT, который подписывает LiveKit при отправке
// webhook'а: alg=HS256, iss=APIKey, sha256=base64(sha256(body)).
func signLiveKitWebhook(t *testing.T, body []byte, apiKey, apiSecret string) string {
	t.Helper()
	bodyHash := sha256.Sum256(body)
	claims := jwt.MapClaims{
		"iss":    apiKey,
		"sha256": base64.StdEncoding.EncodeToString(bodyHash[:]),
		"exp":    time.Now().Add(5 * time.Minute).Unix(),
		"iat":    time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(apiSecret))
	require.NoError(t, err)
	return signed
}

// newWebhookApp поднимает минимальный Fiber-app с одной маршрутом
// POST /api/calls/webhook, привязанным к Handler.HandleWebhook.
// recSvc/service передаются nil — для unit-теста сигнатуры этого достаточно,
// потому что 401 (signature mismatch) случается ДО разбора body и вызова
// service-методов; а для valid-signature теста мы выбираем event типа
// "room_started" с nil-room → handler не дёрнет service.
func newWebhookApp(t *testing.T, cfg config.LiveKitConfig) *fiber.App {
	t.Helper()
	h := &Handler{cfg: cfg}
	app := fiber.New()
	app.Post("/api/calls/webhook", h.HandleWebhook)
	return app
}

// TestLiveKitWebhook_ValidSignature_Accepted — корректно подписанный webhook
// принимается (200). Используется event "room_started" с nil-room, чтобы
// handler не вызывал nil-сервис (service:nil).
func TestLiveKitWebhook_ValidSignature_Accepted(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	// Тело без вложенных полей — handler пройдёт switch без выполнения
	// никаких ответвлений, потому что event.Room/Participant/EgressInfo == nil.
	body := []byte(`{"event":"room_started"}`)
	token := signLiveKitWebhook(t, body, cfg.APIKey, cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode,
		"корректно подписанный webhook должен вернуть 200")
}

// TestLiveKitWebhook_InvalidSignature_Rejected — webhook с JWT, подписанным
// чужим secret'ом, отклоняется (401 "невалидный webhook").
func TestLiveKitWebhook_InvalidSignature_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{"event":"room_started"}`)
	// Подписываем СОВСЕМ ДРУГИМ секретом — JWT.Parse выдаст signature error.
	token := signLiveKitWebhook(t, body, cfg.APIKey, "attacker-controlled-secret")

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
		"webhook с подписью из чужого secret'а должен вернуть 401")
}

// TestLiveKitWebhook_NoAuthHeader_Rejected — webhook без Authorization
// отклоняется (401).
func TestLiveKitWebhook_NoAuthHeader_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{"event":"room_started"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Authorization намеренно не выставлен.

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
		"webhook без Authorization должен вернуть 401")
}

// TestLiveKitWebhook_WrongIssuer_Rejected — JWT подписан правильным secret'ом,
// но iss != APIKey. Должен быть отклонён (401).
func TestLiveKitWebhook_WrongIssuer_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{"event":"room_started"}`)
	// Подписываем правильным secret'ом, но указываем чужой iss.
	token := signLiveKitWebhook(t, body, "another-api-key", cfg.APISecret)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
		"webhook с чужим iss должен вернуть 401")
}

// TestLiveKitWebhook_NoSha256Claim_Rejected — JWT подписан правильно (iss,
// secret), но НЕ содержит claim "sha256". До SEC-02 такой webhook проходил
// (body-hash проверка была опциональной), что позволяло атакующему с
// committed devkey forge'ить произвольный body. После фикса — 401.
func TestLiveKitWebhook_NoSha256Claim_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	body := []byte(`{"event":"room_started"}`)
	// Собираем JWT БЕЗ claim "sha256" — всё остальное валидно.
	claims := jwt.MapClaims{
		"iss": cfg.APIKey,
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tok.SignedString([]byte(cfg.APISecret))
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
		"webhook без claim sha256 должен вернуть 401 (SEC-02)")
}

// TestLiveKitWebhook_TamperedBody_Rejected — тело подменено после подписи.
// SHA256 hash в claims не совпадает с реальным body → 401.
//
// Это критический инвариант: иначе атакующий мог бы перехватить валидный
// webhook и заменить полезную нагрузку (например, поменять egress_id чтобы
// присвоить запись чужой звонок).
func TestLiveKitWebhook_TamperedBody_Rejected(t *testing.T) {
	cfg := liveKitWebhookCfg()
	app := newWebhookApp(t, cfg)

	// Подписываем один body...
	originalBody := []byte(`{"event":"room_started","room":{"name":"original"}}`)
	token := signLiveKitWebhook(t, originalBody, cfg.APIKey, cfg.APISecret)

	// ...а отправляем другой.
	tamperedBody := []byte(`{"event":"egress_ended","egress_info":{"egress_id":"hijack"}}`)

	req := httptest.NewRequest(fiber.MethodPost, "/api/calls/webhook", bytes.NewReader(tamperedBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode,
		"webhook с подменённым body должен вернуть 401")
}
