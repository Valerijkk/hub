package calls

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// TestGenerateToken проверяет генерацию JWT-токена для LiveKit.
// Токен должен быть непустым и содержать три сегмента JWT (header.payload.signature).
func TestGenerateToken(t *testing.T) {
	cfg := config.LiveKitConfig{
		URL:       "http://localhost:7880",
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	}

	svc := &CallService{
		cfg: cfg,
	}

	token, err := svc.GenerateToken("test-room", "user-123", "Test User")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// JWT состоит из трёх частей разделённых точкой
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3, "JWT токен должен содержать 3 сегмента")
}

// TestGenerateToken_DifferentRooms проверяет что токены для разных комнат различаются.
func TestGenerateToken_DifferentRooms(t *testing.T) {
	cfg := config.LiveKitConfig{
		URL:       "http://localhost:7880",
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	}

	svc := &CallService{
		cfg: cfg,
	}

	token1, err := svc.GenerateToken("room-1", "user-1", "User 1")
	require.NoError(t, err)

	token2, err := svc.GenerateToken("room-2", "user-1", "User 1")
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2, "токены для разных комнат должны отличаться")
}

// TestPublishJSONNilSafe проверяет что PublishJSON не паникует при nil NATS client.
func TestPublishJSONNilSafe(t *testing.T) {
	svc := &CallService{
		nats: nil,
	}

	// Не должно быть паники — PublishJSON безопасен при nil receiver
	assert.NotPanics(t, func() {
		svc.nats.PublishJSON("test.subject", map[string]string{"key": "value"})
	})
}

// TestVerifyWebhookSignature_EmptyAuth проверяет отклонение пустого Authorization header.
func TestVerifyWebhookSignature_EmptyAuth(t *testing.T) {
	cfg := config.LiveKitConfig{
		APIKey:    "testkey",
		APISecret: "testsecret",
	}
	h := &Handler{
		cfg: cfg,
	}

	result := h.verifyWebhookSignature("", []byte(`{}`))
	assert.False(t, result, "пустой Authorization должен быть отклонён")
}

// TestVerifyWebhookSignature_InvalidJWT проверяет отклонение невалидного JWT.
func TestVerifyWebhookSignature_InvalidJWT(t *testing.T) {
	cfg := config.LiveKitConfig{
		APIKey:    "testkey",
		APISecret: "testsecret",
	}
	h := &Handler{
		cfg: cfg,
	}

	result := h.verifyWebhookSignature("not-a-jwt", []byte(`{}`))
	assert.False(t, result, "невалидный JWT должен быть отклонён")
}
