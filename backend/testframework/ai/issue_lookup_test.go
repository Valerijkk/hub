package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackerIssueLookup_NilSafe(t *testing.T) {
	var l *ai.TrackerIssueLookup
	title, desc, err := l.IssueContext(context.Background(), "HUB-1")
	require.NoError(t, err)
	assert.Empty(t, title)
	assert.Empty(t, desc)

	l = ai.NewTrackerIssueLookup(nil)
	title, desc, err = l.IssueContext(context.Background(), "HUB-1")
	require.NoError(t, err)
	assert.Empty(t, title)
}

func TestOfflineOnlyEnabled(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	assert.False(t, svc.OfflineOnlyEnabled())
	svc.SetOfflineOnly(true)
	assert.True(t, svc.OfflineOnlyEnabled())
}
