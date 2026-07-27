package calls_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsCallParticipantOrChannelMember_InstantParticipant_A2_007 —
// instant-звонок (ChannelID==nil): активный участник допускается к WS call:{id}.
// Mutation: deny when ChannelID==nil → FAIL.
func TestIsCallParticipantOrChannelMember_InstantParticipant_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "wsinst01")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	ok, err := f.repo.IsCallParticipantOrChannelMember(ctx, callEnt.ID, f.member.ID)
	require.NoError(t, err)
	assert.True(t, ok, "instant call participant must subscribe to call:{id}")
}

// TestIsCallParticipantOrChannelMember_InstantStrangerDenied_A2_007 —
// не-участник instant-звонка по-прежнему deny.
func TestIsCallParticipantOrChannelMember_InstantStrangerDenied_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "wsinst02")
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	ok, err := f.repo.IsCallParticipantOrChannelMember(ctx, callEnt.ID, f.stranger.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestIsCallParticipantOrChannelMember_ChannelMember_A2_007 —
// канальный звонок: член канала допускается без CallParticipant row.
func TestIsCallParticipantOrChannelMember_ChannelMember_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	ok, err := f.repo.IsCallParticipantOrChannelMember(ctx, callEnt.ID, f.member.ID)
	require.NoError(t, err)
	assert.True(t, ok, "channel member must subscribe to call:{id}")
}

// TestIsCallParticipantOrChannelMember_MissingCall_A2_007 — неизвестный call → deny.
func TestIsCallParticipantOrChannelMember_MissingCall_A2_007(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	ok, err := f.repo.IsCallParticipantOrChannelMember(ctx, uuid.New(), f.owner.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}
