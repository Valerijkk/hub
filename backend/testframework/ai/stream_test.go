package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWSStreamPublisher_NilSafe(t *testing.T) {
	var p *ai.WSStreamPublisher
	require.NotPanics(t, func() {
		p.PublishStreamEvent(uuid.New(), "stream-1", "delta", nil)
	})

	p = ai.NewWSStreamPublisher(nil)
	require.NotPanics(t, func() {
		p.PublishStreamEvent(uuid.New(), "stream-1", "done", map[string]interface{}{"content": "hi"})
	})
}

func TestNewOpenAIAdapter(t *testing.T) {
	a := ai.NewOpenAIAdapter()
	require.NotNil(t, a)
	require.NotNil(t, ai.ExportOpenAIAdapterClient(a))
}
