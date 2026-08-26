package messenger

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMongoUnreadAdapterImplementsDurableUnreadContracts(t *testing.T) {
	var _ Phase151UnreadCoordinatorRepository = (*MongoRepository)(nil)
	var _ Phase151UnreadCutoverRepository = (*MongoRepository)(nil)
	var _ Phase151DashboardUnreadRepository = (*MongoRepository)(nil)
}

func TestMongoUnreadScopeTargetsAreCategoryBoundedAndFailClosed(t *testing.T) {
	channelID := uuid.New()
	dmID := uuid.New()
	scope := AuthorizedUnreadScope{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Targets: []AuthorizedUnreadTarget{
			{ChannelID: channelID, Kind: "channel"},
			{ChannelID: dmID, Kind: "dm"},
			{ChannelID: channelID, Kind: "channel"},
		},
	}

	channelTargets, err := phase151MongoUnreadTargets(scope, UnreadCategoryChannel)
	require.NoError(t, err)
	assert.Equal(t, []mongoUnreadTarget{{ChannelID: channelID, Kind: "channel"}}, channelTargets)

	dmTargets, err := phase151MongoUnreadTargets(scope, UnreadCategoryDM)
	require.NoError(t, err)
	assert.Equal(t, []mongoUnreadTarget{{ChannelID: dmID, Kind: "dm"}}, dmTargets)

	mentionTargets, err := phase151MongoUnreadTargets(scope, UnreadCategoryMention)
	require.NoError(t, err)
	assert.ElementsMatch(t, []mongoUnreadTarget{
		{ChannelID: channelID, Kind: "channel"},
		{ChannelID: dmID, Kind: "dm"},
	}, mentionTargets)

	_, err = phase151MongoUnreadTargets(AuthorizedUnreadScope{}, UnreadCategoryChannel)
	require.Error(t, err)

	overflow := AuthorizedUnreadScope{TenantID: uuid.New(), UserID: uuid.New()}
	for index := 0; index < phase151MongoUnreadMaxTargets+1; index++ {
		overflow.Targets = append(overflow.Targets, AuthorizedUnreadTarget{ChannelID: uuid.New(), Kind: "channel"})
	}
	_, err = phase151MongoUnreadTargets(overflow, UnreadCategoryChannel)
	require.Error(t, err)
}

func TestMongoUnreadCandidateFiltersRequireLiveCompletedNonOwnMessagesAfterTuple(t *testing.T) {
	channelID := uuid.New()
	userID := uuid.New()
	position := MessagePosition{
		CreatedAt: time.Date(2026, time.August, 25, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60)),
		MessageID: "message-002",
	}

	filter := phase151MongoUnreadCandidateFilter(
		mongoUnreadTarget{ChannelID: channelID, Kind: "channel"},
		UnreadCategoryChannel,
		userID,
		position,
	)

	assert.Equal(t, bson.M{
		"channel_id":                channelID,
		"deleted_at":                nil,
		"sender_id":                 bson.M{"$ne": userID},
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
		"$or": bson.A{
			bson.M{"thread_id": bson.M{"$exists": false}},
			bson.M{"also_send_to_channel": true},
		},
		"$and": bson.A{
			bson.M{"$or": bson.A{
				bson.M{"created_at": bson.M{"$gt": position.CreatedAt.UTC()}},
				bson.M{"created_at": position.CreatedAt.UTC(), "_id": bson.M{"$gt": position.MessageID}},
			}},
		},
	}, filter)
}

func TestMongoUnreadFollowRootFilterIntersectsOnlyPassedAuthorizedScope(t *testing.T) {
	allowedChannelID := uuid.New()
	foreignChannelID := uuid.New()
	scope := AuthorizedUnreadScope{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Targets:  []AuthorizedUnreadTarget{{ChannelID: allowedChannelID, Kind: "channel"}},
	}

	filter, err := phase151MongoEligibleFollowRootFilter(scope, []string{"root-b", "root-a", "root-b"})
	require.NoError(t, err)
	assert.Equal(t, bson.M{
		"channel_id":                allowedChannelID,
		"_id":                       bson.M{"$in": []string{"root-a", "root-b"}},
		"thread_id":                 bson.M{"$exists": false},
		"deleted_at":                nil,
		"post_create_effects_state": bson.M{"$ne": PostCreateEffectsStatePending},
	}, filter)

	assert.False(t, phase151MongoUnreadScopeAllowsTarget(scope, foreignChannelID, "channel"))
	assert.False(t, phase151MongoUnreadScopeAllowsTarget(scope, allowedChannelID, "dm"))
}

func TestMongoUnreadReadStateFilterUsesTupleIdentity(t *testing.T) {
	tenantID, userID := uuid.New(), uuid.New()
	filter, err := phase151MongoReadStateIdentityFilter(
		tenantID,
		userID,
		UnreadCategoryMention,
		"channel",
		"target-001",
	)
	require.NoError(t, err)
	assert.Equal(t, bson.M{
		"tenant_id":   tenantID,
		"user_id":     userID,
		"category":    UnreadCategoryMention,
		"target_kind": "channel",
		"target_id":   "target-001",
	}, filter)

	_, err = phase151MongoReadStateIdentityFilter(tenantID, userID, UnreadCategory("unknown"), "channel", "target-001")
	require.Error(t, err)
}
