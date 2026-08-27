package messenger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Phase151PostCreateRecoveryMongoStore interface {
	ApplyPostCreateRecoveryClaim(context.Context, bson.D, bson.D) (*MessageDocument, bool, error)
	ApplyPostCreateRecoveryCompletion(context.Context, bson.D, bson.D) (bool, error)
}

type Phase151PostCreateRecoveryMongoRepository interface {
	ClaimPendingPostCreateRecovery(
		context.Context,
		AuthorizedMessageProducerScope,
		*MessageDocument,
		uuid.UUID,
		time.Time,
		time.Duration,
	) (Phase151PostCreateRecoveryAcquisition, bool, error)
	CompleteClaimedPostCreateRecovery(
		context.Context,
		AuthorizedMessageProducerScope,
		*MessageDocument,
		Phase151PostCreateRecoveryClaim,
		time.Time,
	) error
}

type Phase151PostCreateRecoveryAcquisition struct {
	Claim   Phase151PostCreateRecoveryClaim
	Message MessageDocument
}

type Phase151PostCreateRecoveryMongoAdapter struct {
	store Phase151PostCreateRecoveryMongoStore
}

var _ Phase151PostCreateRecoveryMongoRepository = (*Phase151PostCreateRecoveryMongoAdapter)(nil)
var _ Phase151PostCreateRecoveryMongoRepository = (*MongoRepository)(nil)

func NewPhase151PostCreateRecoveryMongoAdapter(
	store Phase151PostCreateRecoveryMongoStore,
) *Phase151PostCreateRecoveryMongoAdapter {
	if phase151PostCreateDependencyNil(store) {
		panic("post-create recovery Mongo store is required")
	}
	return &Phase151PostCreateRecoveryMongoAdapter{store: store}
}

func (a *Phase151PostCreateRecoveryMongoAdapter) ClaimPendingPostCreateRecovery(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	ownerID uuid.UUID,
	now time.Time,
	leaseDuration time.Duration,
) (Phase151PostCreateRecoveryAcquisition, bool, error) {
	if a == nil || phase151PostCreateDependencyNil(a.store) || ctx == nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, err
	}
	plan, err := ExportPhase151BuildPostCreateRecoveryClaimPlan(
		scope, message, ownerID, now, leaseDuration,
	)
	if err != nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, err
	}

	claimed, acquired, err := a.store.ApplyPostCreateRecoveryClaim(ctx, plan.Filter, plan.Update)
	if err != nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, fmt.Errorf("claim pending post-create recovery: %w", err)
	}
	if !acquired {
		if claimed != nil {
			return Phase151PostCreateRecoveryAcquisition{}, false, ErrPostCreatePipelineInvalid
		}
		return Phase151PostCreateRecoveryAcquisition{}, false, nil
	}
	if err := validatePhase151PostCreateRecoveryClaimResult(scope, message, claimed, plan); err != nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, err
	}

	claim := Phase151PostCreateRecoveryClaim{
		TenantID: scope.TenantID, MessageID: claimed.ID, ChannelID: claimed.ChannelID,
		OwnerID: plan.OwnerID, Fence: claimed.PostCreateRecoveryFence, LeaseUntil: plan.LeaseUntil,
	}
	ownedMessage := *claimed
	ownedMessage.ResolvedMentionUserIDs = append([]uuid.UUID(nil), claimed.ResolvedMentionUserIDs...)
	ownedProvenance := *claimed.PostCreateProducerProvenance
	ownedMessage.PostCreateProducerProvenance = &ownedProvenance
	ownedLeaseUntil := plan.LeaseUntil
	ownedMessage.PostCreateRecoveryLeaseUntil = &ownedLeaseUntil
	return Phase151PostCreateRecoveryAcquisition{Claim: claim, Message: ownedMessage}, true, nil
}

func (a *Phase151PostCreateRecoveryMongoAdapter) CompleteClaimedPostCreateRecovery(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	claim Phase151PostCreateRecoveryClaim,
	completedAt time.Time,
) error {
	if a == nil || phase151PostCreateDependencyNil(a.store) || ctx == nil {
		return ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	plan, err := ExportPhase151BuildPostCreateRecoveryCompletionPlan(scope, message, claim, completedAt)
	if err != nil {
		return err
	}
	matched, err := a.store.ApplyPostCreateRecoveryCompletion(ctx, plan.Filter, plan.Update)
	if err != nil {
		return fmt.Errorf("complete claimed post-create recovery: %w", err)
	}
	if !matched {
		return ErrPostCreateRecoveryLeaseLost
	}
	return nil
}

func validatePhase151PostCreateRecoveryClaimResult(
	scope AuthorizedMessageProducerScope,
	expected *MessageDocument,
	claimed *MessageDocument,
	plan Phase151PostCreateRecoveryClaimPlan,
) error {
	if expected == nil || claimed == nil {
		return ErrPostCreatePipelineInvalid
	}
	if err := validatePhase151PendingRecoveryMessage(scope, claimed); err != nil {
		return err
	}
	if claimed.ID != expected.ID || !claimed.CreatedAt.Equal(expected.CreatedAt) ||
		claimed.PostCreateRecoveryOwnerID != plan.OwnerID || claimed.PostCreateRecoveryFence <= 0 ||
		claimed.PostCreateRecoveryLeaseUntil == nil ||
		!claimed.PostCreateRecoveryLeaseUntil.UTC().Equal(plan.LeaseUntil) {
		return ErrPostCreatePipelineInvalid
	}
	return nil
}

type phase151MongoPostCreateRecoveryStore struct {
	collection *mongo.Collection
}

func (s *phase151MongoPostCreateRecoveryStore) ApplyPostCreateRecoveryClaim(
	ctx context.Context,
	filter bson.D,
	update bson.D,
) (*MessageDocument, bool, error) {
	if s == nil || s.collection == nil || ctx == nil {
		return nil, false, ErrPostCreatePipelineInvalid
	}
	var claimed MessageDocument
	err := s.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&claimed)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &claimed, true, nil
}

func (s *phase151MongoPostCreateRecoveryStore) ApplyPostCreateRecoveryCompletion(
	ctx context.Context,
	filter bson.D,
	update bson.D,
) (bool, error) {
	if s == nil || s.collection == nil || ctx == nil {
		return false, ErrPostCreatePipelineInvalid
	}
	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return result.MatchedCount == 1, nil
}

func (r *MongoRepository) ClaimPendingPostCreateRecovery(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	ownerID uuid.UUID,
	now time.Time,
	leaseDuration time.Duration,
) (Phase151PostCreateRecoveryAcquisition, bool, error) {
	if r == nil || r.db == nil {
		return Phase151PostCreateRecoveryAcquisition{}, false, ErrPostCreatePipelineInvalid
	}
	adapter := NewPhase151PostCreateRecoveryMongoAdapter(&phase151MongoPostCreateRecoveryStore{
		collection: r.messages(),
	})
	return adapter.ClaimPendingPostCreateRecovery(ctx, scope, message, ownerID, now, leaseDuration)
}

func (r *MongoRepository) CompleteClaimedPostCreateRecovery(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
	claim Phase151PostCreateRecoveryClaim,
	completedAt time.Time,
) error {
	if r == nil || r.db == nil {
		return ErrPostCreatePipelineInvalid
	}
	adapter := NewPhase151PostCreateRecoveryMongoAdapter(&phase151MongoPostCreateRecoveryStore{
		collection: r.messages(),
	})
	return adapter.CompleteClaimedPostCreateRecovery(ctx, scope, message, claim, completedAt)
}
