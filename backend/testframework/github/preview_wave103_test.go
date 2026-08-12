package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveExternalPreview_BlockedURL(t *testing.T) {
	svc := github.NewLinkPreviewService(&previewMockRepo{}, nil)
	resp := github.ExportLinkPreviewResolveExternalPreview(svc, context.Background(), "http://127.0.0.1:8080/page")
	assert.Nil(t, resp)
}

func TestResolveExternalPreview_ExampleCom(t *testing.T) {
	if testing.Short() {
		t.Skip("network test skipped in -short mode")
	}
	svc := github.NewLinkPreviewService(&previewMockRepo{}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp := github.ExportLinkPreviewResolveExternalPreview(svc, ctx, "https://example.com")
	if resp == nil {
		t.Skip("example.com unreachable in this environment")
	}
	assert.Equal(t, "external", resp.Type)
	assert.Equal(t, "example.com", resp.Domain)
	assert.NotEmpty(t, resp.Title)
}

func TestGetLinkPreview_ExternalHTTPS(t *testing.T) {
	if testing.Short() {
		t.Skip("network test skipped in -short mode")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
			<meta property="og:title" content="OG Title Wave103"/>
			<meta property="og:description" content="Desc"/>
			<meta property="og:image" content="/img/preview.png"/>
		</head><body></body></html>`))
	}))
	t.Cleanup(srv.Close)

	// SafeHTTP блокирует 127.0.0.1 — resolveExternalPreview через GetLinkPreview использует заблокированный хост.
	svc := github.NewLinkPreviewService(&previewMockRepo{}, nil)
	resp, err := svc.GetLinkPreview(context.Background(), uuid.New(), srv.URL)
	require.NoError(t, err)
	// Откатывается на внешнюю заглушку, когда загрузка OG заблокирована
	assert.Equal(t, "external", resp.Type)
}
