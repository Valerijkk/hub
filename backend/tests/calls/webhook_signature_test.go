//go:build hubtest

package calls_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestWebhookSignature_AcceptsValidHS256TokenBoundToBody(t *testing.T) {
	body := []byte(`{"event":"room_started"}`)
	handler := calls.ExportNewHandlerWithCfg(webhookConfig())

	assert.True(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, webhookConfig(), body, true, false), body))
}

func TestWebhookSignature_RejectsMissingOrEmptyBodyHash(t *testing.T) {
	body := []byte(`{"event":"room_finished"}`)
	handler := calls.ExportNewHandlerWithCfg(webhookConfig())

	assert.False(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, webhookConfig(), body, false, false), body))
	assert.False(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, webhookConfig(), body, true, true), body))
}

func TestWebhookSignature_RejectsTamperedBody(t *testing.T) {
	body := []byte(`{"event":"participant_joined"}`)
	handler := calls.ExportNewHandlerWithCfg(webhookConfig())

	assert.False(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, webhookConfig(), body, true, false), append(body, '!')))
}

func TestWebhookSignature_RejectsWrongIssuerOrSecret(t *testing.T) {
	body := []byte(`{"event":"egress_ended"}`)
	handler := calls.ExportNewHandlerWithCfg(webhookConfig())

	wrongIssuer := webhookConfig()
	wrongIssuer.APIKey = "other-key"
	assert.False(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, wrongIssuer, body, true, false), body))

	wrongSecret := webhookConfig()
	wrongSecret.APISecret = "other-secret"
	assert.False(t, calls.ExportVerifyWebhookSignature(handler, signedWebhookToken(t, wrongSecret, body, true, false), body))
}

func TestWebhookSignature_RejectsNonHMACJWT(t *testing.T) {
	body := []byte(`{"event":"room_started"}`)
	cfg := webhookConfig()
	claims := jwt.MapClaims{
		"iss":    cfg.APIKey,
		"sha256": webhookBodyHash(body),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	handler := calls.ExportNewHandlerWithCfg(cfg)
	assert.False(t, calls.ExportVerifyWebhookSignature(handler, token, body))
}

func webhookConfig() config.LiveKitConfig {
	return config.LiveKitConfig{APIKey: "webhook-key", APISecret: "webhook-secret"}
}

func signedWebhookToken(t *testing.T, cfg config.LiveKitConfig, body []byte, includeHash, emptyHash bool) string {
	t.Helper()
	claims := jwt.MapClaims{"iss": cfg.APIKey}
	if includeHash {
		claims["sha256"] = webhookBodyHash(body)
	}
	if emptyHash {
		claims["sha256"] = ""
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.APISecret))
	require.NoError(t, err)
	return token
}

func webhookBodyHash(body []byte) string {
	hash := sha256.Sum256(body)
	return base64.StdEncoding.EncodeToString(hash[:])
}
