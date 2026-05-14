package apigateway

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

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
	d := &WebhookDispatcher{}

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
			status, body, err := d.sendWebhook(u, []byte(`{"ping":true}`), "sig", "test.event")

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
	d := &WebhookDispatcher{}

	cases := []string{
		"file:///etc/passwd",
		"gopher://attacker.local/",
		"http://user:pass@example.com/", // userinfo
		"http://", // empty host
	}

	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			status, _, err := d.sendWebhook(u, []byte("{}"), "sig", "evt")
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
