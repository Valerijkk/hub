package search_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
	"github.com/valerii/hub/backend/internal/search"
)

var _ messenger.AuthorizedMessageSearchScopeResolver = (*phase151GlobalMessageSearchScopeFixture)(nil)
var _ search.AuthorizedMessageSearchExecutor = (*phase151GlobalMessageSearchRecorder)(nil)

type phase151GlobalMessageSearchScopeFixture struct {
	scope messenger.AuthorizedMessageSearchScope
	err   error
	calls []uuid.UUID
}

func (f *phase151GlobalMessageSearchScopeFixture) ResolveAuthorizedMessageSearchScope(
	_ context.Context,
	userID uuid.UUID,
) (messenger.AuthorizedMessageSearchScope, error) {
	f.calls = append(f.calls, userID)
	if f.err != nil {
		return messenger.AuthorizedMessageSearchScope{}, f.err
	}
	return f.scope, nil
}

type phase151GlobalMessageSearchRecorder struct {
	calls  int
	scopes []messenger.AuthorizedMessageSearchScope
}

func (r *phase151GlobalMessageSearchRecorder) SearchAuthorizedMessages(
	_ context.Context,
	scope messenger.AuthorizedMessageSearchScope,
	_ string,
	_ int,
) ([]search.SearchResult, error) {
	r.calls++
	r.scopes = append(r.scopes, scope)
	return nil, nil
}

func TestPhase151GlobalSearchUsesTheSameAuthorizedPublicMemberAndProjectScope(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	publicChannelID := uuid.New()
	memberChannelID := uuid.New()
	projectChannelID := uuid.New()
	scopeResolver := &phase151GlobalMessageSearchScopeFixture{scope: messenger.AuthorizedMessageSearchScope{
		TenantID: tenantID,
		UserID:   userID,
		Targets: []messenger.AuthorizedMessageSearchTarget{
			// The resolver proves public CanRead before returning this pre-join target.
			{ChannelID: publicChannelID, Kind: "channel"},
			{ChannelID: memberChannelID, Kind: "channel"},
			// An inactive ProjectMember must keep this target out of the scope.
			{ChannelID: projectChannelID, Kind: "channel", IsProjectChannel: true},
		},
	}}
	recorder := &phase151GlobalMessageSearchRecorder{}
	service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(scopeResolver, recorder)

	results, err := search.ExportSearchMessages(service, ctx, "authorised", userID)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, []uuid.UUID{userID}, scopeResolver.calls)
	require.Equal(t, 1, recorder.calls)
	require.Len(t, recorder.scopes, 1)
	assert.Equal(t, scopeResolver.scope, recorder.scopes[0])
	assert.Equal(t, []messenger.AuthorizedMessageSearchTarget{
		{ChannelID: publicChannelID, Kind: "channel"},
		{ChannelID: memberChannelID, Kind: "channel"},
		{ChannelID: projectChannelID, Kind: "channel", IsProjectChannel: true},
	}, recorder.scopes[0].Targets)
}

func TestPhase151GlobalSearchEmptyOrDeniedScopeNeverStartsMongoSearch(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	for name, fixture := range map[string]*phase151GlobalMessageSearchScopeFixture{
		"empty": {
			scope: messenger.AuthorizedMessageSearchScope{TenantID: uuid.New(), UserID: userID},
		},
		"denied": {
			err: errors.New("scope denied"),
		},
		"revoked": {
			err: errors.New("membership revoked"),
		},
		"disabled": {
			err: errors.New("account disabled"),
		},
		"cross-tenant": {
			err: errors.New("cross-tenant scope denied"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			recorder := &phase151GlobalMessageSearchRecorder{}
			service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(fixture, recorder)

			results, _ := search.ExportSearchMessages(service, ctx, "must not query Mongo", userID)

			assert.Empty(t, results)
			assert.Equal(t, []uuid.UUID{userID}, fixture.calls)
			assert.Zero(t, recorder.calls, "empty or denied scope must not start a Mongo search or goroutine")
			assert.Empty(t, recorder.scopes)
		})
	}
}
