package messenger_test

// Тесты сервиса мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMessageHook struct {
	notify chan struct{}
}

func (h *stubMessageHook) OnMessageSent(_ context.Context, _ messenger.MessageHookEvent) error {
	h.notify <- struct{}{}
	return nil
}

// TestServiceBot_SendAndUpdateMessage проверяет бизнес-логику: сервис бот отправка and обновление message.
func TestServiceBot_SendAndUpdateMessage(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	botID := uuid.New()
	svc.SetSystemBotID(botID)

	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	msg, err := svc.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{
		ContentText: "hello from bot",
		ContentJSON: map[string]interface{}{"text": "hello from bot"},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello from bot", msg.ContentText)

	require.NoError(t, svc.UpdateBotMessage(ctx, msg.ID, "updated bot text", nil))

	otherUser := uuid.New()
	foreign := &messenger.MessageDocument{
		ID: uuid.New().String(), ChannelID: channelID, SenderID: otherUser,
		ContentText: "user msg", CreatedAt: time.Now(),
	}
	repo.messages[foreign.ID] = foreign
	err = svc.UpdateBotMessage(ctx, foreign.ID, "hack", nil)
	require.Error(t, err)
}

// TestServiceBot_ListAndFindMessages проверяет бизнес-логику: сервис бот список and find сообщения.
func TestServiceBot_ListAndFindMessages(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	botID := uuid.New()
	svc.SetSystemBotID(botID)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	parent, err := svc.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{
		ContentText: "thread",
		ContentJSON: map[string]interface{}{"text": "thread"},
	})
	require.NoError(t, err)

	_, err = svc.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{
		ContentText: "reply",
		ContentJSON: map[string]interface{}{"text": "reply"},
		ThreadID:    parent.ID,
	})
	require.NoError(t, err)

	_, err = svc.ListMessagesForBot(ctx, channelID, nil, 10)
	require.NoError(t, err)

	replies, err := svc.ListThreadRepliesForBot(ctx, parent.ID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, replies, 1)

	found, err := svc.FindMessageForBot(ctx, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, parent.ID, found.ID)

	ch, err := svc.GetChannelEntity(ctx, channelID)
	require.NoError(t, err)
	assert.Equal(t, channelID, ch.ID)

	members, err := svc.GetChannelMemberIDsInternal(ctx, channelID)
	require.NoError(t, err)
	assert.NotEmpty(t, members)
}

// TestServiceBot_NotConfiguredErrors проверяет бизнес-логику: сервис бот not configured errors.
func TestServiceBot_NotConfiguredErrors(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	channelID := uuid.New()

	_, err := svc.SendBotMessage(ctx, channelID, messenger.SendMessageRequest{ContentText: "x"})
	require.Error(t, err)

	_, err = svc.ListMessagesForBot(ctx, channelID, nil, 10)
	require.Error(t, err)

	_, err = svc.ListThreadRepliesForBot(ctx, "msg", nil, 10)
	require.Error(t, err)

	err = svc.UpdateBotMessage(ctx, "any", "text", nil)
	require.Error(t, err)

	// A2W3-2-003: guard на Find/GetChannel/MemberIDs без systemBotID.
	_, err = svc.FindMessageForBot(ctx, "any")
	require.Error(t, err)
	_, err = svc.GetChannelEntity(ctx, channelID)
	require.Error(t, err)
	_, err = svc.GetChannelMemberIDsInternal(ctx, channelID)
	require.Error(t, err)
}

// TestEmitMessageSent_InvokesHook проверяет сценарий: emit message sent invokes hook.
func TestEmitMessageSent_InvokesHook(t *testing.T) {
	svc, repo, _ := newTestService()
	hook := &stubMessageHook{notify: make(chan struct{}, 1)}
	svc.SetMessageSentHook(hook)
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	messenger.ExportMessengerServiceEmitMessageSent(svc, context.Background(), messenger.MessageHookEvent{
		ChannelID: channelID,
		SenderID:  ownerID,
	})
	select {
	case <-hook.notify:
	case <-time.After(2 * time.Second):
		t.Fatal("hook not invoked")
	}
}

// TestEmitMessageSent_SkipsSystemBot проверяет сценарий: emit message sent skips system бот.
func TestEmitMessageSent_SkipsSystemBot(t *testing.T) {
	svc, repo, _ := newTestService()
	botID := uuid.New()
	svc.SetSystemBotID(botID)
	hook := &stubMessageHook{notify: make(chan struct{}, 1)}
	svc.SetMessageSentHook(hook)
	channelID := setupChannelWithMembers(repo, "public", false, botID)

	messenger.ExportMessengerServiceEmitMessageSent(svc, context.Background(), messenger.MessageHookEvent{
		ChannelID: channelID,
		SenderID:  botID,
	})
	select {
	case <-hook.notify:
		t.Fatal("hook should not run for system bot")
	case <-time.After(50 * time.Millisecond):
	}
}
