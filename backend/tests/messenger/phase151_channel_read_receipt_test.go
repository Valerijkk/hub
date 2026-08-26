package messenger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151ReadReceiptRepository struct {
	*mockRepository
	head            messenger.MessagePosition
	headTarget      messenger.AuthorizedUnreadTarget
	headCategory    messenger.UnreadCategory
	advanceCategory messenger.UnreadCategory
	advanceTargetID string
	advancePosition messenger.MessagePosition
	baselineCalls   int
}

func (r *phase151ReadReceiptRepository) CurrentAuthorizedTargetHead(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	target messenger.AuthorizedUnreadTarget,
	category messenger.UnreadCategory,
) (messenger.MessagePosition, error) {
	r.headTarget = target
	r.headCategory = category
	return r.head, nil
}

func (r *phase151ReadReceiptRepository) AdvanceAuthorizedUnreadCursor(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	category messenger.UnreadCategory,
	targetID string,
	position messenger.MessagePosition,
) (bool, error) {
	r.advanceCategory = category
	r.advanceTargetID = targetID
	r.advancePosition = position
	return true, nil
}

func (r *phase151ReadReceiptRepository) PersistUnreadCutoverBaseline(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	_ messenger.UnreadCategory,
	_ string,
	_ messenger.MessagePosition,
) error {
	r.baselineCalls++
	return nil
}

func TestPhase151MarkChannelReadAdvancesDurableAuthorizedHead(t *testing.T) {
	actorID := uuid.New()
	tenantID := uuid.New()
	channelID := uuid.New()
	head := messenger.MessagePosition{CreatedAt: time.Date(2026, 8, 25, 20, 0, 0, 0, time.UTC), MessageID: "message-head"}
	repo := &phase151ReadReceiptRepository{mockRepository: newMockRepository(), head: head}
	unreadScope := &phase151UnreadScopeFixture{scope: messenger.AuthorizedUnreadScope{
		TenantID: tenantID,
		UserID:   actorID,
		Targets: []messenger.AuthorizedUnreadTarget{{
			ChannelID: channelID,
			Kind:      "channel",
		}},
	}}
	service := messenger.NewMessengerServiceWithAuthority(
		repo,
		nil,
		nil,
		&phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)},
		unreadScope,
		&phase151MessageSearchScopeFixture{},
	)

	err := service.MarkChannelRead(context.Background(), actorID.String(), channelID.String())
	require.NoError(t, err)
	assert.Equal(t, messenger.AuthorizedUnreadTarget{ChannelID: channelID, Kind: "channel"}, repo.headTarget)
	assert.Equal(t, messenger.UnreadCategoryChannel, repo.headCategory)
	assert.Equal(t, messenger.UnreadCategoryChannel, repo.advanceCategory)
	assert.Equal(t, channelID.String(), repo.advanceTargetID)
	assert.Equal(t, head, repo.advancePosition)
	assert.Zero(t, repo.baselineCalls)
}
