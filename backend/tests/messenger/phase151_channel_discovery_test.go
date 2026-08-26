package messenger_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/internal/messenger"
)

const phase151DiscoveryLimit = 100

type phase151DiscoveryCall struct {
	actorID uuid.UUID
	limit   int
}

type phase151DiscoveryRepository struct {
	*mockRepository
	candidates []*ent.Channel
	err        error
	calls      []phase151DiscoveryCall
}

func (r *phase151DiscoveryRepository) ListDiscoverableChannels(
	_ context.Context,
	actorID uuid.UUID,
	limit int,
) ([]*ent.Channel, error) {
	r.calls = append(r.calls, phase151DiscoveryCall{actorID: actorID, limit: limit})
	if r.err != nil {
		return nil, r.err
	}
	return append([]*ent.Channel(nil), r.candidates...), nil
}

type phase151DiscoveryAccessCall struct {
	actorID   uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

type phase151DiscoveryAccessResolver struct {
	accessByChannel map[uuid.UUID]messenger.ChannelAccess
	errByChannel    map[uuid.UUID]error
	calls           []phase151DiscoveryAccessCall
}

func (r *phase151DiscoveryAccessResolver) AuthorizeChannelAccess(
	_ context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	r.calls = append(r.calls, phase151DiscoveryAccessCall{
		actorID: actorID, channelID: channelID, action: action,
	})
	if err := r.errByChannel[channelID]; err != nil {
		return messenger.ChannelAccess{}, err
	}
	access, ok := r.accessByChannel[channelID]
	if !ok {
		return messenger.ChannelAccess{}, messenger.ErrChannelAccessDenied
	}
	return access, nil
}

func phase151DiscoveryChannel(id uuid.UUID, channelType, name string, projectID *uuid.UUID) *ent.Channel {
	now := time.Now().UTC()
	return &ent.Channel{
		ID: id, ChannelType: channel.ChannelType(channelType), Name: name,
		ProjectID: projectID, CreatedBy: uuid.New(), CreatedAt: now, UpdatedAt: now,
	}
}

func TestPhase151DiscoverChannelsUsesFreshRecipientProjectionAndOmitsDeniedKinds(t *testing.T) {
	actorID := uuid.New()
	projectID := uuid.New()
	globalID, projectChannelID := uuid.New(), uuid.New()
	privateID, groupID, dmID, deniedProjectID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	staleGlobal := phase151DiscoveryChannel(globalID, "public", "stale global", nil)
	staleProject := phase151DiscoveryChannel(projectChannelID, "public", "stale project", &projectID)
	privateCandidate := phase151DiscoveryChannel(privateID, "private", "private leak", nil)
	groupCandidate := phase151DiscoveryChannel(groupID, "group", "group leak", nil)
	dmCandidate := phase151DiscoveryChannel(dmID, "dm", "dm leak", nil)
	deniedProject := phase151DiscoveryChannel(deniedProjectID, "public", "former project", &projectID)

	freshGlobal := phase151DiscoveryChannel(globalID, "public", "fresh global", nil)
	freshProject := phase151DiscoveryChannel(projectChannelID, "public", "fresh project", &projectID)
	freshPrivate := phase151DiscoveryChannel(privateID, "private", "fresh private", nil)
	freshGroup := phase151DiscoveryChannel(groupID, "group", "fresh group", nil)
	freshDM := phase151DiscoveryChannel(dmID, "dm", "fresh dm", nil)

	base := newMockRepository()
	repo := &phase151DiscoveryRepository{
		mockRepository: base,
		candidates: []*ent.Channel{
			staleGlobal, staleProject, privateCandidate, groupCandidate, dmCandidate, deniedProject,
		},
	}
	resolver := &phase151DiscoveryAccessResolver{
		accessByChannel: map[uuid.UUID]messenger.ChannelAccess{
			globalID: {
				TenantID: uuid.New(), Channel: freshGlobal,
				CanRead: true, CanJoin: true,
			},
			projectChannelID: {
				TenantID: uuid.New(), Channel: freshProject,
				HasActiveProjectMembership: true, CanRead: true, CanJoin: true,
			},
			privateID: {
				TenantID: uuid.New(), Channel: freshPrivate, IsMember: true,
				CanRead: true, CanPost: true, CanReply: true,
			},
			groupID: {
				TenantID: uuid.New(), Channel: freshGroup, IsMember: true, CanRead: true,
			},
			dmID: {
				TenantID: uuid.New(), Channel: freshDM, IsMember: true, CanRead: true,
			},
		},
		errByChannel: map[uuid.UUID]error{deniedProjectID: messenger.ErrChannelAccessDenied},
	}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	channels, err := service.DiscoverChannels(context.Background(), actorID)

	require.NoError(t, err)
	require.Len(t, channels, 2)
	assert.Equal(t, []string{"fresh global", "fresh project"}, []string{channels[0].Name, channels[1].Name})
	assert.Equal(t, phase151DiscoveryCall{actorID: actorID, limit: phase151DiscoveryLimit}, repo.calls[0])
	assert.Zero(t, base.addMembersCalls, "read-only discovery must not mutate membership")

	require.NotNil(t, channels[0].Capabilities)
	assert.False(t, channels[0].Capabilities.IsMember)
	assert.False(t, channels[0].Capabilities.HasActiveProjectMembership)
	assert.True(t, channels[0].Capabilities.CanRead)
	assert.True(t, channels[0].Capabilities.CanJoin)
	assert.False(t, channels[0].Capabilities.CanPost)
	assert.False(t, channels[0].Capabilities.CanReply)
	assert.False(t, channels[0].Capabilities.CanFollow)
	assert.False(t, channels[0].Capabilities.CanViewUnread)

	require.NotNil(t, channels[1].Capabilities)
	assert.False(t, channels[1].Capabilities.IsMember)
	assert.True(t, channels[1].Capabilities.HasActiveProjectMembership)
	assert.True(t, channels[1].Capabilities.CanRead)
	assert.True(t, channels[1].Capabilities.CanJoin)

	for _, call := range resolver.calls {
		assert.Equal(t, actorID, call.actorID)
		assert.Equal(t, messenger.ChannelActionReadTimeline, call.action)
	}
}

func TestPhase151DiscoverChannelsFailsClosedOnRepositoryOrAuthorityFailure(t *testing.T) {
	actorID := uuid.New()
	channelID := uuid.New()
	candidate := phase151DiscoveryChannel(channelID, "public", "candidate", nil)

	t.Run("repository failure", func(t *testing.T) {
		repo := &phase151DiscoveryRepository{
			mockRepository: newMockRepository(), err: errors.New("postgres unavailable"),
		}
		service := messenger.ExportNewMessengerServiceWithChannelAccessTest(
			repo,
			&phase151DiscoveryAccessResolver{},
		)

		channels, err := service.DiscoverChannels(context.Background(), actorID)

		require.Error(t, err)
		assert.Nil(t, channels)
	})

	t.Run("unexpected fresh authority failure", func(t *testing.T) {
		repo := &phase151DiscoveryRepository{
			mockRepository: newMockRepository(), candidates: []*ent.Channel{candidate},
		}
		resolver := &phase151DiscoveryAccessResolver{
			errByChannel: map[uuid.UUID]error{channelID: errors.New("authority unavailable")},
		}
		service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

		channels, err := service.DiscoverChannels(context.Background(), actorID)

		require.Error(t, err)
		assert.Nil(t, channels)
	})
}

func TestPhase151DiscoverChannelsRoutePrecedesChannelIDRoute(t *testing.T) {
	actorID := uuid.New()
	channelID := uuid.New()
	fresh := phase151DiscoveryChannel(channelID, "public", "browse", nil)
	repo := &phase151DiscoveryRepository{
		mockRepository: newMockRepository(), candidates: []*ent.Channel{fresh},
	}
	resolver := &phase151DiscoveryAccessResolver{
		accessByChannel: map[uuid.UUID]messenger.ChannelAccess{
			channelID: {TenantID: uuid.New(), Channel: fresh, CanRead: true, CanJoin: true},
		},
	}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	app := phase151MessengerRouteApp(service, actorID)

	response, err := app.Test(httptest.NewRequest(
		http.MethodGet,
		"/api/messenger/channels/discover",
		http.NoBody,
	))

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
