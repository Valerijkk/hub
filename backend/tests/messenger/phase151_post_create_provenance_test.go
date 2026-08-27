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

func TestPhase151PreparedPendingMessagePersistsTrustedProducerProvenance(t *testing.T) {
	for _, scenario := range []struct {
		name  string
		scope messenger.AuthorizedMessageProducerScope
	}{
		{
			name: "user",
			scope: messenger.AuthorizedMessageProducerScope{
				TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
				Producer: messenger.TrustedMessageProducerUser,
			},
		},
		{
			name: "system bot",
			scope: messenger.AuthorizedMessageProducerScope{
				TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
				Producer: messenger.TrustedMessageProducerSystemBot,
				BotKind:  messenger.BotKindSystem, IntegrationID: uuid.New(),
			},
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			request := messenger.PostCreatePipelineRequest{
				Producer: scenario.scope.Producer,
				Message: messenger.MessageDocument{
					ChannelID: scenario.scope.ChannelID, SenderID: scenario.scope.ActorID,
					ClientNonce: "phase151-provenance-" + scenario.name,
				},
			}

			document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(
				scenario.scope, request, nil, time.Now().UTC(),
			)

			require.NoError(t, err)
			require.NotNil(t, document.PostCreateProducerProvenance)
			assert.Equal(t, messenger.PostCreateProducerProvenanceVersionV1, document.PostCreateProducerProvenance.Version)
			assert.Equal(t, scenario.scope.Producer, document.PostCreateProducerProvenance.Producer)
			assert.Equal(t, scenario.scope.BotKind, document.PostCreateProducerProvenance.BotKind)
			assert.Equal(t, scenario.scope.IntegrationID, document.PostCreateProducerProvenance.IntegrationID)
		})
	}
}

func TestPhase151PreparedPendingMessageRejectsCallerProducerProvenance(t *testing.T) {
	for _, provenance := range []*messenger.PostCreateProducerProvenance{
		{},
		{
			Version:  messenger.PostCreateProducerProvenanceVersionV1,
			Producer: messenger.TrustedMessageProducerUser,
		},
		{
			Version:  messenger.PostCreateProducerProvenanceVersionV1,
			Producer: messenger.TrustedMessageProducerSystemBot,
			BotKind:  messenger.BotKindSystem, IntegrationID: uuid.New(),
		},
	} {
		scope, request := phase151PreparedMessageFixture()
		request.Message.PostCreateProducerProvenance = provenance

		document, err := messenger.ExportPhase151PreparePendingCanonicalMessage(
			scope, request, nil, time.Now().UTC(),
		)

		assert.Nil(t, document)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	}
}

func TestPhase151ReconstructsExactRecoveryProducerScope(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		producer      messenger.TrustedMessageProducer
		botKind       messenger.BotKind
		integrationID uuid.UUID
	}{
		{name: "user", producer: messenger.TrustedMessageProducerUser},
		{name: "forward", producer: messenger.TrustedMessageProducerForward},
		{name: "poll", producer: messenger.TrustedMessageProducerPoll},
		{name: "promotion backlink", producer: messenger.TrustedMessageProducerPromotionBacklink},
		{
			name: "system bot", producer: messenger.TrustedMessageProducerSystemBot,
			botKind: messenger.BotKindSystem, integrationID: uuid.New(),
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			targetAuthority := messenger.AuthorizedMessageProducerScope{
				TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
				Producer: messenger.TrustedMessageProducerUser,
			}
			senderID := uuid.New()
			message := &messenger.MessageDocument{
				ID:        "phase151-recovery-provenance-" + scenario.name,
				ChannelID: targetAuthority.ChannelID, SenderID: senderID,
				PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
				CreatedAt:              time.Now().UTC(),
				PostCreateProducerProvenance: &messenger.PostCreateProducerProvenance{
					Version:  messenger.PostCreateProducerProvenanceVersionV1,
					Producer: scenario.producer, BotKind: scenario.botKind,
					IntegrationID: scenario.integrationID,
				},
			}

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScope(
				targetAuthority, message,
			)

			require.NoError(t, err)
			assert.Equal(t, targetAuthority.TenantID, scope.TenantID)
			assert.Equal(t, targetAuthority.ChannelID, scope.ChannelID)
			assert.Equal(t, senderID, scope.ActorID)
			assert.Equal(t, scenario.producer, scope.Producer)
			assert.Equal(t, scenario.botKind, scope.BotKind)
			assert.Equal(t, scenario.integrationID, scope.IntegrationID)
		})
	}
}

func TestPhase151RecoveryProducerScopeFailsClosedOnMalformedDurableProvenance(t *testing.T) {
	validTarget := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	validMessage := func() *messenger.MessageDocument {
		return &messenger.MessageDocument{
			ID: "phase151-recovery-provenance", ChannelID: validTarget.ChannelID,
			SenderID: uuid.New(), PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
			CreatedAt: time.Now().UTC(),
			PostCreateProducerProvenance: &messenger.PostCreateProducerProvenance{
				Version:  messenger.PostCreateProducerProvenanceVersionV1,
				Producer: messenger.TrustedMessageProducerUser,
			},
		}
	}

	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.AuthorizedMessageProducerScope, *messenger.MessageDocument)
	}{
		{name: "invalid target authority", mutate: func(target *messenger.AuthorizedMessageProducerScope, _ *messenger.MessageDocument) {
			target.ActorID = uuid.Nil
		}},
		{name: "missing message id", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.ID = ""
		}},
		{name: "foreign channel", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.ChannelID = uuid.New()
		}},
		{name: "missing sender", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.SenderID = uuid.Nil
		}},
		{name: "deleted message", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			now := time.Now().UTC()
			message.DeletedAt = &now
		}},
		{name: "complete message", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
		}},
		{name: "missing provenance", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance = nil
		}},
		{name: "unknown version", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.Version++
		}},
		{name: "unknown producer", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducer("unknown")
		}},
		{name: "user with bot proof", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.BotKind = messenger.BotKindSystem
		}},
		{name: "bot without integration", mutate: func(_ *messenger.AuthorizedMessageProducerScope, message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducerSystemBot
			message.PostCreateProducerProvenance.BotKind = messenger.BotKindSystem
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			target := validTarget
			message := validMessage()
			scenario.mutate(&target, message)

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScope(target, message)

			require.Error(t, err)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{}, scope)
		})
	}
}

func TestPhase151CanonicalResultRejectsProducerProvenanceDriftBeforeEffects(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.MessageDocument)
	}{
		{name: "missing", mutate: func(message *messenger.MessageDocument) { message.PostCreateProducerProvenance = nil }},
		{name: "unknown version", mutate: func(message *messenger.MessageDocument) { message.PostCreateProducerProvenance.Version++ }},
		{name: "producer mismatch", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducerForward
		}},
		{name: "bot kind mismatch", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.BotKind = messenger.BotKindSystem
		}},
		{name: "integration mismatch", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.IntegrationID = uuid.New()
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, request := phase151PostCreateCanonicalResultFixture()
			canonical := phase151PostCreateCanonicalResult(
				t, scope, request, false, messenger.PostCreateEffectsStateComplete,
			)
			if canonical.Message.PostCreateProducerProvenance == nil {
				canonical.Message.PostCreateProducerProvenance = &messenger.PostCreateProducerProvenance{
					Version: messenger.PostCreateProducerProvenanceVersionV1, Producer: scope.Producer,
					BotKind: scope.BotKind, IntegrationID: scope.IntegrationID,
				}
			}
			scenario.mutate(canonical.Message)
			recorder := &phase151PostCreateResultRecorder{
				phase151PostCreateRecorder: newPhase151PostCreateRecorder(),
				findFound:                  true, findResult: canonical,
			}

			result, err := messenger.ExportPhase151ExecutePostCreatePipeline(
				context.Background(), scope, request, recorder,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.CreateMessageResult{}, result)
			assert.Equal(t, []string{"idempotency:" + request.Message.ClientNonce}, recorder.calls)
			phase151AssertPostCreateNoEffects(t, recorder.phase151PostCreateRecorder)
		})
	}
}

func TestPhase151DurableManifestIncludesProducerProvenance(t *testing.T) {
	assert.Contains(t, messenger.ExportPhase151DurableMessageBSONFields(), "post_create_producer_provenance")
}

func TestPhase151RecoveryClaimFencesAndOwnsProducerProvenance(t *testing.T) {
	scope, message, now := phase151RecoveryClaimFixture()
	message.PostCreateProducerProvenance = &messenger.PostCreateProducerProvenance{
		Version:  messenger.PostCreateProducerProvenanceVersionV1,
		Producer: messenger.TrustedMessageProducerUser,
	}
	ownerID := uuid.New()
	t.Run("nil message is invalid", func(t *testing.T) {
		_, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
			scope, nil, ownerID, now, time.Minute,
		)
		require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
	})
	plan, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
		scope, message, ownerID, now, time.Minute,
	)
	require.NoError(t, err)
	assert.Contains(t, plan.Filter, bson.E{
		Key: "post_create_producer_provenance", Value: message.PostCreateProducerProvenance,
	})

	claim := messenger.Phase151PostCreateRecoveryClaim{
		TenantID: scope.TenantID, MessageID: message.ID, ChannelID: scope.ChannelID,
		OwnerID: ownerID, Fence: 3, LeaseUntil: plan.LeaseUntil,
	}
	completion, err := messenger.ExportPhase151BuildPostCreateRecoveryCompletionPlan(
		scope, message, claim, now.Add(30*time.Second),
	)
	require.NoError(t, err)
	assert.Contains(t, completion.Filter, bson.E{
		Key: "post_create_producer_provenance", Value: message.PostCreateProducerProvenance,
	})

	t.Run("plan rejects provenance drift", func(t *testing.T) {
		drifted := *message
		drifted.PostCreateProducerProvenance = &messenger.PostCreateProducerProvenance{
			Version:  messenger.PostCreateProducerProvenanceVersionV1,
			Producer: messenger.TrustedMessageProducerForward,
		}
		_, err := messenger.ExportPhase151BuildPostCreateRecoveryClaimPlan(
			scope, &drifted, ownerID, now, time.Minute,
		)
		require.ErrorIs(t, err, messenger.ErrTrustedMessageProducerDenied)
	})

	t.Run("adapter rejects post-update provenance drift", func(t *testing.T) {
		claimed := phase151ClaimedPostCreateDocument(message, ownerID, 3, plan.LeaseUntil)
		claimed.PostCreateProducerProvenance = &messenger.PostCreateProducerProvenance{
			Version:  messenger.PostCreateProducerProvenanceVersionV1,
			Producer: messenger.TrustedMessageProducerForward,
		}
		store := &phase151PostCreateRecoveryMongoStore{claimDocument: claimed, claimAcquired: true}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		acquisition, acquired, err := adapter.ClaimPendingPostCreateRecovery(
			context.Background(), scope, message, ownerID, now, time.Minute,
		)

		require.ErrorIs(t, err, messenger.ErrTrustedMessageProducerDenied)
		assert.False(t, acquired)
		assert.Equal(t, messenger.Phase151PostCreateRecoveryAcquisition{}, acquisition)
	})

	t.Run("acquisition owns provenance copy", func(t *testing.T) {
		claimed := phase151ClaimedPostCreateDocument(message, ownerID, 3, plan.LeaseUntil)
		store := &phase151PostCreateRecoveryMongoStore{claimDocument: claimed, claimAcquired: true}
		adapter := messenger.NewPhase151PostCreateRecoveryMongoAdapter(store)

		acquisition, acquired, err := adapter.ClaimPendingPostCreateRecovery(
			context.Background(), scope, message, ownerID, now, time.Minute,
		)

		require.NoError(t, err)
		require.True(t, acquired)
		require.NotNil(t, acquisition.Message.PostCreateProducerProvenance)
		acquisition.Message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducerForward
		assert.Equal(t, messenger.TrustedMessageProducerUser, claimed.PostCreateProducerProvenance.Producer)
	})
}
