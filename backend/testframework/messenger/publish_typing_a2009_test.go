package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TestPublishTyping_RejectsNonMember_A2_009 — typing только для участников канала.
func TestPublishTyping_RejectsNonMember_A2_009(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)

	owner := uuid.New()
	outsider := uuid.New()
	ch, err := repo.CreateChannel(context.Background(), messenger.CreateChannelInput{
		ChannelType: "public",
		Name:        "general",
		CreatedBy:   owner,
	})
	require.NoError(t, err)
	require.NoError(t, repo.AddMembers(context.Background(), ch.ID, []uuid.UUID{owner}, "owner"))

	err = svc.PublishTyping(context.Background(), ch.ID.String(), outsider)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

// TestPublishTyping_AllowsMember_A2_009 — участник может слать typing.
func TestPublishTyping_AllowsMember_A2_009(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)

	owner := uuid.New()
	ch, err := repo.CreateChannel(context.Background(), messenger.CreateChannelInput{
		ChannelType: "public",
		Name:        "general",
		CreatedBy:   owner,
	})
	require.NoError(t, err)
	require.NoError(t, repo.AddMembers(context.Background(), ch.ID, []uuid.UUID{owner}, "owner"))

	require.NoError(t, svc.PublishTyping(context.Background(), ch.ID.String(), owner))
}
