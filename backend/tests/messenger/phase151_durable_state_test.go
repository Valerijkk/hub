package messenger_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151DurableMongoIndexContract(t *testing.T) {
	expected := []messenger.DurableMongoIndexSpec{
		{Collection: "read_states", Keys: []string{"tenant_id", "user_id", "category", "target_kind", "target_id"}, Unique: true, Purpose: "cursor"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "channel_id", "root_message_id"}, Unique: true, Purpose: "identity"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "channel_id"}, Purpose: "cleanup"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "mode", "channel_id"}, Purpose: "user-list"},
		{Collection: "messages", Keys: []string{"channel_id", "created_at", "_id"}, Purpose: "timeline-head"},
		{Collection: "messages", Keys: []string{"channel_id", "thread_id", "created_at", "_id"}, Purpose: "thread"},
		{Collection: "messages", Keys: []string{"channel_id", "resolved_mention_user_ids", "created_at", "_id"}, Purpose: "mentions"},
		{Collection: "messages", Keys: []string{"channel_id", "post_create_effects_state", "created_at", "_id"}, Purpose: "pending-recovery"},
	}

	assert.ElementsMatch(t, expected, messenger.ExportPhase151DurableMongoIndexSpecs())
}

func TestPhase151ReadStateAdvanceDecisionUsesForwardOnlyTuple(t *testing.T) {
	stamp := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	initial := messenger.MessagePosition{CreatedAt: stamp, MessageID: "message-002"}
	later := messenger.MessagePosition{CreatedAt: stamp.Add(time.Second), MessageID: "message-003"}

	created, advanced, next := messenger.ExportPhase151DecideReadStateAdvance(nil, initial)
	assert.True(t, created)
	assert.False(t, advanced)
	assert.Equal(t, initial, next)

	created, advanced, next = messenger.ExportPhase151DecideReadStateAdvance(&initial, later)
	assert.False(t, created)
	assert.True(t, advanced)
	assert.Equal(t, later, next)

	created, advanced, next = messenger.ExportPhase151DecideReadStateAdvance(&later, later)
	assert.False(t, created)
	assert.False(t, advanced, "equal retry is a successful no-op")
	assert.Equal(t, later, next)

	olderTimestamp := messenger.MessagePosition{CreatedAt: stamp, MessageID: "message-999"}
	_, advanced, next = messenger.ExportPhase151DecideReadStateAdvance(&later, olderTimestamp)
	assert.False(t, advanced, "an older timestamp never moves the cursor backward")
	assert.Equal(t, later, next)

	equalUTCLaterID := messenger.MessagePosition{CreatedAt: stamp, MessageID: "message-003"}
	_, advanced, next = messenger.ExportPhase151DecideReadStateAdvance(&initial, equalUTCLaterID)
	assert.True(t, advanced, "equal UTC timestamps are ordered by canonical string message ID")
	assert.Equal(t, equalUTCLaterID, next)

	equalUTCOlderID := messenger.MessagePosition{CreatedAt: stamp, MessageID: "message-001"}
	_, advanced, next = messenger.ExportPhase151DecideReadStateAdvance(&initial, equalUTCOlderID)
	assert.False(t, advanced)
	assert.Equal(t, initial, next)
}

func TestPhase151ThreadFollowTransitionPreservesManualAuthority(t *testing.T) {
	stamp := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	trigger := messenger.MessagePosition{CreatedAt: stamp, MessageID: "reply-010"}
	head := messenger.MessagePosition{CreatedAt: stamp.Add(time.Minute), MessageID: "reply-011"}
	later := messenger.MessagePosition{CreatedAt: stamp.Add(2 * time.Minute), MessageID: "reply-012"}

	changed, auto := messenger.ExportPhase151DecideThreadFollowTransition(nil, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeAuto,
		Position: trigger,
	})
	require.True(t, changed)
	assert.Equal(t, messenger.ThreadFollowModeAuto, auto.Mode)
	assert.Equal(t, trigger, auto.Cursor, "a new auto-follow baselines the triggering tuple")

	changed, existingAuto := messenger.ExportPhase151DecideThreadFollowTransition(&auto, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeAuto,
		Position: head,
	})
	assert.False(t, changed)
	assert.Equal(t, trigger, existingAuto.Cursor, "an existing eligible follower keeps its cursor")

	changed, manualUnfollow := messenger.ExportPhase151DecideThreadFollowTransition(&auto, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeManualUnfollow,
		Position: head,
	})
	require.True(t, changed)
	assert.Equal(t, messenger.ThreadFollowModeManualUnfollow, manualUnfollow.Mode)

	changed, suppressed := messenger.ExportPhase151DecideThreadFollowTransition(&manualUnfollow, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeAuto,
		Position: later,
	})
	assert.False(t, changed)
	assert.Equal(t, manualUnfollow, suppressed, "manual unfollow persists and suppresses later auto-follow")

	changed, manualFollow := messenger.ExportPhase151DecideThreadFollowTransition(&manualUnfollow, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeManualFollow,
		Position: head,
	})
	require.True(t, changed)
	assert.Equal(t, messenger.ThreadFollowModeManualFollow, manualFollow.Mode)
	assert.Equal(t, head, manualFollow.Cursor, "manual follow baselines the current authorized thread head")

	changed, preservedManual := messenger.ExportPhase151DecideThreadFollowTransition(&manualFollow, messenger.ThreadFollowTransition{
		Mode:     messenger.ThreadFollowModeAuto,
		Position: later,
	})
	assert.False(t, changed)
	assert.Equal(t, manualFollow, preservedManual, "auto never overwrites a manual follow")
}

func TestPhase151PostCreateEffectsClaimDecisionIsIdempotent(t *testing.T) {
	for _, tt := range []struct {
		name    string
		created bool
		state   messenger.PostCreateEffectsState
		claim   bool
	}{
		{name: "new root", created: true, state: messenger.PostCreateEffectsStatePending, claim: true},
		{name: "new direct reply", created: true, state: messenger.PostCreateEffectsStatePending, claim: true},
		{name: "legacy missing state is complete", created: true, state: messenger.PostCreateEffectsState(""), claim: false},
		{name: "created false retry", created: false, state: messenger.PostCreateEffectsStatePending, claim: false},
		{name: "already complete retry", created: true, state: messenger.PostCreateEffectsStateComplete, claim: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := messenger.CreateMessageResult{Created: tt.created, PostCreateEffectsState: tt.state}
			assert.Equal(t, tt.claim, messenger.ExportPhase151ShouldClaimPostCreateEffects(result))
		})
	}
}

func TestPhase151DurableMessageContractPersistsNoUnreadCounter(t *testing.T) {
	fields := messenger.ExportPhase151DurableMessageBSONFields()
	require.Contains(t, fields, "post_create_effects_state")
	require.Contains(t, fields, "post_create_effects_completed_at")
	require.Contains(t, fields, "also_send_to_channel")
	require.Contains(t, fields, "resolved_mention_user_ids")
	assert.NotContains(t, fields, "unread_count")
	assert.NotContains(t, fields, "unread_counter")
}