package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151WSAccessResolver struct {
	access map[uuid.UUID]messenger.ChannelAccess
	errs   map[uuid.UUID]error
	calls  []phase151WSAccessCall
}

type phase151WSAccessCall struct {
	userID    uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

func (r *phase151WSAccessResolver) AuthorizeChannelAccess(
	_ context.Context,
	userID uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	r.calls = append(r.calls, phase151WSAccessCall{userID: userID, channelID: channelID, action: action})
	if err := r.errs[userID]; err != nil {
		return messenger.ChannelAccess{}, err
	}
	return r.access[userID], nil
}

type phase151WSMemberSource struct {
	memberIDs []uuid.UUID
	err       error
	calls     int
}

func (s *phase151WSMemberSource) GetChannelMemberIDs(
	_ context.Context,
	_ uuid.UUID,
) ([]uuid.UUID, error) {
	s.calls++
	return append([]uuid.UUID(nil), s.memberIDs...), s.err
}

func phase151WSAllowedAccess(tenantID, channelID uuid.UUID) messenger.ChannelAccess {
	return messenger.ChannelAccess{
		TenantID:        tenantID,
		Channel:         &ent.Channel{ID: channelID},
		IsMember:        true,
		CanReceiveEvent: true,
	}
}

func TestPhase151WSChannelAuthorityRequiresDependencies(t *testing.T) {
	resolver := &phase151WSAccessResolver{}
	source := &phase151WSMemberSource{}

	assert.Panics(t, func() { messenger.NewWSChannelAuthority(nil, source) })
	assert.Panics(t, func() { messenger.NewWSChannelAuthority(resolver, nil) })
}

func TestPhase151WSChannelAuthorityUsesFreshReceiveEventProof(t *testing.T) {
	tenantID := uuid.New()
	channelID := uuid.New()
	userID := uuid.New()
	resolver := &phase151WSAccessResolver{
		access: map[uuid.UUID]messenger.ChannelAccess{userID: phase151WSAllowedAccess(tenantID, channelID)},
		errs:   map[uuid.UUID]error{},
	}
	authority := messenger.NewWSChannelAuthority(resolver, &phase151WSMemberSource{})

	allowed, err := authority.AuthorizeChannelRoom(context.Background(), userID, channelID)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = authority.AuthorizeUserChannelEvent(context.Background(), userID, channelID, "typing")
	require.NoError(t, err)
	assert.True(t, allowed)
	require.Len(t, resolver.calls, 2)
	for _, call := range resolver.calls {
		assert.Equal(t, userID, call.userID)
		assert.Equal(t, channelID, call.channelID)
		assert.Equal(t, messenger.ChannelActionReceiveEvent, call.action)
	}
}

func TestPhase151WSChannelAudienceOmitsRevokedDeniedAndDuplicateCandidates(t *testing.T) {
	tenantID := uuid.New()
	channelID := uuid.New()
	allowedID := uuid.New()
	revokedID := uuid.New()
	errorID := uuid.New()
	resolver := &phase151WSAccessResolver{
		access: map[uuid.UUID]messenger.ChannelAccess{
			allowedID: phase151WSAllowedAccess(tenantID, channelID),
			revokedID: {TenantID: tenantID, Channel: &ent.Channel{ID: channelID}},
		},
		errs: map[uuid.UUID]error{errorID: errors.New("project membership revoked")},
	}
	source := &phase151WSMemberSource{memberIDs: []uuid.UUID{
		allowedID,
		revokedID,
		errorID,
		uuid.Nil,
		allowedID,
	}}
	authority := messenger.NewWSChannelAuthority(resolver, source)

	recipients, err := authority.ResolveChannelEventAudience(context.Background(), channelID, "message_sent")
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{allowedID}, recipients)
	assert.Equal(t, 1, source.calls)
}

func TestPhase151WSChannelAudienceFailsClosedOnCandidateLookupError(t *testing.T) {
	lookupErr := errors.New("candidate lookup failed")
	authority := messenger.NewWSChannelAuthority(
		&phase151WSAccessResolver{},
		&phase151WSMemberSource{err: lookupErr},
	)

	recipients, err := authority.ResolveChannelEventAudience(context.Background(), uuid.New(), "message_sent")
	require.ErrorIs(t, err, lookupErr)
	assert.Nil(t, recipients)
}
