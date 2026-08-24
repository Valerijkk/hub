package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

func TestSendMessageRejectsForeignThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	foreignChannelID := setupChannelWithMembers(repo, "public", false, senderID)
	parent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: foreignChannelID, SenderID: senderID, ContentText: "foreign root"}
	repo.messages[parent.ID] = parent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: parent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, parent.ReplyCount)
}

func TestSendMessageRejectsMissingThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	untouchedParent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "untouched root"}
	repo.messages[untouchedParent.ID] = untouchedParent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: uuid.NewString()})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, untouchedParent.ReplyCount)
}

func TestSendMessageRejectsDeletedThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	repo.findMessageIncludeDeleted = true
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	deletedAt := time.Now().UTC()
	parent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "deleted root", DeletedAt: &deletedAt}
	repo.messages[parent.ID] = parent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: parent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 404)
	assert.Len(t, repo.messages, 1)
	assert.Zero(t, parent.ReplyCount)
}

func TestSendMessageRejectsNestedThreadParent(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	root := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "root"}
	nestedParent := &messenger.MessageDocument{ID: uuid.NewString(), ChannelID: channelID, SenderID: senderID, ContentText: "nested", ThreadID: root.ID}
	repo.messages[root.ID] = root
	repo.messages[nestedParent.ID] = nestedParent
	service := messenger.ExportNewMessengerServiceTest(repo)

	response, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: nestedParent.ID})

	require.Error(t, err)
	assert.Nil(t, response)
	requireAppErrorCode(t, err, 400)
	assert.Len(t, repo.messages, 2)
	assert.Zero(t, nestedParent.ReplyCount)
}

func TestSendMessageAcceptsLiveRootInSameChannel(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()
	senderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, senderID)
	service := messenger.ExportNewMessengerServiceTest(repo)

	root, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "root"})
	require.NoError(t, err)
	assert.Empty(t, root.ThreadID)

	reply, err := service.SendMessage(ctx, channelID, senderID, messenger.SendMessageRequest{ContentText: "reply", ThreadID: root.ID})
	require.NoError(t, err)
	assert.Equal(t, root.ID, reply.ThreadID)
	require.Len(t, repo.messages, 2)
	assert.Equal(t, 1, repo.messages[root.ID].ReplyCount)

	replies := 0
	for _, document := range repo.messages {
		if document.ThreadID == root.ID {
			replies++
		}
	}
	assert.Equal(t, 1, replies)
}

func requireAppErrorCode(t *testing.T, err error, expectedCode int) {
	t.Helper()
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok, "expected AppError, got %T", err)
	assert.Equal(t, expectedCode, appErr.Code)
}
