package messenger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DeleteMessagesBySender removes every message authored by userID. Messages
// predate tenant-scoped storage, so Hub's singleton tenant is validated but is
// intentionally not added to the Mongo filter.
func (r *MongoRepository) DeleteMessagesBySender(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	if err := r.validateGDPRTarget(tenantID, userID); err != nil {
		return 0, err
	}

	result, err := r.db.Collection(messagesCollection).DeleteMany(ctx, bson.M{
		"sender_id": userID,
	})
	if err != nil {
		return 0, fmt.Errorf("delete messages by sender: %w", err)
	}

	return result.DeletedCount, nil
}

// DeleteReadStatesByUser removes the user's tenant-scoped unread receipts.
func (r *MongoRepository) DeleteReadStatesByUser(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	if err := r.validateGDPRTarget(tenantID, userID); err != nil {
		return 0, err
	}

	result, err := r.db.Collection(readStatesCollection).DeleteMany(ctx, bson.M{
		"tenant_id": tenantID,
		"user_id":   userID,
	})
	if err != nil {
		return 0, fmt.Errorf("delete read states by user: %w", err)
	}

	return result.DeletedCount, nil
}

// DeleteThreadFollowsByUser removes the user's tenant-scoped thread follows.
func (r *MongoRepository) DeleteThreadFollowsByUser(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	if err := r.validateGDPRTarget(tenantID, userID); err != nil {
		return 0, err
	}

	result, err := r.db.Collection(threadFollowsCollection).DeleteMany(ctx, bson.M{
		"tenant_id": tenantID,
		"user_id":   userID,
	})
	if err != nil {
		return 0, fmt.Errorf("delete thread follows by user: %w", err)
	}

	return result.DeletedCount, nil
}

// PullResolvedMentionUserID removes userID from resolved mention projections.
func (r *MongoRepository) PullResolvedMentionUserID(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	if err := r.validateGDPRTarget(tenantID, userID); err != nil {
		return 0, err
	}

	result, err := r.db.Collection(messagesCollection).UpdateMany(
		ctx,
		bson.M{"resolved_mention_user_ids": userID},
		bson.M{"$pull": bson.M{"resolved_mention_user_ids": userID}},
	)
	if err != nil {
		return 0, fmt.Errorf("pull resolved mention user id: %w", err)
	}

	return result.ModifiedCount, nil
}

func (r *MongoRepository) validateGDPRTarget(tenantID, userID uuid.UUID) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("mongo repository is not configured")
	}
	if tenantID == uuid.Nil {
		return fmt.Errorf("tenant id is required")
	}
	if userID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}

	return nil
}
