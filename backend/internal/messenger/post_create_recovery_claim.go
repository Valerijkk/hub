package messenger

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrPostCreateRecoveryLeaseLost = errors.New("post-create recovery lease lost")

type Phase151PostCreateRecoveryClaim struct {
	TenantID   uuid.UUID
	MessageID  string
	ChannelID  uuid.UUID
	OwnerID    uuid.UUID
	Fence      int64
	LeaseUntil time.Time
}

type Phase151PostCreateRecoveryClaimPlan struct {
	Filter     bson.D
	Update     bson.D
	OwnerID    uuid.UUID
	LeaseUntil time.Time
}

type Phase151PostCreateRecoveryCompletionPlan struct {
	Filter bson.D
	Update bson.D
}

func ExportPhase151BuildPostCreateRecoveryClaimPlan(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	ownerID uuid.UUID,
	now time.Time,
	leaseDuration time.Duration,
) (Phase151PostCreateRecoveryClaimPlan, error) {
	if err := validatePhase151PendingRecoveryMessage(scope, message); err != nil {
		return Phase151PostCreateRecoveryClaimPlan{}, err
	}
	if ownerID == uuid.Nil || now.IsZero() || leaseDuration <= 0 {
		return Phase151PostCreateRecoveryClaimPlan{}, ErrPostCreatePipelineInvalid
	}
	now = now.UTC()
	if now.Before(message.CreatedAt.UTC()) {
		return Phase151PostCreateRecoveryClaimPlan{}, ErrPostCreatePipelineInvalid
	}
	leaseUntil := now.Add(leaseDuration)
	if !leaseUntil.After(now) {
		return Phase151PostCreateRecoveryClaimPlan{}, ErrPostCreatePipelineInvalid
	}

	return Phase151PostCreateRecoveryClaimPlan{
		Filter: bson.D{
			{Key: "_id", Value: message.ID},
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "sender_id", Value: scope.ActorID},
			{Key: "post_create_producer_provenance", Value: message.PostCreateProducerProvenance},
			{Key: "deleted_at", Value: nil},
			{Key: "post_create_effects_state", Value: PostCreateEffectsStatePending},
			{Key: "post_create_effects_completed_at", Value: nil},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "post_create_recovery_lease_until", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "post_create_recovery_lease_until", Value: bson.D{{Key: "$lte", Value: now}}}},
			}},
		},
		Update: bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "post_create_recovery_owner_id", Value: ownerID},
				{Key: "post_create_recovery_lease_until", Value: leaseUntil},
			}},
			{Key: "$inc", Value: bson.D{{Key: "post_create_recovery_fence", Value: int64(1)}}},
		},
		OwnerID: ownerID, LeaseUntil: leaseUntil,
	}, nil
}

func ExportPhase151BuildPostCreateRecoveryCompletionPlan(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	claim Phase151PostCreateRecoveryClaim,
	completedAt time.Time,
) (Phase151PostCreateRecoveryCompletionPlan, error) {
	if err := validatePhase151PendingRecoveryMessage(scope, message); err != nil {
		return Phase151PostCreateRecoveryCompletionPlan{}, err
	}
	if claim.TenantID != scope.TenantID || claim.MessageID != message.ID ||
		claim.ChannelID != scope.ChannelID || claim.OwnerID == uuid.Nil || claim.Fence <= 0 ||
		claim.LeaseUntil.IsZero() || completedAt.IsZero() {
		return Phase151PostCreateRecoveryCompletionPlan{}, ErrPostCreatePipelineInvalid
	}
	leaseUntil := claim.LeaseUntil.UTC()
	completedAt = completedAt.UTC()
	if completedAt.Before(message.CreatedAt.UTC()) {
		return Phase151PostCreateRecoveryCompletionPlan{}, ErrPostCreatePipelineInvalid
	}
	if !completedAt.Before(leaseUntil) {
		return Phase151PostCreateRecoveryCompletionPlan{}, ErrPostCreateRecoveryLeaseLost
	}

	return Phase151PostCreateRecoveryCompletionPlan{
		Filter: bson.D{
			{Key: "_id", Value: message.ID},
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "sender_id", Value: scope.ActorID},
			{Key: "post_create_producer_provenance", Value: message.PostCreateProducerProvenance},
			{Key: "deleted_at", Value: nil},
			{Key: "post_create_effects_state", Value: PostCreateEffectsStatePending},
			{Key: "post_create_effects_completed_at", Value: nil},
			{Key: "post_create_recovery_owner_id", Value: claim.OwnerID},
			{Key: "post_create_recovery_fence", Value: claim.Fence},
			{Key: "post_create_recovery_lease_until", Value: bson.D{
				{Key: "$eq", Value: leaseUntil},
				{Key: "$gt", Value: completedAt},
			}},
		},
		Update: bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "post_create_effects_state", Value: PostCreateEffectsStateComplete},
				{Key: "post_create_effects_completed_at", Value: completedAt},
			}},
			{Key: "$unset", Value: bson.D{
				{Key: "post_create_recovery_owner_id", Value: ""},
				{Key: "post_create_recovery_lease_until", Value: ""},
			}},
		},
	}, nil
}

func validatePhase151PendingRecoveryMessage(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) error {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return err
	}
	if message == nil || message.ID == "" || message.ChannelID != scope.ChannelID ||
		message.SenderID != scope.ActorID || message.CreatedAt.IsZero() || message.DeletedAt != nil ||
		message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil {
		return ErrPostCreatePipelineInvalid
	}
	if !postCreateProducerProvenanceMatchesScope(message.PostCreateProducerProvenance, scope) {
		return ErrTrustedMessageProducerDenied
	}
	return nil
}
