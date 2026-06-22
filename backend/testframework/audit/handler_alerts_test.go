package audit_test

// Тесты HTTP-обработчика аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

// TestHandler_ReceiveAlertmanagerWebhook_RecordsAudit проверяет HTTP-слой: HTTP-обработчик receive alertmanager webhook records audit.
func TestHandler_ReceiveAlertmanagerWebhook_RecordsAudit(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	h := audit.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterAlertWebhook(api, "test-alert-secret")

	body, err := json.Marshal(audit.AlertmanagerWebhookPayload{
		Status: "firing",
		Alerts: []audit.AlertmanagerAlert{
			{
				Status:      "firing",
				Fingerprint: "fp-1",
				Labels:      map[string]string{"alertname": "HubHighHTTP5xxRate", "severity": "warning"},
				Annotations: map[string]string{"summary": "5xx elevated"},
			},
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(fiber.MethodPost, "/api/webhooks/alertmanager", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-alert-secret")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	require.Len(t, repo.logs, 1)
	require.Equal(t, "alert.firing", repo.logs[0].Action)
	require.Equal(t, "alert", repo.logs[0].EntityType)
}

// TestHandler_ReceiveAlertmanagerWebhook_RejectsBadSecret проверяет HTTP-слой: HTTP-обработчик receive alertmanager webhook rejects bad секрет.
func TestHandler_ReceiveAlertmanagerWebhook_RejectsBadSecret(t *testing.T) {
	repo := &mockAuditRepo{}
	h := audit.NewHandler(audit.NewAuditService(repo, nil))

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterAlertWebhook(api, "test-alert-secret")

	req := httptest.NewRequest(fiber.MethodPost, "/api/webhooks/alertmanager", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", "Bearer wrong")

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.Empty(t, repo.logs)
}

// TestRegisterAlertWebhook_DisabledWhenSecretEmpty проверяет сценарий: register alert webhook disabled when секрет пустой.
func TestRegisterAlertWebhook_DisabledWhenSecretEmpty(t *testing.T) {
	h := audit.NewHandler(audit.NewAuditService(&mockAuditRepo{}, nil))
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterAlertWebhook(api, "")

	req := httptest.NewRequest(fiber.MethodPost, "/api/webhooks/alertmanager", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
