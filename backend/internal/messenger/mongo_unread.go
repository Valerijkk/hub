package messenger

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	readStatesCollection    = "read_states"
	threadFollowsCollection = "thread_follows"

	phase151MongoUnreadMaxTargets           = userChannelMembershipCap
	phase151MongoUnreadMaxFollowedRoots     = 100
	phase151MongoUnreadMaxCandidates        = 1000
	phase151MongoUnreadMaxCandidatesPerRoot = 250
	phase151MongoUnreadCASAttempts          = 3
)

// mongoUnreadTarget is an already-authorized unread target. It deliberately
// carries no capability beyond the immutable PostgreSQL scope passed to the
// adapter method that uses it.
type mongoUnreadTarget struct {
	ChannelID uuid.UUID
	Kind      string
}

type mongoUnreadReadState struct {
	TenantID        uuid.UUID      `bson:"tenant_id"`
	UserID          uuid.UUID      `bson:"user_id"`
	Category        UnreadCategory `bson:"category"`
	TargetKind      string         `bson:"target_kind"`
	TargetID        string         `bson:"target_id"`
	CursorCreatedAt time.Time      `bson:"cursor_created_at"`
	CursorMessageID string         `bson:"cursor_message_id"`
	UpdatedAt       time.Time      `bson:"updated_at"`
}

func (state mongoUnreadReadState) position() MessagePosition {
	return MessagePosition{CreatedAt: phase151MongoCanonicalTime(state.CursorCreatedAt), MessageID: state.CursorMessageID}
}

type mongoUnreadThreadFollow struct {
	TenantID        uuid.UUID        `bson:"tenant_id"`
	UserID          uuid.UUID        `bson:"user_id"`
	ChannelID       uuid.UUID        `bson:"channel_id"`
	RootMessageID   string           `bson:"root_message_id"`
	Mode            ThreadFollowMode `bson:"mode"`
	CursorCreatedAt time.Time        `bson:"cursor_created_at"`
	CursorMessageID string           `bson:"cursor_message_id"`
	UpdatedAt       time.Time        `bson:"updated_at"`
}

func (r *MongoRepository) readStates() (*mongo.Collection, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("messenger Mongo database is required")
	}
	return r.db.Collection(readStatesCollection), nil
}

func (r *MongoRepository) threadFollows() (*mongo.Collection, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("messenger Mongo database is required")
	}
	return r.db.Collection(threadFollowsCollection), nil
}

func phase151MongoUnreadTargets(scope AuthorizedUnreadScope, category UnreadCategory) ([]mongoUnreadTarget, error) {
	if scope.TenantID == uuid.Nil || scope.UserID == uuid.Nil {
		return nil, fmt.Errorf("authorized unread scope is required")
	}
	if !isKnownUnreadCategory(category) {
		return nil, fmt.Errorf("unknown unread category %q", category)
	}
	if len(scope.Targets) > phase151MongoUnreadMaxTargets {
		return nil, fmt.Errorf("authorized unread scope exceeds %d targets", phase151MongoUnreadMaxTargets)
	}

	targets := make([]mongoUnreadTarget, 0, len(scope.Targets))
	seen := make(map[mongoUnreadTarget]struct{}, len(scope.Targets))
	for _, target := range scope.Targets {
		if target.ChannelID == uuid.Nil || !isKnownUnreadTargetKind(target.Kind) {
			return nil, fmt.Errorf("invalid authorized unread target")
		}
		if !phase151MongoCategoryAllowsTargetKind(category, target.Kind) {
			continue
		}
		converted := mongoUnreadTarget{ChannelID: target.ChannelID, Kind: target.Kind}
		if _, duplicate := seen[converted]; duplicate {
			continue
		}
		seen[converted] = struct{}{}
		targets = append(targets, converted)
	}
	return targets, nil
}

func phase151MongoUnreadScopeAllowsTarget(scope AuthorizedUnreadScope, channelID uuid.UUID, kind string) bool {
	if channelID == uuid.Nil || !isKnownUnreadTargetKind(kind) {
		return false
	}
	for _, target := range scope.Targets {
		if target.ChannelID == channelID && target.Kind == kind {
			return true
		}
	}
	return false
}

func phase151MongoCategoryAllowsTargetKind(category UnreadCategory, kind string) bool {
	switch category {
	case UnreadCategoryChannel, UnreadCategoryFollowedThread:
		return kind == "channel"
	case UnreadCategoryDM:
		return kind == "dm"
	case UnreadCategoryMention:
		return kind == "channel" || kind == "dm"
	default:
		return false
	}
}

func isKnownUnreadCategory(category UnreadCategory) bool {
	switch category {
	case UnreadCategoryChannel, UnreadCategoryDM, UnreadCategoryMention, UnreadCategoryFollowedThread:
		return true
	default:
		return false
	}
}

func isKnownUnreadTargetKind(kind string) bool {
	return kind == "channel" || kind == "dm"
}

func phase151MongoUnreadCandidateFilter(
	target mongoUnreadTarget,
	category UnreadCategory,
	userID uuid.UUID,
	position MessagePosition,
) bson.M {
	filter := bson.M{
		"channel_id":                target.ChannelID,
		"deleted_at":                nil,
		"sender_id":                 bson.M{"$ne": userID},
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
		"$and":                      bson.A{phase151MongoUnreadAfterTupleFilter(position)},
	}
	switch category {
	case UnreadCategoryChannel, UnreadCategoryDM:
		filter["$or"] = bson.A{
			bson.M{"thread_id": bson.M{"$exists": false}},
			bson.M{"also_send_to_channel": true},
		}
	case UnreadCategoryMention:
		filter["resolved_mention_user_ids"] = userID
	}
	return filter
}

func phase151MongoUnreadAfterTupleFilter(position MessagePosition) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"created_at": bson.M{"$gt": phase151MongoCanonicalTime(position.CreatedAt)}},
		bson.M{"created_at": phase151MongoCanonicalTime(position.CreatedAt), "_id": bson.M{"$gt": position.MessageID}},
	}}
}

func phase151MongoEligibleFollowRootFilter(scope AuthorizedUnreadScope, rootMessageIDs []string) (bson.M, error) {
	targets, err := phase151MongoUnreadTargets(scope, UnreadCategoryFollowedThread)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 || len(rootMessageIDs) == 0 {
		return nil, fmt.Errorf("authorized followed root target is required")
	}

	channels := make([]uuid.UUID, 0, len(targets))
	for _, target := range targets {
		channels = append(channels, target.ChannelID)
	}
	roots := phase151MongoSortedUniqueStrings(rootMessageIDs, phase151MongoUnreadMaxFollowedRoots)
	if len(roots) == 0 {
		return nil, fmt.Errorf("eligible root message id is required")
	}

	filter := bson.M{
		"_id":                       bson.M{"$in": roots},
		"thread_id":                 bson.M{"$exists": false},
		"deleted_at":                nil,
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
	}
	if len(channels) == 1 {
		filter["channel_id"] = channels[0]
	} else {
		filter["channel_id"] = bson.M{"$in": channels}
	}
	return filter, nil
}

func phase151MongoReadStateIdentityFilter(
	tenantID, userID uuid.UUID,
	category UnreadCategory,
	targetKind, targetID string,
) (bson.M, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || !isKnownUnreadCategory(category) || targetID == "" {
		return nil, fmt.Errorf("valid read state identity is required")
	}
	if !isKnownUnreadTargetKind(targetKind) && targetKind != "thread" {
		return nil, fmt.Errorf("valid read state target kind is required")
	}
	if (category == UnreadCategoryChannel && targetKind != "channel") ||
		(category == UnreadCategoryDM && targetKind != "dm") ||
		(category == UnreadCategoryFollowedThread && targetKind != "thread") {
		return nil, fmt.Errorf("read state category and target kind do not match")
	}
	return bson.M{
		"tenant_id":   tenantID,
		"user_id":     userID,
		"category":    category,
		"target_kind": targetKind,
		"target_id":   targetID,
	}, nil
}

// ListUnreadCandidates implements the durable, scope-bound unread coordinator.
// It never derives targets from Mongo: a valid non-empty PostgreSQL scope is
// required before it can read either state, follows, or message documents.
func (r *MongoRepository) ListUnreadCandidates(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
) ([]UnreadCandidate, error) {
	targets, err := phase151MongoUnreadTargets(scope, category)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return []UnreadCandidate{}, nil
	}
	if category == UnreadCategoryFollowedThread {
		return r.listFollowedThreadUnreadCandidates(ctx, scope, targets)
	}

	states, err := r.readStates()
	if err != nil {
		return nil, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return nil, err
	}

	result := make([]UnreadCandidate, 0)
	for _, target := range targets {
		if len(result) >= phase151MongoUnreadMaxCandidates {
			return nil, fmt.Errorf("unread candidate limit exceeded")
		}
		state, err := findMongoUnreadReadState(ctx, states, scope.TenantID, scope.UserID, category, target.Kind, target.ChannelID.String())
		if err != nil {
			return nil, err
		}
		if state == nil {
			// No baseline/state must not backfill historical messages as unread.
			continue
		}

		remaining := phase151MongoUnreadMaxCandidates - len(result)
		allowed := min(remaining, phase151MongoUnreadMaxCandidatesPerRoot)
		limit := allowed + 1
		cursor, err := messages.Find(ctx, phase151MongoUnreadCandidateFilter(target, category, scope.UserID, state.position()),
			options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)))
		if err != nil {
			return nil, fmt.Errorf("list %s unread messages: %w", category, err)
		}
		var documents []MessageDocument
		decodeErr := cursor.All(ctx, &documents)
		closeErr := cursor.Close(ctx)
		if decodeErr != nil {
			return nil, decodeErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(documents) > allowed {
			return nil, fmt.Errorf("%s unread candidate limit exceeded", category)
		}
		for _, document := range documents {
			if document.ID == "" || document.SenderID == scope.UserID || document.DeletedAt != nil ||
				document.PostCreateEffectsState == PostCreateEffectsStatePending {
				continue
			}
			result = append(result, UnreadCandidate{
				TargetID: target.ChannelID.String(), MessageID: document.ID, SenderID: document.SenderID,
			})
		}
	}
	return result, nil
}

func (r *MongoRepository) listFollowedThreadUnreadCandidates(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	targets []mongoUnreadTarget,
) ([]UnreadCandidate, error) {
	states, err := r.readStates()
	if err != nil {
		return nil, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return nil, err
	}

	result := make([]UnreadCandidate, 0)
	for _, target := range targets {
		if len(result) >= phase151MongoUnreadMaxCandidates {
			return nil, fmt.Errorf("unread candidate limit exceeded")
		}
		roots, err := r.ListEligiblePersistedFollowedRoots(ctx, scope, AuthorizedUnreadTarget{ChannelID: target.ChannelID, Kind: target.Kind})
		if err != nil {
			return nil, err
		}
		for _, rootMessageID := range roots {
			if len(result) >= phase151MongoUnreadMaxCandidates {
				return nil, fmt.Errorf("unread candidate limit exceeded")
			}
			state, err := findMongoUnreadReadState(ctx, states, scope.TenantID, scope.UserID, UnreadCategoryFollowedThread, "thread", rootMessageID)
			if err != nil {
				return nil, err
			}
			if state == nil {
				continue
			}
			remaining := phase151MongoUnreadMaxCandidates - len(result)
			allowed := min(remaining, phase151MongoUnreadMaxCandidatesPerRoot)
			limit := allowed + 1
			filter := phase151MongoFollowedThreadCandidateFilter(target.ChannelID, rootMessageID, scope.UserID, state.position())
			cursor, err := messages.Find(ctx, filter,
				options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "_id", Value: 1}}).SetLimit(int64(limit)))
			if err != nil {
				return nil, fmt.Errorf("list followed-thread unread messages: %w", err)
			}
			var documents []MessageDocument
			decodeErr := cursor.All(ctx, &documents)
			closeErr := cursor.Close(ctx)
			if decodeErr != nil {
				return nil, decodeErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			if len(documents) > allowed {
				return nil, fmt.Errorf("followed-thread unread candidate limit exceeded")
			}
			for _, document := range documents {
				if document.ID == "" || document.SenderID == scope.UserID || document.DeletedAt != nil ||
					document.PostCreateEffectsState == PostCreateEffectsStatePending {
					continue
				}
				result = append(result, UnreadCandidate{TargetID: rootMessageID, MessageID: document.ID, SenderID: document.SenderID})
			}
		}
	}
	return result, nil
}

func phase151MongoFollowedThreadCandidateFilter(
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
	position MessagePosition,
) bson.M {
	return bson.M{
		"channel_id":                channelID,
		"thread_id":                 rootMessageID,
		"deleted_at":                nil,
		"sender_id":                 bson.M{"$ne": userID},
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
		"$and":                      bson.A{phase151MongoUnreadAfterTupleFilter(position)},
	}
}

// ListEligiblePersistedFollowedRoots intersects stored follows with the exact
// passed scope and then validates live root documents. A stale follow can
// therefore never widen a scope or reveal an inaccessible/deleted root.
func (r *MongoRepository) ListEligiblePersistedFollowedRoots(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	target AuthorizedUnreadTarget,
) ([]string, error) {
	if !phase151MongoUnreadScopeAllowsTarget(scope, target.ChannelID, target.Kind) || target.Kind != "channel" {
		return nil, fmt.Errorf("authorized followed-thread target is required")
	}
	if _, err := phase151MongoUnreadTargets(scope, UnreadCategoryFollowedThread); err != nil {
		return nil, err
	}
	follows, err := r.threadFollows()
	if err != nil {
		return nil, err
	}
	cursor, err := follows.Find(ctx, bson.M{
		"tenant_id":  scope.TenantID,
		"user_id":    scope.UserID,
		"channel_id": target.ChannelID,
		"mode": bson.M{"$in": bson.A{
			ThreadFollowModeAuto, ThreadFollowModeManualFollow,
		}},
	}, options.Find().SetProjection(bson.M{"root_message_id": 1}).SetLimit(phase151MongoUnreadMaxFollowedRoots + 1))
	if err != nil {
		return nil, fmt.Errorf("list persisted thread follows: %w", err)
	}
	var followsForTarget []mongoUnreadThreadFollow
	decodeErr := cursor.All(ctx, &followsForTarget)
	closeErr := cursor.Close(ctx)
	if decodeErr != nil {
		return nil, decodeErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(followsForTarget) > phase151MongoUnreadMaxFollowedRoots {
		return nil, fmt.Errorf("followed root limit exceeded")
	}

	rootIDs := make([]string, 0, len(followsForTarget))
	for _, follow := range followsForTarget {
		if follow.RootMessageID != "" {
			rootIDs = append(rootIDs, follow.RootMessageID)
		}
	}
	filter, err := phase151MongoEligibleFollowRootFilter(AuthorizedUnreadScope{
		TenantID: scope.TenantID,
		UserID:   scope.UserID,
		Targets:  []AuthorizedUnreadTarget{target},
	}, rootIDs)
	if err != nil {
		if len(rootIDs) == 0 {
			return []string{}, nil
		}
		return nil, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return nil, err
	}
	rootCursor, err := messages.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(phase151MongoUnreadMaxFollowedRoots + 1))
	if err != nil {
		return nil, fmt.Errorf("validate followed roots: %w", err)
	}
	var roots []MessageDocument
	decodeErr = rootCursor.All(ctx, &roots)
	closeErr = rootCursor.Close(ctx)
	if decodeErr != nil {
		return nil, decodeErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	result := make([]string, 0, len(roots))
	for _, root := range roots {
		if root.ID != "" {
			result = append(result, root.ID)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (r *MongoRepository) CurrentAuthorizedTargetHead(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	target AuthorizedUnreadTarget,
	category UnreadCategory,
) (MessagePosition, error) {
	if !phase151MongoUnreadScopeAllowsTarget(scope, target.ChannelID, target.Kind) ||
		!phase151MongoCategoryAllowsTargetKind(category, target.Kind) || category == UnreadCategoryFollowedThread {
		return MessagePosition{}, fmt.Errorf("authorized unread target is required")
	}
	if _, err := phase151MongoUnreadTargets(scope, category); err != nil {
		return MessagePosition{}, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return MessagePosition{}, err
	}
	filter := bson.M{
		"channel_id":                target.ChannelID,
		"deleted_at":                nil,
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
		"$or": bson.A{
			bson.M{"thread_id": bson.M{"$exists": false}},
			bson.M{"also_send_to_channel": true},
		},
	}
	return findMongoUnreadHead(ctx, messages, filter)
}

func (r *MongoRepository) CurrentAuthorizedFollowedRootHead(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	target AuthorizedUnreadTarget,
	rootMessageID string,
) (MessagePosition, error) {
	if rootMessageID == "" || !phase151MongoUnreadScopeAllowsTarget(scope, target.ChannelID, target.Kind) || target.Kind != "channel" {
		return MessagePosition{}, fmt.Errorf("authorized followed root is required")
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return MessagePosition{}, err
	}
	rootFilter, err := phase151MongoEligibleFollowRootFilter(AuthorizedUnreadScope{
		TenantID: scope.TenantID,
		UserID:   scope.UserID,
		Targets:  []AuthorizedUnreadTarget{target},
	}, []string{rootMessageID})
	if err != nil {
		return MessagePosition{}, err
	}
	var root MessageDocument
	if err := messages.FindOne(ctx, rootFilter).Decode(&root); err != nil {
		if err == mongo.ErrNoDocuments {
			return MessagePosition{}, nil
		}
		return MessagePosition{}, err
	}
	return findMongoUnreadHead(ctx, messages, bson.M{
		"channel_id":                target.ChannelID,
		"deleted_at":                nil,
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
		"$or": bson.A{
			bson.M{"_id": rootMessageID},
			bson.M{"thread_id": rootMessageID},
		},
	})
}

func findMongoUnreadHead(ctx context.Context, messages *mongo.Collection, filter bson.M) (MessagePosition, error) {
	var document MessageDocument
	err := messages.FindOne(ctx, filter,
		options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
			SetProjection(bson.M{"_id": 1, "created_at": 1})).Decode(&document)
	if err == mongo.ErrNoDocuments {
		return MessagePosition{}, nil
	}
	if err != nil {
		return MessagePosition{}, err
	}
	return MessagePosition{CreatedAt: phase151MongoCanonicalTime(document.CreatedAt), MessageID: document.ID}, nil
}

func (r *MongoRepository) PersistUnreadCutoverBaseline(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
	targetID string,
	position MessagePosition,
) error {
	targetKind, err := r.phase151MongoUnreadTargetKindForState(ctx, scope, category, targetID)
	if err != nil {
		return err
	}
	_, err = r.updateMongoUnreadReadState(ctx, scope.TenantID, scope.UserID, category, targetKind, targetID, position, true)
	return err
}

// LoadAuthorizedUnreadReadState reads only a category/target that is already
// present in the supplied immutable PostgreSQL scope.
func (r *MongoRepository) LoadAuthorizedUnreadReadState(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
	targetID string,
) (MessagePosition, bool, error) {
	targetKind, err := r.phase151MongoUnreadTargetKindForState(ctx, scope, category, targetID)
	if err != nil {
		return MessagePosition{}, false, err
	}
	states, err := r.readStates()
	if err != nil {
		return MessagePosition{}, false, err
	}
	state, err := findMongoUnreadReadState(ctx, states, scope.TenantID, scope.UserID, category, targetKind, targetID)
	if err != nil {
		return MessagePosition{}, false, err
	}
	if state == nil {
		return MessagePosition{}, false, nil
	}
	return state.position(), true, nil
}

// AdvanceAuthorizedUnreadCursor is the scope-bound receipt primitive. It
// atomically advances one category cursor and rejects any stale/equal tuple.
func (r *MongoRepository) AdvanceAuthorizedUnreadCursor(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
	targetID string,
	position MessagePosition,
) (bool, error) {
	if !phase151MongoNonEmptyPosition(position) {
		return false, fmt.Errorf("unread cursor position is required")
	}
	targetKind, err := r.phase151MongoUnreadTargetKindForState(ctx, scope, category, targetID)
	if err != nil {
		return false, err
	}
	return r.updateMongoUnreadReadState(ctx, scope.TenantID, scope.UserID, category, targetKind, targetID, position, false)
}

func (r *MongoRepository) phase151MongoUnreadTargetKindForState(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
	targetID string,
) (string, error) {
	targets, err := phase151MongoUnreadTargets(scope, category)
	if err != nil {
		return "", err
	}
	if targetID == "" {
		return "", fmt.Errorf("unread target id is required")
	}
	if category == UnreadCategoryFollowedThread {
		for _, target := range targets {
			roots, rootsErr := r.ListEligiblePersistedFollowedRoots(ctx, scope, AuthorizedUnreadTarget{ChannelID: target.ChannelID, Kind: target.Kind})
			if rootsErr != nil {
				return "", rootsErr
			}
			for _, rootMessageID := range roots {
				if rootMessageID == targetID {
					return "thread", nil
				}
			}
		}
		return "", fmt.Errorf("followed root is not authorized")
	}
	for _, target := range targets {
		if target.ChannelID.String() == targetID {
			return target.Kind, nil
		}
	}
	return "", fmt.Errorf("unread target is not authorized")
}

func (r *MongoRepository) updateMongoUnreadReadState(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	category UnreadCategory,
	targetKind, targetID string,
	position MessagePosition,
	allowEmptyPosition bool,
) (bool, error) {
	if (!allowEmptyPosition && !phase151MongoNonEmptyPosition(position)) ||
		(allowEmptyPosition && !position.CreatedAt.IsZero() && position.MessageID == "") {
		return false, fmt.Errorf("valid unread cursor position is required")
	}
	identity, err := phase151MongoReadStateIdentityFilter(tenantID, userID, category, targetKind, targetID)
	if err != nil {
		return false, err
	}
	states, err := r.readStates()
	if err != nil {
		return false, err
	}
	position.CreatedAt = phase151MongoCanonicalTime(position.CreatedAt)
	now := time.Now().UTC()

	for attempt := 0; attempt < phase151MongoUnreadCASAttempts; attempt++ {
		filter := cloneMongoFilter(identity)
		filter["$or"] = bson.A{
			bson.M{"cursor_created_at": bson.M{"$exists": false}},
			bson.M{"cursor_created_at": bson.M{"$lt": position.CreatedAt}},
			bson.M{"cursor_created_at": position.CreatedAt, "cursor_message_id": bson.M{"$lt": position.MessageID}},
		}
		update := bson.M{
			"$setOnInsert": identity,
			"$set": bson.M{
				"cursor_created_at": position.CreatedAt,
				"cursor_message_id": position.MessageID,
				"updated_at":        now,
			},
		}
		result, updateErr := states.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
		if updateErr == nil {
			if result.MatchedCount > 0 || result.ModifiedCount > 0 || result.UpsertedCount > 0 {
				return true, nil
			}
			return false, nil
		}
		if !mongo.IsDuplicateKeyError(updateErr) {
			return false, updateErr
		}
		current, readErr := findMongoUnreadReadState(ctx, states, tenantID, userID, category, targetKind, targetID)
		if readErr != nil {
			return false, readErr
		}
		if current != nil && !messagePositionAfter(position, current.position()) {
			return false, nil
		}
	}
	return false, fmt.Errorf("unread cursor compare-and-set contention")
}

func findMongoUnreadReadState(
	ctx context.Context,
	states *mongo.Collection,
	tenantID, userID uuid.UUID,
	category UnreadCategory,
	targetKind, targetID string,
) (*mongoUnreadReadState, error) {
	identity, err := phase151MongoReadStateIdentityFilter(tenantID, userID, category, targetKind, targetID)
	if err != nil {
		return nil, err
	}
	var state mongoUnreadReadState
	err = states.FindOne(ctx, identity).Decode(&state)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func phase151MongoCanonicalTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC().Truncate(time.Millisecond)
}

func phase151MongoNonEmptyPosition(position MessagePosition) bool {
	return !position.CreatedAt.IsZero() && position.MessageID != ""
}

// ListAuthorizedUnreadMessageIDs supplies Dashboard with a set-ready,
// scope-bound projection. It intentionally delegates category selection to the
// same durable coordinator so Dashboard cannot reconstruct unread truth.
func (r *MongoRepository) ListAuthorizedUnreadMessageIDs(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	category UnreadCategory,
) ([]string, error) {
	candidates, err := r.ListUnreadCandidates(ctx, scope, category)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate.MessageID == "" {
			continue
		}
		if _, duplicate := seen[candidate.MessageID]; duplicate {
			continue
		}
		seen[candidate.MessageID] = struct{}{}
		ids = append(ids, candidate.MessageID)
	}
	sort.Strings(ids)
	return ids, nil
}

func (r *MongoRepository) FindThreadFollow(
	ctx context.Context,
	tenantID, userID, channelID uuid.UUID,
	rootMessageID string,
) (string, bool, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || channelID == uuid.Nil || rootMessageID == "" {
		return "", false, fmt.Errorf("valid thread follow identity is required")
	}
	follows, err := r.threadFollows()
	if err != nil {
		return "", false, err
	}
	var follow mongoUnreadThreadFollow
	err = follows.FindOne(ctx, bson.M{
		"tenant_id": tenantID, "user_id": userID, "channel_id": channelID, "root_message_id": rootMessageID,
	}).Decode(&follow)
	if err == mongo.ErrNoDocuments {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !isKnownThreadFollowMode(follow.Mode) {
		return "", false, nil
	}
	return string(follow.Mode), true, nil
}

// AdvanceThreadReadCursor is retained for the exact thread receipt runtime.
// The service establishes fresh ChannelAccess before this call; the adapter
// independently rechecks live root/reply identity and advances only the
// followed-thread cursor for that exact root.
func (r *MongoRepository) AdvanceThreadReadCursor(
	ctx context.Context,
	tenantID, userID, channelID uuid.UUID,
	rootMessageID, messageID string,
) (bool, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || channelID == uuid.Nil || rootMessageID == "" || messageID == "" {
		return false, fmt.Errorf("valid followed-thread receipt is required")
	}
	follows, err := r.threadFollows()
	if err != nil {
		return false, err
	}
	var follow mongoUnreadThreadFollow
	err = follows.FindOne(ctx, bson.M{
		"tenant_id": tenantID, "user_id": userID, "channel_id": channelID, "root_message_id": rootMessageID,
		"mode": bson.M{"$in": bson.A{ThreadFollowModeAuto, ThreadFollowModeManualFollow}},
	}).Decode(&follow)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	messages, err := r.mongoMessages()
	if err != nil {
		return false, err
	}
	rootFilter := bson.M{
		"_id": rootMessageID, "channel_id": channelID, "thread_id": bson.M{"$exists": false},
		"deleted_at": nil, "post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
	}
	var root MessageDocument
	if err := messages.FindOne(ctx, rootFilter).Decode(&root); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	var candidate MessageDocument
	err = messages.FindOne(ctx, bson.M{
		"_id": messageID, "channel_id": channelID, "deleted_at": nil,
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
	}).Decode(&candidate)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	if candidate.ID != rootMessageID && candidate.ThreadID != rootMessageID {
		return false, nil
	}
	position := MessagePosition{CreatedAt: phase151MongoCanonicalTime(candidate.CreatedAt), MessageID: candidate.ID}
	advanced, err := r.updateMongoUnreadReadState(ctx, tenantID, userID, UnreadCategoryFollowedThread, "thread", rootMessageID, position, false)
	if err != nil || !advanced {
		return advanced, err
	}
	_, err = follows.UpdateOne(ctx, bson.M{
		"tenant_id": tenantID, "user_id": userID, "channel_id": channelID, "root_message_id": rootMessageID,
		"mode": bson.M{"$in": bson.A{ThreadFollowModeAuto, ThreadFollowModeManualFollow}},
		"$or": bson.A{
			bson.M{"cursor_created_at": bson.M{"$exists": false}},
			bson.M{"cursor_created_at": bson.M{"$lt": position.CreatedAt}},
			bson.M{"cursor_created_at": position.CreatedAt, "cursor_message_id": bson.M{"$lt": position.MessageID}},
		},
	}, bson.M{"$set": bson.M{
		"cursor_created_at": position.CreatedAt, "cursor_message_id": position.MessageID, "updated_at": time.Now().UTC(),
	}})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MongoRepository) mongoMessages() (*mongo.Collection, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("messenger Mongo database is required")
	}
	return r.messages(), nil
}

func cloneMongoFilter(source bson.M) bson.M {
	clone := make(bson.M, len(source)+1)
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func phase151MongoSortedUniqueStrings(values []string, limit int) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	if len(result) > limit {
		return result[:limit]
	}
	return result
}
