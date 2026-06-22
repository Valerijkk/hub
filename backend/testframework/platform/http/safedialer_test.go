package http_test
import (
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"context"
	"errors"
	"net"
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
}

func TestIsBlockedIP(t *testing.T) {
	assert.True(t, httputil.ExportIsBlockedIP(nil))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("127.0.0.1")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("10.0.0.5")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("192.168.1.1")))
	assert.True(t, httputil.ExportIsBlockedIP(net.ParseIP("169.254.169.254")))
	assert.False(t, httputil.ExportIsBlockedIP(net.ParseIP("8.8.8.8")))
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
