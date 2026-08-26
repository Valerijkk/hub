package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type phase151PinsNoLeakRepository struct {
	*mockRepository
	pins       []*ent.PinnedMessage
	operations []string
}

func (r *phase151PinsNoLeakRepository) ListPins(
	_ context.Context,
	channelID uuid.UUID,
) ([]*ent.PinnedMessage, error) {
	r.operations = append(r.operations, "pg:list_pins:"+channelID.String())
	return append([]*ent.PinnedMessage(nil), r.pins...), nil
}

func TestPhase151ListPinsOmitsForeignChannelRowsAfterFreshProof(t *testing.T) {
	ctx := context.Background()
	base := newMockRepository()
	ownerID := uuid.New()
	viewerID := uuid.New()
	channelID := setupChannelWithMembers(base, "private", false, ownerID, viewerID)
	foreignChannelID := setupChannelWithMembers(base, "private", false, ownerID)
	channel := base.channels[channelID]
	now := time.Now().UTC()
	matching := &ent.PinnedMessage{
		ID: uuid.New(), ChannelID: channelID, MessageID: "visible-pin",
		PinnedBy: ownerID, PinnedAt: now,
	}
	foreign := &ent.PinnedMessage{
		ID: uuid.New(), ChannelID: foreignChannelID, MessageID: "foreign-secret-pin",
		PinnedBy: uuid.New(), PinnedAt: now.Add(-time.Minute),
	}
	repo := &phase151PinsNoLeakRepository{
		mockRepository: base,
		pins:           []*ent.PinnedMessage{matching, foreign},
	}
	resolver := &phase151ChannelAccessResolver{
		grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess),
		onAuthorize: func(key phase151ChannelAccessKey) {
			repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
		},
	}
	access := phase151ReadableAccess(channel, true)
	access.CanListPins = true
	phase151Grant(resolver, viewerID, channel, messenger.ChannelActionListPins, access)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	pins, err := service.ListPins(ctx, channelID, viewerID)

	require.NoError(t, err)
	require.Len(t, pins, 1)
	assert.Equal(t, matching.ID.String(), pins[0].ID)
	assert.Equal(t, matching.MessageID, pins[0].MessageID)
	assert.Equal(t, channelID.String(), pins[0].ChannelID)
	assert.NotEqual(t, foreign.ID.String(), pins[0].ID)
	assert.NotEqual(t, foreign.MessageID, pins[0].MessageID)
	assert.NotEqual(t, foreignChannelID.String(), pins[0].ChannelID)
	assert.Equal(t, []string{
		"pg:authorize:" + channelID.String(),
		"pg:list_pins:" + channelID.String(),
	}, repo.operations)
}

type phase151PinFindErrorRepository struct {
	*mockRepository
	findResult *ent.PinnedMessage
	findErr    error
	findCalls   int
	addCalls    int
	removeCalls int
}

func (r *phase151PinFindErrorRepository) FindPin(_ context.Context, _ uuid.UUID, _ string) (*ent.PinnedMessage, error) {
	r.findCalls++
	return r.findResult, r.findErr
}

func (r *phase151PinFindErrorRepository) AddPin(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	r.addCalls++
	return nil
}

func (r *phase151PinFindErrorRepository) RemovePin(_ context.Context, _ uuid.UUID, _ string) error {
	r.removeCalls++
	return nil
}

func TestPhase151TogglePinFailsClosedOnUnexpectedFindPinError(t *testing.T) {
	sentinelErr := errors.New("pin lookup unavailable")
	testCases := []struct {
		name       string
		findResult *ent.PinnedMessage
	}{
		{name: "nil result", findResult: nil},
		{name: "entity with error", findResult: &ent.PinnedMessage{ID: uuid.New()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			base := newMockRepository()
			ownerID := uuid.New()
			actorID := uuid.New()
			channelID := setupChannelWithMembers(base, "private", false, ownerID, actorID)
			channel := base.channels[channelID]
			messageID := uuid.NewString()
			base.messages[messageID] = &messenger.MessageDocument{
				ID: messageID, ChannelID: channelID, SenderID: ownerID,
			}
			repo := &phase151PinFindErrorRepository{
				mockRepository: base,
				findResult:     tc.findResult,
				findErr:        sentinelErr,
			}
			resolver := &phase151ChannelAccessResolver{
				grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess),
			}
			access := phase151ReadableAccess(channel, true)
			access.CanPin = true
			phase151Grant(resolver, actorID, channel, messenger.ChannelActionPin, access)
			service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

			pinned, err := service.TogglePin(ctx, channelID, messageID, actorID)

			assert.False(t, pinned)
			assert.Equal(t, 1, repo.findCalls)
			assert.Zero(t, repo.addCalls)
			assert.Zero(t, repo.removeCalls)
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 500, appErr.Code)
			assert.Equal(t, "ошибка поиска закреплённого сообщения", appErr.Message)
			assert.NotContains(t, err.Error(), sentinelErr.Error())
			assert.Equal(t, []phase151ChannelAccessKey{{
				actorID: actorID, channelID: channelID, action: messenger.ChannelActionPin,
			}}, resolver.calls)
		})
	}
}
