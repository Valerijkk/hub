package messenger

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMongoUnreadThreadFollowIdentityIsTenantUserChannelRootBound(t *testing.T) {
	tenantID, userID, channelID := uuid.New(), uuid.New(), uuid.New()
	filter, err := phase151MongoThreadFollowIdentityFilter(tenantID, userID, channelID, "root-001")
	require.NoError(t, err)
	assert.Equal(t, bson.M{
		"tenant_id":       tenantID,
		"user_id":         userID,
		"channel_id":      channelID,
		"root_message_id": "root-001",
	}, filter)

	_, err = phase151MongoThreadFollowIdentityFilter(tenantID, userID, uuid.Nil, "root-001")
	require.Error(t, err)
}

func TestMongoUnreadAutoFollowNeverReplacesManualIntent(t *testing.T) {
	current := ThreadFollowState{Mode: ThreadFollowModeManualUnfollow, Cursor: MessagePosition{MessageID: "message-001"}}
	changed, next := phase151MongoThreadFollowTransition(&current, ThreadFollowTransition{
		Mode: ThreadFollowModeAuto, Position: MessagePosition{MessageID: "message-002"},
	})
	assert.False(t, changed)
	assert.Equal(t, current, next)
}
