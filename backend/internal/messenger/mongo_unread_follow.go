package messenger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func phase151MongoThreadFollowIdentityFilter(
	tenantID, userID, channelID uuid.UUID,
	rootMessageID string,
) (bson.M, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || channelID == uuid.Nil || rootMessageID == "" {
		return nil, fmt.Errorf("valid thread follow identity is required")
	}
	return bson.M{
		"tenant_id":       tenantID,
		"user_id":         userID,
		"channel_id":      channelID,
		"root_message_id": rootMessageID,
	}, nil
}

func phase151MongoThreadFollowTransition(
	current *ThreadFollowState,
	transition ThreadFollowTransition,
) (bool, ThreadFollowState) {
	changed, next := ExportPhase151DecideThreadFollowTransition(current, transition)
	if !changed || current == nil {
		return changed, next
	}
	if next.Mode == current.Mode && !messagePositionAfter(next.Cursor, current.Cursor) {
		return false, *current
	}
	return true, next
}

// ApplyAuthorizedThreadFollowTransition persists explicit follow intent only
// after validating the live root against the supplied PostgreSQL unread scope.
// Automatic follow is insert-only, so it can never overwrite manual follow or
// manual unfollow during a retry/race.
func (r *MongoRepository) ApplyAuthorizedThreadFollowTransition(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	channelID uuid.UUID,
	rootMessageID string,
	transition ThreadFollowTransition,
) (bool, error) {
	if !phase151MongoUnreadScopeAllowsTarget(scope, channelID, "channel") ||
		!isKnownThreadFollowMode(transition.Mode) || !phase151MongoNonEmptyPosition(transition.Position) {
		return false, fmt.Errorf("authorized thread follow transition is required")
	}
	if _, err := phase151MongoUnreadTargets(scope, UnreadCategoryFollowedThread); err != nil {
		return false, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return false, err
	}
	rootFilter, err := phase151MongoEligibleFollowRootFilter(AuthorizedUnreadScope{
		TenantID: scope.TenantID,
		UserID:   scope.UserID,
		Targets:  []AuthorizedUnreadTarget{{ChannelID: channelID, Kind: "channel"}},
	}, []string{rootMessageID})
	if err != nil {
		return false, err
	}
	var root MessageDocument
	if err := messages.FindOne(ctx, rootFilter, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&root); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	follows, err := r.threadFollows()
	if err != nil {
		return false, err
	}
	identity, err := phase151MongoThreadFollowIdentityFilter(scope.TenantID, scope.UserID, channelID, rootMessageID)
	if err != nil {
		return false, err
	}
	transition.Position.CreatedAt = phase151MongoCanonicalTime(transition.Position.CreatedAt)

	for attempt := 0; attempt < phase151MongoUnreadCASAttempts; attempt++ {
		current, found, readErr := findMongoUnreadThreadFollow(ctx, follows, identity)
		if readErr != nil {
			return false, readErr
		}
		var currentState *ThreadFollowState
		if found {
			currentState = &ThreadFollowState{Mode: current.Mode, Cursor: MessagePosition{
				CreatedAt: phase151MongoCanonicalTime(current.CursorCreatedAt), MessageID: current.CursorMessageID,
			}}
		}
		changed, next := phase151MongoThreadFollowTransition(currentState, transition)
		if !changed {
			return false, nil
		}

		now := time.Now().UTC()
		update := bson.M{
			"$set": bson.M{
				"mode":              next.Mode,
				"cursor_created_at": phase151MongoCanonicalTime(next.Cursor.CreatedAt),
				"cursor_message_id": next.Cursor.MessageID,
				"updated_at":        now,
			},
		}
		filter := cloneMongoFilter(identity)
		if transition.Mode == ThreadFollowModeAuto {
			filter["mode"] = bson.M{"$exists": false}
			update["$setOnInsert"] = identity
			result, updateErr := follows.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
			if updateErr == nil && result.UpsertedCount == 1 {
				return true, nil
			}
			if updateErr != nil && !mongo.IsDuplicateKeyError(updateErr) {
				return false, updateErr
			}
			continue
		}

		if found {
			filter["mode"] = current.Mode
			result, updateErr := follows.UpdateOne(ctx, filter, update)
			if updateErr != nil {
				return false, updateErr
			}
			if result.MatchedCount == 1 {
				return true, nil
			}
			continue
		}

		update["$setOnInsert"] = identity
		result, updateErr := follows.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
		if updateErr == nil && result.UpsertedCount == 1 {
			return true, nil
		}
		if updateErr != nil && !mongo.IsDuplicateKeyError(updateErr) {
			return false, updateErr
		}
	}
	return false, fmt.Errorf("thread follow compare-and-set contention")
}

func findMongoUnreadThreadFollow(
	ctx context.Context,
	follows *mongo.Collection,
	identity bson.M,
) (*mongoUnreadThreadFollow, bool, error) {
	var follow mongoUnreadThreadFollow
	err := follows.FindOne(ctx, identity).Decode(&follow)
	if err == mongo.ErrNoDocuments {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !isKnownThreadFollowMode(follow.Mode) {
		return nil, false, nil
	}
	return &follow, true, nil
}
