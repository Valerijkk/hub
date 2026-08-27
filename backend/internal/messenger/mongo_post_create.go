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
		document.PostCreateEffectsCompletedAt != nil ||
		!postCreateProducerProvenanceMatchesScope(document.PostCreateProducerProvenance, scope) {
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

type Phase151ReplyCountReconciliationPlan struct {
	ReplyFilter bson.D
	RootFilter  bson.D
}

type Phase151PostCreateCompletionPlan struct {
	Filter bson.D
	Update bson.D
}

func ExportPhase151BuildPostCreateCompletionPlan(
	scope AuthorizedMessageProducerScope,
	messageID string,
	completedAt time.Time,
) (Phase151PostCreateCompletionPlan, error) {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return Phase151PostCreateCompletionPlan{}, err
	}
	if messageID == "" || completedAt.IsZero() {
		return Phase151PostCreateCompletionPlan{}, ErrPostCreatePipelineInvalid
	}
	completedAt = completedAt.UTC()

	return Phase151PostCreateCompletionPlan{
		Filter: bson.D{
			{Key: "_id", Value: messageID},
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "sender_id", Value: scope.ActorID},
			{Key: "post_create_producer_provenance", Value: newPostCreateProducerProvenance(scope)},
			{Key: "deleted_at", Value: nil},
			{Key: "post_create_effects_state", Value: PostCreateEffectsStatePending},
			{Key: "post_create_effects_completed_at", Value: nil},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "post_create_recovery_fence", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "post_create_recovery_fence", Value: int64(0)}},
			}},
		},
		Update: bson.D{{Key: "$set", Value: bson.D{
			{Key: "post_create_effects_state", Value: PostCreateEffectsStateComplete},
			{Key: "post_create_effects_completed_at", Value: completedAt},
		}}},
	}, nil
}

func ExportPhase151BuildReplyCountReconciliationPlan(
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) (Phase151ReplyCountReconciliationPlan, error) {
	if err := ExportPhase151ValidateTrustedMessageProducer(scope, scope.Producer); err != nil {
		return Phase151ReplyCountReconciliationPlan{}, err
	}
	if message == nil || message.ID == "" || message.ThreadID == "" || message.ThreadID == message.ID ||
		message.ChannelID != scope.ChannelID || message.SenderID != scope.ActorID ||
		message.DeletedAt != nil || message.PostCreateEffectsState != PostCreateEffectsStatePending ||
		message.PostCreateEffectsCompletedAt != nil {
		return Phase151ReplyCountReconciliationPlan{}, ErrPostCreatePipelineInvalid
	}

	return Phase151ReplyCountReconciliationPlan{
		ReplyFilter: bson.D{
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "thread_id", Value: message.ThreadID},
			{Key: "deleted_at", Value: nil},
		},
		RootFilter: bson.D{
			{Key: "_id", Value: message.ThreadID},
			{Key: "channel_id", Value: scope.ChannelID},
			{Key: "deleted_at", Value: nil},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "thread_id", Value: bson.D{{Key: "$exists", Value: false}}}},
				bson.D{{Key: "thread_id", Value: ""}},
			}},
		},
	}, nil
}

func ExportPhase151BuildReplyCountRaiseUpdate(replyCount int64) (bson.D, error) {
	if replyCount <= 0 {
		return nil, ErrPostCreatePipelineInvalid
	}
	return bson.D{{Key: "$max", Value: bson.D{{Key: "reply_count", Value: replyCount}}}}, nil
}

// ReconcileReplyCount raises the live root counter to the observed count of
// live direct replies. $max makes a retry a no-op and prevents an older
// concurrent observation from overwriting a newer, larger count.
func (r *MongoRepository) ReconcileReplyCount(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	message *MessageDocument,
) error {
	if r == nil || r.db == nil {
		return ErrPostCreatePipelineInvalid
	}
	plan, err := ExportPhase151BuildReplyCountReconciliationPlan(scope, message)
	if err != nil {
		return err
	}
	replyCount, err := r.messages().CountDocuments(ctx, plan.ReplyFilter)
	if err != nil {
		return err
	}
	update, err := ExportPhase151BuildReplyCountRaiseUpdate(replyCount)
	if err != nil {
		return err
	}
	result, err := r.messages().UpdateOne(ctx, plan.RootFilter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount != 1 {
		return errors.New("live thread root missing during reply-count reconciliation")
	}
	return nil
}

// CompletePostCreateEffects is the durable pending-to-complete CAS. Binding
// channel_id prevents an authorized producer scope from claiming a document
// in another channel; missing legacy state is intentionally not claimable.
func (r *MongoRepository) CompletePostCreateEffects(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	messageID string,
) (bool, error) {
	if r == nil || r.db == nil {
		return false, ErrPostCreatePipelineInvalid
	}
	plan, err := ExportPhase151BuildPostCreateCompletionPlan(scope, messageID, time.Now())
	if err != nil {
		return false, err
	}
	result, err := r.messages().UpdateOne(ctx, plan.Filter, plan.Update)
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
