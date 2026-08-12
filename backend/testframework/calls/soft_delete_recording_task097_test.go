package calls_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
)

// TestEntRepository_RecordingSoftDelete_A2W8_4_005 — SoftDeleteRecording; list/get hide.
// Mutation: drop DeletedAtIsNil from ListRecordingsByProject → soft-deleted listed → FAIL.
func TestEntRepository_RecordingSoftDelete_A2W8_4_005(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	proj, err := f.client.Project.Create().
		SetName("Rec Soft").
		SetKey("REC").
		Save(ctx)
	require.NoError(t, err)

	callEnt, err := f.client.Call.Create().
		SetRoomName("rec-" + uuid.New().String()).
		SetCreatedBy(f.owner.ID).
		SetStatus(call.StatusActive).
		SetNillableChannelID(&f.channel.ID).
		Save(ctx)
	require.NoError(t, err)

	rec, err := f.repo.CreateRecording(ctx, "egress-097", callEnt.ID, proj.ID, f.owner.ID, "recordings/x.mp4")
	require.NoError(t, err)

	require.NoError(t, f.repo.SoftDeleteRecording(ctx, rec.ID))

	_, err = f.repo.GetRecording(ctx, rec.ID)
	require.True(t, ent.IsNotFound(err))

	list, total, err := f.repo.ListRecordingsByProject(ctx, proj.ID, 0, 50)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, list)

	raw, err := f.client.Recording.Get(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "calls", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "SetDeletedAt")
	assert.Contains(t, string(src), "DeletedAtIsNil")
}
