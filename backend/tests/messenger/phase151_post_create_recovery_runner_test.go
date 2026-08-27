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

var _ messenger.Phase151ClaimedPostCreateRecoveryRepository = (*phase151ClaimedRecoveryRunnerRepository)(nil)
var _ messenger.Phase151PostCreateRecoveryClock = (*phase151ClaimedRecoveryRunnerClock)(nil)

type phase151ClaimedRecoveryRunnerRepository struct {
	calls             []string
	acquisition       messenger.Phase151PostCreateRecoveryAcquisition
	acquired          bool
	claimErr          error
	followErr         error
	replyErr          error
	completeErr       error
	invalidationErr   error
	claimScope        messenger.AuthorizedMessageProducerScope
	followScope       messenger.AuthorizedMessageProducerScope
	replyScope        messenger.AuthorizedMessageProducerScope
	completeScope     messenger.AuthorizedMessageProducerScope
	invalidationScope messenger.AuthorizedMessageProducerScope
	claimNow          time.Time
	claimLease        time.Duration
	completedAt       time.Time
}

func (r *phase151ClaimedRecoveryRunnerRepository) ClaimPendingPostCreateRecovery(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
	_ uuid.UUID,
	now time.Time,
	lease time.Duration,
) (messenger.Phase151PostCreateRecoveryAcquisition, bool, error) {
	r.calls = append(r.calls, "claim")
	r.claimScope = scope
	r.claimNow = now
	r.claimLease = lease
	return r.acquisition, r.acquired, r.claimErr
}

func (r *phase151ClaimedRecoveryRunnerRepository) EnsureAutoThreadFollows(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "auto-follow")
	r.followScope = scope
	return r.followErr
}

func (r *phase151ClaimedRecoveryRunnerRepository) ReconcileReplyCount(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "reply-count")
	r.replyScope = scope
	return r.replyErr
}

func (r *phase151ClaimedRecoveryRunnerRepository) CompleteClaimedPostCreateRecovery(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
	_ messenger.Phase151PostCreateRecoveryClaim,
	completedAt time.Time,
) error {
	r.calls = append(r.calls, "complete")
	r.completeScope = scope
	r.completedAt = completedAt
	return r.completeErr
}

func (r *phase151ClaimedRecoveryRunnerRepository) PublishUnreadInvalidation(
	_ context.Context,
	scope messenger.AuthorizedMessageProducerScope,
	_ *messenger.MessageDocument,
) error {
	r.calls = append(r.calls, "publish-unread")
	r.invalidationScope = scope
	return r.invalidationErr
}

type phase151ClaimedRecoveryRunnerClock struct {
	times []time.Time
	calls int
}

func (c *phase151ClaimedRecoveryRunnerClock) Now() time.Time {
	if c == nil || c.calls >= len(c.times) {
		return time.Time{}
	}
	value := c.times[c.calls]
	c.calls++
	return value
}

func phase151ClaimedRecoveryRunnerFixture() (
	messenger.AuthorizedMessageProducerScope,
	*messenger.MessageDocument,
	uuid.UUID,
	time.Time,
	time.Time,
	time.Duration,
	messenger.Phase151PostCreateRecoveryAcquisition,
) {
	base := time.Date(2026, time.August, 27, 9, 0, 0, 0, time.UTC)
	targetAuthority := messenger.AuthorizedMessageProducerScope{
		TenantID: uuid.New(), ActorID: uuid.New(), ChannelID: uuid.New(),
		Producer: messenger.TrustedMessageProducerUser,
	}
	message := &messenger.MessageDocument{
		ID: "phase151-claimed-runner", ChannelID: targetAuthority.ChannelID, SenderID: uuid.New(),
		ClientNonce: "phase151-claimed-runner", CreatedAt: base,
		PostCreateEffectsState: messenger.PostCreateEffectsStatePending,
		PostCreateProducerProvenance: &messenger.PostCreateProducerProvenance{
			Version:  messenger.PostCreateProducerProvenanceVersionV1,
			Producer: messenger.TrustedMessageProducerUser,
		},
	}
	ownerID := uuid.New()
	claimNow := base.Add(time.Minute)
	completedAt := claimNow.Add(30 * time.Second)
	lease := 2 * time.Minute
	leaseUntil := claimNow.Add(lease)
	claimed := *message
	claimed.PostCreateRecoveryOwnerID = ownerID
	claimed.PostCreateRecoveryFence = 4
	claimed.PostCreateRecoveryLeaseUntil = &leaseUntil
	claimedProvenance := *message.PostCreateProducerProvenance
	claimed.PostCreateProducerProvenance = &claimedProvenance
	acquisition := messenger.Phase151PostCreateRecoveryAcquisition{
		Claim: messenger.Phase151PostCreateRecoveryClaim{
			TenantID: targetAuthority.TenantID, MessageID: message.ID,
			ChannelID: targetAuthority.ChannelID, OwnerID: ownerID,
			Fence: claimed.PostCreateRecoveryFence, LeaseUntil: leaseUntil,
		},
		Message: claimed,
	}
	return targetAuthority, message, ownerID, claimNow, completedAt, lease, acquisition
}

func TestPhase151ClaimedRecoveryRunnerExecutesRootEffectsUnderExactClaim(t *testing.T) {
	target, message, ownerID, claimNow, completedAt, lease, acquisition := phase151ClaimedRecoveryRunnerFixture()
	repository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

	acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
		context.Background(), target, message, ownerID, lease, clock, repository,
	)

	require.NoError(t, err)
	require.True(t, acquired)
	assert.Equal(t, []string{"claim", "auto-follow", "complete", "publish-unread"}, repository.calls)
	assert.Equal(t, claimNow, repository.claimNow)
	assert.Equal(t, lease, repository.claimLease)
	assert.Equal(t, completedAt, repository.completedAt)
	for _, scope := range []messenger.AuthorizedMessageProducerScope{
		repository.claimScope, repository.followScope, repository.completeScope, repository.invalidationScope,
	} {
		assert.Equal(t, target.TenantID, scope.TenantID)
		assert.Equal(t, target.ChannelID, scope.ChannelID)
		assert.Equal(t, message.SenderID, scope.ActorID)
		assert.Equal(t, messenger.TrustedMessageProducerUser, scope.Producer)
	}
	assert.NotEqual(t, target.ActorID, repository.claimScope.ActorID)
	assert.Equal(t, 2, clock.calls)
}

func TestPhase151ClaimedRecoveryRunnerReconcilesRepliesBeforeCompletion(t *testing.T) {
	target, message, ownerID, claimNow, completedAt, lease, acquisition := phase151ClaimedRecoveryRunnerFixture()
	message.ThreadID = "phase151-claimed-root"
	acquisition.Message.ThreadID = message.ThreadID
	repository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

	acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
		context.Background(), target, message, ownerID, lease, clock, repository,
	)

	require.NoError(t, err)
	require.True(t, acquired)
	assert.Equal(t, []string{"claim", "auto-follow", "reply-count", "complete", "publish-unread"}, repository.calls)
	assert.Equal(t, message.SenderID, repository.replyScope.ActorID)
}

func TestPhase151ClaimedRecoveryRunnerClaimMissStopsWithoutEffects(t *testing.T) {
	target, message, ownerID, claimNow, _, lease, _ := phase151ClaimedRecoveryRunnerFixture()
	repository := &phase151ClaimedRecoveryRunnerRepository{}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow}}

	acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
		context.Background(), target, message, ownerID, lease, clock, repository,
	)

	require.NoError(t, err)
	assert.False(t, acquired)
	assert.Equal(t, []string{"claim"}, repository.calls)
	assert.Equal(t, 1, clock.calls)
}

func TestPhase151ClaimedRecoveryRunnerRejectsMalformedAcquisitionBeforeEffects(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		mutate func(*messenger.Phase151PostCreateRecoveryAcquisition)
	}{
		{name: "wrong claim owner", mutate: func(value *messenger.Phase151PostCreateRecoveryAcquisition) {
			value.Claim.OwnerID = uuid.New()
		}},
		{name: "wrong claim fence", mutate: func(value *messenger.Phase151PostCreateRecoveryAcquisition) {
			value.Claim.Fence++
		}},
		{name: "wrong claim lease", mutate: func(value *messenger.Phase151PostCreateRecoveryAcquisition) {
			value.Claim.LeaseUntil = value.Claim.LeaseUntil.Add(time.Second)
		}},
		{name: "foreign claimed id", mutate: func(value *messenger.Phase151PostCreateRecoveryAcquisition) {
			value.Message.ID = "foreign"
		}},
		{name: "producer drift", mutate: func(value *messenger.Phase151PostCreateRecoveryAcquisition) {
			value.Message.PostCreateProducerProvenance.Producer = messenger.TrustedMessageProducerForward
		}},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			target, message, ownerID, claimNow, completedAt, lease, acquisition := phase151ClaimedRecoveryRunnerFixture()
			scenario.mutate(&acquisition)
			repository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
			clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

			acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
				context.Background(), target, message, ownerID, lease, clock, repository,
			)

			require.Error(t, err)
			assert.True(t, acquired)
			assert.Equal(t, []string{"claim"}, repository.calls)
			assert.Equal(t, 1, clock.calls)
		})
	}
}

func TestPhase151ClaimedRecoveryRunnerPreservesStageErrorsAndStops(t *testing.T) {
	stageErr := errors.New("phase151 claimed runner stage error")
	for _, scenario := range []struct {
		name          string
		reply         bool
		configure     func(*phase151ClaimedRecoveryRunnerRepository)
		expectedCalls []string
		expectedClaim bool
	}{
		{name: "claim", configure: func(r *phase151ClaimedRecoveryRunnerRepository) { r.claimErr = stageErr }, expectedCalls: []string{"claim"}},
		{name: "auto follow", configure: func(r *phase151ClaimedRecoveryRunnerRepository) { r.followErr = stageErr }, expectedCalls: []string{"claim", "auto-follow"}, expectedClaim: true},
		{name: "reply count", reply: true, configure: func(r *phase151ClaimedRecoveryRunnerRepository) { r.replyErr = stageErr }, expectedCalls: []string{"claim", "auto-follow", "reply-count"}, expectedClaim: true},
		{name: "completion", configure: func(r *phase151ClaimedRecoveryRunnerRepository) { r.completeErr = stageErr }, expectedCalls: []string{"claim", "auto-follow", "complete"}, expectedClaim: true},
		{name: "invalidation", configure: func(r *phase151ClaimedRecoveryRunnerRepository) { r.invalidationErr = stageErr }, expectedCalls: []string{"claim", "auto-follow", "complete", "publish-unread"}, expectedClaim: true},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			target, message, ownerID, claimNow, completedAt, lease, acquisition := phase151ClaimedRecoveryRunnerFixture()
			if scenario.reply {
				message.ThreadID = "phase151-claimed-root"
				acquisition.Message.ThreadID = message.ThreadID
			}
			repository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
			scenario.configure(repository)
			clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

			acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
				context.Background(), target, message, ownerID, lease, clock, repository,
			)

			require.ErrorIs(t, err, stageErr)
			assert.Equal(t, scenario.expectedClaim, acquired)
			assert.Equal(t, scenario.expectedCalls, repository.calls)
		})
	}
}

func TestPhase151ClaimedRecoveryRunnerRejectsInvalidInputsAndCompletionTime(t *testing.T) {
	target, message, ownerID, claimNow, completedAt, lease, acquisition := phase151ClaimedRecoveryRunnerFixture()
	repository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
	clock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}

	for _, scenario := range []struct {
		name       string
		ctx        context.Context
		repository messenger.Phase151ClaimedPostCreateRecoveryRepository
		clock      messenger.Phase151PostCreateRecoveryClock
	}{
		{name: "nil context", repository: repository, clock: clock},
		{name: "nil repository", ctx: context.Background(), clock: clock},
		{name: "nil clock", ctx: context.Background(), repository: repository},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
				scenario.ctx, target, message, ownerID, lease, scenario.clock, scenario.repository,
			)
			require.ErrorIs(t, err, messenger.ErrPostCreatePipelineInvalid)
			assert.False(t, acquired)
		})
	}

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		freshRepository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
		freshClock := &phase151ClaimedRecoveryRunnerClock{times: []time.Time{claimNow, completedAt}}
		acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
			ctx, target, message, ownerID, lease, freshClock, freshRepository,
		)
		require.ErrorIs(t, err, context.Canceled)
		assert.False(t, acquired)
		assert.Empty(t, freshRepository.calls)
		assert.Zero(t, freshClock.calls)
	})

	for _, scenario := range []struct {
		name          string
		times         []time.Time
		expectedClaim bool
		expectedCalls []string
		expectedError error
	}{
		{name: "zero claim clock", times: []time.Time{{}}, expectedCalls: nil, expectedError: messenger.ErrPostCreatePipelineInvalid},
		{name: "zero completion clock", times: []time.Time{claimNow, {}}, expectedClaim: true, expectedCalls: []string{"claim", "auto-follow"}, expectedError: messenger.ErrPostCreatePipelineInvalid},
		{name: "expired completion clock", times: []time.Time{claimNow, acquisition.Claim.LeaseUntil}, expectedClaim: true, expectedCalls: []string{"claim", "auto-follow"}, expectedError: messenger.ErrPostCreateRecoveryLeaseLost},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			freshRepository := &phase151ClaimedRecoveryRunnerRepository{acquisition: acquisition, acquired: true}
			freshClock := &phase151ClaimedRecoveryRunnerClock{times: scenario.times}
			acquired, err := messenger.ExportPhase151ExecuteClaimedPostCreateRecovery(
				context.Background(), target, message, ownerID, lease, freshClock, freshRepository,
			)
			require.ErrorIs(t, err, scenario.expectedError)
			assert.Equal(t, scenario.expectedClaim, acquired)
			assert.Equal(t, scenario.expectedCalls, freshRepository.calls)
		})
	}
}
