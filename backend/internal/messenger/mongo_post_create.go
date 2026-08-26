package messenger

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoRepository) FindCanonicalCreateResult(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	clientNonce string,
) (CreateMessageResult, bool, error) {
	if r == nil || r.db == nil || scope.ActorID == uuid.Nil || clientNonce == "" {
		return CreateMessageResult{}, false, ErrPostCreatePipelineInvalid
	}
	canonical, err := r.FindMessageByClientNonce(ctx, scope.ActorID, clientNonce)
	if err != nil {
		return CreateMessageResult{}, false, err
	}
	if canonical == nil {
		return CreateMessageResult{}, false, nil
	}
	if canonical.ChannelID != scope.ChannelID || canonical.SenderID != scope.ActorID {
		return CreateMessageResult{}, false, ErrTrustedMessageProducerDenied
	}
	return *newCreateMessageResult(canonical, false), true, nil
}

func (r *MongoRepository) PersistPendingCanonicalMessage(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	document *MessageDocument,
) (CreateMessageResult, error) {
	if r == nil || r.db == nil || document == nil || document.ClientNonce == "" ||
		document.ChannelID != scope.ChannelID || document.SenderID != scope.ActorID ||
		document.PostCreateEffectsState != PostCreateEffectsStatePending ||
		document.PostCreateEffectsCompletedAt != nil {
		return CreateMessageResult{}, ErrPostCreatePipelineInvalid
	}
	result, err := r.CreateMessageIdempotent(ctx, document)
	if err != nil {
		return CreateMessageResult{}, err
	}
	if result == nil || result.Message == nil {
		return CreateMessageResult{}, errors.New("canonical pending message missing")
	}
	return *result, nil
}

// CompletePostCreateEffects is the durable pending-to-complete CAS. Binding
// channel_id prevents an authorized producer scope from claiming a document
// in another channel; missing legacy state is intentionally not claimable.
func (r *MongoRepository) CompletePostCreateEffects(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	messageID string,
) (bool, error) {
	if r == nil || r.db == nil || scope.ChannelID == uuid.Nil || messageID == "" {
		return false, ErrPostCreatePipelineInvalid
	}
	completedAt := time.Now().UTC()
	result, err := r.messages().UpdateOne(ctx, bson.D{
		{Key: "_id", Value: messageID},
		{Key: "channel_id", Value: scope.ChannelID},
		{Key: "post_create_effects_state", Value: PostCreateEffectsStatePending},
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "post_create_effects_state", Value: PostCreateEffectsStateComplete},
		{Key: "post_create_effects_completed_at", Value: completedAt},
	}}})
	if err != nil {
		return false, err
	}
	return result.MatchedCount == 1, nil
}

func (r *MongoRepository) ListPendingPostCreateDocuments(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	target Phase151PendingRecoveryTarget,
	limit int,
) ([]MessageDocument, error) {
	if r == nil || r.db == nil || limit <= 0 || scope.TenantID != target.TenantID ||
		scope.ChannelID != target.ChannelID || target.ChannelID == uuid.Nil {
		return nil, ErrPostCreatePipelineInvalid
	}
	if limit > 500 {
		limit = 500
	}
	cursor, err := r.messages().Find(ctx, bson.D{
		{Key: "channel_id", Value: target.ChannelID},
		{Key: "post_create_effects_state", Value: PostCreateEffectsStatePending},
	}, options.Find().SetSort(bson.D{
		{Key: "created_at", Value: 1},
		{Key: "_id", Value: 1},
	}).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	documents := make([]MessageDocument, 0, limit)
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	return documents, nil
}
