package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	searchmod "github.com/valerii/hub/backend/internal/search"
)

func TestSearchDocProvider_NilSearch(t *testing.T) {
	_, err := ai.NewSearchDocProvider(nil).SearchContext(context.Background(), uuid.New(), "query")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	var p *ai.SearchDocProvider
	_, err = p.SearchContext(context.Background(), uuid.New(), "q")
	require.Error(t, err)

	_, err = ai.NewSearchDocProvider(&searchmod.SearchService{}).SearchContext(context.Background(), uuid.New(), "  ")
	require.Error(t, err)
}
