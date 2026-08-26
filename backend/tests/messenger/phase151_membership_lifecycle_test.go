package messenger_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.Phase151MembershipLifecycleRepository = (*phase151MembershipLifecycleRecorder)(nil)

// phase151MembershipLifecycleRecorder records the PG/Mongo/event boundary of
// the planned coordinator. It is intentionally not a local grant algorithm.
type phase151MembershipLifecycleRecorder struct {
	calls              []string
	head               messenger.MessagePosition
	authorizationScope *messenger.AuthorizedGrantBaselineScope
	revocationFence    *messenger.MembershipLifecycleFence
	alreadyMember      bool
	commitCreated      bool
	finalRecheckFails  bool
	cleanupErr         error
	publishRevokeErr   error
	accessDenied       bool
	membership         bool
	staleReadState     bool
	autoFollow         bool
	manualFollow       bool
	manualUnfollow     bool
	baselines          map[messenger.MembershipBaselineKind]messenger.MessagePosition
	events             int
	revocationEvents   int
	compensations      int
}

func newPhase151MembershipLifecycleRecorder() *phase151MembershipLifecycleRecorder {
	return &phase151MembershipLifecycleRecorder{
		head: messenger.MessagePosition{
			CreatedAt: time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC),
			MessageID: "head-010",
		},
		commitCreated: true,
		baselines:     make(map[messenger.MembershipBaselineKind]messenger.MessagePosition),
	}
}

func phase151AuthorizedGrantScope(request messenger.MembershipGrantRequest) messenger.AuthorizedGrantBaselineScope {
	return messenger.AuthorizedGrantBaselineScope{
		TenantID:            request.TenantID,
		ActorID:             request.UserID,
		TargetID:            request.TargetID,
		ChannelID:           request.TargetID,
		Kind:                string(request.TargetKind),
		TargetKind:          request.TargetKind,
		LifecycleGeneration: 1,
	}
}

func (r *phase151MembershipLifecycleRecorder) AuthorizeMembershipGrant(
	_ context.Context,
	request messenger.MembershipGrantRequest,
) (messenger.MembershipGrantAuthorization, error) {
	r.calls = append(r.calls, "pg:authorize:"+string(request.EntryPoint))
	if r.accessDenied {
		return messenger.MembershipGrantAuthorization{}, errors.New("phase151 membership access denied")
	}
	scope := phase151AuthorizedGrantScope(request)
	if r.authorizationScope != nil {
		scope = *r.authorizationScope
	}
	return messenger.MembershipGrantAuthorization{
		AlreadyMember: r.alreadyMember,
		Scope:         scope,
	}, nil
}

func (r *phase151MembershipLifecycleRecorder) ResetStaleMembershipState(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
) error {
	r.calls = append(r.calls, "mongo:reset")
	r.staleReadState = false
	r.autoFollow = false
	r.manualFollow = false
	return nil
}

func (r *phase151MembershipLifecycleRecorder) CaptureAuthorizedMembershipHead(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
) (messenger.MessagePosition, error) {
	r.calls = append(r.calls, "mongo:head")
	return r.head, nil
}

func (r *phase151MembershipLifecycleRecorder) UpsertMembershipBaseline(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
	kind messenger.MembershipBaselineKind,
	position messenger.MessagePosition,
) error {
	r.calls = append(r.calls, "mongo:baseline:"+string(kind))
	r.baselines[kind] = position
	return nil
}

func (r *phase151MembershipLifecycleRecorder) CommitMembershipGrant(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
) (messenger.MembershipGrantCommitResult, error) {
	r.calls = append(r.calls, "pg:final-recheck-and-commit")
	if r.finalRecheckFails {
		return messenger.MembershipGrantCommitResult{}, errors.New("phase151 final grant recheck failed")
	}
	r.membership = r.commitCreated
	return messenger.MembershipGrantCommitResult{Created: r.commitCreated}, nil
}

func (r *phase151MembershipLifecycleRecorder) CompensateUncommittedMembershipGrant(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
) error {
	r.calls = append(r.calls, "mongo:compensate")
	r.compensations++
	return nil
}

func (r *phase151MembershipLifecycleRecorder) PublishMembershipGranted(
	_ context.Context,
	_ messenger.AuthorizedGrantBaselineScope,
) error {
	r.calls = append(r.calls, "event:grant")
	r.events++
	return nil
}

func (r *phase151MembershipLifecycleRecorder) CommitMembershipRevocation(
	_ context.Context,
	request messenger.MembershipRevocationRequest,
) (messenger.MembershipLifecycleFence, error) {
	r.calls = append(r.calls, "pg:revoke:"+string(request.Reason))
	r.accessDenied = true
	r.membership = false
	if r.revocationFence != nil {
		return *r.revocationFence, nil
	}
	generation := request.LifecycleGeneration
	if generation <= 0 {
		generation = 1
	}
	return messenger.MembershipLifecycleFence{
		TenantID:            request.TenantID,
		ActorID:             request.UserID,
		TargetID:            request.TargetID,
		ChannelID:           request.TargetID,
		TargetKind:          request.TargetKind,
		LifecycleGeneration: generation,
	}, nil
}

func (r *phase151MembershipLifecycleRecorder) CleanupRevokedMembershipState(
	_ context.Context,
	_ messenger.MembershipLifecycleFence,
) error {
	r.calls = append(r.calls, "mongo:cleanup")
	if r.cleanupErr != nil {
		return r.cleanupErr
	}
	r.staleReadState = false
	r.autoFollow = false
	r.manualFollow = false
	return nil
}

func (r *phase151MembershipLifecycleRecorder) PublishMembershipRevoked(
	_ context.Context,
	_ messenger.MembershipLifecycleFence,
) error {
	r.calls = append(r.calls, "event:revoke")
	r.revocationEvents++
	return r.publishRevokeErr
}
func TestPhase151AllMembershipGrantEntrypointsShareAuthorizedBaselineCoordinator(t *testing.T) {
	for _, tt := range []struct {
		name       string
		entryPoint messenger.MembershipGrantEntryPoint
		targetKind messenger.MembershipTargetKind
	}{
		{name: "create channel owner", entryPoint: messenger.MembershipGrantEntryPointCreateChannelOwner, targetKind: messenger.MembershipTargetKindChannel},
		{name: "create channel initial member", entryPoint: messenger.MembershipGrantEntryPointCreateChannelInitialMember, targetKind: messenger.MembershipTargetKindChannel},
		{name: "join channel", entryPoint: messenger.MembershipGrantEntryPointJoinChannel, targetKind: messenger.MembershipTargetKindChannel},
		{name: "add members", entryPoint: messenger.MembershipGrantEntryPointAddMembers, targetKind: messenger.MembershipTargetKindChannel},
		{name: "add members internal", entryPoint: messenger.MembershipGrantEntryPointAddMembersInternal, targetKind: messenger.MembershipTargetKindChannel},
		{name: "accept invite", entryPoint: messenger.MembershipGrantEntryPointAcceptInvite, targetKind: messenger.MembershipTargetKindChannel},
		{name: "create or get DM", entryPoint: messenger.MembershipGrantEntryPointCreateOrGetDM, targetKind: messenger.MembershipTargetKindDM},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := newPhase151MembershipLifecycleRecorder()
			result, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
				TenantID:   uuid.New(),
				UserID:     uuid.New(),
				TargetID:   uuid.New(),
				TargetKind: tt.targetKind,
				EntryPoint: tt.entryPoint,
			}, recorder)

			require.NoError(t, err)
			assert.True(t, result.Created)
			assert.True(t, recorder.membership)
			assert.Equal(t, recorder.head, recorder.baselines[messenger.MembershipBaselineChannelOrDM])
			assert.Equal(t, recorder.head, recorder.baselines[messenger.MembershipBaselineMention])
			assert.Equal(t, []string{
				"pg:authorize:" + string(tt.entryPoint),
				"mongo:reset",
				"mongo:head",
				"mongo:baseline:" + string(messenger.MembershipBaselineChannelOrDM),
				"mongo:baseline:" + string(messenger.MembershipBaselineMention),
				"pg:final-recheck-and-commit",
				"event:grant",
			}, recorder.calls)
		})
	}
}

func TestPhase151MembershipGrantRaceRetryAndRejoinBoundaries(t *testing.T) {
	t.Run("message after captured head remains unread eligible", func(t *testing.T) {
		recorder := newPhase151MembershipLifecycleRecorder()
		result, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
			TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
			TargetKind: messenger.MembershipTargetKindChannel,
			EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
		}, recorder)

		require.NoError(t, err)
		assert.True(t, result.Created)
		baseline := recorder.baselines[messenger.MembershipBaselineChannelOrDM]
		assert.Equal(t, "head-010", baseline.MessageID)
		candidate := messenger.MessagePosition{
			CreatedAt: recorder.head.CreatedAt.Add(time.Nanosecond), MessageID: "head-011",
		}
		assert.True(t, messenger.ExportPhase151MessageIsAfterBaseline(baseline, candidate),
			"a message persisted after the captured head remains unread eligible")
		assert.False(t, messenger.ExportPhase151MessageIsAfterBaseline(baseline, baseline),
			"the captured baseline itself is not unread")
	})

	t.Run("created false retry skips reset baseline and publication", func(t *testing.T) {
		recorder := newPhase151MembershipLifecycleRecorder()
		recorder.alreadyMember = true
		result, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
			TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
			TargetKind: messenger.MembershipTargetKindChannel,
			EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
		}, recorder)

		require.NoError(t, err)
		assert.False(t, result.Created)
		assert.Empty(t, recorder.baselines)
		assert.Zero(t, recorder.events)
		assert.Equal(t, []string{"pg:authorize:" + string(messenger.MembershipGrantEntryPointJoinChannel)}, recorder.calls)
	})

	t.Run("final commit created false compensates pending baseline and skips publication", func(t *testing.T) {
		recorder := newPhase151MembershipLifecycleRecorder()
		recorder.commitCreated = false
		result, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
			TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
			TargetKind: messenger.MembershipTargetKindChannel,
			EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
		}, recorder)

		require.NoError(t, err)
		assert.False(t, result.Created)
		assert.Equal(t, 1, recorder.compensations)
		assert.Zero(t, recorder.events)
		assert.Equal(t, []string{
			"pg:authorize:" + string(messenger.MembershipGrantEntryPointJoinChannel),
			"mongo:reset",
			"mongo:head",
			"mongo:baseline:" + string(messenger.MembershipBaselineChannelOrDM),
			"mongo:baseline:" + string(messenger.MembershipBaselineMention),
			"pg:final-recheck-and-commit",
			"mongo:compensate",
		}, recorder.calls)
	})
	t.Run("rejoin resets access bearing state but retains manual unfollow", func(t *testing.T) {
		recorder := newPhase151MembershipLifecycleRecorder()
		recorder.staleReadState = true
		recorder.autoFollow = true
		recorder.manualFollow = true
		recorder.manualUnfollow = true
		_, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
			TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
			TargetKind: messenger.MembershipTargetKindChannel,
			EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
		}, recorder)

		require.NoError(t, err)
		assert.False(t, recorder.staleReadState)
		assert.False(t, recorder.autoFollow)
		assert.False(t, recorder.manualFollow)
		assert.True(t, recorder.manualUnfollow)
	})
}

func TestPhase151FailedFinalGrantRecheckNeverCommitsOrPublishesAndCompensatesIdempotently(t *testing.T) {
	recorder := newPhase151MembershipLifecycleRecorder()
	recorder.finalRecheckFails = true
	request := messenger.MembershipGrantRequest{
		TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
		TargetKind: messenger.MembershipTargetKindChannel,
		EntryPoint: messenger.MembershipGrantEntryPointAddMembers,
	}

	_, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), request, recorder)
	require.Error(t, err)
	assert.False(t, recorder.membership)
	assert.Zero(t, recorder.events)
	assert.Equal(t, 1, recorder.compensations)
	assert.Equal(t, []string{
		"pg:authorize:" + string(request.EntryPoint),
		"mongo:reset",
		"mongo:head",
		"mongo:baseline:" + string(messenger.MembershipBaselineChannelOrDM),
		"mongo:baseline:" + string(messenger.MembershipBaselineMention),
		"pg:final-recheck-and-commit",
		"mongo:compensate",
	}, recorder.calls)

	_, err = messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), request, recorder)
	require.Error(t, err)
	assert.False(t, recorder.membership)
	assert.Zero(t, recorder.events)
	assert.Equal(t, 2, recorder.compensations, "orphan cleanup/reconciliation must be repeat-safe")
}

func TestPhase151MembershipGrantRejectsMismatchedPGScopeBeforeDurableState(t *testing.T) {
	recorder := newPhase151MembershipLifecycleRecorder()
	request := messenger.MembershipGrantRequest{
		TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
		TargetKind: messenger.MembershipTargetKindChannel,
		EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
	}
	scope := phase151AuthorizedGrantScope(request)
	scope.ActorID = uuid.New()
	recorder.authorizationScope = &scope

	_, err := messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), request, recorder)

	require.Error(t, err)
	assert.Equal(t, []string{"pg:authorize:" + string(request.EntryPoint)}, recorder.calls)
	assert.Empty(t, recorder.baselines)
	assert.Zero(t, recorder.events)
	assert.Zero(t, recorder.compensations)
}
func TestPhase151RevokeIsPGFirstAndCleanupFailureCannotRestoreDisclosure(t *testing.T) {
	for _, reason := range []messenger.MembershipRevocationReason{
		messenger.MembershipRevocationReasonLeave,
		messenger.MembershipRevocationReasonProjectDeactivated,
		messenger.MembershipRevocationReasonUserDisabled,
	} {
		t.Run(string(reason), func(t *testing.T) {
			recorder := newPhase151MembershipLifecycleRecorder()
			recorder.membership = true
			recorder.staleReadState = true
			recorder.autoFollow = true
			recorder.manualFollow = true
			recorder.manualUnfollow = true
			recorder.cleanupErr = fmt.Errorf("%s cleanup unavailable", reason)
			request := messenger.MembershipRevocationRequest{
				TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
				TargetKind: messenger.MembershipTargetKindChannel, Reason: reason,
			}

			result, err := messenger.ExportPhase151ExecuteMembershipRevoke(context.Background(), request, recorder)
			require.NoError(t, err)
			assert.True(t, result.CleanupFailed)
			assert.True(t, recorder.accessDenied)
			assert.False(t, recorder.membership)
			assert.True(t, recorder.manualUnfollow, "cleanup never deletes non-disclosive manual suppression")
			assert.Equal(t, []string{"pg:revoke:" + string(reason), "mongo:cleanup", "event:revoke"}, recorder.calls)

			_, err = messenger.ExportPhase151ExecuteMembershipGrant(context.Background(), messenger.MembershipGrantRequest{
				TenantID: request.TenantID, UserID: request.UserID, TargetID: request.TargetID,
				TargetKind: request.TargetKind, EntryPoint: messenger.MembershipGrantEntryPointJoinChannel,
			}, recorder)
			require.Error(t, err, "later membership-scoped disclosure must be denied by PG before any Mongo operation")
			assert.Equal(t, []string{
				"pg:revoke:" + string(reason),
				"mongo:cleanup",
				"event:revoke",
				"pg:authorize:" + string(messenger.MembershipGrantEntryPointJoinChannel),
			}, recorder.calls)
		})
	}
}

func TestPhase151RevokePublishesRemovalEvenWhenCleanupFailsAndSurfacesPublishFailure(t *testing.T) {
	recorder := newPhase151MembershipLifecycleRecorder()
	recorder.cleanupErr = errors.New("cleanup unavailable")
	recorder.publishRevokeErr = errors.New("publish unavailable")
	request := messenger.MembershipRevocationRequest{
		TenantID: uuid.New(), UserID: uuid.New(), TargetID: uuid.New(),
		TargetKind: messenger.MembershipTargetKindChannel,
		Reason: messenger.MembershipRevocationReasonLeave,
		LifecycleGeneration: 3,
	}

	result, err := messenger.ExportPhase151ExecuteMembershipRevoke(context.Background(), request, recorder)

	require.Error(t, err)
	assert.True(t, result.CleanupFailed)
	assert.True(t, result.PublishFailed)
	assert.Equal(t, 1, recorder.revocationEvents)
	assert.Equal(t, []string{"pg:revoke:leave", "mongo:cleanup", "event:revoke"}, recorder.calls)
}
func TestPhase151StaleRevocationCleanupCannotEraseARejoinedMembershipGeneration(t *testing.T) {
	tenantID, userID, targetID := uuid.New(), uuid.New(), uuid.New()
	stale := messenger.MembershipLifecycleFence{
		TenantID: tenantID, ActorID: userID, TargetID: targetID, ChannelID: targetID,
		TargetKind: messenger.MembershipTargetKindChannel, LifecycleGeneration: 7,
	}
	rejoined := stale
	rejoined.LifecycleGeneration = 8

	assert.False(t, messenger.ExportPhase151MembershipCleanupFenceMatches(stale, rejoined),
		"retry from an old lifecycle generation must not clean new rejoin state")
	assert.True(t, messenger.ExportPhase151MembershipCleanupFenceMatches(rejoined, rejoined))

	foreignTenant := rejoined
	foreignTenant.TenantID = uuid.New()
	assert.False(t, messenger.ExportPhase151MembershipCleanupFenceMatches(rejoined, foreignTenant))
}
