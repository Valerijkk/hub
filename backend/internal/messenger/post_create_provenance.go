package messenger

import "github.com/google/uuid"

const PostCreateProducerProvenanceVersionV1 = 1

type PostCreateProducerProvenance struct {
	Version       int                    `bson:"version"`
	Producer      TrustedMessageProducer `bson:"producer"`
	BotKind       BotKind                `bson:"bot_kind,omitempty"`
	IntegrationID uuid.UUID              `bson:"integration_id,omitempty"`
}

func newPostCreateProducerProvenance(
	scope AuthorizedMessageProducerScope,
) *PostCreateProducerProvenance {
	return &PostCreateProducerProvenance{
		Version:       PostCreateProducerProvenanceVersionV1,
		Producer:      scope.Producer,
		BotKind:       scope.BotKind,
		IntegrationID: scope.IntegrationID,
	}
}

func postCreateProducerProvenanceMatchesScope(
	provenance *PostCreateProducerProvenance,
	scope AuthorizedMessageProducerScope,
) bool {
	if provenance == nil || provenance.Version != PostCreateProducerProvenanceVersionV1 ||
		provenance.Producer != scope.Producer || provenance.BotKind != scope.BotKind ||
		provenance.IntegrationID != scope.IntegrationID {
		return false
	}
	return ExportPhase151ValidateTrustedMessageProducer(scope, provenance.Producer) == nil
}

// ExportPhase151ReconstructPostCreateRecoveryProducerScope combines a fresh
// tenant/channel authorization with immutable producer provenance stored on the
// pending canonical message. The reauthorizing actor is never treated as the
// original sender.
func ExportPhase151ReconstructPostCreateRecoveryProducerScope(
	targetAuthority AuthorizedMessageProducerScope,
	message *MessageDocument,
) (AuthorizedMessageProducerScope, error) {
	if err := ExportPhase151ValidateTrustedMessageProducer(targetAuthority, targetAuthority.Producer); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	return reconstructPostCreateRecoveryProducerScope(
		targetAuthority.TenantID,
		targetAuthority.ChannelID,
		message,
	)
}

// ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority
// reconstructs the canonical producer scope from validated server-owned target
// facts and immutable message provenance. It does not authorize a recovery
// target or treat the recovery principal as the original message sender.
func ExportPhase151ReconstructPostCreateRecoveryProducerScopeFromTargetAuthority(
	authority Phase151RecoveryTargetAuthority,
	message *MessageDocument,
) (AuthorizedMessageProducerScope, error) {
	target := Phase151PendingRecoveryTarget{
		TenantID: authority.TenantID, ChannelID: authority.ChannelID, Kind: authority.Kind,
	}
	if err := ExportPhase151ValidateRecoveryTargetAuthority(authority.Principal, target, authority); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	return reconstructPostCreateRecoveryProducerScope(authority.TenantID, authority.ChannelID, message)
}

func reconstructPostCreateRecoveryProducerScope(
	tenantID uuid.UUID,
	channelID uuid.UUID,
	message *MessageDocument,
) (AuthorizedMessageProducerScope, error) {
	if message == nil || message.ID == "" || message.ChannelID != channelID ||
		message.SenderID == uuid.Nil || message.CreatedAt.IsZero() || message.DeletedAt != nil ||
		message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil || message.PostCreateProducerProvenance == nil ||
		message.PostCreateProducerProvenance.Version != PostCreateProducerProvenanceVersionV1 {
		return AuthorizedMessageProducerScope{}, ErrPostCreatePipelineInvalid
	}

	provenance := message.PostCreateProducerProvenance
	scope := AuthorizedMessageProducerScope{
		TenantID:      tenantID,
		ActorID:       message.SenderID,
		ChannelID:     channelID,
		Producer:      provenance.Producer,
		BotKind:       provenance.BotKind,
		IntegrationID: provenance.IntegrationID,
	}
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	return scope, nil
}
