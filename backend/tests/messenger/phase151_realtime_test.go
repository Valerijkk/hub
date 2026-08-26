package messenger_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151UnreadCoordinatorRepository = (*phase151RealtimeUnreadRecorder)(nil)
var _ messenger.Phase151PendingRecoveryRepository = (*phase151PendingRecoveryRecorder)(nil)
var _ messenger.Phase151UnreadCutoverScopePager = (*phase151UnreadCutoverScopes)(nil)
var _ messenger.Phase151UnreadCutoverRepository = (*phase151UnreadCutoverRecorder)(nil)
var _ messenger.Phase151PostCreatePipelineRepository = (*phase151RealtimePostCreateRecorder)(nil)

type phase151RealtimeUnreadRecorder struct {
	calls      []messenger.UnreadCategory
	candidates map[messenger.UnreadCategory][]messenger.UnreadCandidate
}

func (r *phase151RealtimeUnreadRecorder) ListUnreadCandidates(
	_ context.Context,
	_ messenger.AuthorizedUnreadScope,
	category messenger.UnreadCategory,
) ([]messenger.UnreadCandidate, error) {
	r.calls = append(r.calls, category)
	return append([]messenger.UnreadCandidate(nil), r.candidates[category]...), nil
}

func TestPhase151UnreadResponseKeepsFourNonNilInt64MapsAndServerBoolean(t *testing.T) {
	actorID := uuid.New()
	channelID := uuid.New()
	dmID := uuid.New()
	sharedMessageID := "phase151-overlapping-message"
	recorder := &phase151RealtimeUnreadRecorder{candidates: map[messenger.UnreadCategory][]messenger.UnreadCandidate{
		messenger.UnreadCategoryChannel: {
			{TargetID: channelID.String(), MessageID: sharedMessageID, SenderID: uuid.New()},
			{TargetID: channelID.String(), MessageID: "phase151-own-channel-message", SenderID: actorID},
		},
		messenger.UnreadCategoryDM: {
			{TargetID: dmID.String(), MessageID: "phase151-dm-message", SenderID: uuid.New()},
		},
		messenger.UnreadCategoryMention: {
			{TargetID: channelID.String(), MessageID: sharedMessageID, SenderID: uuid.New()},
		},
		messenger.UnreadCategoryFollowedThread: {
			{TargetID: "phase151-followed-root", MessageID: sharedMessageID, SenderID: uuid.New()},
		},
	}}
	scope := messenger.AuthorizedUnreadScope{
		TenantID: uuid.New(), UserID: actorID,
		Targets: []messenger.AuthorizedUnreadTarget{{ChannelID: channelID, Kind: "channel"}, {ChannelID: dmID, Kind: "dm"}},
	}

	response, err := messenger.ExportPhase151BuildUnreadResponse(context.Background(), scope, recorder)

	require.NoError(t, err)
	require.NotNil(t, response.Channels)
	require.NotNil(t, response.DMs)
	require.NotNil(t, response.Mentions)
	require.NotNil(t, response.Threads)
	assert.Equal(t, int64(1), response.Channels[channelID.String()], "own channel messages must be excluded")
	assert.Equal(t, int64(1), response.DMs[dmID.String()])
	assert.Equal(t, int64(1), response.Mentions[channelID.String()], "one message may overlap channel and mention")
	assert.Equal(t, int64(1), response.Threads["phase151-followed-root"], "one message may overlap followed-thread")
	assert.True(t, response.AnyUnread, "global nav state is a server snapshot boolean, never a category sum")
	assert.Equal(t, []messenger.UnreadCategory{
		messenger.UnreadCategoryChannel,
		messenger.UnreadCategoryDM,
		messenger.UnreadCategoryMention,
		messenger.UnreadCategoryFollowedThread,
	}, recorder.calls)
}

type phase151RealtimePostCreateRecorder struct {
	calls       []string
	canonical   *messenger.MessageDocument
	casClaimed  bool
	publishErr  error
}

func (r *phase151RealtimePostCreateRecorder) FindCanonicalCreateResult(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ string) (messenger.CreateMessageResult, bool, error) {
	if r.canonical == nil {
		return messenger.CreateMessageResult{}, false, nil
	}
	return messenger.CreateMessageResult{Created: false, Message: r.canonical, PostCreateEffectsState: r.canonical.PostCreateEffectsState}, true, nil
}

func (r *phase151RealtimePostCreateRecorder) ResolveEligibleMentionUserIDs(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ messenger.PostCreatePipelineRequest) ([]uuid.UUID, error) {
	r.calls = append(r.calls, "resolve-mentions")
	return nil, nil
}

func (r *phase151RealtimePostCreateRecorder) PersistPendingCanonicalMessage(_ context.Context, _ messenger.AuthorizedMessageProducerScope, document *messenger.MessageDocument) (messenger.CreateMessageResult, error) {
	r.calls = append(r.calls, "persist-pending")
	copyDocument := *document
	if copyDocument.ID == "" {
		copyDocument.ID = "phase151-realtime-" + document.ClientNonce
	}
	r.canonical = &copyDocument
	return messenger.CreateMessageResult{Created: true, Message: r.canonical, PostCreateEffectsState: messenger.PostCreateEffectsStatePending}, nil
}

func (r *phase151RealtimePostCreateRecorder) EnsureAutoThreadFollows(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ *messenger.MessageDocument) error {
	r.calls = append(r.calls, "auto-follow")
	return nil
}

func (r *phase151RealtimePostCreateRecorder) ReconcileReplyCount(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ *messenger.MessageDocument) error {
	r.calls = append(r.calls, "reply-count")
	return nil
}

func (r *phase151RealtimePostCreateRecorder) CompletePostCreateEffects(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ string) (bool, error) {
	r.calls = append(r.calls, "complete-cas")
	return r.casClaimed, nil
}

func (r *phase151RealtimePostCreateRecorder) PublishUnreadInvalidation(_ context.Context, _ messenger.AuthorizedMessageProducerScope, _ *messenger.MessageDocument) error {
	r.calls = append(r.calls, "recipient-invalidation")
	return r.publishErr
}

func TestPhase151RecipientInvalidationIsAfterDurableCASAndRetryHasNoEffects(t *testing.T) {
	scope := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(), Producer: messenger.TrustedMessageProducerUser,
	}
	request := messenger.PostCreatePipelineRequest{
		Producer: messenger.TrustedMessageProducerUser,
		Message: messenger.MessageDocument{
			ChannelID: scope.ChannelID, SenderID: scope.ActorID,
			ContentText: "phase151 recipient invalidation", ClientNonce: "phase151-realtime-cas-before-nats",
		},
	}

	t.Run("publication failure leaves the completed canonical state refetchable", func(t *testing.T) {
		recorder := &phase151RealtimePostCreateRecorder{casClaimed: true, publishErr: errors.New("NATS unavailable")}
		result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

		require.Error(t, err)
		assert.True(t, result.Created)
		assert.Equal(t, []string{"resolve-mentions", "persist-pending", "auto-follow", "complete-cas", "recipient-invalidation"}, recorder.calls)
		assert.NotNil(t, recorder.canonical, "the persisted canonical document remains the refetch source after NATS failure")
	})

	t.Run("CAS miss and created-false retry publish nothing and repeat no effects", func(t *testing.T) {
		canonical := phase151PostCreateCanonicalResult(
			t, scope, request, false, messenger.PostCreateEffectsStatePending,
		).Message
		canonical.ID = "phase151-existing-canonical"
		recorder := &phase151RealtimePostCreateRecorder{canonical: canonical, casClaimed: false}
		result, err := messenger.ExportPhase151ExecutePostCreatePipeline(context.Background(), scope, request, recorder)

		require.NoError(t, err)
		assert.False(t, result.Created)
		assert.Same(t, canonical, result.Message)
		assert.Empty(t, recorder.calls, "Created=false retry must not resolve, mutate, CAS, or publish")
	})
}

type phase151PendingRecoveryRecorder struct {
	calls       []string
	pages       []messenger.Phase151PendingRecoveryTargetPage
	pageCalls   int
	authorize   map[messenger.Phase151PendingRecoveryTarget]bool
	pending     map[messenger.Phase151PendingRecoveryTarget][]messenger.MessageDocument
	intervalOK  bool
	intervals   []time.Duration
}

func (r *phase151PendingRecoveryRecorder) TryAcquirePendingRecoveryInterval(_ context.Context, minimum time.Duration) (bool, error) {
	r.calls = append(r.calls, "interval")
	r.intervals = append(r.intervals, minimum)
	return r.intervalOK, nil
}

func (r *phase151PendingRecoveryRecorder) ListPendingRecoveryTargets(_ context.Context, after *messenger.Phase151RecoveryTargetCursor, limit int) (messenger.Phase151PendingRecoveryTargetPage, error) {
	r.calls = append(r.calls, "pg-target-page")
	if r.pageCalls >= len(r.pages) {
		return messenger.Phase151PendingRecoveryTargetPage{}, nil
	}
	page := r.pages[r.pageCalls]
	r.pageCalls++
	return page, nil
}

func (r *phase151PendingRecoveryRecorder) ReauthorizePendingRecoveryTarget(_ context.Context, target messenger.Phase151PendingRecoveryTarget) (messenger.AuthorizedMessageProducerScope, error) {
	r.calls = append(r.calls, "reauthorize:"+target.ChannelID.String())
	if !r.authorize[target] {
		return messenger.AuthorizedMessageProducerScope{}, errors.New("access revoked")
	}
	return messenger.AuthorizedMessageProducerScope{TenantID: target.TenantID, ActorID: uuid.New(), ChannelID: target.ChannelID, Producer: messenger.TrustedMessageProducerUser}, nil
}

func (r *phase151PendingRecoveryRecorder) ListPendingPostCreateDocuments(_ context.Context, _ messenger.AuthorizedMessageProducerScope, target messenger.Phase151PendingRecoveryTarget, limit int) ([]messenger.MessageDocument, error) {
	r.calls = append(r.calls, "mongo-pending:"+target.ChannelID.String())
	return append([]messenger.MessageDocument(nil), r.pending[target]...), nil
}

func (r *phase151PendingRecoveryRecorder) ReconcilePendingPostCreateEffects(_ context.Context, _ messenger.AuthorizedMessageProducerScope, document messenger.MessageDocument) error {
	r.calls = append(r.calls, "reconcile:"+document.ID)
	return nil
}

func TestPhase151PendingRecoveryUsesStablePGTargetsBeforeBoundedAuthorizedMongo(t *testing.T) {
	allowed := messenger.Phase151PendingRecoveryTarget{TenantID: uuid.New(), ChannelID: uuid.New(), Kind: "channel"}
	denied := messenger.Phase151PendingRecoveryTarget{TenantID: allowed.TenantID, ChannelID: uuid.New(), Kind: "channel"}
	duplicate := allowed
	next := &messenger.Phase151RecoveryTargetCursor{TenantID: denied.TenantID, ChannelID: denied.ChannelID, Kind: denied.Kind}
	recorder := &phase151PendingRecoveryRecorder{
		intervalOK: true,
		pages: []messenger.Phase151PendingRecoveryTargetPage{
			{Targets: []messenger.Phase151PendingRecoveryTarget{allowed, denied}, Next: next},
			{Targets: []messenger.Phase151PendingRecoveryTarget{duplicate}},
		},
		authorize: map[messenger.Phase151PendingRecoveryTarget]bool{allowed: true},
		pending: map[messenger.Phase151PendingRecoveryTarget][]messenger.MessageDocument{
			allowed: {
				{ID: "phase151-pending-effect", PostCreateEffectsState: messenger.PostCreateEffectsStatePending},
				{ID: "phase151-legacy-complete", PostCreateEffectsState: ""},
			},
		},
	}

	err := messenger.ExportPhase151ExecutePendingPostCreateRecovery(context.Background(), recorder, messenger.Phase151PendingRecoveryConfig{
		MaxPages: 2, MaxTargets: 3, PendingScanLimit: 10, MinimumInterval: time.Minute,
	})

	require.NoError(t, err)
	assert.Equal(t, []time.Duration{time.Minute}, recorder.intervals)
	assert.Equal(t, 2, recorder.pageCalls, "stable pagination is globally bounded")
	assert.Equal(t, 1, countPhase151Calls(recorder.calls, "mongo-pending:"+allowed.ChannelID.String()))
	assert.Zero(t, countPhase151Calls(recorder.calls, "mongo-pending:"+denied.ChannelID.String()), "denied target must not reach Mongo")
	assert.Equal(t, 1, countPhase151Calls(recorder.calls, "reconcile:phase151-pending-effect"))
	assert.Zero(t, countPhase151Calls(recorder.calls, "reconcile:phase151-legacy-complete"), "legacy missing effect state is complete")
	assert.Equal(t, 1, countPhase151Calls(recorder.calls, "reauthorize:"+allowed.ChannelID.String()), "duplicate PG targets are unique globally")
}

type phase151UnreadCutoverScopes struct {
	scopes []messenger.AuthorizedUnreadScope
	calls  int
}

func (p *phase151UnreadCutoverScopes) ListFreshAuthorizedUnreadScopes(_ context.Context, after *messenger.Phase151UnreadCutoverCursor, limit int) (messenger.Phase151AuthorizedUnreadScopePage, error) {
	p.calls++
	if len(p.scopes) == 0 {
		return messenger.Phase151AuthorizedUnreadScopePage{}, nil
	}
	scope := p.scopes[0]
	p.scopes = p.scopes[1:]
	return messenger.Phase151AuthorizedUnreadScopePage{Scopes: []messenger.AuthorizedUnreadScope{scope}}, nil
}

type phase151UnreadCutoverRecorder struct {
	calls    []string
	followed map[uuid.UUID][]string
}

func (r *phase151UnreadCutoverRecorder) CurrentAuthorizedTargetHead(_ context.Context, _ messenger.AuthorizedUnreadScope, target messenger.AuthorizedUnreadTarget, category messenger.UnreadCategory) (messenger.MessagePosition, error) {
	r.calls = append(r.calls, "mongo-head:"+string(category)+":"+target.ChannelID.String())
	return messenger.MessagePosition{CreatedAt: time.Unix(1, 0).UTC(), MessageID: string(category) + "-head"}, nil
}

func (r *phase151UnreadCutoverRecorder) ListEligiblePersistedFollowedRoots(_ context.Context, _ messenger.AuthorizedUnreadScope, target messenger.AuthorizedUnreadTarget) ([]string, error) {
	r.calls = append(r.calls, "mongo-follows:"+target.ChannelID.String())
	return append([]string(nil), r.followed[target.ChannelID]...), nil
}

func (r *phase151UnreadCutoverRecorder) CurrentAuthorizedFollowedRootHead(_ context.Context, _ messenger.AuthorizedUnreadScope, target messenger.AuthorizedUnreadTarget, rootMessageID string) (messenger.MessagePosition, error) {
	r.calls = append(r.calls, "mongo-thread-head:"+target.ChannelID.String()+":"+rootMessageID)
	return messenger.MessagePosition{CreatedAt: time.Unix(2, 0).UTC(), MessageID: rootMessageID + "-head"}, nil
}

func (r *phase151UnreadCutoverRecorder) PersistUnreadCutoverBaseline(_ context.Context, _ messenger.AuthorizedUnreadScope, category messenger.UnreadCategory, targetID string, position messenger.MessagePosition) error {
	r.calls = append(r.calls, "persist:"+string(category)+":"+targetID+":"+position.MessageID)
	return nil
}

func TestPhase151UnreadCutoverPagesOnlyFreshScopesAndBaselinesAuthorizedHeads(t *testing.T) {
	t.Run("empty or denied scope performs zero Mongo work", func(t *testing.T) {
		scopes := &phase151UnreadCutoverScopes{scopes: []messenger.AuthorizedUnreadScope{{}}}
		recorder := &phase151UnreadCutoverRecorder{}

		err := messenger.ExportPhase151ExecuteUnreadCutover(context.Background(), scopes, recorder, messenger.Phase151UnreadCutoverConfig{MaxPages: 1, MaxScopes: 1})

		require.NoError(t, err)
		assert.Empty(t, recorder.calls)
	})

	t.Run("scope owns channel DM mention and persisted followed-root baselines", func(t *testing.T) {
		channelID := uuid.New()
		dmID := uuid.New()
		scope := messenger.AuthorizedUnreadScope{TenantID: uuid.New(), UserID: uuid.New(), Targets: []messenger.AuthorizedUnreadTarget{{ChannelID: channelID, Kind: "channel"}, {ChannelID: dmID, Kind: "dm"}}}
		scopes := &phase151UnreadCutoverScopes{scopes: []messenger.AuthorizedUnreadScope{scope}}
		recorder := &phase151UnreadCutoverRecorder{followed: map[uuid.UUID][]string{channelID: {"phase151-followed-root"}}}

		err := messenger.ExportPhase151ExecuteUnreadCutover(context.Background(), scopes, recorder, messenger.Phase151UnreadCutoverConfig{MaxPages: 1, MaxScopes: 1})

		require.NoError(t, err)
		assert.Contains(t, recorder.calls, "persist:channel:"+channelID.String()+":channel-head")
		assert.Contains(t, recorder.calls, "persist:dm:"+dmID.String()+":dm-head")
		assert.Contains(t, recorder.calls, "persist:mention:"+channelID.String()+":mention-head")
		assert.Contains(t, recorder.calls, "persist:followed_thread:phase151-followed-root:phase151-followed-root-head")
		assert.NotContains(t, recorder.calls, "redis", "cutover never discovers or imports Redis counters")
	})
}

func countPhase151Calls(calls []string, want string) int {
	count := 0
	for _, call := range calls {
		if call == want {
			count++
		}
	}
	return count
}
