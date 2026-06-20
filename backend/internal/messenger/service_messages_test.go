package messenger

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditMessage_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	sent, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "original"},
		ContentText: "original",
	})
	require.NoError(t, err)

	edited, err := svc.EditMessage(ctx, sent.ID, ownerID, EditMessageRequest{
		ContentJSON: map[string]interface{}{"text": "edited"},
		ContentText: "edited",
	})
	require.NoError(t, err)
	assert.Equal(t, "edited", edited.ContentText)
}

func TestEditMessage_ForbiddenForOtherUser(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	sent, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "mine"},
		ContentText: "mine",
	})
	require.NoError(t, err)

	_, err = svc.EditMessage(ctx, sent.ID, memberID, EditMessageRequest{
		ContentJSON: map[string]interface{}{"text": "hack"},
		ContentText: "hack",
	})
	require.Error(t, err)
}

func TestListMessages_ReturnsChannelMessages(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "one"},
		ContentText: "one",
	})
	require.NoError(t, err)

	msgs, err := svc.ListMessages(ctx, channelID, ownerID, nil, 50)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	assert.Equal(t, "one", msgs[0].ContentText)
}

func TestForwardMessage_ToAnotherChannel(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	sourceID := setupChannelWithMembers(repo, "public", false, ownerID)
	targetID := setupChannelWithMembers(repo, "public", false, ownerID)

	sent, err := svc.SendMessage(ctx, sourceID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "forward me"},
		ContentText: "forward me",
	})
	require.NoError(t, err)

	forwarded, err := svc.ForwardMessage(ctx, sourceID, ownerID, sent.ID, targetID.String())
	require.NoError(t, err)
	assert.Contains(t, forwarded.ContentText, "forward me")
}

func TestSearchMessages_FindsByQuery(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "unique-token-xyz"},
		ContentText: "unique-token-xyz",
	})
	require.NoError(t, err)

	results, err := svc.SearchMessages(ctx, "unique-token", channelID.String(), 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestCreateOrGetDM_CreatesChannel(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userA := uuid.New()
	userB := uuid.New()

	ch, err := svc.CreateOrGetDM(ctx, userA, userB)
	require.NoError(t, err)
	assert.NotEmpty(t, ch.ID)

	channelID, err := uuid.Parse(ch.ID)
	require.NoError(t, err)
	isMember, err := repo.IsMember(ctx, channelID, userA)
	require.NoError(t, err)
	assert.True(t, isMember)
}

func TestListThreadReplies(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	parent, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "thread root"},
		ContentText: "thread root",
	})
	require.NoError(t, err)

	_, err = svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "reply"},
		ContentText: "reply",
		ThreadID:    parent.ID,
	})
	require.NoError(t, err)

	replies, err := svc.ListThreadReplies(ctx, parent.ID, ownerID, nil, 20)
	require.NoError(t, err)
	require.Len(t, replies, 1)
	assert.Equal(t, "reply", replies[0].ContentText)
	_ = time.Now()
}
