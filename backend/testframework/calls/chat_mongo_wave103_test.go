package calls_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

func TestMongoChatStore_Lifecycle(t *testing.T) {
	db := mongotest.Connect(t, "hub_test_calls_")
	coll := db.Collection("call_chat_messages")
	store := calls.ExportNewMongoChatStore(coll)
	ctx := context.Background()

	require.NoError(t, store.EnsureIndexes(ctx))

	msg := &calls.CallChatMessage{
		ID:        bson.NewObjectID(),
		CallID:    "call-wave103",
		UserID:    "user-1",
		UserName:  "Caller",
		Text:      "Hello call chat",
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, store.Insert(ctx, msg))

	listed, err := store.ListByCallID(ctx, "call-wave103", 10)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "Hello call chat", listed[0].Text)

	require.NoError(t, store.DeleteByCallID(ctx, "call-wave103"))
	listed, err = store.ListByCallID(ctx, "call-wave103", 10)
	require.NoError(t, err)
	assert.Empty(t, listed)
}
