package http_test
import (
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutboundURL(t *testing.T) {
	t.Run("valid https", func(t *testing.T) {
		require.NoError(t, httputil.ValidateOutboundURL("https://example.com/path?q=1"))
	})

	t.Run("blocked scheme", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("file:///etc/passwd")
		require.Error(t, err)
		assert.ErrorIs(t, err, httputil.ErrBlockedScheme)
	})

	t.Run("userinfo not allowed", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("http://user:pass@example.com")
		require.ErrorIs(t, err, httputil.ErrUserInfoNotAllowed)
	})

	t.Run("empty host", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("https://")
		require.Error(t, err)
	})

	t.Run("invalid url", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("://bad")
		require.Error(t, err)
	})

	// W4-A1-009 / P3-BE-11: IP-литералы из blocklist → ErrBlocked на Validate (до dial).
	t.Run("blocked link-local IP literal", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("http://169.254.169.254/latest/meta-data/")
		require.ErrorIs(t, err, httputil.ErrBlockedAddress)
	})
	t.Run("blocked loopback IP literal", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("http://127.0.0.1:8080/")
		require.ErrorIs(t, err, httputil.ErrBlockedAddress)
	})
	t.Run("blocked IPv4-mapped literal", func(t *testing.T) {
		err := httputil.ValidateOutboundURL("http://[::ffff:169.254.169.254]/")
		require.ErrorIs(t, err, httputil.ErrBlockedAddress)
	})
	t.Run("public IP literal allowed at validate", func(t *testing.T) {
		require.NoError(t, httputil.ValidateOutboundURL("https://8.8.8.8/"))
	})
}

func TestIsBlockedIP(t *testing.T) {
	assert.True(t, httputil.ExportIsBlockedIP(nil))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("127.0.0.1")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("10.0.0.5")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("192.168.1.1")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("169.254.169.254")))
	assert.False(t, httputil.ExportIsBlockedIP(net.ParseIP("8.8.8.8")))

	// P1-BE-23: IPv4-mapped IPv6 must hit the same IPv4 blocklist.
	mappedBlocked := []string{
		"::ffff:127.0.0.1",
		"::ffff:169.254.169.254",
		"::ffff:10.0.0.1",
		"::ffff:192.168.0.1",
	}
	for _, s := range mappedBlocked {
		ip := net.ParseIP(s)
		require.NotNil(t, ip, s)
		assert.True(t, httputil.ExportIsBlockedIP(ip), "expected blocked: %s", s)
	}
	assert.False(t, httputil.ExportIsBlockedIP(net.ParseIP("::ffff:8.8.8.8")))
}

// TestNormalizeIPForBlocklist_IPv4MappedFourBytes_P1BE23 — мутация: drop To4() → len 16 → RED.
func TestNormalizeIPForBlocklist_IPv4MappedFourBytes_P1BE23(t *testing.T) {
	mapped := net.ParseIP("::ffff:169.254.169.254")
	require.NotNil(t, mapped)
	require.Equal(t, net.IPv6len, len(mapped), "ParseIP даёт 16-байтный mapped")

	n := httputil.ExportNormalizeIPForBlocklist(mapped)
	require.Equal(t, net.IPv4len, len(n), "P1-BE-23: blocklist должен видеть 4-байтный IPv4")
	require.True(t, n.Equal(net.IPv4(169, 254, 169, 254)))
	require.True(t, httputil.ExportIsBlockedIP(mapped))
}

func TestSafeDialContext_BlocksIPv4MappedLoopback(t *testing.T) {
	dial := httputil.SafeDialContext()
	_, err := dial(context.Background(), "tcp", "[::ffff:127.0.0.1]:80")
	require.Error(t, err)
	assert.ErrorIs(t, err, httputil.ErrBlockedAddress)
}

func TestSafeDialContext_BlocksIPv4MappedMetadata(t *testing.T) {
	dial := httputil.SafeDialContext()
	_, err := dial(context.Background(), "tcp", "[::ffff:169.254.169.254]:80")
	require.Error(t, err)
	assert.ErrorIs(t, err, httputil.ErrBlockedAddress)
}

func TestSafeDialContext_BlocksLoopback(t *testing.T) {
	dial := httputil.SafeDialContext()
	_, err := dial(context.Background(), "tcp", "127.0.0.1:80")
	require.Error(t, err)
	assert.True(t, errors.Is(err, httputil.ErrBlockedAddress), "expected ErrBlockedAddress, got %v", err)
}

func TestNewSafeHTTPClient_BlocksPrivateHost(t *testing.T) {
	client := httputil.NewSafeHTTPClient(3 * time.Second)
	_, err := client.Get("http://127.0.0.1:65500/")
	require.Error(t, err)
}

// --- P1-BE-U01 / TASK-101: offline AI loopback consent (не ослабляет SafeDial) ---

func TestValidateOfflineAIURL_Allowlist(t *testing.T) {
	ok := []string{
		"http://127.0.0.1:11434/v1",
		"http://localhost:11434/v1",
		"http://[::1]:11434/v1",
		"https://127.0.0.1/v1",
	}
	for _, u := range ok {
		require.NoError(t, httputil.ValidateOfflineAIURL(u), u)
	}
}

func TestValidateOfflineAIURL_RejectsNonLoopback(t *testing.T) {
	blocked := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.5:8080/",
		"http://192.168.1.1/",
		"https://api.openai.com/v1",
		"http://evil.example/ollama",
		"file:///etc/passwd",
		"http://user:pass@127.0.0.1:11434/",
	}
	for _, u := range blocked {
		err := httputil.ValidateOfflineAIURL(u)
		require.Error(t, err, u)
	}
}

func TestValidateOutboundURL_StillBlocksLoopback_P1BEU01(t *testing.T) {
	// Мутация «ослабить ValidateOutboundURL для loopback» → этот тест RED.
	err := httputil.ValidateOutboundURL("http://127.0.0.1:11434/v1")
	require.ErrorIs(t, err, httputil.ErrBlockedAddress)
}

func TestIsLoopbackIP(t *testing.T) {
	assert.True(t, httputil.ExportIsLoopbackIP(net.ParseIP("127.0.0.1")))
	assert.True(t, httputil.ExportIsLoopbackIP(net.ParseIP("::1")))
	assert.True(t, httputil.ExportIsLoopbackIP(net.ParseIP("::ffff:127.0.0.1")))
	assert.False(t, httputil.ExportIsLoopbackIP(net.ParseIP("169.254.169.254")))
	assert.False(t, httputil.ExportIsLoopbackIP(net.ParseIP("10.0.0.1")))
	assert.False(t, httputil.ExportIsLoopbackIP(net.ParseIP("8.8.8.8")))
	assert.False(t, httputil.ExportIsLoopbackIP(nil))
}

func TestSafeDialContextAllowLoopback_AllowsLoopbackDial(t *testing.T) {
	dial := httputil.SafeDialContextAllowLoopback()
	// Порт закрыт → connection refused, но НЕ ErrBlockedAddress (consent path).
	_, err := dial(context.Background(), "tcp", "127.0.0.1:1")
	require.Error(t, err)
	assert.False(t, errors.Is(err, httputil.ErrBlockedAddress), "loopback must not be SSRF-blocked: %v", err)
}

func TestSafeDialContextAllowLoopback_BlocksMetadata(t *testing.T) {
	dial := httputil.SafeDialContextAllowLoopback()
	_, err := dial(context.Background(), "tcp", "169.254.169.254:80")
	require.ErrorIs(t, err, httputil.ErrBlockedAddress)
}

func TestSafeDialContextAllowLoopback_BlocksRFC1918(t *testing.T) {
	dial := httputil.SafeDialContextAllowLoopback()
	_, err := dial(context.Background(), "tcp", "10.0.0.1:80")
	require.ErrorIs(t, err, httputil.ErrBlockedAddress)
}

func TestNewOfflineAIHTTPClient_DoesNotSSRFBlockLoopback(t *testing.T) {
	client := httputil.NewOfflineAIHTTPClient(2 * time.Second)
	_, err := client.Get("http://127.0.0.1:1/")
	require.Error(t, err)
	assert.False(t, errors.Is(err, httputil.ErrBlockedAddress), "offline client must dial loopback: %v", err)
}

func TestNewSafeHTTPClient_StillBlocksLoopback_P1BEU01(t *testing.T) {
	// Глобальный SafeDial не ослаблен: мутация «SafeDial = AllowLoopback» → RED.
	client := httputil.NewSafeHTTPClient(2 * time.Second)
	_, err := client.Get("http://127.0.0.1:1/")
	require.Error(t, err)
	assert.True(t, errors.Is(err, httputil.ErrBlockedAddress) ||
		strings.Contains(err.Error(), "blocked private address"),
		"safe client must still block loopback, got %v", err)
}
