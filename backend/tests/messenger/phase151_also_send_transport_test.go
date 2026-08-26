package messenger_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestSendMessageAlsoSendPersistsProjectsAndPublishesOneCanonicalRecord(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	publisher := &capturePublisher{}
	service := messenger.ExportNewMessengerServiceWithEventPublisher(repo, publisher)
	request := messenger.SendMessageRequest{
		ContentText:       "visible in thread and timeline",
		ThreadID:          root.ID,
		ClientNonce:       "phase151-also-send-canonical-0001",
		AlsoSendToChannel: true,
	}

	response, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.AlsoSendToChannel)
	assert.True(t, repo.messages[response.ID].AlsoSendToChannel)

	timeline, err := repo.ListMessages(ctx, channelID, nil, 50)
	require.NoError(t, err)
	require.Len(t, timeline, 2)
	assert.Contains(t, []string{timeline[0].ID, timeline[1].ID}, response.ID)

	subject := fmt.Sprintf(messenger.SubjectMessageSent, channelID.String())
	messageSent := 0
	for _, publication := range publisher.snapshot() {
		if publication.subject != subject {
			continue
		}
		event, ok := publication.payload.(messenger.MessageSentEvent)
		require.True(t, ok)
		messageSent++
		assert.Equal(t, response.ID, event.Message.ID)
		assert.True(t, event.Message.AlsoSendToChannel)
	}
	assert.Equal(t, 1, messageSent)

	retry, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, retry)
	assert.Equal(t, response.ID, retry.ID)
	assert.True(t, retry.AlsoSendToChannel)

	messageSent = 0
	for _, publication := range publisher.snapshot() {
		if publication.subject == subject {
			messageSent++
		}
	}
	assert.Equal(t, 1, messageSent, "retry must not publish a second message_sent")
}

func TestSendMessageRejectsAlsoSendWithoutThreadBeforePersistence(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{
		ContentText:       "invalid root",
		AlsoSendToChannel: true,
	})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 400)
	assert.Empty(t, repo.messages)
	assert.Zero(t, repo.createMessageCalls)
}

func TestSendMessageNonceRejectsChangedAlsoSendFlag(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := addLiveThreadRoot(repo, channelID, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)
	request := messenger.SendMessageRequest{
		ContentText:       "same logical reply",
		ThreadID:          root.ID,
		ClientNonce:       "phase151-also-send-fingerprint-0001",
		AlsoSendToChannel: true,
	}

	first, err := service.SendMessage(ctx, channelID, senderID, request)
	require.NoError(t, err)
	require.NotNil(t, first)

	request.AlsoSendToChannel = false
	retry, err := service.SendMessage(ctx, channelID, senderID, request)

	require.Error(t, err)
	assert.Nil(t, retry)
	requireAppErrorCode(t, err, 409)
	assert.Equal(t, 1, repo.createMessageCalls)
	assert.Equal(t, 1, root.ReplyCount)
}
