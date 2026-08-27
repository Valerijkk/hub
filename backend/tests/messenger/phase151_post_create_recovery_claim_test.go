package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func phase151RecoveryClaimFixture() (messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument, time.Time) {
	now := time.Date(2026, time.August, 27, 18, 0, 0, 0, time.UTC)
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	message := &messenger.MessageDocument{
		ID: "phase151-recovery-message", ChannelID: scope.ChannelID, SenderID: scope.ActorID,
		ClientNonce: "phase151-recovery-nonce", CreatedAt: now.Add(-time.Minute),
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
	}
	return scope, message, now
}

func TestPhase151PostCreateRecoveryClaimPlanOwnsOnlyExpiredOrUnclaimedPendingMessage(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	ownerID := uuid.New()
	leaseUntil := now.Add(2 * time.Minute)
	claimNow := now.In(time.FixedZone("phase151-claim", 3*60*60))

	plan, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
		scope, message, ownerID, claimNow, 2*time.Minute,
	)

	require.NoError(t, err)
	assert.Equal(t, ownerID, plan.OwnerID)
	assert.Equal(t, leaseUntil, plan.LeaseUntil)
	assert.Equal(t, bson.D{
		{Key: "_id", Value: message.ID},
		{Key: "channel_id", Value: scope.ChannelID},
		{Key: "sender_id", Value: scope.ActorID},
		{Key: "deleted_at", Value: nil},
		{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStatePending},
		{Key: "post_create_effects_completed_at", Value: nil},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "post_create_recovery_lease_until", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "post_create_recovery_lease_until", Value: bson.D{{Key: "$lte", Value: now}}}},
		}},
	}, plan.Filter)
	assert.Equal(t, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "post_create_recovery_owner_id", Value: ownerID},
			{Key: "post_create_recovery_lease_until", Value: leaseUntil},
		}},
		{Key: "$inc", Value: bson.D{{Key: "post_create_recovery_fence", Value: int64(1)}}},
	}, plan.Update)
}

func TestPhase151PostCreateRecoveryCompletionPlanRequiresExactLiveFence(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	claim := messenger.Phase151PostCreateRecoveryClaim{
		TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID, OwnerID: uuid.New(),
		Fence: 7, LeaseUntil: now.Add(2 * time.Minute).In(time.FixedZone("phase151-lease", -4*60*60)),
	}
	completedAt := now.Add(time.Minute)

	plan, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(
		scope, message, claim, completedAt,
	)

	require.NoError(t, err)
	assert.Equal(t, bson.D{
		{Key: "_id", Value: message.ID},
		{Key: "channel_id", Value: scope.ChannelID},
		{Key: "sender_id", Value: scope.ActorID},
		{Key: "deleted_at", Value: nil},
		{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStatePending},
		{Key: "post_create_effects_completed_at", Value: nil},
		{Key: "post_create_recovery_owner_id", Value: claim.OwnerID},
		{Key: "post_create_recovery_fence", Value: claim.Fence},
		{Key: "post_create_recovery_lease_until", Value: bson.D{
			{Key: "$eq", Value: claim.LeaseUntil.UTC()},
			{Key: "$gt", Value: completedAt},
		}},
	}, plan.Filter)
	assert.Equal(t, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "post_create_effects_state", Value: messenger.PostCreateEffectsStateComplete},
			{Key: "post_create_effects_completed_at", Value: completedAt},
		}},
		{Key: "$unset", Value: bson.D{
			{Key: "post_create_recovery_owner_id", Value: ""},
			{Key: "post_create_recovery_lease_until", Value: ""},
		}},
	}, plan.Update)

	for _, expiredAt := range []time.Time{claim.LeaseUntil, claim.LeaseUntil.Add(time.Nanosecond)} {
		_, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(scope, message, claim, expiredAt)
		require.ErrorIs(t, err, messenger.ErrPostCreateRecoveryLeaseLost)
	}
}

func TestPhase151PostCreateRecoveryPlansRejectInvalidAuthorityStateAndFence(t *testing.T) {
	t.Run("claim rejects incomplete authority and canonical state", func(t *testing.T) {
		scope, message, now := phase151RecoveryClaimFixture()
		ownerID := uuid.New()
		cases := []struct {
			name   string
			mutate func(*messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument)
			owner  uuid.UUID
			now    time.Time
			lease  time.Duration
		}{
			{name: "nil owner", owner: uuid.Nil, now: now, lease: time.Minute},
			{name: "zero now", owner: ownerID, lease: time.Minute},
			{name: "zero lease", owner: ownerID, now: now},
			{name: "clock before message", owner: ownerID, now: message.CreatedAt.Add(-time.Nanosecond), lease: time.Minute},
			{name: "foreign channel", owner: ownerID, now: now, lease: time.Minute, mutate: func(_ *messenger.AuthorizedMessageProducerScope, m *messenger.MessageDocument) {
				m.ChannelID = uuid.New()
			}},
			{name: "foreign sender", owner: ownerID, now: now, lease: time.Minute, mutate: func(_ *messenger.AuthorizedMessageProducerScope, m *messenger.MessageDocument) {
				m.SenderID = uuid.New()
			}},
			{name: "deleted", owner: ownerID, now: now, lease: time.Minute, mutate: func(_ *messenger.AuthorizedMessageProducerScope, m *messenger.MessageDocument) {
				deletedAt := now
				m.DeletedAt = &deletedAt
			}},
			{name: "complete", owner: ownerID, now: now, lease: time.Minute, mutate: func(_ *messenger.AuthorizedMessageProducerScope, m *messenger.MessageDocument) {
				m.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
			}},
			{name: "completed timestamp", owner: ownerID, now: now, lease: time.Minute, mutate: func(_ *messenger.AuthorizedMessageProducerScope, m *messenger.MessageDocument) {
				completedAt := now
				m.PostCreateEffectsCompletedAt = &completedAt
			}},
			{name: "untrusted producer", owner: ownerID, now: now, lease: time.Minute, mutate: func(s *messenger.AuthorizedMessageProducerScope, _ *messenger.MessageDocument) {
				s.Producer = messenger.TrustedMessageProducer("unknown")
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				caseScope := scope
				caseMessage := *message
				if tc.mutate != nil {
					tc.mutate(&caseScope, &caseMessage)
				}
				_, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
					caseScope, &caseMessage, tc.owner, tc.now, tc.lease,
				)
				require.Error(t, err)
			})
		}
	})

	t.Run("completion rejects mismatched token", func(t *testing.T) {
		scope, message, now := phase151RecoveryClaimFixture()
		valid := messenger.Phase151PostCreateRecoveryClaim{
			TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID, OwnerID: uuid.New(), Fence: 1,
			LeaseUntil: now.Add(time.Minute),
		}
		claims := []messenger.Phase151PostCreateRecoveryClaim{
			{},
			{TenantID: uuid.New(), MessageID: valid.MessageID, ChannelID: valid.ChannelID, OwnerID: valid.OwnerID, Fence: valid.Fence, LeaseUntil: valid.LeaseUntil},
			{TenantID: valid.TenantID, MessageID: "foreign", ChannelID: valid.ChannelID, OwnerID: valid.OwnerID, Fence: valid.Fence, LeaseUntil: valid.LeaseUntil},
			{TenantID: valid.TenantID, MessageID: valid.MessageID, ChannelID: uuid.New(), OwnerID: valid.OwnerID, Fence: valid.Fence, LeaseUntil: valid.LeaseUntil},
			{TenantID: valid.TenantID, MessageID: valid.MessageID, ChannelID: valid.ChannelID, OwnerID: uuid.Nil, Fence: valid.Fence, LeaseUntil: valid.LeaseUntil},
			{TenantID: valid.TenantID, MessageID: valid.MessageID, ChannelID: valid.ChannelID, OwnerID: valid.OwnerID, Fence: 0, LeaseUntil: valid.LeaseUntil},
		}
		for _, claim := range claims {
			_, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(scope, message, claim, now)
			require.Error(t, err)
		}
		_, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(
			scope, message, valid, message.CreatedAt.Add(-time.Nanosecond),
		)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	})
}

func TestPhase151PostCreateRecoveryFieldsAreServerOwnedAndDurable(t *testing.T) {
	fields := messenger.ExportPhase151DurableMessageBSONFields()
	assert.Contains(t, fields, "post_create_recovery_owner_id")
	assert.Contains(t, fields, "post_create_recovery_fence")
	assert.Contains(t, fields, "post_create_recovery_lease_until")

	now := time.Date(2026, time.August, 27, 18, 0, 0, 0, time.UTC)
	mutations := map[string]func(*messenger.MessageDocument){
		"owner": func(message *messenger.MessageDocument) { message.PostCreateRecoveryOwnerID = uuid.New() },
		"fence": func(message *messenger.MessageDocument) { message.PostCreateRecoveryFence = 1 },
		"lease": func(message *messenger.MessageDocument) {
			leaseUntil := now.Add(time.Minute)
			message.PostCreateRecoveryLeaseUntil = &leaseUntil
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			scope, _, _ := phase151RecoveryClaimFixture()
			request := messenger.PostCreatePipelineRequest{
				Producer: messenger.TrustedMessageProducerUser,
				Message: messenger.MessageDocument{
					ChannelID: scope.ChannelID, SenderID: scope.ActorID,
					ContentText: "server-owned recovery metadata", ClientNonce: "phase151-recovery-" + name,
				},
			}
			mutate(&request.Message)
			recorder := newPhase151PostCreateRecorder()

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Empty(t, recorder.calls, "server-owned recovery fields fail before repository access")
		})
	}
}
