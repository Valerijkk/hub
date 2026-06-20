package calls

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallChatService_PersistentMessageFlow(t *testing.T) {
	chatRepo := newMemChatRepository()
	f := newCallsTestFixture(t)
	ctx := context.Background()
	f.svc.chat = NewCallChatService(f.svc.repo, chatRepo, f.svc.pub)

	callEnt := f.createChannelCall(t, f.owner)

	msg, err := f.svc.chat.SendMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "persist me", false)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "persist me", msg.Text)

	msgs, err := f.svc.ListChatMessages(ctx, callEnt.ID.String(), 10)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

	f.svc.chat.DeleteCallChat(ctx, callEnt.ID.String())
	msgs, err = f.svc.ListChatMessages(ctx, callEnt.ID.String(), 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}

func TestOnRoomFinished_DeletesCallChat(t *testing.T) {
	chatRepo := newMemChatRepository()
	f := newCallsTestFixture(t)
	ctx := context.Background()
	f.svc.chat = NewCallChatService(f.svc.repo, chatRepo, f.svc.pub)

	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.svc.chat.SendMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "cleanup", false)
	require.NoError(t, err)

	require.NoError(t, f.svc.OnRoomFinished(ctx, callEnt.RoomName))

	msgs, err := f.svc.ListChatMessages(ctx, callEnt.ID.String(), 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}
