package messenger_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(t *testing.T) {
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
			authority := phase151RecoveryTargetAuthorityFixture()
			senderID := uuid.New()
			message := phase151RecoveryTargetCanonicalMessage(
				authority.ChannelID,
				senderID,
				scenario.producer,
				scenario.botKind,
				scenario.integrationID,
			)

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(
				authority,
				message,
			)

			require.NoError(t, err)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{
				TenantID:      authority.TenantID,
				ActorID:       senderID,
				ChannelID:     authority.ChannelID,
				Producer:      scenario.producer,
				BotKind:       scenario.botKind,
				IntegrationID: scenario.integrationID,
			}, scope)
		})
	}
}

func TestPhase151RecoveryTargetAuthorityReconstructionRejectsInvalidAuthority(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.Phase151RecoveryTargetAuthority)
	}{
		{name: "empty principal", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) {
			authority.Principal = ""
		}},
		{name: "arbitrary principal", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) {
			authority.Principal = messenger.Phase151RecoveryServicePrincipal("messenger.user_actor")
		}},
		{name: "nil tenant", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) {
			authority.TenantID = uuid.Nil
		}},
		{name: "nil channel", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) {
			authority.ChannelID = uuid.Nil
		}},
		{name: "empty kind", mutate: func(authority *messenger.Phase151RecoveryTargetAuthority) {
			authority.Kind = ""
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			authority := phase151RecoveryTargetAuthorityFixture()
			scenario.mutate(&authority)
			message := phase151RecoveryTargetCanonicalMessage(
				authority.ChannelID,
				uuid.New(),
				messenger.TrustedMessageProducerUser,
				"",
				uuid.Nil,
			)

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(
				authority,
				message,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{}, scope)
		})
	}
}

func TestPhase151RecoveryTargetAuthorityReconstructionRejectsMalformedCanonicalMessage(t *testing.T) {
	authority := phase151RecoveryTargetAuthorityFixture()
	for _, scenario := range []struct {
		name    string
		nilMessage bool
		mutate  func(*messenger.MessageDocument)
	}{
		{name: "nil message", nilMessage: true},
		{name: "empty id", mutate: func(message *messenger.MessageDocument) { message.ID = "" }},
		{name: "foreign channel", mutate: func(message *messenger.MessageDocument) { message.ChannelID = uuid.New() }},
		{name: "nil sender", mutate: func(message *messenger.MessageDocument) { message.SenderID = uuid.Nil }},
		{name: "zero created at", mutate: func(message *messenger.MessageDocument) { message.CreatedAt = time.Time{} }},
		{name: "deleted", mutate: func(message *messenger.MessageDocument) {
			now := time.Now().UTC()
			message.DeletedAt = &now
		}},
		{name: "complete state", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsState = messenger.PostCreateEffectsStateComplete
		}},
		{name: "unknown state", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateEffectsState = messenger.PostCreateEffectsState("unknown")
		}},
		{name: "completed at set", mutate: func(message *messenger.MessageDocument) {
			now := time.Now().UTC()
			message.PostCreateEffectsCompletedAt = &now
		}},
		{name: "nil provenance", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance = nil
		}},
		{name: "wrong provenance version", mutate: func(message *messenger.MessageDocument) {
			message.PostCreateProducerProvenance.Version++
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			message := phase151RecoveryTargetCanonicalMessage(
				authority.ChannelID,
				uuid.New(),
				messenger.TrustedMessageProducerUser,
				"",
				uuid.Nil,
			)
			if scenario.nilMessage {
				message = nil
			} else {
				scenario.mutate(message)
			}

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(
				authority,
				message,
			)

			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{}, scope)
		})
	}
}

func TestPhase151RecoveryTargetAuthorityReconstructionRejectsMalformedProvenance(t *testing.T) {
	authority := phase151RecoveryTargetAuthorityFixture()
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.PostCreateProducerProvenance)
	}{
		{name: "unknown producer", mutate: func(provenance *messenger.PostCreateProducerProvenance) {
			provenance.Producer = messenger.TrustedMessageProducer("unknown")
		}},
		{name: "non bot with bot kind", mutate: func(provenance *messenger.PostCreateProducerProvenance) {
			provenance.BotKind = messenger.BotKindSystem
		}},
		{name: "non bot with integration", mutate: func(provenance *messenger.PostCreateProducerProvenance) {
			provenance.IntegrationID = uuid.New()
		}},
		{name: "system bot without bot kind", mutate: func(provenance *messenger.PostCreateProducerProvenance) {
			provenance.Producer = messenger.TrustedMessageProducerSystemBot
		}},
		{name: "system bot without integration", mutate: func(provenance *messenger.PostCreateProducerProvenance) {
			provenance.Producer = messenger.TrustedMessageProducerSystemBot
			provenance.BotKind = messenger.BotKindSystem
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			message := phase151RecoveryTargetCanonicalMessage(
				authority.ChannelID,
				uuid.New(),
				messenger.TrustedMessageProducerUser,
				"",
				uuid.Nil,
			)
			scenario.mutate(message.PostCreateProducerProvenance)

			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(
				authority,
				message,
			)

			require.ErrorIs(t, err, messenger.ErrTrustedMessageProducerDenied)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{}, scope)
		})
	}
}

func TestPhase151LegacyRecoveryProducerScopeReconstructionErrorSentinelsRemainStable(t *testing.T) {
	validTarget := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	validMessage := func() *messenger.MessageDocument {
		return phase151RecoveryTargetCanonicalMessage(
			validTarget.ChannelID,
			uuid.New(),
			messenger.TrustedMessageProducerUser,
			"",
			uuid.Nil,
		)
	}

	for _, scenario := range []struct {
		name   string
		target messenger.AuthorizedMessageProducerScope
		message func() *messenger.MessageDocument
		want   error
	}{
		{
			name: "invalid actor shaped target is denied",
			target: messenger.AuthorizedMessageProducerScope{
				TenantID: validTarget.TenantID, ChannelID: validTarget.ChannelID,
				Producer: messenger.TrustedMessageProducerUser,
			},
			message: validMessage,
			want:    messenger.ErrTrustedMessageProducerDenied,
		},
		{
			name:   "malformed canonical message is invalid",
			target: validTarget,
			message: func() *messenger.MessageDocument {
				message := validMessage()
				message.ID = ""
				return message
			},
			want: messenger.ErrPostCreatePipelineInvalid,
		},
		{
			name:   "malformed provenance is denied",
			target: validTarget,
			message: func() *messenger.MessageDocument {
				message := validMessage()
				message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducer("unknown")
				return message
			},
			want: messenger.ErrTrustedMessageProducerDenied,
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			scope, err := messenger.ExportPhase151ReconstructPostCreateRecoveryProducerScope(
				scenario.target,
				scenario.message(),
			)

			require.ErrorIs(t, err, scenario.want)
			assert.Equal(t, messenger.AuthorizedMessageProducerScope{}, scope)
		})
	}
}

func phase151RecoveryTargetAuthorityFixture() messenger.Phase151RecoveryTargetAuthority {
	return messenger.Phase151RecoveryTargetAuthority{
		Principal: messenger.Phase151PostCreateRecoveryServicePrincipal,
		TenantID:  uuid.New(),
		ChannelID: uuid.New(),
		Kind:      "channel_cleanup",
	}
}

func phase151RecoveryTargetCanonicalMessage(
	channelID uuid.UUID,
	senderID uuid.UUID,
	producer messenger.TrustedMessageProducer,
	botKind messenger.BotKind,
	integrationID uuid.UUID,
) *messenger.MessageDocument {
	return &messenger.MessageDocument{
		ID:        "phase151-target-authority-reconstruction",
		ChannelID: channelID,
		SenderID:  senderID,
		CreatedAt: time.Now().UTC(),
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
		PostCreateProducerProvenance: &messenger.PostCreateProducerProvenance{
			Version:       messenger.PostCreateProducerProvenanceVersionV1,
			Producer:      producer,
			BotKind:       botKind,
			IntegrationID: integrationID,
		},
	}
}
