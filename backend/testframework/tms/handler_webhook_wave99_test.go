package tms_test

// Тесты HTTP-обработчика TMS (волна 99): интеграционные и unit-сценарии testframework.
import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_WebhookAllure_MultipartMissingFile проверяет отклонение при HTTP-обработчик webhook allure multipart отсутствующий file.
func TestHandler_WebhookAllure_MultipartMissingFile(t *testing.T) {
	app, projectID := setupTMSWebhookApp(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	require.NoError(t, w.WriteField("title", "CI multipart"))
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, base, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
