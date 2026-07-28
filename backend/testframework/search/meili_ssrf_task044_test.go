package search_test

// TASK-044 / P2-BE-18: Meili host SSRF — metadata / link-local → nil indexer.
import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/search"
)

func TestValidateMeiliHost_BlocksMetadata_P2BE18(t *testing.T) {
	err := search.ExportValidateMeiliHost("http://169.254.169.254/")
	require.Error(t, err)
	assert.ErrorIs(t, err, safehttp.ErrBlockedAddress)

	err = search.ExportValidateMeiliHost("http://metadata.google.internal/")
	require.Error(t, err)
	assert.ErrorIs(t, err, safehttp.ErrBlockedAddress)

	// SafeDial: loopback тоже private — не whitelist для Meili host.
	err = search.ExportValidateMeiliHost("http://127.0.0.1:7700")
	require.Error(t, err)
	assert.ErrorIs(t, err, safehttp.ErrBlockedAddress)

	require.NoError(t, search.ExportValidateMeiliHost("https://meili.example.com"))
}

func TestNewMeiliIndexerFromEnv_RejectsMetadataHost_P2BE18(t *testing.T) {
	t.Setenv("MEILI_HOST", "http://169.254.169.254/")
	t.Setenv("MEILI_MASTER_KEY", "x")
	idx := search.NewMeiliIndexerFromEnv()
	assert.Nil(t, idx, "SSRF guard: metadata host → nil indexer")
}
