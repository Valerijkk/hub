package auth_test

// Тесты Resend email-клиента (magic link). Telegram login suite удалён (ADR-AUTH-01 / TASK-102).
import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/auth"
)

// TestNewResendEmailClient проверяет сценарий: new resend email client.
func TestNewResendEmailClient(t *testing.T) {
	c := auth.NewResendEmailClient("key", "hub@example.com")
	require.NotNil(t, c)
	assert.Equal(t, "key", auth.ExportResendEmailClientAPIKey(c))
	assert.Equal(t, "hub@example.com", auth.ExportResendEmailClientFromEmail(c))
}

// TestResendEmailClient_SendMagicLinkEmail_Success — A3W6-2-004: Method+Bearer и тело (to/code/HTML).
func TestResendEmailClient_SendMagicLinkEmail_Success(t *testing.T) {
	const (
		toEmail = "user@example.com"
		code    = "123456"
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		var payload struct {
			To   []string `json:"to"`
			HTML string   `json:"html"`
		}
		require.NoError(t, json.Unmarshal(raw, &payload))
		assert.Contains(t, payload.To, toEmail)
		assert.Contains(t, payload.HTML, code)
		assert.NotEmpty(t, payload.HTML)

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := auth.NewResendEmailClient("test-key", "hub@example.com")
	auth.ExportResendEmailClientInjectHTTP(client, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	})

	err := client.SendMagicLinkEmail(context.Background(), toEmail, code)
	require.NoError(t, err)
}

// TestResendEmailClient_SendMagicLinkEmail_APIError проверяет сценарий: resend email client отправка magic link email api ошибка.
func TestResendEmailClient_SendMagicLinkEmail_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	client := auth.NewResendEmailClient("test-key", "hub@example.com")
	auth.ExportResendEmailClientInjectHTTP(client, &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = srv.Listener.Addr().String()
			req.URL.Path = "/emails"
			return http.DefaultTransport.RoundTrip(req)
		}),
	})

	err := client.SendMagicLinkEmail(context.Background(), "user@example.com", "123456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
