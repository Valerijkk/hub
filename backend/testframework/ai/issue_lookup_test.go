package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

func TestTrackerIssueLookup_NilSafe(t *testing.T) {
	var l *ai.TrackerIssueLookup
	title, desc, err := l.IssueContext(context.Background(), uuid.New(), "HUB-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issue lookup not configured")
	assert.Empty(t, title)
	assert.Empty(t, desc)

	l = ai.NewTrackerIssueLookup(nil, nil)
	title, desc, err = l.IssueContext(context.Background(), uuid.New(), "HUB-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issue lookup not configured")
	assert.Empty(t, title)
}

// A2W5-A3-007: foreign project + no membership → NotFound (см. также bot_authz_task021).
func TestTrackerIssueLookup_ForeignProjectNotFound_TASK113(t *testing.T) {
	projID := uuid.New()
	l := ai.NewTrackerIssueLookup(&trackermod.TrackerService{}, stubProjectByKey{
		proj: &ent.Project{ID: projID, Key: "FRN"},
	})
	l.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projID: false}})

	title, desc, err := l.IssueContext(context.Background(), uuid.New(), "FRN-99")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	assert.Empty(t, title)
	assert.Empty(t, desc)
}

func TestOfflineOnlyEnabled(t *testing.T) {
	svc := ai.NewService(&mockRepo{}, &mockAdapter{}, nil)
	assert.False(t, svc.OfflineOnlyEnabled())
	svc.SetOfflineOnly(true)
	assert.True(t, svc.OfflineOnlyEnabled())
}
