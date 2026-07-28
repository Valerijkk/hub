package apigateway_test
import (
	"github.com/valerii/hub/backend/internal/apigateway"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// useUnsafeClient временно подменяет пакетный webhookHTTPClient на клиент без
// SafeDialContext, чтобы httptest.NewServer (127.0.0.1) был достижим из тестов.
// Возвращает функцию-восстановитель, которую необходимо отложить с defer.
// Это локальное изменение состояния пакета (не production-конфига) — после
// возвращения управления старый клиент возвращается на место.
func useUnsafeClient(t *testing.T) {
	t.Helper()
	prev := apigateway.ExportSetWebhookHTTPClient(&http.Client{Timeout: 5 * time.Second})
	t.Cleanup(func() {
		apigateway.ExportResetWebhookHTTPClient(prev)
	})
}

// localTestURL — httptest.Server URL с hostname localhost вместо 127.0.0.1.
// W4-A1-009: ValidateOutboundURL режет IP-литералы из blocklist; dial к
// localhost закрывает useUnsafeClient (без SafeDialContext).
func localTestURL(srv *httptest.Server) string {
	u, err := url.Parse(srv.URL)
	if err != nil {
		return srv.URL
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return srv.URL
	}
	if host == "127.0.0.1" || host == "::1" {
		u.Host = net.JoinHostPort("localhost", port)
	}
	return u.String()
}

// TestSendWebhook_BlocksLoopbackSSRF проверяет, что dispatcher отказывается
// дёргать loopback-адрес. Это гард-тест для CR-03: даже если злоумышленник
// (project owner) укажет http://127.0.0.1:<port>/ в качестве webhook URL,
// сервер не должен делать запрос к собственным внутренним сервисам.
//
// LO-06: тест намеренно работает в обход HTTP-mock и реальной сети — он
// гоняет тот же путь, что и production: ValidateOutboundURL до отправки и
// SafeDialContext внутри webhookHTTPClient. Деградация любого из этих
// уровней (например, переход на стандартный http.Client) будет поймана
// этим тестом.
func TestSendWebhook_BlocksLoopbackSSRF(t *testing.T) {
	d := &apigateway.WebhookDispatcher{}

	cases := []string{
		"http://127.0.0.1:9999/",       // IPv4 loopback (CR-03 пример)
		"http://localhost:9999/",       // имя, резолвящееся в loopback
		"http://169.254.169.254/meta",  // link-local: AWS instance metadata
		"http://10.0.0.5/",             // RFC1918
		"http://192.168.0.1/",          // RFC1918
		"http://[::1]/",                // IPv6 loopback
	}

	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			status, body, err := apigateway.ExportWebhookSendWebhook(d, u, []byte(`{"ping":true}`), "sig", "test.event")

			require := assert.New(t)
			require.Error(err, "ожидалась ошибка для %s", u)
			require.Zero(status, "status code должен быть 0 при отказе доставки")
			require.Empty(body)
			msg := strings.ToLower(err.Error())
			// Принимаем любой из двух признаков отказа:
			//   - URL отвергнут ValidateOutboundURL до dial (текст: "блок"/"отклонён")
			//   - dial отвергнут SafeDialContext (текст: "blocked private address")
			ok := strings.Contains(msg, "private") ||
				strings.Contains(msg, "blocked") ||
				strings.Contains(msg, "отклонён небезопасный url")
			require.Truef(ok, "ошибка не похожа на отказ SSRF-защиты: %v", err)
		})
	}
}

// TestSendWebhook_RejectsBadScheme — проверяет, что file:// и пустые
// URL отвергаются ValidateOutboundURL до отправки.
func TestSendWebhook_RejectsBadScheme(t *testing.T) {
	d := &apigateway.WebhookDispatcher{}

	cases := []string{
		"file:///etc/passwd",
		"gopher://attacker.local/",
		"http://user:pass@example.com/", // userinfo
		"http://", // empty host
	}

	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			status, _, err := apigateway.ExportWebhookSendWebhook(d, u, []byte("{}"), "sig", "evt")
			assert.Error(t, err)
			assert.Zero(t, status)
		})
	}
}

// TestValidateOutboundURL_RejectsKnownBad — отдельная проверка, что
// safehttp.ValidateOutboundURL (на которой строится sendWebhook) ловит
// классические SSRF-схемы. Закрывает риск регресса при будущей замене
// импорта/реорганизации helpers.
func TestValidateOutboundURL_RejectsKnownBad(t *testing.T) {
	bad := []string{
		"file:///etc/passwd",
		"javascript:alert(1)",
		"data:text/html,<script>",
		"ftp://example.com/",
		"http://", // пустой host
	}
	for _, u := range bad {
		t.Run(u, func(t *testing.T) {
			assert.Error(t, safehttp.ValidateOutboundURL(u), "ожидалась ошибка для %s", u)
		})
	}
}

// TestSendWebhook_AllowsPublicURL_200 проверяет, что sendWebhook успешно
// доставляет на публичный URL (в тесте — httptest.Server). Зеркало гард-
// теста: SSRF-защита блокирует loopback на production-клиенте, но "обычный"
// HTTP-путь должен работать корректно (статус, тело).
func TestSendWebhook_AllowsPublicURL_200(t *testing.T) {
	useUnsafeClient(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	d := &apigateway.WebhookDispatcher{}
	status, body, err := apigateway.ExportWebhookSendWebhook(d, localTestURL(srv), []byte(`{"ping":true}`), "sig", "test.event")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"ok":true}`, body)
}

// TestSendWebhook_5xxRetries проверяет полный deliverWithRetry path при 5xx
// (A2W4-5-005): не ручной 2× sendWebhook — через ExportWebhookDispatcherDeliverWithRetry.
// Короткие delays — чтобы мутация delays/delivery writes ловилась быстро.
func TestSendWebhook_5xxRetries(t *testing.T) {
	useUnsafeClient(t)

	prev := apigateway.ExportSetWebhookRetryDelays([]time.Duration{
		0, 10 * time.Millisecond, 10 * time.Millisecond,
	})
	t.Cleanup(func() { apigateway.ExportSetWebhookRetryDelays(prev) })

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"err":"down"}`))
	}))
	t.Cleanup(srv.Close)

	epID := uuid.New()
	repo := &stubWebhookRepo{
		endpoints: []*ent.WebhookEndpoint{{
			ID: epID, URL: localTestURL(srv), Secret: "secret", Events: []string{"*"}, IsActive: true,
		}},
	}
	d := apigateway.NewWebhookDispatcher(repo, nil)
	body := []byte(`{"x":1}`)
	payload := map[string]interface{}{"x": float64(1)}

	apigateway.ExportWebhookDispatcherDeliverWithRetry(
		d, context.Background(), repo.endpoints[0], "project.x.issue.created", body, payload,
	)

	assert.GreaterOrEqual(t, int(hits.Load()), 2, "сервер должен быть дёрнут ≥ 2 раз через deliverWithRetry")
	assert.GreaterOrEqual(t, len(repo.deliveries), 2, "каждая попытка пишет webhook_delivery")
	assert.Equal(t, http.StatusServiceUnavailable, repo.deliveries[0].StatusCode)
	assert.Equal(t, epID, repo.deliveries[0].EndpointID)
}

// TestSendWebhook_SignatureHeaderPresent — сервер читает заголовок
// X-Hub-Signature-256 и сравнивает его с собственным HMAC(secret, body).
// Гарантирует контракт подписи на уровне готовых HTTP-запросов
// (а не только signPayload). Регресс этого теста сделает все подписанные
// эндпоинты несовместимыми с подписью.
func TestSendWebhook_SignatureHeaderPresent(t *testing.T) {
	useUnsafeClient(t)

	const secret = "shared-secret-12345"
	body := []byte(`{"event":"issue.created","id":7}`)
	expected := "sha256=" + apigateway.ExportSignPayload(body, secret)

	var receivedHeader, receivedEvent string
	var bodyOK bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Hub-Signature-256")
		receivedEvent = r.Header.Get("X-Hub-Event")
		buf, _ := io.ReadAll(r.Body)
		// Дополнительно: сверяемся в test-server вручную, что подпись валидна
		// против полученного body — это симулирует то, что делает реальный
		// получатель (receiver).
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(buf)
		want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		bodyOK = hmac.Equal([]byte(want), []byte(receivedHeader))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	d := &apigateway.WebhookDispatcher{}
	sig := apigateway.ExportSignPayload(body, secret)
	_, _, err := apigateway.ExportWebhookSendWebhook(d, localTestURL(srv), body, sig, "issue.created")
	require.NoError(t, err)

	assert.Equal(t, expected, receivedHeader, "X-Hub-Signature-256 mismatch")
	assert.Equal(t, "issue.created", receivedEvent, "X-Hub-Event must propagate event type")
	assert.True(t, bodyOK, "server-side HMAC verification must succeed")
}

// TestSendWebhook_ResponseBodyTruncated — даже если эндпоинт вернул
// 10KB ответа, в наш WebhookDelivery попадёт не более
// maxWebhookResponseBodyBytes (4096). CR-03: вебхук как SSRF read-primitive
// (атакующий настраивает «свой» URL и читает внутренние ответы) ограничен
// по размеру.
func TestSendWebhook_ResponseBodyTruncated(t *testing.T) {
	useUnsafeClient(t)

	hugeSize := apigateway.ExportMaxWebhookResponseBodyBytes() * 3 // ~12 КБ
	huge := strings.Repeat("A", hugeSize)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(huge))
	}))
	t.Cleanup(srv.Close)

	d := &apigateway.WebhookDispatcher{}
	status, body, err := apigateway.ExportWebhookSendWebhook(d, localTestURL(srv), []byte(`{}`), "sig", "evt")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.LessOrEqual(t, len(body), apigateway.ExportMaxWebhookResponseBodyBytes(),
		"response body must be truncated to ≤ maxWebhookResponseBodyBytes")
	// Точно ограничено до 4096 (по контракту sendWebhook).
	assert.Equal(t, apigateway.ExportMaxWebhookResponseBodyBytes(), len(body),
		"для тела размером > лимита должна быть точно maxWebhookResponseBodyBytes")
	// Все символы — это 'A' из исходника, никакого мусора.
	assert.Equal(t, strings.Repeat("A", apigateway.ExportMaxWebhookResponseBodyBytes()), body)
}
