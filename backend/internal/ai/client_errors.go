package ai

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// Стабильные client-facing сообщения (A2W5-A1-003 / A2W5-A1-004).
// Внутренности encrypt/base_url/upstream не отдаём клиенту.
const (
	msgInvalidProviderConfig = "invalid provider configuration"
	msgInvalidBaseURL        = "invalid base_url"
	msgEncryptionRequired    = "encryption required"
	msgProviderTestFailed    = "provider test failed"
	msgProviderNotFound      = "provider not found"
	msgInvalidProviderURL    = "invalid provider URL"
)

// mapProviderMutateError — Create/Update provider → HTTP status + стабильный текст.
func mapProviderMutateError(err error) (status int, msg string) {
	if err == nil {
		return 0, ""
	}
	if errors.Is(err, ErrProviderNotFound) {
		return fiber.StatusNotFound, msgProviderNotFound
	}
	if errors.Is(err, ErrEncryptionRequired) {
		return fiber.StatusBadRequest, msgEncryptionRequired
	}
	if errors.Is(err, safehttp.ErrBlockedAddress) ||
		errors.Is(err, safehttp.ErrBlockedScheme) ||
		errors.Is(err, safehttp.ErrUserInfoNotAllowed) ||
		strings.Contains(strings.ToLower(err.Error()), "invalid base url") {
		return fiber.StatusBadRequest, msgInvalidBaseURL
	}
	return fiber.StatusBadRequest, msgInvalidProviderConfig
}

// sanitizeProviderTestError — TestProvider JSON.error без upstream URL/auth/body.
func sanitizeProviderTestError(err error) string {
	if err == nil {
		return ""
	}
	slog.Debug("ai provider test failed", "err", err)
	if strings.Contains(strings.ToLower(err.Error()), "invalid provider url") {
		return msgInvalidProviderURL
	}
	return msgProviderTestFailed
}
