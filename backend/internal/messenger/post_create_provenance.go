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
	if message == nil || message.ID == "" || message.ChannelID != targetAuthority.ChannelID ||
		message.SenderID == uuid.Nil || message.CreatedAt.IsZero() || message.DeletedAt != nil ||
		message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil || message.PostCreateProducerProvenance == nil ||
		message.PostCreateProducerProvenance.Version != PostCreateProducerProvenanceVersionV1 {
		return AuthorizedMessageProducerScope{}, ErrPostCreatePipelineInvalid
	}

	provenance := message.PostCreateProducerProvenance
	scope := AuthorizedMessageProducerScope{
		TenantID:      targetAuthority.TenantID,
		ActorID:       message.SenderID,
		ChannelID:     targetAuthority.ChannelID,
		Producer:      provenance.Producer,
		BotKind:       provenance.BotKind,
		IntegrationID: provenance.IntegrationID,
	}
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return AuthorizedMessageProducerScope{}, err
	}
	return scope, nil
}
