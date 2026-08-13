//go:build hubtest

package calls_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TASK-037 / P2-BE-02 — Ent IsChannelMember stranger denial (guard для scheduled/calls).
// Mutation: IsChannelMember всегда true → stranger пройдёт CreateScheduledCall.

func TestIsChannelMember_RejectsStranger_Ent_TASK037(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	ok, err := f.repo.IsChannelMember(ctx, f.channel.ID, f.stranger.ID)
	require.NoError(t, err)
	require.False(t, ok, "stranger must not be channel member")

	ok, err = f.repo.IsChannelMember(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = f.repo.IsChannelMember(ctx, f.channel.ID, f.member.ID)
	require.NoError(t, err)
	require.True(t, ok)
}
