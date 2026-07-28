//go:build hubtest

package onec

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	hrmod "github.com/valerii/hub/backend/internal/hr"
)

// ExportSettingsKey возвращает сохранённый ключ настроек для hubtest.
func ExportSettingsKey() string {
	return settingsKey
}

// ExportConfigValidate оборачивает Config.validate для hubtest.
func ExportConfigValidate(c Config) error {
	return c.validate()
}

// ExportDecodeConfig оборачивает decodeConfig для hubtest.
func ExportDecodeConfig(raw string) (Config, error) {
	return decodeConfig(raw)
}

// ExportEncodeConfig оборачивает encodeConfig для hubtest.
func ExportEncodeConfig(c Config) (string, error) {
	return encodeConfig(c)
}

// ExportServiceSetClient внедряет HTTP-клиент для hubtest (httptest loopback).
func ExportServiceSetClient(s *Service, client *http.Client) {
	s.client = client
	s.skipOutboundValidate = true
}

// ExportServicePost оборачивает Service.post для hubtest.
func ExportServicePost(s *Service, ctx context.Context, cfg Config, body []byte) (int, error) {
	return s.post(ctx, cfg, body)
}

// ExportBuildLeavePayload оборачивает buildLeavePayload для hubtest.
func ExportBuildLeavePayload(
	req hrmod.LeaveRequestResponse,
	reviewerID uuid.UUID,
	email, phone, displayName string,
	field EmployeeIDField,
) LeaveWebhookPayload {
	return buildLeavePayload(req, reviewerID, email, phone, displayName, field)
}
