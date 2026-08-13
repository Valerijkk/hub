//go:build hubtest

package calls_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/calls"
)

// TestUpdateScheduledCall_SoftDeleted_W3A5011 — Update soft-deleted серии → not found.
// Mutation: снять DeletedAtIsNil из UpdateScheduledCall → 200 на trash → FAIL.
func TestUpdateScheduledCall_SoftDeleted_W3A5011(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	proj, err := f.client.Project.Create().
		SetName("Sched Soft").
		SetKey("SSF").
		Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(48 * time.Hour)
	sc, err := f.repo.CreateScheduledCall(ctx, "Standup", "", at, "", f.channel.ID, proj.ID, f.owner.ID, 0, nil, nil)
	require.NoError(t, err)
	require.NoError(t, f.repo.DeleteScheduledCall(ctx, sc.ID))

	title := "mutated"
	_, err = f.repo.UpdateScheduledCall(ctx, sc.ID, calls.UpdateScheduledCallRequest{Title: &title})
	require.True(t, ent.IsNotFound(err), "soft-deleted scheduled call must not update, got %v", err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	_, err = scheduled.UpdateScheduledCall(ctx, sc.ID, f.owner.ID, calls.UpdateScheduledCallRequest{Title: &title})
	require.ErrorIs(t, err, calls.ErrCallNotFound)
}
