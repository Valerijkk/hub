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
var _ messenger.AuthorizedUnreadScopeResolver = (*phase151GlobalMessageSearchScopeFixture)(nil)
var _ search.AuthorizedMessageSearchExecutor = (*phase151GlobalMessageSearchRecorder)(nil)
var _ search.AuthorizedMessageSearchTaskStarter = (*phase151GlobalMessageSearchTaskStartSpy)(nil)

type phase151GlobalMessageSearchScopeFixture struct {
	scope            messenger.AuthorizedMessageSearchScope
	err              error
	calls            []uuid.UUID
	unreadScopeCalls int
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

func (f *phase151GlobalMessageSearchScopeFixture) ResolveAuthorizedUnreadScope(
	_ context.Context,
	_ uuid.UUID,
) (messenger.AuthorizedUnreadScope, error) {
	f.unreadScopeCalls++
	return messenger.AuthorizedUnreadScope{}, errors.New("global message search must not resolve AuthorizedUnreadScope")
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

type phase151GlobalMessageSearchTaskStartSpy struct {
	startCalls     int
	lateRunCalls   int
	runImmediately bool
	queued         []func()
}

func (s *phase151GlobalMessageSearchTaskStartSpy) StartAuthorizedMessageSearch(task func()) {
	s.startCalls++
	s.queued = append(s.queued, task)
	if s.runImmediately {
		s.RunQueuedTasks()
	}
}

func (s *phase151GlobalMessageSearchTaskStartSpy) RunQueuedTasks() {
	queued := s.queued
	s.queued = nil
	for _, task := range queued {
		s.lateRunCalls++
		task()
	}
}

func TestPhase151GlobalSearchUsesTheSameAuthorizedPublicMemberAndProjectScope(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	publicChannelID := uuid.New()
	memberChannelID := uuid.New()
	activeProjectChannelID := uuid.New()
	scopeResolver := &phase151GlobalMessageSearchScopeFixture{scope: messenger.AuthorizedMessageSearchScope{
		TenantID: tenantID,
		UserID:   userID,
		Targets: []messenger.AuthorizedMessageSearchTarget{
			// The resolver proves public CanRead before returning this pre-join target.
			{ChannelID: publicChannelID, Kind: "channel"},
			{ChannelID: memberChannelID, Kind: "channel"},
			// The resolver returns this target only after fresh active ProjectMember authorization.
			{ChannelID: activeProjectChannelID, Kind: "channel", IsProjectChannel: true},
		},
	}}
	recorder := &phase151GlobalMessageSearchRecorder{}
	starter := &phase151GlobalMessageSearchTaskStartSpy{runImmediately: true}
	service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(scopeResolver, recorder, starter)

	results, err := search.ExportSearchMessages(service, ctx, "authorised", userID)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, []uuid.UUID{userID}, scopeResolver.calls)
	assert.Zero(t, scopeResolver.unreadScopeCalls, "global message search must not derive candidates from AuthorizedUnreadScope")
	assert.Equal(t, 1, starter.startCalls)
	assert.Equal(t, 1, starter.lateRunCalls)
	require.Equal(t, 1, recorder.calls)
	require.Len(t, recorder.scopes, 1)
	assert.Equal(t, scopeResolver.scope, recorder.scopes[0])
	assert.Equal(t, []messenger.AuthorizedMessageSearchTarget{
		{ChannelID: publicChannelID, Kind: "channel"},
		{ChannelID: memberChannelID, Kind: "channel"},
		{ChannelID: activeProjectChannelID, Kind: "channel", IsProjectChannel: true},
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
		"private-nonmember": {
			err: errors.New("private non-member scope denied"),
		},
		"project-ineligible": {
			err: errors.New("project-ineligible scope denied"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			recorder := &phase151GlobalMessageSearchRecorder{}
			starter := &phase151GlobalMessageSearchTaskStartSpy{}
			service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(fixture, recorder, starter)

			results, _ := search.ExportSearchMessages(service, ctx, "must not query Mongo", userID)

			assert.Empty(t, results)
			assert.Equal(t, []uuid.UUID{userID}, fixture.calls)
			assert.Zero(t, fixture.unreadScopeCalls, "global message search must not fall back to AuthorizedUnreadScope")
			assert.Zero(t, starter.startCalls, "empty or denied scope must not start a message-search task")
			assert.Zero(t, recorder.calls, "empty or denied scope must not start a Mongo search or goroutine")
			assert.Empty(t, recorder.scopes)

			starter.RunQueuedTasks()
			assert.Zero(t, starter.lateRunCalls, "denied scope must not leave a message-search task for late execution")
			assert.Zero(t, recorder.calls, "late execution must not reach the message-search executor")
		})
	}
}
func TestPhase151GlobalSearchNilScopeResolverFailsClosedBeforeTaskStartOrExecutor(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	recorder := &phase151GlobalMessageSearchRecorder{}
	starter := &phase151GlobalMessageSearchTaskStartSpy{}
	service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(nil, recorder, starter)

	results, _ := search.ExportSearchMessages(service, ctx, "must not query Mongo", userID)

	assert.Empty(t, results)
	assert.Zero(t, starter.startCalls, "a nil message-search scope resolver must fail closed before task start")
	assert.Zero(t, recorder.calls, "a nil message-search scope resolver must fail closed before executor access")
	starter.RunQueuedTasks()
	assert.Zero(t, starter.lateRunCalls)
	assert.Zero(t, recorder.calls, "late execution must not reach the executor after a nil scope resolver")
}

func TestPhase151GlobalSearchNilExecutorFailsClosedBeforeTaskStart(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	scopeResolver := &phase151GlobalMessageSearchScopeFixture{scope: messenger.AuthorizedMessageSearchScope{
		TenantID: uuid.New(),
		UserID:   userID,
		Targets: []messenger.AuthorizedMessageSearchTarget{
			{ChannelID: uuid.New(), Kind: "channel"},
		},
	}}
	starter := &phase151GlobalMessageSearchTaskStartSpy{}
	service := search.ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(scopeResolver, nil, starter)

	results, _ := search.ExportSearchMessages(service, ctx, "must not query Mongo", userID)

	assert.Empty(t, results)
	assert.Equal(t, []uuid.UUID{userID}, scopeResolver.calls)
	assert.Zero(t, scopeResolver.unreadScopeCalls, "global message search must not fall back to AuthorizedUnreadScope")
	assert.Zero(t, starter.startCalls, "a nil message-search executor must fail closed before task start")
	starter.RunQueuedTasks()
	assert.Zero(t, starter.lateRunCalls)
}