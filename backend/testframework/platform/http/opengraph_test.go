package http_test
import (
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOpenGraphHTML(t *testing.T) {
	doc := `<!DOCTYPE html><html><head>
<title>HTML Title</title>
<meta property="og:title" content="OG Title" />
<meta property="og:description" content="A description" />
<meta property="og:image" content="https://cdn.example/img.png" />
</head><body></body></html>`

	meta := httputil.ExportParseOpenGraphHTML(doc)
	require.NotNil(t, meta)
	assert.Equal(t, "OG Title", meta.Title)
	assert.Equal(t, "A description", meta.Description)
	assert.Equal(t, "https://cdn.example/img.png", meta.ImageURL)
}

func TestParseOpenGraphHTML_FallbackTitle(t *testing.T) {
	doc := `<html><head><title>Plain Title</title></head></html>`
	meta := httputil.ExportParseOpenGraphHTML(doc)
	require.NotNil(t, meta)
	assert.Equal(t, "Plain Title", meta.Title)
}

func TestParseOpenGraphHTML_EmptyReturnsNil(t *testing.T) {
	assert.Nil(t, httputil.ExportParseOpenGraphHTML("<html></html>"))
}

func TestFetchOpenGraphMetadata_RejectsUnsafeURL(t *testing.T) {
	_, err := httputil.FetchOpenGraphMetadata(context.Background(), "file:///etc/passwd")
	require.Error(t, err)
	assert.ErrorIs(t, err, httputil.ErrBlockedScheme)
}

func TestFetchOpenGraphMetadata_BlocksLoopback(t *testing.T) {
	_, err := httputil.FetchOpenGraphMetadata(context.Background(), "http://127.0.0.1:8080/")
	require.Error(t, err)
}
